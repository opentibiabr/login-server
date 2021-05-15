package login

import "github.com/opentibiabr/login-server/src/configs"

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

func LoadWorld() World {
	gameConfigs := configs.GetGameServerConfigs()
	return World{
		ExternalAddress:            gameConfigs.IP,
		ExternalAddressProtected:   gameConfigs.IP,
		ExternalAddressUnprotected: gameConfigs.IP,
		ExternalPort:               gameConfigs.Port,
		ExternalPortProtected:      gameConfigs.Port,
		ExternalPortUnprotected:    gameConfigs.Port,
		Location:                   gameConfigs.Location,
		Name:                       gameConfigs.Name,
	}
}
