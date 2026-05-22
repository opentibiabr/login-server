package logger

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitWritesErrorsToConfiguredLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "login-server.txt")
	t.Cleanup(func() {
		Init(log.InfoLevel)
	})

	Init(log.DebugLevel, logPath)
	Error(errors.New("persisted diagnostic error"))

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "persisted diagnostic error")
}

func TestWithFieldsWritesToConfiguredLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "login-server.txt")
	t.Cleanup(func() {
		Init(log.InfoLevel)
	})

	Init(log.InfoLevel, logPath)
	WithFields(log.Fields{"source": "test"}).Info("persisted field diagnostic")

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "persisted field diagnostic")
	assert.Contains(t, string(content), "[test]")
}
