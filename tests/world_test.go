package tests

import (
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadWorld(t *testing.T) {
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

	c := utils.Configs{}
	c.Load()

	expectedWorld := login.World{
		ExternalAddress:            defaultString,
		ExternalAddressProtected:   defaultString,
		ExternalAddressUnprotected: defaultString,
		ExternalPort:               8080,
		ExternalPortProtected:      8080,
		ExternalPortUnprotected:    8080,
		Location:                   defaultString,
		Name:                       defaultString,
	}
	world := login.LoadWorld(c)

	assert.Equal(t, expectedWorld, world)
}
