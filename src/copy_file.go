package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type FileCopyBody struct {
	FileIn    string `json:"file_in"`
	FileOut   string `json:"file_out"`
	Overwrite bool   `json:"overwrite"`
	CreateDir bool   `json:"create_dir"`
}

func copyFile(reqId string, fileCopySettings FileCopyBody) (int, error) {
	logDebug.Printf("%s - Starting copy.\n", reqId)
	logDebug.Printf("%s - FileIn: %s\n", reqId, fileCopySettings.FileIn)
	logDebug.Printf("%s - FileOut: %s\n", reqId, fileCopySettings.FileOut)
	logDebug.Printf("%s - Overwrite: %v\n", reqId, fileCopySettings.Overwrite)
	logDebug.Printf("%s - CreateDir: %v\n", reqId, fileCopySettings.CreateDir)

	destExists := fileExists(fileCopySettings.FileOut)
	if fileCopySettings.Overwrite || !destExists {
		orig, err := os.Open(fileCopySettings.FileIn)
		if err != nil {
			return 0, err
		}
		about, err := os.Stat(fileCopySettings.FileIn)
		if err != nil {
			return 0, err
		}
		modTime := about.ModTime()
		nowTime := time.Now().Local()
		osFileInfo := about.Sys().(*syscall.Stat_t)
		defer orig.Close()
		if fileCopySettings.CreateDir {
			destDir := filepath.Dir(fileCopySettings.FileOut)
			if destDir != "." {
				err := os.MkdirAll(destDir, 0777)
				if err != nil {
					return 0, err
				}
				err = os.Chown(destDir, int(osFileInfo.Uid), int(osFileInfo.Gid))
				if err != nil {
					return 0, err
				}
			}
		}

		new, err := os.Create(fileCopySettings.FileOut)
		if err != nil {
			return 0, err
		}
		defer new.Close()

		written, err := io.Copy(new, orig)
		if err != nil {
			return 0, err
		}
		err = os.Chown(fileCopySettings.FileOut, int(osFileInfo.Uid), int(osFileInfo.Gid))
		if err != nil {
			return 0, err
		}
		err = os.Chmod(fileCopySettings.FileOut, fs.FileMode(osFileInfo.Mode))
		os.Chtimes(fileCopySettings.FileOut, nowTime, modTime)
		logDebug.Printf("%s - File %s copied to %s successfully\n", reqId, fileCopySettings.FileIn, fileCopySettings.FileOut)
		return int(written), err
	} else {
		logDebug.Printf("%s - Destination file \"%s\" exists. No operations done\n", reqId, fileCopySettings.FileOut)
		return -1, nil
	}

}

func copyInterfaceSync(reqId string, fileCopySettings FileCopyBody) (Envelope, int) {
	nowThreads += 1
	var status int
	written, err := copyFile(reqId, fileCopySettings)
	data := map[string]interface{}{
		"bytes_written": written,
		"body":          fileCopySettings,
	}
	env := Envelope{
		Data:      data,
		RequestId: strings.Replace(reqId, "Request ", "", -1),
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
	nowThreads -= 1
	env.Status = status
	return env, status
}

func copyAsyncWrapper(reqId string, fileCopySettings FileCopyBody, sendStatusTo, sendStatusAuth string) {
	nowThreads += 1
	var status int
	written, err := copyFile(reqId, fileCopySettings)
	env := Envelope{RequestId: strings.Replace(reqId, "Request ", "", -1)}
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
	env.Status = status
	shouldSendStatus(sendStatusTo, reqId, env, sendStatusAuth)
	nowThreads -= 1
}

func copyInterfaceASync(reqId string, fileCopySettings FileCopyBody, sendStatusTo, sendStatusAuth string) (Envelope, int) {
	go copyAsyncWrapper(reqId, fileCopySettings, sendStatusTo, sendStatusAuth)
	data := map[string]interface{}{
		"body": fileCopySettings,
	}
	env := Envelope{
		Message:   "Copy process started",
		Data:      data,
		RequestId: dropReq(reqId),
		Status:    http.StatusAccepted,
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
			logDebug.Printf("%s - Status will be sent to %s", reqId, showIfNotBlank(sendStatusTo))
			if sendStatusToAuth != "" {
				logDebug.Printf("%s - Status will send the following Authorization header: %s", reqId, showToken(sendStatusToAuth))
			}
			env, status = copyInterfaceASync(
				reqId,
				fileCopySettings,
				sendStatusTo,
				sendStatusToAuth,
			)
		}
	}
	envBytes, _ := json.Marshal(env)

	logDebug.Printf("%s - Copy process response: %s", reqId, string(envBytes))
	respond(env, w, status)
}
