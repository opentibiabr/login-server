// Package database provides functionalities for interacting with the database of the system,
// including operations to fetch and update data about boosted creatures and bosses. This package
// encapsulates all SQL queries and data manipulations, making maintenance and future development
// easier.
package database

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/logger"
)

type jsonEvents struct {
	Events []jsonEvent `json:"events"`
}

type jsonEvent struct {
	Name        string      `json:"name"`
	StartDate   string      `json:"startdate"`
	EndDate     string      `json:"enddate"`
	Colors      jsonColors  `json:"colors"`
	Description string      `json:"description"`
	Details     jsonDetails `json:"details"`
}

type jsonColors struct {
	Light string `json:"colorlight"`
	Dark  string `json:"colordark"`
}

type jsonDetails struct {
	DisplayPriority interface{} `json:"displaypriority"`
	IsSeasonal      interface{} `json:"isseasonal"`
	SpecialEvent    interface{} `json:"specialevent"`
}

func loadEventsSchedule(filePath string) (*jsonEvents, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var events jsonEvents
	if err = json.Unmarshal(byteValue, &events); err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Unmarshal from JSON done successfully, %d events loaded.", len(events.Events)))
	return &events, nil
}

func parseDateString(dateStr string) int {
	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"2/1/2006",
		"01/02/2006",
		"1/2/2006",
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.ParseInLocation(layout, dateStr, time.Local)
		if err == nil {
			return int(t.Unix())
		}
	}

	logger.Error(fmt.Errorf("failed to parse date %q: %w", dateStr, err))
	return 0
}

func toInt(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		if math.Trunc(typed) != typed {
			logger.Error(fmt.Errorf("failed to convert %v to int: value is not a whole number", typed))
			return 0
		}
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			logger.Error(fmt.Errorf("failed to convert %q to int: %w", typed, err))
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func toBool(value interface{}) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case float64:
		return typed != 0
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err == nil {
			return parsed
		}

		number, err := strconv.Atoi(typed)
		if err != nil {
			logger.Error(fmt.Errorf("failed to convert %q to bool: %w", typed, err))
			return false
		}
		return number != 0
	default:
		return false
	}
}

func processEvents(events *jsonEvents) []map[string]interface{} {
	eventList := make([]map[string]interface{}, 0, len(events.Events))

	for _, event := range events.Events {
		eventMap := map[string]interface{}{
			"colorlight":      event.Colors.Light,
			"colordark":       event.Colors.Dark,
			"description":     event.Description,
			"displaypriority": toInt(event.Details.DisplayPriority),
			"enddate":         parseDateString(event.EndDate),
			"isseasonal":      toBool(event.Details.IsSeasonal),
			"name":            event.Name,
			"startdate":       parseDateString(event.StartDate),
			"specialevent":    toBool(event.Details.SpecialEvent),
		}
		eventList = append(eventList, eventMap)
	}

	return eventList
}

// HandleEventSchedule loads and processes an event schedule from the provided path.
func HandleEventSchedule(c *gin.Context, eventPath string) {
	events, err := loadEventsSchedule(eventPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventlist":           processEvents(events),
		"lastupdatetimestamp": time.Now().Unix(),
	})
}
