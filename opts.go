package main

import (
	"os"
	"github.com/jessevdk/go-flags"
	"fmt"
	"strings"
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
	"strconv"
)

const dockerFlowPath  = "docker-flow.yml"
const dockerComposePath = "docker-compose.yml"

var getWd = os.Getwd
var parseYml = ParseYml
var parseEnvVars = ParseEnvVars
var parseArgs = ParseArgs
var processOpts = ProcessOpts

type Opts struct {
	Host        			string 		`short:"H" long:"host" description:"Docker daemon socket to connect to. If not specified, DOCKER_HOST environment variable will be used instead."`
	ComposePath 			string 		`short:"f" long:"compose-path" value-name:"docker-compose.yml" description:"Path to the Docker Compose configuration file." yaml:"compose_path" envconfig:"compose_path"`
	BlueGreen   			bool 		`short:"b" long:"blue-green" description:"Perform blue-green desployment." yaml:"blue_green" envconfig:"blue_green"`
	Target					string 		`short:"t" long:"target" description:"Docker Compose target."`
	SideTargets             []string 	`short:"T" long:"side-target" description:"Side or auxiliary Docker Compose targets. Multiple values are allowed." yaml:"side_targets" envconfig:"side_targets"`
	SkipPullTarget          bool		`short:"P" long:"skip-pull-targets" description:"Skip pulling targets." yaml:"skip_pull_target" envconfig:"skip_pull_target"`
	PullSideTargets         bool		`short:"S" long:"pull-side-targets" description:"Pull side or auxiliary targets." yaml:"pull_side_targets" envconfig:"pull_side_targets"`
	Project                 string 		`short:"p" long:"project" description:"Docker Compose project. If not specified, current directory will be used instead."`
	ServiceDiscoveryAddress string 		`short:"c" long:"consul-address" description:"The address of the C	onsul server." yaml:"consul_address" envconfig:"consul_address"`
	Scale                   string		`short:"s" long:"scale" description:"Number of instances to deploy. If value starts with the plug sign (+), the number of instances will be increased by the given number. If value starts with the minus sign (-), the number of instances will be decreased by the given number."`
	Flow		            []string 	`short:"F" long:"flow" choice:"deploy" choice:"scale" choice:"stop-old" description:"The actions that should be performed as the flow. Multiple values are allowed.\ndeploy: Deploys a new release\nscale: Scales currently running release\nstop-old: Stops the old release\n" yaml:"flow" envconfig:"flow"`
	ServiceDiscovery        ServiceDiscovery
	ServiceName             string
	CurrentColor    		string
	NextColor       		string
	CurrentTarget   		string
	NextTarget      		string
}

func GetOpts() (Opts, error) {
	opts := Opts{
		ComposePath: dockerComposePath,
		Flow: []string{"deploy"},
	}
	if err := parseYml(&opts); err != nil {
		return opts, err
	}
	if err := parseEnvVars(&opts); err != nil {
		return opts, err
	}
	if err := parseArgs(&opts); err != nil {
		return opts, err
	}
	if err := processOpts(&opts); err != nil {
		return opts, err
	}
	return opts, nil
}

func ParseYml(opts *Opts) error {
	data, err := readFile(dockerFlowPath)
	if err != nil {
		return fmt.Errorf("Could not read the Docker Flow file %s\n%v", dockerFlowPath, err)
	}
	if err = yaml.Unmarshal([]byte(data), opts); err != nil {
		return fmt.Errorf("Could not parse the Docker Flow file %s\n%v", dockerFlowPath, err)
	}
	return nil
}

func ParseArgs(opts *Opts) error {
	if _, err := flags.ParseArgs(opts, os.Args[1:]); err != nil {
		return fmt.Errorf("Could not parse command line arguments\n%v", err)
	}
	return nil
}


func ParseEnvVars(opts *Opts) error {
	if err := envconfig.Process("flow", opts); err != nil {
		return fmt.Errorf("Could not retrieve environment variables\n%v", err)
	}
	data := []struct{
		key 		string
		value		*[]string
	}{
		{"FLOW_SIDE_TARGETS", &opts.SideTargets},
		{"FLOW", &opts.Flow},
	}
	for _, d := range data {
		value := strings.Trim(os.Getenv(d.key), " ")
		if len(value) > 0 {
			*d.value = strings.Split(value, ",")
		}
	}
	return nil
}

func ProcessOpts(opts *Opts) (err error) {
	if opts.ServiceDiscovery == nil {
		opts.ServiceDiscovery = Consul{}
	}
	if len(opts.Project) == 0 {
		dir, _ := getWd()
		opts.Project = dir[strings.LastIndex(dir, "/") + 1:]
	}
	if len(opts.Target) == 0 {
		return fmt.Errorf("target argument is required")
	}
	if len(opts.ServiceDiscoveryAddress) == 0 {
		return fmt.Errorf("consul-address argument is required")
	}
	if len(opts.Scale) != 0 {
		if _, err := strconv.Atoi(opts.Scale); err != nil {
			return fmt.Errorf("scale must be a number or empty")
		}
	}
	if len(opts.Flow) == 0 {
		opts.Flow = []string{"deploy"}
	}
	if len(opts.ServiceName) == 0 {
		opts.ServiceName = fmt.Sprintf("%s-%s", opts.Project, opts.Target)
	}
	if opts.CurrentColor, err = opts.ServiceDiscovery.GetColor(opts.ServiceDiscoveryAddress, opts.ServiceName); err != nil {
		return err
	}
	if len(opts.Host) == 0 {
		opts.Host = os.Getenv("DOCKER_HOST")
	}
	opts.NextColor = opts.ServiceDiscovery.GetNextColor(opts.CurrentColor)
	if opts.BlueGreen {
		opts.NextTarget = fmt.Sprintf("%s-%s", opts.Target, opts.NextColor)
		opts.CurrentTarget = fmt.Sprintf("%s-%s", opts.Target, opts.CurrentColor)
	} else {
		opts.NextTarget = opts.Target
		opts.CurrentTarget = opts.Target
	}
	return nil
}

