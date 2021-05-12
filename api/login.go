package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"login-server/api/api_errors"
	"login-server/api/login"
	"login-server/config"
	"login-server/database"
	"net/http"
)

func logLoginErrorAndRespond(w http.ResponseWriter, r *http.Request, error api_errors.LoginErrorPayload) {
	logResponse(r, http.StatusOK, error)
	respondWithJSON(w, r, http.StatusOK, error)
}

func (_api *Api) login(w http.ResponseWriter, r *http.Request) {
	payload, err := validateLoginPayload(r)
	if err != nil {
		logErrorAndRespond(w, r, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	acc, apiError := loadAccount(payload, _api.DB)
	if apiError != nil {
		logLoginErrorAndRespond(w, r, *apiError)
		return
	}

	players := database.Players{AccountID: acc.ID}

	err = players.Load(_api.DB)
	if err != nil {
		logErrorAndRespond(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	logAndRespond(w, r, http.StatusOK, BuildLoginResponsePayload(_api.Configs, *acc, players))
}

func validateLoginPayload(r *http.Request) (*login.RequestPayload, error) {
	var payload login.RequestPayload

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		return nil, errors.New("Invalid request payload")
	}

	if payload.Type != "login" {
		r.Body.Close()
		return nil, errors.New("Non-login requests are not accepted")
	}

	return &payload, nil
}

func loadAccount(payload *login.RequestPayload, DB *sql.DB) (*database.Account, *api_errors.LoginErrorPayload) {
	acc := database.Account{Email: payload.Email, Password: payload.Password}
	if err := acc.Authenticate(DB); err != nil {
		apiError := api_errors.LoginErrorPayload{
			ErrorCode:    3,
			ErrorMessage: "Account email or password is not correct.",
		}
		return nil, &apiError
	}

	return &acc, nil
}

func BuildLoginResponsePayload(
	configs config.Configs,
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
			Worlds:     []login.World{login.LoadWorld(configs)},
			Characters: characters,
		},
		Session: session,
	}
}
