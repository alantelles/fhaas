package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type Envelope struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

var fhaasAuthEndpoint string

func respond(data Envelope, w http.ResponseWriter, status int) {
	dataStr, _ := json.Marshal(data)
	w.WriteHeader(status)
	fmt.Fprintln(w, string(dataStr))
}

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
	// execution arguments setting
	authPtr := flag.String("authurl", "", "Default authentication url")
	flag.Parse()

	fhaasAuthEndpoint = *authPtr
	handleRequests()
}
