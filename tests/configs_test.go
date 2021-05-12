package tests

import (
	"fmt"
	"login-server/config"
	"login-server/tests/testlib"
	"os"
	"testing"
)

const TestKey = "TEST_KEY"

func TestGetEnvStr(t *testing.T) {
	a := testlib.Assert{T: *t}

	os.Unsetenv(TestKey)

	defaultValue := "default"
	value := "value"
	c := config.Configs{}

	os.Setenv(TestKey, value)
	actualValue := c.GetEnvStr(TestKey, defaultValue)
	a.Equals(value, actualValue)

	os.Setenv(TestKey, defaultValue)
	actualValue = c.GetEnvStr(TestKey, defaultValue)
	a.Equals(defaultValue, actualValue)
}

func TestGetEnvStrNotSetGetsDefault(t *testing.T) {
	a := testlib.Assert{T: *t}
	os.Unsetenv(TestKey)

	defaultValue := "random_default"
	c := config.Configs{}

	a.Equals(defaultValue, c.GetEnvStr(TestKey, defaultValue))
}

func TestGetEnvInt(t *testing.T) {
	a := testlib.Assert{T: *t}
	os.Unsetenv(TestKey)

	defaultValue := 737
	value := 100
	c := config.Configs{}

	os.Setenv(TestKey, fmt.Sprintf("%d", value))
	a.Equals(value, c.GetEnvInt(TestKey, defaultValue))

	os.Setenv(TestKey, fmt.Sprintf("%d", defaultValue))
	a.Equals(defaultValue, c.GetEnvInt(TestKey, defaultValue))
}

func TestGetEnvIntNotSetGetsDefault(t *testing.T) {
	a := testlib.Assert{T: *t}
	os.Unsetenv(TestKey)

	defaultValue := 333
	c := config.Configs{}

	value := c.GetEnvInt(TestKey, defaultValue)
	a.Equals(defaultValue, value)
}

func TestLoad(t *testing.T) {
	a := testlib.Assert{T: *t}

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
			IP: defaultString,
			Name: defaultString,
			Port: 8080,
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

	a.Equals(expectedConfigs, c)
}
