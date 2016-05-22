package main

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type FlowTestSuite struct {
	suite.Suite
	opts Opts
	dc   DockerComposable
}

func (s *FlowTestSuite) SetupTest() {
	s.opts = Opts{
		ComposePath:   "myComposePath",
		Target:        "myTarget",
		NextColor:     "orange",
		CurrentColor:  "pink",
		NextTarget:    "myNextTarget",
		CurrentTarget: "myCurrentTarget",
		BlueGreen:     true,
		Flow:          []string{"deploy", "scale"},
		ServiceDiscoveryAddress: "myServiceDiscoveryAddress",
		ServiceName:             "myServiceName",
		ProxyHost:               "myProxyHost",
		ProxyDockerHost:         "myProxyDockerHost",
		ProxyDockerCertPath:     "myProxyCertPath",
	}
	GetOptsOrig := GetOpts
	defer func() {
		GetOpts = GetOptsOrig
	}()
	GetOpts = func() (Opts, error) {
		return s.opts, nil
	}
	s.dc = getDockerComposeMock(s.opts, "")
	dockerCompose = s.dc
	flow = getFlowMock("")
	serviceDiscovery = getServiceDiscoveryMock(s.opts, "")
	logFatal = func(v ...interface{}) {}
	logPrintln = func(v ...interface{}) {}
}

// Deploy

func (s FlowTestSuite) Test_DeployReturnsNil() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	s.Nil(actual)
}

// Deploy > CreateFlowFile

func (s FlowTestSuite) Test_Deploy_InvokesDockerComposeCreateFlowFile_WhenDeploy() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Deploy(s.opts, s.dc)

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

func (s MainTestSuite) Test_Deploy_ReturnsError_WhenDeployAndDockerComposeCreateFlowFileFails() {
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
	s.dc = mockObj

	err := Flow{}.Deploy(s.opts, s.dc)

	s.Error(err)
}

// Deploy > PullTargets

func (s FlowTestSuite) Test_DeployInvokesPullTargets() {
	opts := Opts{
		Host:        "myHost",
		Project:     "myProject",
		NextTarget:  "myNextTarget",
		SideTargets: []string{"target1", "target2"},
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	flow := Flow{}

	flow.Deploy(opts, mockObj)

	mockObj.AssertCalled(s.T(), "PullTargets", opts.Host, opts.CertPath, opts.Project, flow.GetPullTargets(opts))
}

func (s FlowTestSuite) Test_DeployReturnsError_WhenPullTargetsFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "PullTargets")
	mockObj.On("PullTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	s.Error(actual)
}

// Deploy > UpTargets

func (s FlowTestSuite) Test_DeployInvokesUpTargets() {
	opts := Opts{
		Host:        "myHost",
		Project:     "myProject",
		SideTargets: []string{"target1", "target2"},
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertCalled(s.T(), "UpTargets", opts.Host, opts.CertPath, opts.Project, append(opts.SideTargets, opts.NextTarget))
}

func (s FlowTestSuite) Test_DeployReturnsError_WhenUpTargetsFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "UpTargets")
	mockObj.On(
		"UpTargets",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	s.Error(actual)
}

// Deploy > RmTargets

func (s FlowTestSuite) Test_DeployInvokesRmTargets() {
	opts := Opts{
		BlueGreen:  true,
		Host:       "myHost",
		Project:    "myProject",
		NextTarget: "myNextTarget",
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertCalled(s.T(), "RmTargets", opts.Host, opts.CertPath, opts.Project, []string{opts.NextTarget})
}

func (s FlowTestSuite) Test_DeployDoesNotInvokeRmTargets_WhenBlueGreenIsFalse() {
	opts := Opts{
		BlueGreen: false,
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertNotCalled(s.T(), "RmTargets", opts.Host, opts.Project, append(opts.SideTargets, opts.NextTarget))
}

func (s FlowTestSuite) Test_DeployReturnsError_WhenRmTargetsFails() {
	opts := Opts{
		BlueGreen: true,
	}
	mockObj := getDockerComposeMock(opts, "RmTargets")
	mockObj.On("RmTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := Flow{}.Deploy(opts, mockObj)
	s.Error(actual)
}

// Deploy > GetScaleCalc

func (s FlowTestSuite) Test_DeployReturnsError_WhenGetScaleCalcFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "GetScaleCalc")
	scMockObj.On("GetScaleCalc", mock.Anything, mock.Anything, mock.Anything).Return(0, fmt.Errorf("This is an error"))
	serviceDiscovery = scMockObj

	actual := Flow{}.Deploy(opts, mockObj)

	s.Error(actual)
}

// Deploy > Scale

func (s FlowTestSuite) Test_DeployDoesNotInvokeScaleTargets() {
	opts := Opts{
		Host:                    "myHost",
		Project:                 "myProject",
		NextTarget:              "myNextTarget",
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName:             "myService",
		Scale:                   "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	flow := Flow{}
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	flow.Deploy(opts, mockObj)

	mockObj.AssertCalled(s.T(), "ScaleTargets", opts.Host, opts.CertPath, opts.Project, opts.NextTarget, scale)
}

func (s FlowTestSuite) Test_DeployReturnsError_WhenScaleTargetsFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "ScaleTargets")
	mockObj.On(
		"ScaleTargets",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	s.Error(actual)
}

func (s FlowTestSuite) Test_DeployInvokesCreateFlowFileOnlyOnce() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertNumberOfCalls(s.T(), "CreateFlowFile", 1)
}

