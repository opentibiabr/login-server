package grpc_login_server

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A regular account that owns a staff character and has no authenticator gets neither the character list nor a session:
// sqlmock fails the test on any statement after the staff character check.
func TestLoginRefusesStaffCharacterOwnerWithoutAuthenticator(t *testing.T) {
	t.Setenv("SERVER_PATH", "")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	in := &login_proto_messages.LoginRequest{Email: "staff@example.invalid", Password: "fixture-password"}
	mock.ExpectQuery("SELECT id, type, premdays, lastday FROM accounts").
		WithArgs(in.Email, in.Email, fmt.Sprintf("%x", sha1.Sum([]byte(in.Password)))).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "premdays", "lastday"}).AddRow(42, 1, 0, 0))
	mock.ExpectQuery(`SELECT id, name, level, sex, vocation`).
		WillReturnRows(ownedCharacterRows().AddRow(5004, "Staff Character", 100, 1, 4, 128, 1, 2, 3, 4, 0, 123))
	mock.ExpectQuery("SELECT secret_encrypted, last_used_step, blocked_until FROM account_authenticators").
		WithArgs(uint32(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT 1 FROM players WHERE account_id = \? AND deletion = 0 AND group_id >= \? LIMIT 1`).
		WithArgs(uint32(42), 4).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	response, err := (&GrpcServer{DB: db}).Login(context.Background(), in)

	require.NoError(t, err)
	require.NotNil(t, response.GetError())
	assert.Equal(t, uint32(serviceerrors.CodeStaffAuthenticatorRequired), response.GetError().Code)
	assert.Nil(t, response.GetPlayData())
	assert.Nil(t, response.GetSession())
	assert.NoError(t, mock.ExpectationsWereMet())
}
