package main

import (
	"testing"
	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/mock"
	"fmt"
)

type MainTestSuite struct {
	suite.Suite
	opts		Opts
	dc			DockerComposable
}

func (s *MainTestSuite) SetupTest() {
	s.opts = Opts{
		ComposePath: "myComposePath",
		Target: "myTarget",
		NextColor: "orange",
		CurrentColor: "pink",
		NextTarget: "myNextTarget",
		CurrentTarget: "myCurrentTarget",
		BlueGreen: true,
		Flow: []string{"deploy", "scale"},
		ServiceDiscoveryAddress: "myServiceDiscoveryAddress",
		ServiceName: "myServiceName",
	}
	GetOpts = func() (Opts, error) {
		return s.opts, nil
	}
	s.dc = getDockerComposeMock(s.opts, "")
	dockerCompose = s.dc
	flow = getFlowMock("")
	serviceDiscovery = getServiceDiscoveryMock(s.opts, "")
	logFatal = func(v ...interface{}) { }
	osExit = func(code int) { }
	logPrintln = func(v ...interface{}) { }
	deployed = false
}

// main

func (s MainTestSuite) Test_Main_Exits_WhenGetOptsFails() {
	GetOpts = func() (Opts, error) {
		return s.opts, fmt.Errorf("This is an error")
	}
	actual := 0
	osExit = func(code int) {
		actual = code
	}

	main()

	s.Equal(1, actual)
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

func (s MainTestSuite) Test_Main_InvokesDockerComposeCreateFlowFile_WhenScaleAndNotDeploy() {
	mockObj := getDockerComposeMock(s.opts, "")
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"scale"}
		return s.opts, nil
	}
	dockerCompose = mockObj

	main()

	mockObj.AssertCalled(
		s.T(),
		"CreateFlowFile",
		s.opts.ComposePath,
		dockerComposeFlowPath,
		s.opts.Target,
		s.opts.CurrentColor,
		s.opts.BlueGreen,
	)
}

func (s MainTestSuite) Test_Main_DoesNotInvokeDockerComposeCreateFlowFile_WhenDeployAndScale() {
	mockObj := getDockerComposeMock(s.opts, "")
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"deploy", "scale"}
		return s.opts, nil
	}
	dockerCompose = mockObj

	main()

	mockObj.AssertNotCalled(
		s.T(),
		"CreateFlowFile",
		s.opts.ComposePath,
		dockerComposeFlowPath,
		s.opts.Target,
		s.opts.CurrentColor,
		s.opts.BlueGreen,
	)
}

func (s MainTestSuite) Test_Main_LogsFatal_WhenScaleAndCreateFlowFileFails() {
	mockObj := getDockerComposeMock(s.opts, "CreateFlowFile")
	mockObj.On(
		"CreateFlowFile",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	dockerCompose = mockObj
	GetOpts = func() (Opts, error) {
		s.opts.Flow = []string{"scale"}
		s.opts.BlueGreen = false
		return s.opts, nil
	}
	deployed = false
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

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
	)
}

func (s MainTestSuite) Test_Main_LogsFatal_WhenScaleAndNotDeployAndScaleFails() {
	mockObj := getFlowMock("Scale")
	mockObj.On(
		"Scale",
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
		dockerComposeFlowPath,
		s.opts.Target,
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
		dockerComposeFlowPath,
		s.opts.Target,
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

// main > cleanup

func (s MainTestSuite) Test_Main_InvokesDockerComposeRemoveFlow() {
	mockObj := getDockerComposeMock(s.opts, "")
	dockerCompose = mockObj

	main()

	mockObj.AssertCalled(s.T(), "RemoveFlow")
}

func (s MainTestSuite) Test_Main_InvokesLogFatal_WhenDockerComposeRemoveFlowFails() {
	mockObj := getDockerComposeMock(s.opts, "RemoveFlow")
	mockObj.On("RemoveFlow").Return(fmt.Errorf("This is an error"))
	dockerCompose = mockObj
	actual := false
	logFatal = func(v ...interface{}) {
		actual = true
	}

	main()

	s.True(actual)
}

// Suite

func TestMainTestSuite(t *testing.T) {
	getOptsOrig := GetOpts
	defer func() {
		GetOpts = getOptsOrig
	}()
	suite.Run(t, new(MainTestSuite))
}
