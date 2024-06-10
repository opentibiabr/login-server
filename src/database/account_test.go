package database

import (
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
)

func TestAccount_GetGrpcSession(t *testing.T) {
	type fields struct {
		ID       uint32
		Email    string
		Password string
		PremDays uint32
		LastDay  uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   *login_proto_messages.Session
	}{{
		name: "Get session no prem days",
		fields: fields{
			PremDays: 0,
			Email:    "a@a.com",
			Password: "123456",
		},
		want: &login_proto_messages.Session{
			IsPremium:    false,
			PremiumUntil: 0,
			SessionKey:   "a@a.com\n123456",
		},
	}, {
		name: "Get session positive prem days",
		fields: fields{
			PremDays: 1,
			Email:    "a@a.com",
			Password: "123456",
		},
		want: &login_proto_messages.Session{
			IsPremium:    true,
			PremiumUntil: 86400,
			SessionKey:   "a@a.com\n123456",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{
				ID:       tt.fields.ID,
				Email:    tt.fields.Email,
				Password: tt.fields.Password,
				PremDays: tt.fields.PremDays,
				LastDay:  tt.fields.LastDay,
			}
			if tt.fields.PremDays > 0 {
				monkey.Patch(time.Now, func() time.Time {
					return time.Unix(0, 0)
				})
			}
			assert.Equal(t, tt.want, acc.GetGrpcSession())
		})
	}
}

func TestAccount_GetPremiumTime(t *testing.T) {
	type fields struct {
		ID       uint32
		Email    string
		Password string
		PremDays uint32
		LastDay  uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{{
		name:   "Premium time 0 returns 0",
		fields: fields{PremDays: 0},
		want:   0,
	}, {
		name:   "Remaining premium returns today + remaining seconds",
		fields: fields{PremDays: 1},
		want:   86400,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{
				ID:       tt.fields.ID,
				Email:    tt.fields.Email,
				Password: tt.fields.Password,
				PremDays: tt.fields.PremDays,
				LastDay:  tt.fields.LastDay,
			}
			if tt.fields.PremDays > 0 {
				monkey.Patch(time.Now, func() time.Time {
					return time.Unix(0, 0)
				})
			}
			assert.Equal(t, tt.want, acc.GetPremiumTime())
		})
	}
}
