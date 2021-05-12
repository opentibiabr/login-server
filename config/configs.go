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

const EnvTypeKey = "ENV_TYPE"

const EnvLoginPortKey = "LOGIN_PORT"

const EnvServerIpKey = "SERVER_IP"
const EnvServerNameKey = "SERVER_NAME"
const EnvServerPortKey = "SERVER_PORT"
const EnvServerLocationKey = "SERVER_LOCATION"

const EnvDBHostKey = "DB_HOSTNAME"
const EnvDBPortKey = "DB_PORT"
const EnvDBNameKey = "DB_DATABASE"
const EnvDBUserKey = "DB_USERNAME"
const EnvDBPassKey = "DB_PASSWORD"

func (c *Configs) Load() {
	envType := c.GetEnvStr(EnvTypeKey, "dev")

	if envType == "dev" {
		godotenv.Load(".env")
	}

	c.LoginPort = c.GetEnvInt(EnvLoginPortKey, 80)

	c.GameServerConfigs.IP = c.GetEnvStr(EnvServerIpKey, "127.0.0.1")
	c.GameServerConfigs.Name = c.GetEnvStr(EnvServerNameKey, "Canary")
	c.GameServerConfigs.Port = c.GetEnvInt(EnvServerPortKey, 7172)
	c.GameServerConfigs.Location = c.GetEnvStr(EnvServerLocationKey, "BRA")

	c.DBConfigs.Host = c.GetEnvStr(EnvDBHostKey, "127.0.0.1")
	c.DBConfigs.Port = c.GetEnvInt(EnvDBPortKey, 3306)
	c.DBConfigs.Name = c.GetEnvStr(EnvDBNameKey, "canary")
	c.DBConfigs.User = c.GetEnvStr(EnvDBUserKey, "canary")
	c.DBConfigs.Pass = c.GetEnvStr(EnvDBPassKey, "canary")
}

func (c *Configs) GetEnvStr(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (c *Configs) GetEnvInt(key string, defaultValue ...int) int {
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
