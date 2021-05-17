package models

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

type World struct {
	AntiCheatProtection        bool   `json:"anticheatprotection"`
	CurrentTournamentPhase     int    `json:"currenttournamentphase"`
	ExternalAddress            string `json:"externaladdress"`
	ExternalAddressProtected   string `json:"externaladdressprotected"`
	ExternalAddressUnprotected string `json:"externaladdressunprotected"`
	ExternalPort               int    `json:"externalport"`
	ExternalPortProtected      int    `json:"externalportprotected"`
	ExternalPortUnprotected    int    `json:"externalportunprotected"`
	ID                         int    `json:"id"`
	IsTournamentWorld          bool   `json:"istournamentworld"`
	Location                   string `json:"location"`
	Name                       string `json:"name"`
	PreviewState               int    `json:"previewstate"`
	PvpType                    int    `json:"pvptype"`
	RestrictedStore            bool   `json:"restrictedstore"`
}

func LoadWorldsFromMessage(worldsMsg []*login_proto_messages.World) []World {
	var worlds []World
	for _, worldMsg := range worldsMsg {
		worlds = append(
			worlds,
			World{
				ExternalAddress:            worldMsg.ExternalAddress,
				ExternalAddressProtected:   worldMsg.ExternalAddressProtected,
				ExternalAddressUnprotected: worldMsg.ExternalAddressUnprotected,
				ExternalPort:               int(worldMsg.ExternalPort),
				ExternalPortProtected:      int(worldMsg.ExternalPortProtected),
				ExternalPortUnprotected:    int(worldMsg.ExternalPortUnprotected),
				Location:                   worldMsg.Location,
				Name:                       worldMsg.Name,
			},
		)
	}

	return worlds
}

func BuildWorldsMessage() []*login_proto_messages.World {
	return []*login_proto_messages.World{buildWorldMessage(configs.GetGameServerConfigs())}
}

func buildWorldMessage(gameConfigs configs.GameServerConfigs) *login_proto_messages.World {
	return &login_proto_messages.World{
		ExternalAddress:            gameConfigs.IP,
		ExternalAddressProtected:   gameConfigs.IP,
		ExternalAddressUnprotected: gameConfigs.IP,
		ExternalPort:               uint32(gameConfigs.Port),
		ExternalPortProtected:      uint32(gameConfigs.Port),
		ExternalPortUnprotected:    uint32(gameConfigs.Port),
		Location:                   gameConfigs.Location,
		Name:                       gameConfigs.Name,
	}
}
