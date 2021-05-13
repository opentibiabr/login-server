package tests

import (
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/config"
	"github.com/opentibiabr/login-server/tests/testlib"
	"os"
	"testing"
)

func TestLoadWorld(t *testing.T) {
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

	c := config.Configs{}
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

	a.Equals(expectedWorld, world)
}
