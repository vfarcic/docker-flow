package main

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
)

type Response struct {
	MyResponse	string `json:"my-response"`
}

// TODO
func startWebServer() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// TODO
func handler(w http.ResponseWriter, r *http.Request) {
	var response = Response{
		MyResponse: "Hello World!",
	}
	data, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(data))
}