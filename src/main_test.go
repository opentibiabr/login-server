package main

import (
	"sync/atomic"
	"testing"

	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type startupProbeServer struct {
	runs *int32
}

func (probe startupProbeServer) GetName() string { return "isolated-startup-probe" }
func (probe startupProbeServer) Run(configs.GlobalConfigs) error {
	atomic.AddInt32(probe.runs, 1)
	return nil
}

func startupTestConfig(t *testing.T) configs.GlobalConfigs {
	t.Helper()
	path := t.TempDir()
	t.Setenv(configs.EnvServerPathKey, path)
	return configs.GlobalConfigs{
		DBConfigs:         configs.DBConfigs{Host: "127.0.0.1", Port: 3306, Name: "fixture_test", User: "fixture_login"},
		GameServerConfigs: configs.GameServerConfigs{Port: 27002},
		LoginServerConfigs: configs.LoginServerConfigs{
			Http: configs.HttpLoginConfigs{Port: 28002},
			TestGodLogin: configs.TestGodLoginConfigs{
				Enabled: "true", Environment: "test", DBHost: "127.0.0.1", DBPort: "3306", DBName: "fixture_test", DBUser: "fixture_login",
				ServerPath: path, GamePort: "27002", HTTPPort: "28002",
			},
		},
	}
}

func TestTestGodLoginStartupRejectsBeforeEitherInitializer(t *testing.T) {
	config := startupTestConfig(t)
	for _, test := range []struct {
		name   string
		change func(*configs.GlobalConfigs)
	}{
		{"invalid flag", func(c *configs.GlobalConfigs) { c.LoginServerConfigs.TestGodLogin.Enabled = "1" }},
		{"production environment", func(c *configs.GlobalConfigs) { c.LoginServerConfigs.TestGodLogin.Environment = "production" }},
		{"missing DB binding", func(c *configs.GlobalConfigs) { c.LoginServerConfigs.TestGodLogin.DBName = "" }},
		{"different database", func(c *configs.GlobalConfigs) { c.DBConfigs.Name = "fixture_live" }},
		{"different path", func(c *configs.GlobalConfigs) { c.LoginServerConfigs.TestGodLogin.ServerPath += "-other" }},
		{"different HTTP listener", func(c *configs.GlobalConfigs) { c.LoginServerConfigs.Http.Port = 8080 }},
		{"different game world", func(c *configs.GlobalConfigs) { c.GameServerConfigs.Port = 7172 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidate := config
			test.change(&candidate)
			var initializations, runs int32
			initializer := func(configs.GlobalConfigs) server.ServerInterface {
				atomic.AddInt32(&initializations, 1)
				return startupProbeServer{runs: &runs}
			}
			// This is the startup path used by main; both real constructors are
			// intercepted before they could create SQL/proxy/listener resources.
			err := runServers(candidate, initializer, initializer, 0)
			require.Error(t, err)
			assert.Equal(t, int32(0), atomic.LoadInt32(&initializations))
			assert.Equal(t, int32(0), atomic.LoadInt32(&runs))
			var configErr *configs.ConfigurationError
			require.ErrorAs(t, err, &configErr)
			assert.Equal(t, configs.ConfigErrorCodeTestGodLoginInvalid, configErr.Code)
		})
	}
}

func TestTestGodLoginStartupKeepsDisabledAndValidInitialization(t *testing.T) {
	for _, enabled := range []string{"", "false", "true"} {
		t.Run(enabled, func(t *testing.T) {
			config := startupTestConfig(t)
			config.LoginServerConfigs.TestGodLogin.Enabled = enabled
			if enabled != "true" {
				// Missing test bindings are irrelevant while the exemption is off.
				config.LoginServerConfigs.TestGodLogin = configs.TestGodLoginConfigs{Enabled: enabled}
			}
			var initializations, runs int32
			initializer := func(actual configs.GlobalConfigs) server.ServerInterface {
				assert.Equal(t, config, actual)
				atomic.AddInt32(&initializations, 1)
				return startupProbeServer{runs: &runs}
			}
			require.NoError(t, runServers(config, initializer, initializer, 0))
			assert.Equal(t, int32(2), atomic.LoadInt32(&initializations))
			assert.Equal(t, int32(2), atomic.LoadInt32(&runs))
		})
	}
}
