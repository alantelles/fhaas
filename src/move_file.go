package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type FileMoveBody struct {
	FileIn    string `json:"file_in"`
	FileOut   string `json:"file_out"`
	Overwrite bool   `json:"overwrite"`
}

func moveFile(reqId string, fileMoveSettings FileMoveBody) (int, error) {
	logDebug.Printf("%s - Starting move.", reqId)
	logDebug.Printf("%s - FileIn: %s", reqId, fileMoveSettings.FileIn)
	logDebug.Printf("%s - FileOut: %s", reqId, fileMoveSettings.FileOut)
	logDebug.Printf("%s - Overwrite: %v", reqId, fileMoveSettings.Overwrite)

	destExists := fileExists(fileMoveSettings.FileOut)
	if fileMoveSettings.Overwrite || !destExists {
		orig, err := os.Open(fileMoveSettings.FileIn)
		if err != nil {
			return 0, err
		}
		about, err := os.Stat(fileMoveSettings.FileIn)
		modTime := about.ModTime()
		nowTime := time.Now().Local()
		if err != nil {
			return 0, err
		}

		new, err := os.Create(fileMoveSettings.FileOut)
		if err != nil {
			return 0, err
		}
		defer new.Close()

		written, err := io.Copy(new, orig)
		if err != nil {
			os.Remove(fileMoveSettings.FileOut)
			return 0, err
		}
		os.Chtimes(fileMoveSettings.FileOut, nowTime, modTime)
		logDebug.Printf("%s - File %s moved to %s successfully\n", reqId, fileMoveSettings.FileIn, fileMoveSettings.FileOut)
		orig.Close()
		err = os.Remove(fileMoveSettings.FileIn)
		logDebug.Printf("%s - Removing %s source file\n", reqId, fileMoveSettings.FileIn)
		if err != nil {
			return 0, err
		}
		return int(written), err
	} else {
		logDebug.Printf("%s - Destination file \"%s\" exists. No operations done\n", reqId, fileMoveSettings.FileOut)
		return -1, nil
	}

}

func moveInterfaceSync(reqId string, fileMoveSettings FileMoveBody) (Envelope, int) {
	nowThreads += 1
	var status int
	written, err := moveFile(reqId, fileMoveSettings)
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileMoveSettings,
	}
	env := Envelope{
		Data:      data,
		RequestId: dropReq(reqId),
	}
	if err != nil {
		logError.Printf("While processing move on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		if written > -1 {
			env.Message = "File moved successfully"
			status = http.StatusCreated
		} else {
			env.Message = "File not moved due to already existing and overwrite was set to false"
			status = http.StatusOK
		}

	}
	nowThreads -= 1
	return env, status
}

func moveAsyncWrapper(reqId string, fileMoveSettings FileMoveBody, sendStatusTo string, sendStatusAuth string) {
	nowThreads += 1
	var status int
	written, err := moveFile(reqId, fileMoveSettings)
	env := Envelope{RequestId: reqId}
	if err != nil {
		logError.Printf("Error while processing move on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		if written > -1 {
			env.Message = "File moved successfully"
			status = http.StatusCreated
		} else {
			env.Message = "File not moved due to already existing and overwrite was set to false"
			status = http.StatusOK
		}
	}
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileMoveSettings,
		"status":       status,
	}
	env.Data = data
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
		req.Header.Set("Content-Type", "application/json")
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

func moveInterfaceASync(reqId string, fileMoveSettings FileMoveBody, sendStatusTo string, sendStatusAuth string) (Envelope, int) {
	go moveAsyncWrapper(reqId, fileMoveSettings, sendStatusTo, sendStatusAuth)
	data := map[string]interface{}{
		"body": fileMoveSettings,
	}
	env := Envelope{
		Message:   "Move process started",
		Data:      data,
		RequestId: reqId,
	}
	return env, http.StatusAccepted
}

func moveFileHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)

	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Printf("%s - Starting move process handling\n", reqId)
	logDebug.Printf("%s - Request body: %s\n", reqId, string(reqBody))
	var fileMoveSettings FileMoveBody
	err := json.Unmarshal(reqBody, &fileMoveSettings)
	if err != nil {
		logError.Printf("%s - Fhaas received a malformed request: %s", reqId, err)
		env, status = createBadRequestResponse(reqBody)
	} else {
		if isSyncRequest(reqId, r) {
			env, status = moveInterfaceSync(reqId, fileMoveSettings)
		} else {
			sendStatusTo := r.Header.Get(H_SEND_STATUS_TO)
			sendStatusToAuth := r.Header.Get(H_SEND_STATUS_TO_AUTH)
			logDebug.Printf("%s - Status will be sent to %s", reqId, showIfNotBlank(sendStatusTo))
			if sendStatusToAuth != "" {
				logDebug.Printf("%s - Status will send the following Authorization header: %s", reqId, showToken(sendStatusToAuth))
			}
			env, status = moveInterfaceASync(
				reqId,
				fileMoveSettings,
				sendStatusTo,
				sendStatusToAuth,
			)
		}
	}

	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Move process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}
