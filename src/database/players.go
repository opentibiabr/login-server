package database

import (
	"database/sql"
	"fmt"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
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

		err := rows.Scan(
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
		)

		if err != nil {
			logger.Error(err)
			return err
		}

		players.Players = append(players.Players, player)
	}

	return nil
}

func (player *Player) ToCharacterMessage() *login_proto_messages.Character {
	info := login_proto_messages.CharacterInfo{
		Name:      player.Name,
		Level:     uint32(player.Level),
		Sex:       uint32(player.Sex),
		Vocation:  configs.GetServerVocations()[player.Vocation],
		LastLogin: uint32(player.LastLogin),
	}

	outfit := login_proto_messages.CharacterOutfit{
		LookType: uint32(player.LookType),
		LookHead: uint32(player.LookHead),
		LookBody: uint32(player.LookBody),
		LookLegs: uint32(player.LookLegs),
		LookFeet: uint32(player.LookFeet),
		Addons:   uint32(player.LookAddons),
	}

	return &login_proto_messages.Character{
		WorldId: 0,
		Info:    &info,
		Outfit:  &outfit,
	}
}
