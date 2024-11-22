package util

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
)

// ExecCmd executes a system call and returns stdout, stderr and exit code of the execution.
func ExecCmd(cmdName string, cmdArgs []string) (stdout, stderr bytes.Buffer, ec int32, err error) {
	// Execute command and catch the stdout and stderr as byte buffer.
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = os.Environ()
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	ec = 0
	if err = cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			ec = int32(ws.ExitStatus())
		} else {
			ec = 1
		}
	}
	return
}
