package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func createClient(timeout int) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: t,
	}
	return client
}

func doPost(url string, body string) (string, int) {
	client := createClient(20)
	r := strings.NewReader(body)
	resp, err := client.Post(url, "application/json", r)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()
	respStr, _ := io.ReadAll(resp.Body)
	return string(respStr), resp.StatusCode
}
