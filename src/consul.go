package main
import (
	"strconv"
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func getConsulScale(consulAddress, serviceName, scale string) int {
	s := 1
	inc := 0
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/scale?raw", consulAddress, serviceName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Please make sure that Consul address is correct\n%v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	if len(data) > 0 {
		s, _ = strconv.Atoi(string(data))
	}
	if len(scale) > 0 {
		if scale[:1] == "+" || scale[:1] == "-" {
			inc, _ = strconv.Atoi(scale)
		} else {
			s, _ = strconv.Atoi(scale)
		}
	}
	total := s + inc
	if total <= 0 {
		return 1
	}
	return total
}

func getConsulNextColor(consulAddress, serviceName string) string {
	if getConsulColor(consulAddress, serviceName) == "blue" {
		return "green"
	}
	return "blue"
}

func getConsulColor(consulAddress, serviceName string) string {
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/color?raw", consulAddress, serviceName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Please make sure that Consul address is correct\n%v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	currColor := "green"
	if len(data) > 0 {
		currColor = string(data)
	}
	return currColor
}

func putConsulScale(consulAddress, serviceName string, scale int) {
	putConsul(consulAddress, serviceName, "scale", strconv.Itoa(scale))
}

func putConsulColor(consulAddress, serviceName, color string) {
	putConsul(consulAddress, serviceName, "color", color)
}

func putConsul(consulAddress, serviceName, key, value string) {
	url := fmt.Sprintf("%s/v1/kv/docker-flow/%s/%s", consulAddress, serviceName, key)
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(value))
	_, err = client.Do(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not store scale information in Consul\n%v", err)
		os.Exit(1)
	}
}
