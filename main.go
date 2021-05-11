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

// FhaaS util headers names
const H_REQUEST_ID = "X-Fhaas-Request-Id"
const H_AUTH_URL = "X-Fhaas-Auth-Url"
const H_AUTH_URL_USED = "X-Fhaas-Auth-Url-Used"
const H_AUTH_TOKEN = "X-Fhaas-Auth-Token"
const H_IS_ASYNC = "X-Fhaas-Async"
const H_AUTH_CONTENT_TYPE = "X-Fhaas-Auth-Content-Type"

// FhaaS util environment variables names
const E_FHAAS_AUTH_URL = "FHAAS_AUTH_URL"

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
	logDebug.Printf(logRequest(w, r))

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
		logWarn.Println("Flag authurl not set. Application will use FHAAS_AUTH_URL environment variable (will be checked in every request). If not set, will use " + H_AUTH_URL + " header of request. This may be potentially dangerous since any url able to authorize operation can be used")
	}
	handleRequests()
	logDebug.Println("Stopping FhaaS")
}
