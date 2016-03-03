package dockerflow
// TODO: Test

import (
	"log"
)

func init() {
	log.SetPrefix(">> Docker Flow: ")
	log.SetFlags(0)
}

func main() {
	flow := FlowImpl{}

	log.Println("Parsing arguments...")
	opts, err := GetOpts()
	if err != nil {
		log.Fatal(err)
	}
	dc := DockerComposeImpl{}
	if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, opts.NextColor, opts.BlueGreen); err != nil {
		log.Fatal(err)
	}

	log.Println("Deploying...")
	if err := flow.Deploy(opts, opts.ServiceDiscovery, dc); err != nil {
		log.Fatal(err)
	}

	log.Println("Cleaning...")
	// TODO: Move
	if _, err := opts.ServiceDiscovery.PutColor(opts.ServiceDiscoveryAddress, opts.ServiceName, opts.NextColor); err != nil {
		log.Fatal(err)
	}
	if opts.BlueGreen {
		if err := dc.CreateFlowFile(opts.ComposePath, dockerComposeFlowPath, opts.Target, opts.CurrentColor, opts.BlueGreen); err != nil {
			log.Fatal(err)
		}
		if err := dc.StopTargets(opts.Host, opts.Project, []string{opts.CurrentTarget}); err != nil {
			log.Fatal(err)
		}
	}
	// TODO: End Move

	if err := dc.RemoveFlow(); err != nil {
		log.Fatal(err)
	}
}
