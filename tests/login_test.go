package tests

import (
	"bou.ke/monkey"
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/opentibiabr/login-server/src/api"
	"github.com/opentibiabr/login-server/src/api/api_errors"
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/config"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/tests/testlib"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a api.Api

func TestMain(m *testing.M) {
	err := os.Setenv(config.EnvRunSilent, "true")
	if err != nil {}

	a = api.Api{}
	a.Initialize()
	code := m.Run()
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func TestBuildLoginResponsePayload(t *testing.T) {
	asserter := testlib.Assert{T: *t}

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

	payload := api.BuildLoginResponsePayload(a.Configs, acc, players)

	expectedSession := acc.GetSession()
	expectedSession.LastLoginTime = 5123412

	expectedWorld := login.LoadWorld(a.Configs)

	asserter.Equals(expectedSession, payload.Session)
	asserter.Equals(1, len(payload.PlayData.Worlds))
	asserter.Equals(expectedWorld, payload.PlayData.Worlds[0])
	asserter.Equals(1, len(payload.PlayData.Characters))
	asserter.Equals(player.ToCharacterPayload(), payload.PlayData.Characters[0])
}

func TestLoginInvalidPayloadReturn400(t *testing.T) {
	var count = 0
	monkey.Patch(api.BuildLoginResponsePayload, func(
		configs config.Configs,
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	asserter := testlib.Assert{T: *t}

	payload := []byte(`{"type"="login"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m map[string]string
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	asserter.Equals(http.StatusBadRequest, response.Code)
	asserter.Equals("Invalid request payload", m["errors"])
	asserter.Equals(0, count)
}

func TestLoginInvalidCredentialsReturnLoginError(t *testing.T) {
	var count = 0
	monkey.Patch(api.BuildLoginResponsePayload, func(
		configs config.Configs,
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	asserter := testlib.Assert{T: *t}

	payload := []byte(`{"type":"login"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m api_errors.LoginErrorPayload
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	asserter.Equals(http.StatusOK, response.Code)
	asserter.Equals("Account email or password is not correct.", m.ErrorMessage)
	asserter.Equals(3, m.ErrorCode)
	asserter.Equals(0, count)
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
		configs config.Configs,
		acc database.Account,
		players database.Players,
	) login.ResponsePayload {
		count++
		return login.ResponsePayload{}
	})

	asserter := testlib.Assert{T: *t}

	payload := []byte(`{"type":"login","email":"@god","password":"2"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m login.ResponsePayload
	err := json.Unmarshal(response.Body.Bytes(), &m)
	if err != nil {
		log.Print("Error on parse bytes")
	}

	asserter.Equals(http.StatusOK, response.Code)
	asserter.Equals(3, count)
}
