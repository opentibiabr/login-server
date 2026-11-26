package configs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opentibiabr/login-server/src/logger"
)

const EnvServerPvpTypeKey = "SERVER_PVP_TYPE"

// Codes of the world PvP type, as the client reads `pvptype` in the world list.
const (
	PvpTypeOpen          uint32 = 0
	PvpTypeOptional      uint32 = 1
	PvpTypeHardcore      uint32 = 2
	PvpTypeRetroOpen     uint32 = 3
	PvpTypeRetroHardcore uint32 = 4
)

// PvpTypeFromWorldType maps a `worldType` value of the game server configuration to the client code. "pvp" is the legacy
// name of "retro-pvp" in Canary, which announces both as retro PvP.
func PvpTypeFromWorldType(worldType string) (uint32, bool) {
	switch strings.ToLower(strings.TrimSpace(worldType)) {
	case "retro-pvp", "pvp":
		return PvpTypeRetroOpen, true
	case "retro-pvp-enforced":
		return PvpTypeRetroHardcore, true
	case "no-pvp":
		return PvpTypeOptional, true
	case "pvp-enforced":
		return PvpTypeHardcore, true
	case "expert-pvp":
		return PvpTypeOpen, true
	}
	return PvpTypeOpen, false
}

// GetServerPvpType returns the PvP type code of the world. SERVER_PVP_TYPE (a `worldType` name or a code from 0 to 4) wins;
// without it, `worldType` is read from the config.lua of SERVER_PATH. Without either, or with an unknown value, the world is
// reported as open PvP, which is what the response carried before.
func GetServerPvpType() uint32 {
	value := strings.TrimSpace(GetEnvStr(EnvServerPvpTypeKey, ""))
	if value == "" {
		value = worldTypeFromServerConfig()
	}
	if value == "" {
		return PvpTypeOpen
	}

	if code, err := strconv.ParseUint(value, 10, 32); err == nil && uint32(code) <= PvpTypeRetroHardcore {
		return uint32(code)
	}
	if code, ok := PvpTypeFromWorldType(value); ok {
		return code
	}

	logger.Warn(fmt.Sprintf("unknown world PvP type %q, reporting open PvP", value))
	return PvpTypeOpen
}

// worldTypeFromServerConfig reads `worldType` from the game server configuration of SERVER_PATH, or "" when SERVER_PATH is
// unset or the configuration cannot be read (the login itself reports configuration problems elsewhere).
func worldTypeFromServerConfig() string {
	serverPath := strings.TrimSpace(GetEnvStr(EnvServerPathKey, ""))
	if serverPath == "" {
		return ""
	}

	configPath, err := findServerConfigPath(serverPath)
	if err != nil {
		return ""
	}

	manager, err := NewLuaConfigManager(configPath)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(manager.GetString("worldType"))
}
