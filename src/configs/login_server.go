package configs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const EnvLoginIpKey = "LOGIN_IP"
const EnvLoginHttpPortKey = "LOGIN_HTTP_PORT"
const EnvLoginGrpcPortKey = "LOGIN_GRPC_PORT"

const EnvRateLimiterBurstKey = "RATE_LIMITER_BURST"
const EnvRateLimiterRateKey = "RATE_LIMITER_RATE"

type LoginServerConfigs struct {
	Http        HttpLoginConfigs
	Grpc        GrpcLoginConfigs
	RateLimiter RateLimiter
	Config
}

type HttpLoginConfigs struct {
	Ip   string
	Port int
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
		Http:        getHttpLoginConfigs(),
		Grpc:        getGrpcLoginConfigs(),
		RateLimiter: GetRateLimiterConfigs(),
	}
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
		Ip:   GetEnvStr(EnvLoginIpKey, ""),
		Port: GetEnvInt(EnvLoginHttpPortKey, 80),
	}
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

func GetLogLevel() logrus.Level {
	defaultLevel, _ := logrus.Level.MarshalText(logrus.InfoLevel)
	level, _ := logrus.ParseLevel(GetEnvStr(EnvLogLevel, string(defaultLevel)))
	return level
}
