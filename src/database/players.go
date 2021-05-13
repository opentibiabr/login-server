package database

import (
	"database/sql"
	"fmt"
	"login-server/src/api/login"
)

var Vocations = []string{
	"None",
	"Sorcerer",
	"Druid",
	"Paladin",
	"Knight",
	"Master Sorcerer",
	"Elder Druid",
	"Royal Paladin",
	"Elite Knight",
	"Sorcerer Dawnport",
	"Druid Dawnport",
	"Paladin Dawnport",
	"Knight Dawnport",
}

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

func (players *Players) Load(db *sql.DB) error {
	statement := fmt.Sprintf(
		`SELECT name, level, sex, vocation, looktype, 
			lookhead, lookbody, looklegs, lookfeet, lookaddons, 
			lastlogin from players where account_id = "%d"`,
		players.AccountID,
	)

	rows, err := db.Query(statement)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		player := Player{}

		err := player.load(rows)
		if err != nil {
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
			Vocation: Vocations[player.Vocation],
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
