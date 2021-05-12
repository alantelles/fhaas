package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func addDefaultHeaders(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := uuid.NewString()
		logDebug.Printf(logRequest(reqId, r))
		fmt.Println("Adding default headers")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(H_REQUEST_ID, reqId)
		endpoint(w, r)
	})
}

func respondNotAuthorized(w http.ResponseWriter, r *http.Request) {
	env := Envelope{
		Message: "Not-Authorized",
		Data:    make(map[string]interface{}),
	}

	respond(env, w, 401)
}

// func checkIfIsAsync(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
// 	return http.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !isSyncRequest(r) {

// 		}
// 	})
// }

func verifyAuth(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := getRequestId(w)
		fmt.Println("Checking authorization")
		authToken := r.Header.Get(H_AUTH_TOKEN)
		authUrlHeader := r.Header.Get(H_AUTH_URL)
		authUrl := selectAuthUrl(authUrlHeader)
		w.Header().Set(H_AUTH_URL_USED, authUrl)
		fmt.Printf("AuthUrl used: %s\n", authUrl)
		logDebug.Printf("%s - Auth URL used for this request: %s\n", reqId, authUrl)
		if authToken != "" && authUrl != "" {
			fmt.Println("Trying to auth")
			resp, code, err := doPost(authUrl, fmt.Sprintf("%s", authToken), getAuthContentType(w))
			if err != nil {
				fmt.Println(err)
				logError.Printf("%s - Authentication process failed: %v", reqId, err)
			}
			if code != 200 {
				fmt.Printf("Not authorized, status code: %d\n", code)
				logWarn.Printf("%s - FhaaS received a request that was not authorized. It may be a token/url mistake but also may have came from an undesired/malicious requester. We suggest you to check your FhaaS network entrypoints\n", reqId)
				logError.Printf("%s - This request was not authorized by endpoint %s\n", reqId, authUrl)
				logDebug.Printf("%s - Authorization response: %s\n", reqId, resp)
				logDebug.Printf("%s - Authorization response code: %d\n", reqId, code)
				logDebug.Printf("%s - Authorization token used: %s\n", reqId, showToken(authToken))
				respondNotAuthorized(w, r)
			} else {
				fmt.Println("Authorized")
				logDebug.Printf("%s - Authorized by %s with token %s. Processing request.", reqId, authUrl, showToken(authToken))
				endpoint(w, r)
			}
		} else {
			fmt.Println("One or more of needed auth headers are missed.")
		}

	})
}
