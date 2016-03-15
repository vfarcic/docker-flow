package main

import (
	"fmt"
)

type Flowable interface {
	Deploy(opts Opts, dc DockerComposable) error
	GetTargets(opts Opts) []string
	Scale(opts Opts, dc DockerComposable, target string) error
	Proxy(opts Opts, proxy Proxy) error
}

type Flow struct{}

var flow Flowable = Flow{}
func getFlow() Flowable {
	return flow
}

func (m Flow) Deploy(opts Opts, dc DockerComposable) error {
	if err := dc.CreateFlowFile(
		opts.ComposePath,
		dockerComposeFlowPath,
		opts.Target,
		opts.NextColor,
		opts.BlueGreen,
	); err != nil {
		return fmt.Errorf("Creationg of the Docker Flow file failed\n%v", err)
	}

	targets := m.GetTargets(opts)
	if err := dc.PullTargets(opts.Host, opts.Project, targets); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	if err := dc.UpTargets(opts.Host, opts.Project, opts.SideTargets); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	if opts.BlueGreen {
		if err := dc.RmTargets(opts.Host, opts.Project, []string{opts.NextTarget}); err != nil {
			return fmt.Errorf("The deployment phase failed\n%v", err)
		}
	}
	if err := m.Scale(opts, dc, opts.NextTarget); err != nil {
		return err
	}
	return nil
}

func (m Flow) Scale(opts Opts, dc DockerComposable, target string) error {
	sc := getServiceDiscovery()
	scale, err := sc.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)
	if err != nil {
		return err
	}
	if err := dc.ScaleTargets(opts.Host, opts.Project, target, scale); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	sc.PutScale(opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
	return nil
}

func (m Flow) Proxy(opts Opts, proxy Proxy) error {
	return proxy.Provision(opts.ProxyHost, opts.ServiceDiscoveryAddress)
}

func (m Flow) GetTargets(opts Opts) []string {
	targets := make([]string, 0);
	if !opts.SkipPullTarget {
		targets = append(targets, opts.NextTarget)
	}
	if opts.PullSideTargets {
		targets = append(targets, opts.SideTargets...)
	}
	return targets
}
