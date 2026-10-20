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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newCellCacheTestSheet returns a sheet holding a single trait of the given name, along with the row the traits table
// shows it in and the index of the table's description column.
func newCellCacheTestSheet(t *testing.T, c check.Checker, name string) (*gurps.Trait, *Node[*gurps.Trait], int) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	c.True(col >= 0, "the traits table must have a description column")
	row := table.RowFromIndex(0)
	c.NotNil(row, "the traits table must show the trait")
	return trait, row, col
}

// TestNodeKeepsCellDataUntilAChangeIsReported verifies that a row derives its cell data once per reported change
// rather than on every call, which is what spares a draw or a sizing pass the cost of the model, and that reporting a
// change is all it takes for the next call to show it.
func TestNodeKeepsCellDataUntilAChangeIsReported(t *testing.T) {
	c := check.New(t)
	trait, row, col := newCellCacheTestSheet(t, c, "Before")
	before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(slices.Contains(uxtest.LabelTexts(before), "Before"), "the cell must show the trait's name")

	trait.Name = "After"
	again := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(again == before, "until a change is reported, the cell must come back as it was")
	c.True(slices.Contains(uxtest.LabelTexts(again), "Before"), "and must still show the name it was built with")

	invalidateDisplayData()
	after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(after != before, "once a change is reported, the cell must be built afresh")
	c.True(slices.Contains(uxtest.LabelTexts(after), "After"), "and must show the new name")
}

// TestSizingPassesReuseCellPanels verifies that the panel built for a cell's width survives the sizing passes, which
// ask for every cell at its natural size before settling on the widths, so that measuring the table and then drawing
// it doesn't build every cell twice.
func TestSizingPassesReuseCellPanels(t *testing.T) {
	c := check.New(t)
	_, row, col := newCellCacheTestSheet(t, c, "Trait")
	table := row.table
	// The sheet is never laid out, so the column is given a width of its own.
	const width = 300
	table.Columns[col].Current = width
	c.True(table.CellWidth(0, col) > 0, "the cell must have a width to fit")
	fitted := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false)
	c.True(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false) == fitted,
		"asking for the cell again at the same width must reuse its panel")

	table.Columns[col].Current = 0
	c.True(table.CellWidth(0, col) <= 0, "a sizing pass asks for the cell without a width")
	natural := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false)
	c.True(natural != fitted, "the natural size must come from a panel of its own")
	c.True(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false) == natural,
		"asking for the natural size again must reuse that panel")

	table.Columns[col].Current = width
	c.True(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false) == fitted,
		"the panel built for the width must have survived the sizing pass")
}

// TestCellKeepsPanelsForFourWidths verifies that a cell keeps the panels built for four widths, since a sync of a page
// list asks for the cells of the column taking the excess width at four: before and after the page reference columns
// take their share, both at the table's preferred width, rounded up to whole pixels, and at the width the layout then
// gives it. Keeping fewer would build those cells anew several times on every sync.
func TestCellKeepsPanelsForFourWidths(t *testing.T) {
	c := check.New(t)
	_, row, col := newCellCacheTestSheet(t, c, "Trait")
	table := row.table
	widths := []float32{300, 300.25, 250, 250.25}
	panels := make([]unison.Paneler, len(widths))
	for i, width := range widths {
		table.Columns[col].Current = width
		panels[i] = row.ColumnCell(0, col, unison.Black, unison.White, false, false, false)
	}
	for i, width := range widths {
		table.Columns[col].Current = width
		c.True(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false) == panels[i],
			"the panel built for width %v must have been kept", width)
	}
}

// TestLibraryListShowsQuantityAdjustmentWithoutSync verifies that adjusting a quantity in an equipment library list
// shows the new quantity on the next draw. Marking a library list as modified only updates its title, so the table
// keeps its rows, and it is the report of the change that has them derive their cell data afresh.
func TestLibraryListShowsQuantityAdjustmentWithoutSync(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	item := gurps.NewEquipment(nil, nil, false)
	item.Name = "Rope"
	item.Quantity = fxp.One
	d := NewEquipmentTableDockable("test"+gurps.EquipmentExt, []*gurps.Equipment{item})
	col := d.table.ColumnIndexForID(gurps.EquipmentQuantityColumn)
	c.True(col >= 0, "the equipment table must have a quantity column")
	row := d.table.RowFromIndex(0)
	c.NotNil(row, "the equipment table must show the item")
	c.Equal([]string{"1"}, uxtest.LabelTexts(row.ColumnCell(0, col, unison.Black, unison.White, false, false,
		false).AsPanel()), "the quantity cell must show the starting quantity")

	d.table.SelectByIndex(0)
	adjustQuantity(d, d.table, true)
	c.Equal(fxp.Two, item.Quantity, "the quantity must have been incremented")
	c.True(d.table.RowFromIndex(0) == row, "nothing may have synced the table, or this test covers nothing")
	c.Equal([]string{"2"}, uxtest.LabelTexts(row.ColumnCell(0, col, unison.Black, unison.White, false, false,
		false).AsPanel()), "the same row must show the new quantity")
}

// TestNotesToggleRebuildsTheCellOfAFilteredRow verifies that showing or hiding a row's notes rebuilds its description
// cell in a filtered table. A filtered table keeps its rows across a sync, and the cell's data doesn't change with the
// disclosure, so only the disclosure state kept with the cell tells the panel built before the toggle from the one
// wanted after it.
func TestNotesToggleRebuildsTheCellOfAFilteredRow(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	trait := newTraitForTest("Charisma", nil)
	trait.LocalNotes = "A note"
	d := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{trait})
	d.table.ApplyFilter(func(*Node[*gurps.Trait]) bool { return false })
	c.True(d.table.IsFiltered(), "the table must be filtered")
	row := d.table.RowFromIndex(0)
	c.NotNil(row, "the filter must keep the trait")
	key := row.noteKey()
	wasClosed := gurps.IsClosed(key)
	t.Cleanup(func() { gurps.SetClosedState(key, wasClosed) })
	gurps.SetClosedState(key, false)
	col := d.table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	c.True(col >= 0, "the traits table must have a description column")

	before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(findMarkdown(before) != nil, "the cell must show the open notes")
	button, ok := uxtest.FirstPanelOfType[*unison.Button](before)
	c.True(ok, "the cell must hold the notes disclosure button")

	button.ClickCallback()
	c.True(gurps.IsClosed(key), "clicking the button must close the notes")
	c.True(d.table.RowFromIndex(0) == row, "the filtered table must have kept its row across the sync")
	after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(after != before, "the cell must be rebuilt once its notes are closed")
	c.True(findMarkdown(after) == nil, "and must no longer show them")
}
