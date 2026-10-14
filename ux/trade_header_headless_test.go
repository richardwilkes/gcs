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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// carriedDescriptionHeader returns the table header of the list's description column, that column's header and its
// index, or a nil column header if the list has no such column or it has another kind of header.
func carriedDescriptionHeader(list *PageList[*gurps.Equipment]) (th *unison.TableHeader[*Node[*gurps.Equipment]],
	header *PageTableColumnHeader[*gurps.Equipment], col int,
) {
	col = list.Table.ColumnIndexForID(gurps.EquipmentDescriptionColumn)
	if col == -1 {
		return list.tableHeader, nil, -1
	}
	var ok bool
	if header, ok = list.tableHeader.ColumnHeaders[col].(*PageTableColumnHeader[*gurps.Equipment]); !ok {
		return list.tableHeader, nil, -1
	}
	return list.tableHeader, header, col
}

// layOutColumnHeader gives the column header its column's frame and lays it out, as the table header does to draw it or
// hand it an event, and returns that frame. UI thread only.
func layOutColumnHeader(th *unison.TableHeader[*Node[*gurps.Equipment]],
	header *PageTableColumnHeader[*gurps.Equipment], col int,
) geom.Rect {
	frame := th.ColumnFrame(col)
	cell := header.AsPanel()
	cell.SetFrameRect(frame)
	cell.ValidateLayout()
	return frame
}

// carriedHeaderPoints scrolls the header of the sheet's carried equipment list into view and returns, on the screen,
// the middle of its description column's title area and of each of that column's trailing buttons, along with whether
// all of them are in view. UI thread only.
func carriedHeaderPoints(wnd *unison.Window, sheet *Sheet) (title geom.Point, buttons []geom.Point, inView bool) {
	th, header, col := carriedDescriptionHeader(sheet.CarriedEquipment)
	if header == nil {
		return title, nil, false
	}
	th.ScrollRectIntoView(th.ColumnFrame(col))
	th.ValidateScrollRoot()
	frame := layOutColumnHeader(th, header, col)
	visible := visibleRect(th.AsPanel())
	inView = true
	toScreen := func(r geom.Rect) geom.Point {
		pt := th.PointToRoot(r.Center().Add(frame.Point))
		inView = inView && pt.In(visible)
		return screenPoint(wnd, pt)
	}
	title = toScreen(header.ContentRect(false))
	for _, b := range header.buttons {
		buttons = append(buttons, toScreen(b.FrameRect()))
	}
	return title, buttons, inView
}

