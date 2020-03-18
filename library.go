package sync

import (
	common "github.com/hatchify/mod-common"
	sort "github.com/hatchify/mod-sort"
)

// Library represents a file and its updated deps
type Library struct {
	File *common.FileWrapper

	updatedDeps sort.FileList
}

// AddDep will ensure go.mod sets specific version of node.file when syncing
func (lib *Library) AddDep(node *sort.FileNode) {
	node.InsertInto(lib.updatedDeps)
}
