package models

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
)

type Account struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	PremDays int    `json:"premdays"`
	LastDay  int    `json:"lastday"`
}

type Player struct {
	Name       int    `json:"name"`
	Level      bool   `json:"level"`
	Sex        bool   `json:"sex"`
	Vocation   bool   `json:"vocation"`
	LookType   int    `json:"looktype"`
	LookHead   string `json:"lookhead"`
	LookBody   bool   `json:"lookbody"`
	LookLegs   string `json:"looklegs"`
	LookFeet   string `json:"lookfeet"`
	LookAddons string `json:"lookaddons"`
	LastLogin  string `json:"lastlogin"`
}

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

type TournamentInfo struct {
	IsTournamentParticipant          bool `json:"istournamentparticipant"`
	RemainingDailyTournamentPlayTime int  `json:"remainingdailytournamentplaytime"`
}

type Outfit struct {
	OutfitID    int `json:"outfitid"`
	AddonsFlags int `json:"addonsflags"`
	DetailColor int `json:"detailcolor"`
	HeadColor   int `json:"headcolor"`
	LegsColor   int `json:"legscolor"`
	TorsoColor  int `json:"torsocolor"`
}

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

type World struct {
	AntiCheatProtection        bool   `json:"anticheatprotection"`
	CurrentTournamentPhase     int    `json:"currenttournamentphase"`
	ExternalAddress            string `json:"externaladdress"`
	ExternalAddressProtected   string `json:"externaladdressprotected"`
	ExternalAddressUnprotected string `json:"externaladdressunprotected"`
	ExternalPort               int    `json:"externalport"`
	ExternalPortProtected      int    `json:"externalportprotected"`
	ExternalPortUnprotected    int    `json:"externalportunprotected"`
	ID                         int    `json:"id"`
	IsTournamentWorld          bool   `json:"istournamentworld"`
	Location                   string `json:"location"`
	Name                       string `json:"name"`
	PreviewState               int    `json:"previewstate"`
	PvpType                    int    `json:"pvptype"`
	RestrictedStore            bool   `json:"restrictedstore"`
}

type PlayData struct {
	Characters []CharacterPayload `json:"characters"`
	Worlds     []World            `json:"worlds"`
}

func (acc *Account) LoadCharacterList(db *sql.DB) ([]CharacterPayload, error) {
	statement := fmt.Sprintf("SELECT name, level, sex, vocation, looktype, lookhead, lookbody, looklegs, lookfeet, lookaddons, lastlogin from players where account_id = '%s'", acc.ID)
	rows, err := db.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	players := []CharacterPayload{}

	for rows.Next() {
		var sexId int
		player := CharacterPayload{WorldID: 0, CharacterInfo: CharacterInfo{IsMale: true}}
		if err := rows.Scan(
			&player.Name,
			&player.Level,
			&sexId,
			&player.Vocation,
			&player.OutfitID,
			&player.HeadColor,
			&player.TorsoColor,
			&player.LegsColor,
			&player.DetailColor,
			&player.AddonsFlags,
			&player.LegsColor,
		); err != nil {
			return nil, err
		}

		if sexId != 1 {
			player.IsMale = false
		}

		players = append(players, player)
	}

	return players, nil
}

func (acc *Account) Authenticate(db *sql.DB) error {
	h := sha1.New()
	h.Write([]byte(acc.Password))

	p := h.Sum(nil)

	statement := fmt.Sprintf(
		"SELECT id, premdays, lastday FROM accounts WHERE email = '%s' AND password = '%x'",
		acc.Email,
		p,
	)

	return db.QueryRow(statement).Scan(&acc.ID, &acc.PremDays, &acc.LastDay)
}
