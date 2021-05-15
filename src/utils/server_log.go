package utils

import (
	"github.com/opentibiabr/login-server/src/configs"
	"log"
)

func Log(format string, v ...interface{}) {
	if configs.GetLogLevel() == configs.LogLevelSilent {
		return
	}

	log.Printf(format, v...)
}
