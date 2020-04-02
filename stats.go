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

	InstalledCount  int
	InstalledOutput string
}

type toString int

func (i toString) string() {
	return
}

// Format returns an formatted output string to print stat report
func (stats ActionStats) Format() (output string) {
	if stats.Options.Action == "list" {
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

	// Print update status
	if stats.UpdateCount == 0 {
		output += "All " + strconv.Itoa(stats.DepCount) + " lib dependencies already up to date!"
		output += ""
	} else {
		output += "Updated mod files in <" + branch + "> for " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
	}

	output += "\n"

	// Print tag status
	if stats.TagCount == 0 {
		output += "All " + strconv.Itoa(stats.DepCount) + " lib tags already up to date!"
		output += ""
	} else {
		output += "Updated tag for " + strconv.Itoa(stats.TagCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.TaggedOutput
	}

	if stats.Options.Commit {
		// Print deploy status
		output += "\n"
		if stats.DeployedCount == 0 {
			output += "No local changes to commit in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
			output += ""
		} else {
			output += "Committed new changes to <" + branch + "> in " + strconv.Itoa(stats.DeployedCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.DeployedOutput
		}
	}

	return
}
