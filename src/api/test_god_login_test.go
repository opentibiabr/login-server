package api

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/configs"
	grpc_login_server "github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestTestGodLoginHTTPUsesServerPolicyAndPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name          string
		enabled       bool
		accountType   uint32
		passwordValid bool
		enrolled      bool
		errorCode     int
	}{
		{"test GOD", true, 6, true, false, 0},
		{"strict GOD cannot opt in through request", false, 6, true, false, serviceerrors.CodeStaffAuthenticatorRequired},
		{"enrolled GOD test login", true, 6, true, true, 0},
		{"GOD invalid password", true, 6, false, false, 3},
		{"gamemaster keeps enrollment", true, 4, true, false, serviceerrors.CodeStaffAuthenticatorRequired},
		{"senior gamemaster keeps token", true, 5, true, true, serviceerrors.CodeAuthenticatorRequired},
		{"enrolled player keeps token", true, 1, true, true, serviceerrors.CodeAuthenticatorRequired},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			configPath := filepath.Join(path, "config.lua")
			require.NoError(t, os.WriteFile(configPath, []byte("serverName = \"OTServBR-Global\"\nauthType = \"session\"\n"), 0600))
			t.Setenv(configs.EnvServerPathKey, path)
			t.Setenv(configs.EnvServerPortKey, "27002")
			gConfig := configs.GetGlobalConfigs()
			gConfig.DBConfigs.Name, gConfig.DBConfigs.User = "fixture_test", "fixture_test_login"
			gConfig.DBConfigs.Host, gConfig.DBConfigs.Port = "127.0.0.1", 3306
			gConfig.LoginServerConfigs.Http.Port = 28002
			gConfig.LoginServerConfigs.TestGodLogin = configs.TestGodLoginConfigs{
				Enabled: fmt.Sprint(test.enabled), Environment: "test", DBName: "fixture_test", DBUser: "fixture_test_login",
				DBHost: "127.0.0.1", DBPort: "3306",
				ServerPath: path, GamePort: "27002", HTTPPort: "28002",
			}
			loginService := grpc_login_server.Initialize(gConfig)
			require.NotNil(t, loginService.DB)
			require.NoError(t, loginService.DB.Close())
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			loginService.DB = db
			query := mock.ExpectQuery("SELECT id, type, premdays, lastday FROM accounts").WithArgs(
				"god@example.invalid", "god@example.invalid", fmt.Sprintf("%x", sha1.Sum([]byte("fixture-password"))))
			if !test.passwordValid {
				query.WillReturnError(sql.ErrNoRows)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"id", "type", "premdays", "lastday"}).AddRow(uint32(42), test.accountType, uint32(0), uint32(0)))
				mock.ExpectQuery("SELECT name, level, sex, vocation").WillReturnRows(sqlmock.NewRows([]string{"name", "level", "sex", "vocation", "looktype", "lookhead", "lookbody", "looklegs", "lookfeet", "lookaddons", "lastlogin"}))
				if !test.enabled || test.accountType != 6 {
					authQuery := mock.ExpectQuery("SELECT secret_encrypted, last_used_step, blocked_until").WithArgs(uint32(42))
					if test.enrolled {
						authQuery.WillReturnRows(sqlmock.NewRows([]string{"secret_encrypted", "last_used_step", "blocked_until"}).AddRow("v1:unused", nil, int64(0)))
					} else {
						authQuery.WillReturnError(sql.ErrNoRows)
					}
				}
				if test.errorCode == 0 {
					mock.ExpectExec("INSERT INTO `account_sessions`").WithArgs(sqlmock.AnyArg(), uint32(42), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			listener := bufconn.Listen(1024 * 1024)
			server := grpc.NewServer()
			login_proto_messages.RegisterLoginServiceServer(server, loginService)
			go func() { _ = server.Serve(listener) }()
			dialCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			connection, err := grpc.DialContext(dialCtx, "fixture-buffer", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}), grpc.WithInsecure(), grpc.WithBlock())
			cancel()
			require.NoError(t, err)
			t.Cleanup(func() { _ = connection.Close(); server.Stop(); _ = listener.Close() })
			manager, err := configs.NewLuaConfigManager(configPath)
			require.NoError(t, err)
			router := gin.New()
			router.POST("/login", (&Api{GrpcConnection: connection, LuaConfigManager: manager}).login)
			// Request fields cannot turn on the server-side deployment policy.
			body := []byte(`{"type":"login","email":"god@example.invalid","password":"fixture-password","trustdevice":true,"REDOT_ENVIRONMENT":"test","REDOT_TEST_GOD_LOGIN_WITHOUT_2FA":true}`)
			request := httptest.NewRequest(http.MethodPost, "https://test.example.invalid/login", bytes.NewReader(body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			assert.Equal(t, http.StatusOK, recorder.Code)
			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			if test.errorCode != 0 {
				assert.Equal(t, float64(test.errorCode), response["errorCode"])
				assert.NotContains(t, response, "session")
			} else {
				assert.NotContains(t, response, "errorCode")
				session, ok := response["session"].(map[string]interface{})
				require.True(t, ok)
				assert.Len(t, session["sessionkey"], 64)
				assert.NotContains(t, response, "trusteddevicetoken")
				assert.NotContains(t, response, "trusteddeviceexpiresat")
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
