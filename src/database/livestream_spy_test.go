package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const spyTargetQuery = `SELECT p.name, p.level, p.sex, p.vocation, p.looktype, p.lookhead, p.lookbody, p.looklegs, p.lookfeet, p.lookaddons, p.lastlogin FROM account_livestream_spy_targets st INNER JOIN players p ON p.id = st.target_id INNER JOIN accounts a ON a.id = st.account_id WHERE st.account_id = \? AND a.type = 6 AND p.deletion = 0 AND NOT EXISTS \( SELECT 1 FROM players owned WHERE owned.account_id = st.account_id AND owned.name = CONCAT\('Spy: ', p.name\) \) ORDER BY p.name ASC LIMIT 64`

func TestLoadLivestreamSpyTargetsGODProjectionAndSessionIsolation(t *testing.T) {
	t.Setenv("VOCATIONS", "None,Sorcerer,Druid,Paladin,Knight")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	acc := &Account{ID: 42, Type: 6, LastLogin: 123}
	// Saved targets need not be online; the query deliberately has no online or stream join.
	rows := sqlmock.NewRows(livestreamCasterColumns()).
		AddRow("Offline Target", 250, 1, 4, 128, 10, 20, 30, 40, 3, 999999).
		AddRow("Unknown Vocation", 50, 0, -1, 129, 1, 2, 3, 4, 0, 456)
	mock.ExpectQuery(spyTargetQuery).WithArgs(acc.ID).WillReturnRows(rows).RowsWillBeClosed()

	targets, err := LoadLivestreamSpyTargets(context.Background(), db, acc)
	require.NoError(t, err)
	require.Len(t, targets, 2)
	assert.Equal(t, "Spy: Offline Target", targets[0].Info.Name)
	assert.Equal(t, uint32(250), targets[0].Info.Level)
	assert.Equal(t, uint32(1), targets[0].Info.Sex)
	assert.Equal(t, "Knight", targets[0].Info.Vocation)
	assert.Equal(t, uint32(999999), targets[0].Info.LastLogin)
	assert.Equal(t, uint32(0), targets[0].WorldId)
	assert.Equal(t, uint32(128), targets[0].Outfit.LookType)
	assert.Equal(t, uint32(10), targets[0].Outfit.LookHead)
	assert.Equal(t, uint32(20), targets[0].Outfit.LookBody)
	assert.Equal(t, uint32(30), targets[0].Outfit.LookLegs)
	assert.Equal(t, uint32(40), targets[0].Outfit.LookFeet)
	assert.Equal(t, uint32(3), targets[0].Outfit.Addons)
	assert.Empty(t, targets[1].Info.Vocation)
	assert.Equal(t, uint32(123), acc.LastLogin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadLivestreamSpyTargetsRequiresExactGODType(t *testing.T) {
	for _, accountType := range []uint32{0, 1, 4, 5, 7} {
		targets, err := LoadLivestreamSpyTargets(context.Background(), nil, &Account{ID: 42, Type: accountType})
		assert.NoError(t, err)
		assert.Nil(t, targets, "GM, access groups and higher numeric types must not authorize spy lists")
	}
}

func TestLoadLivestreamSpyTargetsFreshRoleOrEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectQuery(spyTargetQuery).WithArgs(uint32(42)).
		WillReturnRows(sqlmock.NewRows(livestreamCasterColumns())).RowsWillBeClosed()
	// A cached GOD account may have been demoted; the fresh accounts join yields no rows.
	targets, err := LoadLivestreamSpyTargets(context.Background(), db, &Account{ID: 42, Type: 6})
	assert.NoError(t, err)
	assert.Empty(t, targets)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadLivestreamSpyTargetsErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		err     error
		missing bool
	}{
		{"missing watch table", &mysqlDriver.MySQLError{Number: 1146, Message: "Table 'game.account_livestream_spy_targets' doesn't exist"}, true},
		{"missing players table", &mysqlDriver.MySQLError{Number: 1146, Message: "Table 'game.players' doesn't exist"}, false},
		{"query failed", errors.New("fixture database unavailable"), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectQuery(spyTargetQuery).WithArgs(uint32(42)).WillReturnError(test.err)
			targets, err := LoadLivestreamSpyTargets(context.Background(), db, &Account{ID: 42, Type: 6})
			assert.Nil(t, targets)
			if test.missing {
				assert.ErrorIs(t, err, ErrLivestreamSpyTargetsUnavailable)
			} else {
				assert.ErrorIs(t, err, test.err)
				publicErr, ok := serviceerrors.FromError(err)
				require.True(t, ok)
				assert.Equal(t, serviceerrors.CodeCharacterListLoadFailed, publicErr.Code)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLoadLivestreamSpyTargetsRowErrorsDoNotReturnPartialList(t *testing.T) {
	for _, scanFailure := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		rows := sqlmock.NewRows(livestreamCasterColumns()).
			AddRow("First", 100, 1, 3, 128, 1, 2, 3, 4, 0, 123)
		if scanFailure {
			rows.AddRow("Bad Level", "not a number", 1, 3, 128, 1, 2, 3, 4, 0, 123)
		} else {
			rows.AddRow("Second", 100, 1, 3, 128, 1, 2, 3, 4, 0, 123).RowError(1, errors.New("fixture row error"))
		}
		mock.ExpectQuery(spyTargetQuery).WithArgs(uint32(42)).WillReturnRows(rows).RowsWillBeClosed()
		targets, err := LoadLivestreamSpyTargets(context.Background(), db, &Account{ID: 42, Type: 6})
		assert.Error(t, err)
		assert.Nil(t, targets)
		assert.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}
}
