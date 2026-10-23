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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// newReuseTestSheet returns a sheet whose traits list is the named traits, the one at boxIndex (if any) being a
// container holding a trait named "Inside", along with the traits at the top of the list.
func newReuseTestSheet(t *testing.T, boxIndex int, names ...string) (*Sheet, []*gurps.Trait) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traits := make([]*gurps.Trait, len(names))
	for i, name := range names {
		traits[i] = gurps.NewTrait(entity, nil, i == boxIndex)
		traits[i].Name = name
		if i == boxIndex {
			inside := gurps.NewTrait(entity, traits[i], false)
			inside.Name = "Inside"
			traits[i].Children = []*gurps.Trait{inside}
		}
	}
	entity.Traits = traits
	sheet.Rebuild(true)
	return sheet, traits
}

// updateSheet reports a change to the sheet and then lays its page out, as the window would before drawing it. A sync
// sizes the columns to fit the table's preferred width, which is rounded up to whole pixels, so until the page has
// been laid out again the cells are asked for at widths that differ from the settled ones by that rounding.
func updateSheet(sheet *Sheet) {
	sheet.MarkModified(sheet)
	sheet.page.ValidateLayout()
}

// traitDescriptionCells returns each row the traits table shows, in order, along with the panel of its description
// cell.
func traitDescriptionCells(c check.Checker, table *unison.Table[*Node[*gurps.Trait]]) ([]*Node[*gurps.Trait], []*unison.Panel) {
	col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	c.True(col >= 0, "the traits table must have a description column")
	count := table.LastRowIndex() + 1
	rows := make([]*Node[*gurps.Trait], count)
	panels := make([]*unison.Panel, count)
	for i := range count {
		rows[i] = table.RowFromIndex(i)
		panels[i] = rows[i].ColumnCell(i, col, unison.Black, unison.White, false, false, false).AsPanel()
	}
	return rows, panels
}

// TestSyncKeepsThePanelsOfUnchangedRows verifies that a page list keeps its rows, children included, across the sync
// an update makes, and with them the panels built for cells whose data didn't change, which is what spares an update
// the cost of building every cell on the page anew, while a cell whose data did change is built anew to show it.
func TestSyncKeepsThePanelsOfUnchangedRows(t *testing.T) {
	for _, one := range []struct {
		name     string
		names    []string
		boxIndex int // the trait that is a container, or -1
		rows     int
		renamed  int // the trait renamed to "Changed" before the update, or -1
	}{
		{name: "nothing changed", names: []string{"Alpha", "Box", "Gamma"}, boxIndex: 1, rows: 4, renamed: -1},
		{name: "one trait renamed", names: []string{"Alpha", "Beta", "Gamma"}, boxIndex: -1, rows: 3, renamed: 1},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			sheet, traits := newReuseTestSheet(t, one.boxIndex, one.names...)
			if one.boxIndex >= 0 {
				c.True(traits[one.boxIndex].IsOpen(), "the container must be open, so that its child is shown")
			}
			rows, panels := traitDescriptionCells(c, sheet.Traits.Table)
			c.Equal(one.rows, len(rows), "the table must show the traits and any container's child")

			if one.renamed >= 0 {
				traits[one.renamed].Name = "Changed"
			}
			updateSheet(sheet)
			after, afterPanels := traitDescriptionCells(c, sheet.Traits.Table)
			c.Equal(len(rows), len(after), "the update must leave the same rows on show")
			for i := range rows {
				c.True(after[i] == rows[i], "row %d must be the same object after the sync", i)
				if i == one.renamed {
					c.True(afterPanels[i] != panels[i], "the changed cell must be built anew")
					c.True(slices.Contains(uxtest.LabelTexts(afterPanels[i]), "Changed"), "and must show the new name")
				} else {
					c.True(afterPanels[i] == panels[i], "row %d must keep the panel of its unchanged cell", i)
				}
			}
			if one.boxIndex >= 0 {
				c.True(after[one.boxIndex+1].Parent() == after[one.boxIndex],
					"the kept child must still have its container as its parent")
			}
		})
	}
}

