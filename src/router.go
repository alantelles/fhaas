package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequests() {
	r := mux.NewRouter()
	r.HandleFunc("/", addDefaultHeaders(verifyAuth(indexHandler)))
	r.HandleFunc("/self_auth", addDefaultHeaders(selfAuth))
	//copy routes
	r.HandleFunc("/copy", addDefaultHeaders(verifyAuth(copyFileHandler))).Methods("POST")
	r.HandleFunc("/copy/many", addDefaultHeaders(verifyAuth(copyFileListHandler))).Methods("POST")

	//move routes
	r.HandleFunc("/move", addDefaultHeaders(verifyAuth(moveFileHandler))).Methods("PUT")
	r.HandleFunc("/move/many", addDefaultHeaders(verifyAuth(moveFileListHandler))).Methods("PUT")

	//retrieve routes
	r.HandleFunc("/retrieve", checkThreadLimit(addDefaultHeaders(verifyAuth(retrieveFileContentHandler)))).Methods("GET")

	//list files
	r.HandleFunc("/list", checkThreadLimit(addDefaultHeaders(verifyAuth(listFolderContentContentHandler)))).Methods("GET")

	//get file info
	r.HandleFunc("/file_info", checkThreadLimit(addDefaultHeaders(verifyAuth(retrieveFileInfoHandler)))).Methods("GET")

	//aux
	r.HandleFunc("/threads", addDefaultHeaders(verifyAuth(getThreadsHandler))).Methods("GET")

	fmt.Println("Serving at 8080...")
	http.ListenAndServe(":8080", r)
}
