package models

type PlayData struct {
	Characters []CharacterPayload `json:"characters"`
	Worlds     []World            `json:"worlds"`
}
