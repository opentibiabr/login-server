package grpc_server

import (
	"database/sql"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/definitions"
	"github.com/opentibiabr/login-server/src/grpc/proto"
	"github.com/opentibiabr/login-server/src/logger"
	"google.golang.org/grpc"
	"net"
)

type GrpcServer struct {
	DB *sql.DB
	proto.LoginServiceServer
	definitions.ServerInterface
}

func Initialize(gConfigs configs.GlobalConfigs) *GrpcServer {
	var ls GrpcServer
	err := configs.Init()
	if err != nil {
		logger.Warn("Failed to load '.env' in dev environment, going with default.")
	}

	ls.DB, err = sql.Open("mysql", gConfigs.DBConfigs.GetConnectionString())
	if err != nil {
		logger.Fatal(err)
	}

	return &ls
}

func (ls *GrpcServer) Run(gConfigs configs.GlobalConfigs) error {
	c, err := net.Listen("tcp", gConfigs.LoginServerConfigs.Tcp.Format())

	if err != nil {
		return err
	}

	server := grpc.NewServer()
	proto.RegisterLoginServiceServer(server, ls)

	return server.Serve(c)
}

func (ls *GrpcServer) GetName() string {
	return "gRPC"
}
