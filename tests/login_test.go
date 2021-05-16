package tests

import (
	"bou.ke/monkey"
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/api/api_errors"
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func TestBuildLoginResponsePayload(t *testing.T) {
	player := database.Player{
		Name:       "Test",
		Level:      12,
		Sex:        12,
		Vocation:   12,
		LookType:   12,
		LookHead:   12,
		LookBody:   12,
		LookLegs:   12,
		LookFeet:   12,
		LookAddons: 12,
		LastLogin:  5123412,
	}

	acc := database.Account{
		ID:       1010,
		Email:    "@test",
		Password: "@test",
		PremDays: 1,
		LastDay:  912481920,
	}

	players := database.Players{
		AccountID: acc.ID,
		Players:   []database.Player{player},
	}

	payload := api.BuildLoginResponsePayload(acc, players)

	expectedSession := acc.GetSession()
	expectedSession.LastLoginTime = 5123412

	expectedWorld := login.LoadWorld()

	assert.Equal(t, expectedSession, payload.Session)
	assert.Equal(t, 1, len(payload.PlayData.Worlds))
	assert.Equal(t, expectedWorld, payload.PlayData.Worlds[0])
	assert.Equal(t, 1, len(payload.PlayData.Characters))
	assert.Equal(t, player.ToCharacterPayload(), payload.PlayData.Characters[0])
}

func TestLoginInvalidPayloadReturn400(t *testing.T) {
	var count = 0
	monkey.Patch(api.BuildLoginResponsePayload, func(
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	payload := []byte(`{"type"="login"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m map[string]string
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Invalid request payload", m["errors"])
	assert.Equal(t, 0, count)
}

func TestLoginInvalidCredentialsReturnLoginError(t *testing.T) {
	var count = 0
	monkey.Patch(api.BuildLoginResponsePayload, func(
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	payload := []byte(`{"type":"login"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m api_errors.LoginErrorPayload
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Account email or password is not correct.", m.ErrorMessage)
	assert.Equal(t, 3, m.ErrorCode)
	assert.Equal(t, 0, count)
}

func TestLoginValidCredentials(t *testing.T) {
	var count = 0

	account := database.Account{}

	monkey.Patch(database.LoadPlayers, func(
		DB *sql.DB,
		players *database.Players,
	) error {
		count++
		return nil
	})

	monkey.Patch(api.LoadAccount, func(
		payload *login.RequestPayload,
		DB *sql.DB,
	) (*database.Account, *api_errors.LoginErrorPayload) {
		count++
		return &account, nil
	})

	monkey.Patch(api.BuildLoginResponsePayload, func(
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	payload := []byte(`{"type":"login","email":"@god","password":"2"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m login.ResponsePayload
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 3, count)
}
