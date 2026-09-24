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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var defaultString = "default"
var defaultNumber = uint32(10)

func TestIsSecureLoginRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, test := range []struct {
		name      string
		url       string
		forwarded string
		secure    bool
	}{
		{name: "plain HTTP", url: "http://login.example/login", secure: false},
		{name: "direct HTTPS", url: "https://login.example/login", secure: true},
		{name: "TLS proxy", url: "http://login.example/login", forwarded: "https", secure: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(http.MethodPost, test.url, nil)
			if test.forwarded != "" {
				context.Request.Header.Set("X-Forwarded-Proto", test.forwarded)
			}
			assert.Equal(t, test.secure, isSecureLoginRequest(context))
		})
	}
}

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
		TrustedDeviceToken:     "rotated-trusted-device",
		TrustedDeviceExpiresAt: 123456,
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
		DeviceCookie:           "device-cookie",
		LoginEmail:             "player@example.invalid",
		TrustedDeviceToken:     "rotated-trusted-device",
		TrustedDeviceExpiresAt: 123456,
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

	assert.ElementsMatch(t, []string{
		"devicecookie", "loginemail", "playdata", "session", "trusteddevicetoken", "trusteddeviceexpiresat",
	}, mapKeys(jsonPayload))

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

func Test_hasIncompatibleAuthType(t *testing.T) {
	assert.True(t, (*Api)(nil).hasIncompatibleAuthType())
	assert.True(t, (&Api{}).hasIncompatibleAuthType())

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.lua")
	err := os.WriteFile(configPath, []byte("authType = \"password\"\n"), 0o600)
	assert.Nil(t, err)

	manager, err := configs.NewLuaConfigManager(configPath)
	assert.Nil(t, err)
	assert.True(t, (&Api{LuaConfigManager: manager}).hasIncompatibleAuthType())

	err = os.WriteFile(configPath, []byte("authType = \"session\"\n"), 0o600)
	assert.Nil(t, err)
	manager, err = configs.NewLuaConfigManager(configPath)
	assert.Nil(t, err)
	assert.False(t, (&Api{LuaConfigManager: manager}).hasIncompatibleAuthType())
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
	request  *login_proto_messages.LoginRequest
}

func (svc *testLoginService) Login(_ context.Context, request *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	svc.request = request
	return svc.response, nil
}

func newInMemoryLoginClient(t *testing.T, response *login_proto_messages.LoginResponse) (*grpc.ClientConn, *testLoginService) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()
	service := &testLoginService{response: response}
	login_proto_messages.RegisterLoginServiceServer(grpcServer, service)
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

	return conn, service
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
			authType:   "session",
			assertions: func(t *testing.T, payload loginResponsePayload) {
				assert.Equal(t, "user@example.com\npassword123", payload.Session.SessionKey)
			},
		},
		{
			name:       "random token session key",
			sessionKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			authType:   "session",
			assertions: func(t *testing.T, payload loginResponsePayload) {
				assert.Equal(t, "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff", payload.Session.SessionKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := []string{"/", "/login.php"}

			connection, loginService := newInMemoryLoginClient(t, &login_proto_messages.LoginResponse{
				Session: &login_proto_messages.Session{
					SessionKey: tt.sessionKey,
				},
				PlayData: &login_proto_messages.PlayData{},
			})
			api := &Api{
				GrpcConnection: connection,
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
				Type:               "login",
				Email:              "user@example.com",
				Password:           "password123",
				Token:              "123456",
				DeviceCookie:       "test-device",
				TrustedDeviceToken: "trusted-device",
				TrustDevice:        true,
				DeviceName:         "OTClient (Windows)",
			})

			for _, endpoint := range endpoints {
				t.Run(endpoint, func(t *testing.T) {
					request := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
					request.Header.Set("X-Forwarded-Proto", "https")

					recorder := httptest.NewRecorder()
					router.ServeHTTP(recorder, request)
					assert.Equal(t, http.StatusOK, recorder.Code)

					var payload loginResponsePayload
					err := json.Unmarshal(recorder.Body.Bytes(), &payload)
					assert.NoError(t, err)
					assert.Equal(t, "123456", loginService.request.GetToken())
					assert.Equal(t, "trusted-device", loginService.request.GetTrustedDeviceToken())
					assert.True(t, loginService.request.GetTrustDevice())
					assert.Equal(t, "OTClient (Windows)", loginService.request.GetDeviceName())
					tt.assertions(t, payload)
				})
			}
		})
	}
}

