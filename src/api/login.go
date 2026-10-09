package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/api/limiter"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const temporaryErrorCode = 2
const temporaryErrorMessage = "Internal error. Please try again later or contact customer support if the problem persists."
const loginRequestContextKey = "login-server.request"
const maxLoginRequestBytes = 1 << 20

type boundLoginRequest struct {
	payload models.RequestPayload
	err     error
}

func bindLoginRequest(c *gin.Context) *boundLoginRequest {
	if request, exists := c.Get(loginRequestContextKey); exists {
		return request.(*boundLoginRequest)
	}

	request := &boundLoginRequest{}
	if c.Request.Body != nil {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLoginRequestBytes)
	}
	request.err = c.ShouldBindJSON(&request.payload)
	c.Set(loginRequestContextKey, request)
	return request
}

func loginRateLimitKey(ipLimiter *limiter.IPRateLimiter) func(*gin.Context, string) (string, bool) {
	return func(c *gin.Context, ip string) (string, bool) {
		if c.Request.Method != http.MethodPost {
			return ip, true
		}
		switch c.FullPath() {
		case "/", "/login", "/login.php":
		default:
			return ip, true
		}

		publicKey := "public:" + ip
		if !ipLimiter.CanAllowAny(ip, publicKey) {
			return ip, false
		}
		request := bindLoginRequest(c)
		if request.err == nil {
			switch request.payload.Type {
			case "cacheinfo", "news", "newsviewed", "eventschedule", "boostedcreature":
				return publicKey, true
			}
		}
		return ip, true
	}
}

func (_api *Api) login(c *gin.Context) {
	request := bindLoginRequest(c)
	if request.err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": request.err.Error()})
		return
	}
	payload := request.payload

	switch payload.Type {
	case "cacheinfo", "news", "newsviewed":
		c.JSON(http.StatusOK, buildTemporaryErrorPayload())
	case "eventschedule":
		database.HandleEventSchedule(c, _api.eventSchedulePath())
	case "boostedcreature":
		database.HandleBoostedCreature(c, _api.DB, &_api.BoostedCreatureID, &_api.BoostedBossID)
	case "login":
		if _api.hasIncompatibleAuthType() {
			writePublicError(c, serviceerrors.LoginService(
				serviceerrors.CodeSessionAuthenticationRequired,
				"SESSION_AUTHENTICATION_REQUIRED",
				fmt.Errorf("game server authType is unavailable or not configured as session"),
			))
			return
		}
		if _api.GrpcConnection == nil {
			writePublicError(c, serviceerrors.LoginService(
				serviceerrors.CodeLoginServiceUnavailable,
				"LOGIN_SERVICE_UNAVAILABLE",
				fmt.Errorf("grpc connection is nil"),
			))
			return
		}

		grpcClient := login_proto_messages.NewLoginServiceClient(_api.GrpcConnection)

		trustedDeviceToken := ""
		trustDevice := false
		deviceName := ""
		secureLoginRequest := _api.isSecureLoginRequest(c)
		if secureLoginRequest {
			trustedDeviceToken = payload.TrustedDeviceToken
			trustDevice = payload.TrustDevice
			deviceName = payload.DeviceName
		}

		res, err := grpcClient.Login(
			context.Background(),
			&login_proto_messages.LoginRequest{
				Email:              payload.Email,
				Password:           payload.Password,
				Token:              payload.Token,
				TrustedDeviceToken: trustedDeviceToken,
				TrustDevice:        trustDevice,
				DeviceName:         deviceName,
			},
		)

		if err != nil {
			writePublicError(c, serviceerrors.LoginService(
				serviceerrors.CodeLoginServiceUnavailable,
				"LOGIN_SERVICE_UNAVAILABLE",
				err,
			))
			return
		}

		if res.GetError() != nil {
			c.JSON(http.StatusOK, buildErrorPayloadFromMessage(res))
			return
		}

		response := buildPayloadFromMessage(res, payload)
		if !secureLoginRequest {
			response.TrustedDeviceToken = ""
			response.TrustedDeviceExpiresAt = 0
		}
		c.JSON(http.StatusOK, response)
	default:
		writePublicError(c, serviceerrors.LoginService(
			serviceerrors.CodeUnsupportedRequestType,
			"UNSUPPORTED_REQUEST_TYPE",
			fmt.Errorf("unsupported login request type %q", payload.Type),
		))
	}
}

func (api *Api) hasIncompatibleAuthType() bool {
	if api == nil || api.LuaConfigManager == nil {
		return true
	}
	return api.LuaConfigManager.GetString("authType") != "session"
}

func (api *Api) isSecureLoginRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	if c.Request.TLS != nil {
		return true
	}
	if api == nil || !api.trustedProxies.containsRemoteAddress(c.Request.RemoteAddr) {
		return false
	}
	forwardedValues := c.Request.Header.Values("X-Forwarded-Proto")
	if len(forwardedValues) != 1 || strings.Contains(forwardedValues[0], ",") {
		return false
	}
	forwardedProto := strings.TrimSpace(forwardedValues[0])
	return strings.EqualFold(forwardedProto, "https")
}

func buildPayloadFromMessage(msg *login_proto_messages.LoginResponse, request models.RequestPayload) models.ResponsePayload {
	return models.ResponsePayload{
		DeviceCookie: request.DeviceCookie,
		LoginEmail:   request.Email,
		PlayData: models.PlayData{
			Worlds:     models.LoadWorldsFromMessage(msg.PlayData.Worlds),
			Characters: models.LoadCharactersFromMessage(msg.PlayData.Characters),
		},
		Session:                models.LoadSessionFromMessage(msg.GetSession()),
		TrustedDeviceToken:     msg.GetTrustedDeviceToken(),
		TrustedDeviceExpiresAt: msg.GetTrustedDeviceExpiresAt(),
	}
}

func buildErrorPayloadFromMessage(msg *login_proto_messages.LoginResponse) models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    int(msg.Error.Code),
		ErrorMessage: msg.Error.Message,
	}
}

func buildTemporaryErrorPayload() models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    temporaryErrorCode,
		ErrorMessage: temporaryErrorMessage,
	}
}

func buildErrorPayloadFromPublicError(err *serviceerrors.PublicError) models.LoginErrorPayload {
	return models.LoginErrorPayload{
		ErrorCode:    err.Code,
		ErrorMessage: serviceerrors.MessageWithHint(err),
	}
}

func writePublicError(c *gin.Context, err *serviceerrors.PublicError) {
	if err.Cause != nil {
		logger.Error(err)
	}
	c.JSON(http.StatusOK, buildErrorPayloadFromPublicError(err))
}
