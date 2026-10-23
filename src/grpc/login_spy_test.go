package grpc_login_server

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const loginSpyTargetQuery = `SELECT p.name, p.level, p.sex, p.vocation, p.looktype, p.lookhead, p.lookbody, p.looklegs, p.lookfeet, p.lookaddons, p.lastlogin FROM account_livestream_spy_targets st INNER JOIN players p ON p.id = st.target_id INNER JOIN accounts a ON a.id = st.account_id WHERE st.account_id = \? AND a.type = 6 AND p.deletion = 0 AND NOT EXISTS \( SELECT 1 FROM players owned WHERE owned.account_id = st.account_id AND owned.name = CONCAT\('Spy: ', p.name\) \) ORDER BY p.name ASC LIMIT 64`
const spyFixtureTrustedExpiry = uint64(4102444800)

// All persistence and second-factor calls use sqlmock. No listener or real DB
// is initialized; Login is invoked directly with an isolated account fixture.
func spyLoginFixture(t *testing.T, accountType uint32, owned *sqlmock.Rows, verified bool) (*GrpcServer, sqlmock.Sqlmock, *login_proto_messages.LoginRequest) {
	t.Helper()
	t.Setenv("SERVER_PATH", "")
	t.Setenv("SERVER_NAME", "Spy login fixture")
	t.Setenv("SERVER_IP", "127.0.0.1")
	t.Setenv("SERVER_PORT", "7172")
	t.Setenv("VOCATIONS", "None,Sorcerer,Druid,Paladin,Knight")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	in := &login_proto_messages.LoginRequest{Email: "god@example.invalid", Password: "fixture-password"}
	passwordHash := sha1.Sum([]byte(in.Password))
	mock.ExpectQuery("SELECT id, type, premdays, lastday FROM accounts").
		WithArgs(in.Email, in.Email, fmt.Sprintf("%x", passwordHash)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "premdays", "lastday"}).AddRow(42, accountType, 0, 0))
	mock.ExpectQuery(`SELECT name, level, sex, vocation, looktype, lookhead, lookbody, looklegs, lookfeet, lookaddons, lastlogin from players where account_id = "42"`).
		WillReturnRows(owned).RowsWillBeClosed()
	mock.ExpectQuery("SELECT secret_encrypted, last_used_step, blocked_until FROM account_authenticators").
		WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows([]string{"secret_encrypted", "last_used_step", "blocked_until"}).AddRow("fixture-envelope", nil, 0))
	if verified {
		rawToken := []byte(strings.Repeat("a", 32))
		in.TrustedDeviceToken = base64.RawURLEncoding.EncodeToString(rawToken)
		tokenHash := sha256.Sum256(rawToken)
		mock.ExpectQuery("SELECT expires_at FROM account_trusted_devices").
			WithArgs(uint32(42), tokenHash[:], sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"expires_at"}).AddRow(spyFixtureTrustedExpiry))
		mock.ExpectExec("UPDATE account_trusted_devices").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), uint32(42), tokenHash[:], sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	return &GrpcServer{DB: db}, mock, in
}

type spySessionHash struct {
	value string
}

func (hash *spySessionHash) Match(value driver.Value) bool {
	encoded, ok := value.(string)
	if !ok || len(encoded) != 64 {
		return false
	}
	hash.value = encoded
	return true
}

