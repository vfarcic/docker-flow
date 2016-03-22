package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"fmt"
	"github.com/stretchr/testify/suite"
)

type FlowTestSuite struct {
	suite.Suite
	opts		Opts
	dc			DockerComposable
}

func (s *FlowTestSuite) SetupTest() {
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
}

// Deploy

func Test_DeployReturnsNil(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	assert.Nil(t, actual)
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
		s.opts.Target,
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
	).Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Deploy(s.opts, s.dc)

	s.Error(err)
}

// Deploy > PullTargets

func Test_DeployInvokesPullTargets(t *testing.T) {
	opts := Opts{
		Host: "myHost",
		Project: "myProject",
		NextTarget: "myNextTarget",
		SideTargets: []string{"target1", "target2"},
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	flow := Flow{}

	flow.Deploy(opts, mockObj)

	mockObj.AssertCalled(t, "PullTargets", opts.Host, opts.Project, flow.GetTargets(opts))
}

func Test_DeployReturnsError_WhenPullTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "PullTargets")
	mockObj.On("PullTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	assert.Error(t, actual)
}

// Deploy > UpTargets

func Test_DeployInvokesUpTargets(t *testing.T) {
	opts := Opts{
		Host: "myHost",
		Project: "myProject",
		SideTargets: []string{"target1", "target2"},
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertCalled(t, "UpTargets", opts.Host, opts.Project, opts.SideTargets)
}

func Test_DeployReturnsError_WhenUpTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "UpTargets")
	mockObj.On("UpTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	assert.Error(t, actual)
}

// Deploy > RmTargets

func Test_DeployInvokesRmTargets(t *testing.T) {
	opts := Opts{
		BlueGreen: true,
		Host: "myHost",
		Project: "myProject",
		NextTarget: "myNextTarget",
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertCalled(t, "RmTargets", opts.Host, opts.Project, []string{opts.NextTarget})
}

func Test_DeployDoesNotInvokeRmTargets_WhenBlueGreenIsFalse(t *testing.T) {
	opts := Opts{
		BlueGreen: false,
	}
	mockObj := getDockerComposeMock(opts, "")
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	Flow{}.Deploy(opts, mockObj)

	mockObj.AssertNotCalled(t, "RmTargets", opts.Host, opts.Project, []string{opts.NextTarget})
}

func Test_DeployReturnsError_WhenRmTargetsFails(t *testing.T) {
	opts := Opts{
		BlueGreen: true,
	}
	mockObj := getDockerComposeMock(opts, "RmTargets")
	mockObj.On("RmTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := Flow{}.Deploy(opts, mockObj)
	assert.Error(t, actual)
}

// Deploy > GetScaleCalc

func Test_DeployReturnsError_WhenGetScaleCalcFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "GetScaleCalc")
	scMockObj.On("GetScaleCalc", mock.Anything, mock.Anything, mock.Anything).Return(0, fmt.Errorf("This is an error"))
	serviceDiscovery = scMockObj

	actual := Flow{}.Deploy(opts, mockObj)

	assert.Error(t, actual)
}

// Deploy > ScaleTargets

func Test_DeployInvokesScaleTargets(t *testing.T) {
	opts := Opts{
		Host: "myHost",
		Project: "myProject",
		NextTarget: "myNextTarget",
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName: "myService",
		Scale: "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	flow := Flow{}
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	flow.Deploy(opts, mockObj)

	mockObj.AssertCalled(t, "ScaleTargets", opts.Host, opts.Project, opts.NextTarget, scale)
}

func Test_DeployReturnsError_WhenScaleTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "ScaleTargets")
	mockObj.On("ScaleTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Deploy(opts, mockObj)

	assert.Error(t, actual)
}

// Deploy > PutScale

func Test_DeployInvokesPutScale(t *testing.T) {
	opts := Opts{
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName: "myService",
		Scale: "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "")
	serviceDiscovery = scMockObj
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	Flow{}.Deploy(opts, mockObj)

	scMockObj.AssertCalled(t, "PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
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
	mockObj.On("RemoveFlow",).Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Deploy(s.opts, s.dc)

	s.Error(err)
}

// Scale > CreateFlowFile

func (s FlowTestSuite) Test_Scale_InvokesDockerComposeCreateFlowFile() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget)

	mockObj.AssertCalled(
		s.T(),
		"CreateFlowFile",
		s.opts.ComposePath,
		s.opts.Target,
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
	).Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget)

	s.Error(err)
}


// Scale > GetScaleCalc

func Test_ScaleReturnsError_WhenGetScaleCalcFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "GetScaleCalc")
	serviceDiscovery = scMockObj
	scMockObj.On("GetScaleCalc", mock.Anything, mock.Anything, mock.Anything).Return(0, fmt.Errorf("This is an error"))

	actual := Flow{}.Scale(opts, mockObj, "myTarget")

	assert.Error(t, actual)
}

// Scale > ScaleTargets

