package main

import (
	"github.com/lxn/walk"
)

// DirectoryTreeModel is a model containing a directory tree.
type DirectoryTreeModel struct {
	walk.TreeModelBase
	roots []*Directory
}

var _ walk.TreeModel = new(DirectoryTreeModel)

// NewDirectoryTreeModel returns a new directory tree model.
func NewDirectoryTreeModel() (*DirectoryTreeModel, error) {
	model := new(DirectoryTreeModel)

	drives, err := walk.DriveNames()
	if err != nil {
		return nil, err
	}

	for _, drive := range drives {
		switch drive {
		case "A:\\", "B:\\":
			continue
		}

		model.roots = append(model.roots, NewDirectory(drive, nil))
	}

	return model, nil
}

// LazyPopulation indicates whether the model should be populated lazily.
// It returns true.
func (*DirectoryTreeModel) LazyPopulation() bool {
	// We don't want to eagerly populate our tree view with the whole file system.
	return true
}

// RootCount returns the number of root directories present in the tree.
func (m *DirectoryTreeModel) RootCount() int {
	return len(m.roots)
}

// RootAt returns the root at the given index.
func (m *DirectoryTreeModel) RootAt(index int) walk.TreeItem {
	return m.roots[index]
}
