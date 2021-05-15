package tests

import (
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	"log"
	"os"
	"testing"
)

var a api.Api

const defaultString = "default_string"
const defaultNumberStr = "8080"
const defaultNumber = 8080

func TestMain(m *testing.M) {
	err := os.Setenv(configs.EnvLogLevel, configs.LogLevelSilent)
	if err != nil {
		log.Print("Can't set silent true")
	}
	a = api.Api{}
	a.Initialize()
	code := m.Run()
	os.Exit(code)
}

func SetEnvKeys(strings []string, integers []string) {
	for _, key := range strings {
		err := os.Setenv(key, defaultString)
		if err != nil {
			log.Printf("Cannot set key %s", key)
		}
	}
	for _, key := range integers {
		err := os.Setenv(key, defaultNumberStr)
		if err != nil {
			log.Printf("Cannot set key %s", key)
		}
	}
}

func UnsetEnvKeys(keys []string) {
	for _, key := range keys {
		err := os.Unsetenv(key)
		if err != nil {
			log.Printf("Cannot unset key %s", key)
		}
	}
}
