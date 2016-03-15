package main

import "github.com/stretchr/testify/mock"

// Mock

type ProxyMock struct{
	mock.Mock
}

func (m *ProxyMock) Provision(host, scAddress string) error {
	args := m.Called(host, scAddress)
	return args.Error(0)
}


func getProxyMock(host, scAddress, skipMethod string) *ProxyMock {
	mockObj := new(ProxyMock)
	if skipMethod != "Provision" {
		mockObj.On("Provision", host, scAddress).Return(nil)
	}
	return mockObj
}
