// Package database provides functionalities for interacting with the database of the system,
// including operations to fetch and update data about boosted creatures and bosses. This package
// encapsulates all SQL queries and data manipulations, making maintenance and future development
// easier.
package database

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/logger"
)

func loadBoostedCreatureData(db *sql.DB) (uint32, uint32, error) {
	var creatureRaceID uint32
	var bossRaceID uint32

	query := "SELECT raceid FROM boosted_creature LIMIT 1"
	err := db.QueryRow(query).Scan(&creatureRaceID)
	if err != nil {
		logger.Debug(fmt.Sprintf("Failed to load boosted creature data: %s", err.Error()))
		return 0, 0, err
	}

	query = "SELECT raceid FROM boosted_boss LIMIT 1"
	err = db.QueryRow(query).Scan(&bossRaceID)
	if err != nil {
		logger.Debug(fmt.Sprintf("Failed to load boosted boss data: %s", err.Error()))
		return 0, 0, err
	}

	return creatureRaceID, bossRaceID, nil
}

func checkAndUpdateBoostedCreatureData(db *sql.DB, creatureRaceID *uint32, bossRaceID *uint32) error {
	currentCreatureID, currentBossID, err := loadBoostedCreatureData(db)
	if err != nil {
		return err
	}

	if *creatureRaceID != currentCreatureID || *bossRaceID != currentBossID {
		logger.Info(fmt.Sprintf("Updating creatureid: '%d', bossid: '%d' due to change detected", currentCreatureID, currentBossID))
		*creatureRaceID = currentCreatureID
		*bossRaceID = currentBossID
	}
	return nil
}

// HandleBoostedCreature checks for updated boosted creature and boss race IDs in the database,
// updates the provided race IDs if they are different, and responds with the current boosted creature
// and boss race IDs. It takes a Gin context, a database connection, and pointers to the creature and boss
// race IDs as arguments. If there is an error in checking or updating the boosted creature data,
// it responds with an internal server error. Otherwise, it responds with a JSON containing the
// boosted creature status and the current race IDs for the creature and boss.
func HandleBoostedCreature(c *gin.Context, db *sql.DB, creatureRaceID *uint32, bossRaceID *uint32) {
	if err := checkAndUpdateBoostedCreatureData(db, creatureRaceID, bossRaceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"boostedcreature": true,
		"creatureraceid":  creatureRaceID,
		"bossraceid":      bossRaceID,
	})
}
