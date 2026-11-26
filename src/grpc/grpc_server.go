package grpc_login_server

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/opentibiabr/login-server/src/authenticator"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/passwords"
	"github.com/opentibiabr/login-server/src/server"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	DB                         *sql.DB
	AuthenticatorEncryptionKey string
	// Passwords verifies account passwords; nil uses the defaults, which never rewrite a stored value.
	Passwords *passwords.Hasher
	login_proto_messages.LoginServiceServer
	server.ServerInterface
}

func Initialize(gConfigs configs.GlobalConfigs) *GrpcServer {
	var ls GrpcServer

	ls.DB = database.PullConnection(gConfigs)
	ls.AuthenticatorEncryptionKey = gConfigs.LoginServerConfigs.Authenticator.EncryptionKey
	rehash, concurrency := configs.GetLoginPasswordRehash(), configs.GetLoginPasswordHashConcurrency()
	ls.Passwords = passwords.NewHasher(concurrency, rehash)
	serviceerrors.SetPasswordChangeRequiredMessage(configs.GetLoginPasswordChangeMessage())
	logger.Info(fmt.Sprintf("Account password_change_required enforced: %t", database.DetectPasswordChangeRequired(context.Background(), ls.DB)))
	logger.Info(fmt.Sprintf("Account passwords: upgrade after login %t, %d simultaneous Argon2id computations", rehash, concurrency))

	return &ls
}

func (ls *GrpcServer) Run(gConfigs configs.GlobalConfigs) error {
	if _, err := authenticator.DecodeEncryptionKey(ls.AuthenticatorEncryptionKey); err != nil {
		return fmt.Errorf("invalid authenticator configuration: %w", err)
	}

	c, err := net.Listen("tcp", gConfigs.LoginServerConfigs.Grpc.Format())

	if err != nil {
		return err
	}

	server := grpc.NewServer()
	login_proto_messages.RegisterLoginServiceServer(server, ls)

	return server.Serve(c)
}

func (ls *GrpcServer) GetName() string {
	return "gRPC"
}
