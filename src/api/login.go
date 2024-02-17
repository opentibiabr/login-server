package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

func (_api *Api) login(c *gin.Context) {
	var payload models.RequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch payload.Type {
	case "eventschedule":
		database.HandleEventSchedule(c, _api.ServerPath+"/data/XML/events.xml")
	case "boostedcreature":
		database.HandleBoostedCreature(c, _api.DB, &_api.BoostedCreatureID, &_api.BoostedBossID)
	case "login":
		grpcClient := login_proto_messages.NewLoginServiceClient(_api.GrpcConnection)

		res, err := grpcClient.Login(
			context.Background(),
			&login_proto_messages.LoginRequest{Email: payload.Email, Password: payload.Password},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if res.GetError() != nil {
			c.JSON(http.StatusOK, buildErrorPayloadFromMessage(res))
			return
		}

		c.JSON(http.StatusOK, buildPayloadFromMessage(res))
	default:
		c.JSON(http.StatusNotImplemented, gin.H{"status": "not implemented"})
	}
}

func buildPayloadFromMessage(msg *login_proto_messages.LoginResponse) models.ResponsePayload {
	return models.ResponsePayload{
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
