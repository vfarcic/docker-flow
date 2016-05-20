package main

// TODO: Test

import (
	"fmt"
	"log"
	"strings"
)

func init() {
	log.SetPrefix(">> Docker Flow: ")
	log.SetFlags(0)
}

var logFatal = log.Fatal
var logPrintln = log.Println
var logPrintf = log.Printf
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
		case FLOW_DEPLOY:
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
		case FLOW_SCALE:
			if !deployed {
				logPrintln(fmt.Sprintf("Scaling (%s)...", opts.CurrentTarget))
				if err := flow.Scale(opts, dc, opts.CurrentTarget, true); err != nil {
					logFatal(err)
				}
			}
		case FLOW_STOP_OLD:
			// TODO: Move to flow
			if opts.BlueGreen {
				target := opts.CurrentTarget
				color := opts.CurrentColor
				if !deployed {
					target = opts.NextTarget
					color = opts.NextColor
				}
				logPrintln(fmt.Sprintf("Stopping old (%s)...", target))
				if err := dc.CreateFlowFile(
					opts.ComposePath,
					opts.ServiceName,
					opts.Target,
					opts.SideTargets,
					color,
					opts.BlueGreen,
				); err != nil {
					logFatal(err)
				}
				if err := dc.StopTargets(opts.Host, opts.CertPath, opts.Project, []string{target}); err != nil {
					logFatal(err)
				}
				if err := dc.RemoveFlow(); err != nil {
					logFatal(err)
				}
			}
			// TODO: End Move to flow
		case FLOW_PROXY:
			if err := flow.Proxy(opts, haProxy); err != nil {
				logFatal(err)
			}
		}

	}
}
