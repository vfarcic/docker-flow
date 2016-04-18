package main

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"fmt"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
	"strconv"
	"path/filepath"
)

// Setup

type OptsTestSuite struct {
	suite.Suite
	dir string
	opts Opts
}

func (s *OptsTestSuite) SetupTest() {
	s.dir = "myProjectDir"
	s.opts = Opts{
		Project: "myProject",
		Target: "myTarget",
		ServiceDiscoveryAddress: "http://1.2.3.4:1234",
		ServiceName: "myFancyService",
		CurrentColor: "orange",
	}
	serviceDiscovery = getServiceDiscoveryMock(s.opts, "")
	path := fmt.Sprintf("/some/path/%s", s.dir)
	path = filepath.FromSlash(path)
	getWd = func() (string, error) {
		return path, nil
	}
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), nil
	}
}

// ProcessOpts

func (s OptsTestSuite) Test_ProcessOpts_ReturnsNil() {
	actual := ProcessOpts(&s.opts)

	s.Nil(actual)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsProjectToCurrentDir() {
	s.opts.Project = ""

	ProcessOpts(&s.opts)

	s.Equal(s.dir, s.opts.Project)
}

func (s OptsTestSuite) Test_ProcessOpts_DoesNotSetProjectToCurrentDir_WhenProjectIsNotEmpty() {
	expected := s.opts.Project

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.Project)
}

func (s OptsTestSuite) Test_ProcessOpts_ReturnsError_WhenTargetIsEmpty() {
	s.opts.Target = ""

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) Test_ProcessOpts_ReturnsError_WhenServiceDiscoveryAddressIsEmpty() {
	s.opts.ServiceDiscoveryAddress = ""

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) Test_ProcessOpts_ReturnsError_WhenScaleIsNotNumber() {
	s.opts.Scale = "This Is Not A Number"

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsServiceNameToProjectAndTarget() {
	expected := fmt.Sprintf("%s-%s", s.opts.Project, s.opts.Target)
	mockObj := getServiceDiscoveryMock(s.opts, "GetColor")
	mockObj.On("GetColor", mock.Anything, mock.Anything).Return("orange", fmt.Errorf("This is an error"))
	serviceDiscovery = mockObj
	s.opts.ServiceName = ""

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.ServiceName)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsPortTo8080_WhenEmpty() {
	s.opts.ProxyReconfPort = ""
	ProcessOpts(&s.opts)

	s.Equal(strconv.Itoa(ProxyReconfigureDefaultPort), s.opts.ProxyReconfPort)
}

func (s OptsTestSuite) Test_ProcessOpts_DoesNotSetServiceNameWhenNotEmpty() {
	expected := s.opts.ServiceName

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.ServiceName)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsCurrentColorFromServiceDiscovery() {
	expected, _ := serviceDiscovery.GetColor(s.opts.ServiceDiscoveryAddress, s.opts.ServiceName)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentColor)
}

func (s OptsTestSuite) Test_ProcessOpts_ReturnsError_WhenGetColorFails() {
	mockObj := getServiceDiscoveryMock(s.opts, "GetColor")
	mockObj.On("GetColor", mock.Anything, mock.Anything).Return("orange", fmt.Errorf("This is an error"))
	serviceDiscovery = mockObj

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsNextColorFromServiceDiscovery() {
	expected := serviceDiscovery.GetNextColor(s.opts.CurrentColor)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextColor)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsNextTargetToTarget() {
	expected := s.opts.Target

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextTarget)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsCurrentTargetToTarget() {
	expected := s.opts.Target

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentTarget)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsNextTargetToTargetAndNextColor_WhenBlueGreen() {
	s.opts.BlueGreen = true
	expected := fmt.Sprintf("%s-%s", s.opts.Target, serviceDiscovery.GetNextColor(s.opts.CurrentColor))

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextTarget)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsCurrentTargetToTargetAndCurrentColor_WhenBlueGreen() {
	s.opts.BlueGreen = true
	expected := fmt.Sprintf("%s-%s", s.opts.Target, s.opts.CurrentColor)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentTarget)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsHostFromDockerHostEnv_WhenEmpty() {
	expected := "tcp://5.5.5.5:4444"
	s.opts.Host = ""
	os.Setenv("DOCKER_HOST", expected)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.Host)
}

