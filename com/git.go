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
		err = nil

		// Attempt to create branch
		if err = file.RunCmd("git", "checkout", "-b", branch); err == nil {
			// Success
			file.BranchCreated = true
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

// AddSecret will set a secret for the repository
func (file *FileWrapper) AddSecret(name, secret string) (err error) {
	var comps = strings.Split(file.GetGoURL(), "/")

	switch comps[0] {
	case "github.com":
		// Supported
	default:
		// Not Supported
		err = fmt.Errorf("%s currently not supported for secrets", comps[0])
		return
	}

	var apiURL = "https://api." + comps[0]
	var resource = "/repos/" + strings.Join(comps[1:], "/") + "/actions/secrets/" + name

	var u *url.URL
	if u, err = url.ParseRequestURI(apiURL); err != nil {
		err = fmt.Errorf("Unable to parse url %s", apiURL)
		return
	}

	// Get auth token
	authObject, err := getAuth()
	if err != nil {
		// Get new creds
		return fmt.Errorf("needs github credentials for PR")
	}

	file.Output("Getting encryption key...")
	id, key, err := authObject.GetPublicKey(file.GetGoURL())
	if err != nil {
		file.Output("Error getting public key :(")
		err = fmt.Errorf("Unable get github public key")
		return
	}

	file.Output("Encrypting secret...")
	encrypted, err := authObject.Encrypt(secret, key)
	if err != nil {
		err = fmt.Errorf("Unable to encrypt secret")
		return
	}

	post := &secretRequest{Encrypted: encrypted, KeyID: id}
	data, err := json.Marshal(post)
	if err != nil {
		err = fmt.Errorf("Unable to parse secret request")
		return
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
	file.Output("Setting repository secret...")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	// Return status
	if resp.StatusCode == 200 {
		file.Output("Successfully set repository secret!")
	} else {
		err = fmt.Errorf("Http error %d", resp.StatusCode)
	}

	return
}

// AddGitWorkflow will set an example yml file for the repo
func (file *FileWrapper) AddGitWorkflow(exampleYmlPath string) (err error) {
	// Get source dir and template
	sourceDir, ymlTemplate := path.Split(exampleYmlPath)
	ymlSource := &FileWrapper{Path: sourceDir}
	templateSouce := path.Join(ymlSource.AbsPath(), ymlTemplate)

	// Ignore auto tag for un-tagged libs
	if ymlTemplate == "auto-tag.yml" {
		if file.RunCmd("git-tagger", "--action=get") != nil {
			// No tag set. skip tag
			err = fmt.Errorf("No tag set... Skipping")
			return
		}
	}

	// Prep workflow dir
	workflowPath := path.Join(".github", "workflows")
	file.RunCmd("mkdir", ".github")
	file.RunCmd("mkdir", workflowPath)
	newWorkflow := path.Join(workflowPath, ymlTemplate)

	file.Output("Copying " + exampleYmlPath + " to " + workflowPath + "...")
	// Copy example yml file to workflow dir
	if file.RunCmd("cp", templateSouce, newWorkflow) != nil {
		err = fmt.Errorf("Unable to copy %s to %s", exampleYmlPath, workflowPath)
		return
	}

	if file.Add(path.Join(workflowPath, ymlTemplate)) != nil {
		return fmt.Errorf("Unable to add workflow path")
	}

	if file.Commit("Added workflow: "+ymlTemplate) != nil {
		return fmt.Errorf("Unable to commit workflow")
	}

	if file.Push() != nil {
		return fmt.Errorf("Unable to push workflow changes")
	}

	file.Output("Workflow added successfully!")
	return
}

// PullRequest opens a PR for the specified url on the specified branch
func (file *FileWrapper) PullRequest(title, message, branch, target string) (status *PRResponse, err error) {
	if branch == target {
		err = fmt.Errorf("Cannot create PR from " + branch + " to " + target)
		return
	}

	if err = file.RunCmd("git", "push", "-u", "origin", branch); err != nil {
		err = fmt.Errorf("Unable to set upstream for branch " + branch + " :( Check repo permissions?")
		return
	}

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

	var u *url.URL
	if u, err = url.ParseRequestURI(apiURL); err != nil {
		err = fmt.Errorf("Unable to parse url %s", apiURL)
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
	authObject, err := getAuth()
	if err != nil {
		err = fmt.Errorf("needs github credentials for PR")
		return
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
	resp.Body.Close()

	payload := &PRResponse{}
	err = json.Unmarshal(body, payload)

	// Return status
	payload.HTTPStatus = resp.StatusCode
	status = payload
	if status.HTTPStatus >= 300 {
		err = fmt.Errorf("Http error %d", status.HTTPStatus)
		if len(status.Errors) > 0 {
			file.Output(fmt.Sprintf("Http Error %d: %s", status.HTTPStatus, status.Errors[0].Message))
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
