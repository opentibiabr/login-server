package utils

import (
	"log"
)

const LogLevelVerbose = "verbose"
const LogLevelSilent = "silent"

func Log(format string, v ...interface{}) {
	if getLogLevel() == LogLevelSilent {
		return
	}

	log.Printf(format, v...)
}

func getLogLevel() string {
	return GetEnvStr(EnvLogLevel, LogLevelVerbose)
}
