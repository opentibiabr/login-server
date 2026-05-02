package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventSchedulePathPrefersCurrentCanaryLayout(t *testing.T) {
	corePath := t.TempDir()
	currentPath := filepath.Join(corePath, "json", "eventscheduler", "events.json")
	legacyPath := filepath.Join(corePath, "XML", "events.xml")
	require.NoError(t, os.MkdirAll(filepath.Dir(currentPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(legacyPath), 0o755))
	require.NoError(t, os.WriteFile(currentPath, []byte(`{"events":[]}`), 0o644))
	require.NoError(t, os.WriteFile(legacyPath, []byte("<events/>"), 0o644))

	assert.Equal(t, currentPath, (&Api{CorePath: corePath}).eventSchedulePath())
}

func TestEventSchedulePathFallsBackToLegacyLayout(t *testing.T) {
	corePath := t.TempDir()
	legacyPath := filepath.Join(corePath, "XML", "events.xml")
	require.NoError(t, os.MkdirAll(filepath.Dir(legacyPath), 0o755))
	require.NoError(t, os.WriteFile(legacyPath, []byte("<events/>"), 0o644))

	assert.Equal(t, legacyPath, (&Api{CorePath: corePath}).eventSchedulePath())
}

func TestEventSchedulePathReturnsCurrentLayoutWhenMissing(t *testing.T) {
	corePath := t.TempDir()

	assert.Equal(t, filepath.Join(corePath, "json", "eventscheduler", "events.json"), (&Api{CorePath: corePath}).eventSchedulePath())
}
