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

type FileCopyBody struct {
	FileIn    string `json:"file_in"`
	FileOut   string `json:"file_out"`
	Overwrite bool   `json:"overwrite"`
}

func copyFile(reqId string, fileCopySettings FileCopyBody) (int, error) {
	logDebug.Printf("%s - Starting copy.", reqId)
	logDebug.Printf("%s - FileIn: %s", reqId, fileCopySettings.FileIn)
	logDebug.Printf("%s - FileOut: %s", reqId, fileCopySettings.FileOut)
	logDebug.Printf("%s - Overwrite: %v", reqId, fileCopySettings.Overwrite)

	destExists := fileExists(fileCopySettings.FileOut)
	if fileCopySettings.Overwrite || !destExists {
		orig, err := os.Open(fileCopySettings.FileIn)
		if err != nil {
			return 0, err
		}
		about, err := os.Stat(fileCopySettings.FileIn)
		modTime := about.ModTime()
		nowTime := time.Now().Local()
		if err != nil {
			return 0, err
		}
		defer orig.Close()

		new, err := os.Create(fileCopySettings.FileOut)
		if err != nil {
			return 0, err
		}
		defer new.Close()

		written, err := io.Copy(new, orig)
		if err != nil {
			return 0, err
		}
		os.Chtimes(fileCopySettings.FileOut, nowTime, modTime)
		logDebug.Printf("%s - File %s copied to %s successfully\n", reqId, fileCopySettings.FileIn, fileCopySettings.FileOut)
		return int(written), err
	} else {
		logDebug.Printf("%s - Destination file \"%s\" exists. No operations done\n", reqId, fileCopySettings.FileOut)
		return -1, nil
	}

}

func copyInterfaceSync(reqId string, fileCopySettings FileCopyBody) (Envelope, int) {
	var status int
	written, err := copyFile(reqId, fileCopySettings)
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileCopySettings,
	}
	env := Envelope{
		Data: data,
	}
	if err != nil {
		logError.Printf("While processing copy on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		if written > -1 {
			env.Message = "File copied successfully"
			status = http.StatusCreated
		} else {
			env.Message = "File not copied due to already existing and overwrite was set to false"
			status = http.StatusOK
		}

	}
	return env, status
}

func copyAsyncWrapper(reqId string, fileCopySettings FileCopyBody, sendStatusTo, sendStatusAuth, sendStatusToAuthType string) {
	var status int
	written, err := copyFile(reqId, fileCopySettings)
	env := Envelope{}
	if err != nil {
		logError.Printf("Error while processing copy on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		if written > -1 {
			env.Message = "File copied successfully"
			status = http.StatusCreated
		} else {
			env.Message = "File not copied due to already existing and overwrite was set to false"
			status = http.StatusOK
		}
	}
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileCopySettings,
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
		if sendStatusToAuthType != "" {
			req.Header.Set("Content-Type", sendStatusToAuthType)
		} else {
			req.Header.Set("Content-Type", "application/json")
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
}

func copyInterfaceASync(reqId string, fileCopySettings FileCopyBody, sendStatusToAuthType, sendStatusTo, sendStatusAuth string) (Envelope, int) {
	go copyAsyncWrapper(reqId, fileCopySettings, sendStatusTo, sendStatusAuth, sendStatusToAuthType)
	data := map[string]interface{}{
		"body": fileCopySettings,
	}
	env := Envelope{
		Message: "Copy process started",
		Data:    data,
	}
	return env, http.StatusAccepted
}

func copyFileHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)

	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Printf("%s - Starting copy process handling\n", reqId)
	logDebug.Printf("%s - Request body: %s\n", reqId, string(reqBody))
	var fileCopySettings FileCopyBody
	err := json.Unmarshal(reqBody, &fileCopySettings)
	if err != nil {
		logError.Printf("%s - Fhaas received a malformed request: %s", reqId, err)
		env, status = createBadRequestResponse(reqBody)
	} else {
		if isSyncRequest(reqId, r) {
			env, status = copyInterfaceSync(reqId, fileCopySettings)
		} else {
			sendStatusTo := r.Header.Get(H_SEND_STATUS_TO)
			sendStatusToAuth := r.Header.Get(H_SEND_STATUS_TO_AUTH)
			sendStatusToAuthType := r.Header.Get(H_SEND_STATUS_TO_AUTH_TYPE)
			logDebug.Printf("%s - Status will be sent to %s", reqId, showIfNotBlank(sendStatusTo))
			if sendStatusToAuth != "" {
				logDebug.Printf("%s - Status will send the following Authorization header: %s", reqId, showToken(sendStatusToAuth))
			}
			env, status = copyInterfaceASync(
				reqId,
				fileCopySettings,
				sendStatusTo,
				sendStatusToAuth,
				sendStatusToAuthType,
			)
		}
	}
	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Copy process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}

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
	env.Message = "Copies processed"
	return env, finalStatus
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
		}
		/* else {
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
		*/
	}
	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Copy process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}
