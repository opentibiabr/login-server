package database

import (
	"context"
	"crypto/rand"
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
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/passwords"
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

const passwordUpgradeTimeout = 3 * time.Second

// maxArgon2idPerLogin bounds the hashing one request can ask for when many accounts share an identifier. Legacy
// values cost next to nothing and are all checked.
const maxArgon2idPerLogin = 8

// passwordCheckTimeout bounds how long a login waits for a hashing slot.
var passwordCheckTimeout = 10 * time.Second

// Every account the identifier names is a candidate: the schema does not require unique emails.
const selectCredentialsStatement = "SELECT id, type, premdays, lastday, password FROM accounts WHERE email = ? OR name = ? ORDER BY email = ? DESC, id"

// The column width is part of the condition: without strict SQL mode a narrow column would silently cut the new value
// and lock the account out.
const upgradePasswordStatement = "UPDATE accounts SET password = ? WHERE id = ? AND password = ? AND ? <= " +
	"(SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'accounts' AND COLUMN_NAME = 'password')"

var defaultPasswordHasher = passwords.NewHasher(passwords.DefaultConcurrency, false)

var ErrAccountSessionStorageUnavailable = errors.New("account session persistence unavailable")

// Authenticate finds the account whose email or name is acc.Email and whose stored password matches acc.Password.
// It returns sql.ErrNoRows when there is none. With rehashing enabled, a matching value in an older format is
// replaced; failing to replace it does not fail the login.
func (acc *Account) Authenticate(ctx context.Context, db *sql.DB, hasher *passwords.Hasher) error {
	if db == nil {
		return serviceerrors.LoginService(
			serviceerrors.CodeDatabaseUnavailable,
			"DATABASE_UNAVAILABLE",
			errors.New("database connection is nil"),
		)
	}
	if hasher == nil {
		hasher = defaultPasswordHasher
	}
	if ctx == nil {
		ctx = context.Background()
	}
	// A blank identifier would name every account without an e-mail address; databases ignore trailing spaces.
	if strings.TrimSpace(acc.Email) == "" {
		return sql.ErrNoRows
	}
	candidates, unreadable, err := loadCredentials(ctx, db, acc.Email)
	if err != nil {
		return err
	}
	checkCtx, cancel := context.WithTimeout(ctx, passwordCheckTimeout)
	defer cancel()

	computedArgon2id := 0
	for _, candidate := range candidates {
		usesArgon2id := passwords.IsArgon2id(candidate.password)
		if usesArgon2id && computedArgon2id == maxArgon2idPerLogin {
			continue
		}
		matches, err := hasher.Verify(checkCtx, acc.Password, candidate.password)
		if errors.Is(err, passwords.ErrBusy) {
			return busyError(err)
		}
		if usesArgon2id && err == nil {
			computedArgon2id++
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("Account %d has an unusable stored password: %v", candidate.id, err))
			continue
		}
		if !matches {
			continue
		}

		acc.ID, acc.Type, acc.PremDays, acc.LastDay = candidate.id, candidate.accountType, candidate.premDays, candidate.lastDay
		// Only a correct password reveals the flag, and nothing is rewritten for a flagged account.
		if err := acc.ensurePasswordChangeNotRequired(ctx, db); err != nil {
			return err
		}
		if hasher.RehashEnabled() && passwords.NeedsRehash(candidate.password) {
			if err := upgradeStoredPassword(ctx, db, hasher, candidate, acc.Password); err != nil {
				logger.Warn(fmt.Sprintf("Account %d kept its previous password format: %v", candidate.id, err))
			}
		}
		return nil
	}

	if unreadable != nil {
		return unreadable
	}
	// While values are being upgraded, an unknown account and a wrong password on a legacy value take as long as a
	// wrong password on an Argon2id value, and are refused as busy under the same conditions.
	if computedArgon2id == 0 && hasher.RehashEnabled() {
		if err := hasher.Waste(checkCtx, acc.Password); errors.Is(err, passwords.ErrBusy) {
			return busyError(err)
		}
	}
	return sql.ErrNoRows
}

func busyError(cause error) error {
	return serviceerrors.LoginService(serviceerrors.CodePasswordVerificationBusy, "PASSWORD_VERIFICATION_BUSY", cause)
}

type storedCredentials struct {
	id          uint32
	accountType uint32
	premDays    uint32
	lastDay     uint32
	password    string
}

// loadCredentials reads every candidate before any hashing, so no database connection is held while hashing. A row
// which cannot be read does not hide the others; its error is returned in unreadable. Rows without a password are
// no candidates.
func loadCredentials(ctx context.Context, db *sql.DB, identifier string) (candidates []storedCredentials, unreadable error, err error) {
	rows, err := db.QueryContext(ctx, selectCredentialsStatement, identifier, identifier, identifier)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var candidate storedCredentials
		var password sql.NullString
		if scanErr := rows.Scan(&candidate.id, &candidate.accountType, &candidate.premDays, &candidate.lastDay, &password); scanErr != nil {
			unreadable = scanErr
			continue
		}
		if !password.Valid {
			continue
		}
		candidate.password = password.String
		candidates = append(candidates, candidate)
	}
	return candidates, unreadable, rows.Err()
}

// upgradeStoredPassword replaces the value that has just been verified. The previous value is part of the condition,
// so a password changed in the meantime is never overwritten.
func upgradeStoredPassword(ctx context.Context, db *sql.DB, hasher *passwords.Hasher, current storedCredentials, password string) error {
	upgradeCtx, cancel := context.WithTimeout(ctx, passwordUpgradeTimeout)
	defer cancel()

	upgraded, err := hasher.Hash(upgradeCtx, password)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(upgradeCtx, upgradePasswordStatement, upgraded, current.id, current.password, passwords.MinColumnLength)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return fmt.Errorf("not replaced: the value changed meanwhile or accounts.password holds fewer than %d characters", passwords.MinColumnLength)
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

func LoadAccount(ctx context.Context, email string, password string, DB *sql.DB, hasher *passwords.Hasher) (*Account, error) {
	acc := Account{Email: email, Password: password}
	if err := acc.Authenticate(ctx, DB, hasher); err != nil {
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
