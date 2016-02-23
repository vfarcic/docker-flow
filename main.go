package main

import (
	"log"
)

func init() {
	log.SetPrefix("Docker Flow: ")
	log.SetFlags(0)
}

func main() {
	opts := Opts{
		ComposePath: "docker-compose.yml",
		Scale: "1",
	}
	if err := getArgs(&opts); err != nil {
		log.Fatal(err)
	}

	if err := createDockerComposeFlow(opts.ComposePath, opts.Target, opts.NextColor, opts.BlueGreen); err != nil {
		log.Fatal(err)
	}

	if err := deploy(opts); err != nil {
		log.Fatal(err)
	}

	if err := putConsulColor(opts.ConsulAddress, opts.ServiceName, opts.NextColor); err != nil {
		log.Fatal(err)
	}
	if opts.BlueGreen {
		if err := createDockerComposeFlow(opts.ComposePath, opts.Target, opts.CurrentColor, opts.BlueGreen); err != nil {
			log.Fatal(err)
		}
		if err := stopDockerComposeTargets(opts.Host, opts.Project, []string{opts.CurrentTarget}); err != nil {
			log.Fatal(err)
		}
	}

	if err := removeDockerComposeFlow(); err != nil {
		log.Fatal(err)
	}
}
