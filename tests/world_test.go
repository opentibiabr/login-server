package tests

import (
	"github.com/opentibiabr/login-server/src/api_http/login"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadWorld(t *testing.T) {
	SetEnvKeys(
		[]string{
			configs.EnvServerIpKey,
			configs.EnvServerNameKey,
			configs.EnvServerLocationKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
		},
		[]string{
			configs.EnvLoginHttpPortKey,
			configs.EnvServerPortKey,
			configs.EnvDBPortKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
		},
	)

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
	world := login.LoadWorld()

	assert.Equal(t, expectedWorld, world)

	UnsetEnvKeys(
		[]string{
			configs.EnvServerIpKey,
			configs.EnvServerNameKey,
			configs.EnvServerLocationKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
			configs.EnvLoginHttpPortKey,
			configs.EnvServerPortKey,
			configs.EnvDBPortKey,
			configs.EnvDBHostKey,
			configs.EnvDBNameKey,
			configs.EnvDBUserKey,
			configs.EnvDBPassKey,
		},
	)
}
