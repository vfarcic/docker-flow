package main

import (
	"fmt"
	"os"
	"strings"
)

const dockerComposeFlowPath  = "docker-compose-flow.yml.tmp"

var dockerCompose DockerComposable = DockerCompose{}
var getDockerCompose = func() DockerComposable {
	return dockerCompose
}

type DockerComposable interface {
	CreateFlowFile(dcPath, dfPath, target, color string, blueGreen bool) error
	RemoveFlow() error
	PullTargets(host, project string, targets []string) error
	UpTargets(host, project string, targets []string) error
	ScaleTargets(host, project, target string, scale int) error
	RmTargets(host, project string, targets []string) error
	StopTargets(host, project string, targets []string) error
}

type DockerCompose struct{}

func (dc DockerCompose) CreateFlowFile(dcPath, dfPath, target, color string, blueGreen bool) error {
	data, err := readFile(dcPath)
	if err != nil {
		return fmt.Errorf("Could not read the Docker Compose file %s\n%v", dcPath, err)
	}
	s := string(data)
	if blueGreen {
		old := fmt.Sprintf("%s:", target)
		new := fmt.Sprintf("%s-%s:", target, color)
		s = strings.Replace(string(data), old, new, 1)
	}
	err = writeFile(dfPath, []byte(s), 0644)
	if err != nil {
		return fmt.Errorf("Could not write the Docker Flow file %s\n%v", dockerComposeFlowPath, err)
	}
	return nil
}

func (dc DockerCompose) RemoveFlow() error {
	if err := removeFile(dockerComposeFlowPath); err != nil {
		return fmt.Errorf("Could not remove the temp file %s\n%v", dockerComposeFlowPath, err)
	}
	return nil
}

func (dc DockerCompose) PullTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"pull"}, targets...)
	return dc.runCmd(host, project, args)
}

func (dc DockerCompose) UpTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"up", "-d"}, targets...)
	return dc.runCmd(host, project, args)
}

func (dc DockerCompose) ScaleTargets(host, project, target string, scale int) error {
	if len(target) == 0 {
		return nil
	}
	args := []string{"scale", fmt.Sprintf("%s=%d", target, scale)}
	return dc.runCmd(host, project, args)
}

func (dc DockerCompose) RmTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"rm", "-f"}, targets...)
	return dc.runCmd(host, project, args)
}

func (dc DockerCompose) StopTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"stop"}, targets...)
	return dc.runCmd(host, project, args)
}

func (dc DockerCompose) getArgs(host, project string) []string {
	args := []string{"-f", dockerComposeFlowPath}
	SetDockerHost(host)
	if (len(project) > 0) {
		args = append(args, "-p", project)
	}
	return args
}

func (dc DockerCompose) runCmd(host, project string, args []string) error {
	args = append(dc.getArgs(host, project), args...)
	cmd := execCmd("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose command %v\n%v", cmd, err)
	}
	return nil
}
