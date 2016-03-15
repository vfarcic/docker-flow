package main

type Proxy interface {
	Provision(host, scAddress string) error
}

