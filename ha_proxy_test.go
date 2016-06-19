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
	"./util"
)

type HaProxyTestSuite struct {
	suite.Suite
	ScAddress      string
	CertPath       string
	ExitedMessage  string
	Host           string
	ServiceName    string
	Color          string
	ServicePath    []string
	ReconfPort     string
	DockerHost     string
	DockerCertPath string
	Server         *httptest.Server
}

func (s *HaProxyTestSuite) SetupTest() {
	s.ScAddress = "1.2.3.4:1234"
	s.ServiceName = "my-service"
	s.Color = "purpurina"
	s.ServicePath = []string{"/path/to/my/service", "/path/to/my/other/service"}
	s.ExitedMessage = "Exited (2) 15 seconds ago"
	s.ReconfPort = "5362"
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reconfigureUrl := fmt.Sprintf(
			"/v1/docker-flow-proxy/reconfigure",
			s.ServiceName,
			strings.Join(s.ServicePath, ","),
		)
		actualUrl := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
		if r.Method == "GET" {
			if strings.HasPrefix(actualUrl, reconfigureUrl) {
				w.WriteHeader(http.StatusOK)
			}
		}
	}))
	s.DockerHost = "tcp://my-docker-proxy-host"
	s.DockerCertPath = "/path/to/pem"
	s.Host = "http://my-docker-proxy-host.com"
	runHaProxyRunCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyStartCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		return nil, nil
	}
}

// Provision

func (s HaProxyTestSuite) Test_Provision_SetsDockerHost() {
	actual := ""
	SetDockerHostOrig := util.SetDockerHost
	defer func() { util.SetDockerHost = SetDockerHostOrig }()
	util.SetDockerHost = func(host, certPath string) {
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
		return fmt.Errorf("This is an docker ps error")
	}

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Provision_ReturnsError_WhenProxyFails() {
	runHaProxyPsCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an docker ps error")
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
		return fmt.Errorf("This is an docker start error")
	}

	err := HaProxy{}.Provision(s.Host, s.ReconfPort, s.CertPath, s.ScAddress)

	s.Error(err)
}

