package main

import (
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

func doPost(url string, body string, contentType string) (string, int, error) {
	client := createClient(20)
	r := strings.NewReader(body)
	resp, err := client.Post(url, contentType, r)
	if err != nil {
		return "", 500, err
	}
	defer resp.Body.Close()
	respStr, _ := io.ReadAll(resp.Body)
	return string(respStr), resp.StatusCode, nil
}
