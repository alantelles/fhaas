package main

import (
	"fmt"
	"net/http"
)

func addDefaultHeaders(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Adding default headers")
		w.Header().Set("Content-Type", "application/json")
		endpoint(w, r)
	})
}

func verifyAuth(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("checking authorization")
		w.Header().Set("Content-Type", "application/json")
		endpoint(w, r)
	})
}
