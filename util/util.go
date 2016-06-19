package util

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

var ReadFile = ioutil.ReadFile
var WriteFile = ioutil.WriteFile
var RemoveFile = os.Remove
var ExecCmd = exec.Command
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
var RunCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}
var Sleep = func(d time.Duration) {
	time.Sleep(d)
}
