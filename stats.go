package gomu

import "strconv"

// ActionStats contain stats related to the current action
type ActionStats struct {
	Options  *Options
	DepCount int

	UpdateCount   int
	UpdatedOutput string

	TagCount     int
	TaggedOutput string

	CommitCount    int
	DeployedOutput string

	PRCount  int
	PROutput string

	CreatedCount  int
	CreatedOutput string

	TestFailedCount  int
	TestFailedOutput string
}

type toString int

func (i toString) string() {
	return
}

// Format returns an formatted output string to print stat report
func (stats ActionStats) Format() (output string) {
	if stats.Options.Action == "list" {
		// Already printed
		return
	}

	branch := stats.Options.Branch
	if len(branch) == 0 {
		branch = "Current Branch"
	}

	switch stats.Options.Action {
	case "pull":
		output += "Pulled latest version of <" + branch + "> in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
	case "test":
		if stats.TestFailedCount == 0 {
			output += "All tests passed in " + strconv.Itoa(stats.DepCount) + " lib(s)!\n"
		} else {
			output += "Tests failed in " + strconv.Itoa(stats.TestFailedCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s) :(\n"
			output += stats.TestFailedOutput
		}
	case "replace":
		output += "Replaced local dependencies in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
	case "reset":
		output += "Reset mod files in " + strconv.Itoa(stats.DepCount) + " lib(s)\n"
		// TODO: Count libs with changes here?
		output += "Warning: Local changes will no longer apply\n" //in " + strconv.Itoa(stats.DepCount) + " lib(s)\n"
	case "sync":
		// Print update status
		if stats.UpdateCount == 0 {
			output += "All " + strconv.Itoa(stats.DepCount) + " lib dependencies already up to date!\n"
		} else {
			output += "Updated mod files in <" + branch + "> for " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.UpdatedOutput
		}
	}

	if stats.Options.Tag {
		// Print tag status
		output += "\n"
		if stats.TagCount == 0 {
			output += "All " + strconv.Itoa(stats.DepCount) + " lib tags already up to date!\n"
		} else {
			if len(stats.Options.SetVersion) == 0 {
				output += "Updated tag for " + strconv.Itoa(stats.TagCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			} else {
				output += "Tag set to " + stats.Options.SetVersion + " for " + strconv.Itoa(stats.TagCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			}
			output += stats.TaggedOutput
		}
	}

	if stats.Options.Commit {
		// Print commit status
		output += "\n"
		if stats.CommitCount == 0 {
			output += "No local changes to commit in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
		} else {
			output += "Committed new changes to <" + branch + "> in " + strconv.Itoa(stats.CommitCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.DeployedOutput
		}
	}

	if stats.CreatedCount > 0 {
		output += "\n"
		output += "Created branch <" + branch + "> in " + strconv.Itoa(stats.CreatedCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.CreatedOutput
	}

	if stats.Options.PullRequest {
		// Print pr status
		output += "\n"
		if stats.PRCount == 0 {
			output += "No Pull Requests opened in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
		} else {
			output += "Created Pull Request from <" + branch + "> to <master> in " + strconv.Itoa(stats.PRCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.PROutput
		}
	}

	return
}
