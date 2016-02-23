package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const dockerComposeFlowPath  = "docker-compose-flow.yml.tmp"

func createDockerComposeFlow(dockerComposePath, target, color string, blueGreen bool) error {
	data, err := ioutil.ReadFile(dockerComposePath)
	if err != nil {
		return fmt.Errorf("Could not read the Docker Compose file %s\n%v", dockerComposePath, err)
	}
	s := string(data)
	if blueGreen {
		old := fmt.Sprintf("%s:", target)
		new := fmt.Sprintf("%s-%s:", target, color)
		s = strings.Replace(string(data), old, new, 1)
	}
	err = ioutil.WriteFile(dockerComposeFlowPath, []byte(s), 0644)
	if err != nil {
		return fmt.Errorf("Could not write the Docker Flow file %s\n%v", dockerComposeFlowPath, err)
	}
	return nil
}

func removeDockerComposeFlow() error {
	if err := os.Remove(dockerComposeFlowPath); err != nil {
		return fmt.Errorf("Could not remove the temp file %s\n%v", dockerComposeFlowPath, err)
	}
	return nil
}

func pullDockerComposeTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := getDockerComposeArgs(host, project)
	args = append(args, "pull")
	args = append(args, targets...)
	return runDockerComposeCmd(args)
}

func upDockerComposeTargets(host, project string, targets []string) error {
	args := getDockerComposeArgs(host, project)
	args = append(args, "up", "-d")
	args = append(args, targets...)
	return runDockerComposeCmd(args)
}

func scaleDockerComposeTargets(host, consulAddress, project, target, serviceName, scale string) error {
	args := getDockerComposeArgs(host, project)
	args = append(args, "scale")
	s, err := getConsulScale(consulAddress, serviceName, scale)
	if err != nil {
		return err
	}
	args = append(args, fmt.Sprintf("%s=%d", target, s))
	if err := runDockerComposeCmd(args); err != nil {
		return err
	}
	return putConsulScale(consulAddress, serviceName, s)
}

func rmDockerComposeTargets(host, project string, targets []string) error {
	if err := stopDockerComposeTargets(host, project, targets); err != nil {
		return err
	}
	args := getDockerComposeArgs(host, project)
	args = append(args, "rm", "-f")
	args = append(args, targets...)
	return runDockerComposeCmd(args)
}

func stopDockerComposeTargets(host, project string, targets []string) error {
	args := getDockerComposeArgs(host, project)
	args = append(args, "stop")
	args = append(args, targets...)
	return runDockerComposeCmd(args)
}

func getDockerComposeArgs(host, project string) []string {
	args := []string{"-f", dockerComposeFlowPath}
	if (len(host) > 0) {
		env := os.Environ()
		env = append(env, fmt.Sprintf("DOCKER_HOST=%s", host))
	}
	if (len(project) > 0) {
		args = append(args, "-p", project)
	}
	return args
}

func runDockerComposeCmd(args []string) error {
	cmd := exec.Command("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose command %v\n%v", cmd, err)
	}
	return nil
}
