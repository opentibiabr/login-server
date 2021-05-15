package configs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const EnvLoginIpKey = "LOGIN_IP"
const EnvLoginHttpPortKey = "LOGIN_HTTP_PORT"
const EnvLoginTcpPortKey = "LOGIN_TCP_PORT"

const EnvRateLimiterBurstKey = "RATE_LIMITER_BURST"
const EnvRateLimiterRateKey = "RATE_LIMITER_RATE"

type LoginServerConfigs struct {
	Http        HttpLoginConfigs
	Tcp         TcpLoginConfigs
	RateLimiter RateLimiter
	Config
}

type HttpLoginConfigs struct {
	Ip   string
	Port int
	Config
}

type TcpLoginConfigs struct {
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
		"OTBR Login Server running!!! http: %s | tcp: %s | %s",
		loginServerConfigs.Http.Format(),
		loginServerConfigs.Tcp.Format(),
		loginServerConfigs.RateLimiter.Format(),
	)
}
func GetLoginServerConfigs() LoginServerConfigs {
	return LoginServerConfigs{
		Http:        GetHttpLoginConfigs(),
		Tcp:         GetTcpLoginConfigs(),
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
func GetHttpLoginConfigs() HttpLoginConfigs {
	return HttpLoginConfigs{
		Ip:   GetEnvStr(EnvLoginIpKey, ""),
		Port: GetEnvInt(EnvLoginHttpPortKey, 80),
	}
}

func (tcpLoginConfigs *TcpLoginConfigs) Format() string {
	return fmt.Sprintf(
		"%s:%d",
		tcpLoginConfigs.Ip,
		tcpLoginConfigs.Port,
	)
}
func GetTcpLoginConfigs() TcpLoginConfigs {
	return TcpLoginConfigs{
		Ip:   GetEnvStr(EnvLoginIpKey, ""),
		Port: GetEnvInt(EnvLoginTcpPortKey, 7171),
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
