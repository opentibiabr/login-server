package database

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

type Account struct {
	ID        uint32 `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Type      uint32 `json:"type"`
	PremDays  uint32 `json:"premdays"`
	LastDay   uint32 `json:"lastday"`
	LastLogin uint32
}

const accountTypeGameMaster = 4
const secondsInADay = 24 * 60 * 60
const sessionDuration = 24 * time.Hour
const sessionPersistenceTimeout = 3 * time.Second

var ErrAccountSessionStorageUnavailable = errors.New("account session persistence unavailable")

func (acc *Account) Authenticate(db *sql.DB) error {
	if db == nil {
		return serviceerrors.LoginService(
			serviceerrors.CodeDatabaseUnavailable,
			"DATABASE_UNAVAILABLE",
			errors.New("database connection is nil"),
		)
	}

	h := sha1.New()
	h.Write([]byte(acc.Password))

	p := h.Sum(nil)
	passwordHash := fmt.Sprintf("%x", p)

	statement := "SELECT id, type, premdays, lastday FROM accounts WHERE (email = ? OR name = ?) AND password = ?"

	err := db.QueryRow(statement, acc.Email, acc.Email, passwordHash).Scan(&acc.ID, &acc.Type, &acc.PremDays, &acc.LastDay)
	if err != nil {
		return err
	}

	return nil
}

// GetGrpcSession builds the session the client shows in its character list. The premium state comes from `lastday`, the
// moment the paid time ends, which is what the game server and the website read. The `premdays` column is only a
// cache of the days left that nothing refreshes when the time expires, so it is not used for the account status.
// When freePremium is true, an account without paid time is shown as a free premium account instead of a free one.
func (acc *Account) GetGrpcSession(sessionKey string, freePremium bool) *login_proto_messages.Session {
	isPremium, premiumUntil := acc.PremiumStatus(time.Now(), freePremium)
	return &login_proto_messages.Session{
		IsPremium:    isPremium,
		PremiumUntil: premiumUntil,
		SessionKey:   sessionKey,
		LastLogin:    acc.LastLogin,
	}
}

func (acc *Account) IsAdmin() bool {
	return acc != nil && acc.Type >= accountTypeGameMaster
}

func (acc *Account) CreateSession(ctx context.Context, db *sql.DB) (string, error) {
	if acc == nil {
		return "", serviceerrors.LoginService(
			serviceerrors.CodeSessionCreateFailed,
			"SESSION_CREATE_FAILED",
			errors.New("account is nil"),
		)
	}
	if db == nil {
		return "", serviceerrors.LoginService(
			serviceerrors.CodeDatabaseUnavailable,
			"DATABASE_UNAVAILABLE",
			errors.New("database connection is nil"),
		)
	}
	return acc.createSession(ctx, db)
}

type accountSessionWriter interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

func (acc *Account) createSession(ctx context.Context, writer accountSessionWriter) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", serviceerrors.LoginService(
			serviceerrors.CodeSessionCreateFailed,
			"SESSION_CREATE_FAILED",
			err,
		)
	}

	sessionKey := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(sessionKey))
	expires := time.Now().Add(sessionDuration).Unix()

	if ctx == nil {
		ctx = context.Background()
	}

	writeCtx, cancel := context.WithTimeout(ctx, sessionPersistenceTimeout)
	defer cancel()

	if _, err := writer.ExecContext(
		writeCtx,
		"INSERT INTO `account_sessions` (`id`, `account_id`, `expires`) VALUES (?, ?, ?)",
		fmt.Sprintf("%x", hash),
		acc.ID,
		expires,
	); err != nil {
		if isMissingAccountSessionsTable(err) {
			return "", serviceerrors.LoginService(
				serviceerrors.CodeSessionStorageUnavailable,
				"SESSION_STORAGE_UNAVAILABLE",
				fmt.Errorf("%w: %v", ErrAccountSessionStorageUnavailable, err),
			)
		}
		return "", serviceerrors.LoginService(
			serviceerrors.CodeSessionCreateFailed,
			"SESSION_CREATE_FAILED",
			err,
		)
	}

	return sessionKey, nil
}

// The client turns premiumuntil into whole days, floor((premiumuntil - now) / 86400), and treats exactly 65535 days as
// "free premium". The extra half day keeps that count at 65535 while the client clock is less than half a day behind or
// at most half a day ahead of the login server clock (it shows 65536 or 65534 days outside that window).
const freePremiumDays = 65535
const freePremiumSeconds = freePremiumDays*secondsInADay + secondsInADay/2

// With less than a day (plus the margin for latency) of paid time left the client would count 0 days and print the free
// days text, as if the paid time had already ended, and its expiry highlight would turn off on the day it matters most.
// While less than a day and a half is left, report a day and a half instead, so the client shows one day left.
const minimumPaidSeconds = secondsInADay + secondsInADay/2

// PremiumStatus returns whether the account is premium and until when, for the client. An account with paid time
// (`lastday` in the future) reports that moment, so the client shows the whole days left (rounded down, like Canary's
// remaining days), and at least one day while any paid time is left. Without paid time it reports the free premium
// value when freePremium is true, and no premium otherwise.
func (acc *Account) PremiumStatus(now time.Time, freePremium bool) (bool, uint64) {
	if remaining := int64(acc.LastDay) - now.Unix(); remaining > 0 {
		if remaining < minimumPaidSeconds {
			return true, uint64(now.Unix()) + minimumPaidSeconds
		}
		return true, uint64(acc.LastDay)
	}
	if freePremium {
		return true, uint64(now.Unix()) + freePremiumSeconds
	}
	return false, 0
}

func LoadAccount(email string, password string, DB *sql.DB) (*Account, error) {
	acc := Account{Email: email, Password: password}
	if err := acc.Authenticate(DB); err != nil {
		return nil, classifyAuthenticationError(err)
	}

	return &acc, nil
}

func classifyAuthenticationError(err error) error {
	if _, ok := serviceerrors.FromError(err); ok {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return serviceerrors.InvalidCredentials()
	}

	if isDatabaseUnavailableError(err) {
		return serviceerrors.LoginService(
			serviceerrors.CodeDatabaseUnavailable,
			"DATABASE_UNAVAILABLE",
			err,
		)
	}

	return serviceerrors.LoginService(
		serviceerrors.CodeAccountDataUnavailable,
		"ACCOUNT_DATA_UNAVAILABLE",
		err,
	)
}

func isMissingAccountSessionsTable(err error) bool {
	return isMissingTableError(err, "account_sessions")
}

func isMissingTableError(err error, tableName string) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1146
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, strings.ToLower(tableName)) && strings.Contains(lowered, "doesn't exist")
}

func isDatabaseUnavailableError(err error) bool {
	if errors.Is(err, driver.ErrBadConn) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1045, 1049, 2002, 2003, 2006, 2013:
			return true
		}
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "connection refused") ||
		strings.Contains(lowered, "can't connect") ||
		strings.Contains(lowered, "no such host")
}
