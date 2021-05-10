package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func addDefaultHeaders(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Adding default headers")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(H_REQUEST_ID, uuid.NewString())

		endpoint(w, r)
	})
}

func respondNotAuthorized(w http.ResponseWriter, r *http.Request) {
	env := Envelope{
		Message: "Not-Authorized",
		Data:    make(map[string]interface{}),
	}
	logWarn.Printf("%s - FhaaS received a request that was not authorized. It may be a token/url mistake but also may have came from an undesired/malicious requester. We suggest you to check your FhaaS network entrypoints\n", getRequestId(w))
	logError.Printf("%s - This request was not authorized by endpoint %s\n", getRequestId(w), getAuthUrlUsed(w))
	respond(env, w, 401)
}

func verifyAuth(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Checking authorization")
		authToken := r.Header.Get(H_AUTH_TOKEN)
		authUrlHeader := r.Header.Get(H_AUTH_URL)
		authUrl := selectAuthUrl(authUrlHeader)
		w.Header().Set(H_AUTH_URL_USED, authUrl)
		fmt.Printf("AuthUrl used: %s\n", authUrl)
		logDebug.Printf("%s - Auth URL used for this request: \n", getRequestId(w))
		if authToken != "" && authUrl != "" {
			fmt.Println("Trying to auth")
			_, code, err := doPost(authUrl, fmt.Sprintf("%s", authToken), getAuthContentType(w))
			if err != nil {
				fmt.Println(err)
				logError.Printf("%s - Authentication process failed: %v", getRequestId(w), err)
			}
			if code != 200 {
				fmt.Printf("Not authorized, status code: %d\n", code)
				respondNotAuthorized(w, r)
			} else {
				fmt.Println("Authorized")
				logError.Printf("%s - Authorized. Processing request.", getRequestId(w))
				endpoint(w, r)
			}
		} else {
			fmt.Println("One or more of needed auth headers are missed.")
		}

	})
}
