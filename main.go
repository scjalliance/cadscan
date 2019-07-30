package main

import (
	"fmt"
	"os"
)

func main() {
	scanModel := &ScanModel{}

	treeModel, err := NewDirectoryTreeModel()
	if err != nil {
		fmt.Printf("Failed to prepare directory tree: %v\n", err)
		os.Exit(1)
	}

	scanner := NewScanner(scanModel, 32)
	defer scanner.Stop()

	window, err := NewScanWindow(scanner, treeModel, scanModel)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	window.Run()
}
