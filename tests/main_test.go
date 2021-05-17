package tests

import (
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

var a api.Api

func TestMain(m *testing.M) {
	/* Disable application logs */
	logger.Init(logrus.PanicLevel)

	a = *api.Initialize(configs.GetGlobalConfigs())
	code := m.Run()
	os.Exit(code)
}
