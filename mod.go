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
	for itr := listHead; itr != nil; itr = itr.Next {
		// Create new node to add to independent list on lib
		var node sort.FileNode
		node.File = itr.File
		lib.AddDep(&node)
	}
}

// ModSetDeps adds a dep@version to go.mod to force-update or force-downgrade any deps in the filtered chain
func (lib *Library) ModSetDeps() {
	for itr := lib.updatedDeps; itr != nil; itr = itr.Next {
		if len(itr.File.Version) == 0 {
			lib.File.Output("Error: no version to set for " + itr.File.Path)
		} else {
			lib.File.RunCmd("go", "get", itr.File.Path+"@"+itr.File.Version)
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

	if len(tag) == 0 {
		version := ""
		if !strings.HasSuffix(strings.Trim(lib.File.Path, "/"), "-plugin") {
			version = lib.GetCurrentTag()
		}

		// Set old version of libs in case they weren't updated previously
		lib.File.Version = version
		if lib.File.Commit("Deploy local changes before incrementing version from "+version) == nil {
			deployed = true
		}

	} else {
		if lib.File.Commit("Deploy local changes before updating version to "+tag) == nil {
			deployed = true
		}
	}

	return
}

// ModUpdate will refresh the current dir to master, reset mod files and push changes if there are any
func (lib *Library) ModUpdate(commitMessage string) (err error) {
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

	if err = lib.File.RunCmd("rm", "go.mod"); err != nil {
		lib.File.Output("No mod file found. Skipping.")
		return
	}

	if err = lib.File.RunCmd("rm", "go.sum"); err != nil {
		lib.File.Output("No sum file found.")
	}

	if err = lib.File.RunCmd("go", "mod", "init"); err != nil {
		lib.File.Output("Mod init failed :(")
		return
	}

	lib.ModSetDeps()

	if err = lib.File.RunCmd("go", "mod", "tidy"); err != nil {
		lib.File.Output("Mod tidy failed :(")
		return
	}

	if err = lib.File.Add("go.*"); err != nil {
		lib.File.Output("Git add failed :(")
		return
	}

	if err = lib.File.Commit(commitMessage); err != nil {
		lib.File.Output("Deps up to date!")
		return
	}

	lib.File.Output("Updating mod files...")
	if err = lib.File.Push(); err != nil {
		lib.File.Output("Update failed :(")
		return
	}

	lib.File.Output("Deps updated!")
	return
}
