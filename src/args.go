package main

import (
	"os"
	"github.com/jessevdk/go-flags"
)

var Opts struct {
	Host				string 		`short:"H" long:"host" description:"Docker host"`
	ComposePath 		string 		`short:"c" long:"compose-path" default:"docker-compose.yml" description:"Docker Compose configuration file"`
	WebServer   		bool 		`short:"w" long:"web-server" default:"false" description:"Whether a Web server should be started"`
	BlueGreen			bool 		`short:"b" long:"blue-green" default:"false" description:"Whether to perform blue-green desployment"`
	Target				string 		`short:"t" long:"target" required:"true" description:"Docker Compose target that will be deployed"`
	SideTargets			[]string 	`short:"T" long:"side-targets" description:"Side or auxiliary targets that will be deployed. Multiple values are allowed."`
	PullSideTargets		bool		`short:"P" long:"pull-side-targets" default:"false" description:"Whether side or auxiliary targets should be pulled on each deployment."`
	Project				string 		`short:"p" long:"project" description:"Docker Compose project"`
	Scale				string 		`short:"s" long:"scale" default:"1" description:"Number of instances that should be deployed. If value starts with the plug sign (+), the number of instances will be increased by the given number. If value starts with the minus sign (-), the number of instances will be decreased by the given number."`
}

// TODO
func getArgs() {
	parseArgs()
//	parseYml(&args)
//	parseEnvironmentVars(&args)
//	TODO: Verify that scale is a number (with or without +/-)
}

func parseArgs() {
	if _, err := flags.ParseArgs(&Opts, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

//// TODO
//func parseYml(args *Args) {
//}
//
//// TODO
//func parseEnvironmentVars(args *Args) {
//	args.ComposePath = "xxx"
//}

