package dockerflow

import (
	"strconv"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
)

const ConsulScaleKey = "scale"
const ConsulColorKey = "color"

type Consul struct{}

func (c Consul) GetScaleCalc(address, serviceName, scale string) (int, error) {
	s := 1
	inc := 0
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/scale?raw", address, serviceName))
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
	fmt.Println(string(total))
	if total <= 0 {
		return 1, nil
	}
	return total, nil
}

func (c Consul) GetColor(address, serviceName string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v1/kv/docker-flow/%s/color?raw", address, serviceName))
	if err != nil {
		return "", fmt.Errorf("Could not retrieve the color from Consul. Please make sure that Consul address is correct\n%v", err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	currColor := GreenColor
	if len(data) > 0 {
		currColor = string(data)
	}
	return currColor, nil
}

func (c Consul) GetNextColor(currentColor string) string {
	if currentColor == BlueColor {
		return GreenColor
	}
	return BlueColor
}

func (c Consul) PutScale(address, serviceName string, value int) (string, error) {
	return c.putValue(address, serviceName, ConsulScaleKey, strconv.Itoa(value))
}

func (c Consul) PutColor(address, serviceName string, value string) (string, error) {
	return c.putValue(address, serviceName, ConsulColorKey, value)
}

func (c Consul) putValue(address, serviceName, key, value string) (string, error) {
	url := fmt.Sprintf("%s/v1/kv/docker-flow/%s/%s", address, serviceName, key)
	client := &http.Client{}
	request, _ := http.NewRequest("PUT", url, strings.NewReader(value))
	resp, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("Could not store scale information in Consul\n%v", err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return string(data), nil
}
