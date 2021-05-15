package api

import (
	"encoding/json"
	"github.com/opentibiabr/login-server/src/logger"
	"net/http"
)

func logAndRespond(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	logger.LogRequest(r, code, payload, "OK")
	respondWithJSON(w, code, payload)
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
