package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"

	ui "github.com/lxn/walk/declarative"
)

// ScanWindow is the main scanning window.
type ScanWindow struct {
	scanner         *Scanner
	model           *ScanModel
	ui              *ui.MainWindow
	form            *walk.MainWindow
	tree            *walk.TreeView
	table           *walk.TableView
	splitter        *walk.Splitter
	selection       *walk.LineEdit
	cancel          *walk.PushButton
	actionCopy      *walk.Action
	actionSelectAll *walk.Action
}

// NewScanWindow returns a new scanning window.
func NewScanWindow(scanner *Scanner, treeModel *DirectoryTreeModel, scanModel *ScanModel) (window *ScanWindow, err error) {
	window = &ScanWindow{
		scanner: scanner,
		model:   scanModel,
	}

	icon, err := walk.NewIconFromResourceId(2)
	if err != nil {
		icon = walk.IconInformation()
	}

	window.ui = &ui.MainWindow{
		Icon:     icon,
		Title:    "CAD File Scanner",
		MinSize:  ui.Size{Width: 600, Height: 400},
		Size:     ui.Size{Width: 1024, Height: 640},
		Layout:   ui.VBox{},
		AssignTo: &window.form,
		MenuItems: []ui.MenuItem{
			ui.Menu{
				Text: "&Edit",
				Items: []ui.MenuItem{
					ui.Action{
						AssignTo:    &window.actionSelectAll,
						Text:        "Select &All",
						Shortcut:    ui.Shortcut{Key: walk.KeyA, Modifiers: walk.ModControl},
						OnTriggered: window.onSelectAllResults,
					},
					ui.Action{
						AssignTo:    &window.actionCopy,
						Text:        "Copy",
						Shortcut:    ui.Shortcut{Key: walk.KeyC, Modifiers: walk.ModControl},
						OnTriggered: window.onCopyResults,
					},
				},
			},
		},
		Children: []ui.Widget{
			ui.HSplitter{
				AssignTo:    &window.splitter,
				HandleWidth: 12,
				Children: []ui.Widget{
					ui.TreeView{
						AssignTo:             &window.tree,
						Model:                treeModel,
						OnCurrentItemChanged: window.onSelection,
					},
					ui.TableView{
						Name:           "table",
						AssignTo:       &window.table,
						MultiSelection: true,
						Columns: []ui.TableViewColumn{
							{Title: "File", Width: 300},
							{Title: "Header", Width: 200},
							{Title: "Version", Width: 200},
						},
						ContextMenuItems: []ui.MenuItem{
							ui.ActionRef{Action: &window.actionSelectAll},
							ui.ActionRef{Action: &window.actionCopy},
						},
						Model: scanModel,
					},
				},
			},
			ui.HSplitter{
				Children: []ui.Widget{
					ui.Label{Text: "Directory:"},
					ui.LineEdit{AssignTo: &window.selection},
				},
			},
			ui.HSplitter{
				Children: []ui.Widget{
					ui.PushButton{
						Text:      "Scan",
						OnClicked: window.onScan,
					},
					ui.PushButton{
						AssignTo:  &window.cancel,
						Text:      "Cancel",
						OnClicked: window.onCancel,
						Enabled:   false,
					},
				},
			},
		},
	}

	err = window.ui.Create()

	if err == nil {
		window.splitter.SetFixed(window.tree, true)
		window.splitter.SetFixed(window.table, false)
	}

	return
}

// Run will display the queued lease dialog.
//
// Run blocks until the dialog is closed. The dialog can be closed by the user.
//
// Run returns the result of the dialog.
func (window *ScanWindow) Run() int {
	return window.form.Run()
}

func (window *ScanWindow) onSelection() {
	dir := window.tree.CurrentItem().(*Directory)
	window.selection.SetText(dir.Path())
}

func (window *ScanWindow) onScan() {
	root := window.selection.Text()
	go window.scanner.Scan(
		root,
		func() { window.form.Synchronize(window.onScanStarted) },
		func() { window.form.Synchronize(window.onScanCompleted) },
	)
}

func (window *ScanWindow) onScanStarted() {
	window.cancel.SetEnabled(true)
}

func (window *ScanWindow) onScanCompleted() {
	window.cancel.SetEnabled(false)
}

func (window *ScanWindow) onCancel() {
	go window.scanner.Stop()
}

func (window *ScanWindow) onSelectAllResults() {
	window.table.SetSelectedIndexes([]int{-1})
}

func (window *ScanWindow) onCopyResults() {
	rows := window.table.SelectedIndexes()
	if len(rows) == 0 {
		return
	}

	results := window.model.Results()

	files := make([]File, 0, len(rows))
	for _, row := range rows {
		if row >= len(results) {
			continue
		}
		files = append(files, results[row])
	}

	if len(files) == 0 {
		return
	}

	maxPath := 0
	maxHeader := 0
	for _, file := range files {
		if plen := len(file.Path); plen > maxPath {
			maxPath = plen
		}
		if hlen := len(file.Version.String()); hlen > maxHeader {
			maxHeader = hlen
		}
	}

	var lines []string
	for _, file := range files {
		lines = append(lines, fmt.Sprintf("%-*s  %-*s  %s", maxPath, file.Path, maxHeader, file.Version.String(), file.Version.ReleaseName()))
	}

	text := strings.Join(lines, "\n")

	walk.Clipboard().SetText(text)
}
