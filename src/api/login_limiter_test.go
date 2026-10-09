package api

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const loginTestBurst = 12

type rateLimitLoginService struct {
	login_proto_messages.UnimplementedLoginServiceServer
}

func (*rateLimitLoginService) Login(_ context.Context, _ *login_proto_messages.LoginRequest) (*login_proto_messages.LoginResponse, error) {
	return &login_proto_messages.LoginResponse{
		Session:  &login_proto_messages.Session{SessionKey: "test-session"},
		PlayData: &login_proto_messages.PlayData{},
	}, nil
}

func newRateLimitLoginConnection(t *testing.T) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	login_proto_messages.RegisterLoginServiceServer(server, &rateLimitLoginService{})
	go func() { _ = server.Serve(listener) }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	connection, err := grpc.DialContext(ctx, "bufnet", grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		server.Stop()
		_ = listener.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = connection.Close()
		server.Stop()
		_ = listener.Close()
	})
	return connection
}

func newLimitedLoginRouter(t *testing.T) (*gin.Engine, *limiter.IPRateLimiter) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	corePath := t.TempDir()
	eventDirectory := filepath.Join(corePath, "json", "eventscheduler")
	if err := os.MkdirAll(eventDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventDirectory, "events.json"), []byte(`{"events":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(corePath, "config.lua")
	if err := os.WriteFile(configPath, []byte("authType = \"session\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configManager, err := configs.NewLuaConfigManager(configPath)
	if err != nil {
		t.Fatal(err)
	}

	// Refill one token per day in this fixture so burst boundaries remain
	// deterministic even when the test machine is busy. Production configuration
	// supplies the unchanged RATE_LIMITER_RATE and RATE_LIMITER_BURST values.
	ipLimiter := &limiter.IPRateLimiter{
		Visitors:      make(map[string]*limiter.Visitor),
		Mu:            &sync.RWMutex{},
		Configs:       configs.RateLimiter{Rate: 1.0 / 86400, Burst: loginTestBurst},
		KeyForRequest: loginRateLimitKey,
	}
	api := &Api{
		Router:           gin.New(),
		CorePath:         corePath,
		LuaConfigManager: configManager,
		GrpcConnection:   newRateLimitLoginConnection(t),
	}
	api.Router.Use(ipLimiter.Limit())
	api.initializeRoutes()
	return api.Router, ipLimiter
}

func limitedLoginRequest(router *gin.Engine, endpoint, ip, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	request.RemoteAddr = ip + ":43210"
	request.Header.Set("Content-Type", "application/json")
	// A claimed action header must never select a bucket.
	request.Header.Set("X-Login-Action", "cacheinfo")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertLimitedLoginStatus(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("expected HTTP %d, got %d: %s", status, response.Code, response.Body.String())
	}
	if status == http.StatusTooManyRequests && response.Body.String() != `{"error":"too many requests"}` {
		t.Fatalf("limiter response changed: %s", response.Body.String())
	}
}

func TestLoginRateLimitPublicBootstrapDoesNotConsumeAuthenticationBurst(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	ip := "203.0.113.9"
	// Three clients sharing a NAT each perform four public bootstrap requests.
	actions := []string{"cacheinfo", "eventschedule", "news", "newsviewed"}
	endpoints := []string{"/", "/login", "/login.php"}
	for client, endpoint := range endpoints {
		for _, action := range actions {
			t.Logf("client %d action %s", client, action)
			assertLimitedLoginStatus(t, limitedLoginRequest(router, endpoint, ip, `{"type":"`+action+`"}`), http.StatusOK)
		}
	}
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login.php", ip, `{"type":"news"}`), http.StatusTooManyRequests)
	response := limitedLoginRequest(router, "/login", ip, `{"type":"login","email":"player@example.invalid","password":"secret","devicecookie":"cookie"}`)
	assertLimitedLoginStatus(t, response, http.StatusOK)
	var payload models.ResponsePayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Session.SessionKey != "test-session" || payload.LoginEmail != "player@example.invalid" || payload.DeviceCookie != "cookie" {
		t.Fatalf("authentication did not preserve parsed payload: %+v", payload)
	}
}

func TestLoginRateLimitAuthenticationBurstRemainsBoundedAndIndependent(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	ip := "203.0.113.9"
	endpoints := []string{"/", "/login", "/login.php"}
	for attempt := 0; attempt < loginTestBurst; attempt++ {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, endpoints[attempt%len(endpoints)], ip, `{"type":"login"}`), http.StatusOK)
	}
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"login"}`), http.StatusTooManyRequests)
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"cacheinfo"}`), http.StatusOK)
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", "203.0.113.10", `{"type":"login"}`), http.StatusOK)
}

func TestLoginRateLimitAllPublicActionsShareOneBoundedBucket(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	ip := "203.0.113.9"
	// No database is provided here: the real boosted handler returns its existing
	// named error while admission must still use the public bucket.
	actions := []string{"cacheinfo", "news", "newsviewed", "eventschedule", "boostedcreature"}
	for attempt := 0; attempt < loginTestBurst; attempt++ {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"`+actions[attempt%len(actions)]+`"}`), http.StatusOK)
	}
	for _, action := range actions {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"`+action+`"}`), http.StatusTooManyRequests)
	}
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"login"}`), http.StatusOK)
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", "203.0.113.10", `{"type":"news"}`), http.StatusOK)
}

