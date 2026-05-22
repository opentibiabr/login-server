package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
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

func Test_buildSessionKey(t *testing.T) {
	assert.Equal(t, "user@example.invalid\npassword", buildSessionKey("ignored", true, "user@example.invalid", "password"))
	assert.Equal(t, "abc", buildSessionKey("abc", false, "user@example.invalid", "password"))
}

func Test_authTypeIsPassword(t *testing.T) {
	assert.False(t, (&Api{}).authTypeIsPassword())

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.lua")
	err := os.WriteFile(configPath, []byte("authType = \"password\"\n"), 0o600)
	assert.Nil(t, err)

	manager, err := configs.NewLuaConfigManager(configPath)
	assert.Nil(t, err)
	assert.True(t, (&Api{LuaConfigManager: manager}).authTypeIsPassword())
}

func mapKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

type testLoginService struct {
	login_proto_messages.UnimplementedLoginServiceServer
	response *login_proto_messages.LoginResponse
}

func (svc *testLoginService) Login(_ context.Context, _ *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	return svc.response, nil
}

func newInMemoryLoginClient(t *testing.T, response *login_proto_messages.LoginResponse) *grpc.ClientConn {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()
	login_proto_messages.RegisterLoginServiceServer(grpcServer, &testLoginService{response: response})
	go func() {
		_ = grpcServer.Serve(listener)
	}()

	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(
		dialCtx,
		"bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	dialCancel()
	if err != nil {
		t.Fatalf("failed to create grpc client connection: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
	})

	return conn
}

type loginResponsePayload struct {
	Session struct {
		SessionKey string `json:"sessionkey"`
	} `json:"session"`
}

func Test_loginHandlerReturnsSessionFlowVariants(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		sessionKey string
		authType   string
		assertions func(*testing.T, loginResponsePayload)
	}{
		{
			name:       "legacy session key",
			sessionKey: "user@example.com\npassword123",
			authType:   "",
			assertions: func(t *testing.T, payload loginResponsePayload) {
				assert.Equal(t, "user@example.com\npassword123", payload.Session.SessionKey)
			},
		},
		{
			name:       "random token session key",
			sessionKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			authType:   "",
			assertions: func(t *testing.T, payload loginResponsePayload) {
				assert.Equal(t, "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff", payload.Session.SessionKey)
			},
		},
		{
			name:       "password auth rewrites session key",
			sessionKey: "opaque-from-grpc",
			authType:   "password",
			assertions: func(t *testing.T, payload loginResponsePayload) {
				assert.Equal(t, "user@example.com\npassword123", payload.Session.SessionKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := []string{"/", "/login.php"}

			api := &Api{
				GrpcConnection: newInMemoryLoginClient(t, &login_proto_messages.LoginResponse{
					Session: &login_proto_messages.Session{
						SessionKey: tt.sessionKey,
					},
					PlayData: &login_proto_messages.PlayData{},
				}),
			}
			if tt.authType != "" {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.lua")
				err := os.WriteFile(configPath, []byte("authType = \""+tt.authType+"\"\n"), 0o600)
				assert.NoError(t, err)
				manager, err := configs.NewLuaConfigManager(configPath)
				assert.NoError(t, err)
				api.LuaConfigManager = manager
			}

			router := gin.New()
			router.POST("/", api.login)
			router.POST("/login.php", api.login)

			requestBody, _ := json.Marshal(models.RequestPayload{
				Type:         "login",
				Email:        "user@example.com",
				Password:     "password123",
				DeviceCookie: "test-device",
			})

			for _, endpoint := range endpoints {
				t.Run(endpoint, func(t *testing.T) {
					request := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestBody))

					recorder := httptest.NewRecorder()
					router.ServeHTTP(recorder, request)
					assert.Equal(t, http.StatusOK, recorder.Code)

					var payload loginResponsePayload
					err := json.Unmarshal(recorder.Body.Bytes(), &payload)
					assert.NoError(t, err)
					tt.assertions(t, payload)
				})
			}
		})
	}
}

func Test_loginHandlerReturnsNamedErrorWhenGrpcConnectionIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", (&Api{}).login)

	requestBody, _ := json.Marshal(models.RequestPayload{
		Type:     "login",
		Email:    "user@example.com",
		Password: "password123",
	})
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var payload models.LoginErrorPayload
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, serviceerrors.CodeLoginServiceUnavailable, payload.ErrorCode)
	assert.Equal(t, "Login service error. Please contact support. Error: LOGIN_SERVICE_UNAVAILABLE (LS-3001).", payload.ErrorMessage)
}
