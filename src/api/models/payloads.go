package models

type RequestPayload struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	StayLoggedIn bool   `json:"stayloggedin"`
	Type         string `json:"type"`
}

type ResponsePayload struct {
	PlayData PlayData `json:"playdata"`
	Session  Session  `json:"session"`
}

type LoginErrorPayload struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type PlayData struct {
	Characters []CharacterPayload `json:"characters"`
	Worlds     []World            `json:"worlds"`
}
