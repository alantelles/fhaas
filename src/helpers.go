package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

func dropReq(reqId string) string {
	return strings.Replace(reqId, "Request ", "", -1)
}

func shouldSendStatus(sendStatusTo string, reqId string, env Envelope, sendStatusAuth string) {
	if sendStatusTo != "" {
		msgFmt, err := sendOperationStatus(reqId, env, sendStatusTo, sendStatusAuth)
		if err != nil {
			logError.Printf(msgFmt, err)
		}
		logDebug.Println(msgFmt)
	}
}

func sendOperationStatus(reqId string, env Envelope, sendStatusTo, sendStatusAuth string) (string, error) {
	var msgFmt string
	body, _ := json.Marshal(env)
	logDebug.Printf("%s - Status: %s", reqId, string(body))
	logDebug.Printf("%s - Sending status to %s", reqId, sendStatusTo)
	req, err := http.NewRequest("POST", sendStatusTo, bytes.NewReader(body))
	if err != nil {
		msgFmt = "%s - Error while creating request to send status: %v"
		return msgFmt, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FhaaS/Go-http-client/1.1")
	if sendStatusAuth != "" {
		req.Header.Set("Authorization", sendStatusAuth)
	}
	client := createClient(20)
	resp, err := client.Do(req)
	if err != nil {
		msgFmt = "%s - Error while sending operation status: %v"
		return msgFmt, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	respStr := string(respBytes)
	msgFmt = fmt.Sprintf("%s - Status endpoint returned with: %s", reqId, respStr)
	return msgFmt, nil
}

func fileNameAttendFilters(subject string, filters []string) bool {
	for _, filter := range filters {
		matched, err := regexp.MatchString(filter, subject)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func checkQueryParam(query url.Values, param string) (bool, string, []string) {
	contains := false
	value := ""
	composed := query[param]
	if len(composed) > 0 {
		value = composed[0]
		contains = true
	}
	return contains, value, composed
}

func createBadRequestResponse(body []byte) (Envelope, int) {
	var env Envelope
	var status int
	data := map[string]interface{}{
		"body": string(body),
	}
	env = Envelope{
		Message: "Fhaas received a malformed request",
		Data:    data,
	}
	status = http.StatusBadRequest
	return env, status
}

func getRequestId(w http.ResponseWriter) string {
	return "Request " + w.Header().Get(H_REQUEST_ID)
}

// func getAuthUrlUsed(w http.ResponseWriter) string {
// 	return w.Header().Get(H_AUTH_URL_USED)
// }

func getAuthContentType(w http.ResponseWriter) string {
	content := w.Header().Get(H_AUTH_CONTENT_TYPE)
	if content == "" {
		return "application/json"
	}
	return content
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
