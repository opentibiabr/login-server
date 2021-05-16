package main

import (
	"errors"
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/logger"
	"sync"
)

func main() {
	logger.Init(configs.GetLogLevel())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Loading configurations...")

	var wg sync.WaitGroup
	wg.Add(2)

	gConfigs := configs.GetGlobalConfigs()
	go startHttpServer(&wg, gConfigs)
	go startGrpcServer(&wg, gConfigs)

	// wait until WaitGroup is done
	wg.Wait()
	logger.Info("Good bye...")
}

func startGrpcServer(wg *sync.WaitGroup, globalConfigs configs.GlobalConfigs) {
	logger.Info("Grpc is also running bois...")
	grpc.Run(globalConfigs.LoginServerConfigs.Tcp.Format())
	wg.Done()
	logger.Error(errors.New("Grpc is gone bois..."))
}

func startHttpServer(wg *sync.WaitGroup, globalConfigs configs.GlobalConfigs) {
	httpServer := api.Api{}
	httpServer.Initialize(globalConfigs)
	httpServer.Configs.Display()
	httpServer.Run(httpServer.Configs.LoginServerConfigs.Http.Format())
	wg.Done()
	logger.Error(errors.New("Http is gone..."))
}
