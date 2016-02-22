package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

const dockerComposeFlowPath  = "docker-compose-flow.yml"

func createDockerComposeFlow(dockerComposePath string) error {
	data, err := ioutil.ReadFile(dockerComposePath)
	if err != nil {
		return fmt.Errorf("Could not read the Docker Compose file %s\n%v\n", dockerComposePath, err)
	}
	err = ioutil.WriteFile(dockerComposeFlowPath, data, 0644)
	if err != nil {
		return err
	}
	return nil
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

func scaleDockerComposeTargets(host, project, target, scale string) {
	fmt.Println(scale)
	args := getDockerComposeArgs(host, project)
	args = append(args, "scale")
	// TODO: Change xxx to serviceName
	args = append(args, fmt.Sprintf("%s=%d", target, getConsulScale("xxx", scale)))
	fmt.Println(fmt.Sprintf("%s=%d", target, getConsulScale("xxx", scale)))
	runDockerComposeCmd(args)
}

func rmDockerComposeTargets(host, project string, targets []string) {
	args := getDockerComposeArgs(host, project)
	stopArgs := append(args, "stop")
	stopArgs = append(stopArgs, targets...)
	runDockerComposeCmd(stopArgs)
	rmArgs := append(args, "rm", "-f")
	rmArgs = append(rmArgs, targets...)
	runDockerComposeCmd(rmArgs)
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
