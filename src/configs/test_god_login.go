package configs

import (
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EnvRedotEnvironmentKey             = "REDOT_ENVIRONMENT"
	EnvTestGodLoginWithout2FAKey       = "REDOT_TEST_GOD_LOGIN_WITHOUT_2FA"
	EnvTestGodLoginDBNameKey           = "REDOT_TEST_GOD_LOGIN_DB_NAME"
	EnvTestGodLoginDBUserKey           = "REDOT_TEST_GOD_LOGIN_DB_USER"
	EnvTestGodLoginDBHostKey           = "REDOT_TEST_GOD_LOGIN_DB_HOST"
	EnvTestGodLoginDBPortKey           = "REDOT_TEST_GOD_LOGIN_DB_PORT"
	EnvTestGodLoginServerPathKey       = "REDOT_TEST_GOD_LOGIN_SERVER_PATH"
	EnvTestGodLoginGamePortKey         = "REDOT_TEST_GOD_LOGIN_GAME_PORT"
	EnvTestGodLoginHTTPPortKey         = "REDOT_TEST_GOD_LOGIN_HTTP_PORT"
	ConfigErrorCodeTestGodLoginInvalid = 1004
	ConfigErrorTestGodLoginInvalid     = "TEST_GOD_LOGIN_CONFIGURATION_INVALID"
)

// TestGodLoginConfigs describes an explicit allowlist for one test deployment.
// Every binding is required when the login-only exemption is enabled.
type TestGodLoginConfigs struct {
	Enabled     string
	Environment string
	DBName      string
	DBUser      string
	DBHost      string
	DBPort      string
	ServerPath  string
	GamePort    string
	HTTPPort    string
}

// TestGodLoginPolicy can only be enabled by successful deployment validation.
// Its zero value preserves the ordinary second-factor policy.
type TestGodLoginPolicy struct{ enabled bool }

func (policy TestGodLoginPolicy) Enabled() bool { return policy.enabled }

func getTestGodLoginConfigs() TestGodLoginConfigs {
	return TestGodLoginConfigs{
		Enabled:     GetEnvStr(EnvTestGodLoginWithout2FAKey),
		Environment: GetEnvStr(EnvRedotEnvironmentKey),
		DBName:      GetEnvStr(EnvTestGodLoginDBNameKey),
		DBUser:      GetEnvStr(EnvTestGodLoginDBUserKey),
		DBHost:      GetEnvStr(EnvTestGodLoginDBHostKey),
		DBPort:      GetEnvStr(EnvTestGodLoginDBPortKey),
		ServerPath:  GetEnvStr(EnvTestGodLoginServerPathKey),
		GamePort:    GetEnvStr(EnvTestGodLoginGamePortKey),
		HTTPPort:    GetEnvStr(EnvTestGodLoginHTTPPortKey),
	}
}

func NewTestGodLoginPolicy(gConfigs GlobalConfigs, serverPath string) (TestGodLoginPolicy, error) {
	policy := TestGodLoginPolicy{}
	settings := gConfigs.LoginServerConfigs.TestGodLogin
	if settings.Enabled == "" || settings.Enabled == "false" {
		return policy, nil
	}
	fail := func(message string) (TestGodLoginPolicy, error) {
		return TestGodLoginPolicy{}, &ConfigurationError{
			Code:    ConfigErrorCodeTestGodLoginInvalid,
			Name:    ConfigErrorTestGodLoginInvalid,
			Message: message,
		}
	}
	if settings.Enabled != "true" {
		return fail(EnvTestGodLoginWithout2FAKey + " must be true or false")
	}
	if settings.Environment != "test" {
		return fail("GOD login exemption requires REDOT_ENVIRONMENT=test")
	}
	if !exactNonemptyBinding(settings.DBName, gConfigs.DBConfigs.Name) ||
		!exactNonemptyBinding(settings.DBUser, gConfigs.DBConfigs.User) ||
		!exactNonemptyBinding(settings.DBHost, gConfigs.DBConfigs.Host) ||
		!exactPortBinding(settings.DBPort, gConfigs.DBConfigs.Port) {
		return fail("GOD login exemption database host/port/name/user bindings are missing or do not match")
	}
	if !cleanAbsolutePath(serverPath) || !cleanAbsolutePath(settings.ServerPath) || settings.ServerPath != serverPath {
		return fail("GOD login exemption SERVER_PATH binding must match an absolute clean configured path")
	}
	if !exactPortBinding(settings.GamePort, gConfigs.GameServerConfigs.Port) ||
		!exactPortBinding(settings.HTTPPort, gConfigs.LoginServerConfigs.Http.Port) {
		return fail("GOD login exemption game/HTTP port bindings are missing or do not match")
	}
	policy.enabled = true
	return policy, nil
}

func exactNonemptyBinding(expected, actual string) bool {
	return expected != "" && strings.TrimSpace(expected) == expected && expected == actual
}

func cleanAbsolutePath(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && filepath.IsAbs(value) && filepath.Clean(value) == value
}

func exactPortBinding(value string, actual int) bool {
	return actual > 0 && actual <= 65535 && value == strconv.Itoa(actual)
}