func expectSpyNormalSession(mock sqlmock.Sqlmock) *spySessionHash {
	hash := &spySessionHash{}
	mock.ExpectExec("INSERT INTO `account_sessions`").
		WithArgs(hash, uint32(42), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	return hash
}

func assertSpyNormalSession(t *testing.T, response *login_proto_messages.LoginResponse, hash *spySessionHash, lastLogin uint32) {
	t.Helper()
	require.Nil(t, response.GetError())
	require.NotNil(t, response.GetSession())
	key := response.GetSession().GetSessionKey()
	raw, err := hex.DecodeString(key)
	require.NoError(t, err)
	assert.Len(t, raw, 32)
	keyHash := sha256.Sum256([]byte(key))
	assert.Equal(t, fmt.Sprintf("%x", keyHash), hash.value)
	assert.Equal(t, lastLogin, response.GetSession().GetLastLogin())
	assert.False(t, response.GetSession().GetIsPremium())
	assert.Zero(t, response.GetSession().GetPremiumUntil())
	assert.NotEqual(t, "@livestream", key)
	assert.NotContains(t, key, "\n")
	assert.NotEqual(t, "", response.GetTrustedDeviceToken())
	assert.Equal(t, spyFixtureTrustedExpiry, response.GetTrustedDeviceExpiresAt())
	trustedRaw, err := base64.RawURLEncoding.DecodeString(response.GetTrustedDeviceToken())
	require.NoError(t, err)
	assert.Len(t, trustedRaw, 32)
}

func TestLoginSpyAppendsVirtualCharactersWithoutChangingOwnedCharactersOrSession(t *testing.T) {
	owned := sqlmock.NewRows(livestreamCasterColumns()).
		AddRow("Zulu", 100, 1, 4, 128, 10, 20, 30, 40, 2, 123).
		AddRow("Alpha", 50, 0, 2, 129, 1, 2, 3, 4, 0, 456)
	ls, mock, in := spyLoginFixture(t, 6, owned, true)
	mock.ExpectQuery(loginSpyTargetQuery).WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows(livestreamCasterColumns()).
			AddRow("Offline Target", 250, 1, 3, 130, 11, 22, 33, 44, 3, 999999))
	hash := expectSpyNormalSession(mock)
	response, err := ls.Login(context.Background(), in)
	require.NoError(t, err)
	assertSpyNormalSession(t, response, hash, 456)
	characters := response.GetPlayData().GetCharacters()
	require.Len(t, characters, 3)
	assert.Equal(t, "Zulu", characters[0].Info.Name)
	assert.Equal(t, "Alpha", characters[1].Info.Name)
	assert.Equal(t, "Spy: Offline Target", characters[2].Info.Name)
	assert.Equal(t, uint32(100), characters[0].Info.Level)
	assert.Equal(t, "Knight", characters[0].Info.Vocation)
	assert.Equal(t, uint32(2), characters[0].Outfit.Addons)
	assert.Equal(t, uint32(129), characters[1].Outfit.LookType)
	assert.Equal(t, uint32(44), characters[2].Outfit.LookFeet)
	assert.Equal(t, uint32(999999), characters[2].Info.LastLogin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginSpyDoesNotCreateOwnedCharacterMetadata(t *testing.T) {
	ls, mock, in := spyLoginFixture(t, 6, sqlmock.NewRows(livestreamCasterColumns()), true)
	mock.ExpectQuery(loginSpyTargetQuery).WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows(livestreamCasterColumns()).
			AddRow("Saved Target", 100, 1, 4, 128, 1, 2, 3, 4, 0, 999999))
	hash := expectSpyNormalSession(mock)
	response, err := ls.Login(context.Background(), in)
	require.NoError(t, err)
	assertSpyNormalSession(t, response, hash, 0)
	require.Len(t, response.GetPlayData().GetCharacters(), 1)
	assert.Equal(t, "Spy: Saved Target", response.GetPlayData().GetCharacters()[0].Info.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginSpyKeepsOwnedCharacterWhenDatabaseOmitsCollidingVirtualName(t *testing.T) {
	owned := sqlmock.NewRows(livestreamCasterColumns()).
		AddRow("spy: Offline Target", 100, 1, 4, 128, 1, 2, 3, 4, 0, 123)
	ls, mock, in := spyLoginFixture(t, 6, owned, true)
	// Name collisions use the database collation; the query, rather than Go
	// string comparison, omits the saved target from this mocked result.
	mock.ExpectQuery(loginSpyTargetQuery).WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows(livestreamCasterColumns()))
	hash := expectSpyNormalSession(mock)
	response, err := ls.Login(context.Background(), in)
	require.NoError(t, err)
	assertSpyNormalSession(t, response, hash, 123)
	characters := response.GetPlayData().GetCharacters()
	require.Len(t, characters, 1)
	assert.Equal(t, "spy: Offline Target", characters[0].Info.Name)
	assert.Equal(t, uint32(100), characters[0].Info.Level)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginSpyNeverQueriesTargetsForNonGODAccounts(t *testing.T) {
	for _, accountType := range []uint32{1, 4, 5, 7} {
		t.Run(fmt.Sprintf("type %d", accountType), func(t *testing.T) {
			owned := sqlmock.NewRows(livestreamCasterColumns()).AddRow("Own Character", 100, 1, 4, 128, 1, 2, 3, 4, 0, 123)
			ls, mock, in := spyLoginFixture(t, accountType, owned, true)
			hash := expectSpyNormalSession(mock)
			response, err := ls.Login(context.Background(), in)
			require.NoError(t, err)
			assertSpyNormalSession(t, response, hash, 123)
			require.Len(t, response.GetPlayData().GetCharacters(), 1)
			assert.Equal(t, "Own Character", response.GetPlayData().GetCharacters()[0].Info.Name)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLoginSpyWaitsForSecondFactorBeforeQueryingTargetsOrCreatingSession(t *testing.T) {
	ls, mock, in := spyLoginFixture(t, 6, sqlmock.NewRows(livestreamCasterColumns()), false)
	response, err := ls.Login(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, response.GetError())
	assert.Equal(t, uint32(serviceerrors.CodeAuthenticatorRequired), response.GetError().Code)
	assert.Nil(t, response.GetPlayData())
	assert.Nil(t, response.GetSession())
	assert.Empty(t, response.GetTrustedDeviceToken())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginSpyPreservesNormalLoginForMissingTableOrFreshRoleRevocation(t *testing.T) {
	for _, missingTable := range []bool{false, true} {
		owned := sqlmock.NewRows(livestreamCasterColumns()).AddRow("Own Character", 100, 1, 4, 128, 1, 2, 3, 4, 0, 123)
		ls, mock, in := spyLoginFixture(t, 6, owned, true)
		query := mock.ExpectQuery(loginSpyTargetQuery).WithArgs(uint32(42))
		if missingTable {
			query.WillReturnError(&mysqlDriver.MySQLError{Number: 1146, Message: "Table 'game.account_livestream_spy_targets' doesn't exist"})
		} else {
			query.WillReturnRows(sqlmock.NewRows(livestreamCasterColumns()))
		}
		hash := expectSpyNormalSession(mock)
		response, err := ls.Login(context.Background(), in)
		require.NoError(t, err)
		assertSpyNormalSession(t, response, hash, 123)
		require.Len(t, response.GetPlayData().GetCharacters(), 1)
		assert.Equal(t, "Own Character", response.GetPlayData().GetCharacters()[0].Info.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestLoginSpyStorageErrorsPreserveOwnedCharactersAndSession(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		rows *sqlmock.Rows
	}{
		{name: "query failure", err: errors.New("fixture query failed")},
		{
			name: "scan failure after first target",
			rows: sqlmock.NewRows(livestreamCasterColumns()).
				AddRow("First", 100, 1, 4, 128, 1, 2, 3, 4, 0, 999999).
				AddRow("Bad Level", "not a number", 1, 4, 128, 1, 2, 3, 4, 0, 999999),
		},
		{
			name: "terminal row failure after first target",
			rows: sqlmock.NewRows(livestreamCasterColumns()).
				AddRow("First", 100, 1, 4, 128, 1, 2, 3, 4, 0, 999999).
				AddRow("Second", 100, 1, 4, 128, 1, 2, 3, 4, 0, 999999).
				RowError(1, errors.New("fixture terminal row failure")),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			owned := sqlmock.NewRows(livestreamCasterColumns()).
				AddRow("Own Character", 100, 1, 4, 128, 1, 2, 3, 4, 0, 123)
			ls, mock, in := spyLoginFixture(t, 6, owned, true)
			query := mock.ExpectQuery(loginSpyTargetQuery).WithArgs(uint32(42))
			if test.err != nil {
				query.WillReturnError(test.err)
			} else {
				query.WillReturnRows(test.rows).RowsWillBeClosed()
			}
			hash := expectSpyNormalSession(mock)
			response, err := ls.Login(context.Background(), in)
			require.NoError(t, err)
			assertSpyNormalSession(t, response, hash, 123)
			assert.NotEqual(t, in.TrustedDeviceToken, response.GetTrustedDeviceToken())
			characters := response.GetPlayData().GetCharacters()
			require.Len(t, characters, 1)
			assert.Equal(t, "Own Character", characters[0].Info.Name)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
