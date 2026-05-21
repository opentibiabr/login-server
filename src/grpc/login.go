package grpc_login_server

import (
	"context"
	"errors"

	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
)

const DefaultLoginErrorCode = 3
const temporaryLoginErrorMessage = "Internal error. Please try again later or contact customer support if the problem persists."
const livestreamUnavailableMessage = "No active livestream casters found, or livestream login is disabled."

func (ls *GrpcServer) Login(ctx context.Context, in *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	if database.IsLivestreamLogin(in.Email) {
		return ls.loginLivestream(ctx, in.Password)
	}

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

	sessionKey, err := acc.CreateSession(ctx, ls.DB)
	if err != nil {
		logger.Error(err)
		if errors.Is(err, database.ErrAccountSessionStorageUnavailable) {
			return &login_proto_messages.LoginResponse{
				Error: &login_proto_messages.Error{
					Code:    DefaultLoginErrorCode,
					Message: temporaryLoginErrorMessage,
				},
			}, nil
		}
		return nil, err
	}

	res := login_proto_messages.LoginResponse{
		PlayData: &login_proto_messages.PlayData{
			Characters: characters,
			Worlds:     models.BuildWorldsMessage(configs.GetGameServerConfigs()),
		},
		Session: acc.GetGrpcSession(sessionKey),
	}

	logger.WithFields(logrus.Fields{
		"0": "gRPC",
		"1": "login",
	}).Debug("processed")

	return &res, nil
}

func (ls *GrpcServer) loginLivestream(ctx context.Context, password string) (*login_proto_messages.LoginResponse, error) {
	characters, err := database.LoadLivestreamCasters(ctx, ls.DB)
	if err != nil {
		logger.Error(err)
		if errors.Is(err, database.ErrLivestreamCastersUnavailable) {
			return buildLivestreamUnavailableResponse(), nil
		}
		return nil, err
	}

	if len(characters) == 0 {
		return buildLivestreamUnavailableResponse(), nil
	}

	return &login_proto_messages.LoginResponse{
		PlayData: &login_proto_messages.PlayData{
			Characters: characters,
			Worlds:     models.BuildWorldsMessage(configs.GetGameServerConfigs()),
		},
		Session: buildLivestreamSession(password),
	}, nil
}

func buildLivestreamUnavailableResponse() *login_proto_messages.LoginResponse {
	return &login_proto_messages.LoginResponse{
		Error: &login_proto_messages.Error{
			Code:    DefaultLoginErrorCode,
			Message: livestreamUnavailableMessage,
		},
	}
}

func buildLivestreamSession(password string) *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    false,
		PremiumUntil: 0,
		SessionKey:   database.LivestreamSessionAccount + "\n" + password,
		LastLogin:    0,
	}
}
