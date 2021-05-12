package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Configs struct {
	LoginPort         int
	DBConfigs         DBConfigs
	GameServerConfigs GameServerConfigs
}

type DBConfigs struct {
	Host string
	Port int
	Name string
	User string
	Pass string
}

type GameServerConfigs struct {
	Port     int
	Name     string
	IP       string
	Location string
}

func (c *Configs) Load() {
	envType := c.getEnvStr("ENV_TYPE", "test")

	if envType == "test" {
		if err := godotenv.Load(".env"); err != nil {
			log.Print("Error loading .env file for test env, going with default values")
		}
	}

	c.LoginPort = c.getEnvInt("LOGIN_PORT", 80)

	c.GameServerConfigs.IP = c.getEnvStr("SERVER_IP", "127.0.0.1")
	c.GameServerConfigs.Name = c.getEnvStr("SERVER_NAME", "Canary")
	c.GameServerConfigs.Port = c.getEnvInt("SERVER_PORT", 7172)
	c.GameServerConfigs.Location = c.getEnvStr("SERVER_LOCATION", "BRA")

	c.DBConfigs.Host = c.getEnvStr("DB_HOSTNAME", "127.0.0.1")
	c.DBConfigs.Port = c.getEnvInt("DB_PORT", 3306)
	c.DBConfigs.Name = c.getEnvStr("DB_DATABASE", "canary")
	c.DBConfigs.Pass = c.getEnvStr("DB_PASSWORD", "canary")
	c.DBConfigs.User = c.getEnvStr("DB_USERNAME", "canary")
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

func (c *Configs) Print() {
	log.Printf(
		"Database: %s:%d/%s",
		c.DBConfigs.Host,
		c.DBConfigs.Port,
		c.DBConfigs.Name,
	)
	log.Printf(
		"Connected with %s server (%s:%d)",
		c.GameServerConfigs.Name,
		c.GameServerConfigs.IP,
		c.GameServerConfigs.Port,
	)
}
