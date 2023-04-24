package database

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
	"time"
)

type Account struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	PremDays  uint32 `json:"premdays"`
	LastDay   uint32 `json:"lastday"`
	LastLogin uint32
}

const secondsInADay = 24 * 60 * 60

func (acc *Account) Authenticate(db *sql.DB) error {
	h := sha1.New()
	h.Write([]byte(acc.Password))

	p := h.Sum(nil)

	statement := "SELECT id, premdays, lastday FROM accounts WHERE email = ? AND password = ?"
	row := db.QueryRow(statement, acc.Email, p)

	err := row.Scan(&acc.ID, &acc.PremDays, &acc.LastDay)
	if err != nil {
		return err
	}

	return nil
}

func (acc *Account) GetGrpcSession() *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    acc.PremDays > 0,
		PremiumUntil: acc.GetPremiumTime(),
		SessionKey:   fmt.Sprintf("%s\n%s", acc.Email, acc.Password),
		LastLogin:    acc.LastLogin,
	}
}

func (acc *Account) GetPremiumTime() uint64 {
	if acc.PremDays > 0 {
		return uint64(time.Now().Unix()) + uint64(acc.PremDays * secondsInADay)
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
