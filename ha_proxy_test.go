package main

import (
"github.com/stretchr/testify/suite"
"testing"
	"os/exec"
	"fmt"
)

type HaProxyTestSuite struct {
	suite.Suite
	ScAddress string
	Host      string

}

func (s *HaProxyTestSuite) SetupTest() {
	s.ScAddress = "1.2.3.4:1234"
	s.Host = "tcp://my-docker-proxy-host"
	runHaProxyCmd = func(cmd *exec.Cmd) error {
		return nil
	}
}

// Provision

func (s HaProxyTestSuite) Test_Provision_RunsDockerFlowProxyContainer() {
	var actual []string
	expected := []string{
		"docker", "run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", s.ScAddress),
		"-p", "80:80",
		"vfarcic/docker-flow-proxy", "run",
	}
	runHaProxyCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_SetsDockerHost() {
	actual := ""
	SetDockerHost = func(host string) {
		actual = host
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(s.Host, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenFailure() {
	runHaProxyCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyHostIsEmpty() {
	err := HaProxy{}.Provision("", s.ScAddress)

	s.Error(err)
}

// TODO: Test that container is run only if not present


// Suite

func TestHaProxyTestSuite(t *testing.T) {
	suite.Run(t, new(HaProxyTestSuite))
}
