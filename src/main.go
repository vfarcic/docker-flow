package main

func init() {
	getArgs()
}

func main() {
	if Opts.WebServer {
		startWebServer()
	}

	createDockerComposeFlow(Opts.ComposePath, Opts.Target, Opts.NextColor, Opts.BlueGreen)

	deploy()

	putConsulColor(Opts.ConsulAddress, Opts.ServiceName, Opts.NextColor)
	if Opts.BlueGreen {
		createDockerComposeFlow(Opts.ComposePath, Opts.Target, Opts.CurrentColor, Opts.BlueGreen)
		stopDockerComposeTargets(Opts.Host, Opts.Project, []string{Opts.CurrentTarget})
	}

	removeDockerComposeFlow()
}
