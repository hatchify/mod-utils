package sync

import (
	"strings"
)

// TagLib updates the lib to the provided tag, or increments if git-tagger is able to
func (lib *Library) TagLib(tag string) (newTag string) {
	if len(tag) == 0 {
		lib.File.Output("Updating tag...")

		// Use git-tagger to increment
		if lib.File.RunCmd("git-tagger") != nil {
			lib.File.Output("Unable to increment tag.")
			return
		}

		newTag = lib.GetCurrentTag()
		lib.File.Output("Incremented tag - " + newTag)

	} else {
		lib.File.Output("Setting tag...")

		// Set tag manually
		if lib.File.RunCmd("git", "tag", tag) != nil {
			lib.File.Output("Unable to set tag.")
			return
		}

		// Push new tag
		if lib.File.RunCmd("git", "push", "--tag") != nil {
			lib.File.Output("Unable to push tag.")
			return
		}

		newTag = tag
		lib.File.Output("Set Tag - " + newTag)
	}

	return
}

// ShouldTag returns true if not a plugin and has a tag that is out of date
func (lib *Library) ShouldTag() (shouldTag bool) {
	if strings.HasSuffix(strings.Trim(lib.File.Path, "/"), "-plugin") {
		lib.File.Output("Not tagging plugins. Skipping tag.")
		return
	}

	// Check if tag is up to date
	stdout, err := lib.File.CmdOutput("git-tagger", "--action=get")
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("No tag set. Skipping tag.")
		return
	}
	tag := strings.TrimSpace(string(stdout))

	stdout, err = lib.File.CmdOutput("git", "rev-list", "-n", "1", tag)
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("No revision history. Skipping tag.")
		return
	}
	tagCommit := string(stdout)

	stdout, err = lib.File.CmdOutput("git", "rev-parse", "HEAD")
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

	lib.File.Output("Tags up to date @ " + tag + "!")
	return
}

// GetCurrentTag returns the latest tag for a given dir
func (lib *Library) GetCurrentTag() (currentTag string) {
	output, err := lib.File.CmdOutput("git-tagger", "--action=get")
	if err != nil {
		// No tag set. skip tag
		lib.File.Output("Unable to update tag.")
		return
	}

	return output
}
