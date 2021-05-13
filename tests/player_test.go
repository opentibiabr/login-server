package tests

import (
	"login-server/src/api/login"
	"login-server/src/database"
	"login-server/tests/testlib"
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
		defaultString,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
		defaultNumber,
	}

	character := player.ToCharacterPayload()

	a.Equals(expectedCharacterPayload, character)
}
