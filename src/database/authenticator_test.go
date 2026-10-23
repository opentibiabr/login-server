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
	"github.com/opentibiabr/login-server/src/authenticator"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authenticatorSelect = `SELECT secret_encrypted, last_used_step, blocked_until
		FROM account_authenticators WHERE account_id = ?`

func TestVerifyAuthenticatorAllowsOptionalAccountWithoutEnrollment(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnError(sql.ErrNoRows)

	err := (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, "", "")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorRequiresEnrollmentForStaff(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnError(sql.ErrNoRows)

	err := (&Account{ID: 42, Type: accountTypeGameMaster}).VerifyAuthenticator(context.Background(), db, "", "")

	assertPublicError(t, err, serviceerrors.CodeStaffAuthenticatorRequired, "STAFF_AUTHENTICATOR_REQUIRED")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorRequiresTokenForEnrolledAccount(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnRows(authenticatorRow("v1:unused", nil, 0))

	err := (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, "", "")

	assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyLoginSecondFactorAcceptsAndRotatesTrustedDevice(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(10_000, 0)
	previousNow := trustedDeviceNow
	trustedDeviceNow = func() time.Time { return now }
	t.Cleanup(func() { trustedDeviceNow = previousNow })

	raw := []byte("0123456789abcdef0123456789abcdef")
	token := base64.RawURLEncoding.EncodeToString(raw)
	tokenHash := sha256.Sum256(raw)
	expiresAt := uint64(now.Add(24 * time.Hour).Unix())
	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).
		WillReturnRows(authenticatorRow("v1:unused", nil, 0))
	mock.ExpectQuery("SELECT expires_at FROM account_trusted_devices").
		WithArgs(uint32(42), tokenHash[:], now.Unix()).
		WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(expiresAt))
	mock.ExpectExec("UPDATE account_trusted_devices").
		WithArgs(sqlmock.AnyArg(), now.Unix(), uint32(42), tokenHash[:], now.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	verification, err := (&Account{ID: 42, Type: 1}).VerifyLoginSecondFactor(
		context.Background(), db, "", token, "")

	require.NoError(t, err)
	assert.False(t, verification.VerifiedWithAuthenticator)
	assert.NotEmpty(t, verification.TrustedDeviceToken)
	assert.Equal(t, expiresAt, verification.TrustedDeviceExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorAcceptsUnusedTokenAndResetsFailures(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(3_000, 0)
	previousNow := authenticatorNow
	authenticatorNow = func() time.Time { return now }
	t.Cleanup(func() { authenticatorNow = previousNow })

	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	secret := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	envelope, err := authenticator.EncryptSecret(secret, key, 42, []byte("0123456789ab"))
	require.NoError(t, err)
	token, err := authenticator.TokenAtTime(secret, now.Unix())
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnRows(authenticatorRow(envelope, int64(98), 0))
	mock.ExpectExec("UPDATE account_authenticators").
		WithArgs(int64(100), uint32(42), now.Unix(), int64(100)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, token, key)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorRejectsReplay(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(3_000, 0)
	previousNow := authenticatorNow
	authenticatorNow = func() time.Time { return now }
	t.Cleanup(func() { authenticatorNow = previousNow })

	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	secret := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	envelope, err := authenticator.EncryptSecret(secret, key, 42, []byte("0123456789ab"))
	require.NoError(t, err)
	token, err := authenticator.TokenAtTime(secret, now.Unix())
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnRows(authenticatorRow(envelope, int64(100), 0))

	err = (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, token, key)

	assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorRecordsInvalidTokenAndLocksFifthFailure(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(3_000, 0)
	previousNow := authenticatorNow
	authenticatorNow = func() time.Time { return now }
	t.Cleanup(func() { authenticatorNow = previousNow })

	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	secret := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	envelope, err := authenticator.EncryptSecret(secret, key, 42, []byte("0123456789ab"))
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnRows(authenticatorRow(envelope, nil, 0))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT failed_attempts, blocked_until").
		WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows([]string{"failed_attempts", "blocked_until"}).AddRow(uint32(4), int64(0)))
	mock.ExpectExec("UPDATE account_authenticators").
		WithArgs(uint32(5), now.Unix()+30, uint32(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, "000000", key)

	assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVerifyAuthenticatorRejectsBlockedAccountWithoutDecrypting(t *testing.T) {
	db, mock := newAuthenticatorMock(t)
	now := time.Unix(3_000, 0)
	previousNow := authenticatorNow
	authenticatorNow = func() time.Time { return now }
	t.Cleanup(func() { authenticatorNow = previousNow })

	mock.ExpectQuery(regexp.QuoteMeta(authenticatorSelect)).WithArgs(uint32(42)).WillReturnRows(authenticatorRow("v1:unused", nil, now.Unix()+30))

	err := (&Account{ID: 42, Type: 1}).VerifyAuthenticator(context.Background(), db, "123456", "")

	assertPublicError(t, err, serviceerrors.CodeAuthenticatorRequired, "AUTHENTICATOR_REQUIRED")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func newAuthenticatorMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func authenticatorRow(envelope string, lastUsed interface{}, blocked int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"secret_encrypted", "last_used_step", "blocked_until"}).
		AddRow(envelope, lastUsed, blocked)
}

func assertPublicError(t *testing.T, err error, code int, name string) {
	t.Helper()
	publicErr, ok := serviceerrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, code, publicErr.Code)
	assert.Equal(t, name, publicErr.Name)
}
