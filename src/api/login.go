package api

import (
	"context"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"net/http"
)

func (_api *Api) login(c *gin.Context) {
	var payload models.RequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Type != "login" {
		c.JSON(http.StatusNotImplemented, gin.H{"status": "not implemented"})
		return
	}

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
