package gomu

import (
	common "github.com/hatchify/mod-common"
	sort "github.com/hatchify/mod-sort"
)

// Library represents a file and its updated deps
type Library struct {
	File *common.FileWrapper

	updatedDeps *sort.FileNode
}

// AddDep will ensure go.mod sets specific version of node.file when syncing
func (lib *Library) AddDep(node *sort.FileNode) {
	node.InsertInto(&lib.updatedDeps)
}

func performPull(branch string, itr *sort.FileNode) (success bool) {
	success = true

	if itr.File.CheckoutBranch(branch) != nil {
		itr.File.Output("Failed to checkout " + branch + " :(")
		success = false
	}

	if itr.File.Pull() == nil {
		itr.File.Output("Pull successful!")
	} else {
		itr.File.Output("Failed to pull " + branch + " :(")
		success = false
	}

	return
}
