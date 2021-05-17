package database

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
	"testing"
)

var defaultString = "default"
var defaultNumber = 2

func TestPlayer_ToCharacterMessage(t *testing.T) {
	type fields struct {
		Name       string
		Level      int
		Sex        int
		Vocation   int
		LookType   int
		LookHead   int
		LookBody   int
		LookLegs   int
		LookFeet   int
		LookAddons int
		LastLogin  int
	}
	tests := []struct {
		name   string
		fields fields
		want   *login_proto_messages.Character
	}{{
		name: "Creates gRPC message from db player",
		fields: fields{
			Name:       defaultString,
			Level:      defaultNumber,
			Sex:        defaultNumber,
			Vocation:   defaultNumber,
			LookType:   defaultNumber,
			LookHead:   defaultNumber,
			LookBody:   defaultNumber,
			LookLegs:   defaultNumber,
			LookFeet:   defaultNumber,
			LookAddons: defaultNumber,
			LastLogin:  defaultNumber,
		},
		want: &login_proto_messages.Character{
			WorldId: 0,
			Info: &login_proto_messages.CharacterInfo{
				Name:      defaultString,
				Level:     uint32(defaultNumber),
				Sex:       uint32(defaultNumber),
				LastLogin: uint32(defaultNumber),
				Vocation:  "Druid",
			},
			Outfit: &login_proto_messages.CharacterOutfit{
				LookType: uint32(defaultNumber),
				LookHead: uint32(defaultNumber),
				LookBody: uint32(defaultNumber),
				LookLegs: uint32(defaultNumber),
				LookFeet: uint32(defaultNumber),
				Addons:   uint32(defaultNumber),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{
				Name:       tt.fields.Name,
				Level:      tt.fields.Level,
				Sex:        tt.fields.Sex,
				Vocation:   tt.fields.Vocation,
				LookType:   tt.fields.LookType,
				LookHead:   tt.fields.LookHead,
				LookBody:   tt.fields.LookBody,
				LookLegs:   tt.fields.LookLegs,
				LookFeet:   tt.fields.LookFeet,
				LookAddons: tt.fields.LookAddons,
				LastLogin:  tt.fields.LastLogin,
			}
			assert.Equal(t, tt.want, player.ToCharacterMessage())
		})
	}
}
