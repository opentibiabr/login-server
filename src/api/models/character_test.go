package models

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"reflect"
	"testing"
)

func TestLoadCharactersFromMessage(t *testing.T) {
	type args struct {
		charactersMsg []*login_proto_messages.Character
	}
	tests := []struct {
		name string
		args args
		want []CharacterPayload
	}{{
		"one_character_male",
		args{[]*login_proto_messages.Character{createCharacterMessage(1)}},
		[]CharacterPayload{{
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
	}, {
		"one_character_female",
		args{[]*login_proto_messages.Character{createCharacterMessage(2)}},
		[]CharacterPayload{{
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
	}, {
		"two_characters_male_female",
		args{[]*login_proto_messages.Character{
			createCharacterMessage(2),
			createCharacterMessage(1),
		}},
		[]CharacterPayload{{
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
		}, {
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
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadCharactersFromMessage(tt.args.charactersMsg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadCharactersFromMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
