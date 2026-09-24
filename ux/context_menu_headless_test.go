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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// These tests ask for GCS's context menus every way a user can -- a right-click, shift+F10, the Menu key and a screen
// reader's request -- and check that the menu each opens is the one the panel builds. A headless session uses in-window
// menus, so an open one can be seen with openMenuPopup and closed with Escape.

// contextMenuTitles returns the titles of the items of the popup menu open in wnd, without the separators, or nil when
// none is open. An in-window menu item draws its title rather than holding it in a panel, so the titles are read from
// the menu's accessibility description; the test ends if accessibility support is off, since every check of the titles
// would otherwise fail as though no menu had opened.
func contextMenuTitles(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window) []string {
	t.Helper()
	var items []*unison.Panel
	screen.Do(func() { items = slices.Clone(menuItemPanels(openMenuPopup(wnd))) })
	if len(items) == 0 {
		return nil
	}
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("accessibility support must be on for the menu's items to be read")
	}
	titles := make([]string, 0, len(items))
	for _, item := range items {
		if node := screen.AccessibilityNodeFor(item); node != nil && node.Role == role.MenuItem {
			titles = append(titles, node.Name)
		}
	}
	return titles
}

// closeContextMenu closes the popup menu open in wnd the way a user would, with Escape, and checks that it has gone.
func closeContextMenu(c check.Checker, screen *unison.HeadlessScreen, wnd *unison.Window) {
	c.Helper()
	screen.KeyPress(unison.KeyEscape, mod.None)
	var open bool
	screen.Do(func() { open = openMenuPopup(wnd) != nil })
	c.False(open, "Escape must close the menu")
}

// cellScreenPoint brings the given row into view and returns the middle of its cell in the given column on the screen,
// ending the test if the cell cannot be brought into view.
func cellScreenPoint[T gurps.Node[T]](t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window,
	table *unison.Table[*Node[T]], row, columnID int,
) geom.Point {
	t.Helper()
	var pt geom.Point
	var visible bool
	screen.Do(func() {
		table.ScrollRowIntoView(row)
		table.ValidateScrollRoot()
		frame := table.CellFrame(row, table.ColumnIndexForID(columnID))
		rootPt := table.PointToRoot(frame.Center())
		visible = rootPt.In(visibleRect(table.AsPanel()))
		pt = screenPoint(wnd, rootPt)
	})
	if !visible {
		t.Fatalf("row %d of the list must be in view to be clicked", row)
	}
	return pt
}

