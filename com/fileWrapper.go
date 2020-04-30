package com

import (
	"io/ioutil"
	"path"
	"strings"
)

// FileWrapper represents a file object in a double link list, also contains status update info
type FileWrapper struct {
	// Private cached values
	absPath string
	goURL   string

	// Relative or absolute path to file from working dir
	Path string

	// Optional value to set or match
	Version string

	// Status flags
	Updated       bool
	Tagged        bool
	Committed     bool
	PROpened      bool
	BranchCreated bool
	TestFailed    bool
}

// Debug prints a message to stdout if debug is true
func (file *FileWrapper) Error(message string) {
	var label = file.goURL
	if file.goURL == "" {
		label = file.Path
	}

	Errorln(label, ":ERROR:", message)
}

// Output prints a message to stdout
func (file *FileWrapper) Output(message string) {
	var label = file.goURL
	if file.goURL == "" {
		label = file.Path
	}

	Println(label, "::", message)
}

// Debug prints a message to stdout if debug is true
func (file *FileWrapper) Debug(message string) {
	var label = file.goURL
	if file.goURL == "" {
		label = file.Path
	}

	Debugln(label, ":DEBUG:", message)
}

func (file *FileWrapper) containedIn(modfileContent string) bool {
	return strings.Contains(modfileContent, file.GetGoURL()+" v")
}

// AbsPath returns the current absolute directory of the calling lib
func (file *FileWrapper) AbsPath() string {
	if len(file.absPath) == 0 {
		file.absPath, _ = file.CmdOutput("pwd")
	}

	return file.absPath
}

// GetGoURL will return the format of the dependency version <github.com/hatchify/mod-utils>
func (file *FileWrapper) GetGoURL() string {
	if len(file.goURL) > 0 {
		return file.goURL
	}

	dir := file.AbsPath()

	// Parse go/src out of absolute path
	components := strings.Split(dir, path.Join("go", "src"))

	if len(components) != 2 {
		// We have a problem.. No go url found
		file.goURL = file.Path
		return file.Path
	}

	file.goURL = strings.Trim(components[1], "/")
	return file.goURL
}

// DirectlyImports is used to determine direct dependencies.
// returns true if file/go.mod contains any dep version
func (file *FileWrapper) DirectlyImports(dep *FileWrapper) bool {
	// Read library/go.mod
	if libMod, err := ioutil.ReadFile(path.Join(file.Path, "go.mod")); err == nil {
		return dep.containedIn(string(libMod))
	}

	return false
}

// DirectlyImportsAny returns true if file depends on any of the filter deps. Returns false if slice is empty
func (file *FileWrapper) DirectlyImportsAny(deps []*FileWrapper) bool {
	// Read library/go.sum once
	if libMod, err := ioutil.ReadFile(path.Join(file.Path, "go.mod")); err == nil {
		// Parse sum once
		goMod := string(libMod)

		// Check each dep in parsed sum
		for _, dep := range deps {
			if dep.containedIn(goMod) {
				// This lib is necessary
				return true
			}
		}
	}

	return false
}

// DependsOn is used to determine sort order.
// returns true if file/go.sum contains any dep version
func (file *FileWrapper) DependsOn(dep *FileWrapper) bool {
	// Read library/go.sum
	if libSum, err := ioutil.ReadFile(path.Join(file.Path, "go.sum")); err == nil {
		return dep.containedIn(string(libSum))
	}

	return false
}

// DependsOnAny returns true if file depends on any of the filter deps. Returns false if slice is empty
func (file *FileWrapper) DependsOnAny(deps []*FileWrapper) bool {
	// Read library/go.sum once
	if libSum, err := ioutil.ReadFile(path.Join(file.Path, "go.sum")); err == nil {
		// Parse sum once
		goSum := string(libSum)

		// Check each dep in parsed sum
		for _, dep := range deps {
			if dep.containedIn(goSum) {
				// This lib is necessary
				return true
			}
		}
	}

	return false
}

// MatchesAny returns true if file matches one of the deps
func (file *FileWrapper) MatchesAny(deps []*FileWrapper) bool {
	for _, dep := range deps {
		if strings.HasSuffix(file.GetGoURL(), dep.GetGoURL()) {
			file.Version = dep.Version
			return true
		}
	}

	return false
}
