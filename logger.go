package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func configureLogger() {
	nowTime := time.Now().Format("2006-01-02")
	logFileName := path.Join("logs", nowTime+".log")
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logDebug = log.New(logFile, "DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logWarn = log.New(logFile, "WARN: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logError = log.New(logFile, "ERROR: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

func logRequest(w http.ResponseWriter, r *http.Request) string {
	referer := r.Header.Get("X-Real-IP") // relying in NGINX proxy convention
	if referer == "" {
		// trying from referer
		referer = r.Referer()
	}
	// none worked
	if referer == "" {
		referer = "(referer not sent)"
	}
	out := fmt.Sprintf("[request] id: %s - %s to %s from %s", getRequestId(w), r.Method, r.URL.Path, referer)
	return out
}

func showToken(token string) string {
	if allowLogTokens {
		return token
	} else {
		return "[logtokens disabled]"
	}
}
