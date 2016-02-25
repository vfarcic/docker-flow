package main
// TODO: Switch to methods
// TODO: Test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const dockerComposeFlowPath  = "docker-compose-flow.yml.tmp"
var readFile = ioutil.ReadFile
var writeFile = ioutil.WriteFile
var removeFile = os.Remove
var execCmd = exec.Command

type DockerCompose interface {
	CreateFlow(dcPath, dfPath, target, color string, blueGreen bool) error
	RemoveFlow() error
	PullTargets(host, project string, targets []string) error
	UpTargets(host, project string, targets []string) error
	ScaleTargets(host, project, target string, scale int) error
	RmTargets(host, project string, targets []string) error
	StopTargets(host, project string, targets []string) error
}

type DockerComposeImpl struct{}

func (dc DockerComposeImpl) CreateFlow(dcPath, dfPath, target, color string, blueGreen bool) error {
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

func (dc DockerComposeImpl) RemoveFlow() error {
	if err := removeFile(dockerComposeFlowPath); err != nil {
		return fmt.Errorf("Could not remove the temp file %s\n%v", dockerComposeFlowPath, err)
	}
	return nil
}

// TODO: Test
func (dc DockerComposeImpl) PullTargets(host, project string, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	args := dc.getArgs(host, project)
	args = append(args, "pull")
	args = append(args, targets...)
	return dc.runCmd(args)
}

// TODO: Test
func (dc DockerComposeImpl) UpTargets(host, project string, targets []string) error {
	args := dc.getArgs(host, project)
	args = append(args, "up", "-d")
	args = append(args, targets...)
	return dc.runCmd(args)
}

// TODO: Test
func (dc DockerComposeImpl) ScaleTargets(host, project, target string, scale int) error {
	args := dc.getArgs(host, project)
	args = append(args, "scale")
	args = append(args, fmt.Sprintf("%s=%d", target, scale))
	if err := dc.runCmd(args); err != nil {
		return err
	}
	return nil
}

// TODO: Test
func (dc DockerComposeImpl) RmTargets(host, project string, targets []string) error {
	if err := dc.StopTargets(host, project, targets); err != nil {
		return err
	}
	args := dc.getArgs(host, project)
	args = append(args, "rm", "-f")
	args = append(args, targets...)
	return dc.runCmd(args)
}

// TODO: Test
func (dc DockerComposeImpl) StopTargets(host, project string, targets []string) error {
	args := dc.getArgs(host, project)
	args = append(args, "stop")
	args = append(args, targets...)
	return dc.runCmd(args)
}

func (dc DockerComposeImpl) getArgs(host, project string) []string {
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

func (dc DockerComposeImpl) runCmd(args []string) error {
	cmd := execCmd("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose command %v\n%v", cmd, err)
	}
	return nil
}
