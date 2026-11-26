package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/opentibiabr/login-server/src/passwords"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fixtureIdentifier = "player@example.invalid"
	fixturePassword   = "123456"
	fixtureSHA1       = "7c4a8d09ca3762af61e59520943dc26494f8941b"
	fixtureArgon2id   = "$argon2id$v=19$m=65536,t=3,p=1$L0toZHlaNFNHMlh1NTFEQg$UyY/BT/lQlf/EpFvNtJn/usQFi3APme7vimpMfJh2Fw"
	fixtureWrapped    = "sha1:$argon2id$v=19$m=65536,t=3,p=1$QTNJa2w4U0hQb2VSc1RDNQ$0Tj2i2Lq5+z9b64K+G0j/7hJGeu9GzYJ0KtfXQU+HH0"
	credentialsQuery  = `^SELECT id, type, premdays, lastday, password FROM accounts WHERE email = \? OR name = \? ORDER BY email = \? DESC, id$`
	widthQuery        = `^SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema\.COLUMNS WHERE TABLE_SCHEMA = DATABASE\(\) AND TABLE_NAME = 'accounts' AND COLUMN_NAME = 'password'$`
	upgradeQuery      = `^UPDATE accounts SET password = \? WHERE id = \? AND password = \? AND \? <= \(SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema\.COLUMNS WHERE TABLE_SCHEMA = DATABASE\(\) AND TABLE_NAME = 'accounts' AND COLUMN_NAME = 'password'\)$`
)

// currentArgon2id matches a new value that verifies the fixture password.
type currentArgon2id struct{}

func (currentArgon2id) Match(value driver.Value) bool {
	stored, ok := value.(string)
	if !ok || passwords.NeedsRehash(stored) {
		return false
	}
	matches, err := passwords.NewHasher(1, false).Verify(context.Background(), fixturePassword, stored)
	return err == nil && matches
}

func credentialRows(stored ...string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "type", "premdays", "lastday", "password"})
	for i, value := range stored {
		rows.AddRow(uint32(40+i), uint32(1+i), uint32(2), uint32(3), value)
	}
	return rows
}

func passwordFixture(t *testing.T, stored ...string) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	mock.ExpectQuery(credentialsQuery).
		WithArgs(fixtureIdentifier, fixtureIdentifier, fixtureIdentifier).
		WillReturnRows(credentialRows(stored...)).RowsWillBeClosed()
	return db, mock
}

// expectColumnWidth answers the width check which precedes the hashing of an upgrade.
func expectColumnWidth(mock sqlmock.Sqlmock, width int64) {
	mock.ExpectQuery(widthQuery).WillReturnRows(sqlmock.NewRows([]string{"CHARACTER_MAXIMUM_LENGTH"}).AddRow(width))
}

// forbidUpgrade must be paired with assertNoUpgrade. A statement the mock does not expect only returns an error,
// which the login tolerates, so the absence of the upgrade is proven by an expectation that stays unmet.
func forbidUpgrade(mock sqlmock.Sqlmock) {
	mock.ExpectExec(upgradeQuery).WillReturnResult(sqlmock.NewResult(0, 1))
}

func assertNoUpgrade(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	err := mock.ExpectationsWereMet()
	require.Error(t, err, "the stored value was rewritten")
	assert.Contains(t, err.Error(), "ExpectedExec")
}

func login(db *sql.DB, password string, hasher *passwords.Hasher) (*Account, error) {
	return LoadAccount(context.Background(), fixtureIdentifier, password, db, hasher)
}

func assertInvalidCredentials(t *testing.T, err error) {
	t.Helper()
	publicErr, ok := serviceerrors.FromError(err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, serviceerrors.CodeInvalidCredentials, publicErr.Code)
}

