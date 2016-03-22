package main
// TODO: Test

import (
	"log"
	"strings"
)

func init() {
	log.SetPrefix(">> Docker Flow: ")
	log.SetFlags(0)
}

var logFatal = log.Fatal
var logPrintln = log.Println
var deployed = false

func main() {
//	createdFlow := false
	flow := getFlow()
	sc := getServiceDiscovery()

	opts, err := GetOpts()
	if err != nil {
		logFatal(err)
	}
	dc := getDockerCompose()

	for _, step := range opts.Flow {
		switch strings.ToLower(step) {
		case "deploy":
			logPrintln("Deploying...")
			if err := flow.Deploy(opts, dc); err != nil {
				logFatal(err)
			}
			deployed = true
			// TODO: Move to flow
			logPrintln("Cleaning...")
			if _, err := sc.PutColor(
				opts.ServiceDiscoveryAddress,
				opts.ServiceName,
				opts.NextColor,
			); err != nil {
				logFatal(err)
			}
			// TODO: End Move to flow
		case "scale":
			if !deployed {
				logPrintln("Scaling...")
				if err := flow.Scale(opts, dc, opts.CurrentTarget); err != nil {
					logFatal(err)
				}
			}
		case "stop-old":
			// TODO: Move to flow
			if opts.BlueGreen {
				logPrintln("Stopping old...")
				target := opts.CurrentTarget
				color := opts.CurrentColor
				if !deployed {
					target = opts.NextTarget
					color = opts.NextColor
				}
				if err := dc.CreateFlowFile(opts.ComposePath, opts.Target, color, opts.BlueGreen); err != nil {
					logFatal(err)
				}
				if err := dc.StopTargets(opts.Host, opts.Project, []string{target}); err != nil {
					logFatal(err)
				}
				if err := dc.RemoveFlow(); err != nil {
					logFatal(err)
				}
			}
			// TODO: End Move to flow
		case "proxy":
			if err := flow.Proxy(opts, haProxy); err != nil {
				logFatal(err)
			}
		}

	}
}
