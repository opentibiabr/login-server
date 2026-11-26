package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/opentibiabr/login-server/src/passwords"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const EnvLoginIpKey = "LOGIN_IP"
const EnvLoginHttpPortKey = "LOGIN_HTTP_PORT"
const EnvLoginGrpcPortKey = "LOGIN_GRPC_PORT"
const EnvLoginTrustedProxiesKey = "LOGIN_TRUSTED_PROXIES"
const EnvLoginFreePremiumKey = "LOGIN_FREE_PREMIUM"
const EnvLoginPasswordRehashKey = "LOGIN_PASSWORD_REHASH"
const EnvLoginPasswordChangeMessageKey = "LOGIN_PASSWORD_CHANGE_MESSAGE"
const EnvLoginPasswordHashConcurrencyKey = "LOGIN_PASSWORD_HASH_CONCURRENCY"

const maxPasswordHashConcurrency = 64

const EnvRateLimiterBurstKey = "RATE_LIMITER_BURST"
const EnvRateLimiterRateKey = "RATE_LIMITER_RATE"
const EnvAuthenticatorEncryptionKey = "AUTHENTICATOR_ENCRYPTION_KEY"

type LoginServerConfigs struct {
	Http          HttpLoginConfigs
	Grpc          GrpcLoginConfigs
	RateLimiter   RateLimiter
	Authenticator AuthenticatorConfigs
	Config
}

type AuthenticatorConfigs struct {
	EncryptionKey string
	Config
}

type HttpLoginConfigs struct {
	Ip             string
	Port           int
	TrustedProxies []string
	Config
}

type GrpcLoginConfigs struct {
	Ip   string
	Port int
	Config
}

type RateLimiter struct {
	Burst int
	Rate  rate.Limit
	Config
}

func (loginServerConfigs *LoginServerConfigs) Format() string {
	return fmt.Sprintf(
		"OTBR Login Server running!!! http: %s | gRPC: %s | %s",
		loginServerConfigs.Http.Format(),
		loginServerConfigs.Grpc.Format(),
		loginServerConfigs.RateLimiter.Format(),
	)
}
func GetLoginServerConfigs() LoginServerConfigs {
	return LoginServerConfigs{
		Http:          getHttpLoginConfigs(),
		Grpc:          getGrpcLoginConfigs(),
		RateLimiter:   GetRateLimiterConfigs(),
		Authenticator: getAuthenticatorConfigs(),
	}
}

// GetLoginFreePremium reports whether accounts without paid time are shown as free premium accounts in the client
// (LOGIN_FREE_PREMIUM=true, matching freePremium in the game server config). Unset or invalid means false.
func GetLoginFreePremium() bool {
	enabled, err := strconv.ParseBool(strings.TrimSpace(GetEnvStr(EnvLoginFreePremiumKey, "")))
	return err == nil && enabled
}

// GetLoginPasswordRehash reports whether a successful login replaces an older accounts.password value with Argon2id
// (LOGIN_PASSWORD_REHASH=true). Unset or invalid means false: enable it only after the game server and the account
// website accept Argon2id.
func GetLoginPasswordRehash() bool {
	enabled, err := strconv.ParseBool(strings.TrimSpace(GetEnvStr(EnvLoginPasswordRehashKey, "")))
	return err == nil && enabled
}

// GetLoginPasswordHashConcurrency is how many Argon2id computations may run at once
// (LOGIN_PASSWORD_HASH_CONCURRENCY). Unset or out of range means the default.
func GetLoginPasswordHashConcurrency() int {
	value, err := strconv.Atoi(strings.TrimSpace(GetEnvStr(EnvLoginPasswordHashConcurrencyKey, "")))
	if err != nil || value < 1 || value > maxPasswordHashConcurrency {
		return passwords.DefaultConcurrency
	}
	return value
}

// GetLoginPasswordChangeMessage is the text shown to accounts that must change their password
// (LOGIN_PASSWORD_CHANGE_MESSAGE). Empty means the default message.
func GetLoginPasswordChangeMessage() string {
	return strings.TrimSpace(GetEnvStr(EnvLoginPasswordChangeMessageKey, ""))
}

func getAuthenticatorConfigs() AuthenticatorConfigs {
	return AuthenticatorConfigs{EncryptionKey: GetEnvStr(EnvAuthenticatorEncryptionKey)}
}

func (httpLoginConfigs *HttpLoginConfigs) Format() string {
	return fmt.Sprintf(
		"%s:%d",
		httpLoginConfigs.Ip,
		httpLoginConfigs.Port,
	)
}
func getHttpLoginConfigs() HttpLoginConfigs {
	return HttpLoginConfigs{
		Ip:             GetEnvStr(EnvLoginIpKey, ""),
		Port:           GetEnvInt(EnvLoginHttpPortKey, 80),
		TrustedProxies: getTrustedProxies(),
	}
}

func getTrustedProxies() []string {
	value := strings.TrimSpace(GetEnvStr(EnvLoginTrustedProxiesKey, ""))
	if value == "" {
		return nil
	}

	entries := strings.Split(value, ",")
	proxies := make([]string, 0, len(entries))
	for _, entry := range entries {
		if proxy := strings.TrimSpace(entry); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

func (grpcLoginConfigs *GrpcLoginConfigs) Format() string {
	return fmt.Sprintf(
		"%s:%d",
		grpcLoginConfigs.Ip,
		grpcLoginConfigs.Port,
	)
}
func getGrpcLoginConfigs() GrpcLoginConfigs {
	return GrpcLoginConfigs{
		Ip:   GetEnvStr(EnvLoginIpKey, ""),
		Port: GetEnvInt(EnvLoginGrpcPortKey, 9090),
	}
}

func (rateLimiterConfigs *RateLimiter) Format() string {
	return fmt.Sprintf(
		"rate limit: %.0f/%d",
		rateLimiterConfigs.Rate,
		rateLimiterConfigs.Burst,
	)
}
func GetRateLimiterConfigs() RateLimiter {
	return RateLimiter{
		Burst: GetEnvInt(EnvRateLimiterBurstKey, 5),
		Rate:  rate.Limit(GetEnvInt(EnvRateLimiterRateKey, 2)),
	}
}

const EnvLogLevel = "ENV_LOG_LEVEL"
const EnvLogFile = "ENV_LOG_FILE"

func GetLogLevel() logrus.Level {
	defaultLevel, _ := logrus.Level.MarshalText(logrus.InfoLevel)
	level, _ := logrus.ParseLevel(GetEnvStr(EnvLogLevel, string(defaultLevel)))
	return level
}

func GetLogFile() string {
	if value, exists := os.LookupEnv(EnvLogFile); exists {
		return value
	}

	return "logs/login-server.txt"
}
