package com

import (
	"io/ioutil"
	"path"
	"strings"
)

// FileWrapper represents a file object in a double link list, also contains status update info
type FileWrapper struct {
	Path string

	Version string

	Updated  bool
	Tagged   bool
	Deployed bool
	PROpened bool
}

// Debug prints a message to stdout if debug is true
func (file FileWrapper) Error(message string) {
	Errorln(file.Path, ":ERROR:", message)
}

// Output prints a message to stdout
func (file FileWrapper) Output(message string) {
	Println(file.Path, "::", message)
}

// Debug prints a message to stdout if debug is true
func (file FileWrapper) Debug(message string) {
	Debugln(file.Path, ":DEBUG:", message)
}

// AbsPath returns the current absolute directory of the calling lib
func (file FileWrapper) AbsPath() string {
	dir, _ := file.CmdOutput("pwd")
	return dir
}

// GetGoURL will return the format of the dependency version <github.com/hatchify/mod-common>
func (file FileWrapper) GetGoURL() (url string) {
	dir := file.AbsPath()

	// Parse go/src out of absolute path
	components := strings.Split(dir, path.Join("go", "src"))

	if len(components) != 2 {
		// We have a problem.. No go url found
		file.Output("Unable to parse go url from dir <" + dir + "> defaulting to file path <" + file.Path + ">")

		return file.Path
	}

	url = strings.Trim(components[1], "/")
	return
}

// ImportsDirectly is used to determine direct dependencies.
// returns true if file/go.mod contains any dep version
func (file FileWrapper) ImportsDirectly(dep *FileWrapper) bool {
	// Read library/go.mod
	if libMod, err := ioutil.ReadFile(path.Join(file.Path, "go.mod")); err == nil {
		return strings.Contains(string(libMod), dep.Path+" v")
	}

	return false
}

// DependsOn is used to determine sort order.
// returns true if file/go.sum contains any dep version
func (file FileWrapper) DependsOn(dep *FileWrapper) bool {
	// Read library/go.sum
	if libSum, err := ioutil.ReadFile(path.Join(file.Path, "go.sum")); err == nil {
		return strings.Contains(string(libSum), dep.Path+" v")
	}

	return false
}

// DependsOnAny returns true if file depends on any of the filter deps. Returns false if slice is empty
func (file FileWrapper) DependsOnAny(deps []*FileWrapper) bool {
	// Read library/go.sum once
	if libSum, err := ioutil.ReadFile(path.Join(file.Path, "go.sum")); err == nil {
		// Parse sum once
		goSum := string(libSum)

		// Check each dep in parsed sum
		for i := range deps {
			if strings.Contains(goSum, deps[i].Path+" v") {
				// This lib is necessary
				return true
			}
		}
	}

	return false
}

// MatchesAny returns true if file matches one of the deps
func (file FileWrapper) MatchesAny(deps []*FileWrapper) bool {
	for i := range deps {
		if strings.HasSuffix(file.Path, deps[i].Path) {
			return true
		}
	}

	return false
}
