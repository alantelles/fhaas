package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type AnyData struct {
	Message string                 `json:"message"`
	Data    string                 `json:"data"`
	Torbe   map[string]interface{} `json:"torbe"`
	Status  int                    `json:"status"`
}

type FileCopyBody struct {
	FileIn  string `json:"file_in"`
	FileOut string `json:"file_out"`
}

func addDefaultHeaders(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Adding default headers")
		w.Header().Set("Content-Type", "application/json")
		endpoint(w, r)
	})
}

func receiveWork(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody, _ := ioutil.ReadAll(r.Body)
	var fileCopySettings FileCopyBody
	json.Unmarshal(reqBody, &fileCopySettings)
	fmt.Printf(string(reqBody))
	fmt.Fprintf(w, "{\"message\": \"OK\", \"body\": %s}", string(reqBody))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tt := map[string]interface{}{
		"pika": 905,
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
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
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
}

func handleRequests() {
	r := mux.NewRouter()
	r.HandleFunc("/", addDefaultHeaders(indexHandler))
	r.HandleFunc("/other", addDefaultHeaders(otherHandler))
	r.HandleFunc("/copy", addDefaultHeaders(receiveWork)).Methods("POST")
	fmt.Println("Serving...")
	http.ListenAndServe(":8080", r)
}

func main() {
	handleRequests()
}
