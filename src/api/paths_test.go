package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventSchedulePathReturnsCurrentLayout(t *testing.T) {
	corePath := t.TempDir()

	assert.Equal(t, filepath.Join(corePath, "json", "eventscheduler", "events.json"), (&Api{CorePath: corePath}).eventSchedulePath())
}
