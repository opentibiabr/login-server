package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Configs struct {
	LoginPort int
	DBConfigs
	GameServerConfigs
}

type DBConfigs struct {
	DBHost string
	DBPort int
	DBName string
	DBUser string
	DBPass string
}

type GameServerConfigs struct {
	GameServerPort int
	GameServerName string
	GameServerIP   string
}

func (c *Configs) Load() {
	envType := c.getEnvStr("ENV_TYPE", "test")

	if envType == "test" {
		if err := godotenv.Load(".env"); err != nil {
			log.Print("Error loading .env file for test env, going with default values")
		}
	}

	c.LoginPort = c.getEnvInt("APP_PORT", 80)

	c.GameServerIP = c.getEnvStr("SERVER_IP", "127.0.0.1")
	c.GameServerName = c.getEnvStr("SERVER_NAME", "Canary")
	c.GameServerPort = c.getEnvInt("SERVER_PORT", 7172)

	c.DBHost = c.getEnvStr("DB_HOSTNAME", "127.0.0.1")
	c.DBPort = c.getEnvInt("DB_PORT", 3306)
	c.DBName = c.getEnvStr("DB_DATABASE", "canary")
	c.DBPass = c.getEnvStr("DB_PASSWORD", "canary")
	c.DBUser = c.getEnvStr("DB_USERNAME", "canary")
}

func (c *Configs) getEnvStr(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (c *Configs) getEnvInt(key string, defaultValue ...int) int {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Invalid integer value for %s", key)
		return 0
	}

	return intValue
}
