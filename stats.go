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

	DeployedCount  int
	DeployedOutput string

	PRCount  int
	PROutput string
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

	if stats.Options.Action == "pull" {
		// Print pull status
		output += "Pulled latest version of <" + branch + "> in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
		return
	}

	if stats.Options.Action == "replace" {
		// Print replacement status
		output += "Replaced local dependencies in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
		return
	}

	if stats.Options.Action == "reset" {
		// Print replacement status
		output += "Reset mod files in " + strconv.Itoa(stats.DepCount) + " lib(s)\n"
		// TODO: Count libs with changes here?
		output += "Warning: Local changes will no longer apply" //in " + strconv.Itoa(stats.DepCount) + " lib(s)\n"
		return
	}

	// Print update status
	if stats.UpdateCount == 0 {
		output += "All " + strconv.Itoa(stats.DepCount) + " lib dependencies already up to date!"
	} else {
		output += "Updated mod files in <" + branch + "> for " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
	}

	if stats.Options.Tag {
		// Print tag status
		output += "\n"
		if stats.TagCount == 0 {
			output += "All " + strconv.Itoa(stats.DepCount) + " lib tags already up to date!"
		} else {
			output += "Updated tag for " + strconv.Itoa(stats.TagCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.TaggedOutput
		}
	}

	if stats.Options.Commit {
		// Print deploy status
		output += "\n"
		if stats.DeployedCount == 0 {
			output += "No local changes to commit in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
		} else {
			output += "Committed new changes to <" + branch + "> in " + strconv.Itoa(stats.DeployedCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.DeployedOutput
		}
	}

	if stats.Options.PullRequest {
		// Print deploy status
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
