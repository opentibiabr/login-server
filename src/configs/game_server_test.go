package configs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
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
			Name:     "Canary",
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
