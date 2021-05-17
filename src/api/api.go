package api

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
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
	Router         *mux.Router
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

	_api.Router = mux.NewRouter()
	_api.initializeRoutes()
	_api.Router.Use(ipLimiter.Limit)

	/* Generate HTTP/GRPC reverse proxy */
	_api.GrpcConnection, err = grpc.Dial(":7171", grpc.WithInsecure())
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
	_api.Router.HandleFunc("/login", _api.login).Methods("POST")
	_api.Router.HandleFunc("/login.php", _api.login).Methods("POST")
}
