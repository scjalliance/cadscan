package main

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/lxn/walk"
)

type dirEntry struct {
	scanning bool
}

// Directory is a tree item.
type Directory struct {
	name     string
	parent   *Directory
	ready    chan struct{}
	children []*Directory
	once     sync.Once
}

var _ walk.TreeItem = new(Directory)

// NewDirectory returns a new directory with the given name and parent.
func NewDirectory(name string, parent *Directory) *Directory {
	return &Directory{
		name:   name,
		parent: parent,
		ready:  make(chan struct{}),
	}
}

// Text returns the text for the tree item.
func (d *Directory) Text() string {
	return d.name
}

// Parent returns the parent of the tree item.
func (d *Directory) Parent() walk.TreeItem {
	if d.parent == nil {
		// We can't simply return d.parent in this case, because the interface
		// value then would not be nil.
		return nil
	}

	return d.parent
}

// ChildCount returns the number of children within the directory.
func (d *Directory) ChildCount() int {
	go d.once.Do(d.scan)
	<-d.ready
	return len(d.children)
}

// ChildAt returns the tree item for a particular child.
func (d *Directory) ChildAt(index int) walk.TreeItem {
	go d.once.Do(d.scan)
	<-d.ready
	go d.children[index].Scan() // Warm up the next level down
	return d.children[index]
}

// Scan instructs d to scan its children if it has not already done so.
func (d *Directory) Scan() {
	go d.once.Do(d.scan)
}

func (d *Directory) scan() {
	defer close(d.ready)

	dirname := d.Path()
	f, err := os.Open(dirname)
	if err != nil {
		return
	}

	var names []string
	for {
		files, err := f.Readdir(512)
		for _, info := range files {
			name := info.Name()
			if !info.IsDir() || shouldExclude(name) {
				continue
			}
			names = append(names, name)
		}
		if err != nil {
			break
		}
	}

	f.Close()

	sort.Strings(names)

	for _, name := range names {
		//fmt.Printf("Scanned: %s\n", filepath.Join(dirname, name))
		d.children = append(d.children, NewDirectory(name, d))
	}
}

// Image returns the path of the directory.
func (d *Directory) Image() interface{} {
	return d.Path()
}

// Path returns the absolute path of the directory.
func (d *Directory) Path() string {
	elems := []string{d.name}

	dir, _ := d.Parent().(*Directory)

	for dir != nil {
		elems = append([]string{dir.name}, elems...)
		dir, _ = dir.Parent().(*Directory)
	}

	return filepath.Join(elems...)
}

func shouldExclude(name string) bool {
	switch name {
	case "System Volume Information", "pagefile.sys", "swapfile.sys", "$RECYCLE.BIN":
		return true
	}

	return false
}
