package main

import (
	"io/ioutil"
	"os"
	"os/exec"
)

var readFile = ioutil.ReadFile
var writeFile = ioutil.WriteFile
var removeFile = os.Remove
var execCmd = exec.Command
var SetDockerHost = func(host string) {
	if (len(host) > 0) {
		os.Setenv("DOCKER_HOST", host)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
}
var runCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}