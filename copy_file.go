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

func copyFileHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	var fileCopySettings FileCopyBody
	json.Unmarshal(reqBody, &fileCopySettings)
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
		respond(env, w, http.StatusInternalServerError)
	} else {
		env.Message = "File copied successfully"
		respond(env, w, http.StatusCreated)
	}
}