// TestCarriedEquipmentHeaderButtons verifies that the buy and shopping record buttons in the header of a sheet's
// carried equipment list take clicks of their own without making the header taller or keeping a click on its title
// from sorting the list, that the buy button swaps the list for the buy panel and its Back button swaps it back, and
// that an assistive technology can press the buttons.
func TestCarriedEquipmentHeaderButtons(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}

	var buttonCount int
	var height, plainHeight float32
	var buyTooltip, textTooltip bool
	var buyCursor *unison.Cursor
	screen.Do(func() {
		th, header, col := carriedDescriptionHeader(sheet.CarriedEquipment)
		if header == nil {
			return
		}
		buttonCount = len(header.buttons)
		if buttonCount == 0 {
			return
		}
		_, pref, _ := header.Sizes(geom.Size{})
		height = pref.Height
		_, pref, _ = NewPageTableColumnHeader[*gurps.Equipment](header.Text.String(), "", nil).Sizes(geom.Size{})
		plainHeight = pref.Height
		// The points are in the table header's coordinates, which it translates into the column header's.
		frame := layOutColumnHeader(th, header, col)
		buy := header.buttons[0]
		overBuy := buy.FrameRect().Center().Add(frame.Point)
		th.UpdateTooltipCallback(overBuy, geom.Rect{})
		buyTooltip = header.Tooltip == buy.Tooltip
		buyCursor = th.UpdateCursorCallback(overBuy)
		th.UpdateTooltipCallback(header.ContentRect(false).Center().Add(frame.Point), geom.Rect{})
		textTooltip = header.Tooltip == header.textTooltip
	})
	c.Equal(2, buttonCount, "the carried equipment header of a sheet must have the buy and shopping record buttons")
	if buttonCount != 2 {
		return
	}
	c.Equal(plainHeight, height, "the buttons must not make the header any taller than its title alone")
	c.True(buyTooltip, "the column must show the buy button's tooltip while the pointer is over it")
	c.Equal(unison.PointingCursor(), buyCursor, "the pointer must show the buy button can be clicked")
	c.True(textTooltip, "the column must show its own tooltip again away from the buttons")
	uxtest.CaptureScreen(t, c, screen, "trade_header")

	points := func() (title, buy, record geom.Point) {
		t.Helper()
		var buttons []geom.Point
		var inView bool
		screen.Do(func() { title, buttons, inView = carriedHeaderPoints(wnd, sheet) })
		if !inView || len(buttons) != 2 {
			t.Fatal("the header of the carried equipment list must be in view to be clicked")
		}
		return title, buttons[0], buttons[1]
	}
	sortState := func() unison.SortState {
		var state unison.SortState
		screen.Do(func() {
			if _, header, _ := carriedDescriptionHeader(sheet.CarriedEquipment); header != nil {
				state = header.SortState()
			}
		})
		return state
	}
	var buying, panelOnPage, listOnPage bool
	readState := func() {
		screen.Do(func() {
			buying = sheet.buying()
			panelOnPage = sheet.buyPanel != nil && unison.AncestorIs(sheet.buyPanel, sheet.page)
			listOnPage = unison.AncestorIs(sheet.CarriedEquipment, sheet.page)
		})
	}

	c.Equal(-1, sortState().Order, "the list must start out unsorted")
	title, _, _ := points()
	screen.Click(title)
	sorted := sortState()
	c.Equal(0, sorted.Order, "a click on the title must sort the list on the column")

	_, _, record := points()
	screen.Click(record)
	c.Equal(sorted, sortState(), "a click on the shopping record button must not sort the list")
	readState()
	c.False(buying, "the shopping record button must not start buying")

	_, buy, _ := points()
	screen.Click(buy)
	c.Equal(sorted, sortState(), "a click on the buy button must not sort the list")
	readState()
	c.True(buying, "the buy button must start buying")
	c.True(panelOnPage, "the buy panel must be on the page while buying")
	c.False(listOnPage, "the carried equipment list must leave the page while buying")
	if !buying {
		t.FailNow()
	}

	var back *unison.Button
	var backVisible bool
	screen.Do(func() {
		if back = sheet.buyPanel.back; back != nil {
			back.ScrollIntoView()
			back.ValidateScrollRoot()
			backVisible = fullyVisible(back.AsPanel())
		}
	})
	if back == nil || !backVisible {
		t.Fatal("the buy panel's Back button must be in view to be clicked")
	}
	screen.Click(screen.PanelCenter(back))
	readState()
	c.False(buying, "the Back button must stop buying")
	c.False(panelOnPage, "the buy panel must leave the page")
	c.True(listOnPage, "the carried equipment list must return to the page")

	// An assistive technology reaches the buttons through the column header, which is described as a group holding
	// them, and pressing one clicks it.
	points()
	var buyNode accessibility.NodeID
	found := false
	if tree := screen.AccessibilityTree(wnd); tree != nil {
		tree.Walk(func(n *accessibility.Node) bool {
			if !found && n.Role == role.Button && n.Name == "Buy equipment" {
				buyNode = n.ID
				found = true
			}
			return !found
		})
	}
	if !found {
		t.Fatal("the buy button must be described to assistive technologies")
	}
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: buyNode, Action: accessibility.Press}),
		"the buy button must carry out a press")
	readState()
	c.True(buying, "pressing the buy button must start buying")
	screen.Do(sheet.stopBuying)

	// A template's carried equipment list has no buttons, and its header is just as tall.
	template, ok := uxtest.OpenedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}
	var templateButtons int
	var sheetHeaderHeight, templateHeaderHeight float32
	screen.Do(func() {
		if _, header, _ := carriedDescriptionHeader(template.Equipment); header != nil {
			templateButtons = len(header.buttons)
		}
		_, pref, _ := sheet.CarriedEquipment.tableHeader.Sizes(geom.Size{})
		sheetHeaderHeight = pref.Height
		_, pref, _ = template.Equipment.tableHeader.Sizes(geom.Size{})
		templateHeaderHeight = pref.Height
	})
	c.Equal(0, templateButtons, "a template's carried equipment header must have no buttons")
	c.Equal(templateHeaderHeight, sheetHeaderHeight, "the sheet's header must be as tall as the template's")
}

