package main

import (
	"fmt"
	"net/http"
	"os"
)

func retrieveFileInfo(reqId string, fileRetrieveSettings FileRetrieveQuery, index int) (FileInfoConform, error) {
	logDebug.Printf("%s - Retrieving file info.\n", reqId)
	logDebug.Printf("%s - FileName: %s\n", reqId, fileRetrieveSettings.Files[index])
	info, err := os.Stat(fileRetrieveSettings.Files[index])
	content := fillFileInfoConform(info, info.Name())
	if err != nil {
		return content, err
	}
	logDebug.Printf("%s - File %s info retrieved successfully\n", reqId, fileRetrieveSettings.Files[index])
	return content, err
}

func retrieveFileInfoInterfaceSync(reqId string, fileRetrieveSettings FileRetrieveQuery) (Envelope, int) {
	nowThreads += 1
	var status int
	data := map[string]interface{}{
		"query": fileRetrieveSettings,
	}
	content, err := retrieveFileInfo(reqId, fileRetrieveSettings, 0)
	env := Envelope{
		Data:      data,
		RequestId: dropReq(reqId),
	}
	if err != nil {
		logError.Printf("While processing file info retrieve on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		data["info"] = content
		env.Message = "File info retrieved successfully"
		status = http.StatusOK
	}
	env.Status = status
	nowThreads -= 1
	return env, status
}

func retrieveFileInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int
	)
	reqId := getRequestId(w)
	logDebug.Printf("%s - Starting retrieve file info handling\n", reqId)
	//logDebug.Printf("%s - Request query: %s\n", reqId, string(reqBody))
	filename, ok := r.URL.Query()["filename"]
	if !ok || len(filename) == 0 {
		env.Message = "FhaaS has received a malformed request"
		env.Data = map[string]interface{}{
			"query": r.URL.Query(),
		}
		env.Status = http.StatusBadRequest
		status = http.StatusBadRequest
		env.RequestId = dropReq(reqId)
	} else {

		fileRetrieveSettings := FileRetrieveQuery{
			Files: filename,
		}
		env, status = retrieveFileInfoInterfaceSync(reqId, fileRetrieveSettings)
		env.Status = status
	}
	respond(env, w, status)
}
