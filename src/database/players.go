package database

import (
	"database/sql"
	"fmt"
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/logger"
)

type Players struct {
	AccountID int
	Players   []Player
}

type Player struct {
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Sex        int    `json:"sex"`
	Vocation   int    `json:"vocation"`
	LookType   int    `json:"looktype"`
	LookHead   int    `json:"lookhead"`
	LookBody   int    `json:"lookbody"`
	LookLegs   int    `json:"looklegs"`
	LookFeet   int    `json:"lookfeet"`
	LookAddons int    `json:"lookaddons"`
	LastLogin  int    `json:"lastlogin"`
}

func LoadPlayers(db *sql.DB, players *Players) error {
	statement := fmt.Sprintf(
		`SELECT name, level, sex, vocation, looktype, 
			lookhead, lookbody, looklegs, lookfeet, lookaddons, 
			lastlogin from players where account_id = "%d"`,
		players.AccountID,
	)

	rows, err := db.Query(statement)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		player := Player{}

		err := player.load(rows)
		if err != nil {
			logger.Error(err)
			return err
		}

		players.Players = append(players.Players, player)
	}

	return nil
}

func (player *Player) load(rows *sql.Rows) error {
	if err := rows.Scan(
		&player.Name,
		&player.Level,
		&player.Sex,
		&player.Vocation,
		&player.LookType,
		&player.LookHead,
		&player.LookBody,
		&player.LookLegs,
		&player.LookFeet,
		&player.LookAddons,
		&player.LastLogin,
	); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (player *Player) ToCharacterPayload() login.CharacterPayload {
	return login.CharacterPayload{
		WorldID: 0,
		CharacterInfo: login.CharacterInfo{
			Name:     player.Name,
			Level:    player.Level,
			Vocation: configs.GetServerVocations()[player.Vocation],
			IsMale:   player.Sex == 1,
		},
		Outfit: login.Outfit{
			OutfitID:    player.LookType,
			HeadColor:   player.LookHead,
			TorsoColor:  player.LookBody,
			LegsColor:   player.LookLegs,
			DetailColor: player.LookFeet,
			AddonsFlags: player.LookAddons,
		},
	}
}
