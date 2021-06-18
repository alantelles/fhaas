package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func moveListInterfaceSync(reqId string, fileListMoveSettings []FileMoveBody) (Envelope, int) {
	nowThreads += 1
	var (
		env         Envelope
		envList     []Envelope
		status      int
		finalStatus int
	)
	finalStatus = 0
	works := len(fileListMoveSettings)
	envList = make([]Envelope, works)
	for i, v := range fileListMoveSettings {
		envList[i], status = moveInterfaceSync(reqId, v)
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
		"body":   fileListMoveSettings,
		"result": envList,
	}
	env.Data = data
	env.Message = "Moves processed"
	env.RequestId = strings.Replace(reqId, "Request ", "", -1)
	nowThreads -= 1
	return env, finalStatus
}

func moveListAsyncWrapper(reqId string, fileListMoveSettings []FileMoveBody, sendStatusTo, sendStatusAuth string) {
	nowThreads += 1
	var (
		env         Envelope
		envList     []Envelope
		status      int
		finalStatus int
	)
	finalStatus = 0
	works := len(fileListMoveSettings)
	envList = make([]Envelope, works)
	for i, v := range fileListMoveSettings {
		envList[i], status = moveInterfaceSync(reqId, v)
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
		"body":   fileListMoveSettings,
		"result": envList,
	}
	env.Data = data
	env.RequestId = dropReq(reqId)
	env.Message = "Moves processed"
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
	nowThreads -= 1
}

func moveListInterfaceASync(reqId string, fileListMoveSettings []FileMoveBody, sendStatusTo, sendStatusAuth string) (Envelope, int) {
	go moveListAsyncWrapper(reqId, fileListMoveSettings, sendStatusTo, sendStatusAuth)
	env := Envelope{
		Message: "Move process started",
		Data: map[string]interface{}{
			"body": fileListMoveSettings,
		},
		Status:    http.StatusAccepted,
		RequestId: dropReq(reqId),
	}
	return env, http.StatusAccepted
}

func moveFileListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)

	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Printf("%s - Starting move process handling\n", reqId)
	logDebug.Printf("%s - Request body: %s\n", reqId, string(reqBody))
	var fileMoveListSettings []FileMoveBody
	err := json.Unmarshal(reqBody, &fileMoveListSettings)
	if err != nil {
		logError.Printf("%s - Fhaas received a malformed request: %s", reqId, err)
		env, status = createBadRequestResponse(reqBody)
	} else {
		if isSyncRequest(reqId, r) {
			env, status = moveListInterfaceSync(reqId, fileMoveListSettings)
		} else {
			sendStatusTo := r.Header.Get(H_SEND_STATUS_TO)
			sendStatusToAuth := r.Header.Get(H_SEND_STATUS_TO_AUTH)
			logDebug.Printf("%s - Status will be sent to %s", reqId, showIfNotBlank(sendStatusTo))
			if sendStatusToAuth != "" {
				logDebug.Printf("%s - Status will send the following Authorization header: %s", reqId, showToken(sendStatusToAuth))
			}
			env, status = moveListInterfaceASync(
				reqId,
				fileMoveListSettings,
				sendStatusTo,
				sendStatusToAuth,
			)
		}
	}
	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Move process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}
