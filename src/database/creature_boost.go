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
