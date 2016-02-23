package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const dockerComposeFlowPath  = "docker-compose-flow.yml"

func createDockerComposeFlow(dockerComposePath, target, color string, blueGreen bool) {
	data, err := ioutil.ReadFile(dockerComposePath)
	if err != nil {
		fmt.Errorf("Could not read the Docker Compose file %s\n%v\n", dockerComposePath, err)
		os.Exit(1)
	}
	s := string(data)
	if blueGreen {
		old := fmt.Sprintf("%s:", target)
		new := fmt.Sprintf("%s-%s:", target, color)
		s = strings.Replace(string(data), old, new, 1)
	}
	err = ioutil.WriteFile(dockerComposeFlowPath, []byte(s), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func removeDockerComposeFlow() {
	if err := os.Remove(dockerComposeFlowPath); err != nil {
		fmt.Printf("Could not remove the temp file %s", dockerComposeFlowPath)
	}
}

// TODO
func pullDockerComposeTargets(host, project string, targets []string) {
	args := getDockerComposeArgs(host, project)
	args = append(args, "pull")
	args = append(args, targets...)
	runDockerComposeCmd(args)
}

func upDockerComposeTargets(host, project string, targets []string) {
	args := getDockerComposeArgs(host, project)
	args = append(args, "up", "-d")
	args = append(args, targets...)
	runDockerComposeCmd(args)
}

func scaleDockerComposeTargets(host, consulAddress, project, target, scale string) {
	args := getDockerComposeArgs(host, project)
	args = append(args, "scale")
	// TODO: Change xxx to serviceName
	// TODO: Add color to serviceName
	s := getConsulScale(consulAddress, Opts.ServiceName, scale)
	args = append(args, fmt.Sprintf("%s=%d", target, s))
	runDockerComposeCmd(args)
	putConsulScale(Opts.ConsulAddress, Opts.ServiceName, s)
}

func rmDockerComposeTargets(host, project string, targets []string) {
	stopDockerComposeTargets(host, project, targets)
	args := getDockerComposeArgs(host, project)
	args = append(args, "rm", "-f")
	args = append(args, targets...)
	runDockerComposeCmd(args)
}

func stopDockerComposeTargets(host, project string, targets []string) {
	args := getDockerComposeArgs(host, project)
	args = append(args, "stop")
	args = append(args, targets...)
	runDockerComposeCmd(args)
}

func getDockerComposeArgs(host, project string) []string {
	args := []string{"-f", "docker-compose-flow.yml"}
	if (len(host) > 0) {
		env := os.Environ()
		env = append(env, fmt.Sprintf("DOCKER_HOST=%s", host))
	}
	if (len(project) > 0) {
		args = append(args, "-p", project)
	}
	return args
}

func runDockerComposeCmd(args []string) {
	cmd := exec.Command("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Docker Compose command failed.", err)
		os.Exit(1)
	}
}