// TestSheetListContextMenuHeadless verifies that a list on a character sheet offers its context menu every way a user
// can ask for it. A right-click on one of several selected rows keeps them all selected for the menu's commands to act
// on, while a right-click or a screen reader's request on a row that is not selected makes it the selection alone, and
// either puts the focus on the list, since the menu's commands act on what holds it.
func TestSheetListContextMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	names := []string{"Alpha", "Beta", "Gamma"}
	var table *unison.Table[*Node[*gurps.Trait]]
	screen.Do(func() {
		entity := sheet.Entity()
		for _, name := range names {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = name
			entity.Traits = append(entity.Traits, trait)
		}
		entity.Recalculate()
		sheet.Rebuild(true)
		table = sheet.Traits.Table
	})
	if table == nil {
		t.Fatal("the sheet must have a traits list")
	}
	selectedNames := func() []string {
		var selected []string
		screen.Do(func() {
			for _, row := range table.SelectedRows(false) {
				selected = append(selected, row.Data().Name)
			}
		})
		slices.Sort(selected)
		return selected
	}
	tableFocused := func() bool {
		var focused bool
		screen.Do(func() { focused = wnd.Focus() == table.AsPanel() })
		return focused
	}
	// rowPoint brings the named trait's row into view and returns the middle of its description cell on the screen,
	// clear of the check box at the front of the row that a click would toggle. Rows are found by name, since a new
	// sheet comes with a trait of its own.
	rowPoint := func(name string) geom.Point {
		t.Helper()
		row := -1
		screen.Do(func() {
			for i := 0; i <= table.LastRowIndex(); i++ {
				if table.RowFromIndex(i).Data().Name == name {
					row = i
					break
				}
			}
		})
		if row < 0 {
			t.Fatalf("the traits list must have a row for %s", name)
		}
		return cellScreenPoint(t, screen, wnd, table, row, gurps.TraitDescriptionColumn)
	}
	checkMenu := func(how string) {
		t.Helper()
		titles := contextMenuTitles(t, screen, wnd)
		c.NotNil(titles, "%s must open the list's menu", how)
		c.True(slices.Contains(titles, openEditorAction.Title), "the menu %s opens must offer %q; it offers %v", how,
			openEditorAction.Title, titles)
		c.True(slices.Contains(titles, duplicateAction.Title), "the menu %s opens must offer %q; it offers %v", how,
			duplicateAction.Title, titles)
	}

	screen.Click(rowPoint("Alpha"))
	c.Equal([]string{"Alpha"}, selectedNames(), "clicking a row selects it")
	c.True(tableFocused(), "clicking a row focuses the list")

	screen.KeyPress(unison.KeyF10, mod.Shift)
	checkMenu("shift+F10")
	c.Equal([]string{"Alpha"}, selectedNames(), "opening the menu leaves the selection alone")
	closeContextMenu(c, screen, wnd)
	screen.KeyPress(unison.KeyMenu, mod.None)
	checkMenu("the Menu key")
	closeContextMenu(c, screen, wnd)

	screen.ClickWith(rowPoint("Beta"), unison.ButtonLeft, mod.Shift)
	c.Equal([]string{"Alpha", "Beta"}, selectedNames(), "shift-clicking the next row extends the selection to it")
	screen.ClickWith(rowPoint("Beta"), unison.ButtonRight, mod.None)
	checkMenu("a right-click on a selected row")
	c.Equal([]string{"Alpha", "Beta"}, selectedNames(),
		"a right-click on one of several selected rows must keep them all selected for the menu to act on")
	closeContextMenu(c, screen, wnd)

	// A screen reader asks a row that is not selected for the menu while the focus is elsewhere. Only rows that can be
	// seen, are selected or hold the focus are described, so the row is brought into view first.
	var navigatorFocused bool
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		navigatorFocused = wnd.Focus() == Workspace.Navigator.table.AsPanel()
	})
	c.True(navigatorFocused, "the test needs the focus to start out in the navigator")
	rowPoint("Gamma")
	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the workspace window must be described")
	}
	tableNode := screen.AccessibilityNodeFor(table)
	if tableNode == nil {
		t.Fatal("the traits list must be described")
	}
	var gamma *accessibility.Node
	for _, id := range tableNode.Children {
		if node := tree.Node(id); node != nil && node.Role == role.Row && strings.Contains(node.Name, "Gamma") {
			gamma = node
		}
	}
	if gamma == nil {
		t.Fatal("the row for Gamma must be described")
	}
	c.False(gamma.Selected, "the Gamma row must not start out selected")
	c.True(gamma.Actions.Has(accessibility.ShowContextMenu), "a row of the list must offer its context menu")
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{
		Node:   gamma.ID,
		Action: accessibility.ShowContextMenu,
	}), "the row must carry out the request for its context menu")
	checkMenu("a screen reader's request")
	c.Equal([]string{"Gamma"}, selectedNames(), "asking a row that is not selected for the menu makes it the selection")
	c.True(tableFocused(), "asking a row for the menu puts the focus on the list its commands act on")
	var menuNodes int
	screen.AccessibilityTree(wnd).Walk(func(n *accessibility.Node) bool {
		if n.Role == role.Menu {
			menuNodes++
		}
		return true
	})
	c.Equal(1, menuNodes, "the open menu must be described to the screen reader")
	closeContextMenu(c, screen, wnd)

	// A right-click on a row outside a selection of several, made while the focus is elsewhere, makes that row the
	// whole selection and puts the focus on the list, although unison delivers the list no mouse event for it.
	screen.ClickWith(rowPoint("Beta"), unison.ButtonLeft, mod.Shift)
	c.Equal([]string{"Beta", "Gamma"}, selectedNames(), "shift-clicking the row before extends the selection to it")
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		navigatorFocused = wnd.Focus() == Workspace.Navigator.table.AsPanel()
	})
	c.True(navigatorFocused, "the test needs the focus to be in the navigator again")
	screen.ClickWith(rowPoint("Alpha"), unison.ButtonRight, mod.None)
	checkMenu("a right-click on a row that is not selected")
	c.Equal([]string{"Alpha"}, selectedNames(), "a right-click on a row that is not selected makes it the selection")
	c.True(tableFocused(), "a right-click on a row puts the focus on the list its menu's commands act on")
	closeContextMenu(c, screen, wnd)
}

