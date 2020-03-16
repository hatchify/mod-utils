package main

import (
	"fmt"
	"os/exec"
)

func runCmd(dir, tag, name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return handleError(tag, err)
	}

	return
}

func handleError(command string, ierr error) (err error) {
	return fmt.Errorf("error running \"%s\": %v", command, ierr)
}
