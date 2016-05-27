package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const containerStatusRunning = 1
const containerStatusExited = 2
const containerStatusRemoved = 3
const ProxyReconfigureDefaultPort = 8080
const ConsulTemplatesDir = "/consul_templates"

var haProxy Proxy = HaProxy{}

type HaProxy struct{}

var runHaProxyRunCmd = runCmd
var runHaProxyExecCmd = runCmd
var runHaProxyCpCmd = runCmd
var runHaProxyPsCmd = runCmd
var runHaProxyStartCmd = runCmd
var httpGet = http.Get

func (m HaProxy) Provision(dockerHost, reconfPort, certPath, scAddress string) error {
	if len(dockerHost) == 0 {
		return fmt.Errorf("Proxy docker host is mandatory for the proxy step. Please set the proxy-docker-host argument.")
	}
	if len(scAddress) == 0 {
		return fmt.Errorf("Service Discovery Address is mandatory.")
	}
	SetDockerHost(dockerHost, certPath)
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

// TODO: Change args to struct
func (m HaProxy) Reconfigure(
	dockerHost, dockerCertPath, host, reconfPort, serviceName, serviceColor string,
	servicePath []string,
	consulTemplateFePath string, consulTemplateBePath string,
) error {
	if len(consulTemplateFePath) > 0 {
		if err := m.sendConsulTemplatesToTheProxy(dockerHost, dockerCertPath, consulTemplateFePath, consulTemplateBePath, serviceName, serviceColor); err != nil {
			return err
		}
	} else if len(servicePath) == 0 {
		return fmt.Errorf("It is mandatory to specify servicePath or consulTemplatePath. Please set one of the two.")
	}
	if len(host) == 0 {
		return fmt.Errorf("Proxy host is mandatory for the proxy step. Please set the proxy-host argument.")
	}
	if len(serviceName) == 0 {
		return fmt.Errorf("Service name is mandatory for the proxy step.")
	}
	if len(reconfPort) == 0 && !strings.Contains(host, ":") {
		return fmt.Errorf("Reconfigure port is mandatory.")
	}
	if err := m.sendReconfigureRequest(host, reconfPort, serviceName, serviceColor, servicePath, consulTemplateFePath, consulTemplateBePath); err != nil {
		return err
	}
	return nil
}

func (m HaProxy) sendReconfigureRequest(
	host, reconfPort, serviceName, serviceColor string,
	servicePath []string,
	consulTemplateFePath, consulTemplateBePath string,
) error {
	address := host
	if len(reconfPort) > 0 {
		address = fmt.Sprintf("%s:%s", host, reconfPort)
	}
	if !strings.HasPrefix(strings.ToLower(address), "http") {
		address = fmt.Sprintf("http://%s", address)
	}
	proxyUrl := fmt.Sprintf(
		"%s/v1/docker-flow-proxy/reconfigure?serviceName=%s",
		address,
		serviceName,
	)
	if len(consulTemplateFePath) > 0 {
		proxyUrl = fmt.Sprintf("%s&consulTemplateFePath=%s/%s-fe.tmpl&consulTemplateBePath=%s/%s-be.tmpl", proxyUrl, ConsulTemplatesDir, serviceName, ConsulTemplatesDir, serviceName)
	} else {
		if len(serviceColor) > 0 {
			proxyUrl = fmt.Sprintf("%s&serviceColor=%s", proxyUrl, serviceColor)
		}
		proxyUrl = fmt.Sprintf("%s&servicePath=%s", proxyUrl, strings.Join(servicePath, ","))
	}
	logPrintf("Sending request to %s to reconfigure the proxy", proxyUrl)
	resp, err := httpGet(proxyUrl)
	if err != nil {
		return fmt.Errorf("The request to reconfigure the proxy failed\n%s\n", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("The request to the proxy (%s) failed with status code %d\n", proxyUrl, resp.StatusCode)
	}
	return nil
}

func (m HaProxy) sendConsulTemplatesToTheProxy(dockerHost, dockerCertPath, consulTemplateFePath, consulTemplateBePath, serviceName, color string) error {
	if err := m.sendConsulTemplateToTheProxy(dockerHost, dockerCertPath, consulTemplateFePath, serviceName, color, "fe"); err != nil {
		return err
	}
	if err := m.sendConsulTemplateToTheProxy(dockerHost, dockerCertPath, consulTemplateBePath, serviceName, color, "be"); err != nil {
		return err
	}
	return nil
}

func (m HaProxy) sendConsulTemplateToTheProxy(dockerHost, dockerCertPath, consulTemplatePath, serviceName, color, templateType string) error {
	if err := m.createTempConsulTemplate(consulTemplatePath, serviceName, color); err != nil {
		return err
	}
	file := fmt.Sprintf("%s-%s.tmpl", serviceName, templateType)
	if err := m.copyConsulTemplateToTheProxy(dockerHost, dockerCertPath, consulTemplatePath, file); err != nil {
		return err
	}
	removeFile(fmt.Sprintf("%s.tmp", consulTemplatePath))
	return nil
}

func (m HaProxy) copyConsulTemplateToTheProxy(dockerHost, dockerCertPath, consulTemplatePath, templateName string) error {
	SetDockerHost(dockerHost, dockerCertPath)
	args := []string{"exec", "-i", "docker-flow-proxy", "mkdir", "-p", ConsulTemplatesDir}
	execCmd := exec.Command("docker", args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	// TODO: Remove. Deprecated since Docker Flow: Proxy has that directory by default.
	if err := runHaProxyExecCmd(execCmd); err != nil {
		return err
	}
	args = []string{
		"cp",
		fmt.Sprintf("%s.tmp", consulTemplatePath),
		fmt.Sprintf("docker-flow-proxy:%s/%s", ConsulTemplatesDir, templateName),
	}
	cpCmd := exec.Command("docker", args...)
	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	if err := runHaProxyCpCmd(cpCmd); err != nil {
		return err
	}
	return nil
}

func (m HaProxy) createTempConsulTemplate(consulTemplatePath, serviceName, color string) error {
	fullServiceName := fmt.Sprintf("%s-%s", serviceName, color)
	tmpPath := fmt.Sprintf("%s.tmp", consulTemplatePath)
	data, err := readFile(consulTemplatePath)
	if err != nil {
		return fmt.Errorf("Could not read the Consul template %s\n%s", consulTemplatePath, err.Error())
	}
	if err := writeFile(
		tmpPath,
		[]byte(strings.Replace(string(data), "SERVICE_NAME", fullServiceName, -1)),
		0644,
	); err != nil {
		return fmt.Errorf("Could not write temporary Consul template to %s\n%s", tmpPath, err.Error())
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
	args := []string{"start", "docker-flow-proxy"}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := runHaProxyStartCmd(cmd); err != nil {
		return fmt.Errorf("Docker start command failed\n%v\n%v\n", cmd, err)
	}
	return nil
}
