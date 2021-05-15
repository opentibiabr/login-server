package tests

import (
	"github.com/opentibiabr/login-server/src/api/login"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlayerToCharacterPayload(t *testing.T) {
	defNumber := 12

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
		Level:      defNumber,
		Sex:        defNumber,
		Vocation:   defNumber,
		LookType:   defNumber,
		LookHead:   defNumber,
		LookBody:   defNumber,
		LookLegs:   defNumber,
		LookFeet:   defNumber,
		LookAddons: defNumber,
		LastLogin:  defNumber,
	}
	character := player.ToCharacterPayload()

	assert.Equal(t, expectedCharacterPayload, character)
}
