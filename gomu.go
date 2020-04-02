package gomu

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hatchify/closer"
	com "github.com/hatchify/mod-common"
	sort "github.com/hatchify/mod-sort"
)

// MU represents a Mod Utils instance which sets options from flags and allows actions to be called
type MU struct {
	Options Options

	AllDirectories  sort.StringArray
	SortedLibraries sort.FileList

	Stats ActionStats

	Errors []error

	closer *closer.Closer
}

// Options represents different settings to perform an action
type Options struct {
	Action string

	Branch        string
	CommitMessage string

	Tag         bool
	Commit      bool
	PullRequest bool

	TargetDirectories  sort.StringArray
	FilterDependencies sort.StringArray

	LogLevel com.LogLevel
}

var closed = false

// New returns new Mod Utils struct
func New(options Options) *MU {
	var mu MU
	mu.Stats.Options = &options
	mu.Options = options
	return &mu
}

// Then handles cleanup after func
func cleanupStash(libs sort.StringArray) {
	closed = true

	// Resume working directory
	var f com.FileWrapper
	for i := range libs {
		f.Path = libs[i]
		if f.StashPop() {
			f.Output("Warning - Has local changes")
		}
	}
}

// RunThen runs gomu and then calls closure
func RunThen(mu *MU, complete func(mu *MU)) {
	// Handle closures
	mu.closer = closer.New()

	// Go do the thing
	go mu.Perform()
	// Ensure closure is called
	mu.Close()

	// Callback to completion handler
	complete(mu)
}

// Close handles cleanup
func (mu *MU) Close() {
	mu.closer.Wait()

	if len(mu.Errors) > 0 {
		com.Println("\nEncountered error! Cleaning...")

	} else {
		com.Println("\nFinishing up. Cleaning...")
	}

	cleanupStash(mu.AllDirectories)
}

