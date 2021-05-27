package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func copyListInterfaceSync(reqId string, fileListCopySettings []FileCopyBody) (Envelope, int) {
	var (
		env         Envelope
		envList     []Envelope
		status      int
		finalStatus int
	)
	finalStatus = 0
	works := len(fileListCopySettings)
	envList = make([]Envelope, works)
	for i, v := range fileListCopySettings {
		envList[i], status = copyInterfaceSync(reqId, v)
		envList[i].Status = status
		if finalStatus == 0 {
			finalStatus = status
		} else {
			if status != 201 && finalStatus == 201 {
				finalStatus = 207
			}
			if status == 201 && (finalStatus != 207 && finalStatus != 201) {
				finalStatus = 207
			}
		}
	}
	data := map[string]interface{}{
		"body":   fileListCopySettings,
		"result": envList,
	}
	env.Data = data
	env.Message = "Copies processed"
	env.RequestId = strings.Replace(reqId, "Request ", "", -1)
	return env, finalStatus
}

func copyListAsyncWrapper(reqId string, fileListCopySettings []FileCopyBody, sendStatusTo, sendStatusAuth string) {
	var (
		env         Envelope
		envList     []Envelope
		status      int
		finalStatus int
	)
	finalStatus = 0
	works := len(fileListCopySettings)
	envList = make([]Envelope, works)
	for i, v := range fileListCopySettings {
		envList[i], status = copyInterfaceSync(reqId, v)
		envList[i].Status = status
		if finalStatus == 0 {
			finalStatus = status
		} else {
			if status != 201 && finalStatus == 201 {
				finalStatus = 207
			}
			if status == 201 && finalStatus != 207 {
				finalStatus = 207
			}
		}
	}
	data := map[string]interface{}{
		"body":   fileListCopySettings,
		"result": envList,
	}
	env.Data = data
	env.RequestId = dropReq(reqId)
	env.Message = "Copies processed"
	if sendStatusTo != "" {
		body, _ := json.Marshal(env)
		logDebug.Printf("%s - Status: %s", reqId, string(body))
		logDebug.Printf("%s - Sending status to %s", reqId, sendStatusTo)
		req, err := http.NewRequest("POST", sendStatusTo, bytes.NewReader(body))
		if err != nil {
			logError.Printf("%s - Error while creating request to send status: %v", reqId, err)
		}
		if sendStatusAuth != "" {
			req.Header.Set("Authorization", sendStatusAuth)
		}
		client := createClient(20)
		resp, err := client.Do(req)
		if err != nil {
			logError.Printf("%s - Error while sending operation status: %v", reqId, err)
		}
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(resp.Body)
		respStr := string(respBytes)
		logDebug.Printf("%s - Status endpoint returned with: %s", reqId, respStr)
	}
	// return env, finalStatus
}

func copyListInterfaceASync(reqId string, fileListCopySettings []FileCopyBody, sendStatusTo, sendStatusAuth string) (Envelope, int) {
	go copyListAsyncWrapper(reqId, fileListCopySettings, sendStatusTo, sendStatusAuth)
	env := Envelope{
		Message: "Copy process started",
		Data: map[string]interface{}{
			"body": fileListCopySettings,
		},
		Status:    http.StatusAccepted,
		RequestId: dropReq(reqId),
	}
	return env, http.StatusAccepted
}

func copyFileListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)

	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Printf("%s - Starting copy process handling\n", reqId)
	logDebug.Printf("%s - Request body: %s\n", reqId, string(reqBody))
	var fileCopyListSettings []FileCopyBody
	err := json.Unmarshal(reqBody, &fileCopyListSettings)
	if err != nil {
		logError.Printf("%s - Fhaas received a malformed request: %s", reqId, err)
		env, status = createBadRequestResponse(reqBody)
	} else {
		if isSyncRequest(reqId, r) {
			env, status = copyListInterfaceSync(reqId, fileCopyListSettings)
		} else {
			sendStatusTo := r.Header.Get(H_SEND_STATUS_TO)
			sendStatusToAuth := r.Header.Get(H_SEND_STATUS_TO_AUTH)
			logDebug.Printf("%s - Status will be sent to %s", reqId, showIfNotBlank(sendStatusTo))
			if sendStatusToAuth != "" {
				logDebug.Printf("%s - Status will send the following Authorization header: %s", reqId, showToken(sendStatusToAuth))
			}
			env, status = copyListInterfaceASync(
				reqId,
				fileCopyListSettings,
				sendStatusTo,
				sendStatusToAuth,
			)
		}
	}
	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Copy process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}
