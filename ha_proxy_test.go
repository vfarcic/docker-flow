package main

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type HaProxyTestSuite struct {
	suite.Suite
	ScAddress     string
	Host          string
	CertPath      string
	ExitedMessage string
	ProxyHost     string
	Project       string
	Color         string
	ServicePath   []string
	ReconfPort    string
	Server        *httptest.Server
}

func (s *HaProxyTestSuite) SetupTest() {
	s.ScAddress = "1.2.3.4:1234"
	s.Host = "tcp://my-docker-proxy-host"
	s.ProxyHost = "http://my-docker-proxy-host.com"
	s.Project = "myProject"
	s.Color = "purpurina"
	s.ServicePath = []string{"/path/to/my/service", "/path/to/my/other/service"}
	s.ExitedMessage = "Exited (2) 15 seconds ago"
	s.ReconfPort = "5362"
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reconfigureUrl := fmt.Sprintf(
			"/v1/docker-flow-proxy/reconfigure?serviceName=%s&servicePath=%s",
			s.Project,
			strings.Join(s.ServicePath, ","),
		)
		actualUrl := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
		if r.Method == "GET" {
			if actualUrl == reconfigureUrl {
				w.WriteHeader(http.StatusOK)
			}
		}
	}))
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	logPrintln = func(v ...interface{}) {}
	sleep = func(d time.Duration) {}
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		return nil, nil
	}
}

// Provision

func (s HaProxyTestSuite) Test_Provision_SetsDockerHost() {
	actual := ""
	SetDockerHost = func(host, certPath string) {
		actual = host
	}

	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Equal(s.Host, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyHostIsEmpty() {
	err := HaProxy{}.Provision("", s.ReconfPort, s.CertPath, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenScAddressIsEmpty() {
	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_RunsDockerFlowProxyContainer() {
	var actual []string
	expected := []string{
		"docker", "run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", s.ScAddress),
		"-p", "80:80", "-p", fmt.Sprintf("%s:8080", s.ReconfPort),
		"vfarcic/docker-flow-proxy",
	}
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenFailure() {
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

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

	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenPsFailure() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyFails() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

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
	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

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

	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.True(start)
	s.False(run)
}

func (s HaProxyTestSuite) Test_Provision_StartsDockerFlowProxyContainer_WhenProxyIsExited() {
	var actual []string
	expected := []string{"docker", "start", "docker-flow-proxy"}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		cmd.Stdout.Write([]byte(s.ExitedMessage))
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

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

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Error(err)
}

// Reconfigure

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenProxyHostIsEmpty() {
	err := HaProxy{}.Reconfigure("", s.ReconfPort, s.Project, s.Color, s.ServicePath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenProjectIsEmpty() {
	err := HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, "", s.Color, s.ServicePath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenServicePathIsEmpty() {
	err := HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, s.Project, s.Color, []string{""}, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenReconfPortIsEmpty() {
	err := HaProxy{}.Reconfigure(s.ProxyHost, "", s.Project, s.Color, s.ServicePath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequest() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&serviceColor=%s&servicePath=%s",
		s.ProxyHost,
		s.ReconfPort,
		s.Project,
		s.Color,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an error")
	}

	HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, s.Project, s.Color, s.ServicePath, "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequest_WithOutColor_WhenNotBlueGreen() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&servicePath=%s",
		s.ProxyHost,
		s.ReconfPort,
		s.Project,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an error")
	}

	HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, s.Project, "", s.ServicePath, "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequestWithPrependedHttp() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&servicePath=%s",
		s.ProxyHost,
		s.ReconfPort,
		s.Project,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an error")
	}

	HaProxy{}.Reconfigure("my-docker-proxy-host.com", s.ReconfPort, s.Project, "", s.ServicePath, "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenRequestFails() {
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		return nil, fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, s.Project, s.Color, s.ServicePath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenResponseCodeIsNot2xx() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	err := HaProxy{}.Reconfigure(server.URL, "", s.Project, s.Color, s.ServicePath, "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReadsConsulTemplatePath() {
	consulTemplatePath := "/path/to/consul/template"
	actual := ""
	readConsulTemplate = func(fileName string) ([]byte, error) {
		actual = fileName
		return []byte(""), nil
	}

	HaProxy{}.Reconfigure(s.Server.URL, s.ReconfPort, s.Project, s.Color, s.ServicePath, consulTemplatePath)

	s.Equal(consulTemplatePath, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenReadingConsulTemplatePathFails() {
	consulTemplatePath := "/path/to/consul/template"
	readConsulTemplateOrig := readConsulTemplate
	defer func() { readConsulTemplate = readConsulTemplateOrig }()
	readConsulTemplate = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("This is an error")
	}

	err := HaProxy{}.Reconfigure(s.Server.URL, "", s.Project, s.Color, s.ServicePath, consulTemplatePath)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_TransfersConsulTemplateToTheProxy() {
	consulTemplate := `This is a 'consul' template`
	consulTemplatePath := "/path/to/consul/template"
	var actual []string
	command := fmt.Sprintf("echo '%s'", `This is a \'consul\' template`)
	expected := []string{"docker", "exec", "-it", "docker-flow-proxy", command}
	runHaProxyExecCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}
	readConsulTemplate = func(fileName string) ([]byte, error) {
		return []byte(consulTemplate), nil
	}

	HaProxy{}.Reconfigure(s.Server.URL, s.ReconfPort, s.Project, s.Color, s.ServicePath, consulTemplatePath)

	s.Equal(expected, actual)
}

//func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequestWithConsulTemplatePath_WhenSpecified() {
//	actual := ""
//	expected := fmt.Sprintf(
//		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&consulTemplatePath=%s",
//		s.ProxyHost,
//		s.ReconfPort,
//		s.Project,
//		s.ConsulTemplatePath,
//	)
//	httpGetOrig := httpGet
//	defer func() { httpGet = httpGetOrig }()
//	httpGet = func(url string) (resp *http.Response, err error) {
//		actual = url
//		return nil, fmt.Errorf("This is an error")
//	}
//
//	HaProxy{}.Reconfigure(s.ProxyHost, s.ReconfPort, s.Project, s.Color, s.ServicePath)
//
//	s.Equal(expected, actual)
//}

// Suite

func TestHaProxyTestSuite(t *testing.T) {
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	suite.Run(t, new(HaProxyTestSuite))
}
