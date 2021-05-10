package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func getRequestId(w http.ResponseWriter) string {
	return "Request " + w.Header().Get(H_REQUEST_ID)
}

func getAuthUrlUsed(w http.ResponseWriter) string {
	return w.Header().Get(H_AUTH_URL_USED)
}

func respond(data Envelope, w http.ResponseWriter, status int) {
	dataStr, _ := json.Marshal(data)
	w.WriteHeader(status)
	fmt.Fprintln(w, string(dataStr))
}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logDebug.Println("Ctrl+C pressed in Terminal")
		logDebug.Println("Stopping FhaaS")
		os.Exit(0)
	}()
}
