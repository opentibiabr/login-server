package api

import (
	"awesomeProject/api/models"
	"awesomeProject/configs"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Api struct {
	Router *mux.Router
	DB     *sql.DB
}

type RequestPayload struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	StayLoggedIn bool   `json:"stayloggedin"`
	Type         string `json:"type"`
}

type ErrorPayload struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type ResponsePayload struct {
	PlayData models.PlayData `json:"playdata"`
	Session  models.Session  `json:"session"`
}

func (a *Api) Initialize(configs configs.Configs) {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		configs.DBUser,
		configs.DBPass,
		configs.DBHost,
		configs.DBPort,
		configs.DBName,
	)

	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *Api) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *Api) initializeRoutes() {
	a.Router.HandleFunc("/login", a.login).Methods("GET", "POST", "PUT")
	a.Router.HandleFunc("/login.php", a.login).Methods("GET", "POST", "PUT")
}

func logResponse(r *http.Request, code int, payload interface{}) {
	log.Printf("%s %s %s %d %v\n", r.RemoteAddr, r.Method, r.URL, code, payload)
}

func respondWithError(r *http.Request, w http.ResponseWriter, code int, message string) {
	respondWithJSON(r, w, code, map[string]string{"error": message})
}

func respondWithJSON(r *http.Request, w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)

	switch v := payload.(type) {
	case ErrorPayload:
		logResponse(r, code, "failed login")
	case models.Account:
		logResponse(r, code, "successful login")
	default:
		logResponse(r, code, v)
	}
}

func (a *Api) login(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		var message = "Invalid request payload"

		respondWithError(r, w, http.StatusBadRequest, message)
		return
	}

	if payload.Type != "login" {
		respondWithError(r, w, http.StatusNotImplemented, "Non-login requests are not accepted")
		return
	}

	acc := models.Account{Email: payload.Email, Password: payload.Password}
	if err := acc.Authenticate(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			var response = ErrorPayload{ErrorCode: 3, ErrorMessage: "Account email or password is not correct."}
			respondWithJSON(r, w, http.StatusOK, response)
		default:
			respondWithError(r, w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var premUntil = 0

	if acc.PremDays > 0 {
		premUntil = int(time.Now().UnixNano()/1e6) + acc.PremDays*86400
	}

	session := models.Session{
		EmailCodeRequest:              false,
		FpsTracking:                   false,
		IsPremium:                     acc.PremDays > 0,
		IsReturner:                    false,
		LastLoginTime:                 0,
		OptionTracking:                false,
		PremiumUntil:                  premUntil,
		ReturnerNotification:          false,
		SessionKey:                    fmt.Sprintf("%s\n%s", acc.Email, acc.Password),
		ShowRewardNews:                true,
		Status:                        "active",
		TournamentCyclePhase:          0,
		TournamentTicketPurchaseState: 0,
	}

	world := models.World{
		AntiCheatProtection:        false,
		CurrentTournamentPhase:     0,
		ExternalAddress:            "127.0.0.1",
		ExternalAddressProtected:   "127.0.0.1",
		ExternalAddressUnprotected: "127.0.0.1",
		ExternalPort:               7172,
		ExternalPortProtected:      7172,
		ExternalPortUnprotected:    7172,
		ID:                         0,
		IsTournamentWorld:          false,
		Location:                   "BRA",
		Name:                       "Canary",
		PreviewState:               0,
		PvpType:                    0,
		RestrictedStore:            false,
	}

	characters, err := acc.LoadCharacterList(a.DB)
	if err != nil {
		respondWithError(r, w, http.StatusInternalServerError, err.Error())
		return
	}

	playData := models.PlayData{
		Worlds:     []models.World{world},
		Characters: characters,
	}

	responsePayload := ResponsePayload{
		PlayData: playData,
		Session:  session,
	}

	defer r.Body.Close()
	respondWithJSON(r, w, http.StatusOK, responsePayload)
}
