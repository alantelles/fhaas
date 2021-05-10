package main

import (
	"fmt"
	"os"
)

func selectAuthUrl(authByHeader string) string {
	authByEnv := os.Getenv("FHAAS_AUTH_URL")
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
