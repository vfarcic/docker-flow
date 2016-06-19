package main

import "github.com/stretchr/testify/mock"

type DockerComposeMock struct {
	mock.Mock
}

func (m *DockerComposeMock) CreateFlowFile(
dcPath,
serviceName,
target string,
sideTargets []string,
color string,
blueGreen bool,
) error {
	args := m.Called(dcPath, serviceName, target, sideTargets, color, blueGreen)
	return args.Error(0)
}

func (m *DockerComposeMock) RemoveFlow() error {
	args := m.Called()
	return args.Error(0)
}

func (m *DockerComposeMock) PullTargets(host, certPath, project string, targets []string) error {
	args := m.Called(host, certPath, project, targets)
	return args.Error(0)
}

func (m *DockerComposeMock) UpTargets(host, certPath, project string, targets []string) error {
	args := m.Called(host, certPath, project, targets)
	return args.Error(0)
}

func (m *DockerComposeMock) ScaleTargets(host, certPath, project, target string, scale int) error {
	args := m.Called(host, certPath, project, target, scale)
	return args.Error(0)
}

func (m *DockerComposeMock) RmTargets(host, certPath, project string, targets []string) error {
	args := m.Called(host, certPath, project, targets)
	return args.Error(0)
}

func (m *DockerComposeMock) StopTargets(host, certPath, project string, targets []string) error {
	args := m.Called(host, certPath, project, targets)
	return args.Error(0)
}

func getDockerComposeMock(opts Opts, skipMethod string) *DockerComposeMock {
	mockObj := new(DockerComposeMock)
	if skipMethod != "PullTargets" {
		mockObj.On("PullTargets", opts.Host, opts.CertPath, opts.Project, Flow{}.GetPullTargets(opts)).Return(nil)
	}
	if skipMethod != "UpTargets" {
		mockObj.On("UpTargets", opts.Host, opts.CertPath, opts.Project, append(opts.SideTargets, opts.NextTarget)).Return(nil)
	}
	if skipMethod != "RmTargets" {
		mockObj.On("RmTargets", opts.Host, opts.CertPath, opts.Project, []string{opts.NextTarget}).Return(nil)
	}
	if skipMethod != "ScaleTargets" {
		mockObj.On("ScaleTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	}
	if skipMethod != "CreateFlowFile" {
		mockObj.On(
			"CreateFlowFile",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(nil)
	}
	if skipMethod != "StopTargets" {
		mockObj.On("StopTargets", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	}
	if skipMethod != "RemoveFlow" {
		mockObj.On("RemoveFlow").Return(nil)
	}
	return mockObj
}
