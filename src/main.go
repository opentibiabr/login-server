package main

import (
	"fmt"
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/definitions"
	"github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/logger"
	"sync"
	"time"
)

type LoginServer interface {
	Run(globalConfigs configs.GlobalConfigs) error
	GetName() string
}

func main() {
	logger.Init(configs.GetLogLevel())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Loading configurations...")

	var wg sync.WaitGroup
	wg.Add(2)

	gConfigs := configs.GetGlobalConfigs()

	err := configs.Init()
	if err != nil {
		logger.Warn("Failed to load '.env' in dev environment, going with default.")
	}

	go startServer(&wg, gConfigs, grpc_login_server.Initialize(gConfigs))
	go startServer(&wg, gConfigs, api.Initialize(gConfigs))

	time.Sleep(200 * time.Millisecond)
	gConfigs.Display()

	// wait until WaitGroup is done
	wg.Wait()
	logger.Info("Good bye...")
}

func startServer(
	wg *sync.WaitGroup,
	gConfigs configs.GlobalConfigs,
	server definitions.ServerInterface,
) {
	logger.Info(fmt.Sprintf("Starting %s server...", server.GetName()))
	logger.Error(server.Run(gConfigs))
	wg.Done()
	logger.Warn(fmt.Sprintf("Server %s is gone...", server.GetName()))
}
