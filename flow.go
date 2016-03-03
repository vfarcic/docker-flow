package dockerflow

import (
	"fmt"
)

type Flow interface {
	Deploy(opts Opts, sc ServiceDiscovery, dc DockerCompose) error
	GetTargets(opts Opts) []string
}

type FlowImpl struct{}

func (flow FlowImpl) Deploy(opts Opts, sc ServiceDiscovery, dc DockerCompose) error {
	targets := flow.GetTargets(opts)
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
	scale, err := sc.GetScaleCalc(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.Scale)
	if err != nil {
		return err
	}
	if err := dc.ScaleTargets(opts.Host, opts.Project, opts.NextTarget, scale); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	sc.PutScale(opts.ServiceDiscoveryAddress, opts.ServiceName, scale)
	return nil
}

func (flow FlowImpl) GetTargets(opts Opts) []string {
	targets := make([]string, 0);
	if !opts.SkipPullTarget {
		targets = append(targets, opts.NextTarget)
	}
	if opts.PullSideTargets {
		targets = append(targets, opts.SideTargets...)
	}
	return targets
}
