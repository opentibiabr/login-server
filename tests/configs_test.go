package tests

import (
	"fmt"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log"
	"os"
	"strings"
	"testing"
)

const TestKey = "TEST_KEY"

func TestGetEnvStr(t *testing.T) {
	defaultValue := "default"
	value := "value"

	os.Setenv(TestKey, value)
	actualValue := configs.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, value, actualValue)

	os.Setenv(TestKey, defaultValue)
	actualValue = configs.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, defaultValue, actualValue)
	os.Unsetenv(TestKey)
}

func TestGetEnvStrNotSetGetsDefault(t *testing.T) {

	defaultValue := "random_default"

	assert.Equal(t, defaultValue, configs.GetEnvStr(TestKey, defaultValue))
	os.Unsetenv(TestKey)
}

func TestGetEnvInt(t *testing.T) {
	defaultValue := 737
	value := 100

	os.Setenv(TestKey, fmt.Sprintf("%d", value))
	assert.Equal(t, value, configs.GetEnvInt(TestKey, defaultValue))

	os.Setenv(TestKey, fmt.Sprintf("%d", defaultValue))
	assert.Equal(t, defaultValue, configs.GetEnvInt(TestKey, defaultValue))
	os.Unsetenv(TestKey)
}

func TestGetEnvIntNotSetGetsDefault(t *testing.T) {
	defaultValue := 333

	value := configs.GetEnvInt(TestKey, defaultValue)
	assert.Equal(t, defaultValue, value)
	os.Unsetenv(TestKey)
}

func TestInit(t *testing.T) {
	SetGameConfigs()
	expectedConfigs := configs.GlobalConfigs{
		LoginServerConfigs: configs.LoginServerConfigs{
			Http: configs.HttpLoginConfigs{
				Ip:   defaultString,
				Port: defaultNumber,
			},
			Tcp: configs.TcpLoginConfigs{
				Ip:   defaultString,
				Port: defaultNumber,
			},
			RateLimiter: configs.RateLimiter{
				Burst: defaultNumber,
				Rate:  rate.Limit(defaultNumber),
			},
		},
		GameServerConfigs: configs.GameServerConfigs{
			IP:       defaultString,
			Name:     defaultString,
			Port:     8080,
			Location: defaultString,
		},
		DBConfigs: configs.DBConfigs{
			Host: defaultString,
			Name: defaultString,
			Port: 8080,
			User: defaultString,
			Pass: defaultString,
		},
	}

	assert.Equal(t, expectedConfigs, configs.GetGlobalConfigs())

	UnsetGameConfigs()
}

func TestDefaultGlobalConfigs(t *testing.T) {
	expectedConfigs := configs.GlobalConfigs{
		LoginServerConfigs: configs.LoginServerConfigs{
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
		},
		GameServerConfigs: configs.GameServerConfigs{
			IP:       "127.0.0.1",
			Name:     "Canary",
			Port:     7172,
			Location: "BRA",
		},
		DBConfigs: configs.DBConfigs{
			Host: "127.0.0.1",
			Name: "canary",
			Port: 3306,
			User: "canary",
			Pass: "canary",
		},
	}

	assert.Equal(t, expectedConfigs, configs.GetGlobalConfigs())
}

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

func TestDBConfigsFormat(t *testing.T) {
	expectedFormat := "Database: 127.0.0.1:3306/canary"

	dbConfigs := configs.GetDBConfigs()
	assert.Equal(t, expectedFormat, dbConfigs.Format())
}

func TestDBConfigsConnectionString(t *testing.T) {
	expectedFormat := "canary:canary@tcp(127.0.0.1:3306)/canary"

	dbConfigs := configs.GetDBConfigs()
	assert.Equal(t, expectedFormat, dbConfigs.GetConnectionString())
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

func SetGameConfigs() {
	SetEnvKeys(
		[]string{
			configs.EnvLoginIpKey,
			configs.EnvServerIpKey,
			configs.EnvServerNameKey,
			configs.EnvServerLocationKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
		},
		[]string{
			configs.EnvServerPortKey,
			configs.EnvLoginHttpPortKey,
			configs.EnvLoginTcpPortKey,
			configs.EnvRateLimiterBurstKey,
			configs.EnvRateLimiterRateKey,
			configs.EnvDBPortKey,
		},
	)
}

func UnsetGameConfigs() {
	UnsetEnvKeys(
		[]string{
			configs.EnvLoginIpKey,
			configs.EnvServerIpKey,
			configs.EnvServerNameKey,
			configs.EnvServerLocationKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
			configs.EnvServerPortKey,
			configs.EnvLoginHttpPortKey,
			configs.EnvLoginTcpPortKey,
			configs.EnvRateLimiterBurstKey,
			configs.EnvRateLimiterRateKey,
			configs.EnvDBPortKey,
		},
	)
}
