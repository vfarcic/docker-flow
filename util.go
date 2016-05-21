package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

var readFile = ioutil.ReadFile
var writeFile = ioutil.WriteFile
var removeFile = os.Remove
var execCmd = exec.Command
var SetDockerHost = func(host, certPath string) {
	if len(host) > 0 {
		os.Setenv("DOCKER_HOST", host)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
	if len(certPath) > 0 {
		os.Setenv("DOCKER_CERT_PATH", certPath)
	} else {
		os.Unsetenv("DOCKER_CERT_PATH")
	}
}
var runCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}
var sleep = func(d time.Duration) {
	time.Sleep(d)
}
