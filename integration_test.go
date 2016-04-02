// +build integration
package main

// To run locally on OS X
// $ docker-machine create -d virtualbox testing
// $ eval "$(docker-machine env testing)"
// $ go build && go test --cover --tags integration -v
// $ docker-machine rm -f testing

// TODO: Change books-ms for a "lighter" service

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
	"os/exec"
	"strings"
	"log"
	"bytes"
	"os"
	"time"
	"net/http"
)

type IntegrationTestSuite struct {
	suite.Suite
	ConsulIp			string
	ProxyHost			string
	ProxyDockerHost 	string
	ProxyDockerCertPath string
	ServicePath			string
}

func (s *IntegrationTestSuite) SetupTest() {
	_, ids := s.runCmdWithoutStdOut(true, "docker", "ps", "-a", "--filter", "name=booksms", "--format", "{{.ID}}")
	for _, id := range strings.Split(ids, "\n") {
		s.runCmdWithStdOut(false, "docker", "rm", "-f", string(id))
	}
	s.runCmdWithStdOut(false, "docker", "rm", "-f", "consul", "docker-flow-proxy")
	s.runCmdWithStdOut(
		true,
		"docker", "run", "-d", "--name", "consul",
		"-p", "8500:8500",
		"-h", "consul",
		"progrium/consul", "-server", "-bootstrap",
	)
	time.Sleep(time.Second)
}

// Integration

func (s IntegrationTestSuite) Test_BlueGreenDeployment() {
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
		{"booksms_app-blue_1", "Up" },
		{"books-ms-db", "Up" },
	})

	log.Println("Second deployment (green)")
	os.Setenv("FLOW_CONSUL_ADDRESS", fmt.Sprintf("http://%s:8500", s.ConsulIp))
	s.runCmdWithStdOut(true, "./docker-flow", "--flow", "deploy")
	s.verifyContainer([]ContainerStatus{
		{"booksms_app-blue_1", "Up" },
		{"booksms_app-green_1", "Up" },
	})

	log.Println("Third deployment (blue) with stop old release (green)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--flow", "deploy", "--flow", "stop-old")
	s.verifyContainer([]ContainerStatus{
		{"booksms_app-blue_1", "Up" },
		{"booksms_app-green_1", "Exited" },
	})
}

func (s IntegrationTestSuite) Test_Scaling() {
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
		{"booksms_app-blue_1", "Up" },
		{"booksms_app-blue_2", "Up" },
		{"books-ms-db", "Up" },
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
		{"booksms_app-green_1", "Up" },
		{"booksms_app-green_2", "Up" },
		{"booksms_app-green_3", "Up" },
		{"booksms_app-green_4", "Up" },
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
		{"booksms_app-green_1", "Up" },
		{"booksms_app-green_2", "Up" },
		{"booksms_app-green_3", "Up" },
		{"booksms_app-green_4", "N/A" },
	})
}

func (s IntegrationTestSuite) Test_Proxy() {
	log.Println(">> Integration tests: proxy")

	log.Println("Runs docker-flow-proxy when not present")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--flow", "proxy",
	)
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Up" },
	})

	log.Println("Runs docker-flow-proxy when stopped")
	s.runCmdWithStdOut(false, "docker", "stop", "docker-flow-proxy")
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Exited" },
	})
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--proxy-host", s.ProxyHost,
		"--proxy-docker-host", s.ProxyDockerHost,
		"--proxy-docker-cert-path", s.ProxyDockerCertPath,
		"--flow", "proxy",
	)
	s.verifyContainer([]ContainerStatus{
		{"docker-flow-proxy", "Up" },
	})

	log.Println("Reconfigures proxy when deploy")
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
	fmt.Println(fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath))
	log.Fatal("xxx")
	resp, err := http.Get(fmt.Sprintf("http://%s%s", s.ConsulIp, s.ServicePath))
	s.NoError(err)
	s.Equal(201, resp.StatusCode)
}

// Util
type ContainerStatus struct {
	Name	string
	Status 	string
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

// Suite

func TestIntegrationTestSuite(t *testing.T) {
	s := new(IntegrationTestSuite)
	_, msg := s.runCmdWithoutStdOut(true, "docker-machine", "ip", "testing")
	ip := strings.Trim(msg, "\n")
	s.ConsulIp = ip
	s.ProxyHost = ip
	s.ProxyDockerHost = os.Getenv("DOCKER_HOST")
	s.ProxyDockerCertPath = os.Getenv("DOCKER_CERT_PATH")
	s.ServicePath = "/api/v1/books"
	os.Setenv("FLOW_CONSUL_IP", s.ConsulIp)
	os.Setenv("FLOW_PROJECT", "booksms")
	suite.Run(t, s)
}