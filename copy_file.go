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

func copyFile(fileCopySettings FileCopyBody) (int, error) {
	orig, err := os.Open(fileCopySettings.FileIn)
	about, _ := os.Stat(fileCopySettings.FileIn)
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
	return int(written), err
}

func copyInterfaceSync(reqId string, fileCopySettings FileCopyBody) (Envelope, int) {
	var status int
	written, err := copyFile(fileCopySettings)
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileCopySettings,
	}
	env := Envelope{
		Data: data,
	}
	if err != nil {
		logError.Printf("While processing copy on %s: %v\n", reqId, err)
		env.Message = "Operation failed"
		status = http.StatusInternalServerError
	} else {
		logDebug.Printf("%s - File %s copied to %s successfully\n", reqId, fileCopySettings.FileIn, fileCopySettings.FileOut)
		env.Message = "File copied successfully"
		status = http.StatusCreated
	}
	return env, status
}

func copyAsyncWrapper(reqId string, fileCopySettings FileCopyBody, sendStatusTo string, sendStatusAuth string) {
	var status int
	written, err := copyFile(fileCopySettings)
	env := Envelope{}
	if err != nil {
		logError.Printf("While processing copy on %s: %v\n", reqId, err)
		env.Message = "Operation failed"
		status = http.StatusInternalServerError
	} else {
		logDebug.Printf("%s - File %s copied to %s successfully\n", reqId, fileCopySettings.FileIn, fileCopySettings.FileOut)
		env.Message = "File copied successfully"
		status = http.StatusCreated
	}
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileCopySettings,
		"status":       status,
	}
	env.Data = data
	if sendStatusTo != "" {
		body, _ := json.Marshal(env)
		req, err := http.NewRequest("POST", sendStatusTo, bytes.NewReader(body))
		if err != nil {
			fmt.Println(err)
		}
		if sendStatusAuth != "" {
			req.Header.Set("Authorization", sendStatusAuth)
		}
		req.Header.Set("Content-Type", "application/json")
		client := createClient(20)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(resp.Body)
		respStr := string(respBytes)
		logDebug.Printf("%s - Status endpoint returned with: %s", reqId, respStr)
	}
}

func copyInterfaceASync(reqId string, fileCopySettings FileCopyBody, sendStatusTo string, sendStatusAuth string) (Envelope, int) {
	go copyAsyncWrapper(reqId, fileCopySettings, sendStatusTo, sendStatusAuth)
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
	logDebug.Printf(logRequest(w, r))
	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Println("%s - Starting copy process handling\n", reqId)
	logDebug.Println("%s - Request body: ", reqId, string(reqBody))
	var fileCopySettings FileCopyBody
	json.Unmarshal(reqBody, &fileCopySettings)
	if isSyncRequest(reqId, r) {
		env, status = copyInterfaceSync(reqId, fileCopySettings)
	} else {
		env, status = copyInterfaceASync(
			reqId,
			fileCopySettings,
			w.Header().Get(H_SEND_STATUS_TO),
			w.Header().Get(H_SEND_STATUS_TO_AUTH),
		)
	}
	logDebug.Printf("%s - Copy process response: %v", reqId, env)
	respond(env, w, status)
}
