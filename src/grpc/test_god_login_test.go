package grpc_login_server

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configureTestGodLogin(t *testing.T, enabled bool) configs.GlobalConfigs {
	t.Helper()
	path := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(path, "config.lua"), []byte("serverName = \"OTServBR-Global\"\nauthType = \"session\"\n"), 0600))
	t.Setenv(configs.EnvServerPathKey, path)
	t.Setenv(configs.EnvDBNameKey, "fixture_test")
	t.Setenv(configs.EnvDBUserKey, "fixture_test_login")
	t.Setenv(configs.EnvDBHostKey, "127.0.0.1")
	t.Setenv(configs.EnvDBPortKey, "3306")
	t.Setenv(configs.EnvServerPortKey, "27002")
	t.Setenv(configs.EnvLoginHttpPortKey, "28002")
	t.Setenv(configs.EnvRedotEnvironmentKey, "test")
	t.Setenv(configs.EnvTestGodLoginWithout2FAKey, fmt.Sprint(enabled))
	t.Setenv(configs.EnvTestGodLoginDBNameKey, "fixture_test")
	t.Setenv(configs.EnvTestGodLoginDBUserKey, "fixture_test_login")
	t.Setenv(configs.EnvTestGodLoginDBHostKey, "127.0.0.1")
	t.Setenv(configs.EnvTestGodLoginDBPortKey, "3306")
	t.Setenv(configs.EnvTestGodLoginServerPathKey, path)
	t.Setenv(configs.EnvTestGodLoginGamePortKey, "27002")
	t.Setenv(configs.EnvTestGodLoginHTTPPortKey, "28002")
	return configs.GetGlobalConfigs()
}

func newTestGodLoginServer(t *testing.T, config configs.GlobalConfigs) (*GrpcServer, sqlmock.Sqlmock) {
	t.Helper()
	server := Initialize(config)
	require.NoError(t, server.testGodLoginPolicyError)
	// sql.Open in Initialize is lazy: no live database connection is opened.
	require.NoError(t, server.DB.Close())
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	server.DB = db
	return server, mock
}

func expectTestGodAccount(mock sqlmock.Sqlmock, accountType uint32, validPassword bool) {
	passwordHash := fmt.Sprintf("%x", sha1.Sum([]byte("fixture-password")))
	query := mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, premdays, lastday FROM accounts WHERE (email = ? OR name = ?) AND password = ?")).
		WithArgs("god@example.invalid", "god@example.invalid", passwordHash)
	if !validPassword {
		query.WillReturnError(sql.ErrNoRows)
		return
	}
	query.WillReturnRows(sqlmock.NewRows([]string{"id", "type", "premdays", "lastday"}).AddRow(uint32(42), accountType, uint32(0), uint32(0)))
	mock.ExpectQuery("SELECT name, level, sex, vocation").WillReturnRows(sqlmock.NewRows(livestreamCasterColumns()))
}

func TestTestGodLoginGrpcPropagatesValidatedPolicy(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			config := configureTestGodLogin(t, enabled)
			server, mock := newTestGodLoginServer(t, config)
			assert.Equal(t, enabled, server.testGodLoginPolicy.Enabled())
			expectTestGodAccount(mock, 6, true)
			if enabled {
				mock.ExpectExec("INSERT INTO `account_sessions`").WithArgs(sqlmock.AnyArg(), uint32(42), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				mock.ExpectQuery("SELECT secret_encrypted, last_used_step, blocked_until").WithArgs(uint32(42)).WillReturnError(sql.ErrNoRows)
			}
			response, err := server.Login(context.Background(), &login_proto_messages.LoginRequest{
				Email: "god@example.invalid", Password: "fixture-password", TrustDevice: true,
			})
			require.NoError(t, err)
			if enabled {
				require.Nil(t, response.GetError())
				assert.Len(t, response.GetSession().GetSessionKey(), 64)
				assert.Empty(t, response.GetTrustedDeviceToken())
				assert.Zero(t, response.GetTrustedDeviceExpiresAt())
			} else {
				assert.Equal(t, uint32(serviceerrors.CodeStaffAuthenticatorRequired), response.GetError().GetCode())
				assert.Nil(t, response.GetSession())
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTestGodLoginGrpcStillRejectsInvalidPassword(t *testing.T) {
	server, mock := newTestGodLoginServer(t, configureTestGodLogin(t, true))
	expectTestGodAccount(mock, 6, false)
	response, err := server.Login(context.Background(), &login_proto_messages.LoginRequest{
		Email: "god@example.invalid", Password: "fixture-password",
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(3), response.GetError().GetCode())
	assert.Nil(t, response.GetSession())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTestGodLoginGrpcStillRequiresSessionPersistence(t *testing.T) {
	server, mock := newTestGodLoginServer(t, configureTestGodLogin(t, true))
	expectTestGodAccount(mock, 6, true)
	mock.ExpectExec("INSERT INTO `account_sessions`").WithArgs(sqlmock.AnyArg(), uint32(42), sqlmock.AnyArg()).WillReturnError(errors.New("fixture storage unavailable"))
	response, err := server.Login(context.Background(), &login_proto_messages.LoginRequest{
		Email: "god@example.invalid", Password: "fixture-password",
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(serviceerrors.CodeSessionCreateFailed), response.GetError().GetCode())
	assert.Nil(t, response.GetSession())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTestGodLoginInitializationRejectsInvalidBeforeSideEffects(t *testing.T) {
	config := configureTestGodLogin(t, true)
	config.DBConfigs.Name = "fixture_live"
	server := Initialize(config)
	require.Error(t, server.testGodLoginPolicyError)
	assert.Nil(t, server.DB)
	assert.False(t, server.testGodLoginPolicy.Enabled())
	assert.ErrorIs(t, server.Run(config), server.testGodLoginPolicyError)
	response, err := server.Login(context.Background(), &login_proto_messages.LoginRequest{
		Email: "god@example.invalid", Password: "fixture-password",
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(configs.ConfigErrorCodeTestGodLoginInvalid), response.GetError().GetCode())
	assert.Nil(t, response.GetSession())
}
