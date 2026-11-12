package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	grpc_login_server "github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/server"
)

var initDelay = 200

func main() {
	err := configs.Init()
	logger.Init(configs.GetLogLevel(), configs.GetLogFile())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Configurations loaded")
	if err != nil {
		logger.Debug("Failed to load '.env' in dev environment, going with default.")
	}

	gConfigs := configs.GetGlobalConfigs()

	err = runServers(gConfigs,
		func(config configs.GlobalConfigs) server.ServerInterface { return grpc_login_server.Initialize(config) },
		func(config configs.GlobalConfigs) server.ServerInterface { return api.Initialize(config) },
		time.Duration(initDelay)*time.Millisecond,
	)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("Good bye...")
}

func runServers(
	gConfigs configs.GlobalConfigs,
	initializeGrpc, initializeAPI func(configs.GlobalConfigs) server.ServerInterface,
	startupDelay time.Duration,
) error {
	// Validate before either initializer can create a database connection or proxy.
	if _, err := configs.NewTestGodLoginPolicy(gConfigs, configs.GetEnvStr(configs.EnvServerPathKey)); err != nil {
		return err
	}
	initializers := []func(configs.GlobalConfigs) server.ServerInterface{initializeGrpc, initializeAPI}
	var wg sync.WaitGroup
	wg.Add(len(initializers))
	for _, initialize := range initializers {
		go startServer(&wg, gConfigs, initialize(gConfigs))
	}

	time.Sleep(startupDelay)
	gConfigs.Display()

	// wait until WaitGroup is done
	wg.Wait()
	return nil
}

func startServer(
	wg *sync.WaitGroup,
	gConfigs configs.GlobalConfigs,
	server server.ServerInterface,
) {
	logger.Info(fmt.Sprintf("Starting %s server...", server.GetName()))
	logger.Error(server.Run(gConfigs))
	wg.Done()
	logger.Warn(fmt.Sprintf("Server %s is gone...", server.GetName()))
}
