// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/unison"
)

// hierarchyDiscloser is something whose containers can be opened and closed as a group: a page list, a list dockable's
// table, and a character sheet's attributes and body type.
type hierarchyDiscloser interface {
	// FirstDisclosureState returns the open state of the first container, and whether there is one.
	FirstDisclosureState() (open, exists bool)
	// SetDisclosureState opens or closes every container.
	SetDisclosureState(open bool)
}

// noteDiscloser is something whose notes can be shown and hidden as a group: a page list or a list dockable's table.
type noteDiscloser interface {
	// FirstNoteState returns the state of the first note: -1 is closed, 1 is open, 0 is none found.
	FirstNoteState() int
	// ApplyNoteState shows or hides every note.
	ApplyNoteState(closed bool)
}

// toggleHierarchy opens every container of the disclosers when the first container found among them is closed, and
// closes every one of them otherwise, so that the group as a whole flips between all open and all closed. The state
// they were put in is returned. With no container among them there is nothing to flip, and they are all, harmlessly,
// opened.
func toggleHierarchy[D hierarchyDiscloser](disclosers ...D) (open bool) {
	for _, d := range disclosers {
		var exists bool
		if open, exists = d.FirstDisclosureState(); exists {
			break
		}
	}
	open = !open
	for _, d := range disclosers {
		d.SetDisclosureState(open)
	}
	return open
}

// toggleNotes hides every note of the disclosers when the first note found among them is shown, and shows every one
// of them otherwise, and reports whether there was a note to act on. With none, nothing is changed, and the caller has
// nothing to refresh.
func toggleNotes[D noteDiscloser](disclosers ...D) (changed bool) {
	state := 0
	for _, d := range disclosers {
		if state = d.FirstNoteState(); state != 0 {
			break
		}
	}
	if state == 0 {
		return false
	}
	for _, d := range disclosers {
		d.ApplyNoteState(state == 1)
	}
	return true
}

// firstTableDisclosureState returns the open state of the first root row of the table that can have children, and
// whether there is one.
func firstTableDisclosureState[T gurps.Node[T]](table *unison.Table[*Node[T]]) (open, exists bool) {
	for _, row := range table.RootRows() {
		if row.CanHaveChildren() {
			return row.IsOpen(), true
		}
	}
	return false, false
}

// setTableDisclosureState opens or closes every row of the table that can have children, at every depth.
func setTableDisclosureState[T gurps.Node[T]](table *unison.Table[*Node[T]], open bool) {
	for _, row := range table.RootRows() {
		if row.CanHaveChildren() {
			setRowOpen(row, open)
		}
	}
}

// setRowOpen opens or closes the row and every row beneath it that can have children.
func setRowOpen[T gurps.Node[T]](row *Node[T], open bool) {
	row.SetOpen(open)
	for _, child := range row.Children() {
		if child.CanHaveChildren() {
			setRowOpen(child, open)
		}
	}
}

// firstTableNoteState returns the state of the first note in the table: -1 is closed, 1 is open, 0 is none found.
func firstTableNoteState[T gurps.Node[T]](table *unison.Table[*Node[T]]) int {
	for _, row := range table.RootRows() {
		if state := noteState(row); state != 0 {
			return state
		}
	}
	return 0
}

// applyTableNoteState shows or hides every note in the table, at every depth.
func applyTableNoteState[T gurps.Node[T]](table *unison.Table[*Node[T]], closed bool) {
	for _, row := range table.RootRows() {
		applyNoteState(row, closed)
	}
}

// noteState returns the state of the first note found in the row or, failing that, beneath it: -1 is closed, 1 is
// open, 0 is none found.
func noteState[T gurps.Node[T]](n *Node[T]) int {
	if n.hasNote() {
		if gurps.IsClosed(n.noteKey()) {
			return -1
		}
		return 1
	}
	if n.CanHaveChildren() {
		for _, child := range n.Children() {
			if state := noteState(child); state != 0 {
				return state
			}
		}
	}
	return 0
}

// applyNoteState shows or hides the row's note, if it has one, and those of every row beneath it.
func applyNoteState[T gurps.Node[T]](n *Node[T], closed bool) {
	if n.hasNote() {
		if key := n.noteKey(); gurps.IsClosed(key) != closed {
			gurps.SetClosedState(key, closed)
		}
	}
	if n.CanHaveChildren() {
		for _, child := range n.Children() {
			applyNoteState(child, closed)
		}
	}
}

// hasNote returns true if any of the row's cells carries a note, which is shown as secondary text.
func (n *Node[T]) hasNote() bool {
	for i := range n.table.Columns {
		var data gurps.CellData
		n.data.CellData(n.table.Columns[i].ID, &data)
		if data.Type == cell.Text && data.Secondary != "" {
			return true
		}
	}
	return false
}

// noteKey returns the key under which the closed state of the row's note is kept.
func (n *Node[T]) noteKey() string {
	return "N:" + string(n.ID())
}
