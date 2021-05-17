package database

import (
	"database/sql"
	"fmt"
	"github.com/opentibiabr/login-server/src/api/models"
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

func LoadPlayersGrpc(db *sql.DB, accountId int, players []*login_proto_messages.Character) error {
	statement := fmt.Sprintf(
		`SELECT name, level, sex, vocation, 
			lastlogin, looktype, lookhead, lookbody, looklegs, lookfeet, 
			lookaddons from players where account_id = "%d"`,
		accountId,
	)

	rows, err := db.Query(statement)
	if err != nil {
		logger.Error(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		player := login_proto_messages.Character{}

		err := loadGrpc(&player, rows)
		if err != nil {
			logger.Error(err)
			return err
		}

		players = append(players, &player)
	}

	return nil
}

func loadGrpc(player *login_proto_messages.Character, rows *sql.Rows) error {
	if err := rows.Scan(
		&player.Info.Name,
		&player.Info.Level,
		&player.Info.Sex,
		&player.Info.Vocation,
		&player.Info.LastLogin,
		&player.Outfit.LookType,
		&player.Outfit.LookHead,
		&player.Outfit.LookBody,
		&player.Outfit.LookLegs,
		&player.Outfit.LookFeet,
		&player.Outfit.Addons,
	); err != nil {
		logger.Error(err)
		return err
	}

	return nil
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

func (player *Player) ToCharacterPayload() models.CharacterPayload {
	return models.CharacterPayload{
		WorldID: 0,
		CharacterInfo: models.CharacterInfo{
			Name:     player.Name,
			Level:    player.Level,
			Vocation: configs.GetServerVocations()[player.Vocation],
			IsMale:   player.Sex == 1,
		},
		Outfit: models.Outfit{
			OutfitID:    player.LookType,
			HeadColor:   player.LookHead,
			TorsoColor:  player.LookBody,
			LegsColor:   player.LookLegs,
			DetailColor: player.LookFeet,
			AddonsFlags: player.LookAddons,
		},
	}
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
