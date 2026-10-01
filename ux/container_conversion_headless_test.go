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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestContainerConversionClosesEditorsAndUndoesTheKind verifies that converting a displayed equipment group to a plain
// item closes its editor first, and that undo brings back a group, not a physical container, even though the sheet is
// checked for modification, and so written out, in between. Undo and redo each discard an editor opened on the row in
// between, whose ID's kind the conversion changes, and redo brings back the legality class new equipment starts with.
func TestContainerConversionClosesEditorsAndUndoesTheKind(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	group := gurps.NewEquipmentGroup(entity, nil)
	group.Name = "Kit"
	entity.CarriedEquipment = []*gurps.Equipment{group}
	var sheet *Sheet
	screen.Do(func() {
		sheet = NewSheet("test"+gurps.SheetExt, entity)
		DisplayNewDockable(sheet)
	})
	editorsOpen := func() int {
		var count int
		screen.Do(func() {
			count = len(AllMatchingDockables(func(d unison.Dockable) bool {
				e, ok := d.(*editor[*gurps.Equipment, *gurps.EquipmentEditData])
				return ok && e.target == group
			}))
		})
		return count
	}
	var mgr *unison.UndoManager
	screen.Do(func() {
		mgr = unison.UndoManagerFor(sheet.CarriedEquipment.Table)
		EditEquipment(sheet, group, true)
	})
	c.NotNil(mgr, "the sheet must have an undo manager")
	c.Equal(1, editorsOpen(), "the group's editor must be open")

	screen.Do(func() {
		table := sheet.CarriedEquipment.Table
		table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		ConvertToNonContainer(sheet, table)
	})
	c.False(group.Container(), "the group must have become a plain item")
	c.Equal("4", group.LegalityClass, "the item must start out with the default legality class")
	c.Equal(0, editorsOpen(), "converting must close the group's editor")

	screen.Do(func() { EditEquipment(sheet, group, true) })
	c.Equal(1, editorsOpen(), "the item's editor must be open")
	screen.Do(func() { _ = sheet.Modified() })
	screen.Do(mgr.Undo)
	c.True(group.IsGroup(), "undo must bring back a group, not a physical container")
	c.Equal(0, editorsOpen(), "undo must discard the editor opened on the item")
	c.Equal("", group.LegalityClass, "undo must take the legality class away again")

	screen.Do(func() { EditEquipment(sheet, group, true) })
	c.Equal(1, editorsOpen(), "the group's editor must be open")
	screen.Do(mgr.Redo)
	c.False(group.Container(), "redo must convert the group again")
	c.Equal("4", group.LegalityClass)
	c.Equal(0, editorsOpen(), "redo must discard the editor opened on the group")
}
