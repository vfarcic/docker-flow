package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
)

// TODO
func getCurrentColor() {
	// TODO: Change to param
	resp, err := http.Get("http://technologyconversations.com")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(body))
}