// TestMovedRowGetsANewNodeWithTheRightParent verifies that a row whose data moves to another container is given a new
// node there rather than its old one, which still names the old container as its parent, while the rows that stayed
// put are kept.
func TestMovedRowGetsANewNodeWithTheRightParent(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	boxA := gurps.NewTrait(entity, nil, true)
	boxA.Name = "A"
	boxB := gurps.NewTrait(entity, nil, true)
	boxB.Name = "B"
	child := gurps.NewTrait(entity, boxA, false)
	child.Name = "Child"
	boxA.Children = []*gurps.Trait{child}
	entity.Traits = []*gurps.Trait{boxA, boxB}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	c.True(boxA.IsOpen() && boxB.IsOpen(), "the containers must be open, so that their children are shown")
	rowA, oldChild, rowB := table.RowFromIndex(0), table.RowFromIndex(1), table.RowFromIndex(2)
	c.True(oldChild.Data() == child && oldChild.Parent() == rowA, "the child must start out under A")

	boxA.SetChildren(nil)
	boxB.SetChildren([]*gurps.Trait{child})
	child.SetParent(boxB)
	sheet.Rebuild(true)
	c.True(table == sheet.Traits.Table, "the rebuild must have kept the table, or this test covers nothing")
	c.True(table.RowFromIndex(0) == rowA && table.RowFromIndex(1) == rowB, "the containers must have been kept")
	newChild := table.RowFromIndex(2)
	c.True(newChild.Data() == child, "the child must now be shown under B")
	c.True(newChild.Parent() == rowB, "the child's row must have B's row as its parent")
	c.True(newChild != oldChild, "the child must have been given a new node")
	c.Equal(0, len(rowA.Children()), "A's row must no longer hold the child")
}

// TestNotesToggleRebuildsAReusedCell verifies that showing or hiding a row's notes rebuilds its description cell in a
// list that keeps the row across the sync that follows the toggle, as a sheet's list and a filtered library list do.
// The cell's data stays the same, so only the notes disclosure kept with the panel tells the panel built before the
// toggle from the one wanted after it.
func TestNotesToggleRebuildsAReusedCell(t *testing.T) {
	for _, one := range []struct {
		name  string
		table func(*testing.T, check.Checker) *unison.Table[*Node[*gurps.Trait]]
	}{
		{
			name: "sheet",
			table: func(t *testing.T, _ check.Checker) *unison.Table[*Node[*gurps.Trait]] {
				sheet := newTestSheetForTemplate(t)
				entity := sheet.Entity()
				entity.SheetSettings.NotesDisplay = display.Inline
				trait := gurps.NewTrait(entity, nil, false)
				trait.Name = "Charisma"
				trait.LocalNotes = "A note"
				entity.Traits = []*gurps.Trait{trait}
				sheet.Rebuild(true)
				return sheet.Traits.Table
			},
		},
		{
			name: "filtered library list",
			table: func(t *testing.T, c check.Checker) *unison.Table[*Node[*gurps.Trait]] {
				uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
				registerKeyBindingsOnce.Do(func() { registerActions() })
				trait := newTraitForTest("Charisma", nil)
				trait.LocalNotes = "A note"
				d := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{trait})
				d.table.ApplyFilter(func(*Node[*gurps.Trait]) bool { return false })
				c.True(d.table.IsFiltered(), "the table must be filtered")
				return d.table
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			table := one.table(t, c)
			row := table.RowFromIndex(0)
			c.NotNil(row, "the table must show the trait")
			key := row.noteKey()
			wasClosed := gurps.IsClosed(key)
			t.Cleanup(func() { gurps.SetClosedState(key, wasClosed) })
			gurps.SetClosedState(key, false)
			col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
			c.True(col >= 0, "the traits table must have a description column")

			before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
			c.True(findMarkdown(before) != nil, "the cell must show the open notes")
			button, ok := uxtest.FirstPanelOfType[*unison.Button](before)
			c.True(ok, "the cell must hold the notes disclosure button")

			button.ClickCallback()
			c.True(gurps.IsClosed(key), "clicking the button must close the notes")
			c.True(table.RowFromIndex(0) == row, "the list must have kept its row across the sync")
			after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
			c.True(after != before, "the cell must be rebuilt once its notes are closed")
			c.True(findMarkdown(after) == nil, "and must no longer show them")
		})
	}
}

