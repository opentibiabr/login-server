package api

import (
	"encoding/json"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
	"net/http"
)

func respondAndLog(w http.ResponseWriter, code int, payload interface{}, fields logrus.Fields) {
	respondWithJSON(w, code, payload)
	logger.LogRequest(code, payload, "OK", fields)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"errors": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err := w.Write(response)
	if err != nil {
		logger.Error(err)
	}
}
