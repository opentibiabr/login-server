package models

type RequestPayload struct {
	AssetVersion string `json:"assetversion"`
	ClientType uint32 `json:"clienttype"`
	ClientVersion string `json:"clientversion"`
	DeviceCookie string `json:"devicecookie"`
	Email string `json:"email"`
	FromTimestamp uint64 `json:"fromtimestamp"`
	IsReturner bool `json:"isreturner"`
	Password string `json:"password"`
	ShowRewardNews bool `json:"showrewardnews"`
	StayLoggedIn bool `json:"stayloggedin"`
	Type string `json:"type"`
	ViewedID uint32 `json:"viewedid"`
}

type ResponsePayload struct {
	DeviceCookie string `json:"devicecookie"`
	LoginEmail string `json:"loginemail"`
	PlayData PlayData `json:"playdata"`
	Session Session `json:"session"`
}

type LoginErrorPayload struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type PlayData struct {
	Characters []CharacterPayload `json:"characters"`
	Worlds     []World            `json:"worlds"`
}
