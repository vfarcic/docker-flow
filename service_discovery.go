package main

const BlueColor = "blue"
const GreenColor = "green"

var serviceDiscovery ServiceDiscovery = Consul{}

func getServiceDiscovery() ServiceDiscovery {
	return serviceDiscovery
}

type ServiceDiscovery interface {
	GetScaleCalc(address, serviceName, scale string) (int, error)
	GetNextColor(currentColor string) string
	GetColor(address, serviceName string) (string, error)
	PutScale(address, serviceName string, value int) (string, error)
	PutColor(address, serviceName, value string) (string, error)
}
