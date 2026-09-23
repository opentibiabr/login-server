package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/opentibiabr/login-server/src/authenticator"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const (
	authenticatorAttemptLimit = uint32(5)
	authenticatorMaxLock      = 5 * time.Minute
)

type accountAuthenticator struct {
	secretEnvelope string
	lastUsedStep   sql.NullInt64
	blockedUntil   int64
}

var authenticatorNow = time.Now

func (acc *Account) VerifyAuthenticator(ctx context.Context, db *sql.DB, token, encodedKey string) error {
	if acc == nil || db == nil {
		return serviceerrors.LoginService(
			serviceerrors.CodeAuthenticatorDataUnavailable,
			"AUTHENTICATOR_DATA_UNAVAILABLE",
			errors.New("account or database connection is nil"),
		)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	entry, err := loadAccountAuthenticator(ctx, db, acc.ID)
	if errors.Is(err, sql.ErrNoRows) {
		if acc.IsAdmin() {
			return serviceerrors.StaffAuthenticatorRequired()
		}
		return nil
	}
	if err != nil {
		return classifyAuthenticatorDataError(err)
	}

	now := authenticatorNow().Unix()
	if entry.blockedUntil > now || token == "" {
		return serviceerrors.AuthenticatorRequired()
	}

	secret, err := authenticator.DecryptSecret(entry.secretEnvelope, encodedKey, acc.ID)
	if err != nil {
		return serviceerrors.LoginService(
			serviceerrors.CodeAuthenticatorConfigurationInvalid,
			"AUTHENTICATOR_CONFIGURATION_INVALID",
			err,
		)
	}

	matchedStep, ok := authenticator.MatchingStep(secret, token, now)
	if !ok {
		if err := recordAuthenticatorFailure(ctx, db, acc.ID, now); err != nil {
			return classifyAuthenticatorDataError(err)
		}
		return serviceerrors.AuthenticatorRequired()
	}
	if entry.lastUsedStep.Valid && matchedStep <= entry.lastUsedStep.Int64 {
		return serviceerrors.AuthenticatorRequired()
	}

	result, err := db.ExecContext(ctx, `UPDATE account_authenticators
		SET last_used_step = ?, failed_attempts = 0, blocked_until = 0
		WHERE account_id = ? AND blocked_until <= ? AND (last_used_step IS NULL OR last_used_step < ?)`,
		matchedStep, acc.ID, now, matchedStep)
	if err != nil {
		return classifyAuthenticatorDataError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return classifyAuthenticatorDataError(err)
	}
	if affected != 1 {
		return serviceerrors.AuthenticatorRequired()
	}

	return nil
}

func loadAccountAuthenticator(ctx context.Context, db *sql.DB, accountID uint32) (*accountAuthenticator, error) {
	entry := &accountAuthenticator{}
	err := db.QueryRowContext(ctx, `SELECT secret_encrypted, last_used_step, blocked_until
		FROM account_authenticators WHERE account_id = ?`, accountID).Scan(
		&entry.secretEnvelope,
		&entry.lastUsedStep,
		&entry.blockedUntil,
	)
	return entry, err
}

func recordAuthenticatorFailure(ctx context.Context, db *sql.DB, accountID uint32, now int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var priorFailures uint32
	var priorBlockedUntil int64
	if err := tx.QueryRowContext(ctx, `SELECT failed_attempts, blocked_until
		FROM account_authenticators WHERE account_id = ? FOR UPDATE`, accountID).Scan(&priorFailures, &priorBlockedUntil); err != nil {
		return err
	}
	if priorBlockedUntil > now {
		return tx.Commit()
	}

	failures := priorFailures + 1
	blockedUntil := int64(0)
	if failures >= authenticatorAttemptLimit {
		shift := failures - authenticatorAttemptLimit
		if shift > 4 {
			shift = 4
		}
		lockDuration := 30 * time.Second * time.Duration(uint64(1)<<shift)
		if lockDuration > authenticatorMaxLock {
			lockDuration = authenticatorMaxLock
		}
		blockedUntil = now + int64(lockDuration/time.Second)
	}

	_, err = tx.ExecContext(ctx, `UPDATE account_authenticators
		SET failed_attempts = ?, blocked_until = ?
		WHERE account_id = ?`, failures, blockedUntil, accountID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func classifyAuthenticatorDataError(err error) error {
	if _, ok := serviceerrors.FromError(err); ok {
		return err
	}
	return serviceerrors.LoginService(
		serviceerrors.CodeAuthenticatorDataUnavailable,
		"AUTHENTICATOR_DATA_UNAVAILABLE",
		fmt.Errorf("failed to read or update account authenticator: %w", err),
	)
}
