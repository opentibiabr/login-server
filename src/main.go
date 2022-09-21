package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	grpcsrv "github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/server"
)

const (
	quitChanBuffer = 1
	sleepTime      = 200 * time.Millisecond
)

func main() {
	logger.Init(configs.GetLogLevel())
	logger.Info("Welcome to OTBR Login Server")
	logger.Info("Loading configurations...")

	err := configs.Init()
	if err != nil {
		logger.Debug("Failed to load '.env' in dev environment, going with default.")
	}

	cfg := configs.GetGlobalConfigs()
	errC := make(chan error)
	quit := make(chan os.Signal, quitChanBuffer)

	// listen to server errors
	go func() {
		if err := <-errC; err != nil {
			quit <- os.Kill
		}
	}()

	// start the grpc server
	go func() {
		if err := startServer(cfg, grpcsrv.Initialize(cfg)); err != nil {
			logger.Error(err)
			errC <- err
		}
	}()

	// start the api server
	go func() {
		if err := startServer(cfg, api.Initialize(cfg)); err != nil {
			logger.Error(err)
			errC <- err
		}
	}()

	time.Sleep(sleepTime)
	cfg.Display()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	logger.Info("Login server has been shutdown, goodbye...")
}

func startServer(cfg configs.GlobalConfigs, srv server.ServerInterface) error {
	logger.Info("Starting " + srv.GetName() + " server...")
	return srv.Run(cfg)
}
