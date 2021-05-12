package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func logResponse(r *http.Request, code int, payload interface{}) {
	if os.Getenv("LOGIN_SERVER_SILENT") == "true" {
		return
	}
	log.Printf("%s %s %s %d %v\n", r.RemoteAddr, r.Method, r.URL, code, payload)
}

func logErrorAndRespond(w http.ResponseWriter, r *http.Request, code int, message string) {
	logResponse(r, code, message)
	respondWithError(w, r, code, message)
}

func logAndRespond(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	logResponse(r, code, http.StatusText(code))
	respondWithJSON(w, r, code, payload)
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	respondWithJSON(w, r, code, map[string]string{"errors": message})
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
