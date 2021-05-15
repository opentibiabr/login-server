package api

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/utils"
	"log"
	"net/http"
	"sync"
)

type Api struct {
	Router  *mux.Router
	DB      *sql.DB
	Configs configs.GlobalConfigs
}

func (_api *Api) Initialize() {
	err := configs.Init()
	if err != nil {
		utils.Log("Failed to load '.env' in dev environment, going with default.")
	}

	_api.Configs = configs.GetGlobalConfigs()

	_api.DB, err = sql.Open("mysql", _api.Configs.DBConfigs.GetConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	ipLimiter := &limiter.IPRateLimiter{
		Visitors: make(map[string]*limiter.Visitor),
		Mu:       &sync.RWMutex{},
	}

	ipLimiter.Init()

	_api.Router = mux.NewRouter()
	_api.initializeRoutes()
	_api.Router.Use(ipLimiter.Limit)
}

func (_api *Api) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, _api.Router))
}

func (_api *Api) initializeRoutes() {
	_api.Router.HandleFunc("/login", _api.login).Methods("GET", "POST", "PUT")
	_api.Router.HandleFunc("/login.php", _api.login).Methods("GET", "POST", "PUT")
}
