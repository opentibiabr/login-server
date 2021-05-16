package api_http

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api_http/api_errors"
	"github.com/opentibiabr/login-server/src/api_http/login"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const DefaultLoginErrorCode = 3

func respondAndLogLoginError(w http.ResponseWriter, error api_errors.LoginErrorPayload, fields logrus.Fields) {
	respondWithJSON(w, http.StatusOK, error)
	logger.LogRequest(http.StatusOK, error, "unsuccessful login", fields)
}

func (_api *HttpApi) login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	payload, err := validateLoginPayload(r)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	acc, apiError := LoadAccount(payload, _api.DB)
	if apiError != nil {
		respondAndLogLoginError(w, *apiError, logger.BuildRequestLogFields(r, start))
		return
	}

	players := &database.Players{AccountID: acc.ID}

	err = database.LoadPlayers(_api.DB, players)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondAndLog(
		w,
		http.StatusOK,
		BuildLoginResponsePayload(*acc, *players),
		logger.BuildRequestLogFields(r, start),
	)
}

func validateLoginPayload(r *http.Request) (*login.RequestPayload, error) {
	var payload login.RequestPayload

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

func LoadAccount(payload *login.RequestPayload, DB *sql.DB) (*database.Account, *api_errors.LoginErrorPayload) {
	acc := database.Account{Email: payload.Email, Password: payload.Password}
	if err := acc.Authenticate(DB); err != nil {
		logger.Debug(err.Error())
		return nil, &api_errors.LoginErrorPayload{
			ErrorCode:    DefaultLoginErrorCode,
			ErrorMessage: "Account email or password is not correct.",
		}
	}

	return &acc, nil
}

func BuildLoginResponsePayload(
	acc database.Account,
	players database.Players,
) login.ResponsePayload {
	session := acc.GetSession()
	var characters []login.CharacterPayload
	for _, player := range players.Players {
		if session.LastLoginTime < player.LastLogin {
			session.LastLoginTime = player.LastLogin
		}

		characters = append(characters, player.ToCharacterPayload())
	}

	return login.ResponsePayload{
		PlayData: login.PlayData{
			Worlds:     []login.World{login.LoadWorld()},
			Characters: characters,
		},
		Session: session,
	}
}
