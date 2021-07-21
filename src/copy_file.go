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

func copyFile(reqId string, fileCopySettings FileCopyBody) (int, bool, error) {
	logDebug.Printf("%s - Starting copy.\n", reqId)
	logDebug.Printf("%s - FileIn: %s\n", reqId, fileCopySettings.FileIn)
	logDebug.Printf("%s - FileOut: %s\n", reqId, fileCopySettings.FileOut)
	logDebug.Printf("%s - Overwrite: %v\n", reqId, fileCopySettings.Overwrite)
	logDebug.Printf("%s - CreateDir: %v\n", reqId, fileCopySettings.CreateDir)
	sizeMatch := false
	destExists := fileExists(fileCopySettings.FileOut)
	if fileCopySettings.Overwrite || !destExists {
		orig, err := os.Open(fileCopySettings.FileIn)
		if err != nil {
			return 0, false, err
		}
		about, err := os.Stat(fileCopySettings.FileIn)
		if err != nil {
			return 0, false, err
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
					return 0, false, err
				}
				err = os.Chown(destDir, int(osFileInfo.Uid), int(osFileInfo.Gid))
				if err != nil {
					return 0, false, err
				}
			}
		}

		new, err := os.Create(fileCopySettings.FileOut)
		if err != nil {
			return 0, false, err
		}
		defer new.Close()

		written, err := io.Copy(new, orig)
		if err != nil {			
			return int(written), sizeMatch, err
		}
		
		info_out, err := os.Stat(fileCopySettings.FileOut) 
		if err != nil{
			return int(written), false, err
		}
		sizeMatch = info_out.Size() == about.Size()
		
		os.Chtimes(fileCopySettings.FileOut, nowTime, modTime)
		err = os.Chown(fileCopySettings.FileOut, int(osFileInfo.Uid), int(osFileInfo.Gid))
		if err != nil {
			return int(written), sizeMatch, err
		}
		err = os.Chmod(fileCopySettings.FileOut, fs.FileMode(osFileInfo.Mode))
		logDebug.Printf("%s - File %s copied to %s successfully\n", reqId, fileCopySettings.FileIn, fileCopySettings.FileOut)
		return int(written), sizeMatch, err
	} else {
		logDebug.Printf("%s - Destination file \"%s\" exists. No operations done\n", reqId, fileCopySettings.FileOut)
		return -1, sizeMatch, nil
	}

}

func copyInterfaceSync(reqId string, fileCopySettings FileCopyBody) (Envelope, int) {
	nowThreads += 1
	var status int
	written, sizeMatch, err := copyFile(reqId, fileCopySettings)
	data := map[string]interface{}{
		"bytes_written": written,
		"body":          fileCopySettings,
		"size_match": sizeMatch,
	}
	env := Envelope{
		Data:      data,
		RequestId: strings.Replace(reqId, "Request ", "", -1),
	}
	if err != nil {
		logError.Printf("While processing copy on %s: %v\n", reqId, err)
		
		if sizeMatch {
			env.Message = fmt.Sprintf("Operation partially successful: %v", err)
			logWarn.Printf("%s - The request was partially successful. The file sizes match but permission and/or ownership couldn't be changed", reqId)
			env.Status = http.StatusCreated
			status = http.StatusCreated
		} else {
			env.Message = fmt.Sprintf("Operation failed: %v", err)
			status = http.StatusInternalServerError
		}
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
	written, sizeMatch, err := copyFile(reqId, fileCopySettings)
	env := Envelope{RequestId: strings.Replace(reqId, "Request ", "", -1)}
	if err != nil {
		logError.Printf("While processing copy on %s: %v\n", reqId, err)
		
		if sizeMatch {
			env.Message = fmt.Sprintf("Operation partially successful: %v", err)
			logWarn.Printf("%s - The request was partially successful. The file sizes match but permission and/or ownership couldn't be changed", reqId)
			env.Status = http.StatusCreated
			status = http.StatusCreated
		} else {
			env.Message = fmt.Sprintf("Operation failed: %v", err)
			status = http.StatusInternalServerError
		}
	} else {
		if written > -1 {
			env.Message = "File copied successfully"
			status = http.StatusCreated
		} else {
			env.Message = "File not copied due to already existing and overwrite was set to false"
			status = http.StatusOK
		}

	}
	// if err != nil {
	// 	logError.Printf("Error while processing copy on %s: %v\n", reqId, err)
	// 	env.Message = fmt.Sprintf("Operation failed: %v", err)
	// 	status = http.StatusInternalServerError
	// } else {
	// 	if written > -1 {
	// 		env.Message = "File copied successfully"
	// 		status = http.StatusCreated
	// 	} else {
	// 		env.Message = "File not copied due to already existing and overwrite was set to false"
	// 		status = http.StatusOK
	// 	}
	// }
	data := map[string]interface{}{
		"bytesWritten": written,
		"body":         fileCopySettings,
		"status":       status,
		"size_match":   sizeMatch,
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
