package models

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

type CharacterPayload struct {
	WorldID uint32 `json:"worldid"`
	CharacterInfo
	Outfit
	TournamentInfo
}

type CharacterInfo struct {
	DailyRewardState uint32 `json:"dailyrewardstate"`
	IsHidden         bool   `json:"ishidden"`
	IsMainCharacter  bool   `json:"ismaincharacter"`
	IsMale           bool   `json:"ismale" proto:"Sex" proto_func:"isMale"`
	Level            uint32 `json:"level"`
	Name             string `json:"name"`
	Tutorial         bool   `json:"tutorial"`
	Vocation         string `json:"vocation"`
}

type Outfit struct {
	OutfitID    uint32 `json:"outfitid" proto:"LookType"`
	AddonsFlags uint32 `json:"addonsflags" proto:"Addons"`
	DetailColor uint32 `json:"detailcolor" proto:"LookFeet"`
	HeadColor   uint32 `json:"headcolor" proto:"LookHead"`
	LegsColor   uint32 `json:"legscolor" proto:"LookLegs"`
	TorsoColor  uint32 `json:"torsocolor" proto:"LookBody"`
}

type TournamentInfo struct {
	IsTournamentParticipant          bool   `json:"istournamentparticipant"`
	RemainingDailyTournamentPlayTime uint32 `json:"remainingdailytournamentplaytime"`
}

func LoadCharactersFromMessage(charactersMsg []*login_proto_messages.Character) []CharacterPayload {
	var characters []CharacterPayload
	for _, characterMsg := range charactersMsg {
		characters = append(characters, loadCharacterFromMessage(characterMsg))
	}
	return characters
}

func loadCharacterFromMessage(characterMsg *login_proto_messages.Character) CharacterPayload {
	return CharacterPayload{
		WorldID:       characterMsg.GetWorldId(),
		CharacterInfo: loadCharacterInfoFromMessage(characterMsg.GetInfo()),
		Outfit:        loadOutfitInfoFromMessage(characterMsg.GetOutfit()),
	}
}

func loadCharacterInfoFromMessage(characterInfoMsg *login_proto_messages.CharacterInfo) CharacterInfo {
	return *FromProtoConvertor(characterInfoMsg, &CharacterInfo{}).(*CharacterInfo)
}

func loadOutfitInfoFromMessage(outfitInfoMsg *login_proto_messages.CharacterOutfit) Outfit {
	return *FromProtoConvertor(outfitInfoMsg, &Outfit{}).(*Outfit)
}
