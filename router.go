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
	r.HandleFunc("/copy/many", addDefaultHeaders(verifyAuth(copyFileListHandler))).Methods("POST")
	r.HandleFunc("/copy", addDefaultHeaders(verifyAuth(copyFileHandler))).Methods("POST")
	r.HandleFunc("/move", addDefaultHeaders(verifyAuth(moveFileHandler))).Methods("PUT")

	fmt.Println("Serving at 8080...")
	http.ListenAndServe(":8080", r)
}
