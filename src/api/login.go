package api

import (
	"context"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"net/http"
	"time"
)

func (_api *Api) login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	payload, err := validateLoginPayload(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	defer r.Body.Close()

	grpcClient := login_proto_messages.NewLoginServiceClient(_api.GrpcConnection)

	res, err := grpcClient.Login(
		context.Background(),
		&login_proto_messages.LoginRequest{Email: payload.Email, Password: payload.Password},
	)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	if res.GetError() != nil {
		processErrorResponse(w, buildErrorPayloadFromMessage(res), logger.BuildRequestLogFields(r, start))
		return
	}

	respondAndLog(
		w,
		http.StatusOK,
		buildPayloadFromMessage(res),
		logger.BuildRequestLogFields(r, start),
	)
}

func validateLoginPayload(r *http.Request) (*models.RequestPayload, error) {
	var payload models.RequestPayload

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		logger.Error(err)
		return nil, errors.New("Invalid request payload")
	}

	if payload.Type != "login" {
		r.Body.Close()
		return nil, errors.New("Non-login requests are not accepted")
	}

	return &payload, nil
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
