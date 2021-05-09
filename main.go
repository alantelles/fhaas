package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type Envelope struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func respond(data Envelope, w http.ResponseWriter, status int) {
	dataStr, _ := json.Marshal(data)
	w.WriteHeader(status)
	fmt.Fprintln(w, string(dataStr))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tt := map[string]interface{}{
		"docs": "future-link",
	}
	data := Envelope{
		Message: "FhaaS - File handling as a service",
		Data:    tt,
	}

	respond(data, w, 200)
}

/*func otherHandler(w http.ResponseWriter, r *http.Request) {
	sub := map[string]interface{}{
		"filename": "tops",
		"length":   950,
	}
	tt := map[string]interface{}{
		"jonga":   905,
		"teske":   true,
		"content": "A Marina canta música doida",
		"data":    sub,
	}
	data := &AnyData{
		Message: "O João acordou",
		Data:    "Estes são os dados que quero passar",
		Status:  200,
		Torbe:   tt,
	}
	dataStr, _ := json.Marshal(data)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(dataStr))
}*/

var fhaasAuthEndpoint string

func selectAuthUrl(authByHeader string) string {
	authByEnv := os.Getenv("FHAAS_AUTH_URL")
	if fhaasAuthEndpoint != "" {
		fmt.Println("Authenticating by flag set authurl")
		return fhaasAuthEndpoint
	} else if authByEnv != "" {
		fmt.Println("Authenticating by environment variable set")
		return authByEnv
	} else {
		fmt.Println("Authenticating by header Authorization")
		return authByHeader
	}
}

func main() {
	// execution arguments setting
	authPtr := flag.String("authurl", "", "Default authentication url")
	flag.Parse()

	fhaasAuthEndpoint = *authPtr
	handleRequests()
}
