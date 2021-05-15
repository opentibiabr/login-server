package main

import (
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
)

func main() {
	logger.Init(configs.GetLogLevel())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Loading configurations...")

	app := api.Api{}
	app.Initialize()

	app.Configs.Display()
	app.Run(app.Configs.LoginServerConfigs.Http.Format())
}
