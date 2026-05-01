package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestPayloadAcceptsLoginMetadata(t *testing.T) {
	raw := []byte(`{
		"type": "login",
		"email": "player@example.invalid",
		"password": "secret",
		"stayloggedin": true,
		"clientversion": "15.20.99c34c",
		"clienttype": 2,
		"assetversion": "assets-sha",
		"devicecookie": "device-cookie"
	}`)

	var payload RequestPayload
	assert.Nil(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "login", payload.Type)
	assert.Equal(t, "15.20.99c34c", payload.ClientVersion)
	assert.Equal(t, uint32(2), payload.ClientType)
	assert.Equal(t, "assets-sha", payload.AssetVersion)
	assert.Equal(t, "device-cookie", payload.DeviceCookie)
}

func TestRequestPayloadAcceptsNewsMetadata(t *testing.T) {
	raw := []byte(`{
		"type": "news",
		"fromtimestamp": 123456,
		"isreturner": false,
		"showrewardnews": true
	}`)

	var payload RequestPayload
	assert.Nil(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "news", payload.Type)
	assert.Equal(t, uint64(123456), payload.FromTimestamp)
	assert.False(t, payload.IsReturner)
	assert.True(t, payload.ShowRewardNews)
}

func TestRequestPayloadAcceptsNewsViewedMetadata(t *testing.T) {
	raw := []byte(`{
		"type": "newsviewed",
		"viewedid": 107
	}`)

	var payload RequestPayload
	assert.Nil(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "newsviewed", payload.Type)
	assert.Equal(t, uint32(107), payload.ViewedID)
}
