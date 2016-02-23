package main

import "fmt"

func deploy(opts Opts) error {
	targets := make([]string, 0);
	if !opts.SkipPullTarget {
		targets = append(targets, opts.NextTarget)
	}
	if opts.PullSideTargets {
		targets = append(targets, opts.SideTargets...)
	}
	if err := pullDockerComposeTargets(opts.Host, opts.Project, targets); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	if err := upDockerComposeTargets(opts.Host, opts.Project, opts.SideTargets); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	if opts.BlueGreen {
		if err := rmDockerComposeTargets(opts.Host, opts.Project, []string{opts.NextTarget}); err != nil {
			return fmt.Errorf("The deployment phase failed\n%v", err)
		}
	}
	if err := scaleDockerComposeTargets(opts.Host, opts.ConsulAddress, opts.Project, opts.NextTarget, opts.ServiceName, opts.Scale); err != nil {
		return fmt.Errorf("The deployment phase failed\n%v", err)
	}
	return nil
}
