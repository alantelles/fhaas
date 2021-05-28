package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type FileRetrieveQuery struct {
	Files  []string `json:"file_in"`
	Format string   `json:"format"`
}

func retrieveFile(reqId string, fileRetrieveSettings FileRetrieveQuery, index int) ([]byte, error) {
	content, err := ioutil.ReadFile(fileRetrieveSettings.Files[index])
	if err != nil {
		return content, err
	}
	return content, err
}

func retrieveFileInterfaceSync(reqId string, fileRetrieveSettings FileRetrieveQuery) (Envelope, int) {
	var status int
	data := map[string]interface{}{
		"query": fileRetrieveSettings,
	}
	content, err := retrieveFile(reqId, fileRetrieveSettings, 0)
	logDebug.Printf("%s - Format requested: %s", reqId, fileRetrieveSettings.Format)
	if fileRetrieveSettings.Format == "utf8" {
		data["content"] = string(content)
		/*} else if fileRetrieveSettings.Format == "ascii" {
		contStr := strconv.QuoteToASCII(string(content))

		data["content"] = contStr*/
	} else {
		logDebug.Printf("%s - Format not available. Serving as base64", reqId)
		data["content"] = content
	}
	env := Envelope{
		Data:      data,
		RequestId: dropReq(reqId),
	}
	if err != nil {
		logError.Printf("While processing retrieve on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		env.Message = "File retrieved successfully"
		status = http.StatusOK
	}
	env.Status = status
	return env, status
}

func retrieveFileContentHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)
	logDebug.Printf("%s - Starting retrieve content file handling\n", reqId)
	//logDebug.Printf("%s - Request query: %s\n", reqId, string(reqBody))
	filename, ok := r.URL.Query()["filename"]
	format, okf := r.URL.Query()["format"]
	if !ok || len(filename) == 0 {
		env.Message = "FhaaS has received a malformed request"
		env.Data = map[string]interface{}{
			"query": r.URL.Query(),
		}
		env.Status = http.StatusBadRequest
		env.RequestId = dropReq(reqId)
	} else {
		formatRetr := "base64"
		if okf && len(format) != 0 {
			formatRetr = format[0]
		}
		fileRetrieveSettings := FileRetrieveQuery{
			Files:  filename,
			Format: formatRetr,
		}
		env, status = retrieveFileInterfaceSync(reqId, fileRetrieveSettings)
	}
	respond(env, w, status)
}
