package sync

import (
	"strings"

	sort "github.com/hatchify/mod-sort"
)

// ModInit calls go mod init on a given lib
func (lib *Library) ModInit() error {
	return lib.File.RunCmd("go", "mod", "init")
}

// ModTidy calls go mod tidy on a given lib
func (lib *Library) ModTidy() error {
	return lib.File.RunCmd("go", "mod", "tidy")
}

// ModClear calls rm go.* to remove go mod files on a given lib
// Returns true if go mod file was found
func (lib *Library) ModClear() (hasModFile, hasSumFile bool) {
	if lib.File.RunCmd("rm", "go.mod") == nil {
		hasModFile = true
	}

	if lib.File.RunCmd("rm", "go.sum") == nil {
		hasSumFile = true
	}

	return
}

// ModAddDeps adds a dep@version to go.mod to force-update or force-downgrade any deps in the filtered chain
func (lib *Library) ModAddDeps(listHead *sort.FileNode) {
	for itr := listHead; itr != nil && itr.File.Path != lib.File.Path; itr = itr.Next {
		if lib.File.ImportsDirectly(itr.File) {
			// Create new node to add to independent list on lib
			var node sort.FileNode
			node.File = itr.File
			lib.AddDep(&node)
		}
	}
}

// ModSetDeps adds a dep@version to go.mod to force-update or force-downgrade any deps in the filtered chain
func (lib *Library) ModSetDeps() {
	for itr := lib.updatedDeps; itr != nil; itr = itr.Next {
		if len(itr.File.Version) == 0 {
			lib.File.Output("Error: no version to set for " + itr.File.Path)

		} else {
			url := itr.File.GetGoURL()

			if lib.File.RunCmd("go", "get", url+"@"+itr.File.Version) == nil {
				if itr.File.Updated || itr.File.Tagged || itr.File.Deployed {
					lib.File.Output("Updated " + url + " @ " + itr.File.Version)
				}
			} else {
				lib.File.Output("Error: Failed to get " + url + " @ " + itr.File.Version)
			}
		}
	}
}

// ModDeploy will commit and push local changes to the current branch before switching to master
func (lib *Library) ModDeploy(tag string) (deployed bool) {
	// Handle saving local changes
	lib.File.StashPop()
	lib.File.Add(".")

	// Ignore changes to go mod files (prevents committing local replacements)
	lib.File.Reset("go.*")

	message := ""
	if len(tag) == 0 {
		version := lib.File.Version
		if len(version) == 0 && !strings.HasSuffix(strings.Trim(lib.File.Path, "/"), "-plugin") {
			version = lib.GetCurrentTag()
		}

		if len(version) == 0 {
			message = "GoMu: Deploy local changes"
			// Set old version of libs in case they weren't updated previously
			lib.File.Version = version

		} else {
			message = "GoMu: Deploy local changes before incrementing version from " + version
		}
	} else {
		message = "GoMu: Deploy local changes before updating version to " + tag
	}

	if lib.File.Commit(message) == nil {
		// Successful commit, push changes
		deployed = true
		lib.File.Output("Deploying local changes...")
		if lib.File.Push() == nil {
			lib.File.Output("Deploy Complete!")
		} else {
			lib.File.Output("Deploy Failed :(")
		}
	} else {
		lib.File.Output("No changes to deploy!")
	}

	// Re-stash local changes to go.mod
	lib.File.Stash()

	return
}

// ModUpdate will refresh the current dir to master, reset mod files and push changes if there are any
func (lib *Library) ModUpdate(branch, commitMessage string) (err error) {
	lib.File.Output("Syncing " + branch + " with origin master...")

	if branch != "master" {
		if err = lib.File.CheckoutBranch("master"); err != nil {
			lib.File.Output("Checkout master failed :(")
		}
	}

	if err = lib.File.Fetch(); err != nil {
		lib.File.Output("Fetch failed :(")
	}

	if err = lib.File.Pull(); err != nil {
		lib.File.Output("Pull failed :(")
		return
	}

	if branch != "master" {
		if err = lib.File.CheckoutBranch(branch); err != nil {
			lib.File.Output("Checkout " + branch + " failed :(")
		}
		if err = lib.File.Merge("master"); err != nil {
			lib.File.Output("Merge master into " + branch + " failed :(")
		}
		lib.File.RunCmd("git", "checkout", "--theirs", "go.mod")
		lib.File.RunCmd("git", "checkout", "--theirs", "go.sum")
		if err = lib.File.Merge("master"); err != nil {
			lib.File.Output("Merge master into " + branch + " failed :(")
		}
	}

	lib.File.Output("Checking deps...")

	if err = lib.ModTidy(); err != nil {
		lib.File.Output("No mod file found. Skipping.")
		return
	}

	lib.File.RunCmd("rm", "go.sum")
	lib.ModSetDeps()

	if err = lib.ModTidy(); err != nil {
		lib.File.Output("Mod tidy failed :(")
		return
	}

	if err = lib.File.Add("go.*"); err != nil {
		lib.File.Output("Git add failed :(")
		return
	}

	message := "GoMu: " + commitMessage + "\n"
	for itr := lib.updatedDeps; itr != nil; itr = itr.Next {
		url := itr.File.GetGoURL()

		if itr.File.Updated {
			message += "\nUpdated " + url + "@" + itr.File.Version
		} else {
			message += "\nSet " + url + "@" + itr.File.Version
		}
	}

	if err = lib.File.Commit(message); err == nil {
		lib.File.Output("Updating mod files...")
	} else {
		lib.File.Output("Deps up to date!")
	}

	if pushErr := lib.File.Push(); pushErr != nil {
		lib.File.Output("Push failed :( check local changes")
		return pushErr
	}

	lib.File.Output("Mod Sync Complete!")
	return
}
