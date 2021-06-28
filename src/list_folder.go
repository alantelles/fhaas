package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ListFolderQuery struct {
	Path          string   `json:"path"`
	Recursive     bool     `json:"recursive"`
	IncludeHidden bool     `json:"hidden"`
	IncludeDirs   bool     `json:"dirs"`
	Filters       []string `json:"filter"`
}

type FileInfoConform struct {
	FileName    string      `json:"filename"`
	Size        int         `json:"length"`
	LastMod     time.Time   `json:"last_mod"`
	IsDirectory bool        `json:"is_directory"`
	Sys         interface{} `json:"os_fileinfo"`
}

func fillFileInfoConform(file os.FileInfo, path string) FileInfoConform {
	fileInfo := FileInfoConform{}
	fileInfo.FileName = path
	fileInfo.Size = int(file.Size())
	fileInfo.LastMod = file.ModTime()
	fileInfo.IsDirectory = file.IsDir()
	fileInfo.Sys = file.Sys()
	return fileInfo
}

func listFolderContent(reqId string, listFolderSettings ListFolderQuery, index int) ([]FileInfoConform, error) {

	logDebug.Printf("%s - Listing folder content.\n", reqId)
	logDebug.Printf("%s - Path: %s\n", reqId, listFolderSettings.Path)
	results := make([]FileInfoConform, 0)
	noFilters := len(listFolderSettings.Filters) == 0
	if !listFolderSettings.Recursive {
		files, err := ioutil.ReadDir(listFolderSettings.Path)
		if err != nil {
			return results, err
		}
		for _, file := range files {
			attendFilters := fileNameAttendFilters(file.Name(), listFolderSettings.Filters)
			isHidden := string(file.Name()[0]) == "."
			shouldIncludeIfHidden := (isHidden && listFolderSettings.IncludeHidden) || !isHidden
			isDir := file.IsDir()
			shouldIncludeIfDir := (listFolderSettings.IncludeDirs && isDir) || !isDir
			if (noFilters || attendFilters) && shouldIncludeIfHidden && shouldIncludeIfDir {
				fileInfo := fillFileInfoConform(file, filepath.Join(listFolderSettings.Path, file.Name()))
				results = append(results, fileInfo)
			}
		}
	} else {
		err := filepath.Walk(listFolderSettings.Path, func(path string, file os.FileInfo, err error) error {
			attendFilters := fileNameAttendFilters(file.Name(), listFolderSettings.Filters)
			isHidden := string(file.Name()[0]) == "."
			shouldIncludeIfHidden := (isHidden && listFolderSettings.IncludeHidden) || !isHidden
			isDir := file.IsDir()
			shouldIncludeIfDir := (listFolderSettings.IncludeDirs && isDir) || !isDir
			if (noFilters || attendFilters) && shouldIncludeIfHidden && shouldIncludeIfDir {
				fileInfo := fillFileInfoConform(file, path)
				results = append(results, fileInfo)
			}
			return nil
		})
		if err != nil {
			return results, err
		}

	}
	return results, nil
}

func listFolderContentInterfaceSync(reqId string, listFolderSettings ListFolderQuery) (Envelope, int) {
	nowThreads += 1
	logDebug.Printf("%s - Listing folder content at path %s\n", reqId, listFolderSettings.Path)
	marshalListFolder, _ := json.Marshal(listFolderSettings)
	logDebug.Printf("%s - Operations settings: %v\n", reqId, string(marshalListFolder))
	var status int
	data := map[string]interface{}{
		"query": listFolderSettings,
	}
	content, err := listFolderContent(reqId, listFolderSettings, 0)
	data["files"] = content
	data["count"] = len(content)
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
	logDebug.Printf("%s - %s", reqId, env.Message)
	env.Status = status
	nowThreads -= 1
	return env, status
}

func checkListFolderQuery(r *http.Request) (ListFolderQuery, error) {
	var (
		contains bool
		value    string
		values   []string
	)
	listQuery := ListFolderQuery{}
	query := r.URL.Query()
	contains, value, _ = checkQueryParam(query, "path")
	if !contains {
		return listQuery, errors.New("Path is a required query param but it's not present")
	}
	listQuery.Path = value
	contains, value, _ = checkQueryParam(query, "recursive")
	listQuery.Recursive = contains
	contains, _, _ = checkQueryParam(query, "hidden")
	listQuery.IncludeHidden = contains
	contains, _, _ = checkQueryParam(query, "dirs")
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
