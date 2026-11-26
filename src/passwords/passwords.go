// Package passwords verifies and writes the values stored in accounts.password.
//
// Three formats are accepted, so accounts can be upgraded one at a time without locking anyone out:
//
//   - "$argon2id$v=19$m=...,t=...,p=...$salt$hash": Argon2id of the password, the format written by PHP's
//     password_hash and by libargon2. The parameters are read from the value itself.
//   - "sha1:$argon2id$...": Argon2id of the lowercase hexadecimal SHA-1 of the password. It is a former SHA-1 value
//     wrapped without knowing the password.
//   - 40 hexadecimal characters: the legacy SHA-1 of the password.
package passwords

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Memory (KiB), Time and Threads are the Argon2id parameters of new values. The account website has to write
	// values at least as strong, otherwise each login upgrades them again.
	Memory  uint32 = 64 * 1024
	Time    uint32 = 3
	Threads uint8  = 1

	// DefaultConcurrency bounds how many Argon2id computations run at once. Each one uses the memory of the value it
	// computes, Memory KiB for new values.
	DefaultConcurrency = 2
	// MinColumnLength is the narrowest accounts.password column which holds a new value, with room for the marker of a
	// wrapped one.
	MinColumnLength = 128

	argon2idPrefix   = "$argon2id$"
	wrappedMarker    = "sha1:"
	keyLength        = 32
	saltLength       = 16
	sha1HexLength    = 2 * sha1.Size
	minSaltLength    = 8
	maxSaltLength    = 64
	minKeyLength     = 16
	maxKeyLength     = 64
	maxMemory        = 256 * 1024
	maxTime          = 16
	maxThreads       = 16
	minMemoryPerLane = 8
)

var (
	// ErrBusy reports that no hashing slot became free before the context ended. The password was not checked.
	ErrBusy = errors.New("password hashing is busy")
	// ErrMalformed reports a stored Argon2id value that cannot be used.
	ErrMalformed = errors.New("stored argon2id value is malformed or out of bounds")

	parametersPattern = regexp.MustCompile(`^m=([1-9][0-9]{0,9}),t=([1-9][0-9]{0,9}),p=([1-9][0-9]{0,2})$`)
)

// Hasher verifies and writes password values with a bounded number of simultaneous Argon2id computations.
type Hasher struct {
	slots  chan struct{}
	rehash bool
}

// NewHasher returns a Hasher. rehash enables upgrading stored values after a successful login; keep it disabled until
// everything that reads accounts.password accepts Argon2id.
func NewHasher(concurrency int, rehash bool) *Hasher {
	if concurrency < 1 {
		concurrency = DefaultConcurrency
	}
	return &Hasher{slots: make(chan struct{}, concurrency), rehash: rehash}
}

// RehashEnabled reports whether stored values are upgraded after a successful login.
func (h *Hasher) RehashEnabled() bool {
	return h.rehash
}

// Verify reports whether password matches the stored value. A value in no known format matches nothing.
func (h *Hasher) Verify(ctx context.Context, password, stored string) (bool, error) {
	switch {
	case strings.HasPrefix(stored, wrappedMarker+argon2idPrefix):
		return h.verifyArgon2id(ctx, sha1Hex(password), strings.TrimPrefix(stored, wrappedMarker))
	case strings.HasPrefix(stored, argon2idPrefix):
		return h.verifyArgon2id(ctx, password, stored)
	default:
		// Databases compare text ignoring letter case and trailing spaces, and legacy values were matched that way.
		legacy := strings.ToLower(strings.TrimRight(stored, " "))
		if len(legacy) != sha1HexLength {
			return false, nil
		}
		return subtle.ConstantTimeCompare([]byte(sha1Hex(password)), []byte(legacy)) == 1, nil
	}
}

// IsArgon2id reports whether checking a password against the stored value computes Argon2id.
func IsArgon2id(stored string) bool {
	return strings.HasPrefix(strings.TrimPrefix(stored, wrappedMarker), argon2idPrefix)
}

// NeedsRehash reports whether a stored value that matched a password should be replaced by a current Argon2id value.
// An Argon2id value is replaced when it uses less memory, or the same memory and fewer passes. One with more memory
// is kept whatever its passes.
func NeedsRehash(stored string) bool {
	value, err := parse(stored)
	if err != nil {
		return true
	}
	return value.memory < Memory || (value.memory == Memory && value.time < Time)
}

// Hash returns a new Argon2id value for password.
func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := h.derive(ctx, password, salt, Time, Memory, Threads, keyLength)
	if err != nil {
		return "", err
	}
	encoding := base64.RawStdEncoding
	return fmt.Sprintf("%sv=%d$m=%d,t=%d,p=%d$%s$%s", argon2idPrefix, argon2.Version, Memory, Time, Threads,
		encoding.EncodeToString(salt), encoding.EncodeToString(key)), nil
}

// Waste spends the time of one verification. It is used when there is no stored value to compare with, so that an
// unknown account does not answer faster than a wrong password.
func (h *Hasher) Waste(ctx context.Context, password string) error {
	_, err := h.derive(ctx, password, make([]byte, saltLength), Time, Memory, Threads, keyLength)
	return err
}

func (h *Hasher) verifyArgon2id(ctx context.Context, secret, encoded string) (bool, error) {
	value, err := parse(encoded)
	if err != nil {
		return false, err
	}
	key, err := h.derive(ctx, secret, value.salt, value.time, value.memory, value.threads, uint32(len(value.key)))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(key, value.key) == 1, nil
}

func (h *Hasher) derive(ctx context.Context, secret string, salt []byte, time, memory uint32, threads uint8, length uint32) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ErrBusy
	}
	select {
	case h.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ErrBusy
	}
	defer func() { <-h.slots }()
	key := argon2.IDKey([]byte(secret), salt, time, memory, threads, length)
	// Release the working memory before the slot is handed over. Left to its own pace, the collector keeps finished
	// computations around and the process needs twice the memory the slots allow.
	runtime.GC()
	return key, nil
}

type argon2idValue struct {
	memory  uint32
	time    uint32
	threads uint8
	salt    []byte
	key     []byte
}

// parse reads "$argon2id$v=19$m=..,t=..,p=..$salt$hash". The bounds keep a damaged value from exhausting the memory
// or panicking the process.
func parse(encoded string) (argon2idValue, error) {
	var value argon2idValue
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return value, ErrMalformed
	}
	parameters := parametersPattern.FindStringSubmatch(parts[3])
	if parameters == nil {
		return value, ErrMalformed
	}
	memory, memoryErr := strconv.ParseUint(parameters[1], 10, 32)
	time, timeErr := strconv.ParseUint(parameters[2], 10, 32)
	threads, threadsErr := strconv.ParseUint(parameters[3], 10, 8)
	if memoryErr != nil || timeErr != nil || threadsErr != nil || threads > maxThreads || time > maxTime ||
		memory > maxMemory || memory < minMemoryPerLane*threads {
		return value, ErrMalformed
	}
	encoding := base64.RawStdEncoding.Strict()
	salt, saltErr := encoding.DecodeString(parts[4])
	key, keyErr := encoding.DecodeString(parts[5])
	if saltErr != nil || keyErr != nil || len(salt) < minSaltLength || len(salt) > maxSaltLength ||
		len(key) < minKeyLength || len(key) > maxKeyLength {
		return value, ErrMalformed
	}
	return argon2idValue{memory: uint32(memory), time: uint32(time), threads: uint8(threads), salt: salt, key: key}, nil
}

func sha1Hex(password string) string {
	sum := sha1.Sum([]byte(password))
	return hex.EncodeToString(sum[:])
}
