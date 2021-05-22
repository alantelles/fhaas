package main

import (
	"log"
	"net/http"
)

type Envelope struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	Status  int                    `json:"status"`
}

// FhaaS util headers names
// TODO: explain each header function
const H_REQUEST_ID = "X-Fhaas-Request-Id"
const H_AUTH_URL = "X-Fhaas-Auth-Url"
const H_AUTH_URL_USED = "X-Fhaas-Auth-Url-Used"
const H_AUTH_TOKEN = "X-Fhaas-Auth-Token"
const H_IS_ASYNC = "X-Fhaas-Async"
const H_AUTH_CONTENT_TYPE = "X-Fhaas-Auth-Content-Type"
const H_SEND_STATUS_TO = "X-Fhaas-Send-Status-To"
const H_SEND_STATUS_TO_AUTH = "X-Fhaas-Send-Status-To-Auth"
const H_SEND_STATUS_TO_AUTH_TYPE = "X-Fhaas-Send-Status-To-Auth-Type"

// FhaaS util environment variables names
const E_FHAAS_AUTH_URL = "FHAAS_AUTH_URL"

var (
	fhaasAuthEndpoint string
	allowLogTokens    bool

	logWarn  *log.Logger
	logDebug *log.Logger
	logError *log.Logger
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

	respond(data, w, 200)
}

func main() {
	SetupCloseHandler()
	// execution arguments setting
	configureLogger()
	logDebug.Println("Starting FhaaS")
	setFlags()
	handleRequests()
	logDebug.Println("Stopping FhaaS")
}