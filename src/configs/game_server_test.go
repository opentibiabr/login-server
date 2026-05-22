package configs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameServerConfigs_Format(t *testing.T) {
	type fields struct {
		Port     int
		Name     string
		IP       string
		Location string
		Config   Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Format game server configs",
		fields: fields{
			Port:     7172,
			Name:     "superb",
			IP:       "0.0.0.0",
			Location: "JPN",
		},
		want: "Connected with superb server 0.0.0.0:7172 - JPN",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameServerConfigs := &GameServerConfigs{
				Port:     tt.fields.Port,
				Name:     tt.fields.Name,
				IP:       tt.fields.IP,
				Location: tt.fields.Location,
				Config:   tt.fields.Config,
			}
			assert.Equal(t, tt.want, gameServerConfigs.Format())
		})
	}
}

func TestGetGameServerConfigs(t *testing.T) {
	tests := []struct {
		name string
		want GameServerConfigs
	}{{
		name: "Default Game Server Configs",
		want: GameServerConfigs{
			IP:       "127.0.0.1",
			Name:     DefaultServerName,
			Port:     7172,
			Location: "BRA",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetGameServerConfigs())
		})
	}
}

func TestValidateGameServerName(t *testing.T) {
	t.Run("skips validation without server path", func(t *testing.T) {
		t.Setenv(EnvServerPathKey, "")

		err := ValidateGameServerName(GameServerConfigs{Name: "LoginName"})

		assert.NoError(t, err)
	})

	t.Run("accepts matching canary server name", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(EnvServerPathKey, tempDir)
		err := os.WriteFile(filepath.Join(tempDir, "config.lua"), []byte(`serverName = "Canary"`), 0o600)
		assert.NoError(t, err)

		err = ValidateGameServerName(GameServerConfigs{Name: "Canary"})

		assert.NoError(t, err)
	})

	t.Run("accepts matching canary server name from dist config", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(EnvServerPathKey, tempDir)
		err := os.WriteFile(filepath.Join(tempDir, "config.lua.dist"), []byte(`serverName = "Canary"`), 0o600)
		assert.NoError(t, err)

		err = ValidateGameServerName(GameServerConfigs{Name: "Canary"})

		assert.NoError(t, err)
	})

	t.Run("rejects mismatching canary server name", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(EnvServerPathKey, tempDir)
		err := os.WriteFile(filepath.Join(tempDir, "config.lua"), []byte(`serverName = "Canary"`), 0o600)
		assert.NoError(t, err)

		err = ValidateGameServerName(GameServerConfigs{Name: "OtherWorld"})

		var configErr *ConfigurationError
		assert.ErrorAs(t, err, &configErr)
		assert.Equal(t, ConfigErrorCodeServerNameMismatch, configErr.Code)
		assert.Equal(t, ConfigErrorServerNameMismatch, configErr.Name)
		assert.Equal(t, `login-server SERVER_NAME="OtherWorld" but Canary config.lua serverName="Canary"`, configErr.Message)
	})

	t.Run("returns reportable error when server config is missing", func(t *testing.T) {
		t.Setenv(EnvServerPathKey, t.TempDir())

		err := ValidateGameServerName(GameServerConfigs{Name: "Canary"})

		var configErr *ConfigurationError
		assert.ErrorAs(t, err, &configErr)
		assert.Equal(t, ConfigErrorCodeServerConfigNotFound, configErr.Code)
		assert.Equal(t, ConfigErrorServerConfigNotFound, configErr.Name)
	})

	t.Run("returns reportable error when server name is missing", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(EnvServerPathKey, tempDir)
		err := os.WriteFile(filepath.Join(tempDir, "config.lua"), []byte(`authType = "session"`), 0o600)
		assert.NoError(t, err)

		err = ValidateGameServerName(GameServerConfigs{Name: "Canary"})

		var configErr *ConfigurationError
		assert.ErrorAs(t, err, &configErr)
		assert.Equal(t, ConfigErrorCodeServerConfigInvalid, configErr.Code)
		assert.Equal(t, ConfigErrorServerConfigInvalid, configErr.Name)
	})

	t.Run("returns reportable error when config path cannot be inspected", func(t *testing.T) {
		_, err := findServerConfigPath("bad\x00path")

		var configErr *ConfigurationError
		assert.ErrorAs(t, err, &configErr)
		assert.Equal(t, ConfigErrorCodeServerConfigInvalid, configErr.Code)
		assert.Equal(t, ConfigErrorServerConfigInvalid, configErr.Name)
	})
}

func TestConfigurationErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := &ConfigurationError{Cause: cause}

	assert.ErrorIs(t, err, cause)
}

func TestGetServerVocations(t *testing.T) {
	tests := []struct {
		name   string
		want   []string
		envVoc *[]string
	}{{
		name: "Default Vocations",
		want: []string{
			"None",
			"Sorcerer",
			"Druid",
			"Paladin",
			"Knight",
			"Master Sorcerer",
			"Elder Druid",
			"Royal Paladin",
			"Elite Knight",
			"Sorcerer Dawnport",
			"Druid Dawnport",
			"Paladin Dawnport",
			"Knight Dawnport",
		},
	}, {
		name: "Uses env voc",
		want: []string{
			"artista",
			"professor",
			"engenheiro",
		},
		envVoc: &[]string{
			"artista",
			"professor",
			"engenheiro",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVoc != nil {
				err := os.Setenv(EnvVocations, strings.Join(*tt.envVoc, ","))
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, GetServerVocations())
			if tt.envVoc != nil {
				err := os.Unsetenv(EnvVocations)
				assert.Nil(t, err)
			}
		})
	}
}
