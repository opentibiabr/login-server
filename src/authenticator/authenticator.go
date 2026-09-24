package authenticator

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	envelopeVersion = "v1"
	nonceSize       = 12
	tokenDigits     = 6
	periodSeconds   = int64(30)
)

var (
	ErrInvalidEncryptionKey = errors.New("authenticator encryption key must be base64-encoded and exactly 32 bytes")
	ErrInvalidEnvelope      = errors.New("invalid authenticator secret envelope")
	ErrInvalidSecret        = errors.New("invalid authenticator secret")
)

func DecodeEncryptionKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidEncryptionKey
	}
	return key, nil
}

func EncryptSecret(secret string, encodedKey string, accountID uint32, nonce []byte) (string, error) {
	if len(nonce) != nonceSize {
		return "", ErrInvalidEnvelope
	}

	gcm, err := newGCM(encodedKey)
	if err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, nonce, []byte(normalizeSecret(secret)), associatedData(accountID))
	payload := make([]byte, 0, len(nonce)+len(sealed))
	payload = append(payload, nonce...)
	payload = append(payload, sealed...)
	return envelopeVersion + ":" + base64.StdEncoding.EncodeToString(payload), nil
}

func DecryptSecret(envelope string, encodedKey string, accountID uint32) (string, error) {
	parts := strings.SplitN(envelope, ":", 2)
	if len(parts) != 2 || parts[0] != envelopeVersion {
		return "", ErrInvalidEnvelope
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(payload) <= nonceSize {
		return "", ErrInvalidEnvelope
	}

	gcm, err := newGCM(encodedKey)
	if err != nil {
		return "", err
	}

	plain, err := gcm.Open(nil, payload[:nonceSize], payload[nonceSize:], associatedData(accountID))
	if err != nil {
		return "", ErrInvalidEnvelope
	}

	secret := normalizeSecret(string(plain))
	if _, err := decodeSecret(secret); err != nil {
		return "", err
	}
	return secret, nil
}

func MatchingStep(secret, token string, unixTime int64) (int64, bool) {
	if len(token) != tokenDigits {
		return 0, false
	}
	for _, digit := range token {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}

	currentStep := unixTime / periodSeconds
	for _, offset := range []int64{-1, 0, 1} {
		step := currentStep + offset
		if step < 0 {
			continue
		}
		candidate, err := tokenAtStep(secret, uint64(step))
		if err != nil {
			return 0, false
		}
		if hmac.Equal([]byte(candidate), []byte(token)) {
			return step, true
		}
	}
	return 0, false
}

func TokenAtTime(secret string, unixTime int64) (string, error) {
	if unixTime < 0 {
		return "", ErrInvalidSecret
	}
	return tokenAtStep(secret, uint64(unixTime/periodSeconds))
}

func tokenAtStep(secret string, step uint64) (string, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}

	counter := make([]byte, 8)
	for i := len(counter) - 1; i >= 0; i-- {
		counter[i] = byte(step)
		step >>= 8
	}

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(counter)
	digest := mac.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	value := (uint32(digest[offset])&0x7f)<<24 |
		uint32(digest[offset+1])<<16 |
		uint32(digest[offset+2])<<8 |
		uint32(digest[offset+3])

	value %= 1_000_000
	return fmt.Sprintf("%0"+strconv.Itoa(tokenDigits)+"d", value), nil
}

func newGCM(encodedKey string) (cipher.AEAD, error) {
	key, err := DecodeEncryptionKey(encodedKey)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func associatedData(accountID uint32) []byte {
	return []byte(fmt.Sprintf("otbr-login-authenticator:%s:%d", envelopeVersion, accountID))
}

func normalizeSecret(secret string) string {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	return strings.TrimRight(secret, "=")
}

func decodeSecret(secret string) ([]byte, error) {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalizeSecret(secret))
	if err != nil || len(decoded) < 20 {
		return nil, ErrInvalidSecret
	}
	return decoded, nil
}
