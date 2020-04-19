package gomu

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hatchify/closer"
	"github.com/hatchify/mod-utils/com"
	"github.com/hatchify/mod-utils/sort"
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

	Commit      bool
	PullRequest bool
	Tag         bool
	SetVersion  string

	SourcePath string

	DirectImport       bool
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

// Run runs gomu with configured mu.Options
func (mu *MU) Run() {
	// Handle closures
	mu.closer = closer.New()

	// Go do the thing
	go mu.performThenClose()

	// Ensure clean is called
	mu.waitThenClean()
}

// RunThen runs gomu with configured options and then calls closure
func (mu *MU) RunThen(complete func(mu *MU)) {
	mu.Run()

	// Callback to completion handler
	complete(mu)
}

// WaitThenClean handles cleanup
func (mu *MU) waitThenClean() {
	mu.closer.Wait()

	if len(mu.Errors) > 0 {
		com.Println("\nEncountered error! Cleaning...")

	} else {
		com.Println("\nFinishing up. Cleaning...")
	}

	cleanupStash(mu.AllDirectories)
}

// PerformThenClose executes whatever action is set in mu.Options
func (mu *MU) performThenClose() {
	mu.perform()

	if !mu.closer.Close(nil) {
		mu.Errors = append(mu.Errors, fmt.Errorf("failed to close! Check for local changes and stashes in %v", mu.Options.TargetDirectories))
	}
}

func (mu *MU) perform() {
	if mu.Options.PullRequest {
		authObject, err := com.LoadAuth()
		if err != nil || len(authObject.User) == 0 || len(authObject.Token) == 0 {
			com.Println("")
			com.Println("gomu :: I needs credentials for Pull Requests...")
			if authObject.Setup() != nil {
				com.Println("Error saving :(")
				err = fmt.Errorf("Unable to parse github username and token")
				return
			}
			err = nil
			com.Println("Saved Credentials!")
		}
	}

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
	for _, lib := range libs {
		f.Path = lib
		// Hide local changes to prevent interference with searching/syncing
		f.Stash()
	}

	branch := mu.Options.Branch
	if len(branch) == 0 {
		branch = "\"current\""
	}

	// Sort libs
	var fileHead *sort.FileNode
	if mu.Options.DirectImport {
		// Only check files in go.mod
		fileHead, mu.Stats.DepCount = libs.SortedDirectDeps(mu.Options.FilterDependencies)
	} else {
		// Check all files in go.sum
		fileHead, mu.Stats.DepCount = libs.SortedRecursiveDeps(mu.Options.FilterDependencies)
	}

	if len(mu.Options.FilterDependencies) == 0 || len(mu.Options.FilterDependencies) == 0 {
		com.Println("\nPerforming", mu.Options.Action, "on "+branch+" branch for", mu.Stats.DepCount, "lib(s)")
	} else {
		com.Println("\nPerforming", mu.Options.Action, "on "+branch+" branch for", mu.Stats.DepCount, "lib(s) depending on", mu.Options.FilterDependencies)
	}

	// TODO: Also add check to warn/confirm before pushing? It'd be nice to have a chance to backout both before and after changes took place
	// Eventual "undo" action possibly?
	switch mu.Options.Action {
	case "sync":
		warningLibs := make([]string, mu.Stats.DepCount)
		count := 0
		for itr := fileHead; itr != nil; itr = itr.Next {
			warningLibs[count] = strconv.Itoa(count+1) + ") " + itr.File.GetGoURL()
			count++
		}

		com.Println(strings.Join(warningLibs, "\n"))

		warningActions := []string{"Sync action will:"}
		if mu.Options.Branch != "" {
			warningActions = append(warningActions, "- checkout (or create) branch "+mu.Options.Branch)
		}
		warningActions = append(warningActions, "- update mod files")
		if mu.Options.Commit {
			warningActions = append(warningActions, "- commit local changes (if any)")
		}
		if mu.Options.PullRequest {
			warningActions = append(warningActions, "- open pull request for changes (if any)")
		}
		if mu.Options.Tag {
			if len(mu.Options.SetVersion) > 0 {
				warningActions = append(warningActions, "- tag all dependencies "+mu.Options.SetVersion)
			} else {
				warningActions = append(warningActions, "- increment tag version (if updated)")
			}
		}

		com.Println("\n" + strings.Join(warningActions, "\n  "))

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
		index++

		if closed {
			// Stop execution and clean up
			return
		}

		if mu.Options.Action == "list" {
			// If we're just listing, print 'n go ;)
			com.Println("(", index, "/", mu.Stats.DepCount, ")", itr.File.Path)
			continue
		}

		// Separate output
		com.Println("")
		com.Println("(", index, "/", mu.Stats.DepCount, ")", itr.File.Path)

		// Create sync lib ref from dep file
		var lib Library
		lib.File = itr.File

		switch mu.Options.Action {
		case "pull":
			if len(lib.File.Version) > 0 {
				lib.File.Output("Already has version set: " + lib.File.Version)
			} else {
				mu.pull(lib)
			}
			continue
		case "replace":
			mu.replace(lib, fileHead)
			continue
		case "reset":
			mu.reset(lib)
			continue
		case "test":
			mu.test(lib, fileHead)
			continue
		case "secret":
			mu.addSecret(lib)
			continue
		}

		// Sync
		if len(lib.File.Version) > 0 {
			lib.File.Output("Already has version set: " + lib.File.Version)
			continue
		}

		// Handle branching
		mu.updateOrCreateBranch(lib)

		if closed {
			// Stop execution and clean up
			return
		}

		if mu.Options.Action == "workflow" {
			// Add auto tag workflow
			lib.File.AddGitWorkflow(mu.Options.SourcePath)
		} else {
			// Aggregate updated versions of previously parsed deps
			lib.ModAddDeps(fileHead, false)
		}

		mu.commit(lib)

		if closed {
			// Stop execution and clean up
			return
		}

		commitTitle, commitMessage := mu.getCommitDetails(lib)
		mu.sync(lib, commitTitle, commitMessage)

		if closed {
			// Stop execution and clean up
			return
		}

		// Create PR
		mu.pullRequest(lib, mu.Options.Branch, commitTitle, commitMessage)

		if closed {
			// Stop execution and clean up
			return
		}

		mu.removeBranchIfUnused(lib)

		if closed {
			// Stop execution and clean up
			return
		}

		mu.tag(lib)
	}

	if com.GetLogLevel() == com.NAMEONLY {
		// Print names and quit
		for fileItr := fileHead; fileItr != nil; fileItr = fileItr.Next {
			if fileItr.File.Tagged || fileItr.File.Committed || fileItr.File.Updated || fileItr.File.PROpened || mu.Options.Action == "list" {
				com.Outputln(com.NAMEONLY, fileItr.File.GetGoURL())
			}
		}
	}
}
