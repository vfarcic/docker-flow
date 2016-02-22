package main

import (
	"os"
	"fmt"
)

func main() {
	getArgs()
	if Opts.WebServer {
		startWebServer()
	}

	if err := createDockerComposeFlow(Opts.ComposePath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

//	targets := []string{Opts.Target};
//	if Opts.PullSideTargets {
//		targets = append(targets, Opts.SideTargets...)
//	}
//	pullDockerComposeTargets(Opts.Host, Opts.Project, targets)
//	upDockerComposeTargets(Opts.Host, Opts.Project, Opts.SideTargets)
//	rmDockerComposeTargets(Opts.Host, Opts.Project, []string{Opts.Target})
	scaleDockerComposeTargets(Opts.Host, Opts.Project, Opts.Target, Opts.Scale)

//	removeDockerComposeFlow()
}
