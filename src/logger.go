package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func configureLogger() {
	//nowTime := time.Now().Format("2006-01-02")
	// logFileName := path.Join("logs", nowTime+".log")
	os.Mkdir("logs", 0777)
	logFileName := "logs/fhaas.log"
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logDebug = log.New(logFile, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logWarn = log.New(logFile, "WARN: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logError = log.New(logFile, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

func logRequest(reqId string, r *http.Request) string {
	referer := r.Header.Get("X-Real-IP") // relying in NGINX proxy convention
	if referer == "" {
		// trying from referer
		referer = r.Referer()
	}
	// none worked
	if referer == "" {
		referer = "(referer not sent)"
	}
	out := fmt.Sprintf("[request] id: %s - %s to %s from %s", reqId, r.Method, r.URL.Path, referer)
	return out
}
