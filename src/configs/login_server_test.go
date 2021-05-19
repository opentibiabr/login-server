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
			Grpc: GrpcLoginConfigs{
				Ip:   "",
				Port: 9090,
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

func TestGetGrpcLoginConfigs(t *testing.T) {
	tests := []struct {
		name string
		want GrpcLoginConfigs
	}{{
		name: "Default Grpc Configs",
		want: GrpcLoginConfigs{
			Ip:   "",
			Port: 9090,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getGrpcLoginConfigs())
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
		Grpc        GrpcLoginConfigs
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
			Grpc: GrpcLoginConfigs{
				Ip:   "0.0.0.0",
				Port: 9090,
			},
			RateLimiter: RateLimiter{
				Rate:  10,
				Burst: 30,
			},
		},
		want: "OTBR Login Server running!!! http: 0.0.0.0:8080 |" +
			" gRPC: 0.0.0.0:9090 | rate limit: 10/30",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginServerConfigs := &LoginServerConfigs{
				Http:        tt.fields.Http,
				Grpc:        tt.fields.Grpc,
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

func TestGrpcLoginConfigs_Format(t *testing.T) {
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
		name: "Grpc Login Configs Format",
		fields: fields{
			Ip:   "1.1.1.1",
			Port: 53201,
		},
		want: "1.1.1.1:53201",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcLoginConfigs := &GrpcLoginConfigs{
				Ip:     tt.fields.Ip,
				Port:   tt.fields.Port,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, grpcLoginConfigs.Format())
		})
	}
}
