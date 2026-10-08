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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func columnIndexForID[T unison.TableRowConstraint[T]](table *unison.Table[T], id int) int {
	for i, col := range table.Columns {
		if col.ID == id {
			return i
		}
	}
	return -1
}

// The table synthesizes no click for Space at cell level, so page reference and check cells must answer the press
// themselves; a cell with nothing to press leaves Space to open the row's editor.
func TestSpaceOnACellWorksWhatTheCellHolds(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	const ref = "https://example.com/ref"
	var traits *unison.Table[*Node[*gurps.Trait]]
	var equipment *unison.Table[*Node[*gurps.Equipment]]
	var eq *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Alpha"
		tr.PageRef = ref
		// Replaces the new sheet's own traits, whose book references would prompt for a PDF if opened.
		entity.Traits = []*gurps.Trait{tr}
		eq = gurps.NewEquipment(entity, nil, false)
		eq.Name = "Rope"
		eq.Equipped = true
		entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
		entity.Recalculate()
		sheet.Rebuild(true)
		traits, _ = uxtest.FirstPanelOfType[*unison.Table[*Node[*gurps.Trait]]](sheet.AsPanel())
		for _, one := range uxtest.PanelsOfType[*unison.Table[*Node[*gurps.Equipment]]](sheet.AsPanel()) {
			if one.RootRowCount() > 0 {
				equipment = one
				break
			}
		}
	})
	if traits == nil || equipment == nil {
		t.Fatal("the character sheet must show the traits and the carried equipment")
	}
	dockables := func() (count int) {
		screen.Do(func() { count = len(AllDockables()) })
		return count
	}
	before := dockables()
	spaceOn := func(table unison.Paneler, place func() bool) {
		t.Helper()
		var placed bool
		screen.Do(func() {
			placed = place()
			table.AsPanel().RequestFocus()
		})
		if !placed {
			t.Fatal("the cell cursor must go where it is put")
		}
		screen.KeyPress(unison.KeySpace, mod.None)
	}

	refCol := columnIndexForID(traits, gurps.TraitReferenceColumn)
	if refCol < 0 {
		t.Fatal("the traits must show their page references")
	}
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")
	spaceOn(traits, func() bool { return traits.SetLeadCell(0, refCol) })
	c.Equal([]string{ref}, screen.OpenedURLs(), "space on a page reference opens it")
	c.Equal(before, dockables(), "and opens no editor")

	equippedCol := columnIndexForID(equipment, gurps.EquipmentEquippedColumn)
	if equippedCol < 0 {
		t.Fatal("the carried equipment must show whether it is equipped")
	}
	spaceOn(equipment, func() bool { return equipment.SetLeadCell(0, equippedCol) })
	var equipped bool
	screen.Do(func() { equipped = eq.Equipped })
	c.False(equipped, "space on a check cell unchecks it")
	c.Equal(before, dockables(), "and opens no editor")
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() { equipped = eq.Equipped })
	c.True(equipped, "and checks it again")
	c.Nil(screen.OpenedURLs())

	nameCol := columnIndexForID(traits, gurps.TraitDescriptionColumn)
	spaceOn(traits, func() bool { return traits.SetLeadCell(0, nameCol) })
	c.Equal(before+1, dockables(), "space on a cell with nothing to press opens the editor")
}
