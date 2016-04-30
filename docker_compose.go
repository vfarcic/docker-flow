package main

import (
	"fmt"
	"os"
	"strings"
)

const dockerComposeFlowPath = "docker-compose-flow.yml.tmp"

var dockerCompose DockerComposable = DockerCompose{}
var getDockerCompose = func() DockerComposable {
	return dockerCompose
}

type DockerComposable interface {
	CreateFlowFile(dcPath []string, serviceName, target string, sideTargets []string, color string, blueGreen bool) error
	RemoveFlow() error
	PullTargets(host, certPath, project string, targets []string) error
	UpTargets(host, certPath, project string, targets []string) error
	ScaleTargets(host, certPath, project, target string, scale int) error
	RmTargets(host, certPath, project string, targets []string) error
	StopTargets(host, certPath, project string, targets []string) error
}

type DockerCompose struct{}

func (dc DockerCompose) CreateFlowFile(dcPath []string, serviceName, target string, sideTargets []string, color string, blueGreen bool) error {
	data, err := readFile(dcPath[0])
	if err != nil {
		return fmt.Errorf("Could not read the Docker Compose file %s\n%v", dcPath, err)
	}
	extendedTarget := target
	if blueGreen {
		extendedTarget = fmt.Sprintf("%s-%s", target, color)
	}
	s := ""
	dcData := strings.Trim(string(data), " ")
	firstLine := strings.Split(dcData, "\n")[0]
	indent := ""
	dcTemplate := `
%s%s:
%s  extends:
%s    file: %s
%s    service: %s`
	dcTemplateTarget := dcTemplate + `
%s  environment:
%s    - SERVICE_NAME=%s-%s`
	if strings.Contains(strings.ToLower(firstLine), "version") && strings.Contains(firstLine, "2") {
		indent = "  "
		s = `version: '2'

services:`
	}
	s += fmt.Sprintf(
		dcTemplateTarget,
		indent,
		extendedTarget,
		indent,
		indent,
		dcPath[0],
		indent,
		target,
		indent,
		indent,
		serviceName,
		color,
	)
	for _, sideTarget := range sideTargets {
		s += fmt.Sprintf(
			dcTemplate,
			indent,
			sideTarget,
			indent,
			indent,
			dcPath[0],
			indent,
			sideTarget,
		)
	}
	err = writeFile(dockerComposeFlowPath, []byte(strings.Trim(s, "\n")), 0644)
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

func (dc DockerCompose) PullTargets(host, certPath, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"pull"}, targets...)
	return dc.runCmd(host, certPath, project, args)
}

func (dc DockerCompose) UpTargets(host, certPath, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"up", "-d"}, targets...)
	return dc.runCmd(host, certPath, project, args)
}

func (dc DockerCompose) ScaleTargets(host, certPath, project, target string, scale int) error {
	if len(target) == 0 {
		return nil
	}
	args := []string{"scale", fmt.Sprintf("%s=%d", target, scale)}
	return dc.runCmd(host, certPath, project, args)
}

func (dc DockerCompose) RmTargets(host, certPath, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"rm", "-f"}, targets...)
	return dc.runCmd(host, certPath, project, args)
}

func (dc DockerCompose) StopTargets(host, certPath, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := append([]string{"stop"}, targets...)
	return dc.runCmd(host, certPath, project, args)
}

func (dc DockerCompose) getArgs(host, certPath, project string) []string {
	args := []string{"-f", dockerComposeFlowPath}
	SetDockerHost(host, certPath)
	if len(project) > 0 {
		args = append(args, "-p", project)
	}
	return args
}

func (dc DockerCompose) runCmd(host, certPath, project string, args []string) error {
	args = append(dc.getArgs(host, certPath, project), args...)
	cmd := execCmd("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose command %v\n%v", cmd, err)
	}
	return nil
}
