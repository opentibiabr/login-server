package models

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"reflect"
	"testing"
)

func TestLoadSessionFromMessage(t *testing.T) {
	type args struct {
		sessionMsg *login_proto_messages.Session
	}
	tests := []struct {
		name string
		args args
		want Session
	}{{
		"is_not_premium",
		args{createSessionMessage(false)},
		Session{
			IsPremium:     false,
			PremiumUntil:  defaultNumber,
			SessionKey:    defaultString,
			LastLoginTime: defaultNumber,
		}}, {
		"is_premium",
		args{createSessionMessage(true)},
		Session{
			IsPremium:     true,
			PremiumUntil:  defaultNumber,
			SessionKey:    defaultString,
			LastLoginTime: defaultNumber,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadSessionFromMessage(tt.args.sessionMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadSessionFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
