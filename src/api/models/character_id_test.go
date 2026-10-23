package models

import (
	"encoding/json"
	"testing"

	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/stretchr/testify/assert"
)

func TestCharacterPayloadCharacterID(t *testing.T) {
	message := createCharacterMessage(1)
	message.CharacterId = 1073741825

	payload := LoadCharactersFromMessage([]*login_proto_messages.Character{message})[0]
	assert.Equal(t, uint32(1073741825), payload.CharacterID)

	encoded, err := json.Marshal(payload)
	assert.Nil(t, err)
	var decoded map[string]interface{}
	assert.Nil(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, float64(1073741825), decoded["characterid"])
}

func TestCharacterPayloadOmitsMissingCharacterID(t *testing.T) {
	payload := LoadCharactersFromMessage([]*login_proto_messages.Character{createCharacterMessage(1)})[0]
	encoded, err := json.Marshal(payload)
	assert.Nil(t, err)
	var decoded map[string]interface{}
	assert.Nil(t, json.Unmarshal(encoded, &decoded))
	_, present := decoded["characterid"]
	assert.False(t, present)
}
