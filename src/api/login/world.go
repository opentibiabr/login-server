package login

import "github.com/opentibiabr/login-server/src/config"

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

func LoadWorld(configs config.Configs) World {
	return World{
		ExternalAddress:            configs.GameServerConfigs.IP,
		ExternalAddressProtected:   configs.GameServerConfigs.IP,
		ExternalAddressUnprotected: configs.GameServerConfigs.IP,
		ExternalPort:               configs.GameServerConfigs.Port,
		ExternalPortProtected:      configs.GameServerConfigs.Port,
		ExternalPortUnprotected:    configs.GameServerConfigs.Port,
		Location:                   configs.GameServerConfigs.Location,
		Name:                       configs.GameServerConfigs.Name,
	}
}