// TestSheetLayoutEditorContextMenuHeadless verifies that the block layout editor's menu opens for a right-click on a
// block, offering to hide that block, and for shift+F10 while the editor has the focus.
func TestSheetLayoutEditorContextMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	editLayout := i18n.Text("Edit Layout")
	var pt geom.Point
	var editing, visible, overlayFocused bool
	screen.Do(func() {
		sheet.toggleLayoutEditing()
		editing = sheet.layoutEditing()
		if !editing {
			return
		}
		overlay := sheet.layoutEditor.overlay
		overlayFocused = wnd.Focus() == overlay
		leaf := sheet.layoutEditor.ensureRegions().leafFor(gurps.BlockTraitsKey)
		if leaf == nil {
			return
		}
		overlay.ScrollRectIntoView(leaf.rect)
		overlay.ValidateScrollRoot()
		leaf = sheet.layoutEditor.ensureRegions().leafFor(gurps.BlockTraitsKey)
		rootPt := overlay.PointToRoot(leaf.rect.Center())
		visible = rootPt.In(visibleRect(overlay))
		pt = screenPoint(wnd, rootPt)
	})
	t.Cleanup(func() {
		screen.Do(func() {
			if sheet.layoutEditing() {
				sheet.toggleLayoutEditing()
			}
		})
	})
	if !editing {
		t.Fatal("the sheet must be in layout editing mode")
	}
	c.True(overlayFocused, "the layout editor takes the focus")
	if !visible {
		t.Fatal("the traits block must be in view to be clicked")
	}

	screen.ClickWith(pt, unison.ButtonRight, mod.None)
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, editLayout), "a right-click on a block must open the layout menu; it offers %v",
		titles)
	hideTraits := fmt.Sprintf(i18n.Text("Hide %s"), gurps.BlockTitle(gurps.BlockTraitsKey))
	c.True(slices.Contains(titles, hideTraits), "the menu must offer to hide the block clicked on; it offers %v",
		titles)
	closeContextMenu(c, screen, wnd)

	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles = contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, editLayout), "shift+F10 must open the layout menu; it offers %v", titles)
	closeContextMenu(c, screen, wnd)
	screen.Do(func() { editing = sheet.layoutEditing() })
	c.True(editing, "closing the menu with Escape must leave layout editing on")
}

