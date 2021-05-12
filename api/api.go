package api

import (
	"awesomeProject/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Api struct {
	Router  *mux.Router
	DB      *sql.DB
	Configs config.Configs
}

func (api *Api) Initialize() {
	api.Configs.Load()

	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		api.Configs.DBConfigs.User,
		api.Configs.DBConfigs.Pass,
		api.Configs.DBConfigs.Host,
		api.Configs.DBConfigs.Port,
		api.Configs.DBConfigs.Name,
	)

	var err error
	api.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	api.Router = mux.NewRouter()
	api.initializeRoutes()
}

func (api *Api) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, api.Router))
}

func (api *Api) initializeRoutes() {
	api.Router.HandleFunc("/login", api.login).Methods("GET", "POST", "PUT")
	api.Router.HandleFunc("/login.php", api.login).Methods("GET", "POST", "PUT")
}
