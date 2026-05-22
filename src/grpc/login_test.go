package grpc_login_server

import (
	"testing"

	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
)

func TestLoginReturnsConfigurationErrorWhenServerNameMismatches(t *testing.T) {
	response := buildConfigurationErrorResponse(&configs.ConfigurationError{
		Code: 1001,
		Name: configs.ConfigErrorServerNameMismatch,
	}, false)

	assert.Equal(t, uint32(1001), response.GetError().GetCode())
	assert.Equal(t, "Game world configuration error. Please contact support. Error: SERVER_NAME_MISMATCH (LS-1001).", response.GetError().GetMessage())
}

func TestBuildConfigurationErrorResponseIncludesAdminHint(t *testing.T) {
	response := buildConfigurationErrorResponse(&configs.ConfigurationError{
		Code: 1001,
		Name: configs.ConfigErrorServerNameMismatch,
	}, true)

	assert.Equal(t, uint32(1001), response.GetError().GetCode())
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
