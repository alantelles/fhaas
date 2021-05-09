package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequests() {
	r := mux.NewRouter()
	r.HandleFunc("/", addDefaultHeaders(verifyAuth(indexHandler)))
	r.HandleFunc("/copy", addDefaultHeaders(verifyAuth(receiveCopyWork))).Methods("POST")
	fmt.Println("Serving at 8080...")
	http.ListenAndServe(":8080", r)
}
