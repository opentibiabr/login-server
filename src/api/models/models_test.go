package models

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

var defaultString = "default"
var defaultNumber = uint32(10)

var outfitMsg = login_proto_messages.CharacterOutfit{
	LookType: defaultNumber,
	LookHead: defaultNumber,
	LookBody: defaultNumber,
	LookLegs: defaultNumber,
	LookFeet: defaultNumber,
	Addons:   defaultNumber,
}

func createCharacterInfo(sex uint32) *login_proto_messages.CharacterInfo {
	return &login_proto_messages.CharacterInfo{
		Name:     defaultString,
		Level:    defaultNumber,
		Vocation: defaultString,
		Sex:      sex,
	}
}

func createCharacterMessage(sex uint32) *login_proto_messages.Character {
	return &login_proto_messages.Character{
		WorldId: defaultNumber,
		Outfit:  &outfitMsg,
		Info:    createCharacterInfo(sex),
	}
}

func createSessionMessage(isPremium bool) *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    isPremium,
		PremiumUntil: uint64(defaultNumber),
		SessionKey:   defaultString,
		LastLogin:    defaultNumber,
	}
}

func createWorldMessage(id uint32) *login_proto_messages.World {
	return &login_proto_messages.World{
		Id:                         id,
		ExternalPort:               defaultNumber,
		ExternalPortProtected:      defaultNumber,
		ExternalPortUnprotected:    defaultNumber,
		ExternalAddress:            defaultString,
		ExternalAddressProtected:   defaultString,
		ExternalAddressUnprotected: defaultString,
		Location:                   defaultString,
		Name:                       defaultString,
	}
}
