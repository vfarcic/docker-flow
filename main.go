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

func main() {
	flow := FlowImpl{}

	opts, err := GetOpts()
	if err != nil {
		os.Exit(1)
	}
	dc := DockerComposeImpl{}

	deployed := false
	for _, step := range opts.Flow {
		switch strings.ToLower(step) {
		case "deploy":
			log.Println("Deploying...")
			if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, opts.NextColor, opts.BlueGreen); err != nil {
				log.Fatal(err)
			}
			if err := flow.Deploy(opts, opts.ServiceDiscovery, dc); err != nil {
				log.Fatal(err)
			}
			deployed = true
			// TODO: Move to flow
			log.Println("Cleaning...")
			if _, err := opts.ServiceDiscovery.PutColor(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.NextColor); err != nil {
				log.Fatal(err)
			}
			// TODO: End Move to flow
		case "scale":
			if !deployed {
				log.Println("Scaling...")
				if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, opts.CurrentColor, opts.BlueGreen); err != nil {
					log.Fatal(err)
				}
				if err := flow.Scale(opts, opts.ServiceDiscovery, dc, opts.CurrentTarget); err != nil {
					log.Fatal(err)
				}
			}
		case "stop-old":
			// TODO: Move to flow
			if opts.BlueGreen {
				target := opts.CurrentTarget
				color := opts.CurrentColor
				if !deployed {
					target = opts.NextTarget
					color = opts.NextColor
				}
				if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, color, opts.BlueGreen); err != nil {
					log.Fatal(err)
				}
				if err := dc.StopTargets(opts.Host, opts.Project, []string{target}); err != nil {
					log.Fatal(err)
				}
			}
			// TODO: End Move to flow
		}
	}

	if err := dc.RemoveFlow(); err != nil {
		log.Fatal(err)
	}
}
