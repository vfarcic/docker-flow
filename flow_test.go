package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"fmt"
)

// Deploy

func Test_DeployReturnsNil(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "")

	actual := FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

	assert.Nil(t, actual)
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
	flow := FlowImpl{}

	flow.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

	mockObj.AssertCalled(t, "PullTargets", opts.Host, opts.Project, flow.GetTargets(opts))
}

func Test_DeployReturnsError_WhenPullTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "PullTargets")
	mockObj.On("PullTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

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
	flow := FlowImpl{}

	flow.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

	mockObj.AssertCalled(t, "UpTargets", opts.Host, opts.Project, opts.SideTargets)
}

func Test_DeployReturnsError_WhenUpTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "UpTargets")
	mockObj.On("UpTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

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

	FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

	mockObj.AssertCalled(t, "RmTargets", opts.Host, opts.Project, []string{opts.NextTarget})
}

func Test_DeployDoesNotInvokeRmTargets_WhenBlueGreenIsFalse(t *testing.T) {
	opts := Opts{
		BlueGreen: false,
	}
	mockObj := getDockerComposeMock(opts, "")

	FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

	mockObj.AssertNotCalled(t, "RmTargets", opts.Host, opts.Project, []string{opts.NextTarget})
}

func Test_DeployReturnsError_WhenRmTargetsFails(t *testing.T) {
	opts := Opts{
		BlueGreen: true,
	}
	mockObj := getDockerComposeMock(opts, "RmTargets")
	mockObj.On("RmTargets", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)
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
	flow := FlowImpl{}
	sc := getServiceDiscoveryMock(opts)
	scale, _ := sc.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	flow.Deploy(opts, sc, mockObj)

	mockObj.AssertCalled(t, "ScaleTargets", opts.Host, opts.Project, opts.NextTarget, scale)
}

func Test_DeployReturnsError_WhenScaleTargetsFails(t *testing.T) {
	opts := Opts{}
	mockObj := getDockerComposeMock(opts, "ScaleTargets")
	mockObj.On("ScaleTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("This is an error"))

	actual := FlowImpl{}.Deploy(opts, getServiceDiscoveryMock(opts), mockObj)

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
	sc := getServiceDiscoveryMock(opts)
	scale, _ := sc.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)

	FlowImpl{}.Deploy(opts, sc, mockObj)

	sc.AssertCalled(t, "PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
}

// GetTargets

func Test_GetTargetsReturnsAllTargets(t *testing.T) {
	opts := Opts{
		NextTarget: "myNextTarget",
		SideTargets: []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: true,
	}
	expected := append([]string{opts.NextTarget}, opts.SideTargets...)

	actual := FlowImpl{}.GetTargets(opts)

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

	actual := FlowImpl{}.GetTargets(opts)

	assert.Equal(t, expected, actual)
}

func Test_GetTargetsExcludesSideTargets_WhenNotPullSideTargets(t *testing.T) {
	opts := Opts{
		NextTarget: "myNextTarget",
		SideTargets: []string{"sideTarget1", "sideTarget2"},
		PullSideTargets: false,
	}
	expected := []string{opts.NextTarget}

	actual := FlowImpl{}.GetTargets(opts)

	assert.Equal(t, expected, actual)
}
