package configs

import (
	"github.com/joho/godotenv"
	"github.com/opentibiabr/login-server/src/logger"
	"os"
	"strconv"
)

type Config interface {
	Format() string
}

type GlobalConfigs struct {
	DBConfigs          DBConfigs
	GameServerConfigs  GameServerConfigs
	LoginServerConfigs LoginServerConfigs
}

// Init only works for variables that are not yet defined. /*
func Init() error {
	return godotenv.Load(".env")
}

func (c *GlobalConfigs) Display() {
	logger.Info(c.DBConfigs.Format())
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
