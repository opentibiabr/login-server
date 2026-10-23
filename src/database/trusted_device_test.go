package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT expires_at FROM account_trusted_devices
		WHERE account_id = ? AND token_hash = ? AND expires_at > ?`)).
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expiresAt))
	mock.ExpectExec("UPDATE account_trusted_devices").
		WithArgs(sqlmock.AnyArg(), now.Unix(), uint32(42), tokenHash[:], now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))

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
	mock.ExpectQuery("SELECT expires_at FROM account_trusted_devices").
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnError(sql.ErrNoRows)

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

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM accounts").
		WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint32(42)))
	mock.ExpectExec("DELETE FROM account_trusted_devices").
		WithArgs(uint32(42), now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM account_trusted_devices").
		WithArgs(uint32(42), uint32(42), trustedDeviceMaxCount-1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO account_trusted_devices").
		WithArgs(uint32(42), sqlmock.AnyArg(), "OTClient Windows", now.Unix(), now.Unix(), expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	token, actualExpiry, err := (&Account{ID: 42}).IssueTrustedDevice(context.Background(), db, "OTClient\n Windows")

	require.NoError(t, err)
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	assert.Len(t, decoded, trustedDeviceTokenBytes)
	assert.Equal(t, uint64(expiresAt), actualExpiry)
	assert.NoError(t, mock.ExpectationsWereMet())
}
