package grpc_login_server

import (
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
	"testing"
)

var defaultString = "default"
var defaultNumber = 5

func TestBuildGrpcLoginResponsePayload(t *testing.T) {
	type args struct {
		session *login_proto_messages.Session
		players database.Players
	}
	tests := []struct {
		name string
		args args
		want *login_proto_messages.LoginResponse
	}{{
		name: "Session last login  gets player one if smaller",
		args: args{
			session: &login_proto_messages.Session{
				IsPremium:    false,
				PremiumUntil: 0,
				SessionKey:   "",
				LastLogin:    2,
			},
			players: database.Players{
				AccountID: defaultNumber,
				Players: []database.Player{
					{
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
				},
			},
		},
		want: &login_proto_messages.LoginResponse{
			Session: &login_proto_messages.Session{
				IsPremium:    false,
				PremiumUntil: 0,
				SessionKey:   "",
				LastLogin:    uint32(defaultNumber),
			},
			PlayData: &login_proto_messages.PlayData{
				Characters: []*login_proto_messages.Character{{
					WorldId: 0,
					Info: &login_proto_messages.CharacterInfo{
						Name:      defaultString,
						Vocation:  "Master Sorcerer",
						Level:     uint32(defaultNumber),
						LastLogin: uint32(defaultNumber),
						Sex:       uint32(defaultNumber),
					},
					Outfit: &login_proto_messages.CharacterOutfit{
						LookType: uint32(defaultNumber),
						LookHead: uint32(defaultNumber),
						LookBody: uint32(defaultNumber),
						LookLegs: uint32(defaultNumber),
						LookFeet: uint32(defaultNumber),
						Addons:   uint32(defaultNumber),
					},
				}},
				Worlds: []*login_proto_messages.World{{
					Id:                         0,
					ExternalPort:               7172,
					ExternalPortProtected:      7172,
					ExternalPortUnprotected:    7172,
					Name:                       "Canary",
					ExternalAddress:            "127.0.0.1",
					ExternalAddressProtected:   "127.0.0.1",
					ExternalAddressUnprotected: "127.0.0.1",
					Location:                   "BRA",
				}},
			}},
	}, {
		name: "Session last login keeps the same if bigger",
		args: args{
			session: &login_proto_messages.Session{
				IsPremium:    false,
				PremiumUntil: 0,
				SessionKey:   "",
				LastLogin:    10,
			},
			players: database.Players{
				AccountID: defaultNumber,
				Players: []database.Player{
					{
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
				},
			},
		},
		want: &login_proto_messages.LoginResponse{
			Session: &login_proto_messages.Session{
				IsPremium:    false,
				PremiumUntil: 0,
				SessionKey:   "",
				LastLogin:    uint32(10),
			},
			PlayData: &login_proto_messages.PlayData{
				Characters: []*login_proto_messages.Character{{
					WorldId: 0,
					Info: &login_proto_messages.CharacterInfo{
						Name:      defaultString,
						Vocation:  "Master Sorcerer",
						Level:     uint32(defaultNumber),
						LastLogin: uint32(defaultNumber),
						Sex:       uint32(defaultNumber),
					},
					Outfit: &login_proto_messages.CharacterOutfit{
						LookType: uint32(defaultNumber),
						LookHead: uint32(defaultNumber),
						LookBody: uint32(defaultNumber),
						LookLegs: uint32(defaultNumber),
						LookFeet: uint32(defaultNumber),
						Addons:   uint32(defaultNumber),
					},
				}},
				Worlds: []*login_proto_messages.World{{
					Id:                         1,
					ExternalPort:               7172,
					ExternalPortProtected:      7172,
					ExternalPortUnprotected:    7172,
					Name:                       "Canary",
					ExternalAddress:            "127.0.0.1",
					ExternalAddressProtected:   "127.0.0.1",
					ExternalAddressUnprotected: "127.0.0.1",
					Location:                   "BRA",
				}},
			},
		}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BuildGrpcLoginResponsePayload(tt.args.session, tt.args.players))
		})
	}
}