func TestLoginRateLimitMalformedAndUnknownRequestsConsumeAuthenticationBurst(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "malformed JSON", body: `{"type":"cacheinfo"`, status: http.StatusBadRequest},
		{name: "invalid public payload field", body: `{"type":"cacheinfo","clienttype":"wrong"}`, status: http.StatusBadRequest},
		{name: "unknown action", body: `{"type":"unknown"}`, status: http.StatusOK},
		{name: "absent action", body: `{}`, status: http.StatusOK},
		{name: "oversized payload", body: `{"type":"news","email":"` + strings.Repeat("a", maxLoginRequestBytes) + `"}`, status: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _ := newLimitedLoginRouter(t)
			ip := "203.0.113.9"
			for attempt := 0; attempt < loginTestBurst; attempt++ {
				assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, test.body), test.status)
			}
			assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, test.body), http.StatusTooManyRequests)
			assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"login"}`), http.StatusTooManyRequests)
			assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"cacheinfo"}`), http.StatusOK)
		})
	}
}

func TestLoginRateLimitDoesNotClassifyUnrelatedRoutesAsPublic(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	ip := "203.0.113.9"
	for attempt := 0; attempt < loginTestBurst; attempt++ {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, "/crash-report", ip, `{"type":"news"}`), http.StatusNoContent)
	}
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"login"}`), http.StatusTooManyRequests)
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"news"}`), http.StatusOK)
}

type countedLoginBody struct {
	reader io.Reader
	reads  int
}

func (body *countedLoginBody) Read(buffer []byte) (int, error) {
	body.reads++
	return body.reader.Read(buffer)
}

func (body *countedLoginBody) Close() error { return nil }

func TestLoginRateLimitParsesRequestOnceAndPreservesHandlerErrors(t *testing.T) {
	for _, body := range []string{`{"type":"cacheinfo"}`, `{"type":"login"}`, `{"type":"login"`, `{"type":"news","clienttype":"wrong"}`} {
		t.Run(body, func(t *testing.T) {
			router, _ := newLimitedLoginRouter(t)
			request := httptest.NewRequest(http.MethodPost, "/login", nil)
			countedBody := &countedLoginBody{reader: strings.NewReader(body)}
			request.Body = countedBody
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			baseline := gin.New()
			baseline.POST("/login", func(c *gin.Context) {
				var payload models.RequestPayload
				if err := c.ShouldBindJSON(&payload); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.Status(http.StatusOK)
			})
			baselineResponse := limitedLoginRequest(baseline, "/login", "203.0.113.9", body)
			assertLimitedLoginStatus(t, response, baselineResponse.Code)
			if response.Code == http.StatusBadRequest && response.Body.String() != baselineResponse.Body.String() {
				t.Fatalf("binding error changed: got %s, want %s", response.Body.String(), baselineResponse.Body.String())
			}
			// Invalid/incomplete JSON needs an EOF read; a second decoder would
			// perform one additional EOF read in either case.
			expectedReads := 1
			if body == `{"type":"login"` {
				expectedReads = 2
			}
			if countedBody.reads != expectedReads {
				t.Fatalf("body read %d times; expected one decoder (%d reads)", countedBody.reads, expectedReads)
			}
		})
	}
}

func TestLoginRateLimitUnknownActionKeepsNamedError(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	response := limitedLoginRequest(router, "/login.php", "203.0.113.9", `{"type":"unknown"}`)
	var payload models.LoginErrorPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ErrorCode != serviceerrors.CodeUnsupportedRequestType {
		t.Fatalf("unsupported request error changed: %+v", payload)
	}
}

func TestLoginRateLimitRetainsConfiguredRateAndBurstForBothBuckets(t *testing.T) {
	router, ipLimiter := newLimitedLoginRouter(t)
	ipLimiter.Configs.Rate = 2
	for _, action := range []string{"login", "news"} {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", "203.0.113.9", `{"type":"`+action+`"}`), http.StatusOK)
	}
	if len(ipLimiter.Visitors) != 2 {
		t.Fatalf("expected two request classes, got %d buckets", len(ipLimiter.Visitors))
	}
	for _, visitor := range ipLimiter.Visitors {
		if visitor.Limiter.Limit() != 2 || visitor.Limiter.Burst() != loginTestBurst {
			t.Fatal("request class changed the configured token rate or burst")
		}
	}
}

func TestLoginRateLimitUsesExactlyThePayloadDispatched(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	ip := "203.0.113.9"
	// Encoding/json resolves duplicate/case-insensitive type fields. Admission
	// must use that same result rather than independently scanning the raw body.
	for attempt := 0; attempt < loginTestBurst; attempt++ {
		assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"news","type":"login"}`), http.StatusOK)
	}
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"news","type":"login"}`), http.StatusTooManyRequests)
	assertLimitedLoginStatus(t, limitedLoginRequest(router, "/login", ip, `{"type":"login","TYPE":"news"}`), http.StatusOK)
}
