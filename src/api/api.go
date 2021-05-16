package api

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/definitions"
	"github.com/opentibiabr/login-server/src/logger"
	"net/http"
	"sync"
)

type Api struct {
	Router *mux.Router
	DB     *sql.DB
	definitions.ServerInterface
}

func Initialize(gConfigs configs.GlobalConfigs) *Api {
	var _api Api
	err := configs.Init()
	if err != nil {
		logger.Warn("Failed to load '.env' in dev environment, going with default.")
	}

	_api.DB, err = sql.Open("mysql", gConfigs.DBConfigs.GetConnectionString())
	if err != nil {
		logger.Fatal(err)
	}

	ipLimiter := &limiter.IPRateLimiter{
		Visitors: make(map[string]*limiter.Visitor),
		Mu:       &sync.RWMutex{},
	}

	ipLimiter.Init()

	_api.Router = mux.NewRouter()
	_api.initializeRoutes()
	_api.Router.Use(ipLimiter.Limit)

	return &_api
}

func (_api *Api) Run(gConfigs configs.GlobalConfigs) error {
	return http.ListenAndServe(gConfigs.LoginServerConfigs.Http.Format(), _api.Router)
}

func (_api *Api) GetName() string {
	return "api"
}

func (_api *Api) initializeRoutes() {
	_api.Router.HandleFunc("/login", _api.login).Methods("GET", "POST", "PUT")
	_api.Router.HandleFunc("/login.php", _api.login).Methods("GET", "POST", "PUT")
	_api.Router.HandleFunc("/login2", _api.login2).Methods("GET", "POST", "PUT")
}
