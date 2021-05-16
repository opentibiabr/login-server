package login

type Session struct {
	EmailCodeRequest              bool   `json:"emailcoderequest"`
	FpsTracking                   bool   `json:"fpstracking"`
	IsPremium                     bool   `json:"ispremium"`
	IsReturner                    bool   `json:"isreturner"`
	LastLoginTime                 int    `json:"lastlogintime"`
	OptionTracking                bool   `json:"optiontracking"`
	PremiumUntil                  int    `json:"premiumuntil"`
	ReturnerNotification          bool   `json:"returnernotification"`
	SessionKey                    string `json:"sessionkey"`
	ShowRewardNews                bool   `json:"showrewardnews"`
	Status                        string `json:"status"`
	TournamentTicketPurchaseState int    `json:"tournamentticketpurchasestate"`
	TournamentCyclePhase          int    `json:"tournamentcyclephase"`
}
