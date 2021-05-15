package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api/api_errors"
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/logger"
	"net/http"
)

const DefaultLoginErrorCode = 3

func logLoginErrorAndRespond(w http.ResponseWriter, r *http.Request, error api_errors.LoginErrorPayload) {
	logger.LogRequest(r, http.StatusOK, error, "unsuccessful login")
	respondWithJSON(w, http.StatusOK, error)
}

func (_api *Api) login(w http.ResponseWriter, r *http.Request) {
	payload, err := validateLoginPayload(r)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	acc, apiError := LoadAccount(payload, _api.DB)
	if apiError != nil {
		logLoginErrorAndRespond(w, r, *apiError)
		return
	}

	players := &database.Players{AccountID: acc.ID}

	err = database.LoadPlayers(_api.DB, players)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logAndRespond(w, r, http.StatusOK, BuildLoginResponsePayload(*acc, *players))
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
