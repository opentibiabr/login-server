package grpc_login_server

import (
	"context"
	"errors"
	"fmt"

	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/sirupsen/logrus"
)

const DefaultLoginErrorCode = 3
const configurationErrorMessage = "Game world configuration error. Please contact support. Error: %s (LS-%d)."

func (ls *GrpcServer) Login(ctx context.Context, in *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	acc, err := database.LoadAccount(in.Email, in.Password, ls.DB)
	if err != nil {
		return buildLoginErrorResponse(err, false), nil
	}

	if err := configs.ValidateGameServerName(configs.GetGameServerConfigs()); err != nil {
		logger.Error(err)
		configErr := toConfigurationError(err)
		return buildConfigurationErrorResponse(configErr, acc.IsAdmin()), nil
	}

	characters, err := database.LoadPlayers(ls.DB, acc)
	if err != nil {
		return buildLoginErrorResponse(err, acc.IsAdmin()), nil
	}

	sessionKey, err := acc.CreateSession(ctx, ls.DB)
	if err != nil {
		return buildLoginErrorResponse(err, acc.IsAdmin()), nil
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

func buildLoginErrorResponse(err error, includeAdminHint bool) *login_proto_messages.LoginResponse {
	publicErr, ok := serviceerrors.FromError(err)
	if !ok {
		publicErr = serviceerrors.LoginService(
			serviceerrors.CodeLoginServiceUnavailable,
			"LOGIN_SERVICE_UNAVAILABLE",
			err,
		)
	}

	if publicErr.Cause != nil {
		logger.Error(publicErr)
	}
	if includeAdminHint {
		publicErr = serviceerrors.WithHint(publicErr, serviceerrors.AdminHint(publicErr.Name))
	}

	return &login_proto_messages.LoginResponse{
		Error: &login_proto_messages.Error{
			Code:    uint32(publicErr.Code),
			Message: publicErr.Message,
		},
	}
}

func buildConfigurationErrorResponse(configErr *configs.ConfigurationError, includeAdminHint bool) *login_proto_messages.LoginResponse {
	message := fmt.Sprintf(configurationErrorMessage, configErr.Name, configErr.Code)
	if includeAdminHint {
		message = serviceerrors.WithHint(&serviceerrors.PublicError{
			Code:    configErr.Code,
			Name:    configErr.Name,
			Message: message,
		}, serviceerrors.AdminHint(configErr.Name)).Message
	}

	return &login_proto_messages.LoginResponse{
		Error: &login_proto_messages.Error{
			Code:    uint32(configErr.Code),
			Message: message,
		},
	}
}

func toConfigurationError(err error) *configs.ConfigurationError {
	var configErr *configs.ConfigurationError
	if errors.As(err, &configErr) {
		return configErr
	}

	return &configs.ConfigurationError{
		Code:    1000,
		Name:    "UNKNOWN_CONFIG_ERROR",
		Message: err.Error(),
		Cause:   err,
	}
}