// Perform executes whatever action is set in mu.Options
func (mu *MU) Perform() {
	if len(mu.Options.TargetDirectories) > 0 {
		com.Println("\nSearching", mu.Options.TargetDirectories, "for git repositories...")
	} else {
		com.Println("\nSearching for git repositories in current directory...")
	}

	// Get all libs within target dirs
	mu.PopulateLibsFromTargets()
	libs := mu.AllDirectories

	com.Println("\nFound", len(libs)+1, "file(s). Scanning for dependencies...")

	var f com.FileWrapper
	for i := range libs {
		f.Path = libs[i]
		// Hide local changes to prevent interference with searching/syncing
		f.Stash()
	}

	branch := mu.Options.Branch
	if len(branch) == 0 {
		branch = "Current Branch"
	}

	// Sort libs
	var fileHead *sort.FileNode
	fileHead, mu.Stats.DepCount = libs.SortedDependingOnAny(mu.Options.FilterDependencies)
	if len(mu.Options.FilterDependencies) == 0 || len(mu.Options.FilterDependencies) == 0 {
		com.Println("\nPerforming", mu.Options.Action, "on branch <"+branch+"> for", mu.Stats.DepCount, "lib(s)")
	} else {
		com.Println("\nPerforming", mu.Options.Action, "on branch <"+branch+"> for", mu.Stats.DepCount, "lib(s) depending on", mu.Options.FilterDependencies)
	}

	switch mu.Options.Action {
	case "sync":
		msg := []string{"This will update mod files"}
		if mu.Options.Commit {
			msg = append(msg, "&& commit local changes (if any)")
		}
		if mu.Options.Tag {
			msg = append(msg, "&& tag new versions (if updated)")
		}
		com.Println("\n" + strings.Join(msg, ", "))

		if !ShowWarning("\nIs this ok?") {
			cleanupStash(libs)
			os.Exit(-1)
		}
	default:
		// No worries
	}

	// Perform action on sorted libs
	index := 0
	for itr := fileHead; itr != nil; itr = itr.Next {
		if closed {
			// Stop execution and clean up
			return
		}

		index++

		// If we're just listing files, we don't need to do anything else :)
		if mu.Options.Action == "list" {
			com.Println(strconv.Itoa(index) + ") " + itr.File.GetGoURL())
			continue
		}

		// Separate output
		com.Println("")
		com.Println("(", index, "/", mu.Stats.DepCount, ")", itr.File.Path)

		if mu.Options.Action == "pull" {
			// Check out branch if provided
			if len(mu.Options.Branch) > 0 {
				itr.File.Output("Checking out " + mu.Options.Branch + "...")
				if itr.File.CheckoutBranch(mu.Options.Branch) != nil {
					itr.File.Output("Failed to check out branch :(")
					switch mu.Options.Action {
					case "deploy", "sync":
						// Quit. We failed
					}
				}
			}

			// Only git pull.
			itr.File.Output("Pulling latest changes...")
			if itr.File.Pull() == nil {
				itr.File.Updated = true
				mu.Stats.UpdateCount++
				mu.Stats.UpdatedOutput += strconv.Itoa(mu.Stats.UpdateCount) + ") " + itr.File.Path

				mu.Stats.UpdatedOutput += "\n"
			}

			continue
		}

		// Create sync lib ref from dep file
		var lib Library
		lib.File = itr.File

		if mu.Options.Action == "replace" {
			// Append local replacements for all libs in lib.updatedDeps
			lib.File.Output("Setting local replacements...")

			// Aggregate updated versions of previously parsed deps
			lib.ModAddDeps(fileHead)

			if lib.ModReplaceLocal() {
				lib.File.Updated = true
				mu.Stats.UpdateCount++
				mu.Stats.UpdatedOutput += strconv.Itoa(mu.Stats.UpdateCount) + ") " + lib.File.Path + "\n"

				lib.File.Output("Local replacements set!")
			} else {
				lib.File.Output("Failed to set local deps :(")
			}
			continue
		}

		if mu.Options.Action == "reset" {
			if len(mu.Options.Branch) > 0 {
				lib.File.Output("Reverting mod files to <" + mu.Options.Branch + "> ref...")
			} else {
				lib.File.Output("Reverting mod files to last-committed ref...")
			}

			lib.File.StashPop()

			// Revert any changes to mod files
			lib.File.RunCmd("git", "checkout", mu.Options.Branch, "go.mod")
			lib.File.RunCmd("git", "checkout", mu.Options.Branch, "go.sum")

			lib.File.Output("Reverted mod files!")

			if lib.File.HasChanges() {
				lib.File.Output("Warning! Has local changes.")
			}

			continue
		}

		// Handle branching
		switched := false
		created := false
		var branchErr error
		if len(mu.Options.Branch) > 0 {
			switched, created, branchErr = itr.File.CheckoutOrCreateBranch(mu.Options.Branch)
			if branchErr != nil {
				itr.File.Error("Failed to checkout " + mu.Options.Branch + " :(")
			} else if !switched {
				itr.File.Output("Already on " + mu.Options.Branch)
			}
		}

		itr.File.Output("Pulling latest changes...")
		if itr.File.Pull() != nil {
			itr.File.Output("Failed to pull " + mu.Options.Branch + " :(")
		}

		// Aggregate updated versions of previously parsed deps
		lib.ModAddDeps(fileHead)

		if mu.Options.Commit {
			// TODO: Branch and PR? Diff?
			lib.File.Output("Checking for local changes...")
			lib.File.Deployed = lib.ModDeploy("")

			if lib.File.Deployed {
				mu.Stats.DeployedCount++
				mu.Stats.DeployedOutput += strconv.Itoa(mu.Stats.DeployedCount) + ") " + itr.File.Path + "\n"
			}
		}

		// Update the dep if necessary
		if err := lib.ModUpdate(mu.Options.Branch, mu.Options.CommitMessage); err == nil {
			// Dep was updated
			lib.File.Updated = true
			mu.Stats.UpdateCount++
			mu.Stats.UpdatedOutput += strconv.Itoa(mu.Stats.UpdateCount) + ") " + lib.File.Path + "\n"
		}

		// Check if created a branch we didn't need
		if !lib.File.Tagged && !lib.File.Updated && !lib.File.Deployed {
			switch mu.Options.Branch {
			case "master", "develop", "staging", "beta", "prod", "":
				// Ignore protected branches and empty branch
			default:
				// Delete branch
				if created {
					lib.File.CheckoutBranch("master")
					lib.File.RunCmd("git", "branch", "-D", mu.Options.Branch)
				}
			}
		}

		if !mu.Options.Tag || strings.HasSuffix(strings.Trim(itr.File.Path, "/"), "-plugin") {
			// Ignore tagging entirely
			continue
		} else {
			// Tag if forced or if able to increment
			if mu.Options.Tag && lib.ShouldTag() {
				// TODO: Support explicit versions?
				itr.File.Version = lib.TagLib("")
				itr.File.Tagged = true
				mu.Stats.TagCount++
				mu.Stats.TaggedOutput += strconv.Itoa(mu.Stats.TagCount) + ") " + lib.File.Path + " " + lib.File.Version + "\n"
			}
		}

		// Set tag for next lib if not set
		if len(itr.File.Version) == 0 {
			itr.File.Version = lib.GetCurrentTag()
		}
	}

	if com.GetLogLevel() == com.NAMEONLY {
		// Print names and quit
		for fileItr := fileHead; fileItr != nil; fileItr = fileItr.Next {
			if fileItr.File.Tagged || fileItr.File.Deployed || fileItr.File.Updated || fileItr.File.Installed || mu.Options.Action == "list" {
				com.Outputln(com.NAMEONLY, fileItr.File.GetGoURL())
			}
		}
	}

	if !mu.closer.Close(nil) {
		mu.Errors = append(mu.Errors, fmt.Errorf("failed to close! Check for local changes and stashes in %v", mu.Options.TargetDirectories))
	}
}
