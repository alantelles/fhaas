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

func isSyncRequest(reqId string, r *http.Request) bool {
	isAsync := r.Header.Get(H_IS_ASYNC)
	if isAsync == "false" || isAsync == "" {
		logDebug.Printf("%s - Request is sync\n", reqId)
		return true
	}
	logDebug.Printf("%s - Request is async\n", reqId)
	return false
}

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func showToken(token string) string {
	if allowLogTokens {
		return token
	} else {
		return "[logtokens disabled]"
	}
}

func showIfNotBlank(value string) string {
	if value == "" {
		return "[value not set]"
	} else {
		return value
	}
}
