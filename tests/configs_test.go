package tests

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log"
	"os"
	"strings"
	"testing"
)

func TestGetGameServerConfigs(t *testing.T) {
	expectedConfigs := configs.GameServerConfigs{
		IP:       "127.0.0.1",
		Name:     "Canary",
		Port:     7172,
		Location: "BRA",
	}

	assert.Equal(t, expectedConfigs, configs.GetGameServerConfigs())
}

func TestGameServerConfigsFormat(t *testing.T) {
	expectedFormat := "Connected with Canary server 127.0.0.1:7172 - BRA"

	gameConfigs := configs.GetGameServerConfigs()
	assert.Equal(t, expectedFormat, gameConfigs.Format())
}

func TestGetServerVocationsDefault(t *testing.T) {
	defaultVoc := []string{
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
	}

	assert.Equal(t, defaultVoc, configs.GetServerVocations())
}

func TestGetServerVocationsFromEnv(t *testing.T) {
	newVoc := []string{
		"artista",
		"professor",
		"engenheiro",
	}

	err := os.Setenv(configs.EnvVocations, strings.Join(newVoc, ","))
	if err != nil {
		log.Print("Error trying to get vocations from env vars.")
	}

	assert.Equal(t, newVoc, configs.GetServerVocations())
	os.Unsetenv(configs.EnvVocations)
}

func TestGetDBConfigs(t *testing.T) {
	expectedConfigs := configs.DBConfigs{
		Host: "127.0.0.1",
		Name: "canary",
		Port: 3306,
		User: "canary",
		Pass: "canary",
	}

	assert.Equal(t, expectedConfigs, configs.GetDBConfigs())
}

func TestGetLoginServerConfigs(t *testing.T) {
	expectedConfigs := configs.LoginServerConfigs{
		Http: configs.HttpLoginConfigs{
			Ip:   "",
			Port: 80,
		},
		Tcp: configs.TcpLoginConfigs{
			Ip:   "",
			Port: 7171,
		},
		RateLimiter: configs.RateLimiter{
			Burst: 5,
			Rate:  rate.Limit(2),
		},
	}

	assert.Equal(t, expectedConfigs, configs.GetLoginServerConfigs())
}

func TestLoginServerConfigsFormat(t *testing.T) {
	expectedFormat := "OTBR Login Server running!!! http: :80 | tcp: :7171 | rate limit: 2/5"

	loginConfigs := configs.GetLoginServerConfigs()
	assert.Equal(t, expectedFormat, loginConfigs.Format())
}
