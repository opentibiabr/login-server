package database

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"time"
)

type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	PremDays int    `json:"premdays"`
	LastDay  int    `json:"lastday"`
}

const secondsInADay = 24 * 60 * 60
const million = 1e6

func (acc *Account) Authenticate(db *sql.DB) error {
	h := sha1.New()
	h.Write([]byte(acc.Password))

	p := h.Sum(nil)

	statement := fmt.Sprintf(
		"SELECT id, premdays, lastday FROM accounts WHERE email = '%s' AND password = '%x'",
		acc.Email,
		p,
	)

	err := db.QueryRow(statement).Scan(&acc.ID, &acc.PremDays, &acc.LastDay)
	if err != nil {
		return err
	}

	return nil
}

func (acc *Account) GetSession() models.Session {
	return models.Session{
		IsPremium:      acc.PremDays > 0,
		PremiumUntil:   acc.GetPremiumTime(),
		SessionKey:     fmt.Sprintf("%s\n%s", acc.Email, acc.Password),
		ShowRewardNews: true,
		Status:         "active",
	}
}

func (acc *Account) GetGrpcSession() *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    acc.PremDays > 0,
		PremiumUntil: uint32(acc.GetPremiumTime()),
		SessionKey:   fmt.Sprintf("%s\n%s", acc.Email, acc.Password),
	}
}

func (acc *Account) GetPremiumTime() int {
	if acc.PremDays > 0 {
		return int(time.Now().UnixNano()/million) + acc.PremDays*secondsInADay
	}
	return 0
}

func LoadAccount(email string, password string, DB *sql.DB) (*Account, error) {
	acc := Account{Email: email, Password: password}
	if err := acc.Authenticate(DB); err != nil {
		logger.Debug(err.Error())
		return nil, errors.New("Account email or password is not correct.")
	}

	return &acc, nil
}
