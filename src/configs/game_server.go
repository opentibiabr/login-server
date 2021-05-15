package configs

import (
	"fmt"
	"strings"
)

const EnvServerIpKey = "SERVER_IP"
const EnvServerLocationKey = "SERVER_LOCATION"
const EnvServerNameKey = "SERVER_NAME"
const EnvServerPortKey = "SERVER_PORT"

type GameServerConfigs struct {
	Port     int
	Name     string
	IP       string
	Location string
	Config
}

func (gameServerConfigs *GameServerConfigs) Format() string {
	return fmt.Sprintf(
		"Connected with %s server %s:%d - %s",
		gameServerConfigs.Name,
		gameServerConfigs.IP,
		gameServerConfigs.Port,
		gameServerConfigs.Location,
	)
}
func GetGameServerConfigs() GameServerConfigs {
	return GameServerConfigs{
		IP:       GetEnvStr(EnvServerIpKey, "127.0.0.1"),
		Name:     GetEnvStr(EnvServerNameKey, "Canary"),
		Port:     GetEnvInt(EnvServerPortKey, 7172),
		Location: GetEnvStr(EnvServerLocationKey, "BRA"),
	}
}

const EnvVocations = "VOCATIONS"

func GetServerVocations() []string {
	vocationsStr := GetEnvStr(EnvVocations, "")
	vocations := strings.Split(vocationsStr, ",")

	if len(vocationsStr) == 0 || len(vocations) == 0 {
		return []string{
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
	}

	return vocations
}
