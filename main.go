package main

import (
	"flag"
	"log"
	"net/http"
)

type Envelope struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

var (
	fhaasAuthEndpoint string
	logWarn           *log.Logger
	logDebug          *log.Logger
	logError          *log.Logger
)

func selfAuth(w http.ResponseWriter, r *http.Request) {
	respDate := map[string]interface{}{}
	data := Envelope{
		Message: "Self authorization endpoint",
		Data:    respDate,
	}
	respond(data, w, http.StatusOK)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tt := map[string]interface{}{
		"docs": "future-link",
	}
	data := Envelope{
		Message: "FhaaS - File handling as a service",
		Data:    tt,
	}
	logDebug.Printf(logRequest(r))

	respond(data, w, 200)
}

func main() {
	SetupCloseHandler()
	// execution arguments setting
	configureLogger()
	logDebug.Println("Starting FhaaS")
	authPtr := flag.String("authurl", "", "Default authentication url")
	flag.Parse()
	fhaasAuthEndpoint = *authPtr
	if fhaasAuthEndpoint == "" {
		logWarn.Println("Flag authurl not set. Application will use FHAAS_AUTH_URL environment variable")
	}
	handleRequests()
	logDebug.Println("Stopping FhaaS")
}
