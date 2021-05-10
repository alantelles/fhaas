package main

import (
	"encoding/json"
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

func copyInterfaceSync(fileCopySettings FileCopyBody) (Envelope, int) {
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
		env.Message = "Operation failed"
		status = http.StatusInternalServerError
	} else {
		env.Message = "File copied successfully"
		status = http.StatusCreated
	}
	return env, status
}

func copyFileHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	logDebug.Printf(logRequest(w, r))
	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	logDebug.Println()
	var fileCopySettings FileCopyBody
	json.Unmarshal(reqBody, &fileCopySettings)
	if isSyncRequest(r) {
		env, status = copyInterfaceSync(fileCopySettings)
	}
	respond(env, w, status)
}
