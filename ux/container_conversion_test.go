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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// newSheetWithConvertibleNotes returns a sheet whose notes list holds a plain note, an empty container and a container
// with a child, so that a selection of all three exercises every branch of the conversion checks: the plain note can
// only become a container, the empty container can only become a plain note, and the container with a child cannot be
// converted at all.
func newSheetWithConvertibleNotes(t *testing.T) (sheet *Sheet, plain, empty, parent *gurps.Note) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	plain = gurps.NewNote(entity, nil, false)
	plain.MarkDown = "Plain"
	empty = gurps.NewNote(entity, nil, true)
	empty.MarkDown = "Empty"
	parent = gurps.NewNote(entity, nil, true)
	parent.MarkDown = "Parent"
	child := gurps.NewNote(entity, parent, false)
	child.MarkDown = "Child"
	parent.Children = []*gurps.Note{child}
	entity.Notes = []*gurps.Note{plain, empty, parent}
	sheet.Rebuild(true)
	return sheet, plain, empty, parent
}

// TestContainerConversionChecksHonorDirectionAndChildren verifies that the enablement checks only report a convertible
// selection when a selected row can actually move in the requested direction, and that a container holding children is
// never offered up for conversion to a non-container.
func TestContainerConversionChecksHonorDirectionAndChildren(t *testing.T) {
	c := check.New(t)
	sheet, plain, empty, parent := newSheetWithConvertibleNotes(t)
	table := sheet.Notes.Table

	table.SetSelectionMap(nil)
	c.False(CanConvertToContainer(table), "an empty selection cannot be converted to containers")
	c.False(CanConvertToNonContainer(table), "an empty selection cannot be converted to non-containers")

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true})
	c.True(CanConvertToContainer(table), "a plain note can become a container")
	c.False(CanConvertToNonContainer(table), "a plain note is already a non-container")

	table.SetSelectionMap(map[tid.TID]bool{empty.ID(): true})
	c.False(CanConvertToContainer(table), "an empty container is already a container")
	c.True(CanConvertToNonContainer(table), "an empty container can become a plain note")

	table.SetSelectionMap(map[tid.TID]bool{parent.ID(): true})
	c.False(CanConvertToContainer(table), "a container with children is already a container")
	c.False(CanConvertToNonContainer(table), "a container with children cannot become a plain note")

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	c.True(CanConvertToContainer(table), "a mixed selection can be converted to containers if any row can")
	c.True(CanConvertToNonContainer(table), "a mixed selection can be converted to non-containers if any row can")
}

// TestContainerConversionIsUndoableAsOneEdit verifies that converting a mixed selection only touches the rows that can
// move in the requested direction, that the whole conversion is recorded as a single undo edit carrying the action's
// title, and that undo and redo each put every touched row back where it belongs.
func TestContainerConversionIsUndoableAsOneEdit(t *testing.T) {
	c := check.New(t)
	sheet, plain, empty, parent := newSheetWithConvertibleNotes(t)
	table := sheet.Notes.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.False(mgr.CanUndo(), "nothing must be undoable before any conversion")
	sheet.Entity().ModifiedOn = jio.Time{}

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	ConvertToContainer(sheet, table)
	c.True(plain.Container(), "the plain note must have become a container")
	c.True(empty.Container(), "the empty container must remain a container")
	c.True(parent.Container(), "the container with children must remain a container")
	c.Equal(1, len(parent.Children), "the container's children must be untouched")
	c.True(mgr.CanUndo(), "the conversion must be undoable")
	c.Equal("Undo "+convertToContainerAction.Title, mgr.UndoTitle(), "the undo edit must carry the action's title")
	c.NotEqual(jio.Time{}, sheet.Entity().ModifiedOn, "the conversion must mark the sheet as modified")

	mgr.Undo()
	c.False(plain.Container(), "undo must turn the note back into a plain note")
	c.True(empty.Container(), "undo must leave the untouched container alone")
	c.False(mgr.CanUndo(), "the whole conversion must have been a single undo edit")

	c.True(mgr.CanRedo(), "the conversion must be redoable")
	mgr.Redo()
	c.True(plain.Container(), "redo must turn the note back into a container")
	mgr.Undo()

	// Now the other direction: both the plain note and the empty container are selected along with the container that
	// has children, but only the empty container can become a plain note.
	table = sheet.Notes.Table
	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	ConvertToNonContainer(sheet, table)
	c.False(plain.Container(), "the plain note must remain a plain note")
	c.False(empty.Container(), "the empty container must have become a plain note")
	c.True(parent.Container(), "the container with children must not be converted")
	c.Equal("Undo "+convertToNonContainerAction.Title, mgr.UndoTitle(), "the undo edit must carry the action's title")

	mgr.Undo()
	c.True(empty.Container(), "undo must turn the note back into a container")
	c.False(plain.Container(), "undo must leave the untouched plain note alone")
	c.True(parent.Container(), "undo must leave the container with children alone")

	// A selection with nothing convertible must not record an edit.
	table = sheet.Notes.Table
	table.SetSelectionMap(map[tid.TID]bool{parent.ID(): true})
	ConvertToNonContainer(sheet, table)
	c.True(parent.Container(), "a container with children must not be converted")
	c.True(mgr.CanRedo(), "a conversion that changes nothing must not have recorded an edit that clears the redo stack")
}
