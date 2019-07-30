package main

import (
	"sort"
	"strings"
	"sync"

	"github.com/lxn/walk"
)

// ScanModel is a view model for the scan results.
//
// ScanModel is not threadsafe. Its operation should be managed by a single
// goroutine, such as the Sync function.
type ScanModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder

	mutex sync.RWMutex
	files []File
}

// RowCount returns the number of rows in the model.
func (m *ScanModel) RowCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.files)
}

// Value returns the value for the cell at the given row and column.
func (m *ScanModel) Value(row, col int) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if row >= len(m.files) {
		return nil
	}

	file := m.files[row]

	switch col {
	case 0:
		return file.Path
	case 1:
		return file.Version.String()
	case 2:
		return file.Version.ReleaseName()
	default:
		return nil
	}
}

// Checked is called by the TableView to retrieve if a given row is checked.
func (m *ScanModel) Checked(row int) bool {
	return false
}

// SetChecked is called by the TableView when the user toggled the check box
// of a given row.
func (m *ScanModel) SetChecked(row int, checked bool) error {
	return nil
}

// Sort is called by the TableView to sort the model.
func (m *ScanModel) Sort(col int, order walk.SortOrder) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sortColumn, m.sortOrder = col, order

	sort.SliceStable(m.files, func(i, j int) bool {
		a, b := m.files[i], m.files[j]

		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}

			return !ls
		}

		switch m.sortColumn {
		case 0:
			return c(strings.Compare(a.Path, b.Path) < 0)
		case 1:
			return c(a.Version.Release() < b.Version.Release())
		case 2:
			return c(a.Version.Release() < b.Version.Release())
		}

		panic("unexpected table sort column number")
	})

	return m.SorterBase.Sort(col, order)
}

// Clear removes all results from the model.
func (m *ScanModel) Clear() {
	m.mutex.Lock()
	m.files = nil
	m.mutex.Unlock()

	m.PublishRowsReset()
}

// Append adds the given results to the model.
func (m *ScanModel) Append(results ...File) {
	if len(results) == 0 {
		return
	}

	m.mutex.Lock()
	start := len(m.files)
	m.files = append(m.files, results...)
	end := len(m.files) - 1
	m.mutex.Unlock()

	if start <= end {
		m.PublishRowsInserted(start, end)
	}
}

// Results returns the current set of results in the model.
func (m *ScanModel) Results() []File {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.files
}
