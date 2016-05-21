// +build integration

package main

// Without Docker Machine
// $ export HOST_IP=<HOST_IP>
// $ go build && go test --cover --tags integration

// With Docker Machine
// $ docker-machine create -d virtualbox docker-flow-test
// $ eval "$(docker-machine env docker-flow-test)"
// $ go build && go test --cover --tags integration | tee tests.log
// $ docker-machine rm -f docker-flow-test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite
	ConsulIp            string
	ProxyIp             string
	ProxyHost           string
	ProxyDockerHost     string
	ProxyDockerCertPath string
	ServicePath         string
	ServiceName         string
}

func (s *IntegrationTestSuite) SetupTest() {
	s.removeAll()
	time.Sleep(time.Second)
}

// Integration

func (s IntegrationTestSuite) XTest_BlueGreenDeployment() {
	origConsulAddress := os.Getenv("FLOW_CONSUL_ADDRESS")
	defer func() {
		os.Setenv("FLOW_CONSUL_ADDRESS", origConsulAddress)
	}()

	log.Println(">> Integration tests: deployment")

	log.Println("First deployment (blue)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--target", "app",
		"--side-target", "db",
		"--blue-green",
	)
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-blue_1", "Up"},
		{"godemo_db", "Up"},
	})

	log.Println("Second deployment (green)")
	os.Setenv("FLOW_CONSUL_ADDRESS", fmt.Sprintf("http://%s:8500", s.ConsulIp))
	s.runCmdWithStdOut(true, "./docker-flow", "--flow", "deploy")
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-blue_1", "Up"},
		{"godemo_app-green_1", "Up"},
	})

	log.Println("Third deployment (blue) with stop old release (green)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--flow", "deploy", "--flow", "stop-old")
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-blue_1", "Up"},
		{"godemo_app-green_1", "Exited"},
	})
}

func (s IntegrationTestSuite) XTest_Scaling() {
	log.Println(">> Integration tests: scaling")

	log.Println("First deployment (blue, 2 instances)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--flow", "deploy",
		"--scale", "2",
	)
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-blue_1", "Up"},
		{"godemo_app-blue_2", "Up"},
		{"godemo_db", "Up"},
	})

	log.Println("Second deployment (green, 4 (+2) instances)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--flow", "deploy",
		"--scale", "+2",
	)
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-green_1", "Up"},
		{"godemo_app-green_2", "Up"},
		{"godemo_app-green_3", "Up"},
		{"godemo_app-green_4", "Up"},
	})

	log.Println("Scaling (green, 3 (-1) instances)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--flow", "scale",
		"--scale", "\"-1\"",
	)
	s.verifyContainer([]ContainerStatus{
		{"godemo_app-green_1", "Up"},
		{"godemo_app-green_2", "Up"},
		{"godemo_app-green_3", "Up"},
		{"godemo_app-green_4", "N/A"},
	})
}

func (s IntegrationTestSuite) XTest_Proxy() {
	log.Println(">> Integration tests: proxy")

	s.runCmdWithStdOut(
		true,
		"docker", "run", "-d", "--name", "registrator",
		"-v", "/var/run/docker.sock:/tmp/docker.sock",
		"gliderlabs/registrator",
		"-ip", s.ConsulIp, fmt.Sprintf("consul://%s:8500", s.ConsulIp),
	)

	log.Println("Runs proxy when not present and reconfigures it when deploy")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--service-path", s.ServicePath,
		"--flow", "deploy", "--flow", "proxy",
	)
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Up"},
	})
	url := fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath)
	resp, err := http.Get(url)
	s.NoError(err)
	s.Equal(200, resp.StatusCode, "Failed to send the request %s", url)

	log.Println("Runs proxy when stopped and reconfigures it when scale")
	s.runCmdWithStdOut(false, "docker", "stop", "docker-flow-proxy")
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Exited"},
	})
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--service-path", s.ServicePath,
		"--scale", "+1",
		"--flow", "scale", "--flow", "proxy",
	)
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Up"},
	})
	resp, err = http.Get(fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath))
	s.NoError(err)
	s.Equal(200, resp.StatusCode)
	s.runCmdWithStdOut(true, "docker", "rm", "-f", "godemo_app-blue_1")
	resp, err = http.Get(fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath))
	s.NoError(err)
	s.Equal(200, resp.StatusCode)

	log.Println("Works as a standalone")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--flow", "deploy", "--flow", "stop-old",
	)
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--service-path", s.ServicePath,
		"--flow", "proxy",
	)
	resp, err = http.Get(fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath))
	s.NoError(err)
	s.Equal(200, resp.StatusCode)
}

