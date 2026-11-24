package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/opentibiabr/login-server/src/grpc/login_proto_messages"
	"github.com/opentibiabr/login-server/src/serviceerrors"
	"github.com/stretchr/testify/assert"
)

func TestAccount_PremiumStatus(t *testing.T) {
	now := time.Unix(1_795_482_505, 0)
	future := uint32(now.Unix() + 12*secondsInADay + 3600)
	expired := uint32(now.Unix() - 5*secondsInADay)

	tests := []struct {
		name            string
		account         Account
		freePremium     bool
		wantIsPremium   bool
		wantPremiumTime uint64
	}{{
		name:            "paid time in the future reports that moment",
		account:         Account{LastDay: future},
		wantIsPremium:   true,
		wantPremiumTime: uint64(future),
	}, {
		name:            "paid time in the future wins over the free premium value",
		account:         Account{LastDay: future},
		freePremium:     true,
		wantIsPremium:   true,
		wantPremiumTime: uint64(future),
	}, {
		name:            "expired time without free premium is not premium",
		account:         Account{LastDay: expired},
		wantIsPremium:   false,
		wantPremiumTime: 0,
	}, {
		name:            "expired time with free premium reports the free premium value",
		account:         Account{LastDay: expired},
		freePremium:     true,
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + freePremiumSeconds,
	}, {
		name:            "an account that never had paid time follows the same rule",
		account:         Account{},
		freePremium:     true,
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + freePremiumSeconds,
	}, {
		name:            "a stale premdays cache is ignored (the 12 days shown after the paid time ended)",
		account:         Account{PremDays: 12, LastDay: expired},
		wantIsPremium:   false,
		wantPremiumTime: 0,
	}, {
		name:            "a stale premdays cache does not hide the free premium value either",
		account:         Account{PremDays: 12, LastDay: expired},
		freePremium:     true,
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + freePremiumSeconds,
	}, {
		name:            "paid time ending exactly now has ended",
		account:         Account{LastDay: uint32(now.Unix())},
		wantIsPremium:   false,
		wantPremiumTime: 0,
	}, {
		name:            "one second of paid time left is reported as a day and a half",
		account:         Account{LastDay: uint32(now.Unix() + 1)},
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + minimumPaidSeconds,
	}, {
		name:            "three hours of paid time left is reported as a day and a half",
		account:         Account{LastDay: uint32(now.Unix() + 3*3600)},
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + minimumPaidSeconds,
	}, {
		name:            "just over a day of paid time left is still reported as a day and a half",
		account:         Account{LastDay: uint32(now.Unix() + secondsInADay + 1)},
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + minimumPaidSeconds,
	}, {
		name:            "a day and a half of paid time left is reported as it is",
		account:         Account{LastDay: uint32(now.Unix() + minimumPaidSeconds)},
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix()) + minimumPaidSeconds,
	}, {
		name:            "two days of paid time left is reported as it is",
		account:         Account{LastDay: uint32(now.Unix() + 2*secondsInADay)},
		wantIsPremium:   true,
		wantPremiumTime: uint64(now.Unix() + 2*secondsInADay),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPremium, premiumTime := tt.account.PremiumStatus(now, tt.freePremium)
			assert.Equal(t, tt.wantIsPremium, isPremium)
			assert.Equal(t, tt.wantPremiumTime, premiumTime)
		})
	}
}

// The client counts whole days with floor((premiumuntil - clientNow) / 86400) and prints "free days" only for exactly 65535.
func clientDayCount(premiumUntil uint64, clientNow int64) int64 {
	seconds := int64(premiumUntil) - clientNow
	days := seconds / secondsInADay
	if seconds%secondsInADay != 0 && seconds < 0 {
		days--
	}
	return days
}

// delta is how far the client clock is ahead of the login server clock (negative: behind), in seconds.
func TestAccount_FreePremiumValueKeepsTheClientDayCount(t *testing.T) {
	now := time.Unix(1_795_482_505, 0)
	_, premiumUntil := (&Account{}).PremiumStatus(now, true)

	for _, delta := range []int64{-secondsInADay/2 + 1, -11 * 3600, -3600, -60, -1, 0, 1, 60, 3600, 11 * 3600, secondsInADay / 2} {
		assert.Equal(t, int64(freePremiumDays), clientDayCount(premiumUntil, now.Unix()+delta), "client clock %+ds from the server", delta)
	}

	// Outside half a day the count moves, and the client would print the days instead of the free text.
	assert.Equal(t, int64(freePremiumDays)+1, clientDayCount(premiumUntil, now.Unix()-secondsInADay/2))
	assert.Equal(t, int64(freePremiumDays)-1, clientDayCount(premiumUntil, now.Unix()+secondsInADay/2+1))
}

