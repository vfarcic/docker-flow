package main

import (
	"github.com/stretchr/testify/mock"
)

// Mock

type ServiceDiscoveryMock struct {
	mock.Mock
}

func (m *ServiceDiscoveryMock) GetScaleCalc(address, serviceName, scale string) (int, error) {
	args := m.Called(address, serviceName, scale)
	return args.Int(0), args.Error(1)
}

func (m *ServiceDiscoveryMock) GetNextColor(currentColor string) string {
	args := m.Called(currentColor)
	return args.String(0)
}

func (m *ServiceDiscoveryMock) GetColor(address, serviceName string) (string, error) {
	args := m.Called(address, serviceName)
	return args.String(0), args.Error(1)
}

func (m *ServiceDiscoveryMock) PutScale(address, serviceName string, value int) (string, error) {
	args := m.Called(address, serviceName, value)
	return args.String(0), args.Error(1)
}

func (m *ServiceDiscoveryMock) PutColor(address, serviceName, value string) (string, error) {
	args := m.Called(address, serviceName, value)
	return args.String(0), args.Error(1)
}

func getServiceDiscoveryMock(opts Opts, skipMethod string) *ServiceDiscoveryMock {
	mockObj := new(ServiceDiscoveryMock)
	scaleCalc := 5
	if skipMethod != "GetScaleCalc" {
		mockObj.On("GetScaleCalc", opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale).Return(scaleCalc, nil)
	}
	if skipMethod != "PutScale" {
		mockObj.On("PutScale", opts.ServiceDiscoveryAddress, opts.ServiceName, scaleCalc).Return("", nil)
	}
	if skipMethod != "GetColor" {
		mockObj.On("GetColor", opts.ServiceDiscoveryAddress, opts.ServiceName).Return("orange", nil)
	}
	if skipMethod != "GetNextColor" {
		mockObj.On("GetNextColor", opts.CurrentColor).Return("pink")
	}
	if skipMethod != "PutColor" {
		mockObj.On("PutColor", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
	}
	return mockObj
}
