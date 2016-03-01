package main

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"os"
	"fmt"
	"os/exec"
)

type DockerComposeTestSuite struct {
	suite.Suite
	dockerComposePath string
	target            string
	color             string
	blueGreen         bool
	host 			  string
	project 		  string
}

func (s *DockerComposeTestSuite) SetupTest() {
	s.dockerComposePath = "test-docker-compose.yml"
	s.target = "my-target"
	s.color = "red"
	s.blueGreen = false
	s.host = "tcp://1.2.3.4:1234"
	s.project = "my-project"
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), nil
	}
	writeFile = func(fileName string, data []byte, perm os.FileMode) error {
		return nil
	}
	removeFile = func(name string) error {
		return nil
	}
	execCmd = func(name string, arg ...string) *exec.Cmd {
		return &exec.Cmd{}
	}
}

// CreateFlow

func (s DockerComposeTestSuite) Test_CreateFlow_ReturnsNil() {
	actual := DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, s.target, s.color, s.blueGreen)

	s.Nil(actual)
}

func (s DockerComposeTestSuite) Test_CreateFlow_ReturnsError_WhenReadFile() {
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("Some error")
	}

	err := DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, s.target, s.color, s.blueGreen)

	s.Error(err)
}

func (s DockerComposeTestSuite) Test_CreateFlow_CreatesTheFile() {
	var actual string
	writeFile = func(filename string, data []byte, perm os.FileMode) error {
		actual = filename
		return nil
	}

	DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, s.target, s.color, s.blueGreen)

	s.Equal(dockerComposeFlowPath, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlow_CreatesDockerComposeReplica() {
	var actual string
	readFile = func(filename string) ([]byte, error) {
		actual = filename
		return []byte(""), nil
	}

	DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, s.target, s.color, s.blueGreen)

	s.Equal(s.dockerComposePath, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlow_ReplacesTarget_WhenBlueGreen() {
	color := "orange"
	target := "app"
	var actual string
	var dcTemplate = `
services:
  %s%s:
    image: vfarcic/books-ms`
	var expected = fmt.Sprintf(dcTemplate, target, "-" + color)
	readFile = func(filename string) ([]byte, error) {
		return []byte(fmt.Sprintf(dcTemplate, target, "")), nil
	}
	writeFile = func(filename string, data []byte, perm os.FileMode) error {
		actual = string(data)
		return nil
	}

	DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, target, color, true)

	s.Equal(expected, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlow_ReturnsError_WhenWriteFile() {
	writeFile = func(filename string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("Some error")
	}

	err := DockerComposeImpl{}.CreateFlow(s.dockerComposePath, dockerComposeFlowPath, s.target, s.color, s.blueGreen)

	s.Error(err)
}

// RemoveFlow

func (s DockerComposeTestSuite) Test_RemoveFlow_RemovesTheFile() {
	var actual string
	removeFile = func(name string) error {
		actual = name
		return nil
	}

	DockerComposeImpl{}.RemoveFlow()

	s.Equal(dockerComposeFlowPath, actual)
}

func (s DockerComposeTestSuite) Test_RemoveFlow_ReturnsError() {
	removeFile = func(name string) error {
		return fmt.Errorf("Some error")
	}

	err := DockerComposeImpl{}.RemoveFlow()

	s.Error(err)
}

// PullTargets

func (s DockerComposeTestSuite) Test_PullTargets_ReturnsNil_WhenTargetsAreEmpty() {
	actual := DockerComposeImpl{}.PullTargets(s.host, s.project, []string{})

	s.Nil(actual)
}

// PullTargets

func (s DockerComposeTestSuite) Test_PullTargets() {
	s.testCmd(DockerComposeImpl{}.PullTargets, "pull", s.target)
}

// UpTargets

func (s DockerComposeTestSuite) Test_UpTargets() {
	s.testCmd(DockerComposeImpl{}.UpTargets, "up", "-d", s.target)
}

// ScaleTargets

func (s DockerComposeTestSuite) Test_ScaleTargets_ReturnsNil_WhenTargetIsEmpty() {
	actual := DockerComposeImpl{}.ScaleTargets(s.host, s.project, "", 8)

	s.Nil(actual)
}

func (s DockerComposeTestSuite) Test_ScaleTargets_CreatesTheCommand() {
	var scale = 7
	expected := []string{"docker-compose", "-f", dockerComposeFlowPath, "-p", s.project, "scale", fmt.Sprintf("%s=%d", s.target, scale)}
	actual := s.mockExecCmd()

	DockerComposeImpl{}.ScaleTargets(s.host, s.project, s.target, scale)

	s.Equal(expected, *actual)
}

// RmTargets

func (s DockerComposeTestSuite) Test_RmTargets() {
	s.testCmd(DockerComposeImpl{}.RmTargets, "rm", "-f", s.target)
}

// StopTargets

func (s DockerComposeTestSuite) Test_StopTargets() {
	s.testCmd(DockerComposeImpl{}.StopTargets, "stop", s.target)
}

// Suite

func TestDockerComposeTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeTestSuite))
}

// Helper

func (s DockerComposeTestSuite) mockExecCmd() *[]string {
	var actualCommand []string
	execCmd = func(name string, arg ...string) *exec.Cmd {
		actualCommand = append([]string{name}, arg...)
		cmd := &exec.Cmd{}
		return cmd
	}
	return &actualCommand
}

type testCmdType func(host, project string, targets []string) error

func (s DockerComposeTestSuite) testCmd(f testCmdType, args ...string) {
	var expected []string
	var actual *[]string

	// Returns nil when targets are empty
	s.Nil(f(s.host, s.project, []string{}))

	// Creates command
	expected = append([]string{"docker-compose", "-f", dockerComposeFlowPath, "-p", s.project}, args...)
	actual = s.mockExecCmd()
	f(s.host, s.project, []string{s.target})
	s.Equal(expected, *actual)


	// Does not add project when empty
	expected = append([]string{"docker-compose", "-f", dockerComposeFlowPath}, args...)
	actual = s.mockExecCmd()
	f(s.host, "", []string{s.target})
	s.Equal(expected, *actual)

	// Adds DOCKER_HOST variable
	f(s.host, s.project, []string{s.target})
	s.Contains(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", s.host))

	// Does not add DOCKER_HOST variable when empty
	f("", s.project, []string{s.target})
	s.NotContains(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", s.host))
}