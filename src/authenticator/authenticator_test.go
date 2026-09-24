package authenticator

import (
	"encoding/base32"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchingStepUsesRFC6238SHA1Vector(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))

	step, ok := MatchingStep(secret, "287082", 59)

	assert.True(t, ok)
	assert.Equal(t, int64(1), step)
	_, ok = MatchingStep(secret, "287082", 149)
	assert.False(t, ok)
}

func TestMatchingStepAcceptsAdjacentWindow(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))
	token, err := tokenAtStep(secret, 100)
	require.NoError(t, err)

	step, ok := MatchingStep(secret, token, 101*periodSeconds)

	assert.True(t, ok)
	assert.Equal(t, int64(100), step)
}

func TestMatchingStepRejectsMalformedToken(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))

	for _, token := range []string{"", "12345", "1234567", "12a456"} {
		_, ok := MatchingStep(secret, token, 0)
		assert.False(t, ok, token)
	}
}

func TestSecretEnvelopeRoundTripAndAccountBinding(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	nonce := []byte("0123456789ab")
	secret := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

	envelope, err := EncryptSecret(secret, key, 42, nonce)
	require.NoError(t, err)
	assert.Equal(t, "v1:MDEyMzQ1Njc4OWFi0+u0bRr53EuwTXP2gFQkdwjEh5hO+y47YcefUvqowfm20UEprJIafeZd9qNxyQU2", envelope)

	decrypted, err := DecryptSecret(envelope, key, 42)
	require.NoError(t, err)
	assert.Equal(t, secret, decrypted)

	_, err = DecryptSecret(envelope, key, 43)
	assert.ErrorIs(t, err, ErrInvalidEnvelope)
}

func TestDecodeEncryptionKeyRejectsWrongLength(t *testing.T) {
	_, err := DecodeEncryptionKey(base64.StdEncoding.EncodeToString([]byte("too short")))
	assert.ErrorIs(t, err, ErrInvalidEncryptionKey)
}
