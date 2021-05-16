package grpc_server

import (
	"context"
	"fmt"
	"github.com/opentibiabr/login-server/src/grpc/proto"
	"github.com/opentibiabr/login-server/src/logger"
	"google.golang.org/grpc"
)

func (ls *GrpcServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	return &proto.LoginResponse{Name: fmt.Sprintf("%s%s", in.GetName(), "Res")}, nil
}

func Client(request *proto.LoginRequest) *proto.LoginResponse {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":7171", grpc.WithInsecure())
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	c := proto.NewLoginServiceClient(conn)

	response, err := c.Login(context.Background(), request)
	if err != nil {
		logger.Fatal(err)
	}
	return response
}