// TestSellInCarriedEquipmentContextMenu verifies that the context menu of a sheet's carried equipment list offers to
// sell the selection, and that the command hands the sheet and the list to the sell dialog, while a template's carried
// equipment list offers no such thing.
func TestSellInCarriedEquipmentContextMenu(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sell := contextMenuItemFor(sellEquipmentAction).Title
	var soldFrom *Sheet
	var soldTable *unison.Table[*Node[*gurps.Equipment]]
	var soldNames []string
	uxtest.SwapForTest(t, &sellSelectedEquipment, func(sheet *Sheet, table *unison.Table[*Node[*gurps.Equipment]]) {
		soldFrom = sheet
		soldTable = table
		for _, row := range table.SelectedRows(true) {
			soldNames = append(soldNames, row.Data().Name)
		}
	})

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var table *unison.Table[*Node[*gurps.Equipment]]
	row := -1
	screen.Do(func() {
		entity := sheet.Entity()
		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Shortsword"
		entity.CarriedEquipment = append(entity.CarriedEquipment, eqp)
		entity.Recalculate()
		sheet.Rebuild(true)
		table = sheet.CarriedEquipment.Table
		for i := 0; i <= table.LastRowIndex(); i++ {
			if table.RowFromIndex(i).Data() == eqp {
				row = i
			}
		}
	})
	if row == -1 {
		t.Fatal("the carried equipment list must show the new item")
	}
	screen.Click(cellScreenPoint(t, screen, wnd, table, row, gurps.EquipmentDescriptionColumn))
	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, sell), "the sheet's carried equipment menu must offer %q; it offers %v", sell,
		titles)
	closeContextMenu(c, screen, wnd)
	screen.Do(func() { table.PerformCmd(table, SellEquipmentItemID) })
	c.Equal(sheet, soldFrom, "selling must be handed the sheet")
	c.True(soldTable == table, "selling must be handed the carried equipment table")
	c.Equal([]string{"Shortsword"}, soldNames, "selling must be handed the selection")

	template, ok := uxtest.OpenedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}
	var templateTable *unison.Table[*Node[*gurps.Equipment]]
	screen.Do(func() {
		eqp := gurps.NewEquipment(nil, nil, false)
		eqp.Name = "Shortsword"
		template.template.Equipment = append(template.template.Equipment, eqp)
		template.Rebuild(true)
		templateTable = template.Equipment.Table
	})
	screen.Click(cellScreenPoint(t, screen, wnd, templateTable, 0, gurps.EquipmentDescriptionColumn))
	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles = contextMenuTitles(t, screen, wnd)
	c.NotNil(titles, "the template's carried equipment list must have a menu")
	c.False(slices.Contains(titles, sell), "a template's carried equipment menu must not offer %q; it offers %v",
		sell, titles)
	closeContextMenu(c, screen, wnd)
}