func (s IntegrationTestSuite) Test_Proxy_Templates() {
	log.Println(">> Integration tests: proxy with templates")

	s.runCmdWithStdOut(
		true,
		"docker", "run", "-d", "--name", "registrator",
		"-v", "/var/run/docker.sock:/tmp/docker.sock",
		"gliderlabs/registrator",
		"-ip", s.ConsulIp, fmt.Sprintf("consul://%s:8500", s.ConsulIp),
	)
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--service-path", "INCORRECT",
		"--consul-template-path", "test_configs/tmpl/go-demo-app.tmpl",
		"--flow", "deploy", "--flow", "proxy",
	)
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Up"},
	})
	url := fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath)
	resp, err := http.Get(url)
	s.NoError(err)
	s.Equal(200, resp.StatusCode, "Failed to send the request %s", url)
}

// Util

type ContainerStatus struct {
	Name   string
	Status string
}

func (s IntegrationTestSuite) verifyContainer(csList []ContainerStatus) {
	s.runCmdWithStdOut(false, "docker", "ps", "-a")
	for _, cs := range csList {
		_, msg := s.runCmdWithoutStdOut(
			true,
			"docker", "ps", "-a",
			"--filter", fmt.Sprintf("name=%s", cs.Name),
			"--format", "\"{{.Names}} {{.Status}}\"",
		)

		if cs.Status == "N/A" {
			s.NotContains(msg, cs.Name)
		} else {
			s.Contains(msg, cs.Name)
			s.Contains(msg, cs.Status)
		}
	}
}

func (s IntegrationTestSuite) runCmd(failOnError, stdOut bool, command string, args ...string) (bool, string) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	msg := ""
	if stdOut {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &out
		cmd.Stderr = &out
	}
	err := cmd.Run()
	if !stdOut {
		msg = string(out.Bytes())
	}
	fmt.Printf("$ %s %s\n", command, strings.Join(args, " "))
	fmt.Println(msg)
	if err != nil {
		msgWithError := fmt.Sprintf("%s %s\n%s\n", command, strings.Join(args, " "), err.Error())
		if failOnError {
			log.Fatal(msgWithError)
		}
		return false, msgWithError
	}
	return true, msg
}

func (s IntegrationTestSuite) runCmdWithoutStdOut(failOnError bool, command string, args ...string) (bool, string) {
	return s.runCmd(failOnError, false, command, args...)
}

func (s IntegrationTestSuite) runCmdWithStdOut(failOnError bool, command string, args ...string) (bool, string) {
	return s.runCmd(failOnError, true, command, args...)
}

func (s *IntegrationTestSuite) removeAll() {
	_, ids := s.runCmdWithoutStdOut(true, "docker", "ps", "-a", "--filter", "name=dockerflow", "--format", "{{.ID}}")
	for _, id := range strings.Split(ids, "\n") {
		s.runCmdWithStdOut(false, "docker", "rm", "-f", string(id))
	}
	s.runCmdWithStdOut(false, "docker", "rm", "-f", "consul", "docker-flow-proxy", "registrator")
	s.runCmdWithStdOut(false, "docker-compose", "-f", "docker-compose-setup.yml", "-p", "tests-setup", "down")
	s.runCmdWithStdOut(
		true,
		"docker-compose", "-f", "docker-compose-setup.yml", "-p", "tests-setup", "up", "-d", "consul",
	)
}

// Suite

func TestIntegrationTestSuite(t *testing.T) {
	s := new(IntegrationTestSuite)
	ip := os.Getenv("HOST_IP")
	if len(ip) == 0 {
		_, msg := s.runCmdWithoutStdOut(true, "docker-machine", "ip", "docker-flow-test")
		ip = strings.Trim(msg, "\n")
	}
	s.ConsulIp = ip
	s.ProxyIp = ip
	s.ProxyHost = ip
	s.ProxyDockerHost = os.Getenv("DOCKER_HOST")
	s.ProxyDockerCertPath = os.Getenv("DOCKER_CERT_PATH")
	s.ServicePath = "/demo/hello"
	s.ServiceName = "go-demo"
	os.Setenv("FLOW_CONSUL_IP", s.ConsulIp)
	os.Setenv("FLOW_PROJECT", s.ServiceName)
	suite.Run(t, s)
}
