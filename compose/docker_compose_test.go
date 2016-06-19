package compose

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"../util"
"testing"
)

type DockerComposeTestSuite struct {
	suite.Suite
	dockerComposePath string
	serviceName       string
	target            string
	sideTargets       []string
	color             string
	blueGreen         bool
	host              string
	certPath          string
	project           string
}

func (s *DockerComposeTestSuite) SetupTest() {
	s.dockerComposePath = "test-docker-compose.yml"
	s.serviceName = "myService"
	s.target = "my-target"
	s.sideTargets = []string{"my-side-target-1", "my-side-target-2"}
	s.color = "red"
	s.blueGreen = false
	s.host = "tcp://1.2.3.4:1234"
	s.certPath = "/path/to/docker/cert"
	s.project = "my-project"
	util.ReadFile = func(fileName string) ([]byte, error) {
		return []byte(""), nil
	}
	util.WriteFile = func(fileName string, data []byte, perm os.FileMode) error {
		return nil
	}
	util.RemoveFile = func(name string) error {
		return nil
	}
	util.ExecCmd = func(name string, arg ...string) *exec.Cmd {
		return &exec.Cmd{}
	}
}

// GetDockerCompose

func (s DockerComposeTestSuite) Test_GetDockerCompose_ReturnsDockerCompose() {
	actual := GetDockerCompose()

	s.Equal(dockerCompose, actual)
}

// CreateFlow

func (s DockerComposeTestSuite) Test_CreateFlowFile_ReturnsNil() {
	actual := DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, s.color, s.blueGreen)

	s.Nil(actual)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_ReturnsError_WhenReadFile() {
	util.ReadFile = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("Some error")
	}

	err := DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, s.color, s.blueGreen)

	s.Error(err)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_CreatesTheFile() {
	var actual string
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		actual = filename
		return nil
	}

	DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, s.color, s.blueGreen)

	s.Equal(dockerComposeFlowPath, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_CreatesDockerComposeReplica() {
	var actual string
	util.ReadFile = func(filename string) ([]byte, error) {
		actual = filename
		return []byte(""), nil
	}

	DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, s.color, s.blueGreen)

	s.Equal(s.dockerComposePath, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_CreatesNewTarget_WhenBlueGreen() {
	color := "orange"
	var actual string
	var dcContent = fmt.Sprintf(`
%s:
  image: vfarcic/books-ms`,
		s.target,
	)
	newTarget := fmt.Sprintf("%s-%s", s.target, color)
	expected := fmt.Sprintf(`%s:
  extends:
    file: %s
    service: %s
  environment:
    - SERVICE_NAME=%s-%s
%s:
  extends:
    file: %s
    service: %s
%s:
  extends:
    file: %s
    service: %s`,
		newTarget,
		s.dockerComposePath,
		s.target,
		s.serviceName,
		color,
		s.sideTargets[0],
		s.dockerComposePath,
		s.sideTargets[0],
		s.sideTargets[1],
		s.dockerComposePath,
		s.sideTargets[1],
	)
	util.ReadFile = func(filename string) ([]byte, error) {
		return []byte(dcContent), nil
	}
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		actual = string(data)
		return nil
	}

	DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, color, true)

	s.Equal(expected, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_UsesV2_WhenBlueGreen() {
	color := "orange"
	var actual string
	newTarget := fmt.Sprintf("%s-%s", s.target, color)
	var dcContent = fmt.Sprintf(`version: '2'

services:
  %s:
    image: vfarcic/books-ms`,
		s.target,
	)
	expected := fmt.Sprintf(`version: '2'

services:
  %s:
    extends:
      file: %s
      service: %s
    environment:
      - SERVICE_NAME=%s-%s
  %s:
    extends:
      file: %s
      service: %s
  %s:
    extends:
      file: %s
      service: %s`,
		newTarget,
		s.dockerComposePath,
		s.target,
		s.serviceName,
		color,
		s.sideTargets[0],
		s.dockerComposePath,
		s.sideTargets[0],
		s.sideTargets[1],
		s.dockerComposePath,
		s.sideTargets[1],
	)
	util.ReadFile = func(filename string) ([]byte, error) {
		return []byte(dcContent), nil
	}
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		actual = string(data)
		return nil
	}

	DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, color, true)

	s.Equal(expected, actual)
}

func (s DockerComposeTestSuite) Test_CreateFlowFile_ReturnsError_WhenWriteFile() {
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("Some error")
	}

	err := DockerCompose{}.CreateFlowFile(s.dockerComposePath, s.serviceName, s.target, s.sideTargets, s.color, s.blueGreen)

	s.Error(err)
}

// RemoveFlow

