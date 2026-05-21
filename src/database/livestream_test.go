package database

import (
	"errors"
	"testing"

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
		{name: "accepted typo", email: "@livesteam", want: true},
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
	assert.True(t, isMissingLivestreamCastersTable(errors.New("table `active_livestream_casters` doesn't exist")))
	assert.False(t, isMissingLivestreamCastersTable(errors.New("table `players` doesn't exist")))
	assert.False(t, isMissingLivestreamCastersTable(errors.New("connection refused")))
}
