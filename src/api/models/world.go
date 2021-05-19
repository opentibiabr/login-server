package models

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

type World struct {
	ID                         uint32 `json:"id" proto:"Id"`
	Name                       string `json:"name"`
	ExternalAddress            string `json:"externaladdress"`
	ExternalAddressProtected   string `json:"externaladdressprotected"`
	ExternalAddressUnprotected string `json:"externaladdressunprotected"`
	ExternalPort               uint32 `json:"externalport"`
	ExternalPortProtected      uint32 `json:"externalportprotected"`
	ExternalPortUnprotected    uint32 `json:"externalportunprotected"`
	Location                   string `json:"location"`
	AntiCheatProtection        bool   `json:"anticheatprotection"`
	CurrentTournamentPhase     uint32 `json:"currenttournamentphase"`
	IsTournamentWorld          bool   `json:"istournamentworld"`
	PreviewState               uint32 `json:"previewstate"`
	PvpType                    uint32 `json:"pvptype"`
	RestrictedStore            bool   `json:"restrictedstore"`
}

func LoadWorldsFromMessage(worldsMsg []*login_proto_messages.World) []World {
	var worlds []World

	for _, worldMsg := range worldsMsg {
		worlds = append(worlds, *FromProtoConvertor(worldMsg, &World{}).(*World))
	}

	return worlds
}

func BuildWorldsMessage(gameConfigs configs.GameServerConfigs) []*login_proto_messages.World {
	return []*login_proto_messages.World{buildWorldMessage(gameConfigs, 0)}
}

func buildWorldMessage(gameConfigs configs.GameServerConfigs, worldId uint32) *login_proto_messages.World {
	return &login_proto_messages.World{
		Id:                         worldId,
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
