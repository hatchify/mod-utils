package sync

import (
	"fmt"
	"os/exec"
	"strings"
)

func runCmd(dir, tag, name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return handleError(tag, err)
	}

	return
}

func cmdOutput(dir, tag, name string, args ...string) (output string, err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		err = handleError(tag, err)
		return
	}

	output = strings.TrimSpace(string(stdout))

	return
}

func handleError(command string, ierr error) (err error) {
	return fmt.Errorf("error running \"%s\": %v", command, ierr)
}
