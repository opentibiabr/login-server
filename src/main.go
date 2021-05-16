package main

import (
	"errors"
	"github.com/opentibiabr/login-server/src/api_http"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
	//"github.com/opentibiabr/login-server/src/tcp"
	"sync"
)

func main() {
	logger.Init(configs.GetLogLevel())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Loading configurations...")

	var wg sync.WaitGroup
	wg.Add(2)

	go startHttpServer(&wg)
	go startTcpServer(&wg)

	// wait until WaitGroup is done
	wg.Wait()
	logger.Info("Good bye...")
}

func startTcpServer(wg *sync.WaitGroup) {
	logger.Info("Tcp is also running bois...")
	//tcp.Run()
	wg.Done()
	logger.Error(errors.New("TCP is gone bois..."))
}

func startHttpServer(wg *sync.WaitGroup) {
	httpServer := api_http.HttpApi{}
	httpServer.Initialize()
	httpServer.Configs.Display()
	httpServer.Run(httpServer.Configs.LoginServerConfigs.Http.Format())
	wg.Done()
	logger.Error(errors.New("Http is gone..."))
}
