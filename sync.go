package sync

import (
	"os/exec"
	"strings"

	common "github.com/hatchify/mod-common"
	sort "github.com/hatchify/mod-sort"
)

// Library represents a file and its updated deps
type Library struct {
	File *common.FileWrapper

	updatedDeps *sort.FileNode
}

// Update will refresh the current dir to master, reset mod files and push changes if there are any
func (lib Library) Update(commitMessage string) (err error) {
	lib.File.Output("Checking out master...")
	if err = lib.File.CheckoutBranch("master"); err != nil {
		lib.File.Output("Checkout failed :(")
		return
	}

	if err = lib.File.Fetch(); err != nil {
		lib.File.Output("Fetch failed :(")
		return
	}

	if err = lib.File.Pull(); err != nil {
		lib.File.Output("Pull failed :(")
		return
	}

	lib.File.Output("Checking deps...")
	if err = lib.File.RunCmd("remove go.mod", "rm", "go.mod"); err != nil {
		lib.File.Output("No mod file found. Skipping.")
		return
	}

	if err = lib.File.RunCmd("remove go.sum", "rm", "go.sum"); err != nil {
		lib.File.Output("No sum file found.")
	}

	if err = lib.File.RunCmd("go mod init", "go", "mod", "init"); err != nil {
		lib.File.Output("Mod init failed :(")
		return
	}

	if err = lib.File.RunCmd("go mod tidy", "go", "mod", "tidy"); err != nil {
		lib.File.Output("Mod init failed :(")
		return
	}

	if err = lib.File.RunCmd("git add", "git", "add", "go.*"); err != nil {
		lib.File.Output("Git add failed :(")
		return
	}

	if err = lib.File.RunCmd("git commit", "git", "commit", "-m", commitMessage); err != nil {
		lib.File.Output("Deps up to date!")
		return
	}

	lib.File.Output("Updating mod files...")
	if err = lib.File.RunCmd("git push", "git", "push"); err != nil {
		lib.File.Output("Update failed :(")
		return
	}

	lib.File.Output("Deps updated!")
	return
}

// TagLib updates the lib to the provided tag, or increments if git-tagger is able to
func (lib Library) TagLib(tag string) (newTag string) {
	if len(tag) == 0 {
		// Use git-tagger to increment
		lib.File.Output("Updating tag...")

		if lib.File.RunCmd("git-tagger", "git-tagger") != nil {
			lib.File.Output("Unable to increment tag.")
			return
		}

		newTag = lib.GetCurrentTag()
		lib.File.Output("Incremented tag - " + newTag)

	} else {
		lib.File.Output("Setting tag...")

		// Set tag manually
		if lib.File.RunCmd("git tag", "git", "tag", tag) != nil {
			lib.File.Output("Unable to set tag.")
			return
		}

		// Push new tag
		if lib.File.RunCmd("git push tag", "git", "push", "--tag") != nil {
			lib.File.Output("Unable to push tag.")
			return
		}

		newTag = tag
		lib.File.Output("Set Tag - " + newTag)
	}

	return
}

// ShouldTag returns true if not a plugin and has a tag that is out of date
func (lib Library) ShouldTag() (shouldTag bool) {
	if strings.HasSuffix(strings.Trim(lib.File.Path, "/"), "-plugin") {
		lib.File.Output("Not tagging plugins. Skipping tag.")
		return
	}

	// Check if tag is up to date
	cmd := exec.Command("git-tagger", "--action=get")
	cmd.Dir = lib.File.Path
	stdout, err := cmd.Output()
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("No tag set. Skipping tag.")
		return
	}
	tag := strings.TrimSpace(string(stdout))

	cmd = exec.Command("git", "rev-list", "-n", "1", tag)
	cmd.Dir = lib.File.Path
	stdout, err = cmd.Output()
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("No revision history. Skipping tag.")
		return
	}
	tagCommit := string(stdout)

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = lib.File.Path
	stdout, err = cmd.Output()
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("No revision head. Skipping tag.")
		return
	}
	headCommit := string(stdout)

	if tagCommit != headCommit {
		// Tag out of date
		lib.File.Output("Tag outdated...")
		return true
	}

	lib.File.Output("Tags up to date!")
	return
}

// GetCurrentTag returns the latest tag for a given dir
func (lib Library) GetCurrentTag() (currentTag string) {
	output, err := lib.File.CmdOutput("git-tagger", "git-tagger", "--action=get")
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("Unable to update tag.")
		return
	}

	return output
}
