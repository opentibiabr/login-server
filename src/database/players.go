package database

import (
	"database/sql"
	"fmt"

	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

func LoadPlayers(db *sql.DB, acc *Account) ([]*login_proto_messages.Character, error) {
	var players []*login_proto_messages.Character

	statement := fmt.Sprintf(
		`SELECT name, level, sex, vocation, looktype, lookhead, lookbody, looklegs,
			lookfeet, lookaddons, lastlogin from players where account_id = "%d"`,
		acc.ID,
	)

	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	vocations := configs.GetServerVocations()
	for rows.Next() {
		player := login_proto_messages.Character{
			WorldId: 0,
			Info:    &login_proto_messages.CharacterInfo{},
			Outfit:  &login_proto_messages.CharacterOutfit{},
		}

		var vocation int

		err := rows.Scan(
			&player.Info.Name, &player.Info.Level, &player.Info.Sex, &vocation,
			&player.Outfit.LookType, &player.Outfit.LookHead, &player.Outfit.LookBody,
			&player.Outfit.LookLegs, &player.Outfit.LookFeet, &player.Outfit.Addons,
			&player.Info.LastLogin,
		)

		if err != nil {
			return nil, err
		}

		if acc.LastLogin < player.Info.LastLogin {
			acc.LastLogin = player.Info.LastLogin
		}

		if vocation < len(vocations) {
			player.Info.Vocation = vocations[vocation]
		}

		players = append(players, &player)
	}

	return players, nil
}
