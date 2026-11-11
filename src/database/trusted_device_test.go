package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gameDeviceSelect = `SELECT expires_at FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND token_hash = ? AND expires_at > ?`
const gameDeviceUpdate = `UPDATE account_trusted_devices
		SET token_hash = ?, last_used_at = ?
		WHERE account_id = ? AND purpose = 'game' AND token_hash = ? AND expires_at > ?`

type recordedDeviceHash struct{ value []byte }

func (hash *recordedDeviceHash) Match(value driver.Value) bool {
	bytes, ok := value.([]byte)
	if !ok || len(bytes) != sha256.Size {
		return false
	}
	hash.value = append([]byte(nil), bytes...)
	return true
}

func expectGameDevicePolicy(mock sqlmock.Sqlmock, mode string, policyError error) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM accounts WHERE id = ? FOR UPDATE`)).
		WithArgs(uint32(42)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	query := mock.ExpectQuery(regexp.QuoteMeta(`SELECT game_mode FROM account_device_policies WHERE account_id = ?`)).WithArgs(uint32(42))
	if policyError != nil {
		query.WillReturnError(policyError)
	} else {
		query.WillReturnRows(sqlmock.NewRows([]string{"game_mode"}).AddRow(mode))
	}
}

func TestRotateTrustedDeviceRejectsMalformedTokenWithoutDatabaseAccess(t *testing.T) {
	db, mock := newAuthenticatorMock(t)

	replacement, expiresAt, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, "not-a-device-token")

	assert.NoError(t, err)
	assert.Empty(t, replacement)
	assert.Zero(t, expiresAt)
	assert.False(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateTrustedDeviceReplacesSingleUseToken(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })

	raw := []byte("0123456789abcdef0123456789abcdef")
	token := base64.RawURLEncoding.EncodeToString(raw)
	tokenHash := sha256.Sum256(raw)
	expiresAt := uint64(now.Add(24 * time.Hour).Unix())
	expectGameDevicePolicy(mock, "locked", nil)
	mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expiresAt))
	mock.ExpectExec(regexp.QuoteMeta(gameDeviceUpdate)).
		WithArgs(sqlmock.AnyArg(), now.Unix(), uint32(42), tokenHash[:], now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	replacement, actualExpiry, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, token)

	require.NoError(t, err)
	decoded, err := base64.RawURLEncoding.DecodeString(replacement)
	require.NoError(t, err)
	assert.Len(t, decoded, trustedDeviceTokenBytes)
	assert.Equal(t, expiresAt, actualExpiry)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateTrustedDeviceRejectsExpiredOrRevokedToken(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })

	raw := []byte("0123456789abcdef0123456789abcdef")
	token := base64.RawURLEncoding.EncodeToString(raw)
	tokenHash := sha256.Sum256(raw)
	expectGameDevicePolicy(mock, "open", nil)
	mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	replacement, expiresAt, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, token)

	assert.NoError(t, err)
	assert.Empty(t, replacement)
	assert.Zero(t, expiresAt)
	assert.False(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIssueTrustedDeviceStoresOnlyTokenHashWithAbsoluteExpiry(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })
	expiresAt := now.Add(trustedDeviceDuration).Unix()
	storedHash := &recordedDeviceHash{}

	expectGameDevicePolicy(mock, "open", nil)
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND expires_at <= ?`)).
		WithArgs(uint32(42), now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM account_trusted_devices
		WHERE account_id = ? AND purpose = 'game' AND id NOT IN (
			SELECT id FROM (
				SELECT id FROM account_trusted_devices
				WHERE account_id = ? AND purpose = 'game' ORDER BY last_used_at DESC, id DESC LIMIT ?
			) AS recent_devices
		)`)).
		WithArgs(uint32(42), uint32(42), trustedDeviceMaxCount-1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO account_trusted_devices
		(account_id, purpose, token_hash, device_name, created_at, last_used_at, expires_at)
		VALUES (?, 'game', ?, ?, ?, ?, ?)`)).
		WithArgs(uint32(42), storedHash, "OTClient Windows", now.Unix(), now.Unix(), expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	token, actualExpiry, err := (&Account{ID: 42}).IssueTrustedDevice(context.Background(), db, "OTClient\n Windows")

	require.NoError(t, err)
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	assert.Len(t, decoded, trustedDeviceTokenBytes)
	expectedHash := sha256.Sum256(decoded)
	assert.Equal(t, expectedHash[:], storedHash.value)
	assert.Equal(t, uint64(expiresAt), actualExpiry)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIssueTrustedDevicePolicyDeniesEnrollment(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode string
		err  error
	}{
		{name: "locked", mode: "locked"},
		{name: "disabled", mode: "disabled"},
		{name: "unknown mode", mode: "unexpected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			expectGameDevicePolicy(mock, tc.mode, tc.err)
			mock.ExpectRollback()
			token, expiry, err := (&Account{ID: 42}).IssueTrustedDevice(context.Background(), db, "fixture")
			assert.NoError(t, err)
			assert.Empty(t, token)
			assert.Zero(t, expiry)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestIssueTrustedDeviceWithoutSavedPolicyEnrolls(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })
	storedHash := &recordedDeviceHash{}

	expectGameDevicePolicy(mock, "", sql.ErrNoRows)
	mock.ExpectExec(`DELETE FROM account_trusted_devices\s+WHERE account_id = \? AND purpose = 'game' AND expires_at <= \?`).
		WithArgs(uint32(42), now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM account_trusted_devices\s+WHERE account_id = \? AND purpose = 'game' AND id NOT IN`).
		WithArgs(uint32(42), uint32(42), trustedDeviceMaxCount-1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO account_trusted_devices`).
		WithArgs(uint32(42), storedHash, "OTClient", now.Unix(), now.Unix(), now.Add(trustedDeviceDuration).Unix()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	token, expiry, err := (&Account{ID: 42}).IssueTrustedDevice(context.Background(), db, "")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint64(now.Add(trustedDeviceDuration).Unix()), expiry)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateTrustedDeviceWithoutSavedPolicyAcceptsApprovedDevice(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })

	raw := []byte("0123456789abcdef0123456789abcdef")
	tokenHash := sha256.Sum256(raw)
	expiresAt := uint64(now.Add(24 * time.Hour).Unix())
	expectGameDevicePolicy(mock, "", sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expiresAt))
	mock.ExpectExec(regexp.QuoteMeta(gameDeviceUpdate)).
		WithArgs(sqlmock.AnyArg(), now.Unix(), uint32(42), tokenHash[:], now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	replacement, _, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, base64.RawURLEncoding.EncodeToString(raw))

	require.NoError(t, err)
	assert.NotEmpty(t, replacement)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrustedDevicePolicyUnavailableFailsClosed(t *testing.T) {
	for _, issue := range []bool{false, true} {
		t.Run(map[bool]string{false: "rotate", true: "issue"}[issue], func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			failure := errors.New("policy storage unavailable")
			expectGameDevicePolicy(mock, "", failure)
			mock.ExpectRollback()
			acc := &Account{ID: 42}
			if issue {
				token, expiry, err := acc.IssueTrustedDevice(context.Background(), db, "fixture")
				assert.ErrorIs(t, err, failure)
				assert.Empty(t, token)
				assert.Zero(t, expiry)
			} else {
				token, expiry, valid, err := acc.RotateTrustedDevice(context.Background(), db, base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
				assert.ErrorIs(t, err, failure)
				assert.Empty(t, token)
				assert.Zero(t, expiry)
				assert.False(t, valid)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRotateTrustedDeviceDisabledOrUnknownPolicyCannotUseCredential(t *testing.T) {
	for _, mode := range []string{"disabled", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			expectGameDevicePolicy(mock, mode, nil)
			mock.ExpectRollback()
			token, expiry, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
			assert.NoError(t, err)
			assert.Empty(t, token)
			assert.Zero(t, expiry)
			assert.False(t, valid)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRotateTrustedDeviceDoesNotReturnUncommittedOrReplayedCredential(t *testing.T) {
	for _, tc := range []struct {
		name        string
		affected    int64
		commitError error
	}{
		{name: "compare and swap lost", affected: 0},
		{name: "commit failed", affected: 1, commitError: errors.New("commit failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			raw := make([]byte, 32)
			hash := sha256.Sum256(raw)
			expectGameDevicePolicy(mock, "open", nil)
			mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).WithArgs(uint32(42), hash[:], sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(uint64(4102444800)))
			mock.ExpectExec(regexp.QuoteMeta(gameDeviceUpdate)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint32(42), hash[:], sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, tc.affected))
			if tc.affected == 0 {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit().WillReturnError(tc.commitError)
			}
			token, expiry, valid, err := (&Account{ID: 42}).RotateTrustedDevice(context.Background(), db, base64.RawURLEncoding.EncodeToString(raw))
			assert.ErrorIs(t, err, tc.commitError)
			assert.Empty(t, token)
			assert.Zero(t, expiry)
			assert.False(t, valid)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTrustedDeviceSessionRechecksRevocationAfterRotation(t *testing.T) {
	for _, revoked := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled after verification", true: "device revoked after verification"}[revoked], func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			acc := &Account{ID: 42}
			raw := make([]byte, 32)
			hash := sha256.Sum256(raw)
			expectGameDevicePolicy(mock, "locked", nil)
			mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).WithArgs(uint32(42), hash[:], sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(uint64(4102444800)))
			mock.ExpectExec(regexp.QuoteMeta(gameDeviceUpdate)).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint32(42), hash[:], sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			rotated, _, valid, err := acc.RotateTrustedDevice(context.Background(), db, base64.RawURLEncoding.EncodeToString(raw))
			require.NoError(t, err)
			require.True(t, valid)
			if revoked {
				expectGameDevicePolicy(mock, "locked", nil)
				mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).WithArgs(uint32(42), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			} else {
				expectGameDevicePolicy(mock, "disabled", nil)
			}
			mock.ExpectRollback()
			session, err := acc.CreateTrustedDeviceSession(context.Background(), db, rotated)
			assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
			assert.Empty(t, session)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTrustedDeviceSessionCommitsOnlyWithCurrentGameCredential(t *testing.T) {
	for _, commitError := range []error{nil, errors.New("session commit failed")} {
		name := "success"
		if commitError != nil {
			name = "commit failed"
		}
		t.Run(name, func(t *testing.T) {
			db, mock := newAuthenticatorMock(t)
			raw := make([]byte, 32)
			hash := sha256.Sum256(raw)
			expectGameDevicePolicy(mock, "locked", nil)
			mock.ExpectQuery(regexp.QuoteMeta(gameDeviceSelect)).WithArgs(uint32(42), hash[:], sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(uint64(4102444800)))
			mock.ExpectExec("INSERT INTO `account_sessions`").
				WithArgs(sqlmock.AnyArg(), uint32(42), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			if commitError != nil {
				mock.ExpectCommit().WillReturnError(commitError)
			} else {
				mock.ExpectCommit()
			}
			session, err := (&Account{ID: 42}).CreateTrustedDeviceSession(context.Background(), db, base64.RawURLEncoding.EncodeToString(raw))
			if commitError != nil {
				assertPublicError(t, err, serviceerrors.CodeSessionCreateFailed, "SESSION_CREATE_FAILED")
				assert.Empty(t, session)
			} else {
				assert.NoError(t, err)
				assert.Len(t, session, 64)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
