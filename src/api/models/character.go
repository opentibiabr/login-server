package models

import (
	"github.com/opentibiabr/login-server/src/configs"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

type CharacterPayload struct {
	WorldID int `json:"worldid"`
	CharacterInfo
	Outfit
	TournamentInfo
}

type CharacterInfo struct {
	DailyRewardState int    `json:"dailyrewardstate"`
	IsHidden         bool   `json:"ishidden"`
	IsMainCharacter  bool   `json:"ismaincharacter"`
	IsMale           bool   `json:"ismale"`
	Level            int    `json:"level"`
	Name             string `json:"name"`
	Tutorial         bool   `json:"tutorial"`
	Vocation         string `json:"vocation"`
}

type Outfit struct {
	OutfitID    int `json:"outfitid"`
	AddonsFlags int `json:"addonsflags"`
	DetailColor int `json:"detailcolor"`
	HeadColor   int `json:"headcolor"`
	LegsColor   int `json:"legscolor"`
	TorsoColor  int `json:"torsocolor"`
}

type TournamentInfo struct {
	IsTournamentParticipant          bool `json:"istournamentparticipant"`
	RemainingDailyTournamentPlayTime int  `json:"remainingdailytournamentplaytime"`
}

func LoadCharactersFromMessage(charactersMsg []*login_proto_messages.Character) []CharacterPayload {
	var characters []CharacterPayload
	for _, characterMsg := range charactersMsg {
		characters = append(
			characters,
			CharacterPayload{
				WorldID: 0,
				CharacterInfo: CharacterInfo{
					Name:     characterMsg.GetInfo().GetName(),
					Level:    int(characterMsg.GetInfo().GetLevel()),
					Vocation: configs.GetServerVocations()[int(characterMsg.GetInfo().GetLevel())],
					IsMale:   int(characterMsg.GetInfo().GetSex()) == 1,
				},
				Outfit: Outfit{
					OutfitID:    int(characterMsg.GetOutfit().GetLookType()),
					HeadColor:   int(characterMsg.GetOutfit().GetLookHead()),
					TorsoColor:  int(characterMsg.GetOutfit().GetLookBody()),
					LegsColor:   int(characterMsg.GetOutfit().GetLookLegs()),
					DetailColor: int(characterMsg.GetOutfit().GetLookFeet()),
					AddonsFlags: int(characterMsg.GetOutfit().GetAddons()),
				},
			})
	}

	return characters
}
