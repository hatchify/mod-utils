package sort

// FileList represents a head to a list of FileNodes.
// Used to pass a pointer to a list such that it may be modified in-line
type FileList **FileNode

// InsertInto adds file to the provided file list in-line.
// NOTE: listHead will be modified if files are inserted at the beginning of list
func (node *FileNode) InsertInto(listHead FileList) {
	// Iterate over existing list
	for itr := *listHead; itr != nil; itr = itr.Next {
		if itr.File.Path == node.File.Path {
			// Don't need to add a file we already have in the list
			return // Done
		}

		if itr.File.DependsOn(node.File) {
			// Insert file before first lib that depends on it
			node.insertBefore(itr)

			break // Fall through to bottom to handle insert at index 0
		}

		if itr.Next == nil {
			// End of the list
			node.insertAfter(itr)

			return // Done
		}
	}

	// Handle file inserted at index 0, update file head
	if node.Last == nil {
		// Also handles empty list init case by default
		*listHead = node
	}
}

// insertAfter is used to add a fileNode to the end of the list
func (node *FileNode) insertAfter(itr *FileNode) {
	node.Last = itr
	itr.Next = node
}

// insertBefore is used to add a file before a given dep
func (node *FileNode) insertBefore(itr *FileNode) {
	// Set links from file
	node.Next = itr
	node.Last = itr.Last

	// Set links to file
	itr.Last = node
	if node.Last != nil {
		// Middle of list
		node.Last.Next = node
	}
}
