package grpc_server

import (
	"context"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/proto"
	"github.com/opentibiabr/login-server/src/logger"
	"google.golang.org/grpc"
)

func (ls *GrpcServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	acc, err := database.LoadAccount(in.Email, in.Password, ls.DB)
	if err != nil {
		return &proto.LoginResponse{
			Error: &proto.Error{
				Code:    3,
				Message: err.Error(),
			},
		}, nil
	}

	players := &database.Players{AccountID: acc.ID}

	err = database.LoadPlayers(ls.DB, players)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		PlayData: &proto.PlayData{
			Characters: []*proto.Character{
				{
					Info: &proto.CharacterInfo{
						Name: "Junior",
					},
				},
			},
		},
		Session: &proto.Session{},
	}, nil
}

func Client(request *proto.LoginRequest) (*proto.LoginResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":7171", grpc.WithInsecure())
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	c := proto.NewLoginServiceClient(conn)

	return c.Login(context.Background(), request)
}
