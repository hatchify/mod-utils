package sync

import (
	"fmt"
	"os/exec"
	"strings"
)

// SyncDeps will refresh the current dir to master, reset mod files and push changes if there are any
func SyncDeps(dir, commitMessage string) (err error) {
	fmt.Println(dir + ": Checking out master...")
	if err = runCmd(dir, "git checkout", "git", "checkout", "master"); err != nil {
		fmt.Println(dir + ": Checkout failed :(")
		return
	}

	if err = runCmd(dir, "git fetch", "git", "fetch", "origin", "--prune", "--prune-tags", "--tags"); err != nil {
		fmt.Println(dir + ": Fetch failed :(")
		return
	}

	if err = runCmd(dir, "git pull", "git", "pull"); err != nil {
		fmt.Println(dir + ": Pull failed :(")
		return
	}

	fmt.Println(dir + ": Checking deps...")
	if err = runCmd(dir, "remove go.mod", "rm", "go.mod"); err != nil {
		fmt.Println(dir + ": No mod file found. Skipping.")
		return
	}

	if err = runCmd(dir, "remove go.sum", "rm", "go.sum"); err != nil {
		fmt.Println(dir + ": No sum file found.")
	}

	if err = runCmd(dir, "go mod init", "go", "mod", "init"); err != nil {
		fmt.Println(dir + ": Mod init failed :(")
		return
	}

	if err = runCmd(dir, "go mod tidy", "go", "mod", "tidy"); err != nil {
		fmt.Println(dir + ": Mod init failed :(")
		return
	}

	if err = runCmd(dir, "git add", "git", "add", "go.*"); err != nil {
		fmt.Println(dir + ": Git add failed :(")
		return
	}

	if err = runCmd(dir, "git commit", "git", "commit", "-m", commitMessage); err != nil {
		fmt.Println(dir + ": Deps up to date!")
		return
	}

	fmt.Println(dir + ": Updating mod files...")
	if err = runCmd(dir, "git push", "git", "push"); err != nil {
		fmt.Println(dir + ": Update failed :(")
		return
	}

	fmt.Println(dir + ": Deps updated!")
	return
}

// TagLib updates the lib to the provided tag, or increments if git-tagger is able to
func TagLib(dir, tag string) (newTag string) {
	if len(tag) == 0 {
		// Use git-tagger to increment
		fmt.Println(dir + ": Updating tag...")

		if runCmd(dir, "git-tagger", "git-tagger") != nil {
			fmt.Println(dir + ": Unable to increment tag.")
			return
		}

		newTag = GetCurrentTag(dir)
		fmt.Println(dir+": Updated tag -", newTag)

	} else {
		fmt.Println(dir + ": Setting tag...")

		// Set tag manually
		if runCmd(dir, "git tag", "git", "tag", tag) != nil {
			fmt.Println(dir + ": Unable to set tag.")
			return
		}

		// Push new tag
		if runCmd(dir, "git push tag", "git", "push", "--tag") != nil {
			fmt.Println(dir + ": Unable to push tag.")
			return
		}

		newTag = tag
		fmt.Println(dir+": Set Tag -", newTag)
	}

	return
}

// ShouldTag returns true if not a plugin and has a tag that is out of date
func ShouldTag(dir string) (shouldTag bool) {
	if strings.HasSuffix(strings.Trim(dir, "/"), "-plugin") {
		fmt.Println(dir + ": Not tagging plugins. Skipping tag.")
		return
	}

	// Check if tag is up to date
	cmd := exec.Command("git-tagger", "--action=get")
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		// No tag set. skip tag
		fmt.Println(dir + ": No tag set. Skipping tag.")
		return
	}
	tag := strings.TrimSpace(string(stdout))

	cmd = exec.Command("git", "rev-list", "-n", "1", tag)
	cmd.Dir = dir
	stdout, err = cmd.Output()
	if err != nil {
		// No tag set. skip tag
		fmt.Println(dir + ": No revision history. Skipping tag.")
		return
	}
	tagCommit := string(stdout)

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	stdout, err = cmd.Output()
	if err != nil {
		// No tag set. skip tag
		fmt.Println(dir + ": No revision head. Skipping tag.")
		return
	}
	headCommit := string(stdout)

	if tagCommit != headCommit {
		// Tag out of date
		fmt.Println(dir + ": Tag outdated...")
		return true
	}

	fmt.Println(dir + ": Tags up to date!")
	return
}

// GetCurrentTag returns the latest tag for a given dir
func GetCurrentTag(dir string) (currentTag string) {
	output, err := cmdOutput(dir, "git-tagger", "git-tagger", "--action=get")
	if err != nil {
		// No tag set. skip tag
		fmt.Println(dir + ": Unable to update tag.")
		return
	}

	return output
}
