package models

import "github.com/opentibiabr/login-server/src/grpc/login_proto_messages"

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

func LoadSessionFromMessage(sessionMsg *login_proto_messages.Session) Session {
	return Session{
		IsPremium:     sessionMsg.IsPremium,
		PremiumUntil:  int(sessionMsg.PremiumUntil),
		SessionKey:    sessionMsg.SessionKey,
		LastLoginTime: int(sessionMsg.LastLogin),
	}
}
