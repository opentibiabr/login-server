package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

const LivestreamSessionAccount = "@livestream"
const LivestreamStatusActive = 1

var ErrLivestreamCastersUnavailable = errors.New("active livestream caster table unavailable")

func IsLivestreamLogin(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	return email == LivestreamSessionAccount
}

func LoadLivestreamCasters(ctx context.Context, db *sql.DB) ([]*login_proto_messages.Character, error) {
	const query = `
		SELECT p.name, p.level, p.sex, p.vocation, p.looktype, p.lookhead, p.lookbody,
			p.looklegs, p.lookfeet, p.lookaddons, p.lastlogin
		FROM players p
		INNER JOIN active_livestream_casters lc ON p.id = lc.caster_id
		WHERE lc.livestream_status >= ?
		ORDER BY lc.livestream_viewers DESC, p.name ASC`

	rows, err := db.QueryContext(ctx, query, LivestreamStatusActive)
	if err != nil {
		if isMissingLivestreamCastersTable(err) {
			return nil, fmt.Errorf("%w: %v", ErrLivestreamCastersUnavailable, err)
		}
		return nil, err
	}
	defer rows.Close()

	vocations := configs.GetServerVocations()
	casters := make([]*login_proto_messages.Character, 0)
	for rows.Next() {
		caster := login_proto_messages.Character{
			WorldId: 0,
			Info:    &login_proto_messages.CharacterInfo{},
			Outfit:  &login_proto_messages.CharacterOutfit{},
		}

		var vocation int
		if err := rows.Scan(
			&caster.Info.Name, &caster.Info.Level, &caster.Info.Sex, &vocation,
			&caster.Outfit.LookType, &caster.Outfit.LookHead, &caster.Outfit.LookBody,
			&caster.Outfit.LookLegs, &caster.Outfit.LookFeet, &caster.Outfit.Addons,
			&caster.Info.LastLogin,
		); err != nil {
			return nil, err
		}

		if vocation >= 0 && vocation < len(vocations) {
			caster.Info.Vocation = vocations[vocation]
		}

		casters = append(casters, &caster)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return casters, nil
}

func isMissingLivestreamCastersTable(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1146 && strings.Contains(strings.ToLower(mysqlErr.Message), "active_livestream_casters")
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "active_livestream_casters") && strings.Contains(lowered, "doesn't exist")
}