func (s FlowTestSuite) Test_DeployInvokesRemoveFlowFileOnlyOnce() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertNumberOfCalls(s.T(), "CreateFlowFile", 1)
}

// Deploy > PutScale

func (s FlowTestSuite) Test_DeployInvokesPutScale() {
	opts := Opts{
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName:             "myService",
		Scale:                   "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "")
	serviceDiscovery = scMockObj
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	Flow{}.Deploy(opts, mockObj)

	scMockObj.AssertCalled(s.T(), "PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
}

// Deploy > RemoveFlow

func (s FlowTestSuite) Test_Deploy_InvokesDockerComposeRemoveFlow() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Deploy(s.opts, s.dc)

	mockObj.AssertCalled(s.T(), "RemoveFlow")
}

func (s FlowTestSuite) Test_Deploy_ReturnsError_WhenDockerComposeRemoveFlowFails() {
	mockObj := getDockerComposeMock(s.opts, "RemoveFlow")
	mockObj.On("RemoveFlow").Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Deploy(s.opts, s.dc)

	s.Error(err)
}

// Scale > CreateFlowFile

func (s FlowTestSuite) Test_Scale_InvokesDockerComposeCreateFlowFile() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget, true)

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

func (s MainTestSuite) Test_Scale_ReturnsError_WhenDeployAndDockerComposeCreateFlowFileFails() {
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
	s.dc = mockObj

	err := Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget, true)

	s.Error(err)
}

// Scale > GetScaleCalc

func (s FlowTestSuite) Test_ScaleReturnsError_WhenGetScaleCalcFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "GetScaleCalc")
	serviceDiscovery = scMockObj
	scMockObj.On("GetScaleCalc", mock.Anything, mock.Anything, mock.Anything).Return(0, fmt.Errorf("This is an error"))

	actual := Flow{}.Scale(opts, mockObj, "myTarget", true)

	s.Error(actual)
}

// Scale > ScaleTargets

func (s FlowTestSuite) Test_ScaleInvokesScaleTargets() {
	opts := Opts{
		Host:                    "myHost",
		Project:                 "myProject",
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName:             "myService",
		Scale:                   "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	flow := Flow{}
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	target := "myTarget"
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	flow.Scale(opts, mockObj, target, true)

	mockObj.AssertCalled(s.T(), "ScaleTargets", opts.Host, opts.CertPath, opts.Project, target, scale)
}

func (s FlowTestSuite) Test_ScaleReturnsError_WhenScaleTargetsFails() {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "ScaleTargets")
	mockObj.On(
		"ScaleTargets",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Scale(opts, mockObj, "myTarget", true)

	s.Error(actual)
}

// Scale > PutScale

func (s FlowTestSuite) Test_ScaleInvokesPutScale() {
	opts := Opts{
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName:             "myService",
		Scale:                   "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "")
	serviceDiscovery = scMockObj
	scale, _ := scMockObj.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	Flow{}.Scale(opts, mockObj, "myTarget", true)

	scMockObj.AssertCalled(s.T(), "PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
}

// Deploy > RemoveFlow

func (s FlowTestSuite) Test_Scale_InvokesDockerComposeRemoveFlow() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget, true)

	mockObj.AssertCalled(s.T(), "RemoveFlow")
}

func (s FlowTestSuite) Test_Scale_ReturnsError_WhenDockerComposeRemoveFlowFails() {
	mockObj := getDockerComposeMock(s.opts, "RemoveFlow")
	mockObj.On("RemoveFlow").Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget, true)

	s.Error(err)
}

// GetTargets

