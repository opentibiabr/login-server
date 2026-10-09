package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

func TestLoginRateLimitPreservesDeployedSecondFactorAndTransportRules(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name      string
		origin    string
		ip        string
		forwarded string
		secure    bool
	}{
		{name: "direct HTTPS", origin: "https://login.example", ip: "203.0.113.9", secure: true},
		{name: "plain HTTP", origin: "http://login.example", ip: "203.0.113.9"},
		{name: "forwarded HTTPS", origin: "http://login.example", ip: "10.2.3.4", forwarded: "https", secure: true},
		{name: "deployed first forwarded value", origin: "http://login.example", ip: "10.2.3.4", forwarded: "https, http", secure: true},
	} {
		for _, endpoint := range []string{"/", "/login", "/login.php"} {
			t.Run(test.name+endpoint, func(t *testing.T) {
				connection, service := newInMemoryLoginClient(t, &login_proto_messages.LoginResponse{
					Session:                &login_proto_messages.Session{SessionKey: "opaque-session"},
					PlayData:               &login_proto_messages.PlayData{},
					TrustedDeviceToken:     "rotated-device",
					TrustedDeviceExpiresAt: 123456,
				})
				configPath := filepath.Join(t.TempDir(), "config.lua")
				if err := os.WriteFile(configPath, []byte("authType = \"session\"\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				config, err := configs.NewLuaConfigManager(configPath)
				if err != nil {
					t.Fatal(err)
				}
				api := &Api{Router: gin.New(), GrpcConnection: connection, LuaConfigManager: config}
				ipLimiter := &limiter.IPRateLimiter{
					Visitors:      make(map[string]*limiter.Visitor),
					Mu:            &sync.RWMutex{},
					Configs:       configs.RateLimiter{Rate: 1.0 / 86400, Burst: loginTestBurst},
					KeyForRequest: loginRateLimitKey,
				}
				api.Router.Use(ipLimiter.Limit())
				api.initializeRoutes()
				for attempt := 0; attempt < loginTestBurst; attempt++ {
					assertLimitedLoginStatus(t, limitedLoginRequest(api.Router, endpoint, test.ip, `{"type":"news"}`), http.StatusOK)
				}
				body, err := json.Marshal(models.RequestPayload{
					Type:               "login",
					Email:              "player@example.invalid",
					Password:           "password",
					Token:              "123456",
					TrustedDeviceToken: "trusted-device",
					TrustDevice:        true,
					DeviceName:         "OTClient",
				})
				if err != nil {
					t.Fatal(err)
				}
				request := httptest.NewRequest(http.MethodPost, test.origin+endpoint, bytes.NewReader(body))
				request.RemoteAddr = test.ip + ":43210"
				if test.forwarded != "" {
					request.Header.Set("X-Forwarded-Proto", test.forwarded)
				}
				response := httptest.NewRecorder()
				api.Router.ServeHTTP(response, request)
				assertLimitedLoginStatus(t, response, http.StatusOK)
				if service.request == nil || service.request.GetEmail() != "player@example.invalid" || service.request.GetPassword() != "password" || service.request.GetToken() != "123456" {
					t.Fatalf("parsed second-factor credentials did not reach gRPC: %+v", service.request)
				}
				var payload models.ResponsePayload
				if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if payload.Session.SessionKey != "opaque-session" {
					t.Fatal("session response was changed by quota admission")
				}
				if test.secure {
					if service.request.GetTrustedDeviceToken() != "trusted-device" || !service.request.GetTrustDevice() || service.request.GetDeviceName() != "OTClient" || payload.TrustedDeviceToken != "rotated-device" || payload.TrustedDeviceExpiresAt != 123456 {
						t.Fatal("secure trusted-device request or response metadata was lost")
					}
				} else if service.request.GetTrustedDeviceToken() != "" || service.request.GetTrustDevice() || service.request.GetDeviceName() != "" || payload.TrustedDeviceToken != "" || payload.TrustedDeviceExpiresAt != 0 {
					t.Fatal("untrusted transport exposed trusted-device metadata")
				}
			})
		}
	}
}