// TestNotesToggleFollowsAConvertedRow verifies that the notes disclosure button of a kept cell acts on the row's notes
// as they are keyed now. Converting an item to or from a container changes its ID, and with it the key its notes'
// disclosure is kept under, while its row and, when the disclosure is the same under both keys, its cell are kept.
func TestNotesToggleFollowsAConvertedRow(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.SheetSettings.NotesDisplay = display.Inline
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = "Backpack"
	eqp.LocalNotes = "A note"
	// Already a container, so that the list leaves room for the disclosure triangle and the conversion doesn't change
	// the width of the cell.
	bag := gurps.NewEquipment(entity, nil, true)
	bag.Name = "Bag"
	entity.CarriedEquipment = []*gurps.Equipment{eqp, bag}
	sheet.Rebuild(true)
	sheet.page.ValidateLayout()
	table := sheet.CarriedEquipment.Table
	row := table.RowFromIndex(0)
	c.NotNil(row, "the table must show the equipment")
	oldKey := row.noteKey()
	col := table.ColumnIndexForID(gurps.EquipmentDescriptionColumn)
	c.True(col >= 0, "the equipment table must have a description column")
	before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()

	table.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
	ConvertToContainer(sheet, table)
	sheet.page.ValidateLayout()
	c.True(eqp.Container(), "the equipment must have become a container")
	c.True(table.RowFromIndex(0) == row, "the list must have kept its row across the conversion")
	newKey := row.noteKey()
	c.NotEqual(oldKey, newKey, "the conversion must have changed the key the notes' disclosure is kept under")
	for _, key := range []string{oldKey, newKey} {
		wasClosed := gurps.IsClosed(key)
		t.Cleanup(func() { gurps.SetClosedState(key, wasClosed) })
		gurps.SetClosedState(key, false)
	}
	kept := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(kept == before, "the cell must have been kept, or this test covers nothing")
	button, ok := uxtest.FirstPanelOfType[*unison.Button](kept)
	c.True(ok, "the cell must hold the notes disclosure button")

	button.ClickCallback()
	c.True(gurps.IsClosed(newKey), "clicking the button must close the notes of the converted row")
	c.False(gurps.IsClosed(oldKey), "and must leave the key the row no longer has alone")
	after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(findMarkdown(after) == nil, "the cell must no longer show the notes")
}

// TestFontChangeRebuildsCells verifies that a change of fonts rebuilds the panels a row keeps, although the data of
// its cells is unchanged, since their text was wrapped and measured with the old fonts.
func TestFontChangeRebuildsCells(t *testing.T) {
	c := check.New(t)
	sheet, _ := newReuseTestSheet(t, -1, "Alpha")
	rows, panels := traitDescriptionCells(c, sheet.Traits.Table)
	again, againPanels := traitDescriptionCells(c, sheet.Traits.Table)
	c.True(again[0] == rows[0] && againPanels[0] == panels[0], "until the fonts change, the cell must be kept")

	invalidateCellPanels()
	after, afterPanels := traitDescriptionCells(c, sheet.Traits.Table)
	c.True(after[0] == rows[0], "the row must be the same object")
	c.True(afterPanels[0] != panels[0], "the cell must be built anew once the fonts change")
	c.True(slices.Contains(uxtest.LabelTexts(afterPanels[0]), "Alpha"), "and must still show the trait")
}

// TestRootRowsReturnsTheTablesOwnRows verifies that a page list's provider hands out the rows the table already shows
// rather than new ones, while a provider without reuse, as off a page, builds new rows on every call.
func TestRootRowsReturnsTheTablesOwnRows(t *testing.T) {
	c := check.New(t)
	sheet, traits := newReuseTestSheet(t, -1, "Alpha", "Beta")
	table := sheet.Traits.Table
	roots := table.RootRows()
	c.Equal(2, len(roots), "the table must have both traits at the top")
	c.True(roots[0] == table.RowFromIndex(0) && roots[1] == table.RowFromIndex(1),
		"the provider must return the rows the table shows")

	lists := &listsForTest{traits: traits}
	for _, p := range []TableProvider[*gurps.Trait]{NewTraitsProvider(lists, false), NewTraitsProvider(lists, true)} {
		if tp, ok := p.(*traitsProvider); ok {
			// Reuse switched off for the page provider too, which must then behave as one off a page.
			tp.reuse = nil
		}
		p.SetTable(unison.NewTable[*Node[*gurps.Trait]](p))
		first, second := p.RootRows(), p.RootRows()
		c.Equal(2, len(first), "the provider must return a row for each trait")
		for i := range first {
			c.True(first[i].Data() == traits[i] && second[i].Data() == traits[i], "row %d must hold its trait", i)
			c.True(first[i] != second[i], "without reuse, row %d must be built anew on each call", i)
		}
	}
}
