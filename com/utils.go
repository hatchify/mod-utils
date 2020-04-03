package com

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCmd executes a shell command at the file's path
func (file *FileWrapper) RunCmd(args ...string) (err error) {
	name := args[0]
	params := args[1:]

	tag := name + " " + strings.Join(params, " ")
	file.Debug(tag)

	cmd := exec.Command(name, params...)
	cmd.Dir = file.Path
	if err = cmd.Run(); err != nil {
		return file.handleError(tag, err)
	}

	return
}

// CmdOutput returns output of a shell command at the file's path
func (file *FileWrapper) CmdOutput(args ...string) (output string, err error) {
	name := args[0]
	params := args[1:]

	tag := name + " " + strings.Join(params, " ")
	file.Debug(tag)

	cmd := exec.Command(name, params...)
	cmd.Dir = file.Path
	stdout, err := cmd.Output()
	if err != nil {
		err = file.handleError(tag, err)
		return
	}

	output = strings.TrimSpace(string(stdout))
	return
}

func (file *FileWrapper) handleError(command string, ierr error) (err error) {
	return fmt.Errorf("Error running command `" + command + "` - " + ierr.Error())
}
