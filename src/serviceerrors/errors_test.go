package serviceerrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithHintKeepsMessageImmutable(t *testing.T) {
	err := LoginService(CodeSessionStorageUnavailable, "SESSION_STORAGE_UNAVAILABLE", assert.AnError)

	withHint := WithHint(err, "create the table")
	withSameHint := WithHint(withHint, "create the table")

	assert.Equal(t, err.Message, withHint.Message)
	assert.Equal(t, err.Message, withSameHint.Message)
	assert.Equal(t, "create the table", withHint.Hint)
	assert.Equal(t, "Login service error. Please contact support. Error: SESSION_STORAGE_UNAVAILABLE (LS-2004). Admin hint: create the table", MessageWithHint(withHint))
}
