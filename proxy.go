package main

type Proxy interface {
	Provision(host, reconfPort, certPath, scAddress string) error
	Reconfigure(domain, reconfPort, project, servicePath string) error
}

