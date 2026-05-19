package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateString(t *testing.T) {
	ddmmyyyy, err := time.ParseInLocation("2/1/2006", "03/04/2024", time.Local)
	assert.NoError(t, err)

	yyyymmdd, err := time.ParseInLocation("2006-01-02", "2024-04-03", time.Local)
	assert.NoError(t, err)

	assert.Equal(t, int(ddmmyyyy.Unix()), parseDateString("03/04/2024"))
	assert.Equal(t, int(yyyymmdd.Unix()), parseDateString("2024-04-03"))
	assert.Equal(t, 0, parseDateString("12/31/2024"))
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
