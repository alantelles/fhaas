package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ListFolderQuery struct {
	Path          string   `json:"path"`
	Recursive     bool     `json:"recursive"`
	IncludeHidden bool     `json:"show_hidden"`
	IncludeDirs   bool     `json:"show_dirs"`
	Filters       []string `json:"filter"`
}

type FileInfoConform struct {
	FileName    string    `json:"filename"`
	Size        int       `json:"length"`
	LastMod     time.Time `json:"last_mod"`
	IsDirectory bool      `json:"is_directory"`
}

func listFolderContent(reqId string, listFolderSettings ListFolderQuery, index int) ([]FileInfoConform, error) {

	nowThreads += 1
	logDebug.Printf("%s - Listing folder content.\n", reqId)
	logDebug.Printf("%s - Path: %s\n", reqId, listFolderSettings.Path)
	files, err := ioutil.ReadDir(listFolderSettings.Path)
	results := make([]FileInfoConform, 0)
	if err != nil {
		nowThreads -= 1
		return results, err
	}

	for _, file := range files {
		attendFilters := fileNameAttendFilters(file.Name(), listFolderSettings.Filters)
		if attendFilters {
			fileInfo := FileInfoConform{}
			fileInfo.FileName = file.Name()
			fileInfo.Size = int(file.Size())
			fileInfo.LastMod = file.ModTime()
			fileInfo.IsDirectory = file.IsDir()
			results = append(results, fileInfo)
		}
	}
	nowThreads -= 1
	return results, err
}

func listFolderContentInterfaceSync(reqId string, listFolderSettings ListFolderQuery) (Envelope, int) {
	var status int
	data := map[string]interface{}{
		"query": listFolderSettings,
	}
	content, err := listFolderContent(reqId, listFolderSettings, 0)
	data["files"] = content
	env := Envelope{
		Data:      data,
		RequestId: dropReq(reqId),
	}
	if err != nil {
		logError.Printf("While processing folder content list on %s: %v\n", reqId, err)
		env.Message = fmt.Sprintf("Operation failed: %v", err)
		status = http.StatusInternalServerError
	} else {
		env.Message = "Folder content retrieved successfully"
		status = http.StatusOK
	}
	env.Status = status
	return env, status
}

func checkListFolderQuery(r *http.Request) (ListFolderQuery, error) {
	var (
		contains bool
		value    string
		values   []string
	)
	listQuery := ListFolderQuery{
		Recursive: false,
	}
	query := r.URL.Query()
	contains, value, _ = checkQueryParam(query, "path")
	if !contains {
		return listQuery, errors.New("Path is a required query param but it's not present")
	}
	listQuery.Path = value
	contains, value, _ = checkQueryParam(query, "recursive")
	listQuery.Recursive = contains
	contains, _, _ = checkQueryParam(query, "show_hidden")
	listQuery.IncludeHidden = contains
	contains, _, _ = checkQueryParam(query, "show_dirs")
	listQuery.IncludeDirs = contains
	_, _, values = checkQueryParam(query, "filter")
	listQuery.Filters = values
	return listQuery, nil
}

func listFolderContentContentHandler(w http.ResponseWriter, r *http.Request) {
	var (
		env    Envelope
		status int = http.StatusOK
	)
	reqId := getRequestId(w)
	logDebug.Printf("%s - Starting list folder content file handling\n", reqId)
	listFolderSettings, err := checkListFolderQuery(r)
	if err != nil {
		logError.Printf("%s - FhaaS has received a malformed request: %v\n", reqId, err)
		env.Message = "FhaaS has received a malformed request"
		env.Data = map[string]interface{}{
			"query": r.URL.Query(),
		}
		env.Status = http.StatusBadRequest
		env.RequestId = dropReq(reqId)
	} else {
		env, status = listFolderContentInterfaceSync(reqId, listFolderSettings)
	}
	respond(env, w, status)
}
