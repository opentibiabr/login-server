package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEventsScheduleFromJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"events": [{
			"name": "Double XP",
			"startdate": "2024-11-12",
			"enddate": "2024-11-17",
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

func TestParseDateString(t *testing.T) {
	ddmmyyyy, err := time.ParseInLocation("2/1/2006", "03/04/2024", time.Local)
	assert.NoError(t, err)

	yyyymmdd, err := time.ParseInLocation("2006-01-02", "2024-04-03", time.Local)
	assert.NoError(t, err)

	mmddyyyy, err := time.ParseInLocation("1/2/2006", "11/17/2024", time.Local)
	assert.NoError(t, err)

	assert.Equal(t, int(ddmmyyyy.Unix()), parseDateString("03/04/2024"))
	assert.Equal(t, int(yyyymmdd.Unix()), parseDateString("2024-04-03"))
	assert.Equal(t, int(mmddyyyy.Unix()), parseDateString("11/17/2024"))
	assert.Equal(t, 0, parseDateString("31/31/2024"))
}

func TestToInt(t *testing.T) {
	assert.Equal(t, 10, toInt(float64(10)))
	assert.Equal(t, 0, toInt(float64(10.5)))
	assert.Equal(t, 42, toInt("42"))
	assert.Equal(t, 0, toInt("abc"))
}

func TestToBool(t *testing.T) {
	assert.True(t, toBool(true))
	assert.True(t, toBool(float64(1)))
	assert.True(t, toBool("true"))
	assert.True(t, toBool("1"))
	assert.False(t, toBool("0"))
	assert.False(t, toBool("abc"))
}
