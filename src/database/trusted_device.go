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
)

const (
	trustedDeviceTokenBytes = 32
	trustedDeviceDuration   = 30 * 24 * time.Hour
	trustedDeviceMaxCount   = 10
	trustedDeviceNameLength = 80
)

var trustedDeviceNow = time.Now

func (acc *Account) RotateTrustedDevice(ctx context.Context, db *sql.DB, token string) (string, uint64, bool, error) {
	if acc == nil || db == nil || token == "" {
		return "", 0, false, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != trustedDeviceTokenBytes {
		return "", 0, false, nil
	}

	tokenHash := sha256.Sum256(raw)
	now := trustedDeviceNow().Unix()
	var expiresAt uint64
	if err := db.QueryRowContext(ctx, `SELECT expires_at FROM account_trusted_devices
		WHERE account_id = ? AND token_hash = ? AND expires_at > ?`, acc.ID, tokenHash[:], now).Scan(&expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}

	replacement, replacementHash, err := newTrustedDeviceToken()
	if err != nil {
		return "", 0, false, err
	}

	result, err := db.ExecContext(ctx, `UPDATE account_trusted_devices
		SET token_hash = ?, last_used_at = ?
		WHERE account_id = ? AND token_hash = ? AND expires_at > ?`,
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
	return replacement, expiresAt, true, nil
}

func (acc *Account) IssueTrustedDevice(ctx context.Context, db *sql.DB, deviceName string) (string, uint64, error) {
	if acc == nil || db == nil {
		return "", 0, errors.New("account or database connection is nil")
	}

	token, tokenHash, err := newTrustedDeviceToken()
	if err != nil {
		return "", 0, err
	}
	now := trustedDeviceNow().Unix()
	expiresAt := now + int64(trustedDeviceDuration/time.Second)
	deviceName = normalizeTrustedDeviceName(deviceName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var lockedAccountID uint32
	if err = tx.QueryRowContext(ctx, `SELECT id FROM accounts WHERE id = ? FOR UPDATE`, acc.ID).Scan(&lockedAccountID); err != nil {
		return "", 0, err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM account_trusted_devices
		WHERE account_id = ? AND expires_at <= ?`, acc.ID, now); err != nil {
		return "", 0, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM account_trusted_devices
		WHERE account_id = ? AND id NOT IN (
			SELECT id FROM (
				SELECT id FROM account_trusted_devices
				WHERE account_id = ? ORDER BY last_used_at DESC, id DESC LIMIT ?
			) AS recent_devices
		)`, acc.ID, acc.ID, trustedDeviceMaxCount-1); err != nil {
		return "", 0, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO account_trusted_devices
		(account_id, token_hash, device_name, created_at, last_used_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)`, acc.ID, tokenHash[:], deviceName, now, now, expiresAt); err != nil {
		return "", 0, err
	}
	if err = tx.Commit(); err != nil {
		return "", 0, err
	}

	return token, uint64(expiresAt), nil
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