func Test_ScaleInvokesScaleTargets(t *testing.T) {
	opts := Opts{
		Host: "myHost",
		Project: "myProject",
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName: "myService",
		Scale: "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	flow := Flow{}
	serviceDiscovery = getServiceDiscoveryMock(opts, "")
	target := "myTarget"
	scale, _ := serviceDiscovery.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	flow.Scale(opts, mockObj, target)

	mockObj.AssertCalled(t, "ScaleTargets", opts.Host, opts.Project, target, scale)
}

func Test_ScaleReturnsError_WhenScaleTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "ScaleTargets")
	mockObj.On("ScaleTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))
	serviceDiscovery = getServiceDiscoveryMock(opts, "")

	actual := Flow{}.Scale(opts, mockObj, "myTarget")

	assert.Error(t, actual)
}

// Scale > PutScale

func Test_ScaleInvokesPutScale(t *testing.T) {
	opts := Opts{
		ServiceDiscoveryAddress: "mySeviceDiscoveryAddress",
		ServiceName: "myService",
		Scale: "34",
	}
	mockObj := getDockerComposeMock(opts, "")
	scMockObj := getServiceDiscoveryMock(opts, "")
	serviceDiscovery = scMockObj
	scale, _ := scMockObj.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	Flow{}.Scale(opts, mockObj, "myTarget")

	scMockObj.AssertCalled(t, "PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
}

// Deploy > RemoveFlow

func (s FlowTestSuite) Test_Scale_InvokesDockerComposeRemoveFlow() {
	mockObj := getDockerComposeMock(s.opts, "")
	s.dc = mockObj

	Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget)

	mockObj.AssertCalled(s.T(), "RemoveFlow")
}

func (s FlowTestSuite) Test_Scale_ReturnsError_WhenDockerComposeRemoveFlowFails() {
	mockObj := getDockerComposeMock(s.opts, "RemoveFlow")
	mockObj.On("RemoveFlow",).Return(fmt.Errorf("This is an error"))
	s.dc = mockObj

	err := Flow{}.Scale(s.opts, s.dc, s.opts.CurrentTarget)

	s.Error(err)
}

// GetTargets

func Test_GetTargetsReturnsAllTargets(t *testing.T) {
	opts := Opts{
		NextTarget: "myNextTarget",
		SideTargets: []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: true,
	}
	expected := append([]string{opts.NextTarget}, opts.SideTargets...)

	actual := Flow{}.GetTargets(opts)

	assert.Equal(t, expected, actual)
}

func Test_GetTargetsExcludesNextTarget_WhenSkipPullTarget(t *testing.T) {
	opts := Opts{
		NextTarget: "myNextTarget",
		SideTargets: []string{"sideTarget1", "sideTarget2"},
		SkipPullTarget: true,
		PullSideTargets: true,
	}
	expected := opts.SideTargets

	actual := Flow{}.GetTargets(opts)

	assert.Equal(t, expected, actual)
}

func Test_GetTargetsExcludesSideTargets_WhenNotPullSideTargets(t *testing.T) {
	opts := Opts{
		NextTarget: "myNextTarget",
		SideTargets: []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: false,
	}
	expected := []string{opts.NextTarget}

	actual := Flow{}.GetTargets(opts)

	assert.Equal(t, expected, actual)
}

// Proxy

func Test_Proxy_InvokesProvision(t *testing.T) {
	opts := Opts{
		ProxyDockerHost: "myProxyDockerHost",
		ProxyDockerCertPath: "myProxyCertPath",
		ServiceDiscoveryAddress: "myServiceDiscoveryAddress",
	}
	mockObj := getProxyMock("")

	Flow{}.Proxy(opts, mockObj)

	mockObj.AssertCalled(t, "Provision", opts.ProxyDockerHost, opts.ProxyDockerCertPath, opts.ServiceDiscoveryAddress)
}

func Test_Proxy_ReturnsError_WhenFailure(t *testing.T) {
	opts := Opts{}
	mockObj := getProxyMock("Provision")
	mockObj.On("Provision", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := Flow{}.Proxy(opts, mockObj)

	assert.Error(t, actual)
}

// Suite

func TestFlowTestSuite(t *testing.T) {
	suite.Run(t, new(FlowTestSuite))
}

// Mock

type FlowMock struct{
	mock.Mock
}

func (m *FlowMock) Deploy(opts Opts, dc DockerComposable) error {
	args := m.Called(opts, dc)
	return args.Error(0)
}

func (m *FlowMock) GetTargets(opts Opts) []string {
//	args := m.Called(opts)
	return []string{}
}

func (m *FlowMock) Scale(opts Opts, dc DockerComposable, target string) error {
	args := m.Called(opts, dc, target)
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
	if skipMethod != "GetTargets" {
		mockObj.On("GetTargets", mock.Anything).Return(nil)
	}
	if skipMethod != "Scale" {
		mockObj.On("Scale", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	}
	if skipMethod != "Proxy" {
		mockObj.On("Proxy", mock.Anything, mock.Anything).Return(nil)
	}
	return mockObj
}