// TestBuyModeGatesListCommands verifies that while the buy panel stands in for a sheet's carried equipment list, the
// commands that would add to that list are disabled, the panel takes the list's place for the layout editor and for
// drops, and buying ends if the layout stops showing the carried equipment.
func TestBuyModeGatesListCommands(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	newCarriedIDs := []int{
		NewCarriedEquipmentItemID,
		NewCarriedEquipmentContainerItemID,
		NewCarriedEquipmentGroupItemID,
	}
	type gates struct {
		canStart, buyEnabled, moveToCarried bool
		newCarried                          []bool
	}
	readGates := func() gates {
		var g gates
		g.canStart = sheet.canStartBuying()
		g.buyEnabled = buyEquipmentAction.EnabledCallback(buyEquipmentAction, nil)
		g.moveToCarried = sheet.OtherEquipment.CanPerformCmd(nil, MoveToCarriedEquipmentItemID)
		for _, id := range newCarriedIDs {
			g.newCarried = append(g.newCarried, sheet.CanPerformCmd(nil, id))
		}
		return g
	}
	var before, during, after gates
	var blockPanelIsBuyPanel, dropTargetIsBuyPanel, editorFindsBuyPanel, listsBuilt, canStartWhileEditing bool
	var buyingAfterHide, canStartAfterHide, canStartAfterUndo bool
	screen.Do(func() {
		entity := sheet.Entity()
		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Rope"
		entity.OtherEquipment = append(entity.OtherEquipment, eqp)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.OtherEquipment.Table.SelectByIndex(0)
		before = readGates()

		sheet.startBuying()
		during = readGates()
		blockPanelIsBuyPanel = sheet.blockPanel(gurps.BlockEquipmentKey) == unison.Paneler(sheet.buyPanel)
		dropTargetIsBuyPanel = sheet.keyToPanel(equipmentDragKey) == sheet.buyPanel.AsPanel()
		listsBuilt = !slices.ContainsFunc(sheet.lists(), func(l sheetList) bool { return xreflect.IsNil(l) })
		sheet.toggleLayoutEditing()
		if sheet.layoutEditing() {
			editorFindsBuyPanel = sheet.layoutEditor.ensureRegions().leafFor(gurps.BlockEquipmentKey) != nil
			sheet.toggleLayoutEditing()
		}
		sheet.stopBuying()
		after = readGates()

		sheet.toggleLayoutEditing()
		if sheet.layoutEditing() {
			canStartWhileEditing = sheet.canStartBuying()
			sheet.toggleLayoutEditing()
		}

		sheet.startBuying()
		sheet.hideLayoutBlock(gurps.BlockEquipmentKey)
		buyingAfterHide = sheet.buying()
		canStartAfterHide = sheet.canStartBuying()
		sheet.UndoManager().Undo()
		canStartAfterUndo = sheet.canStartBuying()
	})
	c.True(before.canStart, "a new sheet must be able to start buying")
	c.True(before.buyEnabled, "Buy Equipment must be enabled for a new sheet")
	c.True(before.moveToCarried, "moving the selected other equipment to the carried list must be enabled")
	c.Equal([]bool{true, true, true}, before.newCarried, "the New Carried Equipment commands must be enabled")

	c.False(during.canStart, "buying must not start again while under way")
	c.False(during.buyEnabled, "Buy Equipment must be disabled while buying")
	c.False(during.moveToCarried, "moving equipment to the carried list must be disabled while buying")
	c.Equal([]bool{false, false, false}, during.newCarried,
		"the New Carried Equipment commands must be disabled while buying")
	c.True(blockPanelIsBuyPanel, "the buy panel must stand for the carried equipment block in the layout editor")
	c.True(dropTargetIsBuyPanel, "equipment dropped on the sheet must go to the buy panel")
	c.True(listsBuilt, "every list must still exist while buying")
	c.True(editorFindsBuyPanel, "the layout editor must find the carried equipment block while buying")

	c.Equal(before, after, "stopping must enable everything again")
	c.False(canStartWhileEditing, "buying must not start while the layout is being edited")
	c.False(buyingAfterHide, "hiding the carried equipment block must end buying")
	c.False(canStartAfterHide, "buying must not start while the layout hides the carried equipment")
	c.True(canStartAfterUndo, "buying must be possible again once the block is shown")
}

// TestCarriedEquipmentHeaderKeepsItsOwnDescription checks that a trailing button's tooltip, lent to the column header
// while the pointer is over the button (see PageTableColumnHeader.buttonTooltip), doesn't become the column's
// description for assistive technologies, neither of the column header node the table header describes the column
// with nor of the header's own node within it, which both keep the header's own tooltip although nothing has redrawn
// the header.
func TestCarriedEquipmentHeaderKeepsItsOwnDescription(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		// The header has a tooltip of its own, the exact totals, only while the display format rounds them.
		entity := sheet.Entity()
		entity.CarriedEquipment = []*gurps.Equipment{newTradeTestItem(entity, "Gem", "10.5", fxp.One)}
		entity.SheetSettings.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.ZeroPlaces}
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	// The redraw the rebuild asked for is over, and nothing below redraws the header before the window is described.
	var tooltip, buyTooltip string
	lent := false
	screen.Do(func() {
		th, header, col := carriedDescriptionHeader(sheet.CarriedEquipment)
		if header == nil || len(header.buttons) == 0 {
			return
		}
		tooltip = header.TooltipText()
		frame := layOutColumnHeader(th, header, col)
		buy := header.buttons[0]
		buyTooltip = buy.Accessibility.Name
		th.UpdateTooltipCallback(buy.FrameRect().Center().Add(frame.Point), geom.Rect{})
		lent = header.Tooltip == buy.Tooltip
	})
	if !lent {
		t.Fatal("the column must lend the buy button's tooltip while the pointer is over it")
	}
	c.NotEqual("", tooltip, "the header must have a tooltip of its own")

	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the window must be described")
	}
	var group, column *accessibility.Node
	tree.Walk(func(n *accessibility.Node) bool {
		if n.Role == role.Button && n.Name == buyTooltip {
			group = tree.Node(n.Parent)
			return false
		}
		return true
	})
	if group == nil {
		t.Fatal("the column header holding the buy button must be described")
	}
	for column = tree.Node(group.Parent); column != nil && column.Role != role.ColumnHeader; column = tree.Node(column.Parent) {
	}
	if column == nil {
		t.Fatal("the column header must be described as a column header")
	}
	c.Equal(tooltip, group.Description, "the header must be described by its own tooltip, not the button's")
	c.Equal(tooltip, column.Description, "the column must be described by its own tooltip, not the button's")
}
