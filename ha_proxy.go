package main

import (
	"os/exec"
	"os"
	"fmt"
	"bytes"
	"strings"
)

const containerStatusRunning = 1
const containerStatusExited = 2
const containerStatusRemoved = 3

var haProxy Proxy = HaProxy{}
type HaProxy struct {}

var runHaProxyRunCmd = runCmd
var runHaProxyPsCmd = runCmd
var runHaProxyStartCmd = runCmd

func (m HaProxy) Provision(host, scAddress string) error {
	if len(host) == 0 {
		return fmt.Errorf("Proxy host is mandatory for the proxy step. Please set the proxy-host argument.")
	}
	SetDockerHost(host)
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
	default:
		if err := m.run(scAddress); err != nil {
			return err
		}
	}
	return nil
}

func (m HaProxy) run(scAddress string) error {
	logPrintln("Running the docker-flow-proxy container...")
	args := []string{
		"run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", scAddress),
		"-p", "80:80",
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
