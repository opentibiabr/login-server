package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestIsLivestreamLogin(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "canonical", email: "@livestream", want: true},
		{name: "canonical with spaces and case", email: " @LiveStream ", want: true},
		{name: "plain descriptor", email: "livestream", want: false},
		{name: "empty", email: "", want: false},
		{name: "livestream substring", email: "user@livestream2.com", want: false},
		{name: "regular account", email: "player@example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsLivestreamLogin(tt.email))
		})
	}
}

func TestIsMissingLivestreamCastersTable(t *testing.T) {
	assert.True(t, isMissingLivestreamCastersTable(&mysqlDriver.MySQLError{
		Number:  1146,
		Message: "Table 'otserv.active_livestream_casters' doesn't exist",
	}))
	assert.False(t, isMissingLivestreamCastersTable(&mysqlDriver.MySQLError{
		Number:  1146,
		Message: "Table 'otserv.players' doesn't exist",
	}))
	assert.True(t, isMissingLivestreamCastersTable(errors.New("table `active_livestream_casters` doesn't exist")))
	assert.False(t, isMissingLivestreamCastersTable(errors.New("table `players` doesn't exist")))
	assert.False(t, isMissingLivestreamCastersTable(errors.New("connection refused")))
}

func TestLoadLivestreamCasters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows(livestreamCasterColumns()).
		AddRow("Alice", uint32(100), uint32(1), 3, uint32(128), uint32(10), uint32(20), uint32(30), uint32(40), uint32(0), uint32(123)).
		AddRow("Invalid Vocation", uint32(50), uint32(0), -1, uint32(129), uint32(1), uint32(2), uint32(3), uint32(4), uint32(0), uint32(456))
	mock.ExpectQuery("SELECT p.name").
		WithArgs(LivestreamStatusActive).
		WillReturnRows(rows)

	casters, err := LoadLivestreamCasters(context.Background(), db)
	assert.NoError(t, err)
	assert.Len(t, casters, 2)
	assert.Equal(t, "Alice", casters[0].Info.Name)
	assert.Equal(t, uint32(100), casters[0].Info.Level)
	assert.Equal(t, "Paladin", casters[0].Info.Vocation)
	assert.Equal(t, "Invalid Vocation", casters[1].Info.Name)
	assert.Empty(t, casters[1].Info.Vocation)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadLivestreamCastersWrapsMissingTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT p.name").
		WithArgs(LivestreamStatusActive).
		WillReturnError(&mysqlDriver.MySQLError{
			Number:  1146,
			Message: "Table 'otserv.active_livestream_casters' doesn't exist",
		})

	casters, err := LoadLivestreamCasters(context.Background(), db)
	assert.Nil(t, casters)
	assert.ErrorIs(t, err, ErrLivestreamCastersUnavailable)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func livestreamCasterColumns() []string {
	return []string{
		"name",
		"level",
		"sex",
		"vocation",
		"looktype",
		"lookhead",
		"lookbody",
		"looklegs",
		"lookfeet",
		"lookaddons",
		"lastlogin",
	}
}
