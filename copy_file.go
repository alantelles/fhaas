package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type FileCopyBody struct {
	FileIn  string `json:"file_in"`
	FileOut string `json:"file_out"`
}

func receiveCopyWork(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	var fileCopySettings FileCopyBody
	json.Unmarshal(reqBody, &fileCopySettings)
	orig, err := os.Open(fileCopySettings.FileIn)
	about, _ := os.Stat(fileCopySettings.FileIn)
	modTime := about.ModTime()
	nowTime := time.Now().Local()
	if err != nil {
		log.Fatal(err)
	}
	defer orig.Close()

	new, err := os.Create(fileCopySettings.FileOut)
	if err != nil {
		log.Fatal(err)
	}
	defer new.Close()

	written, err := io.Copy(new, orig)
	if err != nil {
		log.Fatal(err)
	}
	os.Chtimes(fileCopySettings.FileOut, nowTime, modTime)
	data := map[string]interface{}{
		"bytesWritten": written,
	}
	env := Envelope{
		Message: "Copied successfully",
		Data:    data,
	}
	respond(env, w)
}