// TestSheetLayoutEditorMenuDuringADragHeadless verifies that shift+F10 in the middle of a divider drag ends the drag
// where the pointer last was, as one undoable edit, rather than at the release unison delivers off the page before
// asking for the menu, and then opens the menu.
func TestSheetLayoutEditorMenuDuringADragHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var start geom.Point
	var editing, visible bool
	var editor *sheetLayoutEditor
	screen.Do(func() {
		sheet.toggleLayoutEditing()
		editing = sheet.layoutEditing()
		if !editing {
			return
		}
		editor = sheet.layoutEditor
		overlay := editor.overlay
		divider := findTestDivider(editor.ensureRegions(), gurps.BlockTraitsKey)
		if divider == nil {
			return
		}
		overlay.ScrollRectIntoView(divider.rect)
		overlay.ValidateScrollRoot()
		divider = findTestDivider(editor.ensureRegions(), gurps.BlockTraitsKey)
		rootPt := overlay.PointToRoot(divider.rect.Center())
		visible = rootPt.In(visibleRect(overlay))
		start = screenPoint(wnd, rootPt)
	})
	t.Cleanup(func() {
		screen.Do(func() {
			if sheet.layoutEditing() {
				sheet.toggleLayoutEditing()
			}
		})
	})
	if !editing {
		t.Fatal("the sheet must be in layout editing mode")
	}
	if !visible {
		t.Fatal("the divider between the traits and skills blocks must be in view to be dragged")
	}
	traitsWeight := func() fxp.Int {
		var weight fxp.Int
		screen.Do(func() {
			traits, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockTraitsKey)
			weight = traits.Weight
		})
		return weight
	}
	mode := func() layoutEditMode {
		var m layoutEditMode
		screen.Do(func() { m = editor.mode })
		return m
	}

	screen.MouseDown(start, unison.ButtonLeft, mod.None)
	screen.MouseMove(geom.NewPoint(start.X+40, start.Y), mod.None)
	c.Equal(layoutDraggingDivider, mode(), "pressing on the divider and moving must be a divider drag")
	dragged := traitsWeight()
	c.True(dragged > fxp.One, "the drag must have widened the traits block")

	screen.KeyPress(unison.KeyF10, mod.Shift)
	c.Equal(layoutIdle, mode(), "shift+F10 must end the drag")
	c.Equal(dragged, traitsWeight(), "the divider must stay where the drag last put it")
	var edits int
	screen.Do(func() { edits = undoEditCount(sheet.UndoManager()) })
	c.Equal(1, edits, "the resize must be recorded as one undoable edit")
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, i18n.Text("Edit Layout")), "shift+F10 must then open the layout menu; it offers %v",
		titles)
	closeContextMenu(c, screen, wnd)
	// The window has already seen the release, so letting go of the button changes nothing.
	screen.MouseUp(geom.NewPoint(start.X+40, start.Y), unison.ButtonLeft, mod.None)
	c.Equal(layoutIdle, mode(), "the release of a button the window no longer holds down must change nothing")
	c.Equal(dragged, traitsWeight(), "and must leave the layout alone")
	screen.Do(func() { edits = undoEditCount(sheet.UndoManager()) })
	c.Equal(1, edits, "and record nothing")
}

