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
const EnvRunSilent = "ENV_RUN_SILENT"

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
	if IsDevEnvironment() {
		err := godotenv.Load(".env")
		if err != nil && !RunSilent() {
			log.Print("Failed to load '.env' in dev environment, going with default.")
		}
	}

	c.LoginPort = GetEnvInt(EnvLoginPortKey, 80)

	c.GameServerConfigs.IP = GetEnvStr(EnvServerIpKey, "127.0.0.1")
	c.GameServerConfigs.Name = GetEnvStr(EnvServerNameKey, "Canary")
	c.GameServerConfigs.Port = GetEnvInt(EnvServerPortKey, 7172)
	c.GameServerConfigs.Location = GetEnvStr(EnvServerLocationKey, "BRA")

	c.DBConfigs.Host = GetEnvStr(EnvDBHostKey, "127.0.0.1")
	c.DBConfigs.Port = GetEnvInt(EnvDBPortKey, 3306)
	c.DBConfigs.Name = GetEnvStr(EnvDBNameKey, "canary")
	c.DBConfigs.User = GetEnvStr(EnvDBUserKey, "canary")
	c.DBConfigs.Pass = GetEnvStr(EnvDBPassKey, "canary")
}

func GetEnvStr(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func GetEnvInt(key string, defaultValue ...int) int {
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
		"Connected with %s server %s:%d - %s",
		c.GameServerConfigs.Name,
		c.GameServerConfigs.IP,
		c.GameServerConfigs.Port,
		c.GameServerConfigs.Location,
	)
}

func RunSilent() bool {
	return len(GetEnvStr(EnvRunSilent, "")) == 0
}

func IsDevEnvironment() bool {
	return GetEnvStr(EnvTypeKey, "dev") == "dev"
}