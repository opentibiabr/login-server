package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestLoadPlayersReturnsStableCharacterID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	columns := []string{"id", "name", "level", "sex", "vocation", "looktype", "lookhead", "lookbody", "looklegs", "lookfeet", "lookaddons", "lastlogin"}
	mock.ExpectQuery(`SELECT id, name, level, sex, vocation`).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1073741825, "Example Knight", 100, 1, 0, 128, 1, 2, 3, 4, 3, 1700000000))

	players, err := LoadPlayers(db, &Account{ID: 7})
	assert.Nil(t, err)
	if assert.Len(t, players, 1) {
		assert.Equal(t, uint32(1073741825), players[0].GetCharacterId())
		assert.Equal(t, "Example Knight", players[0].GetInfo().GetName())
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}
