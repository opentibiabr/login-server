package tests

import (
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/tests/testlib"
	"testing"
)

func TestPlayerToCharacterPayload(t *testing.T) {
	a := testlib.Assert{T: *t}

	defaultString := "default_string"
	defaultNumber := 12

	expectedCharacterPayload := login.CharacterPayload{
		WorldID: 0,
		CharacterInfo: login.CharacterInfo{
			Name:     "default_string",
			Level:    12,
			Vocation: "Knight Dawnport",
			IsMale:   false,
		},
		Outfit: login.Outfit{
			OutfitID:    12,
			HeadColor:   12,
			TorsoColor:  12,
			LegsColor:   12,
			DetailColor: 12,
			AddonsFlags: 12,
		},
	}

	player := database.Player{
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
	}

	character := player.ToCharacterPayload()

	a.Equals(expectedCharacterPayload, character)
}
