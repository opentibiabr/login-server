package configs

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
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

func TestGetHttpLoginConfigsTrustedProxies(t *testing.T) {
	t.Setenv(EnvLoginTrustedProxiesKey, "127.0.0.1, 10.0.0.0/8, ::1")

	configs := getHttpLoginConfigs()
	assert.Equal(t, []string{"127.0.0.1", "10.0.0.0/8", "::1"}, configs.TrustedProxies)
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

func TestGetLogFile(t *testing.T) {
	t.Run("defaults when unset", func(t *testing.T) {
		original, exists := os.LookupEnv(EnvLogFile)
		assert.NoError(t, os.Unsetenv(EnvLogFile))
		t.Cleanup(func() {
			if exists {
				assert.NoError(t, os.Setenv(EnvLogFile, original))
				return
			}
			assert.NoError(t, os.Unsetenv(EnvLogFile))
		})

		assert.Equal(t, "logs/login-server.txt", GetLogFile())
	})

	t.Run("uses configured path", func(t *testing.T) {
		t.Setenv(EnvLogFile, "custom.log")
		assert.Equal(t, "custom.log", GetLogFile())
	})

	t.Run("can be disabled with explicit empty value", func(t *testing.T) {
		t.Setenv(EnvLogFile, "")
		assert.Equal(t, "", GetLogFile())
	})
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

func TestGetAuthenticatorConfigs(t *testing.T) {
	t.Setenv(EnvAuthenticatorEncryptionKey, "encoded-key")

	assert.Equal(t, AuthenticatorConfigs{EncryptionKey: "encoded-key"}, getAuthenticatorConfigs())
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

func TestGetLoginFreePremium(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  bool
	}{
		{name: "unset means off", set: false, want: false},
		{name: "true", value: "true", set: true, want: true},
		{name: "one", value: "1", set: true, want: true},
		{name: "true with spaces and capitals", value: " True ", set: true, want: true},
		{name: "false", value: "false", set: true, want: false},
		{name: "empty", value: "", set: true, want: false},
		{name: "not a boolean means off", value: "yes please", set: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Setenv registers the restore of the original value, then an unset case really unsets it.
			t.Setenv(EnvLoginFreePremiumKey, tt.value)
			if !tt.set {
				os.Unsetenv(EnvLoginFreePremiumKey)
			}
			assert.Equal(t, tt.want, GetLoginFreePremium())
		})
	}
}

func TestGetLoginPasswordRehashIsOffUnlessEnabled(t *testing.T) {
	for value, want := range map[string]bool{"": false, "false": false, "yes": false, "true": true, " TRUE ": true, "1": true} {
		t.Setenv(EnvLoginPasswordRehashKey, value)
		assert.Equal(t, want, GetLoginPasswordRehash(), "%q", value)
	}
}

func TestGetLoginPasswordHashConcurrency(t *testing.T) {
	for value, want := range map[string]int{"": 2, "0": 2, "-1": 2, "many": 2, "65": 2, "1": 1, "8": 8, "64": 64} {
		t.Setenv(EnvLoginPasswordHashConcurrencyKey, value)
		assert.Equal(t, want, GetLoginPasswordHashConcurrency(), "%q", value)
	}
}
