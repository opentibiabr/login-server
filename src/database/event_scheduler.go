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

type Events struct {
	XMLName xml.Name `xml:"events"`
	Events  []Event  `xml:"event"`
}

type Description struct {
	Text string `xml:"description,attr"`
}

type Event struct {
	Name        string      `xml:"name,attr"`
	StartDate   string      `xml:"startdate,attr"`
	EndDate     string      `xml:"enddate,attr"`
	Colors      Colors      `xml:"colors"`
	Description Description `xml:"description"`
	Details     Details     `xml:"details"`
}

type Colors struct {
	Light string `xml:"colorlight,attr"`
	Dark  string `xml:"colordark,attr"`
}

type Details struct {
	DisplayPriority string `xml:"displaypriority,attr"`
	IsSeasonal      string `xml:"isseasonal,attr"`
	SpecialEvent    string `xml:"specialevent,attr"`
}

func loadEventsSchedule(filePath string) (*Events, error) {
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

	var events Events
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

func processEvents(events *Events) []map[string]interface{} {
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
