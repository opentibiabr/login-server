package api

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
)

var defaultString = "default"
var defaultNumber = uint32(10)

func Test_buildErrorPayloadFromMessage(t *testing.T) {
	type args struct {
		msg *login_proto_messages.LoginResponse
	}
	tests := []struct {
		name string
		args args
		want models.LoginErrorPayload
	}{{
		"default_error_only_message",
		args{&login_proto_messages.LoginResponse{
			Error: &login_proto_messages.Error{
				Code:    10,
				Message: "Failed",
			},
		}},
		models.LoginErrorPayload{
			ErrorCode:    10,
			ErrorMessage: "Failed",
		},
	}, {
		"error_payload_with_more_info",
		args{&login_proto_messages.LoginResponse{
			Error: &login_proto_messages.Error{
				Code:    10,
				Message: "Failed",
			},
			PlayData: &login_proto_messages.PlayData{
				Characters: []*login_proto_messages.Character{
					{WorldId: 0},
					{WorldId: 2},
				},
			},
		}},
		models.LoginErrorPayload{
			ErrorCode:    10,
			ErrorMessage: "Failed",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildErrorPayloadFromMessage(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildErrorPayloadFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildPayloadFromMessage(t *testing.T) {
	request := models.RequestPayload{
		DeviceCookie: "device-cookie",
		Email:        "player@example.invalid",
	}
	msg := &login_proto_messages.LoginResponse{
		Session: &login_proto_messages.Session{
			IsPremium:    true,
			PremiumUntil: 20,
			SessionKey:   "opaque-session",
			LastLogin:    30,
		},
		PlayData: &login_proto_messages.PlayData{
			Characters: []*login_proto_messages.Character{{
				WorldId: defaultNumber,
				Info: &login_proto_messages.CharacterInfo{
					Name:     defaultString,
					Vocation: defaultString,
					Level:    defaultNumber,
					Sex:      1,
				},
				Outfit: &login_proto_messages.CharacterOutfit{
					LookType: defaultNumber,
					LookHead: defaultNumber,
					LookBody: defaultNumber,
					LookLegs: defaultNumber,
					LookFeet: defaultNumber,
					Addons:   defaultNumber,
				},
			}},
			Worlds: []*login_proto_messages.World{{
				Id:                         defaultNumber,
				Name:                       defaultString,
				ExternalAddress:            "should-not-be-exported",
				ExternalAddressProtected:   defaultString,
				ExternalAddressUnprotected: defaultString,
				ExternalPort:               9999,
				ExternalPortProtected:      defaultNumber,
				ExternalPortUnprotected:    defaultNumber,
				Location:                   defaultString,
			}},
		},
	}

	want := models.ResponsePayload{
		DeviceCookie: "device-cookie",
		LoginEmail:   "player@example.invalid",
		PlayData: models.PlayData{
			Characters: []models.CharacterPayload{{
				WorldID: defaultNumber,
				CharacterInfo: models.CharacterInfo{
					Name:             defaultString,
					Level:            defaultNumber,
					Vocation:         defaultString,
					IsMale:           true,
					Tutorial:         false,
					IsMainCharacter:  false,
					IsHidden:         false,
					DailyRewardState: 0,
				},
				Outfit: models.Outfit{
					OutfitID:    defaultNumber,
					HeadColor:   defaultNumber,
					TorsoColor:  defaultNumber,
					LegsColor:   defaultNumber,
					DetailColor: defaultNumber,
					AddonsFlags: defaultNumber,
				},
			}},
			Worlds: []models.World{{
				ID:                         defaultNumber,
				Name:                       defaultString,
				ExternalAddressProtected:   defaultString,
				ExternalAddressUnprotected: defaultString,
				ExternalPortProtected:      defaultNumber,
				ExternalPortUnprotected:    defaultNumber,
				Location:                   defaultString,
				AntiCheatProtection:        false,
				PreviewState:               0,
				PvpType:                    0,
			}},
		},
		Session: models.Session{
			IsPremium:             true,
			PremiumUntil:          20,
			SessionKey:            "opaque-session",
			LastLoginTime:         30,
			FpsTracking:           false,
			IsReturner:            false,
			OptionTracking:        false,
			RecoverySetupComplete: false,
			ReturnerNotification:  false,
			ShowRewardNews:        false,
			Status:                "active",
		},
	}

	payload := buildPayloadFromMessage(msg, request)
	assert.Equal(t, want, payload)

	var jsonPayload map[string]interface{}
	bytes, err := json.Marshal(payload)
	assert.Nil(t, err)
	assert.Nil(t, json.Unmarshal(bytes, &jsonPayload))

	assert.ElementsMatch(t, []string{"devicecookie", "loginemail", "playdata", "session"}, mapKeys(jsonPayload))

	session := jsonPayload["session"].(map[string]interface{})
	assert.ElementsMatch(t, []string{
		"fpstracking",
		"ispremium",
		"isreturner",
		"lastlogintime",
		"optiontracking",
		"premiumuntil",
		"recoverysetupcomplete",
		"returnernotification",
		"sessionkey",
		"showrewardnews",
		"status",
	}, mapKeys(session))

	playData := jsonPayload["playdata"].(map[string]interface{})
	world := playData["worlds"].([]interface{})[0].(map[string]interface{})
	assert.ElementsMatch(t, []string{
		"anticheatprotection",
		"externaladdressprotected",
		"externaladdressunprotected",
		"externalportprotected",
		"externalportunprotected",
		"id",
		"location",
		"name",
		"previewstate",
		"pvptype",
	}, mapKeys(world))
	_, hasExternalAddress := world["externaladdress"]
	_, hasExternalPort := world["externalport"]
	assert.False(t, hasExternalAddress)
	assert.False(t, hasExternalPort)

	character := playData["characters"].([]interface{})[0].(map[string]interface{})
	assert.ElementsMatch(t, []string{
		"addonsflags",
		"dailyrewardstate",
		"detailcolor",
		"headcolor",
		"ishidden",
		"ismaincharacter",
		"ismale",
		"legscolor",
		"level",
		"name",
		"outfitid",
		"torsocolor",
		"tutorial",
		"vocation",
		"worldid",
	}, mapKeys(character))
	_, hasTournamentParticipant := character["istournamentparticipant"]
	_, hasTournamentPlayTime := character["remainingdailytournamentplaytime"]
	assert.False(t, hasTournamentParticipant)
	assert.False(t, hasTournamentPlayTime)
}

func Test_buildTemporaryErrorPayload(t *testing.T) {
	assert.Equal(t, models.LoginErrorPayload{
		ErrorCode:    2,
		ErrorMessage: "Internal error. Please try again later or contact customer support if the problem persists.",
	}, buildTemporaryErrorPayload())
}

func mapKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}
