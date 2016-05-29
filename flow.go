package main

import (
	"fmt"
)

type Flowable interface {
	Deploy(opts Opts, dc DockerComposable) error
	GetPullTargets(opts Opts) []string
	Scale(opts Opts, dc DockerComposable, target string, createFlowFile bool) error
	Proxy(opts Opts, proxy Proxy) error
}

const FLOW_DEPLOY = "deploy"
const FLOW_SCALE = "scale"
const FLOW_STOP_OLD = "stop-old"
const FLOW_PROXY = "proxy"

type Flow struct{}

var flow Flowable = Flow{}

func getFlow() Flowable {
	return flow
}

func (m Flow) Deploy(opts Opts, dc DockerComposable) error {
	if err := dc.CreateFlowFile(
		opts.ComposePath,
		opts.ServiceName,
		opts.Target,
		opts.SideTargets,
		opts.NextColor,
		opts.BlueGreen,
	); err != nil {
		return fmt.Errorf("Failed to create the Docker Flow file\n%s\n", err.Error())
	}
	logPrintln(fmt.Sprintf("Deploying (%s)...", opts.NextTarget))

	if err := dc.PullTargets(opts.Host, opts.CertPath, opts.Project, m.GetPullTargets(opts)); err != nil {
		return fmt.Errorf("The deployment phase failed (pull)\n%s", err.Error())
	}
	if opts.BlueGreen {
		if err := dc.RmTargets(opts.Host, opts.CertPath, opts.Project, []string{opts.NextTarget}); err != nil {
			return fmt.Errorf("The deployment phase failed (rm)\n%s", err.Error())
		}
	}
	targets := append(opts.SideTargets, opts.NextTarget)
	if err := dc.UpTargets(opts.Host, opts.CertPath, opts.Project, targets); err != nil {
		return fmt.Errorf("The deployment phase failed (up)\n%s", err.Error())
	}
	if err := m.Scale(opts, dc, opts.NextTarget, false); err != nil {
		return err
	}
	if err := dc.RemoveFlow(); err != nil {
		return err
	}
	return nil
}

func (m Flow) Scale(opts Opts, dc DockerComposable, target string, createFlowFile bool) error {
	if createFlowFile {
		if err := dc.CreateFlowFile(
			opts.ComposePath,
			opts.ServiceName,
			opts.Target,
			opts.SideTargets,
			opts.CurrentColor,
			opts.BlueGreen,
		); err != nil {
			return fmt.Errorf("Failed to create the Docker Flow file\n%s\n", err.Error())
		}
	}
	sc := getServiceDiscovery()
	scale, err := sc.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)
	if err != nil {
		return err
	}
	if err := dc.ScaleTargets(opts.Host, opts.CertPath, opts.Project, target, scale); err != nil {
		return fmt.Errorf("The scale phase failed\n%s", err.Error())
	}
	sc.PutScale(opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
	if createFlowFile {
		if err := dc.RemoveFlow(); err != nil {
			return err
		}
	}
	return nil
}

func (m Flow) Proxy(opts Opts, proxy Proxy) error {
	if err := proxy.Provision(
		opts.ProxyDockerHost,
		opts.ProxyReconfPort,
		opts.ProxyDockerCertPath,
		opts.ServiceDiscoveryAddress,
	); err != nil {
		return err
	}
	color := opts.CurrentColor
	if m.contains(opts.Flow, FLOW_DEPLOY) {
		color = opts.NextColor
	}
	if err := proxy.Reconfigure(
		opts.ProxyDockerHost,
		opts.ProxyDockerCertPath,
		opts.ProxyHost,
		opts.ProxyReconfPort,
		opts.ServiceName,
		color,
		opts.ServicePath,
		opts.ConsulTemplateFePath,
		opts.ConsulTemplateBePath,
	); err != nil {
		return err
	}
	return nil
}

func (m Flow) GetPullTargets(opts Opts) []string {
	targets := make([]string, 0)
	targets = append(targets, opts.NextTarget)
	if opts.PullSideTargets {
		targets = append(targets, opts.SideTargets...)
	}
	return targets
}

func (m Flow) contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