func TestLoadAccountAcceptsEveryStoredFormat(t *testing.T) {
	for name, stored := range map[string]string{"sha1": fixtureSHA1, "argon2id": fixtureArgon2id, "wrapped": fixtureWrapped} {
		t.Run(name, func(t *testing.T) {
			// Without rehashing, which is also the default, nothing is rewritten.
			for _, hasher := range []*passwords.Hasher{nil, passwords.NewHasher(1, false)} {
				db, mock := passwordFixture(t, stored)
				forbidUpgrade(mock)

				acc, err := login(db, fixturePassword, hasher)

				require.NoError(t, err)
				assert.Equal(t, uint32(40), acc.ID)
				assert.Equal(t, uint32(1), acc.Type)
				assert.Equal(t, uint32(2), acc.PremDays)
				assert.Equal(t, uint32(3), acc.LastDay)
				assertNoUpgrade(t, mock)
			}
		})
	}
}

func TestLoadAccountRejectsAWrongPassword(t *testing.T) {
	for name, stored := range map[string]string{"sha1": fixtureSHA1, "argon2id": fixtureArgon2id, "wrapped": fixtureWrapped, "locked": "locked-1", "malformed": "$argon2id$broken"} {
		t.Run(name, func(t *testing.T) {
			db, mock := passwordFixture(t, stored)
			forbidUpgrade(mock)

			_, err := login(db, "wrong-password", passwords.NewHasher(1, true))

			assertInvalidCredentials(t, err)
			assertNoUpgrade(t, mock)
		})
	}
}