// Reconfigure

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenProxyHostIsEmpty() {
	err := HaProxy{}.Reconfigure("", "", "", s.ReconfPort, s.ServiceName, s.Color, s.ServicePath, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenProjectIsEmpty() {
	err := HaProxy{}.Reconfigure("", "", s.Host, s.ReconfPort, "", s.Color, s.ServicePath, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenServicePathAndConsulTemplatePathAreEmpty() {
	err := HaProxy{}.Reconfigure("", "", s.Host, s.ReconfPort, s.ServiceName, s.Color, []string{""}, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenReconfPortIsEmpty() {
	err := HaProxy{}.Reconfigure("", "", s.Host, "", s.ServiceName, s.Color, s.ServicePath, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequest() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&serviceColor=%s&servicePath=%s",
		s.Host,
		s.ReconfPort,
		s.ServiceName,
		s.Color,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an HTTP error")
	}

	HaProxy{}.Reconfigure("", "", s.Host, s.ReconfPort, s.ServiceName, s.Color, s.ServicePath, "", "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequestWithOutColor_WhenNotBlueGreen() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&servicePath=%s",
		s.Host,
		s.ReconfPort,
		s.ServiceName,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an HTTP error")
	}

	HaProxy{}.Reconfigure("", "", s.Host, s.ReconfPort, s.ServiceName, "", s.ServicePath, "", "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequestWithPrependedHttp() {
	actual := ""
	expected := fmt.Sprintf(
		"%s:%s/v1/docker-flow-proxy/reconfigure?serviceName=%s&servicePath=%s",
		s.Host,
		s.ReconfPort,
		s.ServiceName,
		strings.Join(s.ServicePath, ","),
	)
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		actual = url
		return nil, fmt.Errorf("This is an HTTP error")
	}

	HaProxy{}.Reconfigure("", "", "my-docker-proxy-host.com", s.ReconfPort, s.ServiceName, "", s.ServicePath, "", "")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenRequestFails() {
	httpGetOrig := httpGet
	defer func() { httpGet = httpGetOrig }()
	httpGet = func(url string) (resp *http.Response, err error) {
		return nil, fmt.Errorf("This is an HTTP error")
	}

	err := HaProxy{}.Reconfigure("", "", s.Host, s.ReconfPort, s.ServiceName, s.Color, s.ServicePath, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenResponseCodeIsNot2xx() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	err := HaProxy{}.Reconfigure("", "", server.URL, "", s.ServiceName, s.Color, s.ServicePath, "", "")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_SetsDockerHost_WhenConsulTemplatePathIsPresent() {
	os.Unsetenv("DOCKER_HOST")

	err := HaProxy{}.Reconfigure(s.DockerHost, s.DockerCertPath, s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.NoError(err)
	s.Equal(s.DockerHost, os.Getenv("DOCKER_HOST"))
	s.Equal(s.DockerCertPath, os.Getenv("DOCKER_CERT_PATH"))
}

func (s HaProxyTestSuite) Test_Reconfigure_CreatesConsulTemplatesDirectory_WhenConsulTemplatePathIsPresent() {
	var actual []string
	expected := []string{"docker", "exec", "-i", "docker-flow-proxy", "mkdir", "-p", "/consul_templates"}
	runHaProxyExecCmd = func(cmd *exec.Cmd) error {
		actual = cmd.Args
		return nil
	}

	HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenDirectoryCreationFails() {
	runHaProxyExecCmdOrig := runHaProxyExecCmd
	defer func() { runHaProxyExecCmd = runHaProxyExecCmdOrig }()
	runHaProxyExecCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an docker exec error")
	}

	actual := HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Error(actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_CopiesTemplates_WhenConsulTemplatePathIsPresent() {
	fePath := "/path/to/consul/fe/template"
	bePath := "/path/to/consul/be/template"
	var actual [][]string
	feExpected := []string{
		"docker",
		"cp",
		fmt.Sprintf("%s.tmp", fePath),
		fmt.Sprintf("docker-flow-proxy:/consul_templates/%s-fe.tmpl", s.ServiceName),
	}
	beExpected := []string{
		"docker",
		"cp",
		fmt.Sprintf("%s.tmp", bePath),
		fmt.Sprintf("docker-flow-proxy:/consul_templates/%s-be.tmpl", s.ServiceName),
	}
	runHaProxyCpCmdOrig := runHaProxyCpCmd
	defer func() { runHaProxyCpCmd = runHaProxyCpCmdOrig }()
	runHaProxyCpCmd = func(cmd *exec.Cmd) error {
		actual = append(actual, cmd.Args)
		return nil
	}

	HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, fePath, bePath)

	s.Equal(feExpected, actual[0])
	s.Equal(beExpected, actual[1])
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenTemplateCopyFails() {
	runHaProxyCpCmdOrig := runHaProxyCpCmd
	defer func() { runHaProxyCpCmd = runHaProxyCpCmdOrig }()
	runHaProxyCpCmd = func(cmd *exec.Cmd) error {
		return fmt.Errorf("This is an docker cp error")
	}

	actual := HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Error(actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_SendsHttpRequestWithConsulTemplatePath_WhenSpecified() {
	actual := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual = fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
	}))
	expected := fmt.Sprintf(
		"/v1/docker-flow-proxy/reconfigure?serviceName=%s&consulTemplateFePath=/consul_templates/%s-fe.tmpl&consulTemplateBePath=/consul_templates/%s-be.tmpl",
		s.ServiceName,
		s.ServiceName,
		s.ServiceName,
	)

	HaProxy{}.Reconfigure("", "", server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Equal(expected, actual)
}

func (s HaProxyTestSuite) Test_Reconfigure_CreatesTempTemplateFile() {
	fePath := "/path/to/consul/fe/template"
	bePath := "/path/to/consul/be/template"
	var actualFilenames []string
	actualData := ""
	data := "This is a %s template"
	expectedData := fmt.Sprintf(data, s.ServiceName+"-"+s.Color)
	writeFileOrig := util.WriteFile
	defer func() { util.WriteFile = writeFileOrig }()
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		actualFilenames = append(actualFilenames, filename)
		actualData = string(data)
		return nil
	}
	readFileOrig := util.ReadFile
	defer func() { util.ReadFile = readFileOrig }()
	util.ReadFile = func(fileName string) ([]byte, error) {
		return []byte(fmt.Sprintf(data, "SERVICE_NAME")), nil
	}

	HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, fePath, bePath)

	s.Equal(fePath+".tmp", actualFilenames[0])
	s.Equal(bePath+".tmp", actualFilenames[1])
	s.Equal(expectedData, actualData)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenTemplateFileReadFails() {
	readFileOrig := util.ReadFile
	defer func() { util.ReadFile = readFileOrig }()
	util.ReadFile = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("This is an read file error")
	}

	err := HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_ReturnsError_WhenTempTemplateFileCreationFails() {
	writeFileOrig := util.WriteFile
	defer func() { util.WriteFile = writeFileOrig }()
	util.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("This is an write file error")
	}

	err := HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, "/path/to/consul/fe/template", "/path/to/consul/be/template")

	s.Error(err)
}

func (s HaProxyTestSuite) Test_Reconfigure_RemovesTempTemplateFile() {
	fePath := "/path/to/consul/fe/template"
	bePath := "/path/to/consul/be/template"
	expected := []string{
		fmt.Sprintf("%s.tmp", fePath),
		fmt.Sprintf("%s.tmp", bePath),
	}
	var actual []string
	removeFileOrig := util.RemoveFile
	defer func() { util.RemoveFile = removeFileOrig }()
	util.RemoveFile = func(name string) error {
		actual = append(actual, name)
		return nil
	}

	HaProxy{}.Reconfigure("", "", s.Server.URL, "", s.ServiceName, s.Color, s.ServicePath, fePath, bePath)

	s.Equal(expected, actual)
}

// Suite

func TestHaProxyTestSuite(t *testing.T) {
	logPrintln = func(v ...interface{}) {}
	logPrintf = func(format string, v ...interface{}) {}
	util.Sleep = func(d time.Duration) {}
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	runHaProxyExecCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	runHaProxyCpCmd = func(cmd *exec.Cmd) error {
		return nil
	}
	util.WriteFile = func(fileName string, data []byte, perm os.FileMode) error {
		return nil
	}
	util.ReadFile = func(fileName string) ([]byte, error) {
		return []byte(""), nil
	}
	util.RemoveFile = func(name string) error {
		return nil
	}
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	suite.Run(t, new(HaProxyTestSuite))
}
