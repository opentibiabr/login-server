package grpc_login_server

import (
	"testing"

	"github.com/opentibiabr/login-server/src/database"
	"github.com/stretchr/testify/assert"
)

func TestBuildLivestreamSession(t *testing.T) {
	session := buildLivestreamSession("any-password")

	assert.False(t, session.IsPremium)
	assert.Zero(t, session.PremiumUntil)
	assert.Zero(t, session.LastLogin)
	assert.Equal(t, database.LivestreamSessionAccount+"\nany-password", session.SessionKey)
}

func TestBuildLivestreamUnavailableResponse(t *testing.T) {
	response := buildLivestreamUnavailableResponse()

	assert.NotNil(t, response.GetError())
	assert.Equal(t, uint32(DefaultLoginErrorCode), response.GetError().Code)
	assert.Equal(t, livestreamUnavailableMessage, response.GetError().Message)
}
