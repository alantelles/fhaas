package main

import "flag"

func setFlags() {

	authPtr := getAuthUrlFlag()
	logTokensPtr := getAllowLogToken()
	maxThreadsPtr := getMaxThreads()
	flag.Parse()

	setAuthUrlFlag(authPtr)
	setAllowLogToken(logTokensPtr)
	setMaxThreads(maxThreadsPtr)
}

func getMaxThreads() *int {
	return flag.Int("maxthreads", 10, "Maximum number of simultaneous processes allowed. 0 for no limit [use with caution]")
}

func setMaxThreads(value *int) {
	maxThreads = *value
	logDebug.Printf("maxthreads flag is set to %v\n", maxThreads)
	if maxThreads != 0 {
		logDebug.Printf("Maximum simultaneous processes allowed is %v. Requests received when this limit is reached will return a 503 (Service unavailable) as response status.\n", maxThreads)
	} else {
		logDebug.Printf("FhaaS is operating with no limit of simultaneous threads. This is highly not recommended. Use at your own risk.\n")
	}
}

func getAuthUrlFlag() *string {
	return flag.String("authurl", "", "Default authentication url")
}

func setAuthUrlFlag(value *string) {
	fhaasAuthEndpoint = *value
	if fhaasAuthEndpoint == "" {
		logWarn.Println("Flag authurl not set. Application will use FHAAS_AUTH_URL environment variable (will be checked in every request). If not set, will use " + H_AUTH_URL + " header of request. This may be potentially dangerous since any url able to authorize operation can be used")
	} else {
		logDebug.Printf("Using %s as url for authorization.\n", fhaasAuthEndpoint)
	}
}

func getAllowLogToken() *string {
	return flag.String("logtokens", "false", "Allow authentication tokens to be logged.")
}

func setAllowLogToken(value *string) {
	logTokens := *value
	if logTokens == "true" {
		allowLogTokens = true
		logDebug.Println("logtokens is set to true.")
		logWarn.Println("logtokens is set to true. Though it is useful to verify authorization fails it exposes \"credential-like\" informations on logs.")
	} else if logTokens == "false" {
		allowLogTokens = false
		logDebug.Println("logtokens is set to false.")
	} else {
		allowLogTokens = false
		logDebug.Println("Invalid value for logtokens. Will be set to false.")
	}
}
