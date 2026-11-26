package grpc_login_server

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func flaggedLoginFixture(t *testing.T, flag int64, in *login_proto_messages.LoginRequest) (*GrpcServer, sqlmock.Sqlmock) {
	t.Helper()
	t.Setenv("SERVER_PATH", "")
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// The detection query only answers about the schema; the mock reports the column as present.
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	require.True(t, database.DetectPasswordChangeRequired(context.Background(), db))
	t.Cleanup(func() {
		off, offMock, _ := sqlmock.New()
		defer off.Close()
		offMock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		database.DetectPasswordChangeRequired(context.Background(), off)
	})

	expectAccountCredentials(mock, in, 42, 1)
	mock.ExpectQuery("SELECT password_change_required FROM accounts WHERE id").
		WithArgs(uint32(42)).WillReturnRows(sqlmock.NewRows([]string{"password_change_required"}).AddRow(flag))
	return &GrpcServer{DB: db}, mock
}

// A flagged account is refused after the password check: the character list, the second factor, the trusted device
// rotation and the session insert are all statements sqlmock would reject, since none is expected.
func TestLoginRefusesAFlaggedAccountEvenWithATrustedDeviceToken(t *testing.T) {
	in := &login_proto_messages.LoginRequest{
		Email:              "flagged@example.invalid",
		Password:           "fixture-password",
		TrustedDeviceToken: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		TrustDevice:        true,
		DeviceName:         "pc",
	}
	server, mock := flaggedLoginFixture(t, 1, in)

	response, err := server.Login(context.Background(), in)

	require.NoError(t, err)
	require.NotNil(t, response.GetError())
	assert.Equal(t, uint32(serviceerrors.CodePasswordChangeRequired), response.GetError().GetCode())
	assert.Equal(t, serviceerrors.DefaultPasswordChangeRequiredMessage, response.GetError().GetMessage())
	assert.Nil(t, response.GetPlayData())
	assert.Nil(t, response.GetSession())
	assert.Empty(t, response.GetTrustedDeviceToken())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginRefusesAFlaggedAccountWithATokenToo(t *testing.T) {
	in := &login_proto_messages.LoginRequest{Email: "flagged@example.invalid", Password: "fixture-password", Token: "123456"}
	server, mock := flaggedLoginFixture(t, 1, in)

	response, err := server.Login(context.Background(), in)

	require.NoError(t, err)
	assert.Equal(t, uint32(serviceerrors.CodePasswordChangeRequired), response.GetError().GetCode())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginUsesTheConfiguredPasswordChangeMessage(t *testing.T) {
	serviceerrors.SetPasswordChangeRequiredMessage("Your password must be changed on example.invalid before you can play.")
	t.Cleanup(func() { serviceerrors.SetPasswordChangeRequiredMessage("") })
	in := &login_proto_messages.LoginRequest{Email: "flagged@example.invalid", Password: "fixture-password"}
	server, _ := flaggedLoginFixture(t, 1, in)

	response, err := server.Login(context.Background(), in)

	require.NoError(t, err)
	assert.Equal(t, "Your password must be changed on example.invalid before you can play.", response.GetError().GetMessage())
}
