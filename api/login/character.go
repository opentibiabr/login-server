package login

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
