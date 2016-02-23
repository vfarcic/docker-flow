package main

import (
	"strconv"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
)

func getConsulScale(consulAddress, serviceName, scale string) (int, error) {
	s := 1
	inc := 0
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/scale?raw", consulAddress, serviceName))
	if err != nil {
		return 0, fmt.Errorf("Please make sure that Consul address is correct\n%v", err)
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
		return 1, nil
	}
	return total, nil
}

func getConsulNextColor(consulAddress, serviceName string) (string, error) {
	color, err := getConsulColor(consulAddress, serviceName)
	if err != nil {
		return "", err
	}
	if color == "blue" {
		return "green", nil
	}
	return "blue", nil
}

func getConsulColor(consulAddress, serviceName string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/color?raw", consulAddress, serviceName))
	if err != nil {
		return "", fmt.Errorf("Could not retrieve the color from Consul. Please make sure that Consul address is correct\n%v", err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	currColor := "green"
	if len(data) > 0 {
		currColor = string(data)
	}
	return currColor, nil
}

func putConsulScale(consulAddress, serviceName string, scale int) error {
	return putConsul(consulAddress, serviceName, "scale", strconv.Itoa(scale))
}

func putConsulColor(consulAddress, serviceName, color string) error {
	return putConsul(consulAddress, serviceName, "color", color)
}

func putConsul(consulAddress, serviceName, key, value string) error {
	url := fmt.Sprintf("%s/v1/kv/docker-flow/%s/%s", consulAddress, serviceName, key)
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(value))
	_, err = client.Do(request)
	if err != nil {
		return fmt.Errorf("Could not store scale information in Consul\n%v", err)
	}
	return nil
}
