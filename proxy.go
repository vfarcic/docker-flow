package main

type Proxy interface {
	Provision(dockerHost, reconfPort, certPath, scAddress string) error
	Reconfigure(dockerHost, proxyCertPath, host, reconfPort, serviceName, serviceColor string, servicePath []string, consulTemplatePath string) error
}
