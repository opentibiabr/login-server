package api

import (
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/api/api_errors"
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/database"
	grpc_server "github.com/opentibiabr/login-server/src/grpc"
	"github.com/opentibiabr/login-server/src/grpc/proto"
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

func (_api *Api) login2(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	payload, err := validateLoginPayload(r)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	resp, err := grpc_server.Client(&proto.LoginRequest{Email: payload.Email, Password: payload.Password})

	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if resp.GetError() != nil {
		respondAndLogLoginError(
			w,
			api_errors.LoginErrorPayload{
				ErrorCode:    int(resp.Error.Code),
				ErrorMessage: resp.Error.Message,
			},
			logger.BuildRequestLogFields(r, start),
		)
		return
	}

	session := resp.GetSession()
	_session := login.Session{
		IsPremium:    session.GetIsPremium(),
		PremiumUntil: int(session.GetPremiumUntil()),
		SessionKey:   session.GetSessionKey(),
	}

	var characters []login.CharacterPayload
	for _, character := range resp.GetPlayData().GetCharacters() {
		lastLogin := int(character.Info.GetLastLogin())
		if _session.LastLoginTime < lastLogin {
			_session.LastLoginTime = lastLogin
		}

		characters = append(characters, login.CharacterPayload{
			WorldID: 0,
			CharacterInfo: login.CharacterInfo{
				Name:     character.GetInfo().GetName(),
				Level:    int(character.GetInfo().GetLevel()),
				Vocation: configs.GetServerVocations()[int(character.GetInfo().GetLevel())],
				IsMale:   int(character.GetInfo().GetSex()) == 1,
			},
			Outfit: login.Outfit{
				OutfitID:    int(character.GetOutfit().GetLookType()),
				HeadColor:   int(character.GetOutfit().GetLookHead()),
				TorsoColor:  int(character.GetOutfit().GetLookBody()),
				LegsColor:   int(character.GetOutfit().GetLookLegs()),
				DetailColor: int(character.GetOutfit().GetLookFeet()),
				AddonsFlags: int(character.GetOutfit().GetAddons()),
			},
		})
	}

	respPayload := login.ResponsePayload{
		PlayData: login.PlayData{
			Worlds:     []login.World{login.LoadWorld()},
			Characters: characters,
		},
		Session: _session,
	}

	respondAndLog(
		w,
		http.StatusOK,
		respPayload,
		logger.BuildRequestLogFields(r, start),
	)
}

func (_api *Api) login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	payload, err := validateLoginPayload(r)
	if err != nil {
		logger.Error(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	acc, err := database.LoadAccount(payload.Email, payload.Password, _api.DB)
	if err != nil {
		respondAndLogLoginError(
			w,
			api_errors.LoginErrorPayload{
				ErrorCode:    DefaultLoginErrorCode,
				ErrorMessage: err.Error(),
			},
			logger.BuildRequestLogFields(r, start),
		)
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
