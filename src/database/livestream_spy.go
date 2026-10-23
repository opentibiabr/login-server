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
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const livestreamSpyAccountType = 6
const LivestreamSpyCharacterPrefix = "Spy: "

var ErrLivestreamSpyTargetsUnavailable = errors.New("livestream spy target table unavailable")

// LoadLivestreamSpyTargets adds saved targets to the login list without treating
// them as owned characters or changing the account's session metadata.
func LoadLivestreamSpyTargets(ctx context.Context, db *sql.DB, acc *Account) ([]*login_proto_messages.Character, error) {
	if acc == nil {
		return nil, livestreamSpyListError(errors.New("account is nil"))
	}
	if acc.Type != livestreamSpyAccountType {
		return nil, nil
	}
	if db == nil {
		return nil, livestreamSpyListError(errors.New("database connection is nil"))
	}
	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT p.name, p.level, p.sex, p.vocation, p.looktype, p.lookhead, p.lookbody,
			p.looklegs, p.lookfeet, p.lookaddons, p.lastlogin
		FROM account_livestream_spy_targets st
		INNER JOIN players p ON p.id = st.target_id
		INNER JOIN accounts a ON a.id = st.account_id
		WHERE st.account_id = ? AND a.type = 6 AND p.deletion = 0
		ORDER BY p.name ASC LIMIT 64`
	rows, err := db.QueryContext(ctx, query, acc.ID)
	if err != nil {
		if isMissingLivestreamSpyTargetsTable(err) {
			return nil, fmt.Errorf("%w: %v", ErrLivestreamSpyTargetsUnavailable, err)
		}
		return nil, livestreamSpyListError(err)
	}
	defer rows.Close()

	vocations := configs.GetServerVocations()
	targets := make([]*login_proto_messages.Character, 0)
	for rows.Next() {
		target := login_proto_messages.Character{
			WorldId: 0,
			Info:    &login_proto_messages.CharacterInfo{},
			Outfit:  &login_proto_messages.CharacterOutfit{},
		}
		var vocation int
		if err := rows.Scan(
			&target.Info.Name, &target.Info.Level, &target.Info.Sex, &vocation,
			&target.Outfit.LookType, &target.Outfit.LookHead, &target.Outfit.LookBody,
			&target.Outfit.LookLegs, &target.Outfit.LookFeet, &target.Outfit.Addons,
			&target.Info.LastLogin,
		); err != nil {
			return nil, livestreamSpyListError(err)
		}
		if vocation >= 0 && vocation < len(vocations) {
			target.Info.Vocation = vocations[vocation]
		}
		target.Info.Name = LivestreamSpyCharacterPrefix + target.Info.Name
		targets = append(targets, &target)
	}
	if err := rows.Err(); err != nil {
		return nil, livestreamSpyListError(err)
	}
	return targets, nil
}

func livestreamSpyListError(err error) error {
	return serviceerrors.LoginService(
		serviceerrors.CodeCharacterListLoadFailed,
		"CHARACTER_LIST_LOAD_FAILED",
		fmt.Errorf("load livestream spy targets: %w", err),
	)
}

func isMissingLivestreamSpyTargetsTable(err error) bool {
	const table = "account_livestream_spy_targets"
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1146 && strings.Contains(strings.ToLower(mysqlErr.Message), table)
	}
	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, table) && strings.Contains(lowered, "doesn't exist")
}
