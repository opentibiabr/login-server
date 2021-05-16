package grpc_server

import (
	"context"
	"fmt"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/definitions"
	"github.com/opentibiabr/login-server/src/grpc/proto"
	"google.golang.org/grpc"
	"net"
)

type GrpcServer struct {
	proto.LoginServiceServer
	definitions.ServerInterface
}

func (ls *GrpcServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	return &proto.LoginResponse{Name: fmt.Sprintf("%s%s", in.GetName(), "Res")}, nil
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