// With paid time the client shows the whole days left, and never the "free days" text (a count of 0) while any is left.
func TestAccount_PaidTimeNeverCountsAsZeroDaysInTheClient(t *testing.T) {
	now := time.Unix(1_795_482_505, 0)

	for _, remaining := range []int64{1, 60, 3 * 3600, secondsInADay - 1, secondsInADay, secondsInADay + 1, minimumPaidSeconds - 1, minimumPaidSeconds, 2 * secondsInADay, 12*secondsInADay + 3600} {
		acc := &Account{LastDay: uint32(now.Unix() + remaining)}
		isPremium, premiumUntil := acc.PremiumStatus(now, false)
		assert.True(t, isPremium, "remaining %ds", remaining)
		for _, delta := range []int64{-11 * 3600, -60, 0, 1, 60, 3600, 11 * 3600} {
			assert.GreaterOrEqual(t, clientDayCount(premiumUntil, now.Unix()+delta), int64(1), "remaining %ds, client clock %+ds from the server", remaining, delta)
		}
	}

	// From a day and a half on, the count is the whole days left, as Canary counts them.
	for _, remaining := range []int64{minimumPaidSeconds, 2 * secondsInADay, 12*secondsInADay + 3600, 30 * secondsInADay} {
		acc := &Account{LastDay: uint32(now.Unix() + remaining)}
		_, premiumUntil := acc.PremiumStatus(now, false)
		assert.Equal(t, remaining/secondsInADay, clientDayCount(premiumUntil, now.Unix()), "remaining %ds", remaining)
	}
}

func TestAccount_GetGrpcSession(t *testing.T) {
	future := uint32(time.Now().Unix() + 30*secondsInADay)
	expired := uint32(time.Now().Unix() - 5*secondsInADay)

	t.Run("paid time keeps the session key and the last login", func(t *testing.T) {
		acc := &Account{LastDay: future, LastLogin: 77}
		session := acc.GetGrpcSession("opaque-session", true)
		assert.True(t, session.IsPremium)
		assert.Equal(t, uint64(future), session.PremiumUntil)
		assert.Equal(t, "opaque-session", session.SessionKey)
		assert.Equal(t, uint32(77), session.LastLogin)
	})

	t.Run("expired time with a stale premdays is a free account when free premium is off", func(t *testing.T) {
		acc := &Account{PremDays: 12, LastDay: expired}
		assert.Equal(t, &login_proto_messages.Session{SessionKey: "opaque-session"}, acc.GetGrpcSession("opaque-session", false))
	})

	t.Run("expired time with free premium on reports the free premium value", func(t *testing.T) {
		acc := &Account{PremDays: 12, LastDay: expired}
		session := acc.GetGrpcSession("opaque-session", true)
		assert.True(t, session.IsPremium)
		assert.GreaterOrEqual(t, session.PremiumUntil, uint64(time.Now().Unix())+freePremiumDays*secondsInADay)
	})
}

func TestAccount_IsAdmin(t *testing.T) {
	assert.False(t, (&Account{Type: 1}).IsAdmin())
	assert.False(t, (&Account{Type: 3}).IsAdmin())
	assert.True(t, (&Account{Type: 4}).IsAdmin())
	assert.True(t, (&Account{Type: 6}).IsAdmin())
}

func TestAccount_CreateSessionRejectsNilReceiver(t *testing.T) {
	var acc *Account

	_, err := acc.CreateSession(context.Background(), nil)

	publicErr, ok := serviceerrors.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serviceerrors.CodeSessionCreateFailed, publicErr.Code)
	assert.Equal(t, "SESSION_CREATE_FAILED", publicErr.Name)
	assert.ErrorContains(t, publicErr, "account is nil")
}

func TestIsMissingAccountSessionsTable(t *testing.T) {
	assert.True(t, isMissingAccountSessionsTable(&mysqlDriver.MySQLError{
		Number:  1146,
		Message: "Table 'otserv.account_sessions' doesn't exist",
	}))
	assert.True(t, isMissingAccountSessionsTable(errors.New("table `account_sessions` doesn't exist")))
	assert.False(t, isMissingAccountSessionsTable(errors.New("connection refused")))
}

func TestClassifyAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantName string
	}{
		{
			name:     "invalid credentials",
			err:      sql.ErrNoRows,
			wantCode: serviceerrors.CodeInvalidCredentials,
			wantName: "INVALID_CREDENTIALS",
		},
		{
			name: "database unavailable",
			err: &mysqlDriver.MySQLError{
				Number:  1049,
				Message: "Unknown database 'canary'",
			},
			wantCode: serviceerrors.CodeDatabaseUnavailable,
			wantName: "DATABASE_UNAVAILABLE",
		},
		{
			name: "account data unavailable",
			err: &mysqlDriver.MySQLError{
				Number:  1146,
				Message: "Table 'canary.accounts' doesn't exist",
			},
			wantCode: serviceerrors.CodeAccountDataUnavailable,
			wantName: "ACCOUNT_DATA_UNAVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publicErr, ok := serviceerrors.FromError(classifyAuthenticationError(tt.err))

			assert.True(t, ok)
			assert.Equal(t, tt.wantCode, publicErr.Code)
			assert.Equal(t, tt.wantName, publicErr.Name)
		})
	}
}