func TestLoadAccountRejectsAnUnknownAccount(t *testing.T) {
	for _, rehash := range []bool{false, true} {
		db, mock := passwordFixture(t)

		_, err := login(db, fixturePassword, passwords.NewHasher(1, rehash))

		assertInvalidCredentials(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestLoadAccountPicksTheCandidateThatMatches(t *testing.T) {
	db, mock := passwordFixture(t, "locked-1", fixtureSHA1)

	acc, err := login(db, fixturePassword, passwords.NewHasher(1, false))

	require.NoError(t, err)
	assert.Equal(t, uint32(41), acc.ID)
	assert.Equal(t, uint32(2), acc.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadAccountUpgradesOlderValuesAfterAMatch(t *testing.T) {
	for name, stored := range map[string]string{"sha1": fixtureSHA1, "wrapped": fixtureWrapped} {
		t.Run(name, func(t *testing.T) {
			db, mock := passwordFixture(t, stored)
			expectColumnWidth(mock, 255)
			mock.ExpectExec(upgradeQuery).WithArgs(currentArgon2id{}, uint32(40), stored, passwords.MinColumnLength).
				WillReturnResult(sqlmock.NewResult(0, 1))

			_, err := login(db, fixturePassword, passwords.NewHasher(1, true))

			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLoadAccountKeepsACurrentValue(t *testing.T) {
	db, mock := passwordFixture(t, fixtureArgon2id)
	forbidUpgrade(mock)

	_, err := login(db, fixturePassword, passwords.NewHasher(1, true))

	require.NoError(t, err)
	assertNoUpgrade(t, mock)
}

func TestLoadAccountSucceedsWhenTheUpgradeFails(t *testing.T) {
	db, mock := passwordFixture(t, fixtureSHA1)
	expectColumnWidth(mock, 255)
	mock.ExpectExec(upgradeQuery).WithArgs(currentArgon2id{}, uint32(40), fixtureSHA1, passwords.MinColumnLength).
		WillReturnError(errors.New("UPDATE command denied"))

	acc, err := login(db, fixturePassword, passwords.NewHasher(1, true))

	require.NoError(t, err)
	assert.Equal(t, uint32(40), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadAccountSkipsTheHashWhenTheColumnIsTooNarrow(t *testing.T) {
	db, mock := passwordFixture(t, fixtureSHA1)
	expectColumnWidth(mock, 40)
	forbidUpgrade(mock)

	acc, err := login(db, fixturePassword, passwords.NewHasher(1, true))

	require.NoError(t, err)
	assert.Equal(t, uint32(40), acc.ID)
	assertNoUpgrade(t, mock)
}

func TestLoadAccountChecksEveryAccountWithTheIdentifier(t *testing.T) {
	stored := make([]string, 20)
	for i := range stored {
		stored[i] = "locked-1"
	}
	stored[19] = fixtureSHA1
	db, mock := passwordFixture(t, stored...)

	acc, err := login(db, fixturePassword, nil)

	require.NoError(t, err)
	assert.Equal(t, uint32(59), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadAccountRejectsAnEmptyIdentifierWithoutAQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectQuery(credentialsQuery).WillReturnRows(credentialRows(fixtureSHA1))

	for _, identifier := range []string{"", " ", "   "} {
		_, err = LoadAccount(context.Background(), identifier, fixturePassword, db, nil)
		assertInvalidCredentials(t, err)
	}

	require.Error(t, mock.ExpectationsWereMet(), "no account may be looked up")
}

func TestLoadAccountBoundsTheHashingOfOneRequest(t *testing.T) {
	// A cheap value which does not match the fixture password, so the test stays fast.
	const other = "$argon2id$v=19$m=1024,t=2,p=1$SUpqdExXMGlzelFod05DRA$mlCOM0J6Wf9zIek97/yk2RagJiE55cyBk3SHDeiNRZw"
	stored := make([]string, maxArgon2idPerLogin)
	for i := range stored {
		stored[i] = other
	}

	// Beyond the bound an Argon2id value is not computed, even when it would match.
	db, mock := passwordFixture(t, append(append([]string{}, stored...), fixtureArgon2id)...)
	_, err := login(db, fixturePassword, nil)
	assertInvalidCredentials(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// A legacy value after them is still checked.
	db, mock = passwordFixture(t, append(append([]string{}, stored...), fixtureArgon2id, fixtureSHA1)...)
	acc, err := login(db, fixturePassword, nil)
	require.NoError(t, err)
	assert.Equal(t, uint32(40+maxArgon2idPerLogin+1), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	// The last value within the bound is computed.
	db, mock = passwordFixture(t, append(append([]string{}, stored[1:]...), fixtureArgon2id)...)
	acc, err = login(db, fixturePassword, nil)
	require.NoError(t, err)
	assert.Equal(t, uint32(40+maxArgon2idPerLogin-1), acc.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadAccountIsNotHiddenByABrokenSibling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	rows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "type", "premdays", "lastday", "password"}).
			AddRow(uint32(1), uint32(1), -5, uint32(0), fixtureSHA1).
			AddRow(uint32(2), uint32(1), uint32(0), uint32(0), nil).
			AddRow(uint32(3), uint32(1), uint32(0), uint32(0), fixtureSHA1)
	}
	mock.ExpectQuery(credentialsQuery).WillReturnRows(rows()).RowsWillBeClosed()
	mock.ExpectQuery(credentialsQuery).WillReturnRows(rows()).RowsWillBeClosed()

	acc, err := login(db, fixturePassword, nil)
	require.NoError(t, err)
	assert.Equal(t, uint32(3), acc.ID)

	// Without a match the unreadable row is reported, as the account may be that one.
	_, err = login(db, "wrong-password", nil)
	publicErr, ok := serviceerrors.FromError(err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, serviceerrors.CodeAccountDataUnavailable, publicErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadAccountReportsABusyHasherApartFromBadCredentials(t *testing.T) {
	previous := passwordCheckTimeout
	passwordCheckTimeout = 0
	defer func() { passwordCheckTimeout = previous }()

	for name, test := range map[string]struct {
		stored []string
		rehash bool
		busy   bool
	}{
		"argon2id value":                        {stored: []string{fixtureArgon2id}, busy: true},
		"unknown account while upgrading":       {rehash: true, busy: true},
		"wrong legacy password while upgrading": {stored: []string{"7110eda4d09e062aa5e4a390b0a572ac0d2c0220"}, rehash: true, busy: true},
		"unknown account without upgrading":     {},
	} {
		t.Run(name, func(t *testing.T) {
			db, mock := passwordFixture(t, test.stored...)

			_, err := login(db, fixturePassword, passwords.NewHasher(1, test.rehash))

			publicErr, ok := serviceerrors.FromError(err)
			require.True(t, ok, "%v", err)
			want := serviceerrors.CodeInvalidCredentials
			if test.busy {
				want = serviceerrors.CodePasswordVerificationBusy
			}
			assert.Equal(t, want, publicErr.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