// TestCheckCellRightClickOpensMenuHeadless verifies that a right-click on the check box at the front of a row -- a
// trait's switch or an item's equipped box -- opens the list's menu for that row without toggling the box or recording
// anything to undo.
func TestCheckCellRightClickOpensMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var trait *gurps.Trait
	var eqp *gurps.Equipment
	var mgr *unison.UndoManager
	traitRow, eqpRow := -1, -1
	var switchColumnFirst, equippedColumnFirst bool
	screen.Do(func() {
		entity := sheet.Entity()
		trait = newSwitchableTrait(entity, "Claws")
		entity.Traits = append(entity.Traits, trait)
		eqp = gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Equipped = false
		entity.CarriedEquipment = append(entity.CarriedEquipment, eqp)
		entity.Recalculate()
		sheet.Rebuild(true)
		mgr = sheet.UndoManager()
		traits := sheet.Traits.Table
		switchColumnFirst = traits.Columns[0].ID == gurps.TraitSwitchColumn
		for i := 0; i <= traits.LastRowIndex(); i++ {
			if traits.RowFromIndex(i).Data() == trait {
				traitRow = i
			}
		}
		carried := sheet.CarriedEquipment.Table
		equippedColumnFirst = carried.Columns[0].ID == gurps.EquipmentEquippedColumn
		for i := 0; i <= carried.LastRowIndex(); i++ {
			if carried.RowFromIndex(i).Data() == eqp {
				eqpRow = i
			}
		}
	})
	c.True(switchColumnFirst, "the switch column must come first in the traits list")
	c.True(equippedColumnFirst, "the equipped column must come first in the carried equipment list")
	if traitRow < 0 || eqpRow < 0 {
		t.Fatal("the lists must have rows for the trait and the equipment")
	}
	c.NotNil(mgr, "the sheet must have an undo manager")

	screen.ClickWith(cellScreenPoint(t, screen, wnd, sheet.Traits.Table, traitRow, gurps.TraitSwitchColumn),
		unison.ButtonRight, mod.None)
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, openEditorAction.Title),
		"a right-click on the switch must open the traits list's menu; it offers %v", titles)
	var switchedOn, canUndo bool
	var selected []*gurps.Trait
	var bonus fxp.Int
	screen.Do(func() {
		switchedOn = trait.SwitchedOn
		bonus = stBonusFor(sheet.Entity())
		canUndo = mgr.CanUndo()
		for _, row := range sheet.Traits.Table.SelectedRows(false) {
			selected = append(selected, row.Data())
		}
	})
	c.False(switchedOn, "a right-click on the switch must not throw it")
	c.Equal(fxp.Int(0), bonus, "nor bring the switchable bonus into play")
	c.False(canUndo, "nor record anything to undo")
	c.Equal([]*gurps.Trait{trait}, selected, "a right-click on the switch must select its row for the menu")
	closeContextMenu(c, screen, wnd)

	screen.ClickWith(cellScreenPoint(t, screen, wnd, sheet.CarriedEquipment.Table, eqpRow,
		gurps.EquipmentEquippedColumn), unison.ButtonRight, mod.None)
	titles = contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, openEditorAction.Title),
		"a right-click on the equipped box must open the equipment list's menu; it offers %v", titles)
	var equipped bool
	var selectedEqp []*gurps.Equipment
	screen.Do(func() {
		equipped = eqp.Equipped
		canUndo = mgr.CanUndo()
		for _, row := range sheet.CarriedEquipment.Table.SelectedRows(false) {
			selectedEqp = append(selectedEqp, row.Data())
		}
	})
	c.False(equipped, "a right-click on the equipped box must not equip the item")
	c.False(canUndo, "nor record anything to undo")
	c.Equal([]*gurps.Equipment{eqp}, selectedEqp, "a right-click on the equipped box must select its row for the menu")
	closeContextMenu(c, screen, wnd)
}

// TestNavigatorContextMenuHeadless verifies that shift+F10 opens the library navigator's menu for the selected file,
// and that it opens nothing while nothing is selected.
func TestNavigatorContextMenuHeadless(t *testing.T) {
	c := check.New(t)
	// Every library row starts out disclosed, so that the file put in the user library below is a row of its own.
	swapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, wnd := startHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	user := gurps.GlobalSettings().Libraries.User()
	c.NoError(os.MkdirAll(user.Path(), 0o750))
	path := filepath.Join(user.Path(), "test"+gurps.SheetExt)
	c.NoError(gurps.NewEntity().Save(path))

	var n *Navigator
	var selected []string
	var focused bool
	screen.Do(func() {
		n = Workspace.Navigator
		n.Reload()
		n.ApplySelectedPaths([]string{path})
		n.InitialFocus()
		selected = n.SelectedPaths()
		focused = wnd.Focus() == n.table.AsPanel()
	})
	c.Equal([]string{path}, selected, "the sheet in the user library must be selected")
	c.True(focused, "the navigator's list must have the focus")

	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles := contextMenuTitles(t, screen, wnd)
	showOnDisk := i18n.Text("Show on Disk")
	c.True(slices.Contains(titles, showOnDisk), "shift+F10 must open the menu for the selected file; it offers %v",
		titles)
	c.True(slices.Contains(titles, cloneSheetAction.Title), "the menu for a sheet must offer %q; it offers %v",
		cloneSheetAction.Title, titles)
	closeContextMenu(c, screen, wnd)

	screen.Do(func() {
		n.table.ClearSelection()
		focused = wnd.Focus() == n.table.AsPanel()
	})
	c.True(focused, "the navigator's list must still have the focus for the chord to be asking it")
	screen.KeyPress(unison.KeyF10, mod.Shift)
	var open bool
	screen.Do(func() { open = openMenuPopup(wnd) != nil })
	c.False(open, "shift+F10 must open nothing while nothing is selected")
}
