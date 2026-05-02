package api

import (
	"os"
	"path/filepath"
)

func (_api *Api) eventSchedulePath() string {
	candidates := []string{
		filepath.Join(_api.CorePath, "json", "eventscheduler", "events.json"),
		filepath.Join(_api.CorePath, "XML", "events.xml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return candidates[0]
}
