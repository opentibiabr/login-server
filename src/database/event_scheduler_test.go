package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEventsScheduleFromJson(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"events": [{
			"name": "Double XP",
			"startdate": "11/12/2024",
			"enddate": "11/17/2024",
			"description": "Double experience.",
			"colors": {
				"colordark": "#001122",
				"colorlight": "#334455"
			},
			"details": {
				"displaypriority": 6,
				"isseasonal": 1,
				"specialevent": 0
			}
		}]
	}`), 0o644))

	events, err := loadEventsSchedule(path)
	require.NoError(t, err)

	payload := processEvents(events)
	require.Len(t, payload, 1)
	assert.Equal(t, "Double XP", payload[0]["name"])
	assert.Equal(t, "Double experience.", payload[0]["description"])
	assert.Equal(t, "#001122", payload[0]["colordark"])
	assert.Equal(t, "#334455", payload[0]["colorlight"])
	assert.Equal(t, 6, payload[0]["displaypriority"])
	assert.Equal(t, true, payload[0]["isseasonal"])
	assert.Equal(t, false, payload[0]["specialevent"])
	assert.Equal(t, int(time.Date(2024, 11, 12, 0, 0, 0, 0, time.Local).Unix()), payload[0]["startdate"])
	assert.Equal(t, int(time.Date(2024, 11, 17, 0, 0, 0, 0, time.Local).Unix()), payload[0]["enddate"])
}

func TestNormalizeJsonDateLeavesUnknownFormatUntouched(t *testing.T) {
	assert.Equal(t, "2024-11-12", normalizeJsonDate("2024-11-12"))
}
