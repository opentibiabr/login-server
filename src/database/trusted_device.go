package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const (
	trustedDeviceTokenBytes = 32
	trustedDeviceDuration   = 30 * 24 * time.Hour
	trustedDeviceMaxCount   = 10
	trustedDeviceNameLength = 80
)

var trustedDeviceNow = time.Now

// Both website policy updates and game credentials lock the account first.
// Accounts that never saved a policy keep the original behavior: a fresh second
// factor may enroll a remembered device. Owners restrict it on the website.
func lockGameDevicePolicy(ctx context.Context, tx *sql.Tx, accountID uint32) (string, error) {
	var lockedAccountID uint32
	if err := tx.QueryRowContext(ctx, `SELECT id FROM accounts WHERE id = ? FOR UPDATE`, accountID).Scan(&lockedAccountID); err != nil {
		return "", err
	}
	var mode string
	err := tx.QueryRowContext(ctx, `SELECT game_mode FROM account_device_policies WHERE account_id = ?`, accountID).Scan(&mode)
	if errors.Is(err, sql.ErrNoRows) {
		return "open", nil
	}
	return mode, err
}

func (acc *Account) RotateTrustedDevice(ctx context.Context, db *sql.DB, token string) (string, uint64, bool, error) {
	if acc == nil || db == nil || token == "" {
		return "", 0, false, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != trustedDeviceTokenBytes {
		return "", 0, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, false, err
	}
	defer func() { _ = tx.Rollback() }()
	mode, err := lockGameDevicePolicy(ctx, tx, acc.ID)
	if err != nil {
		return "", 0, false, err
	}
	if mode != "open" && mode != "locked" {
		return "", 0, false, nil
	}

	tokenHash := sha256.Sum256(raw)
	now := trustedDeviceNow().Unix()
	var expiresAt uint64
	if err := tx.QueryRowContext(ctx, `SELECT expires_at FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND token_hash = ? AND expires_at > ?`, acc.ID, tokenHash[:], now).Scan(&expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}

	replacement, replacementHash, err := newTrustedDeviceToken()
	if err != nil {
		return "", 0, false, err
	}

	result, err := tx.ExecContext(ctx, `UPDATE account_trusted_devices
		SET token_hash = ?, last_used_at = ?
		WHERE account_id = ? AND purpose = 'game' AND token_hash = ? AND expires_at > ?`,
		replacementHash[:], now, acc.ID, tokenHash[:], now)
	if err != nil {
		return "", 0, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", 0, false, err
	}
	if affected != 1 {
		return "", 0, false, nil
	}
	if err := tx.Commit(); err != nil {
		return "", 0, false, err
	}
	return replacement, expiresAt, true, nil
}

func (acc *Account) IssueTrustedDevice(ctx context.Context, db *sql.DB, deviceName string) (string, uint64, error) {
	if acc == nil || db == nil {
		return "", 0, errors.New("account or database connection is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = tx.Rollback() }()
	mode, err := lockGameDevicePolicy(ctx, tx, acc.ID)
	if err != nil {
		return "", 0, err
	}
	if mode != "open" {
		return "", 0, nil
	}
	token, tokenHash, err := newTrustedDeviceToken()
	if err != nil {
		return "", 0, err
	}
	now := trustedDeviceNow().Unix()
	expiresAt := now + int64(trustedDeviceDuration/time.Second)
	deviceName = normalizeTrustedDeviceName(deviceName)

	if _, err = tx.ExecContext(ctx, `DELETE FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND expires_at <= ?`, acc.ID, now); err != nil {
		return "", 0, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND id NOT IN (
			SELECT id FROM (
				SELECT id FROM account_trusted_devices
				WHERE account_id = ? AND purpose = 'game' ORDER BY last_used_at DESC, id DESC LIMIT ?
			) AS recent_devices
		)`, acc.ID, acc.ID, trustedDeviceMaxCount-1); err != nil {
		return "", 0, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO account_trusted_devices
		(account_id, purpose, token_hash, device_name, created_at, last_used_at, expires_at)
		VALUES (?, 'game', ?, ?, ?, ?, ?)`, acc.ID, tokenHash[:], deviceName, now, now, expiresAt); err != nil {
		return "", 0, err
	}
	if err = tx.Commit(); err != nil {
		return "", 0, err
	}

	return token, uint64(expiresAt), nil
}

// Revalidate the rotated credential while persisting its game session. Otherwise
// revocation between second-factor verification and INSERT could leave a new
// session alive after the owner removed all trusted access.
func (acc *Account) CreateTrustedDeviceSession(ctx context.Context, db *sql.DB, token string) (string, error) {
	if acc == nil || db == nil {
		return "", classifyAuthenticatorDataError(errors.New("account or database connection is nil"))
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != trustedDeviceTokenBytes {
		return "", serviceerrors.AuthenticatorRequired()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	writeCtx, cancel := context.WithTimeout(ctx, sessionPersistenceTimeout)
	defer cancel()
	tx, err := db.BeginTx(writeCtx, nil)
	if err != nil {
		return "", classifyAuthenticatorDataError(err)
	}
	defer func() { _ = tx.Rollback() }()
	mode, err := lockGameDevicePolicy(writeCtx, tx, acc.ID)
	if err != nil {
		return "", classifyAuthenticatorDataError(err)
	}
	if mode != "open" && mode != "locked" {
		return "", serviceerrors.AuthenticatorRequired()
	}
	tokenHash := sha256.Sum256(raw)
	var expiresAt uint64
	if err := tx.QueryRowContext(writeCtx, `SELECT expires_at FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND token_hash = ? AND expires_at > ?`,
		acc.ID, tokenHash[:], trustedDeviceNow().Unix()).Scan(&expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", serviceerrors.AuthenticatorRequired()
		}
		return "", classifyAuthenticatorDataError(err)
	}
	session, err := acc.createSession(writeCtx, tx)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", serviceerrors.LoginService(serviceerrors.CodeSessionCreateFailed, "SESSION_CREATE_FAILED", err)
	}
	return session, nil
}

func newTrustedDeviceToken() (string, [sha256.Size]byte, error) {
	raw := make([]byte, trustedDeviceTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", [sha256.Size]byte{}, err
	}
	return base64.RawURLEncoding.EncodeToString(raw), sha256.Sum256(raw), nil
}

func normalizeTrustedDeviceName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	if name == "" {
		name = "OTClient"
	}
	if utf8.RuneCountInString(name) <= trustedDeviceNameLength {
		return name
	}
	runes := []rune(name)
	return string(runes[:trustedDeviceNameLength])
}
