package models

type RequestPayload struct {
	AssetVersion       string `json:"assetversion"`
	ClientType         uint32 `json:"clienttype"`
	ClientVersion      string `json:"clientversion"`
	DeviceCookie       string `json:"devicecookie"`
	Email              string `json:"email"`
	FromTimestamp      uint64 `json:"fromtimestamp"`
	IsReturner         bool   `json:"isreturner"`
	Password           string `json:"password"`
	ShowRewardNews     bool   `json:"showrewardnews"`
	StayLoggedIn       bool   `json:"stayloggedin"`
	Token              string `json:"token"`
	TrustedDeviceToken string `json:"trusteddevicetoken"`
	TrustDevice        bool   `json:"trustdevice"`
	DeviceName         string `json:"devicename"`
	Type               string `json:"type"`
	ViewedID           uint32 `json:"viewedid"`
}

type ResponsePayload struct {
	DeviceCookie           string   `json:"devicecookie"`
	LoginEmail             string   `json:"loginemail"`
	PlayData               PlayData `json:"playdata"`
	Session                Session  `json:"session"`
	TrustedDeviceToken     string   `json:"trusteddevicetoken,omitempty"`
	TrustedDeviceExpiresAt uint64   `json:"trusteddeviceexpiresat,omitempty"`
}

type LoginErrorPayload struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	// TrustDeviceAllowed is set only on an authenticator challenge, when the
	// server knows whether this login may remember the device. Clients that do
	// not read it keep offering the option.
	TrustDeviceAllowed *bool `json:"trustdeviceallowed,omitempty"`
}

type PlayData struct {
	Characters []CharacterPayload `json:"characters"`
	Worlds     []World            `json:"worlds"`
}
