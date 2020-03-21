package gomu

import "strconv"

// ActionStats contain stats related to the current action
type ActionStats struct {
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
func (stats ActionStats) Format(action, branch string) (output string) {
	if action == "list" {
		return
	}

	if action == "pull" {
		// Print pull status
		output += "Pulled latest version of " + branch + " in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
		return
	}

	if action == "replace-local" {
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
		output += "Updated mod files in " + strconv.Itoa(stats.UpdateCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.UpdatedOutput
	}

	output += "\n"

	// Print tag status
	if stats.TagCount == 0 {
		output += "All " + strconv.Itoa(stats.DepCount) + " lib tags already up to date!"
		output += ""
	} else {
		output += "Updated tag in " + strconv.Itoa(stats.TagCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
		output += stats.TaggedOutput
	}

	if action == "deploy" {
		// Print deploy status
		output += "\n"
		if stats.DeployedCount == 0 {
			output += "No local changes to deploy in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
			output += ""
		} else {
			output += "Deployed new changes to <" + branch + "> in " + strconv.Itoa(stats.DeployedCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.DeployedOutput
		}
	} else if action == "install" {
		// Print install status
		output += "\n"
		if stats.InstalledCount == 0 {
			output += "No packages installed in " + strconv.Itoa(stats.DepCount) + " lib(s).\n"
			output += ""
		} else {
			output += "Installed " + strconv.Itoa(stats.InstalledCount) + "/" + strconv.Itoa(stats.DepCount) + " lib(s):\n"
			output += stats.InstalledOutput
		}
	}

	return
}
