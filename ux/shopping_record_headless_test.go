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
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestShoppingRecordHeadless drives a sheet's shopping record: the sheet's toolbar button opens it; it lists the
// entries most recent first with the agreed cell texts; Delete removes the selected entries as one edit on the sheet's
// undo manager, which Edit > Undo and Redo reach while the record has the focus; Select All then Delete empties it;
// changes made to the record from elsewhere show up with the selection kept; showing it again brings the open one
// forward; and it closes with its sheet.
func TestShoppingRecordHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	base := time.Date(2026, time.October, 13, 9, 40, 0, 0, time.Local)
	knives := &gurps.ShoppingRecordEntry{
		When:      jio.Time(base.Add(25 * time.Minute)),
		Name:      "Sold small knives",
		Kind:      gurps.ShoppingSold,
		Items:     []gurps.ShoppingRecordItem{{Name: "Small Knife", Quantity: fxp.Two, Value: fxp.FromInteger(30)}},
		Exchanged: fxp.FromInteger(31),
		Money: []gurps.ShoppingRecordItem{
			{Name: "Silver Eagle", Quantity: fxp.One, Value: fxp.FromInteger(20)},
			{Name: "Copper Penny", Quantity: fxp.FromInteger(11), Value: fxp.One},
		},
	}
	rope := &gurps.ShoppingRecordEntry{
		When: jio.Time(base),
		Name: "Bought rope",
		Kind: gurps.ShoppingBought,
		Items: []gurps.ShoppingRecordItem{
			{Name: "Rope (50 ft)", Quantity: fxp.One, Value: fxp.FromInteger(90)},
			{Name: "Torch", Quantity: fxp.FromInteger(4), Value: fxp.Two},
		},
		Exchanged: fxp.FromInteger(98),
		Money: []gurps.ShoppingRecordItem{
			{Name: "Silver Eagle", Quantity: fxp.FromInteger(-5), Value: fxp.FromInteger(20)},
			{Name: "Copper Penny", Quantity: fxp.Two, Value: fxp.One},
		},
	}
	shortsword := &gurps.ShoppingRecordEntry{
		When:      jio.Time(base.Add(-24 * time.Hour)),
		Name:      "Sold shortsword",
		Kind:      gurps.ShoppingSold,
		Items:     []gurps.ShoppingRecordItem{{Name: "Shortsword", Quantity: fxp.One, Value: fxp.FromInteger(400)}},
		Exchanged: fxp.FromInteger(200),
		Money:     []gurps.ShoppingRecordItem{{Name: "Silver Eagle", Quantity: fxp.FromInteger(10), Value: fxp.FromInteger(20)}},
	}
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Profile.Name = "Rudolf"
		// Added out of order, since the record keeps itself most recent first.
		entity.AddShoppingRecord(rope)
		entity.AddShoppingRecord(shortsword)
		entity.AddShoppingRecord(knives)
		sheet.Rebuild(true)
		// A new sheet counts as modified until saved, so it is marked unmodified for the test to tell whether removing
		// entries changed it.
		sheet.markUnmodified()
	})

	isRecord := func(d unison.Dockable) bool {
		_, isOne := d.AsPanel().Self.(*shoppingRecordDockable)
		return isOne
	}
	var recordButton *unison.Button
	screen.Do(func() { recordButton = buttonWithTooltip(sheet.toolbar, "Shopping Record") })
	if recordButton == nil {
		t.Fatal("the sheet's toolbar must have a Shopping Record button")
	}
	screen.Click(screen.PanelCenter(recordButton))
	record := uxtest.SoleEditor[*shoppingRecordDockable](t, screen, isRecord)

	// recordedNames returns the names of the sheet's record entries, in the order the sheet keeps them.
	recordedNames := func() []string {
		var names []string
		screen.Do(func() {
			list := sheet.Entity().ShoppingRecord
			names = make([]string, len(list))
			for i, entry := range list {
				names[i] = entry.Name
			}
		})
		return names
	}
	// shownRows returns the text of every cell the record's table shows, row by row.
	shownRows := func() [][]string {
		var rows [][]string
		screen.Do(func() {
			rows = make([][]string, record.table.LastRowIndex()+1)
			for i := range rows {
				row := record.table.RowFromIndex(i)
				rows[i] = make([]string, len(record.table.Columns))
				for col := range rows[i] {
					rows[i][col] = row.CellDataForSort(col)
				}
			}
		})
		return rows
	}
	// shownNames returns the transactions the record's table shows, in order.
	shownNames := func() []string {
		rows := shownRows()
		names := make([]string, len(rows))
		for i, row := range rows {
			names[i] = row[shoppingRecordTransactionColumn]
		}
		return names
	}
	// selectedNames returns the transactions selected in the record's table, in order.
	selectedNames := func() []string {
		var names []string
		screen.Do(func() {
			selected := record.table.SelectedRows(false)
			names = make([]string, len(selected))
			for i, row := range selected {
				names[i] = row.CellDataForSort(shoppingRecordTransactionColumn)
			}
		})
		return names
	}
	// state returns whether the sheet is modified, whether the record's table has the focus and the sheet's undo title.
	state := func() (modified, focused bool, undoTitle string) {
		screen.Do(func() {
			modified = sheet.Modified()
			focused = record.table.Focused()
			undoTitle = sheet.undoMgr.UndoTitle()
		})
		return modified, focused, undoTitle
	}

	var title, accessibleName, cellText string
	screen.Do(func() {
		title = record.Title()
		accessibleName = record.table.Accessibility.Name
		if label, isLabel := record.table.RowFromIndex(1).ColumnCell(1, shoppingRecordMoneyColumn, unison.ThemeOnSurface,
			unison.ThemeSurface, false, false, false).(*unison.Label); isLabel {
			cellText = label.String()
		}
	})
	c.Equal("Shopping Record: Rudolf", title)
	c.Equal("Shopping Record", accessibleName)
	c.Equal([][]string{
		{
			knives.When.String(), "Sold small knives", "Sold", "Small Knife ×2",
			"$31 (+1 Silver Eagle, +11 Copper Penny)",
		},
		{
			rope.When.String(), "Bought rope", "Bought", "Rope (50 ft) ×1, Torch ×4",
			"$98 (−5 Silver Eagle, +2 Copper Penny)",
		},
		{shortsword.When.String(), "Sold shortsword", "Sold", "Shortsword ×1", "$200 (+10 Silver Eagle)"},
	}, shownRows(), "the entries are listed most recent first, with signed money")
	c.Equal("$98 (−5 Silver Eagle, +2 Copper Penny)", cellText, "a cell shows its text in a label")
	_, focused, _ := state()
	c.True(focused, "the record's table takes the focus when shown")
	audit := uxtest.NewAXNameAudit(t, screen, wnd)
	audit.Check("Shopping Record", record)
	audit.Report()

	// A click on the second row selects it, and Delete removes it from the sheet as one edit.
	var pt geom.Point
	var visible bool
	screen.Do(func() {
		record.table.ValidateScrollRoot()
		rootPt := record.table.PointToRoot(record.table.RowFrame(1).Center())
		visible = rootPt.In(visibleRect(record.table.AsPanel()))
		pt = screenPoint(wnd, rootPt)
	})
	if !visible {
		t.Fatal("the second row of the record must be in view to be clicked")
	}
	screen.Click(pt)
	c.Equal([]string{"Bought rope"}, selectedNames(), "the click selects the row")
	screen.KeyPress(unison.KeyDelete, mod.None)
	c.Equal([]string{"Sold small knives", "Sold shortsword"}, recordedNames(), "Delete removes the entry")
	c.Equal([]string{"Sold small knives", "Sold shortsword"}, shownNames(), "the removed row is gone")
	modified, focused, undoTitle := state()
	c.True(modified, "removing an entry modifies the sheet")
	c.True(focused, "the record keeps the focus")
	c.Equal("Undo Remove Shopping Record Entry", undoTitle)

	// Edit > Undo and Redo reach the sheet's undo manager from the record.
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Undo Remove Shopping Record Entry")
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, recordedNames(), "undo restores the entry")
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, shownNames(), "undo shows the entry again")
	modified, focused, _ = state()
	c.False(modified, "undo leaves the sheet as it was")
	c.True(focused, "the record keeps the focus through undo")
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Redo Remove Shopping Record Entry")
	c.Equal([]string{"Sold small knives", "Sold shortsword"}, recordedNames(), "redo removes the entry again")
	c.Equal([]string{"Sold small knives", "Sold shortsword"}, shownNames(), "redo takes the row away again")
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, recordedNames(),
		"the Undo key restores the entry")

	// Select All then Delete empties the record in one edit, and one undo brings everything back.
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, selectedNames(), "Select All selects all")
	screen.KeyPress(unison.KeyBackspace, mod.None)
	c.Equal(0, len(recordedNames()), "Delete removes every selected entry")
	c.Equal(0, len(shownRows()), "the table is emptied")
	_, _, undoTitle = state()
	c.Equal("Undo Remove Shopping Record Entries", undoTitle)
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Undo Remove Shopping Record Entries")
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, recordedNames(), "undo restores all")
	c.Equal([]string{"Sold small knives", "Bought rope", "Sold shortsword"}, shownNames(), "undo shows all again")

	// A change made to the record elsewhere, such as undoing a sale, is shown with the remaining selection kept, and
	// the title follows the character's name.
	screen.Do(func() {
		record.table.SetSelectionMap(map[tid.TID]bool{knives.ID: true, shortsword.ID: true})
		entity := sheet.Entity()
		entity.SetShoppingRecord(slices.DeleteFunc(gurps.CloneShoppingRecordList(entity.ShoppingRecord),
			func(entry *gurps.ShoppingRecordEntry) bool { return entry.ID == shortsword.ID }))
		entity.Profile.Name = "Rudolf Vautour"
		rebuildAsModified(sheet, false)
	})
	c.Equal([]string{"Sold small knives", "Bought rope"}, shownNames(), "the rows follow the sheet")
	c.Equal([]string{"Sold small knives"}, selectedNames(), "the remaining selection is kept")
	screen.Do(func() { title = record.title })
	c.Equal("Shopping Record: Rudolf Vautour", title, "the title follows the character's name")

	// Showing the record again brings the open one forward rather than opening another.
	screen.Do(func() { showShoppingRecord(sheet) })
	again := uxtest.SoleEditor[*shoppingRecordDockable](t, screen, isRecord)
	c.True(again == record, "the open record is reused")

	// Closing the sheet closes its record.
	screen.Do(func() { sheet.markUnmodified() })
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
	var open int
	screen.Do(func() {
		for _, d := range AllDockables() {
			if isRecord(d) {
				open++
			}
		}
	})
	c.Equal(0, open, "the record closes with its sheet")
}
