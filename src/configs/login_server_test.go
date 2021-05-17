package configs

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"os"
	"testing"
)

func TestGetHttpLoginConfigs(t *testing.T) {
	tests := []struct {
		name string
		want HttpLoginConfigs
	}{{
		name: "Default Http Configs",
		want: HttpLoginConfigs{
			Ip:   "",
			Port: 80,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getHttpLoginConfigs())
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		want     logrus.Level
		envValue string
	}{{
		name: "LogLevel defaults to INFO",
		want: logrus.InfoLevel,
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.TraceLevel,
		envValue: "trace",
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.DebugLevel,
		envValue: "debug",
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.WarnLevel,
		envValue: "warn",
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.ErrorLevel,
		envValue: "error",
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.FatalLevel,
		envValue: "fatal",
	}, {
		name:     "LogLevel defaults to INFO",
		want:     logrus.PanicLevel,
		envValue: "panic",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				err := os.Setenv(EnvLogLevel, tt.envValue)
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, GetLogLevel())
			if tt.envValue != "" {
				err := os.Unsetenv(EnvLogLevel)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetLoginServerConfigs(t *testing.T) {
	tests := []struct {
		name string
		want LoginServerConfigs
	}{{
		name: "Default Login Configs",
		want: LoginServerConfigs{
			Http: HttpLoginConfigs{
				Ip:   "",
				Port: 80,
			},
			Tcp: TcpLoginConfigs{
				Ip:   "",
				Port: 7171,
			},
			RateLimiter: RateLimiter{
				Burst: 5,
				Rate:  rate.Limit(2),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetLoginServerConfigs())
		})
	}
}

func TestGetRateLimiterConfigs(t *testing.T) {
	tests := []struct {
		name string
		want RateLimiter
	}{{
		name: "Default Rate Limiter Configs",
		want: RateLimiter{
			Burst: 5,
			Rate:  2,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetRateLimiterConfigs())
		})
	}
}

func TestGetTcpLoginConfigs(t *testing.T) {
	tests := []struct {
		name string
		want TcpLoginConfigs
	}{{
		name: "Default Tcp Configs",
		want: TcpLoginConfigs{
			Ip:   "",
			Port: 7171,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getTcpLoginConfigs())
		})
	}
}

func TestHttpLoginConfigs_Format(t *testing.T) {
	type fields struct {
		Ip     string
		Port   int
		Config Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Test Http Configs format",
		fields: fields{
			Ip:   "0.0.0.0",
			Port: 8080,
		},
		want: "0.0.0.0:8080",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpLoginConfigs := &HttpLoginConfigs{
				Ip:     tt.fields.Ip,
				Port:   tt.fields.Port,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, httpLoginConfigs.Format())
		})
	}
}

func TestLoginServerConfigs_Format(t *testing.T) {
	type fields struct {
		Http        HttpLoginConfigs
		Tcp         TcpLoginConfigs
		RateLimiter RateLimiter
		Config      Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Test Http Configs format",
		fields: fields{
			Http: HttpLoginConfigs{
				Ip:   "0.0.0.0",
				Port: 8080,
			},
			Tcp: TcpLoginConfigs{
				Ip:   "0.0.0.0",
				Port: 7171,
			},
			RateLimiter: RateLimiter{
				Rate:  10,
				Burst: 30,
			},
		},
		want: "OTBR Login Server running!!! http: 0.0.0.0:8080 |" +
			" tcp: 0.0.0.0:7171 | rate limit: 10/30",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginServerConfigs := &LoginServerConfigs{
				Http:        tt.fields.Http,
				Tcp:         tt.fields.Tcp,
				RateLimiter: tt.fields.RateLimiter,
				Config:      tt.fields.Config,
			}
			assert.Equal(t, tt.want, loginServerConfigs.Format())
		})
	}
}

func TestRateLimiter_Format(t *testing.T) {
	type fields struct {
		Burst  int
		Rate   rate.Limit
		Config Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Rate Limit Configs Format",
		fields: fields{
			Burst: 10,
			Rate:  7,
		},
		want: "rate limit: 7/10",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimiterConfigs := &RateLimiter{
				Burst:  tt.fields.Burst,
				Rate:   tt.fields.Rate,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, rateLimiterConfigs.Format())
		})
	}
}

func TestTcpLoginConfigs_Format(t *testing.T) {
	type fields struct {
		Ip     string
		Port   int
		Config Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Tcp Login Configs Format",
		fields: fields{
			Ip:   "1.1.1.1",
			Port: 53201,
		},
		want: "1.1.1.1:53201",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcpLoginConfigs := &TcpLoginConfigs{
				Ip:     tt.fields.Ip,
				Port:   tt.fields.Port,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, tcpLoginConfigs.Format())
		})
	}
}
