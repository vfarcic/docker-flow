// +build integration
package main

// To run locally on OS X
// $ docker-machine create -d virtualbox testing
// $ eval "$(docker-machine env testing)"
// $ go build && go test --cover --tags integration -v
// $ docker-machine rm -f testing

// TODO: Change books-ms for a lighter service

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
	"os/exec"
	"strings"
	"log"
	"bytes"
	"os"
)

type IntegrationTestSuite struct {
	suite.Suite
	ConsulIp	string
}

func (s *IntegrationTestSuite) SetupTest() {
}

// Integration

func (s IntegrationTestSuite) Test_Integration_BlueGreenDeployment() {
	log.Println("First deployment (blue)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--consul-address", fmt.Sprintf("http://%s:8500", s.ConsulIp),
		"--target", "app",
		"--side-target", "db",
		"--blue-green",
	)
	s.runCmdWithStdOut(false, "docker", "ps", "-a")
	s.verifyContainer("booksms_app-blue_1", "Up")
	s.verifyContainer("books-ms-db", "Up")

	log.Println("Second deployment (green)")
	origConsulAddress := os.Getenv("FLOW_CONSUL_ADDRESS")
	defer func() {
		os.Setenv("FLOW_CONSUL_ADDRESS", origConsulAddress)
	}()
	os.Setenv("FLOW_CONSUL_ADDRESS", fmt.Sprintf("http://%s:8500", s.ConsulIp))
	s.runCmdWithStdOut(true, "./docker-flow", "--flow", "deploy")
	s.runCmdWithStdOut(false, "docker", "ps", "-a")
	s.verifyContainer("booksms_app-blue_1", "Up")
	s.verifyContainer("booksms_app-green_1", "Up")

	log.Println("Third deployment (blue) with stop old release (green)")
	s.runCmdWithStdOut(
		true,
		"./docker-flow",
		"--flow", "deploy", "--flow", "stop-old")
	s.runCmdWithStdOut(false, "docker", "ps", "-a")
	s.verifyContainer("booksms_app-blue_1", "Up")
	s.verifyContainer("booksms_app-green_1", "Exited")

	log.Println("Cleanup")
	s.runCmdWithStdOut(
		false,
		"docker", "rm", "-f",
		"booksms_app-blue_1", "booksms_app-green_1", "books-ms-db", "consul")
}

func (s IntegrationTestSuite) verifyContainer(name, status string) {
	_, msg := s.runCmdWithoutStdOut(
		true,
		"docker", "ps", "-a",
		"--filter", fmt.Sprintf("name=%s", name),
		"--format", "\"{{.Names}} {{.Status}}\"",
	)

	s.Contains(msg, name)
	s.Contains(msg, status)
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
	var msg string
	s := new(IntegrationTestSuite)
	s.runCmdWithStdOut(false, "docker", "rm", "-f", "consul")
	s.runCmdWithStdOut(
		true,
		"docker", "run", "-d", "--name", "consul",
		"-p", "8500:8500",
		"-h", "consul",
		"progrium/consul", "-server", "-bootstrap",
	)
	_, msg = s.runCmdWithoutStdOut(true, "docker-machine", "ip", "testing")
	s.ConsulIp = strings.Trim(msg, "\n")
	os.Setenv("CONSUL_IP", s.ConsulIp)
	os.Setenv("FLOW_PROJECT", "booksms")
	suite.Run(t, s)
}