package serviceerrors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminHintDefaultDoesNotHardcodeLogPath(t *testing.T) {
	hint := AdminHint("UNKNOWN_ERROR")

	assert.Contains(t, hint, "configured login-server log output")
	assert.False(t, strings.Contains(hint, "logs/login-server.txt"))
}
