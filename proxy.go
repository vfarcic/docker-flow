package main

type Proxy interface {
	Provision(host, reconfPort, certPath, scAddress string) error
	Reconfigure(host, reconfPort, serviceName, serviceColor string, servicePath []string) error
}