func (s FlowTestSuite) Test_GetTargetsReturnsAllTargets() {
	opts := Opts{
		NextTarget:      "myNextTarget",
		SideTargets:     []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: true,
	}
	expected := append([]string{opts.NextTarget}, opts.SideTargets...)

	actual := Flow{}.GetPullTargets(opts)

	s.Equal(expected, actual)
}

func (s FlowTestSuite) Test_GetTargetsExcludesSideTargets_WhenNotPullSideTargets() {
	opts := Opts{
		NextTarget:      "myNextTarget",
		SideTargets:     []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: false,
	}
	expected := []string{opts.NextTarget}

	actual := Flow{}.GetPullTargets(opts)

	s.Equal(expected, actual)
}

// Proxy

func (s FlowTestSuite) Test_Proxy_InvokesProvision() {
	mockObj := getProxyMock("")

	Flow{}.Proxy(s.opts, mockObj)

	mockObj.AssertCalled(
		s.T(),
		"Provision",
		s.opts.ProxyDockerHost,
		s.opts.ProxyDockerCertPath,
		s.opts.ServiceDiscoveryAddress,
	)
}

func (s FlowTestSuite) Test_Proxy_ReturnsError_WhenProvisionFails() {
	opts := Opts{}
	mockObj := getProxyMock("Provision")
	mockObj.On("Provision", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := Flow{}.Proxy(opts, mockObj)

	s.Error(actual)
}

func (s FlowTestSuite) Test_Proxy_InvokesReconfigure_WhenDeploy() {
	mockObj := getProxyMock("")
	s.opts.Flow = []string{FLOW_DEPLOY}

	Flow{}.Proxy(s.opts, mockObj)

	mockObj.AssertCalled(
		s.T(),
		"Reconfigure",
		s.opts.ProxyDockerHost,
		s.opts.ProxyDockerCertPath,
		s.opts.ProxyHost,
		s.opts.ProxyReconfPort,
		s.opts.ServiceName,
		s.opts.NextColor,
		s.opts.ServicePath,
		"",
	)
}

func (s FlowTestSuite) Test_Proxy_InvokesReconfigure_WhenScale() {
	mockObj := getProxyMock("")
	s.opts.Flow = []string{FLOW_SCALE}

	Flow{}.Proxy(s.opts, mockObj)

	mockObj.AssertCalled(
		s.T(),
		"Reconfigure",
		s.opts.ProxyDockerHost,
		s.opts.ProxyDockerCertPath,
		s.opts.ProxyHost,
		s.opts.ProxyReconfPort,
		s.opts.ServiceName,
		s.opts.CurrentColor,
		s.opts.ServicePath,
		"",
	)
}

func (s FlowTestSuite) Test_Proxy_ReturnsError_WhenReconfigureFails() {
	s.opts.Flow = []string{FLOW_DEPLOY}
	mockObj := getProxyMock("Reconfigure")
	mockObj.On(
		"Reconfigure",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(fmt.Errorf("This is an error"))

	actual := Flow{}.Proxy(s.opts, mockObj)

	s.Error(actual)
}

// Suite

func TestFlowTestSuite(t *testing.T) {
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	defer func() {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("DOCKER_CERT_PATH", dockerCertPath)
	}()
	suite.Run(t, new(FlowTestSuite))
}

// Mock

type FlowMock struct {
	mock.Mock
}

func (m *FlowMock) Deploy(opts Opts, dc DockerComposable) error {
	args := m.Called(opts, dc)
	return args.Error(0)
}

func (m *FlowMock) GetPullTargets(opts Opts) []string {
	//	args := m.Called(opts)
	return []string{}
}

func (m *FlowMock) Scale(opts Opts, dc DockerComposable, target string, createFlowFile bool) error {
	args := m.Called(opts, dc, target, createFlowFile)
	return args.Error(0)
}

func (m *FlowMock) Proxy(opts Opts, proxy Proxy) error {
	args := m.Called(opts, proxy)
	return args.Error(0)
}

func getFlowMock(skipMethod string) *FlowMock {
	mockObj := new(FlowMock)
	if skipMethod != "Deploy" {
		mockObj.On("Deploy", mock.Anything, mock.Anything).Return(nil)
	}
	if skipMethod != "GetPullTargets" {
		mockObj.On("GetPullTargets", mock.Anything).Return(nil)
	}
	if skipMethod != "Scale" {
		mockObj.On("Scale", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	}
	if skipMethod != "Proxy" {
		mockObj.On("Proxy", mock.Anything, mock.Anything).Return(nil)
	}
	return mockObj
}
