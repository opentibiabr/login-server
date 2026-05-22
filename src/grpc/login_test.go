package grpc_login_server

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
)

func TestLoginReturnsConfigurationErrorWhenServerNameMismatches(t *testing.T) {
	response := buildConfigurationErrorResponse(&configs.ConfigurationError{
		Code: configs.ConfigErrorCodeServerNameMismatch,
		Name: configs.ConfigErrorServerNameMismatch,
	}, false)

	assert.Equal(t, uint32(configs.ConfigErrorCodeServerNameMismatch), response.GetError().GetCode())
	assert.Equal(t, "Game world configuration error. Please contact support. Error: SERVER_NAME_MISMATCH (LS-1001).", response.GetError().GetMessage())
}

func TestBuildConfigurationErrorResponseIncludesAdminHint(t *testing.T) {
	response := buildConfigurationErrorResponse(&configs.ConfigurationError{
		Code: configs.ConfigErrorCodeServerNameMismatch,
		Name: configs.ConfigErrorServerNameMismatch,
	}, true)

	assert.Equal(t, uint32(configs.ConfigErrorCodeServerNameMismatch), response.GetError().GetCode())
	assert.Contains(t, response.GetError().GetMessage(), "SERVER_NAME_MISMATCH (LS-1001)")
	assert.Contains(t, response.GetError().GetMessage(), "Admin hint: Make SERVER_NAME in the login-server .env match serverName in Canary config.lua")
}

func TestBuildLoginErrorResponseIncludesAdminHint(t *testing.T) {
	response := buildLoginErrorResponse(serviceerrors.LoginService(
		serviceerrors.CodeSessionStorageUnavailable,
		"SESSION_STORAGE_UNAVAILABLE",
		assert.AnError,
	), true)

	assert.Equal(t, uint32(serviceerrors.CodeSessionStorageUnavailable), response.GetError().GetCode())
	assert.Contains(t, response.GetError().GetMessage(), "SESSION_STORAGE_UNAVAILABLE (LS-2004)")
	assert.Contains(t, response.GetError().GetMessage(), "Admin hint: Create or migrate the account_sessions table")
}

func TestBuildLoginErrorResponseOmitsAdminHintForNormalAccounts(t *testing.T) {
	response := buildLoginErrorResponse(serviceerrors.LoginService(
		serviceerrors.CodeSessionStorageUnavailable,
		"SESSION_STORAGE_UNAVAILABLE",
		assert.AnError,
	), false)

	assert.Equal(t, uint32(serviceerrors.CodeSessionStorageUnavailable), response.GetError().GetCode())
	assert.Contains(t, response.GetError().GetMessage(), "SESSION_STORAGE_UNAVAILABLE (LS-2004)")
	assert.NotContains(t, response.GetError().GetMessage(), "Admin hint:")
}

func TestBuildLivestreamSession(t *testing.T) {
	session := buildLivestreamSession("any-password")

	assert.False(t, session.IsPremium)
	assert.Zero(t, session.PremiumUntil)
	assert.Zero(t, session.LastLogin)
	assert.Equal(t, database.LivestreamSessionAccount+"\nany-password", session.SessionKey)

	emptyPasswordSession := buildLivestreamSession("")
	assert.Equal(t, database.LivestreamSessionAccount+"\n", emptyPasswordSession.SessionKey)
}

func TestBuildLivestreamUnavailableResponse(t *testing.T) {
	response := buildLivestreamUnavailableResponse()

	assert.NotNil(t, response.GetError())
	assert.Equal(t, uint32(DefaultLoginErrorCode), response.GetError().Code)
	assert.Equal(t, livestreamUnavailableMessage, response.GetError().Message)
}

func TestLoginLivestreamSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows(livestreamCasterColumns()).
		AddRow("Caster", uint32(250), uint32(1), 4, uint32(128), uint32(10), uint32(20), uint32(30), uint32(40), uint32(0), uint32(123))
	mock.ExpectQuery("SELECT p.name").
		WithArgs(database.LivestreamStatusActive).
		WillReturnRows(rows)

	response, err := (&GrpcServer{DB: db}).loginLivestream(context.Background(), "stream-password")
	assert.NoError(t, err)
	assert.Nil(t, response.GetError())
	assert.NotNil(t, response.GetPlayData())
	assert.Len(t, response.GetPlayData().GetCharacters(), 1)
	assert.NotNil(t, response.GetSession())
	assert.Equal(t, database.LivestreamSessionAccount+"\nstream-password", response.GetSession().GetSessionKey())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginLivestreamUnavailableForEmptyCasters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows(livestreamCasterColumns())
	mock.ExpectQuery("SELECT p.name").
		WithArgs(database.LivestreamStatusActive).
		WillReturnRows(rows)

	response, err := (&GrpcServer{DB: db}).loginLivestream(context.Background(), "stream-password")
	assert.NoError(t, err)
	assert.Equal(t, buildLivestreamUnavailableResponse(), response)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginLivestreamUnavailableForMissingTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT p.name").
		WithArgs(database.LivestreamStatusActive).
		WillReturnError(&mysqlDriver.MySQLError{
			Number:  1146,
			Message: "Table 'otserv.active_livestream_casters' doesn't exist",
		})

	response, err := (&GrpcServer{DB: db}).loginLivestream(context.Background(), "stream-password")
	assert.NoError(t, err)
	assert.Equal(t, buildLivestreamUnavailableResponse(), response)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginLivestreamReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	dbErr := errors.New("connection refused")
	mock.ExpectQuery("SELECT p.name").
		WithArgs(database.LivestreamStatusActive).
		WillReturnError(dbErr)

	response, err := (&GrpcServer{DB: db}).loginLivestream(context.Background(), "stream-password")
	assert.Nil(t, response)
	assert.ErrorIs(t, err, dbErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func livestreamCasterColumns() []string {
	return []string{
		"name",
		"level",
		"sex",
		"vocation",
		"looktype",
		"lookhead",
		"lookbody",
		"looklegs",
		"lookfeet",
		"lookaddons",
		"lastlogin",
	}
}
