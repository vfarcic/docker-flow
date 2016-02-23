package main

func deploy() {
	targets := make([]string, 0);
	if !Opts.SkipPullTarget {
		targets = append(targets, Opts.NextTarget)
	}
	if Opts.PullSideTargets {
		targets = append(targets, Opts.SideTargets...)
	}
	if len(targets) > 0 {
		pullDockerComposeTargets(Opts.Host, Opts.Project, targets)
	}
	upDockerComposeTargets(Opts.Host, Opts.Project, Opts.SideTargets)
	if Opts.BlueGreen {
		rmDockerComposeTargets(Opts.Host, Opts.Project, []string{Opts.NextTarget})
	}
	scaleDockerComposeTargets(Opts.Host, Opts.ConsulAddress, Opts.Project, Opts.NextTarget, Opts.Scale)
}
