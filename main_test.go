package main

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type MainTestSuite struct {
	suite.Suite
	opts Opts
	dc   DockerComposable
}

func (s *MainTestSuite) SetupTest() {
	s.opts = Opts{
		ComposePath:   []string{"myComposePath"},
		Target:        "myTarget",
		NextColor:     "orange",
		CurrentColor:  "pink",
		NextTarget:    "myNextTarget",
		CurrentTarget: "myCurrentTarget",
		BlueGreen:     true,
		Flow:          []string{"deploy", "scale"},
		ServiceDiscoveryAddress: "myServiceDiscoveryAddress",
		ServiceName:             "myServiceName",
	}
	GetOpts = func() (Opts, error) {
		return s.opts, nil
	}
	s.dc = getDockerComposeMock(s.opts, "")
	dockerCompose = s.dc
	flow = getFlowMock("")
	haProxy = getProxyMock("")
	serviceDiscovery = getServiceDiscoveryMock(s.opts, "")
	logFatal = func(v ...interface{}) {}
	logPrintln = func(v ...interface{}) {}
	deployed = false
}

// main

func (s MainTestSuite) Test_Main_Exits_WhenGetOptsFails() {
	GetOpts = func() (Opts, error) {
		return s.opts, fmt.Errorf("This is an error")
	}
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

// main > deploy

func (s MainTestSuite) Test_Main_InvokesFlowDeploy_WhenDeploy() {
	mockObj := getFlowMock("")
	flow = mockObj

	main()

	mockObj.AssertCalled(
		s.T(),
		"Deploy",
		s.opts,
		s.dc,
	)
}

func (s MainTestSuite) Test_Main_LogsError_WhenDeployAndFlowDeployFails() {
	mockObj := getFlowMock("Deploy")
	mockObj.On(
		"Deploy",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	flow = mockObj
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

func (s MainTestSuite) Test_Main_InvokesServiceDiscoveryPutColor_WhenDeploy() {
	mockObj := getServiceDiscoveryMock(s.opts, "")
	serviceDiscovery = mockObj

	main()

	mockObj.AssertCalled(
		s.T(),
		"PutColor",
		s.opts.ServiceDiscoveryAddress,
		s.opts.ServiceName,
		s.opts.NextColor,
	)
}

func (s MainTestSuite) Test_Main_LogsFatal_WhenDeployAndServiceDiscoveryPutColorFails() {
	mockObj := getServiceDiscoveryMock(s.opts, "PutColor")
	mockObj.On(
		"PutColor",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return("", fmt.Errorf("This is an error"))
	serviceDiscovery = mockObj
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

// main > scale

func (s MainTestSuite) Test_Main_InvokesFlowScale_WhenScaleAndNotDeploy() {
	mockObj := getFlowMock("")
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"scale"}
		return s.opts, nil
	}
	flow = mockObj

	main()

	mockObj.AssertCalled(
		s.T(),
		"Scale",
		s.opts,
		s.dc,
		s.opts.CurrentTarget,
		true,
	)
}

func (s MainTestSuite) Test_Main_LogsFatal_WhenScaleAndNotDeployAndScaleFails() {
	mockObj := getFlowMock("Scale")
	mockObj.On(
		"Scale",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"scale"}
		return s.opts, nil
	}
	flow = mockObj
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

// main > stop-old

func (s MainTestSuite) Test_Main_InvokesDockerComposeCreateFlowFileWithCurrentColor_WhenStopOldAndDeployed() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}
	deployed = true

	main()

	mockObj.AssertCalled(
		s.T(),
		"CreateFlowFile",
		s.opts.ComposePath,
		s.opts.ServiceName,
		s.opts.Target,
		s.opts.SideTargets,
		s.opts.CurrentColor,
		s.opts.BlueGreen,
	)
}

func (s MainTestSuite) Test_Main_InvokesDockerComposeCreateFlowFileWithNextColor_WhenStopOldAndNotDeployed() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}

	main()

	mockObj.AssertCalled(
		s.T(),
		"CreateFlowFile",
		s.opts.ComposePath,
		s.opts.ServiceName,
		s.opts.Target,
		s.opts.SideTargets,
		s.opts.NextColor,
		s.opts.BlueGreen,
	)
}

func (s MainTestSuite) Test_Main_InvokesLogFatal_WhenStopOldAndDockerComposeCreateFlowFileFails() {
	mockObj := getDockerComposeMock(s.opts, "CreateFlowFile")
	mockObj.On(
		"CreateFlowFile",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

func (s MainTestSuite) Test_Main_InvokesDockerComposeStopTargetWithCurrentTarget_WhenStopOldAndDeployed() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}
	deployed = true

	main()

	mockObj.AssertCalled(
		s.T(),
		"StopTargets",
		s.opts.Host,
		s.opts.CertPath,
		s.opts.Project,
		[]string{s.opts.CurrentTarget},
	)
}

func (s MainTestSuite) Test_Main_InvokesDockerComposeStopTargetWithNextColor_WhenStopOldAndNotDeployed() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}

	main()

	mockObj.AssertCalled(
		s.T(),
		"StopTargets",
		s.opts.Host,
		s.opts.CertPath,
		s.opts.Project,
		[]string{s.opts.NextTarget},
	)
}

func (s MainTestSuite) Test_Main_InvokesLogFatal_WhenStopOldAndDockerComposeStopTargetsFails() {
	mockObj := getDockerComposeMock(s.opts, "StopTargets")
	mockObj.On(
		"StopTargets",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

func (s MainTestSuite) Test_Main_DoesNotRunStopOld_WhenStopOldAndNotBlueGreen() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		s.opts.BlueGreen = false
		return s.opts, nil
	}

	main()

	mockObj.AssertNotCalled(
		s.T(),
		"StopTargets",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	)
}

func (s MainTestSuite) Test_Main_InvokesDockerComposeRemoveFlow_WhenStopOld() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}

	main()

	mockObj.AssertCalled(s.T(), "RemoveFlow")
}

func (s MainTestSuite) Test_Main_InvokesLogFatal_WhenStopOldAndDockerComposeRemoveFlowFails() {
	mockObj := getDockerComposeMock(s.opts, "RemoveFlow")
	mockObj.On("RemoveFlow").Return(fmt.Errorf("This is an error"))
	dockerCompose = mockObj
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"stop-old"}
		return s.opts, nil
	}

	main()

	s.True(actual)
}

// main > proxy

func (s MainTestSuite) Test_Main_InvokesFlawProxy() {
	mockObj := getFlowMock("")
	flow = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"proxy"}
		return s.opts, nil
	}

	main()

	mockObj.AssertCalled(
		s.T(),
		"Proxy",
		s.opts,
		haProxy,
	)
}

func (s MainTestSuite) Test_Main_InvokesLogFatal_WhenFlawProxyFails() {
	mockObj := getFlowMock("Proxy")
	mockObj.On("Proxy", mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	flow = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"proxy"}
		return s.opts, nil
	}
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

// Suite

func TestMainTestSuite(t *testing.T) {
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	getOptsOrig := GetOpts
	defer func() {
		GetOpts = getOptsOrig
	}()
	suite.Run(t, new(MainTestSuite))
}
