package tests

import (
	"bou.ke/monkey"
	"github.com/opentibiabr/login-server/src/database"
	"github.com/opentibiabr/login-server/src/http_api/login"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetSession(t *testing.T) {
	monkey.Patch(time.Now, func() time.Time {
		return time.Unix(0, 0)
	})

	expectedSession := login.Session{
		IsPremium:      true,
		PremiumUntil:   86400,
		SessionKey:     "a\nb",
		ShowRewardNews: true,
		Status:         "active",
	}

	acc := database.Account{
		PremDays: 1,
		Email:    "a",
		Password: "b",
	}

	session := acc.GetSession()

	assert.Equal(t, expectedSession, session)

	acc.PremDays = 0
	assert.False(t, acc.GetSession().IsPremium)
}

func TestGetPremiumTime(t *testing.T) {
	monkey.Patch(time.Now, func() time.Time {
		return time.Unix(1621103616, 0)
	})

	acc := database.Account{PremDays: -1000}
	assert.Equal(t, 0, acc.GetPremiumTime())

	acc = database.Account{PremDays: 0}
	assert.Equal(t, 0, acc.GetPremiumTime())

	acc = database.Account{PremDays: 1}
	assert.Equal(t, 1621103702400, acc.GetPremiumTime())
}
