package main

// Refresh will refresh the current dir
func Refresh(dir, commitMessage string) (err error) {
	if err = runCmd(dir, "git checkout", "git", "checkout", "master"); err != nil {
		return
	}

	if err = runCmd(dir, "git fetch", "git", "fetch", "origin", "--prune", "--prune-tags", "--tags"); err != nil {
		return
	}

	if err = runCmd(dir, "git pull", "git", "pull"); err != nil {
		return
	}

	if err = runCmd(dir, "remove gomod files", "rm", "go.mod", "go.sum"); err != nil {
		return
	}

	if err = runCmd(dir, "go mod init", "go", "mod", "init"); err != nil {
		return
	}

	if err = runCmd(dir, "go mod tidy", "go", "mod", "tidy"); err != nil {
		return
	}

	if err = runCmd(dir, "git add", "git", "add", "go.mod", "go.sum", "*.go"); err != nil {
		return
	}

	if err = runCmd(dir, "git commit", "git", "commit", "-m", commitMessage); err != nil {
		return
	}

	if err = runCmd(dir, "git push", "git", "push"); err != nil {
		return
	}

	return
}
