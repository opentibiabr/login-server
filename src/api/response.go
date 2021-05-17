package api

import (
	"encoding/json"
	"github.com/opentibiabr/login-server/src/api/models"
	"github.com/opentibiabr/login-server/src/logger"
	"github.com/sirupsen/logrus"
	"net/http"
)

func processErrorResponse(
	w http.ResponseWriter,
	loginError models.LoginErrorPayload,
	fields logrus.Fields,
) {
	logger.LogRequest(
		http.StatusOK,
		loginError,
		"unsuccessful login",
		fields,
	)

	respondWithJSON(w, http.StatusOK, loginError)
}

func respondAndLog(w http.ResponseWriter, code int, payload interface{}, fields logrus.Fields) {
	respondWithJSON(w, code, payload)
	logger.LogRequest(code, payload, "OK", fields)
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	logger.Error(err)
	respondWithJSON(w, code, map[string]string{"errors": err.Error()})
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