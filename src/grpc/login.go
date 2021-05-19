package grpc_login_server

import (
	"context"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
)

const DefaultLoginErrorCode = 3

func (ls *GrpcServer) Login(ctx context.Context, in *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	acc, err := database.LoadAccount(in.Email, in.Password, ls.DB)
	if err != nil {
		return &login_proto_messages.LoginResponse{
			Error: &login_proto_messages.Error{
				Code:    DefaultLoginErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	characters, err := database.LoadPlayers(ls.DB, acc)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	res := login_proto_messages.LoginResponse{
		PlayData: &login_proto_messages.PlayData{
			Characters: characters,
			Worlds:     models.BuildWorldsMessage(configs.GetGameServerConfigs()),
		},
		Session: acc.GetGrpcSession(),
	}

	logger.WithFields(logrus.Fields{
		"0": "gRPC",
		"1": "login",
	}).Debug("processed")

	return &res, nil
}