func TestLoginHandlerDropsTrustedDeviceCredentialsOverPlainHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	connection, loginService := newInMemoryLoginClient(t, &login_proto_messages.LoginResponse{
		Session:                &login_proto_messages.Session{SessionKey: "opaque-session"},
		PlayData:               &login_proto_messages.PlayData{},
		TrustedDeviceToken:     "must-not-cross-plain-http-response",
		TrustedDeviceExpiresAt: 123456,
	})
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.lua")
	require.NoError(t, os.WriteFile(configPath, []byte("authType = \"session\"\n"), 0o600))
	manager, err := configs.NewLuaConfigManager(configPath)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/login", (&Api{GrpcConnection: connection, LuaConfigManager: manager}).login)
	requestBody, err := json.Marshal(models.RequestPayload{
		Type:               "login",
		Email:              "user@example.com",
		Password:           "password123",
		Token:              "123456",
		TrustedDeviceToken: "must-not-cross-plain-http",
		TrustDevice:        true,
		DeviceName:         "OTClient (Windows)",
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "http://login.example/login", bytes.NewBuffer(requestBody))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, loginService.request.GetTrustedDeviceToken())
	assert.False(t, loginService.request.GetTrustDevice())
	assert.Empty(t, loginService.request.GetDeviceName())
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.NotContains(t, response, "trusteddevicetoken")
	assert.NotContains(t, response, "trusteddeviceexpiresat")
}

func Test_loginHandlerReturnsNamedErrorWhenGrpcConnectionIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.lua")
	err := os.WriteFile(configPath, []byte("authType = \"session\"\n"), 0o600)
	require.NoError(t, err)
	manager, err := configs.NewLuaConfigManager(configPath)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/login", (&Api{LuaConfigManager: manager}).login)

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
	err = json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, serviceerrors.CodeLoginServiceUnavailable, payload.ErrorCode)
	assert.Equal(t, "Login service error. Please contact support. Error: LOGIN_SERVICE_UNAVAILABLE (LS-3001).", payload.ErrorMessage)
}

func Test_loginHandlerRejectsMissingAuthenticationConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", (&Api{}).login)

	requestBody, err := json.Marshal(models.RequestPayload{
		Type:     "login",
		Email:    "user@example.com",
		Password: "password123",
		Token:    "123456",
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var payload models.LoginErrorPayload
	err = json.Unmarshal(recorder.Body.Bytes(), &payload)
	require.NoError(t, err)
	assert.Equal(t, serviceerrors.CodeSessionAuthenticationRequired, payload.ErrorCode)
}

func Test_loginHandlerRejectsPasswordAuthenticationMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.lua")
	err := os.WriteFile(configPath, []byte("authType = \"password\"\n"), 0o600)
	require.NoError(t, err)
	manager, err := configs.NewLuaConfigManager(configPath)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/login", (&Api{LuaConfigManager: manager}).login)

	requestBody, err := json.Marshal(models.RequestPayload{
		Type:     "login",
		Email:    "user@example.com",
		Password: "password123",
		Token:    "123456",
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var payload models.LoginErrorPayload
	err = json.Unmarshal(recorder.Body.Bytes(), &payload)
	require.NoError(t, err)
	assert.Equal(t, serviceerrors.CodeSessionAuthenticationRequired, payload.ErrorCode)
}
