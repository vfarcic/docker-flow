package main

import (
	"os"
	"github.com/jessevdk/go-flags"
	"fmt"
	"strings"
)

var Opts struct {
	Host				string 		`short:"H" long:"host" description:"Docker host"`
	ComposePath 		string 		`short:"f" long:"compose-path" default:"docker-compose.yml" description:"Docker Compose configuration file"`
	WebServer   		bool 		`short:"w" long:"web-server" default:"false" description:"Whether a Web server should be started"`
	BlueGreen			bool 		`short:"b" long:"blue-green" default:"false" description:"Whether to perform blue-green desployment"`
	Target				string 		`short:"t" long:"target" required:"true" description:"Docker Compose target that will be deployed"`
	SideTargets			[]string 	`short:"T" long:"side-targets" description:"Side or auxiliary targets that will be deployed. Multiple values are allowed."`
	SkipPullTarget		bool		`short:"P" long:"skip-pull-targets" default:"false" description:"Whether to skip pulling target."`
	PullSideTargets		bool		`short:"S" long:"pull-side-targets" default:"false" description:"Whether side or auxiliary targets should be pulled."`
	Project				string 		`short:"p" long:"project" description:"Docker Compose project. If not specified, current directory will be used instead."`
	Scale         		string 		`short:"s" long:"scale" default:"1" description:"Number of instances that should be deployed. If value starts with the plug sign (+), the number of instances will be increased by the given number. If value starts with the minus sign (-), the number of instances will be decreased by the given number."`
	ConsulAddress 		string 		`short:"c" long:"consul-address" required:"true" description:"The address of the consul server."`
	ServiceName   		string
	CurrentColor  		string
	NextColor     		string
	CurrentTarget 		string
	NextTarget    		string
}

// TODO
func getArgs() {
	parseArgs()
//	parseYml(&args)
//	parseEnvironmentVars(&args)
	if len(Opts.Project) == 0 {
		dir, _ := os.Getwd()
		Opts.Project = dir[strings.LastIndex(dir, "/") + 1:]
	}
	//	TODO: Verify that scale is a number (with or without +/-)
	Opts.ServiceName = fmt.Sprintf("%s_%s", Opts.Project, Opts.Target)
	Opts.CurrentColor = getConsulColor(Opts.ConsulAddress, Opts.ServiceName)
	Opts.NextColor = getConsulNextColor(Opts.ConsulAddress, Opts.ServiceName)
	if Opts.BlueGreen {
		Opts.NextTarget = fmt.Sprintf("%s-%s", Opts.Target, Opts.NextColor)
		Opts.CurrentTarget = fmt.Sprintf("%s-%s", Opts.Target, Opts.CurrentColor)
	}
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

