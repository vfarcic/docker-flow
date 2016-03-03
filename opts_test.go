package dockerflow

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"fmt"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
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
	s.opts.ServiceDiscovery = getServiceDiscoveryMock(s.opts, "")
	path := fmt.Sprintf("/some/path/%s", s.dir)
	getWd = func() (string, error) {
		return path, nil
	}
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), nil
	}
}

func TestOptsTestSuite(t *testing.T) {
	suite.Run(t, new(OptsTestSuite))
}

// processOpts

func (s OptsTestSuite) TestProcessOpts_ReturnsNil() {
	actual := ProcessOpts(&s.opts)

	s.Nil(actual)
}

func (s OptsTestSuite) TestProcessOpts_SetsServiceDiscoveryToConsul() {
	s.opts.Target = "" // So that it fails instead of running the whole method
	s.opts.ServiceDiscovery = nil

	ProcessOpts(&s.opts)

	s.IsType(Consul{}, s.opts.ServiceDiscovery)
}

func (s OptsTestSuite) TestProcessOpts_DoesNotSetServiceDiscoveryToConsul_WhenNotEmpty() {
	ProcessOpts(&s.opts)

	s.IsType(getServiceDiscoveryMock(s.opts, ""), s.opts.ServiceDiscovery)
}

func (s OptsTestSuite) TestProcessOpts_SetsProjectToCurrentDir() {
	s.opts.Project = ""

	ProcessOpts(&s.opts)

	s.Equal(s.dir, s.opts.Project)
}

func (s OptsTestSuite) TestProcessOpts_DoesNotSetProjectToCurrentDir_WhenProjectIsNotEmpty() {
	expected := s.opts.Project

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.Project)
}

func (s OptsTestSuite) TestProcessOpts_ReturnsError_WhenTargetIsEmpty() {
	s.opts.Target = ""

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestProcessOpts_ReturnsError_WhenServiceDiscoveryAddressIsEmpty() {
	s.opts.ServiceDiscoveryAddress = ""

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestProcessOpts_ReturnsError_WhenScaleIsNotNumber() {
	s.opts.Scale = "This Is Not A Number"

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestProcessOpts_SetsServiceNameToProjectAndTarget() {
	expected := fmt.Sprintf("%s-%s", s.opts.Project, s.opts.Target)
	mockObj := getServiceDiscoveryMock(s.opts, "GetColor")
	mockObj.On("GetColor", mock.Anything, mock.Anything).Return("orange", fmt.Errorf("This is an error"))
	s.opts.ServiceDiscovery = mockObj
	s.opts.ServiceName = ""

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.ServiceName)
}

func (s OptsTestSuite) TestProcessOpts_DoesNotSetServiceNameWhenNotEmpty() {
	expected := s.opts.ServiceName

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.ServiceName)
}

func (s OptsTestSuite) TestProcessOpts_SetsCurrentColorFromServiceDiscovery() {
	expected, _ := s.opts.ServiceDiscovery.GetColor(s.opts.ServiceDiscoveryAddress, s.opts.ServiceName)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentColor)
}

func (s OptsTestSuite) TestProcessOpts_ReturnsError_WhenGetColorFails() {
	mockObj := getServiceDiscoveryMock(s.opts, "GetColor")
	mockObj.On("GetColor", mock.Anything, mock.Anything).Return("orange", fmt.Errorf("This is an error"))
	s.opts.ServiceDiscovery = mockObj

	actual := ProcessOpts(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestProcessOpts_SetsNextColorFromServiceDiscovery() {
	expected := s.opts.ServiceDiscovery.GetNextColor(s.opts.CurrentColor)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextColor)
}

func (s OptsTestSuite) TestProcessOpts_SetsNextTargetToTarget() {
	expected := s.opts.Target

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextTarget)
}

func (s OptsTestSuite) TestProcessOpts_SetsCurrentTargetToTarget() {
	expected := s.opts.Target

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentTarget)
}

func (s OptsTestSuite) TestProcessOpts_SetsNextTargetToTargetAndNextColor_WhenBlueGreen() {
	s.opts.BlueGreen = true
	expected := fmt.Sprintf("%s-%s", s.opts.Target, s.opts.ServiceDiscovery.GetNextColor(s.opts.CurrentColor))

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.NextTarget)
}

func (s OptsTestSuite) TestProcessOpts_SetsCurrentTargetToTargetAndCurrentColor_WhenBlueGreen() {
	s.opts.BlueGreen = true
	expected := fmt.Sprintf("%s-%s", s.opts.Target, s.opts.CurrentColor)

	ProcessOpts(&s.opts)

	s.Equal(expected, s.opts.CurrentTarget)
}

// ParseEnvVars

func (s OptsTestSuite) TestParseEnvVars_Strings() {
	data := []struct{
		expected	string
		key 		string
		value		*string
	}{
		{"myHost", 				"FLOW_HOST", 			&s.opts.Host},
		{"myComposePath", 		"FLOW_COMPOSE_PATH", 	&s.opts.ComposePath},
		{"myTarget", 			"FLOW_TARGET", 			&s.opts.Target},
		{"myProject", 			"FLOW_PROJECT", 		&s.opts.Project},
		{"mySDAddress", 		"FLOW_CONSUL_ADDRESS", 	&s.opts.ServiceDiscoveryAddress},
		{"myScale", 			"FLOW_SCALE", 			&s.opts.Scale},
	}
	for _, d := range data {
		os.Setenv(d.key, d.expected)
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.Equal(d.expected, *d.value)
	}
}

