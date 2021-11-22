package models

import "github.com/opentibiabr/login-server/src/grpc/login_proto_messages"

type Session struct {
	EmailCodeRequest              bool   `json:"emailcoderequest"`
	FpsTracking                   bool   `json:"fpstracking"`
	IsPremium                     bool   `json:"ispremium"`
	IsReturner                    bool   `json:"isreturner"`
	LastLoginTime                 uint32 `json:"lastlogintime" proto:"LastLogin"`
	OptionTracking                bool   `json:"optiontracking"`
	PremiumUntil                  uint64 `json:"premiumuntil"`
	ReturnerNotification          bool   `json:"returnernotification"`
	SessionKey                    string `json:"sessionkey"`
	ShowRewardNews                bool   `json:"showrewardnews"`
	Status                        string `json:"status"`
	TournamentTicketPurchaseState uint32 `json:"tournamentticketpurchasestate"`
	TournamentCyclePhase          uint32 `json:"tournamentcyclephase"`
}

func LoadSessionFromMessage(sessionMsg *login_proto_messages.Session) Session {
	return *FromProtoConvertor(sessionMsg, &Session{Status: "active"}).(*Session)
}
