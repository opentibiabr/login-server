package api

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"login-server/src/api/limiter"
	"login-server/src/config"
	"net/http"
	"sync"
)

type Api struct {
	Router  *mux.Router
	DB      *sql.DB
	Configs config.Configs
}

func (_api *Api) Initialize() {
	_api.Configs.Load()

	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		_api.Configs.DBConfigs.User,
		_api.Configs.DBConfigs.Pass,
		_api.Configs.DBConfigs.Host,
		_api.Configs.DBConfigs.Port,
		_api.Configs.DBConfigs.Name,
	)

	var err error
	_api.DB, err = sql.Open("mysql", connectionString)
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
