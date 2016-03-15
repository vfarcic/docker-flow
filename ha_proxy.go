package main
import (
	"os/exec"
	"os"
	"fmt"
)

type HaProxy struct {}

var runHaProxyCmd = runCmd
//var lsHaProxyCmd = runCmd

func (m HaProxy) Provision(host, scAddress string) error {
	args := []string{
		"run", "-d",
		"--name", "docker-flow-proxy",
		"-e", fmt.Sprintf("%s=%s", "CONSUL_ADDRESS", scAddress),
		"-p", "80:80",
		"vfarcic/docker-flow-proxy", "run",
	}
	if len(host) == 0 {
		return fmt.Errorf("Proxy host is mandatory for the proxy step")
	}
	SetDockerHost(host)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := runHaProxyCmd(cmd); err != nil {
		return fmt.Errorf("Docker Compose command %v\n%v", cmd, err)
	}
	return nil
}
