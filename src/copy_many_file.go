package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func copyListInterfaceSync(reqId string, fileListCopySettings []FileCopyBody) (Envelope, int) {
	nowThreads += 1
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
	env.Status = finalStatus
	env.Message = "Copies processed"
	env.RequestId = strings.Replace(reqId, "Request ", "", -1)
	nowThreads -= 1
	return env, finalStatus
}

func copyListAsyncWrapper(reqId string, fileListCopySettings []FileCopyBody, sendStatusTo, sendStatusAuth string) {
	nowThreads += 1
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
			if status == 201 && finalStatus == 201 {
				finalStatus = 201
			} else {
				if status != 201 && finalStatus == 201 {
					finalStatus = 207
				}
				if status == 201 && finalStatus != 207 {
					finalStatus = 207
				}
			}

		}
	}
	data := map[string]interface{}{
		"body":   fileListCopySettings,
		"result": envList,
	}
	env.Data = data
	env.RequestId = dropReq(reqId)
	env.Status = finalStatus
	env.Message = "Copies processed"
	shouldSendStatus(sendStatusTo, reqId, env, sendStatusAuth)
	// return env, finalStatus
	nowThreads -= 1
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
