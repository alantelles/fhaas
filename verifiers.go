package main

import (
	"fmt"
	"net/http"
	"os"
)

func selectAuthUrl(authByHeader string) string {
	authByEnv := os.Getenv(E_FHAAS_AUTH_URL)
	if fhaasAuthEndpoint != "" {
		fmt.Println("Authenticating by flag set authurl")
		return fhaasAuthEndpoint
	} else if authByEnv != "" {
		fmt.Println("Authenticating by environment variable set")
		return authByEnv
	} else {
		fmt.Println("Authenticating by header Auth")
		return authByHeader
	}
}

func isSyncRequest(reqId int, r *http.Request) bool {
	isAsync := r.Header.Get(H_IS_ASYNC)
	if isAsync == "false" || isAsync == "" {
		logDebug.Printf("%s - Request is sync\n", reqId)
		return true
	}
	logDebug.Printf("%s - Request is async\n", reqId)
	return false
}
