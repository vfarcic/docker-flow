package dockerflow

import (
	"io/ioutil"
	"os"
	"os/exec"
)

var readFile = ioutil.ReadFile
var writeFile = ioutil.WriteFile
var removeFile = os.Remove
var execCmd = exec.Command
