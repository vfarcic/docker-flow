package main

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"os/exec"
	"fmt"
)

type HaProxyTestSuite struct {
	suite.Suite
	ScAddress 		string
	Host      		string
	ExitedMessage 	string

}

func (s *HaProxyTestSuite) SetupTest() {
	s.ScAddress = "1.2.3.4:1234"
	s.Host = "tcp://my-docker-proxy-host"
	s.ExitedMessage = "Exited (2) 15 seconds ago"
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	logPrintln = func(v ...interface{}) { }
}

// Provision

func (s HaProxyTestSuite) Test_Provision_SetsDockerHost() {
	actual := ""
	SetDockerHost = func(host string) {
		actual = host
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(s.Host, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyHostIsEmpty() {
	err := HaProxy{}.Provision("", s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_RunsDockerFlowProxyContainer() {
	var actual []string
	expected := []string{
		"docker", "run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", s.ScAddress),
		"-p", "80:80",
		"vfarcic/docker-flow-proxy",
	}
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenFailure() {
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_RunsDockerPs() {
	var actual []string
	expected := []string{
		"docker", "ps", "-a",
		"--filter", "name=docker-flow-proxy",
		"--format", "{{.Status}}",
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenPsFailure() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyFails() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_DoesNotRun_WhenProxyExists() {
	actual := false
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		actual = true
		return nil
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		cmd.Stdout.Write([]byte("Up 3 hours"))
		return nil
	}
	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.False(actual)
}

func (s HaProxyTestSuite) Test_Provision_StartsAndDoesNotRun_WhenProxyIsExited() {
	start := false
	run := false
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		start = true
		return nil
	}
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		run = true
		return nil
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		cmd.Stdout.Write([]byte(s.ExitedMessage))
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.True(start)
	s.False(run)
}

func (s HaProxyTestSuite) Test_Provision_StartsDockerFlowProxyContainer_WhenProxyIsExited() {
	var actual []string
	expected := []string{ "docker", "start", "docker-flow-proxy" }
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		cmd.Stdout.Write([]byte(s.ExitedMessage))
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenStartFailure() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		cmd.Stdout.Write([]byte(s.ExitedMessage))
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ScAddress)

	s.Error(err)
}


// TODO: Test that container is run only if not present


// Suite

func TestHaProxyTestSuite(t *testing.T) {
	suite.Run(t, new(HaProxyTestSuite))
}