func (s OptsTestSuite) Test_ProcessOpts_DoesNotSetHostFromDockerHostEnv_WhenNotEmpty() {
	expected := "tcp://5.5.5.5:4444"
	s.opts.Host = expected
	os.Setenv("DOCKER_HOST", "myHost")

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.Host)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsCertPathFromDockerCertPathEnv_WhenEmpty() {
	expected := "/path/to/docker/cert"
	s.opts.Host = ""
	os.Setenv("DOCKER_CERT_PATH", expected)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CertPath)
}

func (s OptsTestSuite) Test_ProcessOpts_DoesNotSetCertPathFromDockerCertPathEnv_WhenEmpty() {
	expected := "/path/to/docker/cert"
	s.opts.CertPath = expected
	os.Setenv("DOCKER_CERT_PATH", "/my/cert/path")

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CertPath)
}

func (s OptsTestSuite) Test_ProcessOpts_SetsFlowToDeploy_WhenEmpty() {
	expected := []string{"deploy"}
	s.opts.Flow = []string{}
	ProcessOpts(&s.opts)
	s.Equal(expected, s.opts.Flow)
}

// ParseEnvVars

func (s OptsTestSuite) Test_ParseEnvVars_Strings() {
	data := []struct{
		expected	string
		key 		string
		value		*string
	}{
		{"myHost", 				"FLOW_HOST", 					&s.opts.Host},
		{"myCertPath", 			"FLOW_CERT_PATH", 				&s.opts.CertPath},
		{"myComposePath", 		"FLOW_COMPOSE_PATH", 			&s.opts.ComposePath},
		{"myTarget", 			"FLOW_TARGET", 					&s.opts.Target},
		{"myProject", 			"FLOW_PROJECT", 				&s.opts.Project},
		{"mySDAddress", 		"FLOW_CONSUL_ADDRESS", 			&s.opts.ServiceDiscoveryAddress},
		{"myScale", 			"FLOW_SCALE", 					&s.opts.Scale},
		{"myProxyHost", 		"FLOW_PROXY_HOST", 				&s.opts.ProxyHost},
		{"myProxyDockerHost", 	"FLOW_PROXY_DOCKER_HOST", 		&s.opts.ProxyDockerHost},
		{"myProxyCertPath", 	"FLOW_PROXY_DOCKER_CERT_PATH",	&s.opts.ProxyDockerCertPath},
		{"4357",				"FLOW_PROXY_RECONF_PORT",		&s.opts.ProxyReconfPort},
	}
	for _, d := range data {
		os.Setenv(d.key, d.expected)
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) Test_ParseEnvVars_Bools() {
	data := []struct{
		key 		string
		value		*bool
	}{
		{"FLOW_BLUE_GREEN", 		&s.opts.BlueGreen},
		{"FLOW_PULL_SIDE_TARGETS", 	&s.opts.PullSideTargets},
	}
	for _, d := range data {
		os.Setenv(d.key, "true")
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.True(*d.value)
	}
}

func (s OptsTestSuite) Test_ParseEnvVars_Slices() {
	data := []struct{
		expected	string
		key 		string
		value		*[]string
	}{
		{"myTarget1,myTarget2", "FLOW_SIDE_TARGETS", &s.opts.SideTargets},
		{"deploy,stop-old", "FLOW", &s.opts.Flow},
		{"path1,path2", "FLOW_SERVICE_PATH", &s.opts.ServicePath},
	}
	for _, d := range data {
		os.Setenv(d.key, d.expected)
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.Equal(strings.Split(d.expected, ","), *d.value)
	}
}

func (s OptsTestSuite) Test_ParseEnvVars_DoesNotParseSlices_WhenEmpty() {
	data := []struct{
		key 		string
		value		*[]string
	}{
		{"FLOW_SIDE_TARGETS", &s.opts.SideTargets},
	}
	for _, d := range data {
		s.opts.SideTargets = []string{}
		os.Unsetenv(d.key)
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.Len(*d.value, 0)
	}
}

func (s OptsTestSuite) Test_ParseEnvVars_ReturnsError_WhenFailure() {
	os.Setenv("FLOW_BLUE_GREEN", "This is not a bool")

	actual := ParseEnvVars(&s.opts)

	s.Error(actual)
	os.Unsetenv("FLOW_BLUE_GREEN")
}

// ParseArgs

func (s OptsTestSuite) Test_ParseArgs_LongStrings() {
	data := []struct{
		expected	string
		key 		string
		value		*string
	}{
		{"hostFromArgs", 			"host", 					&s.opts.Host},
		{"certPathFromArgs", 		"cert-path", 				&s.opts.CertPath},
		{"composePathFromArgs",		"compose-path", 			&s.opts.ComposePath},
		{"targetFromArgs", 			"target", 					&s.opts.Target},
		{"projectFromArgs", 		"project", 					&s.opts.Project},
		{"addressFromArgs", 		"consul-address", 			&s.opts.ServiceDiscoveryAddress},
		{"scaleFromArgs", 			"scale", 					&s.opts.Scale},
		{"proxyDomainFromArgs",		"proxy-host", 				&s.opts.ProxyHost},
		{"proxyHostFromArgs", 		"proxy-docker-host", 		&s.opts.ProxyDockerHost},
		{"proxyCertPathFromArgs", 	"proxy-docker-cert-path", 	&s.opts.ProxyDockerCertPath},
		{"1234", 					"proxy-reconf-port", 		&s.opts.ProxyReconfPort},
	}

	for _, d := range data {
		os.Args = []string{"myProgram", fmt.Sprintf("--%s=%s", d.key, d.expected)}
		ParseArgs(&s.opts)
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) Test_ParseArgs_ParsesLongSlices() {
	os.Args = []string{"myProgram"}
	data := []struct{
		expected	[]string
		key 		string
		value		*[]string
	}{
		{[]string{"path1", "path2"}, "service-path", &s.opts.ServicePath},
	}

	for _, d := range data {
		for _, v := range d.expected {
			os.Args = append(os.Args, fmt.Sprintf("--%s", d.key), v)
		}

	}

	ParseArgs(&s.opts)

	for _, d := range data {
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_ShortStrings() {
	data := []struct{
		expected	string
		key 		string
		value		*string
	}{
		{"hostFromArgs", "H", &s.opts.Host},
		{"composePathFromArgs", "f", &s.opts.ComposePath},
		{"targetFromArgs", "t", &s.opts.Target},
		{"projectFromArgs", "p", &s.opts.Project},
		{"addressFromArgs", "c", &s.opts.ServiceDiscoveryAddress},
		{"scaleFromArgs", "s", &s.opts.Scale},
	}

	for _, d := range data {
		os.Args = []string{"myProgram", fmt.Sprintf("-%s=%s", d.key, d.expected)}
		ParseArgs(&s.opts)
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_LongBools() {
	data := []struct{
		key 		string
		value		*bool
	}{
		{"blue-green", &s.opts.BlueGreen},
		{"pull-side-targets", &s.opts.PullSideTargets},
	}

	for _, d := range data {
		os.Args = []string{"myProgram", fmt.Sprintf("--%s", d.key)}
		ParseArgs(&s.opts)
		s.True(*d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_ShortBools() {
	data := []struct{
		key 		string
		value		*bool
	}{
		{"b", &s.opts.BlueGreen},
		{"S", &s.opts.PullSideTargets},
	}

	for _, d := range data {
		os.Args = []string{"myProgram", fmt.Sprintf("-%s", d.key)}
		ParseArgs(&s.opts)
		s.True(*d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_LongSlices() {
	data := []struct{
		expected	[]string
		key 		string
		value		*[]string
	}{
		{[]string{"target1", "target2"}, "side-target", &s.opts.SideTargets},
		{[]string{"deploy", "stop-old"}, "flow", &s.opts.Flow},
	}

	for _, d := range data {
		os.Args = []string{"myProgram"}
		for _, v := range d.expected {
			os.Args = append(os.Args, fmt.Sprintf("--%s=%s", d.key, v))
		}
		ParseArgs(&s.opts)
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_ShortSlices() {
	data := []struct{
		expected	[]string
		key 		string
		value		*[]string
	}{
		{[]string{"target1", "target2"}, "T", &s.opts.SideTargets},
		{[]string{"flow", "stop-old"}, "F", &s.opts.Flow},
	}

	for _, d := range data {
		os.Args = []string{"myProgram"}
		for _, v := range d.expected {
			os.Args = append(os.Args, fmt.Sprintf("-%s=%s", d.key, v))
		}
		ParseArgs(&s.opts)
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) TestParseArgs_ReturnsError_WhenFailure() {
	os.Args = []string{"myProgram", "--this-flag-does-not-exist=something"}

	actual := ParseArgs(&s.opts)

	s.Error(actual)
}

// ParseYml

func (s OptsTestSuite) Test_ParseYml_ReturnsNil() {
	actual := ParseYml(&s.opts)

	s.Nil(actual)
}

func (s OptsTestSuite) Test_ParseYml_ReturnsNil_WhenReadFileFails() {
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("This is an error")
	}

	actual := ParseYml(&s.opts)

	s.Nil(actual)
}

func (s OptsTestSuite) Test_ParseYml_ReturnsError_WhenUnmarshalFails() {
	readFile = func(fileName string) ([]byte, error) {
		return []byte("This is not a proper YML"), nil
	}

	actual := ParseYml(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) Test_ParseYml_SetsOpts() {
	host := "hostFromYml"
	certPath := "certPathFromYml"
	composePath := "composePathFromYml"
	target := "targetFromYml"
	sideTarget1 := "sideTarget1FromYml"
	sideTarget2 := "sideTarget2FromYml"
	project := "projectFromYml"
	consulAddress := "consulAddressFromYml"
	scale := "scaleFromYml"
	flow1 := "deploy"
	flow2 := "stop-old"
	path1 := "path1"
	path2 := "path2"
	proxyHost := "proxyHostFromYml"
	proxyDockerHost := "proxyDomainFromYml"
	proxyDockerCertPath := "proxyCertPathFromYml"
	proxyReconfPort := "1245"
	yml := fmt.Sprintf(`
host: %s
cert_path: %s
compose_path: %s
blue_green: true
target: %s
side_targets:
  - %s
  - %s
skip_pull_target: true
pull_side_targets: true
project: %s
consul_address: %s
scale: %s
proxy_host: %s
proxy_docker_host: %s
proxy_docker_cert_path: %s
proxy_reconf_port: %s
flow:
  - %s
  - %s
service_path:
  - %s
  - %s
`,
		host, certPath, composePath, target, sideTarget1, sideTarget2,
		project, consulAddress, scale, proxyHost, proxyDockerHost,
		proxyDockerCertPath, proxyReconfPort, flow1, flow2, path1,
		path2,
	)
	readFile = func(fileName string) ([]byte, error) {
		return []byte(yml), nil
	}

	ParseYml(&s.opts)

	s.Equal(host, s.opts.Host)
	s.Equal(composePath, s.opts.ComposePath)
	s.True(s.opts.BlueGreen)
	s.Equal(target, s.opts.Target)
	s.Equal([]string{sideTarget1, sideTarget2}, s.opts.SideTargets)
	s.True(s.opts.PullSideTargets)
	s.Equal(project, s.opts.Project)
	s.Equal(consulAddress, s.opts.ServiceDiscoveryAddress)
	s.Equal(scale, s.opts.Scale)
	s.Equal(proxyHost, s.opts.ProxyHost)
	s.Equal(proxyDockerHost, s.opts.ProxyDockerHost)
	s.Equal(proxyDockerCertPath, s.opts.ProxyDockerCertPath)
	s.Equal(proxyReconfPort, s.opts.ProxyReconfPort)
	s.Equal([]string{flow1, flow2}, s.opts.Flow)
	s.Equal([]string{path1, path2}, s.opts.ServicePath)
}

// GetOpts

func (s OptsTestSuite) TestGetOpts_SetsComposePath() {
	opts, _ := GetOpts()

	s.Equal(dockerComposePath, opts.ComposePath)
}

func (s OptsTestSuite) TestGetOpts_InvokesParseYml() {
	called := false
	processOpts = func (*Opts) error {
		return nil
	}
	parseYml = func (*Opts) error {
		called = true
		return nil
	}

	_, err := GetOpts()

	s.Nil(err)
	s.True(called)
}

func (s OptsTestSuite) Test_GetOpts_ReturnsError_WhenParseYmlFails() {
	restore := parseYml
	processOpts = func (*Opts) error {
		return nil
	}
	parseYml = func (*Opts) error {
		return fmt.Errorf("This is an error from ParseYml")
	}

	_, actual := GetOpts()

	s.Error(actual)
	parseYml = restore
}

func (s OptsTestSuite) Test_GetOpts_InvokesParseEnvVars() {
	called := false
	processOpts = func (*Opts) error {
		return nil
	}
	parseEnvVars = func (*Opts) error {
		called = true
		return nil
	}

	_, err := GetOpts()

	s.Nil(err)
	s.True(called)
}

func (s OptsTestSuite) Test_GetOpts_ReturnsError_WhenParseEnvVarsFails() {
	restore := parseEnvVars
	processOpts = func (*Opts) error {
		return nil
	}
	parseEnvVars = func (*Opts) error {
		return fmt.Errorf("This is an error from ParseEnvVars")
	}

	_, actual := GetOpts()

	s.Error(actual)
	parseEnvVars = restore
}

func (s OptsTestSuite) Test_GetOpts_InvokesParseArgs() {
	called := false
	processOpts = func (*Opts) error {
		return nil
	}
	parseArgs = func (*Opts) error {
		called = true
		return nil
	}

	_, err := GetOpts()

	s.Nil(err)
	s.True(called)
}

func (s OptsTestSuite) Test_GetOpts_ReturnsError_WhenParseArgsFails() {
	restore := parseArgs
	processOpts = func (*Opts) error {
		return nil
	}
	parseArgs = func (*Opts) error {
		return fmt.Errorf("This is an error from ParseArgs")
	}

	_, actual := GetOpts()

	s.Error(actual)
	parseArgs = restore
}

func (s OptsTestSuite) Test_GetOpts_InvokesProcessOpts() {
	called := false
	processOpts = func (*Opts) error {
		called = true
		return nil
	}

	_, err := GetOpts()

	s.Nil(err)
	s.True(called)
}

func (s OptsTestSuite) Test_GetOpts_ReturnsError_WhenProcessOptsFails() {
	restore := processOpts
	processOpts = func (*Opts) error {
		return fmt.Errorf("This is an error from ProcessOpts")
	}

	_, actual := GetOpts()

	s.Error(actual)
	processOpts = restore
}

// Suite

func TestOptsTestSuite(t *testing.T) {
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	suite.Run(t, new(OptsTestSuite))
}

