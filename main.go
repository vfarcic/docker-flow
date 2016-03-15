package main
// TODO: Test

import (
	"log"
	"strings"
	"os"
)

func init() {
	log.SetPrefix(">> Docker Flow: ")
	log.SetFlags(0)
}

var logFatal = log.Fatal
var osExit = os.Exit
var logPrintln = log.Println
var deployed = false

func main() {
	flow := getFlow()
	sc := getServiceDiscovery()

	opts, err := GetOpts()
	if err != nil {
		osExit(1)
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
				// TODO: Move to flow
				if err := dc.CreateFlowFile(
					opts.ComposePath,
					dockerComposeFlowPath,
					opts.Target,
					opts.CurrentColor,
					opts.BlueGreen,
				); err != nil {
					logFatal(err)
				}
				// TODO: End Move to flow
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
				if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, color, opts.BlueGreen); err != nil {
					logFatal(err)
				}
				if err := dc.StopTargets(opts.Host, opts.Project, []string{target}); err != nil {
					logFatal(err)
				}
			}
			// TODO: End Move to flow
		}
	}
	// cleanup
	if err := dc.RemoveFlow(); err != nil {
		logFatal(err)
	}
}
