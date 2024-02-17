// Package database provides functionalities for interacting with the database of the system,
// including operations to fetch and update data about boosted creatures and bosses. This package
// encapsulates all SQL queries and data manipulations, making maintenance and future development
// easier.
package database

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentibiabr/login-server/src/logger"
)

type events struct {
	XMLName xml.Name `xml:"events"`
	Events  []event  `xml:"event"`
}

type description struct {
	Text string `xml:"description,attr"`
}

type event struct {
	Name        string      `xml:"name,attr"`
	StartDate   string      `xml:"startdate,attr"`
	EndDate     string      `xml:"enddate,attr"`
	Colors      colors      `xml:"colors"`
	Description description `xml:"description"`
	Details     details     `xml:"details"`
}

type colors struct {
	Light string `xml:"colorlight,attr"`
	Dark  string `xml:"colordark,attr"`
}

type details struct {
	DisplayPriority string `xml:"displaypriority,attr"`
	IsSeasonal      string `xml:"isseasonal,attr"`
	SpecialEvent    string `xml:"specialevent,attr"`
}

func loadEventsSchedule(filePath string) (*events, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		logger.Error(fmt.Errorf(err.Error()))
		return nil, err
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		logger.Error(fmt.Errorf(err.Error()))
		return nil, err
	}

	var events events
	err = xml.Unmarshal(byteValue, &events)
	if err != nil {
		logger.Error(fmt.Errorf(err.Error()))
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Unmarshal from XML done successfully, %d events loaded.", len(events.Events)))
	return &events, nil
}

func parseDateString(dateStr string) int {
	layouts := []string{"02/01/2006", "2/01/2006", "02/1/2006", "2/1/2006"}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.ParseInLocation(layout, dateStr, time.Local)
		if err == nil {
			return int(t.Unix())
		}
	}
	logger.Error(fmt.Errorf(err.Error()))
	return 0
}

func processEvents(events *events) []map[string]interface{} {
	eventList := make([]map[string]interface{}, 0)

	for _, event := range events.Events {
		displayPriority, _ := strconv.Atoi(event.Details.DisplayPriority)
		isSeasonal, err := strconv.ParseBool(event.Details.IsSeasonal)
		if err != nil {
			isSeasonal = false
		}
		specialEvent, err := strconv.ParseBool(event.Details.SpecialEvent)
		if err != nil {
			specialEvent = false
		}
		eventMap := map[string]interface{}{
			"colorlight":      event.Colors.Light,
			"colordark":       event.Colors.Dark,
			"description":     event.Description.Text,
			"displaypriority": displayPriority,
			"enddate":         parseDateString(event.EndDate),
			"isseasonal":      isSeasonal,
			"name":            event.Name,
			"startdate":       parseDateString(event.StartDate),
			"specialevent":    specialEvent,
		}
		eventList = append(eventList, eventMap)
	}

	return eventList
}

// HandleEventSchedule loads and processes an event schedule from a specified XML file.
// It takes a Gin context and a string path to the event XML file as arguments.
// If the event schedule is loaded and processed successfully, it sends back a JSON response
// with the list of events and the last update timestamp.
// If there is an error loading or processing the event schedule, it responds with an internal server error.
func HandleEventSchedule(c *gin.Context, eventPath string) {
	events, err := loadEventsSchedule(eventPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	eventList := processEvents(events)
	c.JSON(http.StatusOK, gin.H{
		"eventlist":           eventList,
		"lastupdatetimestamp": time.Now().Unix(),
	})
}
