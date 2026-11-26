package passwords

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const utf8Password = "Sênha çom ação #1"
const legacySHA1 = "7c4a8d09ca3762af61e59520943dc26494f8941b"

type vector struct {
	password string
	stored   string
}

// Half written by PHP password_hash(), half by golang.org/x/crypto/argon2.
var argon2idVectors = []vector{
	{"123456", "$argon2id$v=19$m=65536,t=3,p=1$L0toZHlaNFNHMlh1NTFEQg$UyY/BT/lQlf/EpFvNtJn/usQFi3APme7vimpMfJh2Fw"},
	{utf8Password, "$argon2id$v=19$m=65536,t=3,p=1$OXhTMnBJUGw1aElWeGcvSg$IqBlzGjFDABGfDm35M5Ch54Qt/yJDlbLv8CSoFo6e1g"},
	{"123456", "$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$icckntc5Bk4fxs9C/0ALeRJmTMinHqsrSoO/ByU5R+E"},
	{utf8Password, "$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$GVwVfi90wPcmSjNfhvPj9n8EgS/8lspzdNo5QGjGtvk"},
	{"123456", "$argon2id$v=19$m=1024,t=2,p=1$VkdrdjgwcWVRYmwuNzRjSg$3FemK5cMa5yYB3kpk++7cFO73TnP7PazDQCaXHBfqN8"},
}

var wrappedVectors = []vector{
	{"123456", "sha1:$argon2id$v=19$m=65536,t=3,p=1$QTNJa2w4U0hQb2VSc1RDNQ$0Tj2i2Lq5+z9b64K+G0j/7hJGeu9GzYJ0KtfXQU+HH0"},
	{utf8Password, "sha1:$argon2id$v=19$m=65536,t=3,p=1$dUtSbjd6YnAuTFV5UUxTUQ$StjmIZbZ0pJS+8PGq9/lyrzr4ArR9s31Tby6xp9vSyQ"},
	{"123456", "sha1:$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$LxBtEz7HhVvFglsxIBqqsAn7sGjAuIf8c7y87lUgr+s"},
	{utf8Password, "sha1:$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$c0yhKdVXVsh9q89v9UtiS4u/OH1Y1VBZpk0uvBaghoo"},
}

func assertVerify(t *testing.T, want bool, password, stored string) {
	t.Helper()
	got, err := NewHasher(1, false).Verify(context.Background(), password, stored)
	require.NoError(t, err, stored)
	assert.Equal(t, want, got, "password %q against %q", password, stored)
}

func TestVerifyArgon2id(t *testing.T) {
	for _, v := range argon2idVectors {
		assertVerify(t, true, v.password, v.stored)
		assertVerify(t, false, v.password+" ", v.stored)
		assertVerify(t, false, v.stored, v.stored)
		assertVerify(t, false, sha1Hex(v.password), v.stored)
	}
}

func TestVerifyWrappedSHA1(t *testing.T) {
	for _, v := range wrappedVectors {
		assertVerify(t, true, v.password, v.stored)
		assertVerify(t, false, v.password+" ", v.stored)
		// Someone who only has the former SHA-1 value cannot log in with it.
		assertVerify(t, false, sha1Hex(v.password), v.stored)
		assertVerify(t, false, v.password, strings.TrimPrefix(v.stored, wrappedMarker))
	}
}

func TestVerifyLegacySHA1(t *testing.T) {
	assertVerify(t, true, "123456", legacySHA1)
	assertVerify(t, true, "123456", strings.ToUpper(legacySHA1))
	assertVerify(t, true, "123456", legacySHA1+"  ")
	assertVerify(t, false, "123456", " "+legacySHA1)
	assertVerify(t, false, "1234567", legacySHA1)
	assertVerify(t, false, legacySHA1, legacySHA1)
}

func TestVerifyUnknownValuesMatchNothing(t *testing.T) {
	for _, stored := range []string{"", "123456", "locked-1", "sha1:", "sha1:" + legacySHA1, "$2y$10$abcdefghijklmnopqrstuv", "$argon2i$v=19$m=1024,t=2,p=1$c2FsdHNhbHQ$aGFzaGhhc2hoYXNoaGFzaA"} {
		assertVerify(t, false, "123456", stored)
		assertVerify(t, false, "", stored)
	}
}

