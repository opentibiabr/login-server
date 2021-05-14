package tests

import (
	"fmt"
	"github.com/opentibiabr/login-server/src/utils"
	"github.com/stretchr/testify/assert"
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
	actualValue := utils.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, value, actualValue)

	os.Setenv(TestKey, defaultValue)
	actualValue = utils.GetEnvStr(TestKey, defaultValue)
	assert.Equal(t, defaultValue, actualValue)
	os.Unsetenv(TestKey)
}

func TestGetEnvStrNotSetGetsDefault(t *testing.T) {

	defaultValue := "random_default"

	assert.Equal(t, defaultValue, utils.GetEnvStr(TestKey, defaultValue))
	os.Unsetenv(TestKey)
}

func TestGetEnvInt(t *testing.T) {
	defaultValue := 737
	value := 100

	os.Setenv(TestKey, fmt.Sprintf("%d", value))
	assert.Equal(t, value, utils.GetEnvInt(TestKey, defaultValue))

	os.Setenv(TestKey, fmt.Sprintf("%d", defaultValue))
	assert.Equal(t, defaultValue, utils.GetEnvInt(TestKey, defaultValue))
	os.Unsetenv(TestKey)
}

func TestGetEnvIntNotSetGetsDefault(t *testing.T) {
	defaultValue := 333

	value := utils.GetEnvInt(TestKey, defaultValue)
	assert.Equal(t, defaultValue, value)
	os.Unsetenv(TestKey)
}

func TestLoad(t *testing.T) {
	defaultString := "default_string"
	defaultNumber := "8080"

	os.Setenv(utils.EnvLoginPortKey, defaultNumber)
	os.Setenv(utils.EnvServerIpKey, defaultString)
	os.Setenv(utils.EnvServerNameKey, defaultString)
	os.Setenv(utils.EnvServerPortKey, defaultNumber)
	os.Setenv(utils.EnvServerLocationKey, defaultString)
	os.Setenv(utils.EnvDBHostKey, defaultString)
	os.Setenv(utils.EnvDBPortKey, defaultNumber)
	os.Setenv(utils.EnvDBNameKey, defaultString)
	os.Setenv(utils.EnvDBUserKey, defaultString)
	os.Setenv(utils.EnvDBPassKey, defaultString)

	expectedConfigs := utils.Configs{
		LoginPort: 8080,
		GameServerConfigs: utils.GameServerConfigs{
			IP:       defaultString,
			Name:     defaultString,
			Port:     8080,
			Location: defaultString,
		},
		DBConfigs: utils.DBConfigs{
			Host: defaultString,
			Name: defaultString,
			Port: 8080,
			User: defaultString,
			Pass: defaultString,
		},
	}

	c := utils.Configs{}
	c.Load()

	assert.Equal(t, expectedConfigs, c)
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

	assert.Equal(t, defaultVoc, utils.GetServerVocations())
}

func TestGetServerVocationsFromEnv(t *testing.T) {
	newVoc := []string{
		"artista",
		"professor",
		"engenheiro",
	}

	err := os.Setenv(utils.EnvVocations, strings.Join(newVoc, ","))
	if err != nil {
		log.Print("Error trying to get vocations from env vars.")
	}

	assert.Equal(t, newVoc, utils.GetServerVocations())
	os.Unsetenv(utils.EnvVocations)
}
