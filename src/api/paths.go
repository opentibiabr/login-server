package api

import (
	"path/filepath"
)

func (_api *Api) eventSchedulePath() string {
	return filepath.Join(_api.CorePath, "json", "eventscheduler", "events.json")
}