func TestVerifyRejectsMalformedAndOutOfBoundsValues(t *testing.T) {
	salt := "$Zml4ZWQtc2FsdC0xNmJ5dA$icckntc5Bk4fxs9C/0ALeRJmTMinHqsrSoO/ByU5R+E"
	for _, stored := range []string{
		"$argon2id$",
		"$argon2id$v=16$m=65536,t=3,p=1" + salt,
		"$argon2id$v=19$m=65536,t=0,p=1" + salt,
		"$argon2id$v=19$m=65536,t=3,p=0" + salt,
		"$argon2id$v=19$m=65536,t=17,p=1" + salt,
		"$argon2id$v=19$m=65536,t=3,p=17" + salt,
		"$argon2id$v=19$m=262145,t=3,p=1" + salt,
		"$argon2id$v=19$m=4294967295,t=3,p=1" + salt,
		"$argon2id$v=19$m=4,t=3,p=1" + salt,
		"$argon2id$v=19$m=65536,t=3,p=1,keyid=1" + salt,
		"$argon2id$v=19$m=65536,t=3,p=1$short$icckntc5Bk4fxs9C/0ALeRJmTMinHqsrSoO/ByU5R+E",
		"$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$",
		"$argon2id$v=19$m=65536,t=3,p=1$Zml4ZWQtc2FsdC0xNmJ5dA$***",
		"sha1:$argon2id$v=19$m=65536,t=3,p=1",
	} {
		matches, err := NewHasher(1, false).Verify(context.Background(), "123456", stored)
		assert.False(t, matches, stored)
		assert.ErrorIs(t, err, ErrMalformed, stored)
	}
}

func TestHashWritesACurrentValue(t *testing.T) {
	hasher := NewHasher(1, true)

	first, err := hasher.Hash(context.Background(), utf8Password)
	require.NoError(t, err)
	second, err := hasher.Hash(context.Background(), utf8Password)
	require.NoError(t, err)

	assert.Regexp(t, `^\$argon2id\$v=19\$m=65536,t=3,p=1\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`, first)
	assert.NotEqual(t, first, second, "each value has its own salt")
	assert.False(t, NeedsRehash(first))
	assertVerify(t, true, utf8Password, first)
	assertVerify(t, false, "123456", first)
}

func TestNeedsRehash(t *testing.T) {
	salt := "$Zml4ZWQtc2FsdC0xNmJ5dA$icckntc5Bk4fxs9C/0ALeRJmTMinHqsrSoO/ByU5R+E"
	assert.True(t, NeedsRehash(legacySHA1))
	assert.True(t, NeedsRehash(wrappedVectors[0].stored))
	assert.True(t, NeedsRehash("$argon2id$v=19$m=1024,t=3,p=1"+salt))
	assert.True(t, NeedsRehash("$argon2id$v=19$m=65536,t=2,p=1"+salt))
	assert.False(t, NeedsRehash("$argon2id$v=19$m=65536,t=3,p=1"+salt))
	assert.False(t, NeedsRehash("$argon2id$v=19$m=131072,t=4,p=1"+salt), "stronger values are kept")
	assert.False(t, NeedsRehash("$argon2id$v=19$m=131072,t=2,p=1"+salt), "more memory is kept with fewer passes")
	assert.False(t, NeedsRehash("$argon2id$v=19$m=65536,t=4,p=1"+salt))
	assert.True(t, NeedsRehash("$argon2id$v=19$m=32768,t=8,p=1"+salt), "less memory is replaced")
}

func TestBusyHasherDoesNotAnswer(t *testing.T) {
	hasher := NewHasher(1, false)
	hasher.slots <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	matches, err := hasher.Verify(ctx, "123456", argon2idVectors[0].stored)
	assert.False(t, matches)
	assert.ErrorIs(t, err, ErrBusy)
	_, err = hasher.Hash(ctx, "123456")
	assert.ErrorIs(t, err, ErrBusy)
	assert.ErrorIs(t, hasher.Waste(ctx, "123456"), ErrBusy)

	// A legacy value needs no hashing slot.
	matches, err = hasher.Verify(ctx, "123456", legacySHA1)
	assert.True(t, matches)
	assert.NoError(t, err)
}

func TestNewHasherDefaults(t *testing.T) {
	assert.Equal(t, DefaultConcurrency, cap(NewHasher(0, false).slots))
	assert.Equal(t, 5, cap(NewHasher(5, false).slots))
	assert.False(t, NewHasher(1, false).RehashEnabled())
	assert.True(t, NewHasher(1, true).RehashEnabled())
}
