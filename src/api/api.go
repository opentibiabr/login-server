package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/server"
	"google.golang.org/grpc"
	"net/http"
	"sync"
)

type Api struct {
	Router         *gin.Engine
	DB             *sql.DB
	GrpcConnection *grpc.ClientConn
	server.ServerInterface
}

func Initialize(gConfigs configs.GlobalConfigs) *Api {
	var _api Api
	var err error

	_api.DB = database.PullConnection(gConfigs)

	ipLimiter := &limiter.IPRateLimiter{
		Visitors: make(map[string]*limiter.Visitor),
		Mu:       &sync.RWMutex{},
	}

	ipLimiter.Init()

	gin.SetMode(gin.ReleaseMode)

	_api.Router = gin.New()
	_api.Router.Use(logger.LogRequest())
	_api.Router.Use(gin.Recovery())
	_api.Router.Use(ipLimiter.Limit())

	_api.initializeRoutes()

	/* Generate HTTP/GRPC reverse proxy */

	_api.GrpcConnection, err = grpc.Dial(gConfigs.LoginServerConfigs.Tcp.Format(), grpc.WithInsecure())
	if err != nil {
		logger.Error(errors.New("Couldn't start GRPC reverse proxy."))
	}

	return &_api
}

func (_api *Api) Run(gConfigs configs.GlobalConfigs) error {
	err := http.ListenAndServe(gConfigs.LoginServerConfigs.Http.Format(), _api.Router)

	/* Make sure we free the reverse proxy connection */
	if _api.GrpcConnection != nil {
		closeErr := _api.GrpcConnection.Close()
		if closeErr != nil {
			logger.Error(closeErr)
		}
	}

	return err
}

func (_api *Api) GetName() string {
	return "api"
}

func (_api *Api) initializeRoutes() {
	_api.Router.POST("/login", _api.login)
	_api.Router.POST("/login.php", _api.login)
}
