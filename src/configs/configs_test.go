package configs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log"
	"os"
	"testing"
)

const defaultString = "default_string"
const defaultNumberStr = "8080"
const defaultNumber = 8080

func setGameConfigs() {
	setEnvKeys(
		[]string{
			EnvLoginIpKey,
			EnvServerIpKey,
			EnvServerNameKey,
			EnvServerLocationKey,
			EnvDBHostKey,
			EnvDBNameKey,
			EnvDBUserKey,
			EnvDBPassKey,
		},
		[]string{
			EnvServerPortKey,
			EnvLoginHttpPortKey,
			EnvLoginGrpcPortKey,
			EnvRateLimiterBurstKey,
			EnvRateLimiterRateKey,
			EnvDBPortKey,
		},
	)
}

func unsetGameConfigs() {
	unsetEnvKeys(
		[]string{
			EnvLoginIpKey,
			EnvServerIpKey,
			EnvServerNameKey,
			EnvServerLocationKey,
			EnvDBHostKey,
			EnvDBNameKey,
			EnvDBUserKey,
			EnvDBPassKey,
			EnvServerPortKey,
			EnvLoginHttpPortKey,
			EnvLoginGrpcPortKey,
			EnvRateLimiterBurstKey,
			EnvRateLimiterRateKey,
			EnvDBPortKey,
		},
	)
}

func setEnvKeys(strings []string, integers []string) {
	for _, key := range strings {
		err := os.Setenv(key, defaultString)
		if err != nil {
			log.Printf("Cannot set key %s", key)
		}
	}
	for _, key := range integers {
		err := os.Setenv(key, defaultNumberStr)
		if err != nil {
			log.Printf("Cannot set key %s", key)
		}
	}
}

func unsetEnvKeys(keys []string) {
	for _, key := range keys {
		err := os.Unsetenv(key)
		if err != nil {
			log.Printf("Cannot unset key %s", key)
		}
	}
}

func TestGetEnvInt(t *testing.T) {
	type args struct {
		key          string
		defaultValue []int
	}
	tests := []struct {
		name   string
		args   args
		want   int
		envKey int
	}{{
		name: "no key set, uses default",
		args: args{key: "my_key", defaultValue: []int{10}},
		want: 10,
	}, {
		name: "no key set, no default",
		args: args{key: "my_key", defaultValue: []int{}},
		want: 0,
	}, {
		name:   "key set, does not use default",
		args:   args{key: "my_key", defaultValue: []int{10}},
		want:   8,
		envKey: 8,
	}, {
		name:   "key set, works without default",
		args:   args{key: "my_key", defaultValue: []int{}},
		want:   8,
		envKey: 8,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != 0 {
				err := os.Setenv(tt.args.key, fmt.Sprintf("%d", tt.envKey))
				assert.Nil(t, err)
			}
			got := GetEnvInt(tt.args.key, tt.args.defaultValue...)
			assert.Equal(t, got, tt.want)
			if tt.envKey != 0 {
				err := os.Unsetenv(tt.args.key)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetEnvStr(t *testing.T) {
	type args struct {
		key          string
		defaultValue []string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		envKey string
	}{{
		name: "no key set, uses default",
		args: args{key: "my_key", defaultValue: []string{"default"}},
		want: "default",
	}, {
		name: "no key set, no default",
		args: args{key: "my_key", defaultValue: []string{}},
		want: "",
	}, {
		name:   "key set, does not use default",
		args:   args{key: "my_key", defaultValue: []string{"default"}},
		want:   "set",
		envKey: "set",
	}, {
		name:   "key set, works without default",
		args:   args{key: "my_key", defaultValue: []string{}},
		want:   "set",
		envKey: "set",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				err := os.Setenv(tt.args.key, tt.envKey)
				assert.Nil(t, err)
			}
			got := GetEnvStr(tt.args.key, tt.args.defaultValue...)
			assert.Equal(t, got, tt.want)
			if tt.envKey != "" {
				err := os.Unsetenv(tt.args.key)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetGlobalConfigs(t *testing.T) {
	tests := []struct {
		name   string
		want   GlobalConfigs
		setEnv bool
	}{{
		name: "default global configs",
		want: GlobalConfigs{
			LoginServerConfigs: LoginServerConfigs{
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
			GameServerConfigs: GameServerConfigs{
				IP:       "127.0.0.1",
				Name:     "Canary",
				Port:     7172,
				Location: "BRA",
			},
			DBConfigs: DBConfigs{
				Host: "127.0.0.1",
				Name: "canary",
				Port: 3306,
				User: "canary",
				Pass: "canary",
			},
		},
	}, {
		name: "default global configs",
		want: GlobalConfigs{
			LoginServerConfigs: LoginServerConfigs{
				Http: HttpLoginConfigs{
					Ip:   defaultString,
					Port: defaultNumber,
				},
				Grpc: GrpcLoginConfigs{
					Ip:   defaultString,
					Port: defaultNumber,
				},
				RateLimiter: RateLimiter{
					Burst: defaultNumber,
					Rate:  rate.Limit(defaultNumber),
				},
			},
			GameServerConfigs: GameServerConfigs{
				IP:       defaultString,
				Name:     defaultString,
				Port:     8080,
				Location: defaultString,
			},
			DBConfigs: DBConfigs{
				Host: defaultString,
				Name: defaultString,
				Port: 8080,
				User: defaultString,
				Pass: defaultString,
			},
		},
		setEnv: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				setGameConfigs()
			}
			assert.Equal(t, tt.want, GetGlobalConfigs())
			if tt.setEnv {
				unsetGameConfigs()
			}
		})
	}
}
