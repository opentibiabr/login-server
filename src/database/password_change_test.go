package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/passwords"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	flagQuery   = `^SELECT password_change_required FROM accounts WHERE id = \?$`
	detectQuery = `^SELECT COUNT\(\*\) FROM information_schema\.COLUMNS WHERE TABLE_SCHEMA = DATABASE\(\) AND TABLE_NAME = 'accounts' AND COLUMN_NAME = 'password_change_required'$`
)

func withPasswordChangeMode(t *testing.T, mode passwordChangeMode) {
	t.Helper()
	setPasswordChangeMode(mode)
	t.Cleanup(func() { setPasswordChangeMode(passwordChangeOff) })
}

func flagRows(value int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"password_change_required"}).AddRow(value)
}

func assertPasswordChangeRequired(t *testing.T, err error) {
	t.Helper()
	publicErr, ok := serviceerrors.FromError(err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, serviceerrors.CodePasswordChangeRequired, publicErr.Code)
	assert.Equal(t, "PASSWORD_CHANGE_REQUIRED", publicErr.Name)
}

func TestFlaggedAccountWithTheRightPasswordIsRefused(t *testing.T) {
	for name, stored := range map[string]string{"sha1": fixtureSHA1, "argon2id": fixtureArgon2id, "wrapped": fixtureWrapped} {
		t.Run(name, func(t *testing.T) {
			withPasswordChangeMode(t, passwordChangeOn)
			db, mock := passwordFixture(t, stored)
			mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).WillReturnRows(flagRows(1))
			// An upgrade or a session row would be an unexpected statement, which the next assertion catches.
			forbidUpgrade(mock)

			acc, err := login(db, fixturePassword, passwords.NewHasher(1, true))

			assertPasswordChangeRequired(t, err)
			assert.Nil(t, acc)
			assertNoUpgrade(t, mock)
		})
	}
}

func TestFlaggedAccountDoesNotCreateASession(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOn)
	db, mock := passwordFixture(t, fixtureSHA1)
	mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).WillReturnRows(flagRows(1))

	_, err := login(db, fixturePassword, nil)

	assertPasswordChangeRequired(t, err)
	// No INSERT INTO account_sessions was expected, and none was issued.
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFlaggedAccountWithAWrongPasswordGetsInvalidCredentials(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOn)
	db, mock := passwordFixture(t, fixtureSHA1)

	_, err := login(db, "wrong-password", nil)

	assertInvalidCredentials(t, err)
	// The flag was never read.
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnflaggedAccountLogsInAndIsUpgraded(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOn)
	db, mock := passwordFixture(t, fixtureSHA1)
	mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).WillReturnRows(flagRows(0))
	expectColumnWidth(mock, 255)
	mock.ExpectExec(upgradeQuery).WithArgs(currentArgon2id{}, uint32(40), fixtureSHA1, passwords.MinColumnLength).
		WillReturnResult(sqlmock.NewResult(0, 1))

	acc, err := login(db, fixturePassword, passwords.NewHasher(1, true))

	require.NoError(t, err)
	assert.Equal(t, uint32(40), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithoutTheColumnNothingIsReadAndLoginWorks(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOff)
	db, mock := passwordFixture(t, fixtureSHA1)

	acc, err := login(db, fixturePassword, nil)

	require.NoError(t, err)
	assert.Equal(t, uint32(40), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAMissingColumnErrorSwitchesTheFeatureOff(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOn)
	db, mock := passwordFixture(t, fixtureSHA1)
	mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).
		WillReturnError(&mysqlDriver.MySQLError{Number: 1054, Message: "Unknown column"})

	acc, err := login(db, fixturePassword, nil)

	require.NoError(t, err)
	assert.Equal(t, uint32(40), acc.ID)
	assert.Equal(t, passwordChangeOff, currentPasswordChangeMode())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnUnreadableFlagRefusesTheLogin(t *testing.T) {
	withPasswordChangeMode(t, passwordChangeOn)
	db, mock := passwordFixture(t, fixtureSHA1)
	mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).WillReturnError(errors.New("lost connection"))

	acc, err := login(db, fixturePassword, nil)

	require.Error(t, err)
	assert.Nil(t, acc)
	publicErr, ok := serviceerrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, serviceerrors.CodeAccountDataUnavailable, publicErr.Code)
}

func TestDetectionTurnsTheFeatureOnOnlyWhenTheColumnExists(t *testing.T) {
	for name, test := range map[string]struct {
		count int
		want  bool
	}{"present": {1, true}, "missing": {0, false}} {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(func() { setPasswordChangeMode(passwordChangeOff) })
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { db.Close() })
			mock.ExpectQuery(detectQuery).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(test.count))

			assert.Equal(t, test.want, DetectPasswordChangeRequired(context.Background(), db))
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAFailedDetectionIsRetriedAtTheNextLogin(t *testing.T) {
	t.Cleanup(func() { setPasswordChangeMode(passwordChangeOff) })
	db, mock := passwordFixture(t, fixtureSHA1)
	// The startup lookup fails (it used another connection pool state), so the mode stays pending.
	setPasswordChangeMode(passwordChangePending)
	mock.ExpectQuery(detectQuery).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(flagQuery).WithArgs(uint32(40)).WillReturnRows(flagRows(1))

	_, err := login(db, fixturePassword, nil)

	assertPasswordChangeRequired(t, err)
	assert.Equal(t, passwordChangeOn, currentPasswordChangeMode())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPendingDetectionThatFailsAgainRefusesTheLogin(t *testing.T) {
	withPasswordChangeMode(t, passwordChangePending)
	db, mock := passwordFixture(t, fixtureSHA1)
	mock.ExpectQuery(detectQuery).WillReturnError(sql.ErrConnDone)

	acc, err := login(db, fixturePassword, nil)

	require.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, passwordChangePending, currentPasswordChangeMode())
}

func TestPasswordChangeRequiredStateNamesEveryMode(t *testing.T) {
	defer setPasswordChangeMode(passwordChangeOff)
	for mode, want := range map[passwordChangeMode]string{passwordChangeOff: "off", passwordChangeOn: "on", passwordChangePending: "pending"} {
		setPasswordChangeMode(mode)
		if got := PasswordChangeRequiredState(); got != want {
			t.Fatalf("mode %d: got %q, want %q", mode, got, want)
		}
	}
}
