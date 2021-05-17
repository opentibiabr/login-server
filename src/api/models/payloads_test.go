package models

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"reflect"
	"testing"
)

func TestBuildWorldsMessage(t *testing.T) {
	tests := []struct {
		name string
		want []*login_proto_messages.World
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildWorldsMessage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildWorldsMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadCharactersFromMessage(t *testing.T) {
	type args struct {
		charactersMsg []*login_proto_messages.Character
	}
	tests := []struct {
		name string
		args args
		want []CharacterPayload
	}{
		{
			name: "one_character_male",
			args: args{[]*login_proto_messages.Character{createCharacterMessage(1)}},
			want: []CharacterPayload{{
				WorldID: defaultNumber,
				CharacterInfo: CharacterInfo{
					Name:     defaultString,
					Level:    defaultNumber,
					Vocation: defaultString,
					IsMale:   true,
				},
				Outfit: Outfit{
					OutfitID:    defaultNumber,
					HeadColor:   defaultNumber,
					TorsoColor:  defaultNumber,
					LegsColor:   defaultNumber,
					DetailColor: defaultNumber,
					AddonsFlags: defaultNumber,
				},
			}},
		},
		{
			name: "one_character_female",
			args: args{[]*login_proto_messages.Character{createCharacterMessage(2)}},
			want: []CharacterPayload{{
				WorldID: defaultNumber,
				CharacterInfo: CharacterInfo{
					Name:     defaultString,
					Level:    defaultNumber,
					Vocation: defaultString,
					IsMale:   false,
				},
				Outfit: Outfit{
					OutfitID:    defaultNumber,
					HeadColor:   defaultNumber,
					TorsoColor:  defaultNumber,
					LegsColor:   defaultNumber,
					DetailColor: defaultNumber,
					AddonsFlags: defaultNumber,
				},
			}},
		},
		{
			name: "two_characters_male_female",
			args: args{[]*login_proto_messages.Character{
				createCharacterMessage(2),
				createCharacterMessage(1),
			}},
			want: []CharacterPayload{
				{
					WorldID: defaultNumber,
					CharacterInfo: CharacterInfo{
						Name:     defaultString,
						Level:    defaultNumber,
						Vocation: defaultString,
						IsMale:   false,
					},
					Outfit: Outfit{
						OutfitID:    defaultNumber,
						HeadColor:   defaultNumber,
						TorsoColor:  defaultNumber,
						LegsColor:   defaultNumber,
						DetailColor: defaultNumber,
						AddonsFlags: defaultNumber,
					},
				},
				{
					WorldID: defaultNumber,
					CharacterInfo: CharacterInfo{
						Name:     defaultString,
						Level:    defaultNumber,
						Vocation: defaultString,
						IsMale:   true,
					},
					Outfit: Outfit{
						OutfitID:    defaultNumber,
						HeadColor:   defaultNumber,
						TorsoColor:  defaultNumber,
						LegsColor:   defaultNumber,
						DetailColor: defaultNumber,
						AddonsFlags: defaultNumber,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadCharactersFromMessage(tt.args.charactersMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadCharactersFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadSessionFromMessage(t *testing.T) {
	type args struct {
		sessionMsg *login_proto_messages.Session
	}
	tests := []struct {
		name string
		args args
		want Session
	}{
		{
			name: "is_not_premium",
			args: args{createSessionMessage(false)},
			want: Session{
				IsPremium:     false,
				PremiumUntil:  defaultNumber,
				SessionKey:    defaultString,
				LastLoginTime: defaultNumber,
			},
		},
		{
			name: "is_premium",
			args: args{createSessionMessage(true)},
			want: Session{
				IsPremium:     true,
				PremiumUntil:  defaultNumber,
				SessionKey:    defaultString,
				LastLoginTime: defaultNumber,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadSessionFromMessage(tt.args.sessionMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadSessionFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadWorldsFromMessage(t *testing.T) {
	type args struct {
		worldsMsg []*login_proto_messages.World
	}
	tests := []struct {
		name string
		args args
		want []World
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadWorldsFromMessage(tt.args.worldsMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadWorldsFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildWorldMessage(t *testing.T) {
	type args struct {
		gameConfigs configs.GameServerConfigs
	}
	tests := []struct {
		name string
		args args
		want *login_proto_messages.World
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildWorldMessage(tt.args.gameConfigs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildWorldMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
