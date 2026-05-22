package database

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/logger"
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
const sessionDuration = 24 * time.Hour
const sessionPersistenceTimeout = 3 * time.Second

var ErrAccountSessionStorageUnavailable = errors.New("account session persistence unavailable")

func (acc *Account) Authenticate(db *sql.DB) error {
	h := sha1.New()
	h.Write([]byte(acc.Password))

	p := h.Sum(nil)
	passwordHash := fmt.Sprintf("%x", p)

	statement := "SELECT id, premdays, lastday FROM accounts WHERE (email = ? OR name = ?) AND password = ?"

	err := db.QueryRow(statement, acc.Email, acc.Email, passwordHash).Scan(&acc.ID, &acc.PremDays, &acc.LastDay)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (acc *Account) GetGrpcSession(sessionKey string) *login_proto_messages.Session {
	return &login_proto_messages.Session{
		IsPremium:    acc.PremDays > 0,
		PremiumUntil: acc.GetPremiumTime(),
		SessionKey:   sessionKey,
		LastLogin:    acc.LastLogin,
	}
}

func (acc *Account) CreateSession(ctx context.Context, db *sql.DB) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	sessionKey := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(sessionKey))
	expires := time.Now().Add(sessionDuration).Unix()

	if ctx == nil {
		ctx = context.Background()
	}

	writeCtx, cancel := context.WithTimeout(ctx, sessionPersistenceTimeout)
	defer cancel()

	if _, err := db.ExecContext(
		writeCtx,
		"INSERT INTO `account_sessions` (`id`, `account_id`, `expires`) VALUES (?, ?, ?)",
		fmt.Sprintf("%x", hash),
		acc.ID,
		expires,
	); err != nil {
		if isMissingAccountSessionsTable(err) {
			logger.Error(err)
			return "", ErrAccountSessionStorageUnavailable
		}
		return "", err
	}

	return sessionKey, nil
}

func (acc *Account) GetPremiumTime() uint64 {
	if acc.PremDays > 0 {
		return uint64(time.Now().Unix()) + uint64(acc.PremDays*secondsInADay)
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

func isMissingAccountSessionsTable(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1146
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "account_sessions") && strings.Contains(lowered, "doesn't exist")
}
