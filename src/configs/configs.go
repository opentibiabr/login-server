package configs

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/opentibiabr/login-server/src/logger"
)

type Config interface {
	Format() string
}

type GlobalConfigs struct {
	DBConfigs          DBConfigs
	GameServerConfigs  GameServerConfigs
	LoginServerConfigs LoginServerConfigs
}

// LuaConfigManager is a structure that manages configurations loaded from a Lua file.
// It stores the configurations as key-value pairs, where both keys and values are strings.
type LuaConfigManager struct {
	configs map[string]string
}

// Init only works for variables that are not yet defined. /*
func Init() error {
	return godotenv.Load(".env")
}

func (c *GlobalConfigs) Display() {
	logger.Info(c.DBConfigs.format())
	logger.Info(c.GameServerConfigs.Format())
	logger.Info(c.LoginServerConfigs.Format())
}

func GetGlobalConfigs() GlobalConfigs {
	return GlobalConfigs{
		DBConfigs:          GetDBConfigs(),
		GameServerConfigs:  GetGameServerConfigs(),
		LoginServerConfigs: GetLoginServerConfigs(),
	}
}

func GetEnvInt(key string, defaultValue ...int) int {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		logger.Error(err)
		return 0
	}

	return intValue
}

func GetEnvStr(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (manager *LuaConfigManager) loadConfigs(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	re := regexp.MustCompile(`^(\w+)\s*=\s*(["']?.*["']?)$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := cleanLuaLine(scanner.Text())
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			manager.configs[matches[1]] = matches[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func cleanLuaLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "--") {
		return ""
	}

	var result strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	for i := 0; i < len(line); i++ {
		escapedQuote := isEscapedLuaQuote(line, i)
		if line[i] == '\'' && !inDoubleQuotes && !escapedQuote {
			inSingleQuotes = !inSingleQuotes
		} else if line[i] == '"' && !inSingleQuotes && !escapedQuote {
			inDoubleQuotes = !inDoubleQuotes
		}
		if i+1 < len(line) && !inSingleQuotes && !inDoubleQuotes && line[i] == '-' && line[i+1] == '-' {
			break
		}
		result.WriteByte(line[i])
	}

	return strings.TrimSpace(result.String())
}

func isEscapedLuaQuote(line string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}

	return backslashes%2 == 1
}

// NewLuaConfigManager creates a new instance of LuaConfigManager by loading configurations from the specified Lua file.
// It returns the LuaConfigManager instance and any error encountered during the file loading process.
func NewLuaConfigManager(filePath string) (*LuaConfigManager, error) {
	manager := &LuaConfigManager{
		configs: make(map[string]string),
	}
	err := manager.loadConfigs(filePath)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

// GetString retrieves the string value for a given key from the configuration.
// If the key does not exist or the value is not a string, it returns an empty string.
func (manager *LuaConfigManager) GetString(key string) string {
	if manager == nil {
		return ""
	}

	value, exists := manager.configs[key]
	if !exists {
		return ""
	}

	if exists && len(value) > 1 && (value[0] == '"' || value[0] == '\'') && value[0] == value[len(value)-1] {
		value = value[1 : len(value)-1]
	}
	return value
}

// GetInt retrieves the integer value for a given key from the configuration.
// It returns 0 if the key does not exist, the value is not an integer, or an error occurs during conversion.
func (manager *LuaConfigManager) GetInt(key string) int {
	if manager == nil {
		return 0
	}

	valueStr := manager.GetString(key)
	if valueStr == "" {
		return 0
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0
	}
	return value
}

// GetBool retrieves the boolean value for a given key from the configuration.
// It returns false if the key does not exist, the value is not a boolean, or an error occurs during conversion.
func (manager *LuaConfigManager) GetBool(key string) bool {
	if manager == nil {
		return false
	}

	valueStr := manager.GetString(key)
	if valueStr == "" {
		return false
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false
	}
	return value
}
