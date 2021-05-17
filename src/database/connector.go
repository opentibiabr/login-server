package database

import (
	"database/sql"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
)

const DefaultMaxDbOpenConns = 50

func PullConnection(gConfigs configs.GlobalConfigs) *sql.DB {
	DB, err := sql.Open("mysql", gConfigs.DBConfigs.GetConnectionString())
	if err != nil {
		logger.Fatal(err)
	}

	DB.SetMaxOpenConns(DefaultMaxDbOpenConns)

	return DB
}
