package tests

import (
	"fmt"
	"github.com/opentibiabr/login-server/src/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const TestKey = "TEST_KEY"

func TestGetEnvStr(t *testing.T) {
	os.Unsetenv(TestKey)

	defaultValue := "default"
	value := "value"

	os.Setenv(TestKey, value)
	actualValue := config.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, value, actualValue)

	os.Setenv(TestKey, defaultValue)
	actualValue = config.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, defaultValue, actualValue)
}

func TestGetEnvStrNotSetGetsDefault(t *testing.T) {
	os.Unsetenv(TestKey)

	defaultValue := "random_default"

	assert.Equal(t, defaultValue, config.GetEnvStr(TestKey, defaultValue))
}

func TestGetEnvInt(t *testing.T) {
	os.Unsetenv(TestKey)

	defaultValue := 737
	value := 100

	os.Setenv(TestKey, fmt.Sprintf("%d", value))
	assert.Equal(t, value, config.GetEnvInt(TestKey, defaultValue))

	os.Setenv(TestKey, fmt.Sprintf("%d", defaultValue))
	assert.Equal(t, defaultValue, config.GetEnvInt(TestKey, defaultValue))
}

func TestGetEnvIntNotSetGetsDefault(t *testing.T) {
	os.Unsetenv(TestKey)

	defaultValue := 333

	value := config.GetEnvInt(TestKey, defaultValue)
	assert.Equal(t, defaultValue, value)
}

func TestLoad(t *testing.T) {
	defaultString := "default_string"
	defaultNumber := "8080"

	os.Setenv(config.EnvLoginPortKey, defaultNumber)
	os.Setenv(config.EnvServerIpKey, defaultString)
	os.Setenv(config.EnvServerNameKey, defaultString)
	os.Setenv(config.EnvServerPortKey, defaultNumber)
	os.Setenv(config.EnvServerLocationKey, defaultString)
	os.Setenv(config.EnvDBHostKey, defaultString)
	os.Setenv(config.EnvDBPortKey, defaultNumber)
	os.Setenv(config.EnvDBNameKey, defaultString)
	os.Setenv(config.EnvDBUserKey, defaultString)
	os.Setenv(config.EnvDBPassKey, defaultString)

	expectedConfigs := config.Configs{
		LoginPort: 8080,
		GameServerConfigs: config.GameServerConfigs{
			IP:       defaultString,
			Name:     defaultString,
			Port:     8080,
			Location: defaultString,
		},
		DBConfigs: config.DBConfigs{
			Host: defaultString,
			Name: defaultString,
			Port: 8080,
			User: defaultString,
			Pass: defaultString,
		},
	}

	c := config.Configs{}
	c.Load()

	assert.Equal(t, expectedConfigs, c)
}
