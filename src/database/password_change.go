package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/opentibiabr/login-server/src/serviceerrors"
)

const passwordChangeCheckTimeout = 3 * time.Second

const detectPasswordChangeColumnStatement = "SELECT COUNT(*) FROM information_schema.COLUMNS " +
	"WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'accounts' AND COLUMN_NAME = 'password_change_required'"

const selectPasswordChangeRequiredStatement = "SELECT password_change_required FROM accounts WHERE id = ?"

type passwordChangeMode int

const (
	// passwordChangeOff is the default: the column is unknown or missing, and nothing is read.
	passwordChangeOff passwordChangeMode = iota
	// passwordChangeOn: the column exists and verified accounts are checked.
	passwordChangeOn
	// passwordChangePending: the startup detection failed; it is repeated on the next login, which fails meanwhile.
	passwordChangePending
)

var passwordChange = struct {
	sync.Mutex
	mode passwordChangeMode
}{}

func currentPasswordChangeMode() passwordChangeMode {
	passwordChange.Lock()
	defer passwordChange.Unlock()
	return passwordChange.mode
}

func setPasswordChangeMode(mode passwordChangeMode) {
	passwordChange.Lock()
	passwordChange.mode = mode
	passwordChange.Unlock()
}

// PasswordChangeRequiredState names the current mode for the startup log: "on", "off" or "pending" (the lookup failed
// and is repeated at the next login, which is refused until it succeeds).
func PasswordChangeRequiredState() string {
	switch currentPasswordChangeMode() {
	case passwordChangeOn:
		return "on"
	case passwordChangePending:
		return "pending"
	default:
		return "off"
	}
}

// DetectPasswordChangeRequired looks once for accounts.password_change_required in the current schema. Databases
// without the column keep the feature off. When the lookup itself fails the feature is retried at the next login
// instead of being silently left off. It returns whether verified accounts are going to be checked.
func DetectPasswordChangeRequired(ctx context.Context, db *sql.DB) bool {
	mode, err := detectPasswordChangeMode(ctx, db)
	if err != nil {
		logger.Warn(fmt.Sprintf("Could not look for accounts.password_change_required, retrying at the next login: %v", err))
	}
	setPasswordChangeMode(mode)
	return mode != passwordChangeOff
}

func detectPasswordChangeMode(ctx context.Context, db *sql.DB) (passwordChangeMode, error) {
	if db == nil {
		return passwordChangePending, errors.New("database connection is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	detectCtx, cancel := context.WithTimeout(ctx, passwordChangeCheckTimeout)
	defer cancel()

	var count int
	if err := db.QueryRowContext(detectCtx, detectPasswordChangeColumnStatement).Scan(&count); err != nil {
		return passwordChangePending, err
	}
	if count > 0 {
		return passwordChangeOn, nil
	}
	return passwordChangeOff, nil
}

// ensurePasswordChangeNotRequired runs after the password of acc has been verified, and before anything is issued or
// rewritten for the account.
func (acc *Account) ensurePasswordChangeNotRequired(ctx context.Context, db *sql.DB) error {
	mode := currentPasswordChangeMode()
	if mode == passwordChangeOff {
		return nil
	}
	if mode == passwordChangePending {
		detected, err := detectPasswordChangeMode(ctx, db)
		if err != nil {
			return serviceerrors.LoginService(serviceerrors.CodeAccountDataUnavailable, "ACCOUNT_DATA_UNAVAILABLE", err)
		}
		setPasswordChangeMode(detected)
		if detected == passwordChangeOff {
			return nil
		}
	}

	checkCtx, cancel := context.WithTimeout(ctx, passwordChangeCheckTimeout)
	defer cancel()
	var required sql.NullInt64
	if err := db.QueryRowContext(checkCtx, selectPasswordChangeRequiredStatement, acc.ID).Scan(&required); err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1054 {
			// The column disappeared (or never was there): this database does not use the feature.
			setPasswordChangeMode(passwordChangeOff)
			return nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		// The password is right and the flag cannot be read: refuse instead of letting a flagged account in.
		return serviceerrors.LoginService(serviceerrors.CodeAccountDataUnavailable, "ACCOUNT_DATA_UNAVAILABLE", err)
	}
	if required.Valid && required.Int64 != 0 {
		return serviceerrors.PasswordChangeRequired()
	}
	return nil
}
