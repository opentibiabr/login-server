package tests

import (
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildWorldMessage(t *testing.T) {
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

	expectedWorld := &login_proto_messages.World{
		ExternalAddress:            defaultString,
		ExternalAddressProtected:   defaultString,
		ExternalAddressUnprotected: defaultString,
		ExternalPort:               8080,
		ExternalPortProtected:      8080,
		ExternalPortUnprotected:    8080,
		Location:                   defaultString,
		Name:                       defaultString,
	}
	world := models.buildWorldMessage(configs.GetGameServerConfigs())

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
