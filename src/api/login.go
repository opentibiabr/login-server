package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const temporaryErrorCode = 2
const temporaryErrorMessage = "Internal error. Please try again later or contact customer support if the problem persists."

func (_api *Api) login(c *gin.Context) {
	var payload models.RequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch payload.Type {
	case "cacheinfo", "news", "newsviewed":
		c.JSON(http.StatusOK, buildTemporaryErrorPayload())
	case "eventschedule":
		database.HandleEventSchedule(c, _api.eventSchedulePath())
	case "boostedcreature":
		database.HandleBoostedCreature(c, _api.DB, &_api.BoostedCreatureID, &_api.BoostedBossID)
	case "login":
		if _api.GrpcConnection == nil {
			writePublicError(c, serviceerrors.LoginService(
				serviceerrors.CodeLoginServiceUnavailable,
				"LOGIN_SERVICE_UNAVAILABLE",
				fmt.Errorf("grpc connection is nil"),
			))
			return
		}

		grpcClient := login_proto_messages.NewLoginServiceClient(_api.GrpcConnection)

		res, err := grpcClient.Login(
			context.Background(),
			&login_proto_messages.LoginRequest{Email: payload.Email, Password: payload.Password},
		)

		if err != nil {
			writePublicError(c, serviceerrors.LoginService(
				serviceerrors.CodeLoginServiceUnavailable,
				"LOGIN_SERVICE_UNAVAILABLE",
				err,
			))
			return
		}

		if res.GetError() != nil {
			c.JSON(http.StatusOK, buildErrorPayloadFromMessage(res))
			return
		}

		response := buildPayloadFromMessage(res, payload)
		response.Session.SessionKey = buildSessionKey(response.Session.SessionKey, _api.authTypeIsPassword(), payload.Email, payload.Password)
		c.JSON(http.StatusOK, response)
	default:
		writePublicError(c, serviceerrors.LoginService(
			serviceerrors.CodeUnsupportedRequestType,
			"UNSUPPORTED_REQUEST_TYPE",
			fmt.Errorf("unsupported login request type %q", payload.Type),
		))
	}
}

func (api *Api) authTypeIsPassword() bool {
	if api == nil || api.LuaConfigManager == nil {
		return false
	}
	return api.LuaConfigManager.GetString("authType") == "password"
}

func buildSessionKey(defaultSessionKey string, authTypeIsPassword bool, email, password string) string {
	if !authTypeIsPassword {
		return defaultSessionKey
	}

	return fmt.Sprintf("%s\n%s", email, password)
}

func buildPayloadFromMessage(msg *login_proto_messages.LoginResponse, request models.RequestPayload) models.ResponsePayload {
	return models.ResponsePayload{
		DeviceCookie: request.DeviceCookie,
		LoginEmail:   request.Email,
		PlayData: models.PlayData{
			Worlds:     models.LoadWorldsFromMessage(msg.PlayData.Worlds),
			Characters: models.LoadCharactersFromMessage(msg.PlayData.Characters),
		},
		Session: models.LoadSessionFromMessage(msg.GetSession()),
	}
}

func buildErrorPayloadFromMessage(msg *login_proto_messages.LoginResponse) models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    int(msg.Error.Code),
		ErrorMessage: msg.Error.Message,
	}
}

func buildTemporaryErrorPayload() models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    temporaryErrorCode,
		ErrorMessage: temporaryErrorMessage,
	}
}

func buildErrorPayloadFromPublicError(err *serviceerrors.PublicError) models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    err.Code,
		ErrorMessage: serviceerrors.MessageWithHint(err),
	}
}

func writePublicError(c *gin.Context, err *serviceerrors.PublicError) {
	if err.Cause != nil {
		logger.Error(err)
	}
	c.JSON(http.StatusOK, buildErrorPayloadFromPublicError(err))
}