func (s DockerComposeTestSuite) Test_RemoveFlow_RemovesTheFile() {
	var actual string
	util.RemoveFile = func(name string) error {
		actual = name
		return nil
	}

	DockerCompose{}.RemoveFlow()

	s.Equal(dockerComposeFlowPath, actual)
}

func (s DockerComposeTestSuite) Test_RemoveFlow_ReturnsError() {
	util.RemoveFile = func(name string) error {
		return fmt.Errorf("Some error")
	}

	err := DockerCompose{}.RemoveFlow()

	s.Error(err)
}

// PullTargets

func (s DockerComposeTestSuite) Test_PullTargets_ReturnsNil_WhenTargetsAreEmpty() {
	actual := DockerCompose{}.PullTargets(s.host, s.certPath, s.project, []string{})

	s.Nil(actual)
}

func (s DockerComposeTestSuite) Test_PullTargets() {
	s.testCmd(DockerCompose{}.PullTargets, "pull", s.target)
}

func (s DockerComposeTestSuite) Test_PullTargets_ReturnsError_WhenCommandFails() {
	runCmdOrig := util.RunCmd
	defer func() { util.RunCmd = runCmdOrig }()
	util.RunCmd = func(cmd *exec.Cmd) error { return fmt.Errorf("This is an error") }

	actual := DockerCompose{}.PullTargets(s.host, s.certPath, s.project, []string{s.target})
	s.Error(actual)
}

// UpTargets

func (s DockerComposeTestSuite) Test_UpTargets() {
	s.testCmd(DockerCompose{}.UpTargets, "up", "-d", s.target)
}

// ScaleTargets

func (s DockerComposeTestSuite) Test_ScaleTargets_ReturnsNil_WhenTargetIsEmpty() {
	actual := DockerCompose{}.ScaleTargets(s.host, s.certPath, s.project, "", 8)

	s.Nil(actual)
}

func (s DockerComposeTestSuite) Test_ScaleTargets_CreatesTheCommand() {
	var scale = 7
	expected := []string{"docker-compose", "-f", dockerComposeFlowPath, "-p", s.project, "scale", fmt.Sprintf("%s=%d", s.target, scale)}
	actual := s.mockExecCmd()

	DockerCompose{}.ScaleTargets(s.host, s.certPath, s.project, s.target, scale)

	s.Equal(expected, *actual)
}

// RmTargets

func (s DockerComposeTestSuite) Test_RmTargets() {
	s.testCmd(DockerCompose{}.RmTargets, "rm", "-f", s.target)
}

// StopTargets

func (s DockerComposeTestSuite) Test_StopTargets() {
	s.testCmd(DockerCompose{}.StopTargets, "stop", s.target)
}

// Suite

func TestDockerComposeTestSuite(t *testing.T) {
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	runCmdOrig := util.RunCmd
	defer func() { util.RunCmd = runCmdOrig }()
	util.RunCmd = func(cmd *exec.Cmd) error { return nil }
	suite.Run(t, new(DockerComposeTestSuite))
}

// Helper

func (s DockerComposeTestSuite) mockExecCmd() *[]string {
	var actualCommand []string
	util.ExecCmd = func(name string, arg ...string) *exec.Cmd {
		actualCommand = append([]string{name}, arg...)
		cmd := &exec.Cmd{}
		return cmd
	}
	return &actualCommand
}

type testCmdType func(host, certPath, project string, targets []string) error

func (s DockerComposeTestSuite) testCmd(f testCmdType, args ...string) {
	var expected []string
	var actual *[]string

	// Returns nil when targets are empty
	s.Nil(f(s.host, s.certPath, s.project, []string{}))

	// Creates command
	expected = append([]string{"docker-compose", "-f", dockerComposeFlowPath, "-p", s.project}, args...)
	actual = s.mockExecCmd()
	f(s.host, s.certPath, s.project, []string{s.target})
	s.Equal(expected, *actual)

	// Does not add project when empty
	expected = append([]string{"docker-compose", "-f", dockerComposeFlowPath}, args...)
	actual = s.mockExecCmd()
	f(s.host, s.certPath, "", []string{s.target})
	s.Equal(expected, *actual)

	// Adds DOCKER_HOST variable
	f(s.host, s.certPath, s.project, []string{s.target})
	host := s.host
	s.Equal(host, s.host)

	// Does not add DOCKER_HOST variable when empty
	f("", s.certPath, s.project, []string{s.target})
	s.NotEqual(os.Getenv("DOCKER_HOST"), s.host)

	// Adds DOCKER_CERT_PATH variable
	f(s.host, s.certPath, s.project, []string{s.target})
	s.Equal(os.Getenv("DOCKER_CERT_PATH"), s.certPath)

}
