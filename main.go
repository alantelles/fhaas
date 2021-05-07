package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func main() {
	handleRequests()
}
