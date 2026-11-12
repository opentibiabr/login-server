package configs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validTestGodLoginConfigs(t *testing.T) (GlobalConfigs, string) {
	t.Helper()
	serverPath := t.TempDir()
	return GlobalConfigs{
		DBConfigs:         DBConfigs{Host: "127.0.0.1", Port: 3306, Name: "fixture_test", User: "fixture_test_login"},
		GameServerConfigs: GameServerConfigs{Port: 27002},
		LoginServerConfigs: LoginServerConfigs{
			Http: HttpLoginConfigs{Port: 28002},
			TestGodLogin: TestGodLoginConfigs{
				Enabled: "true", Environment: "test",
				DBName: "fixture_test", DBUser: "fixture_test_login", ServerPath: serverPath,
				DBHost: "127.0.0.1", DBPort: "3306",
				GamePort: "27002", HTTPPort: "28002",
			},
		},
	}, serverPath
}

func TestTestGodLoginPolicyDefaultsToStrict(t *testing.T) {
	for _, value := range []string{"", "false"} {
		policy, err := NewTestGodLoginPolicy(GlobalConfigs{LoginServerConfigs: LoginServerConfigs{
			TestGodLogin: TestGodLoginConfigs{Enabled: value},
		}}, "")
		require.NoError(t, err)
		assert.False(t, policy.Enabled())
	}
	assert.False(t, (TestGodLoginPolicy{}).Enabled())
}

func TestTestGodLoginPolicyRequiresMatchingTestDeployment(t *testing.T) {
	config, serverPath := validTestGodLoginConfigs(t)
	policy, err := NewTestGodLoginPolicy(config, serverPath)
	require.NoError(t, err)
	assert.True(t, policy.Enabled())

	tests := []struct {
		name   string
		change func(*GlobalConfigs, *string)
	}{
		{"unknown flag", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.Enabled = "1" }},
		{"missing environment", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.Environment = "" }},
		{"production environment", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.Environment = "production" }},
		{"ambiguous environment", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.Environment = "TEST" }},
		{"missing database binding", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.DBName = "" }},
		{"different actual database", func(c *GlobalConfigs, _ *string) { c.DBConfigs.Name = "fixture_live" }},
		{"padded database binding", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.DBName += " " }},
		{"missing user binding", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.DBUser = "" }},
		{"different actual user", func(c *GlobalConfigs, _ *string) { c.DBConfigs.User = "fixture_live_login" }},
		{"missing database host", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.DBHost = "" }},
		{"different actual database host", func(c *GlobalConfigs, _ *string) { c.DBConfigs.Host = "192.0.2.1" }},
		{"missing database port", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.DBPort = "" }},
		{"different actual database port", func(c *GlobalConfigs, _ *string) { c.DBConfigs.Port = 3307 }},
		{"missing path binding", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.ServerPath = "" }},
		{"missing actual path", func(_ *GlobalConfigs, p *string) { *p = "" }},
		{"different actual path", func(_ *GlobalConfigs, p *string) { *p = filepath.Join(*p, "different") }},
		{"relative path", func(c *GlobalConfigs, p *string) {
			*p = "fixture-test"
			c.LoginServerConfigs.TestGodLogin.ServerPath = *p
		}},
		{"unclean path", func(c *GlobalConfigs, p *string) {
			*p += string(filepath.Separator) + "."
			c.LoginServerConfigs.TestGodLogin.ServerPath = *p
		}},
		{"missing game port", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.GamePort = "" }},
		{"different actual game port", func(c *GlobalConfigs, _ *string) { c.GameServerConfigs.Port = 7172 }},
		{"invalid actual game port", func(c *GlobalConfigs, _ *string) {
			c.GameServerConfigs.Port = 0
			c.LoginServerConfigs.TestGodLogin.GamePort = "0"
		}},
		{"missing HTTP port", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.HTTPPort = "" }},
		{"different actual HTTP port", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.Http.Port = 8080 }},
		{"noncanonical HTTP binding", func(c *GlobalConfigs, _ *string) { c.LoginServerConfigs.TestGodLogin.HTTPPort = "028002" }},
		{"invalid actual HTTP port", func(c *GlobalConfigs, _ *string) {
			c.LoginServerConfigs.Http.Port = 65536
			c.LoginServerConfigs.TestGodLogin.HTTPPort = "65536"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate, path := config, serverPath
			test.change(&candidate, &path)
			policy, err := NewTestGodLoginPolicy(candidate, path)
			require.Error(t, err)
			assert.False(t, policy.Enabled())
			var configErr *ConfigurationError
			require.ErrorAs(t, err, &configErr)
			assert.Equal(t, ConfigErrorCodeTestGodLoginInvalid, configErr.Code)
			assert.Equal(t, ConfigErrorTestGodLoginInvalid, configErr.Name)
		})
	}
}

func TestGetTestGodLoginConfigs(t *testing.T) {
	t.Setenv(EnvRedotEnvironmentKey, "test")
	t.Setenv(EnvTestGodLoginWithout2FAKey, "true")
	t.Setenv(EnvTestGodLoginDBNameKey, "fixture_test")
	t.Setenv(EnvTestGodLoginDBUserKey, "fixture_test_login")
	t.Setenv(EnvTestGodLoginDBHostKey, "127.0.0.1")
	t.Setenv(EnvTestGodLoginDBPortKey, "3306")
	t.Setenv(EnvTestGodLoginServerPathKey, "configured-test-path")
	t.Setenv(EnvTestGodLoginGamePortKey, "27002")
	t.Setenv(EnvTestGodLoginHTTPPortKey, "28002")
	assert.Equal(t, TestGodLoginConfigs{
		Enabled: "true", Environment: "test", DBName: "fixture_test", DBUser: "fixture_test_login",
		DBHost: "127.0.0.1", DBPort: "3306",
		ServerPath: "configured-test-path", GamePort: "27002", HTTPPort: "28002",
	}, GetLoginServerConfigs().TestGodLogin)
}
