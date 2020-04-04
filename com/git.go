package com

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"path"
	"strings"

	"github.com/hatchify/simply"
)

// CheckoutBranch calls git checkout on provided branch in provided dir. Creates new branch if necessary
func (file *FileWrapper) CheckoutBranch(branch string) (err error) {
	return file.RunCmd("git", "checkout", branch)
}

// CheckoutOrCreateBranch calls git checkout on provided branch in provided dir. Creates new branch if necessary
func (file *FileWrapper) CheckoutOrCreateBranch(branch string) (switched, created bool, err error) {
	if len(branch) == 0 {
		// Not switched, not created, no error
		return
	}

	// Attempt checkout branch
	if err = file.RunCmd("git", "checkout", branch); err != nil {
		// Attempt to create branch
		err = nil

		if err = file.RunCmd("git", "checkout", "-b", branch); err == nil {
			// Success
			created = true
			switched = true
		}
	} else {
		// Switch succeeded
		switched = true
	}

	return
}

// Fetch calls git fetch in provided dir
func (file *FileWrapper) Fetch() (err error) {
	file.RunCmd("git", "fetch")
	return file.RunCmd("git", "fetch", "origin", "--prune", "--prune-tags", "--tags")
}

// Merge merges other branch into current branch
func (file *FileWrapper) Merge(otherBranch string) error {
	return file.RunCmd("git", "merge", otherBranch)
}

// Pull calls git pull in provided dir
func (file *FileWrapper) Pull() (err error) {
	return file.RunCmd("git", "pull")
}

// Push calls git push in provided dir
func (file *FileWrapper) Push() (err error) {
	return file.RunCmd("git", "push", "-u", "origin")
}

// Stash calls git stash in provided dir
func (file *FileWrapper) Stash() (err error) {
	return file.RunCmd("git", "stash")
}

// StashPop calls git stash pop in provided dir
func (file *FileWrapper) StashPop() (localChanges bool) {
	// Hide mod file changes to prevent stash pop issues
	file.RunCmd("mv", "go.mod", "go.mod.bak")
	file.RunCmd("mv", "go.sum", "go.sum.bak")

	// Pop
	file.RunCmd("git", "stash", "pop")

	// Hide mod file changes to prevent stash pop issues
	file.RunCmd("mv", "go.mod.bak", "go.mod")
	file.RunCmd("mv", "go.sum.bak", "go.sum")

	// Handle conflicts
	localChanges = file.HasChanges()

	return
}

// HasChanges is true if files are able to be committed
func (file *FileWrapper) HasChanges() bool {
	file.Add(".")

	if file.Commit("revert me") == nil {
		file.RunCmd("git", "reset", "HEAD~1")
		return true
	}

	return false
}

// Add calls git add on each filename proved in provided dir
func (file *FileWrapper) Add(filename ...string) (err error) {
	var args = []string{"git", "add"}
	args = append(args, filename...)

	return file.RunCmd(args...)
}

// Commit calls git commit with provided message provided in provided dir
func (file *FileWrapper) Commit(message string) (err error) {
	return file.RunCmd("git", "commit", "-m", message)
}

// Reset calls git reset with provided args in provieded in provided dir
func (file *FileWrapper) Reset(args ...string) (err error) {
	params := append([]string{"git", "reset"}, args...)
	return file.RunCmd(params...)
}

// CurrentBranch returns current branch for a given file or an error if it can't be determined
func (file *FileWrapper) CurrentBranch() (branch string, err error) {
	branch, err = file.CmdOutput("git", "branch", "--show-current")
	branch = strings.TrimSpace(branch)
	return
}

// PullRequest opens a PR for the specified url on the specified branch
func (file *FileWrapper) PullRequest(title, message, branch, target string) (status *PRResponse, err error) {
	if branch == target {
		err = fmt.Errorf("Cannot create PR from " + branch + " to " + target)
		return
	}

	file.RunCmd("git", "push", "-u", "origin", branch)

	// Get git host
	comps := strings.Split(file.GetGoURL(), "/")
	switch comps[0] {
	case "github.com":
		// Supported
	default:
		// Not Supported
		err = fmt.Errorf("%s currently not supported for pull requests", comps[0])
		return
	}

	// GitHub api url
	apiURL := "https://api." + comps[0]
	resource := "/repos/" + strings.Join(comps[1:], "/") + "/pulls"

	u, err := url.ParseRequestURI(apiURL)
	if err != nil {
		err = fmt.Errorf("Unable to url %s", apiURL)
		return
	}

	// Parse request
	if len(branch) == 0 {
		branch, err = file.CurrentBranch()
		if err != nil {
			err = fmt.Errorf("Unable to get current branch :(")
			return
		}
	}
	post := &prRequest{title, message, branch, target}
	data, err := json.Marshal(post)
	if err != nil {
		err = fmt.Errorf("Unable to parse pull request params")
		return
	}

	// Get auth token
	authObject, err := LoadAuth()
	if err != nil || len(authObject.User) == 0 || len(authObject.Token) == 0 {
		// Get new creds
		file.Output("Needs github credentials for PR...")
		if authObject.Setup() != nil {
			file.Output("Error saving :(")
			err = fmt.Errorf("Unable to parse github username and token")
			return
		}
		err = nil
		file.Output("Saved Credentials!")
	}

	// Make request
	u.Path = resource
	urlStr := u.String()
	client := &http.Client{}
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "token "+authObject.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// Execute Request
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	// Read response
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	payload := &PRResponse{}
	err = json.Unmarshal(body, payload)
	resp.Body.Close()

	// Return status
	payload.HTTPStatus = resp.StatusCode
	status = payload
	if status.HTTPStatus >= 300 {
		err = fmt.Errorf("Http error %d", status.HTTPStatus)
		if len(status.Errors) > 0 {
			file.Output(fmt.Sprintf("Http Error %d: %s", status.HTTPStatus, simply.Stringify(status)))
		}
	}

	if status.HTTPStatus == 401 {
		// Bad credentials.. clear file
		usr, _ := user.Current()
		file.RunCmd("rm", path.Join(usr.HomeDir, configName))
		file.Output("Bad credentials cleared.")
		// Try again
		file.PullRequest(title, message, branch, target)
	}

	return
}