func (s OptsTestSuite) TestParseEnvVars_Bools() {
	data := []struct{
		key 		string
		value		*bool
	}{
		{"FLOW_WEB_SERVER", 		&s.opts.WebServer},
		{"FLOW_BLUE_GREEN", 		&s.opts.BlueGreen},
		{"FLOW_SKIP_PULL_TARGET", 	&s.opts.SkipPullTarget},
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

func (s OptsTestSuite) TestParseEnvVars_Slices() {
	data := []struct{
		expected	string
		key 		string
		value		*[]string
	}{
		{"myTarget1,myTarget2", "FLOW_SIDE_TARGETS", &s.opts.SideTargets},
	}
	for _, d := range data {
		os.Setenv(d.key, d.expected)
	}
	ParseEnvVars(&s.opts)
	for _, d := range data {
		s.Equal(strings.Split(d.expected, ","), *d.value)
	}
}

func (s OptsTestSuite) TestParseEnvVars_ReturnsError_WhenFailure() {
	os.Setenv("FLOW_WEB_SERVER", "This is not a bool")

	actual := ParseEnvVars(&s.opts)

	s.Error(actual)
	os.Unsetenv("FLOW_WEB_SERVER")
}

// ParseArgs

func (s OptsTestSuite) TestParseArgs_ReturnsError_WhenFailure() {
	os.Args = []string{"myProgram", "--this-flag-does-not-exist=something"}

	actual := ParseArgs(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestParseArgs_LongStrings() {
	data := []struct{
		expected	string
		key 		string
		value		*string
	}{
		{"hostFromArgs", "host", &s.opts.Host},
		{"composePathFromArgs", "compose-path", &s.opts.ComposePath},
		{"targetFromArgs", "target", &s.opts.Target},
		{"projectFromArgs", "project", &s.opts.Project},
		{"addressFromArgs", "consul-address", &s.opts.ServiceDiscoveryAddress},
		{"scaleFromArgs", "scale", &s.opts.Scale},
	}

	for _, d := range data {
		os.Args = []string{"myProgram", fmt.Sprintf("--%s=%s", d.key, d.expected)}
		ParseArgs(&s.opts)
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
		{"web-server", &s.opts.WebServer},
		{"blue-green", &s.opts.BlueGreen},
		{"skip-pull-targets", &s.opts.SkipPullTarget},
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
		{"w", &s.opts.WebServer},
		{"b", &s.opts.BlueGreen},
		{"P", &s.opts.SkipPullTarget},
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

// ParseYml

func (s OptsTestSuite) TestParseYml_ReturnsNil() {
	actual := ParseYml(&s.opts)

	s.Nil(actual)
}

func (s OptsTestSuite) TestParseYml_ReturnsError_WhenReadFileFails() {
	readFile = func(fileName string) ([]byte, error) {
		return []byte(""), fmt.Errorf("This is an error")
	}

	actual := ParseYml(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestParseYml_ReturnsError_WhenUnmarshalFails() {
	readFile = func(fileName string) ([]byte, error) {
		return []byte("This is not a proper YML"), nil
	}

	actual := ParseYml(&s.opts)

	s.Error(actual)
}

func (s OptsTestSuite) TestParseYml_SetsOpts() {
	host := "hostFromYml"
	composePath := "composePathFromYml"
	target := "targetFromYml"
	sideTarget1 := "sideTarget1FromYml"
	sideTarget2 := "sideTarget2FromYml"
	project := "projectFromYml"
	consulAddress := "consulAddressFromYml"
	scale := "scaleFromYml"
	yml := fmt.Sprintf(`
host: %s
compose_path: %s
web_server: true
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
`, host, composePath, target, sideTarget1, sideTarget2, project, consulAddress, scale)
	readFile = func(fileName string) ([]byte, error) {
		return []byte(yml), nil
	}

	ParseYml(&s.opts)

	s.Equal(host, s.opts.Host)
	s.Equal(composePath, s.opts.ComposePath)
	s.True(s.opts.WebServer)
	s.True(s.opts.BlueGreen)
	s.Equal(target, s.opts.Target)
	s.Equal([]string{sideTarget1, sideTarget2}, s.opts.SideTargets)
	s.True(s.opts.SkipPullTarget)
	s.True(s.opts.PullSideTargets)
	s.Equal(project, s.opts.Project)
	s.Equal(consulAddress, s.opts.ServiceDiscoveryAddress)
	s.Equal(scale, s.opts.Scale)
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

func (s OptsTestSuite) TestGetOpts_ReturnsError_WhenParseYmlFails() {
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

func (s OptsTestSuite) TestGetOpts_InvokesParseEnvVars() {
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

func (s OptsTestSuite) TestGetOpts_ReturnsError_WhenParseEnvVarsFails() {
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

func (s OptsTestSuite) TestGetOpts_InvokesParseArgs() {
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

func (s OptsTestSuite) TestGetOpts_ReturnsError_WhenParseArgsFails() {
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

func (s OptsTestSuite) TestGetOpts_InvokesProcessOpts() {
	called := false
	processOpts = func (*Opts) error {
		called = true
		return nil
	}

	_, err := GetOpts()

	s.Nil(err)
	s.True(called)
}

func (s OptsTestSuite) TestGetOpts_ReturnsError_WhenProcessOptsFails() {
	restore := processOpts
	processOpts = func (*Opts) error {
		return fmt.Errorf("This is an error from ProcessOpts")
	}

	_, actual := GetOpts()

	s.Error(actual)
	processOpts = restore
}
