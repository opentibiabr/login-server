package models

import (
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
)

var defaultString = "default"
var defaultNumber = 10

var outfitMsg = login_proto_messages.CharacterOutfit{
	LookType: uint32(defaultNumber),
	LookHead: uint32(defaultNumber),
	LookBody: uint32(defaultNumber),
	LookLegs: uint32(defaultNumber),
	LookFeet: uint32(defaultNumber),
	Addons:   uint32(defaultNumber),
}

func createCharacterInfo(sex uint32) *login_proto_messages.CharacterInfo {
	return &login_proto_messages.CharacterInfo{
		Name:     defaultString,
		Level:    uint32(defaultNumber),
		Vocation: defaultString,
		Sex:      uint32(sex),
	}
}

func createCharacterMessage(sex uint32) *login_proto_messages.Character {
	return &login_proto_messages.Character{
		WorldId: uint32(defaultNumber),
		Outfit:  &outfitMsg,
		Info:    createCharacterInfo(sex),
	}
}

func createSessionMessage(isPremium bool) *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    isPremium,
		PremiumUntil: uint32(defaultNumber),
		SessionKey:   defaultString,
		LastLogin:    uint32(defaultNumber),
	}
}

func createWorldMessage(id int) *login_proto_messages.World {
	return &login_proto_messages.World{
		Id:                         uint32(id),
		ExternalPort:               uint32(defaultNumber),
		ExternalPortProtected:      uint32(defaultNumber),
		ExternalPortUnprotected:    uint32(defaultNumber),
		ExternalAddress:            defaultString,
		ExternalAddressProtected:   defaultString,
		ExternalAddressUnprotected: defaultString,
		Location:                   defaultString,
		Name:                       defaultString,
	}
}
