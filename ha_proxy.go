package main

import (
	"os/exec"
	"os"
	"fmt"
	"bytes"
	"strings"
	"net/http"
	"time"
)


const containerStatusRunning = 1
const containerStatusExited = 2
const containerStatusRemoved = 3
const ProxyReconfigureDefaultPort = 8080

var haProxy Proxy = HaProxy{}
type HaProxy struct {}

var runHaProxyRunCmd = runCmd
var runHaProxyPsCmd = runCmd
var runHaProxyStartCmd = runCmd
var httpGet = http.Get

func (m HaProxy) Provision(host, reconfPort, certPath, scAddress string) error {
	if len(host) == 0 {
		return fmt.Errorf("Proxy docker host is mandatory for the proxy step. Please set the proxy-docker-host argument.")
	}
	if len(scAddress) == 0 {
		return fmt.Errorf("Service Discovery Address is mandatory.")
	}
	SetDockerHost(host, certPath)
	status, err := m.ps()
	if err != nil {
		return err
	}
	switch status {
	case containerStatusRunning:
		return nil
	case containerStatusExited:
		if err := m.start(); err != nil {
			return err
		}
		sleep(time.Second * 5)
	default:
		if err := m.run(reconfPort, scAddress); err != nil {
			return err
		}
		sleep(time.Second * 5)
	}
	return nil
}

func (m HaProxy) Reconfigure(host, reconfPort, serviceName, serviceColor string, servicePath []string) error {
	if len(host) == 0 {
		return fmt.Errorf("Proxy host is mandatory for the proxy step. Please set the proxy-host argument.")
	}
	if len(serviceName) == 0 {
		return fmt.Errorf("Service name is mandatory for the proxy step.")
	}
	if len(servicePath) == 0 {
		return fmt.Errorf("Service path is mandatory.")
	}
	if len(reconfPort) == 0 {
		return fmt.Errorf("Reconfigure port is mandatory.")
	}
	address := fmt.Sprintf("%s:%s", host, reconfPort)
	if !strings.HasPrefix(strings.ToLower(address), "http") {
		address = fmt.Sprintf("http://%s", address)
	}
	colorQuery := ""
	if (len(serviceColor) > 0) {
		colorQuery = fmt.Sprintf("&serviceColor=%s", serviceColor)
	}
	resp, err := httpGet(fmt.Sprintf(
		"%s/v1/docker-flow-proxy/reconfigure?serviceName=%s%s&servicePath=%s",
		address,
		serviceName,
		colorQuery,
		strings.Join(servicePath, ","),
	))
	if err != nil {
		return fmt.Errorf("The request to reconfigure the proxy failed\n%s\n", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("The response from the proxy was incorrect\n%s\n", err.Error())
	}
	return nil
}

func (m HaProxy) run(reconfPort, scAddress string) error {
	logPrintln("Running the docker-flow-proxy container...")
	args := []string{
		"run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", scAddress),
		"-p", "80:80", "-p", fmt.Sprintf("%s:8080", reconfPort),
		"vfarcic/docker-flow-proxy",
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := runHaProxyRunCmd(cmd); err != nil {
		return fmt.Errorf("Docker run command failed\n%v\n%v\n", cmd, err)
	}
	return nil
}

func (m HaProxy) ps() (int, error) {
	logPrintln("Checking status of the docker-flow-proxy container...")
	args := []string{
		"ps", "-a",
		"--filter", "name=docker-flow-proxy",
		"--format", "{{.Status}}",
	}
	cmd := exec.Command("docker", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := runHaProxyPsCmd(cmd); err != nil {
		return 0, fmt.Errorf("Docker ps command failed\n%v\n%v\n", cmd, err)
	}
	status := string(out.Bytes())
	if strings.HasPrefix(status, "Exited") {
		return containerStatusExited, nil
	}
	if len(status) == 0 {
		return containerStatusRemoved, nil
	}
	return containerStatusRunning, nil
}

func (m HaProxy) start() error {
	logPrintln("Starting the docker-flow-proxy container...")
	args := []string{ "start", "docker-flow-proxy" }
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := runHaProxyStartCmd(cmd); err != nil {
		return fmt.Errorf("Docker start command failed\n%v\n%v\n", cmd, err)
	}
	return nil
}
