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
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// buyTestItem returns a piece of equipment owned by owner with the name, value, quantity and tags.
func buyTestItem(owner gurps.DataOwner, name, value string, quantity int, tags ...string) *gurps.Equipment {
	eqp := gurps.NewEquipment(owner, nil, false)
	eqp.Name = name
	eqp.BaseValue = value
	eqp.Quantity = fxp.FromInteger(quantity)
	eqp.Tags = tags
	return eqp
}

// buyTestSheet starts a headless workspace, with modifier prompts forbidden and drag feedback kept from drawing, and
// opens a new character sheet carrying what carried returns for its entity.
func buyTestSheet(t *testing.T, c check.Checker, carried func(entity *gurps.Entity) []*gurps.Equipment) (*unison.HeadlessScreen, *unison.Window, *Sheet) {
	t.Helper()
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		entity.CarriedEquipment = carried(entity)
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	return screen, wnd, sheet
}

// buyTestCoins returns Silver Eagles worth $20 and Copper Pennies worth $1, both tagged as money, in the quantities
// given and owned by the entity.
func buyTestCoins(entity *gurps.Entity, eagles, pennies int) []*gurps.Equipment {
	return []*gurps.Equipment{
		buyTestItem(entity, "Silver Eagle", "20", eagles, gurps.MoneyTag),
		buyTestItem(entity, "Copper Penny", "1", pennies, gurps.MoneyTag),
	}
}

// buyTestLibrary opens an equipment list dockable showing the items, at path, or in a temporary file when path is
// empty, and returns its table.
func buyTestLibrary(t *testing.T, screen *unison.HeadlessScreen, path string, items ...*gurps.Equipment) *unison.Table[*Node[*gurps.Equipment]] {
	t.Helper()
	if path == "" {
		path = filepath.Join(t.TempDir(), "Gear"+gurps.EquipmentExt)
	}
	var table *unison.Table[*Node[*gurps.Equipment]]
	screen.Do(func() {
		d := NewEquipmentTableDockable(path, items)
		DisplayNewDockable(d)
		table = d.table
	})
	if table == nil {
		t.Fatal("the equipment list could not be opened")
	}
	return table
}

// buyTestDrop drags the table's rows holding the items over the sheet and drops them there, through the sheet's
// rerouting of drops, as the window does for a drop anywhere on the sheet, and returns the operation offered and
// whether the drop was taken.
func buyTestDrop(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) (op drag.Op, dropped bool) {
	t.Helper()
	var rows []*Node[*gurps.Equipment]
	screen.Do(func() {
		for _, eqp := range items {
			for _, row := range table.RootRows() {
				if row.Data() == eqp {
					rows = append(rows, row)
					break
				}
			}
		}
	})
	if len(rows) != len(items) {
		t.Fatal("the table must show every item to be dragged")
	}
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.Equipment]]{
		Table: table,
		Rows:  rows,
	}))
	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	screen.Do(func() {
		root := sheet.AsPanel()
		where := geom.NewPoint(10, 10)
		if op = root.DragEnteredCallback(di, where, 0); op != drag.None {
			dropped = root.DropCallback(di, where, 0)
		} else {
			root.DragExitedCallback()
		}
	})
	return op, dropped
}

// buyTestState is what a test reads from the buy panel.
type buyTestState struct {
	summary     string
	notice      string
	name        string
	rows        []string
	finalizeOn  bool
	manualOn    bool
	noticeShown bool
}

// readBuyPanel reads the state of the sheet's buy panel, failing the test if it has none.
func readBuyPanel(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) buyTestState {
	t.Helper()
	var state buyTestState
	found := false
	screen.Do(func() {
		p := sheet.buyPanel
		if p == nil {
			return
		}
		found = true
		state.summary = p.summary.Title()
		state.noticeShown = p.notice.Parent() != nil
		if state.noticeShown {
			state.notice = p.notice.Title()
		}
		state.name = p.nameField.Text()
		state.finalizeOn = p.finalize.Enabled()
		state.manualOn = p.manualButton.Enabled()
		for _, item := range p.items {
			state.rows = append(state.rows, fmt.Sprintf("%s | %s | %s | %s", item.field.Text(), item.name.Text.String(),
				item.each.Text.String(), item.total.Text.String()))
		}
	})
	if !found {
		t.Fatal("the sheet must be buying")
	}
	return state
}

// buyTestCarried describes the sheet's top-level carried equipment, each as its name and quantity.
func buyTestCarried(screen *unison.HeadlessScreen, sheet *Sheet) []string {
	var list []string
	screen.Do(func() {
		for _, eqp := range sheet.Entity().CarriedEquipment {
			list = append(list, fmt.Sprintf("%s ×%s", eqp.Name, eqp.Quantity.String()))
		}
	})
	return list
}

// buyTestSetQuantity types text into the quantity field of the buy panel's row at index.
func buyTestSetQuantity(screen *unison.HeadlessScreen, sheet *Sheet, index int, text string) {
	screen.Do(func() { sheet.buyPanel.items[index].field.SetText(text) })
}

// TestBuyPanelDropAndAutomaticPayment checks that equipment dropped on the sheet while buying lands in the buy panel
// with a quantity of one and its price, that the carried money pays for it automatically, that the payment follows the
// quantities and the percentage, that a purchase the money can't cover can't be finalized, that a group's quantity
// can't be changed, and that a row can be removed again, all without touching the sheet or its undo.
func TestBuyPanelDropAndAutomaticPayment(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	longbow := buyTestItem(nil, "Longbow", "45", 1)
	kit := gurps.NewEquipmentGroup(nil, nil)
	kit.Name = "Kit"
	rope := buyTestItem(nil, "Rope", "5", 2)
	rope.SetParent(kit)
	kit.Children = []*gurps.Equipment{rope}
	library := buyTestLibrary(t, screen, "", longbow, kit)

	var modified bool
	screen.Do(func() {
		sheet.markUnmodified()
		sheet.startBuying()
	})
	state := readBuyPanel(t, screen, sheet)
	c.Equal("Nothing to buy yet.", state.summary)
	c.False(state.finalizeOn, "nothing can be bought before something is dropped")
	c.False(state.manualOn, "there is nothing to pay for yet")

	op, dropped := buyTestDrop(t, screen, sheet, library, longbow)
	c.Equal(drag.Copy, op, "a drop from an equipment list is offered as a copy")
	c.True(dropped, "the drop is taken")
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45"}, state.rows)
	c.Equal("Pay automatically: Silver Eagle ×2, Copper Penny ×5 ($45)", state.summary)
	c.False(state.noticeShown, "the payment is exact, so there is no change to warn about")
	c.Equal("Bought Longbow", state.name, "the transaction is named after what is bought")
	c.True(state.finalizeOn, "the purchase can be finalized")
	c.True(state.manualOn, "the money to pay with can be chosen")

	buyTestSetQuantity(screen, sheet, 0, "2")
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"2 | Longbow | $45 each | Total $90"}, state.rows)
	c.Equal("Pay automatically: Silver Eagle ×4, Copper Penny ×10 ($90)", state.summary)

	buyTestSetQuantity(screen, sheet, 0, "4")
	state = readBuyPanel(t, screen, sheet)
	c.Equal("Not enough money carried: $180 needed, $151 carried.", state.summary)
	c.False(state.finalizeOn, "a purchase the money carried can't cover can't be finalized")

	buyTestSetQuantity(screen, sheet, 0, "1")
	screen.Do(func() { sheet.buyPanel.percentField.SetText("200") })
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $90 each | Total $90"}, state.rows, "the prices follow the percentage")
	screen.Do(func() { sheet.buyPanel.percentField.SetText("100") })

	op, dropped = buyTestDrop(t, screen, sheet, library, kit)
	c.Equal(drag.Copy, op)
	c.True(dropped)
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45", "1 | Kit | $10 each | Total $10"}, state.rows,
		"a group is bought once, its price that of what it holds")
	c.Equal("Pay automatically: Silver Eagle ×3 ($60)", state.summary,
		"the smallest total the money carried can make that covers the price is paid")
	c.Equal("$5 of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		state.notice)
	c.Equal("Bought Longbow, Kit", state.name)
	var groupFieldEnabled bool
	screen.Do(func() { groupFieldEnabled = sheet.buyPanel.items[1].field.Enabled() })
	c.False(groupFieldEnabled, "a group's quantity can't be changed")

	screen.Do(func() {
		sheet.buyPanel.nameField.SetText("Archery")
		sheet.buyPanel.items[1].remove.Click()
	})
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45"}, state.rows, "the row is removed")
	c.Equal("Archery", state.name, "a name the user typed is kept")
	screen.Do(func() { sheet.buyPanel.items[0].remove.Click() })
	state = readBuyPanel(t, screen, sheet)
	c.Nil(state.rows)
	c.Equal("Nothing to buy yet.", state.summary)
	c.False(state.finalizeOn)

	screen.Do(func() { modified = sheet.Modified() })
	c.False(modified, "nothing done in the buy panel marks the sheet as modified")
	c.False(sheet.UndoManager().CanUndo(), "nothing done in the buy panel is recorded for undo")
	c.Equal([]string{"Silver Eagle ×7", "Copper Penny ×11"}, buyTestCarried(screen, sheet))
}

// TestBuyPanelHoldsTheHeightOfTheCarriedEquipmentList checks that the buy panel takes the place of the carried
// equipment list at the list's height, so that what lies below it on the page stays put, unless it needs more room, and
// that it grows past that height once it does.
func TestBuyPanelHoldsTheHeightOfTheCarriedEquipmentList(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		items := buyTestCoins(entity, 7, 11)
		for i := range 20 {
			items = append(items, buyTestItem(entity, fmt.Sprintf("Item %d", i), "1", 1))
		}
		return items
	})
	var listHeight, otherTop float32
	screen.Do(func() {
		wnd.Content().ValidateLayout()
		listHeight = sheet.CarriedEquipment.FrameRect().Height
		otherTop = sheet.OtherEquipment.FrameRect().Y
		sheet.startBuying()
		wnd.Content().ValidateLayout()
	})
	c.True(listHeight > 0)
	var natural, held, frame geom.Size
	var otherTopAfter float32
	screen.Do(func() {
		p := sheet.buyPanel
		frame = p.FrameRect().Size
		_, natural, _ = p.held.Layout.LayoutSizes(p.AsPanel(), geom.Size{Width: frame.Width})
		_, held, _ = p.held.LayoutSizes(p.AsPanel(), geom.Size{Width: frame.Width})
		otherTopAfter = sheet.OtherEquipment.FrameRect().Y
	})
	c.True(natural.Height < listHeight, "the empty panel must want less room than a long list, or the test proves nothing")
	c.Equal(listHeight, held.Height, "the panel asks for the list's height")
	c.Equal(listHeight, frame.Height, "the panel stands at the list's height")
	c.Equal(otherTop, otherTopAfter, "what lies below the list stays put")

	// Once the panel needs more room than the list had, it grows.
	items := make([]*gurps.Equipment, 0, 40)
	for i := range 40 {
		items = append(items, buyTestItem(nil, fmt.Sprintf("Gear %d", i), "1", 1))
	}
	library := buyTestLibrary(t, screen, "", items...)
	for _, item := range items {
		if _, dropped := buyTestDrop(t, screen, sheet, library, item); !dropped {
			t.Fatalf("%s must be dropped", item.Name)
		}
	}
	screen.Do(func() {
		wnd.Content().ValidateLayout()
		frame = sheet.buyPanel.FrameRect().Size
		_, natural, _ = sheet.buyPanel.held.Layout.LayoutSizes(sheet.buyPanel.AsPanel(), geom.Size{Width: frame.Width})
	})
	c.True(natural.Height > listHeight, "the full panel must want more room than the list, or the test proves nothing")
	c.Equal(natural.Height, frame.Height, "the panel grows to what it needs")

	// Going back puts the list back at its own height.
	screen.Do(func() {
		sheet.stopBuying()
		wnd.Content().ValidateLayout()
	})
	c.Equal(listHeight, sheet.CarriedEquipment.FrameRect().Height)
}

// TestBuyPanelBuysTheSheetsOwnEquipment checks that equipment dragged from the sheet's own other equipment list is
// bought like any other, both when it reaches the buy panel through the sheet's rerouting of drops and when it reaches
// the panel directly, and that finalizing carries a copy while the original stays in the other equipment list.
func TestBuyPanelBuysTheSheetsOwnEquipment(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var rope *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		rope = buyTestItem(entity, "Rope", "5", 1)
		entity.OtherEquipment = []*gurps.Equipment{rope}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.startBuying()
	})
	op, dropped := buyTestDrop(t, screen, sheet, sheet.OtherEquipment.Table, rope)
	c.Equal(drag.Copy, op, "the sheet's own equipment is offered as a copy")
	c.True(dropped, "the drop is taken")
	state := readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Rope | $5 each | Total $5"}, state.rows)
	c.Equal("Pay automatically: Copper Penny ×5 ($5)", state.summary)

	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	var canAccept, droppedDirectly bool
	var updated drag.Op
	screen.Do(func() {
		p := sheet.buyPanel.AsPanel()
		canAccept = p.CanAcceptDropCallback(di)
		updated = p.DragUpdatedCallback(di, geom.Point{}, 0)
		droppedDirectly = p.DropCallback(di, geom.Point{}, 0)
	})
	c.True(canAccept, "the panel accepts the sheet's own equipment")
	c.Equal(drag.Copy, updated, "moving over the panel offers to take it")
	c.True(droppedDirectly, "dropping it on the panel adds it")
	state = readBuyPanel(t, screen, sheet)
	c.Equal([]string{"1 | Rope | $5 each | Total $5", "1 | Rope | $5 each | Total $5"}, state.rows)
	c.Equal("Pay automatically: Copper Penny ×10 ($10)", state.summary)

	screen.Do(func() { sheet.buyPanel.finalize.Click() })
	c.Equal([]string{"Silver Eagle ×7", "Copper Penny ×1", "Rope ×2"}, buyTestCarried(screen, sheet),
		"the money is spent and the ropes bought are carried")
	var other []*gurps.Equipment
	screen.Do(func() { other = sheet.Entity().OtherEquipment })
	c.Equal([]*gurps.Equipment{rope}, other, "the other equipment list keeps the original")
	c.Equal(fxp.One, rope.Quantity, "the original's quantity is unchanged")
}

// buyTestMoneyLibrary writes Silver Eagles and Copper Pennies into a file of the user library and returns the sources
// of the two, in that order.
func buyTestMoneyLibrary(t *testing.T, c check.Checker) []gurps.Source {
	t.Helper()
	user := gurps.GlobalSettings().Libraries.User()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Buy Money" + gurps.EquipmentExt}
	path := filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))
	eagle := buyTestItem(nil, "Silver Eagle", "20", 1, gurps.MoneyTag)
	penny := buyTestItem(nil, "Copper Penny", "1", 1, gurps.MoneyTag)
	c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{eagle, penny}, path))
	return []gurps.Source{{LibraryFile: libFile, TID: eagle.TID}, {LibraryFile: libFile, TID: penny.TID}}
}

// TestBuyPanelReturnsChange checks that an overpayment warns that the change can't be returned while the sheet has no
// sell money, that the change is worked out from the sell money once the sheet has some, and that finalizing spends
// the money, carries the item and the change, records the purchase, ends buying with the focus on the carried
// equipment list and can be undone and redone as one edit.
func TestBuyPanelReturnsChange(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	sources := buyTestMoneyLibrary(t, c)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = nil
		sheet.startBuying()
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	state := readBuyPanel(t, screen, sheet)
	c.Equal("Pay automatically: Silver Eagle ×1 ($20)", state.summary)
	c.True(state.noticeShown, "change that can't be returned is warned about")
	c.Equal("$5 of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		state.notice)
	c.True(state.finalizeOn, "the purchase can still be finalized")

	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = sources[:1]
		sheet.Rebuild(true)
	})
	state = readBuyPanel(t, screen, sheet)
	c.Equal("$5 of change can't be returned with the sheet's sell money or the money paid.", state.notice,
		"the panel follows a change to the sell money")

	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = sources
		sheet.Rebuild(true)
	})
	state = readBuyPanel(t, screen, sheet)
	c.Equal("Pay automatically: Silver Eagle ×1 ($20); change: Copper Penny ×5 ($5)", state.summary)
	c.False(state.noticeShown, "the change can be returned in full")

	screen.Do(func() {
		sheet.buyPanel.finalize.RequestFocus()
		sheet.buyPanel.finalize.Click()
	})
	var buying, listOnPage, listFocused bool
	screen.Do(func() {
		buying = sheet.buying()
		listOnPage = unison.AncestorIs(sheet.CarriedEquipment, sheet.page)
		listFocused = wnd.Focus() == sheet.CarriedEquipment.Table.AsPanel()
	})
	c.False(buying, "finalizing ends buying")
	c.True(listFocused, "the focus moves from the button that finalized to the carried equipment list")
	c.True(listOnPage, "the carried equipment list is back on the page")
	after := []string{"Silver Eagle ×6", "Dagger ×1", "Copper Penny ×5"}
	c.Equal(after, buyTestCarried(screen, sheet))
	var record []*gurps.ShoppingRecordEntry
	var pennySource gurps.Source
	var daggerEquipped bool
	screen.Do(func() {
		entity := sheet.Entity()
		record = gurps.CloneShoppingRecordList(entity.ShoppingRecord)
		pennySource = entity.CarriedEquipment[2].Source
		daggerEquipped = entity.CarriedEquipment[1].Equipped
	})
	c.Equal(sources[1], pennySource, "the change comes from the sell money in the library")
	c.True(daggerEquipped, "the item bought is carried equipped")
	c.Equal(1, len(record))
	if len(record) == 1 {
		entry := record[0]
		c.Equal(gurps.ShoppingBought, entry.Kind)
		c.Equal("Bought Dagger", entry.Name)
		c.Equal([]gurps.ShoppingRecordItem{{Name: "Dagger", Quantity: fxp.One, Value: fxp.FromInteger(15)}}, entry.Items)
		c.Equal([]gurps.ShoppingRecordItem{
			{Name: "Silver Eagle", Quantity: -fxp.One, Value: fxp.FromInteger(20)},
			{Name: "Copper Penny", Quantity: fxp.Five, Value: fxp.One},
		}, entry.Money, "the money spent is negative and the change positive")
		c.Equal(fxp.FromInteger(15), entry.Exchanged, "the money exchanged is what was paid less the change")
	}

	mgr := sheet.UndoManager()
	var undoTitle string
	screen.Do(func() { undoTitle = mgr.UndoTitle() })
	c.Equal("Undo Buy Equipment", undoTitle, "the purchase can be undone as one edit")
	var recordCount int
	screen.Do(func() {
		mgr.Undo()
		recordCount = len(sheet.Entity().ShoppingRecord)
	})
	c.Equal([]string{"Silver Eagle ×7"}, buyTestCarried(screen, sheet), "undo puts the carried equipment back")
	c.Equal(0, recordCount, "undo takes the record of the purchase away")
	screen.Do(func() {
		mgr.Redo()
		recordCount = len(sheet.Entity().ShoppingRecord)
	})
	c.Equal(after, buyTestCarried(screen, sheet), "redo makes the purchase again")
	c.Equal(1, recordCount, "redo records the purchase again")
}

// TestBuyPanelFinalizeMerges checks that a purchase merges into a matching carried item when asked to and is carried
// separately otherwise.
func TestBuyPanelFinalizeMerges(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return append(buyTestCoins(entity, 7, 11), buyTestItem(entity, "Leather Gloves", "30", 1))
	})
	gloves := buyTestItem(nil, "Leather Gloves", "30", 1)
	library := buyTestLibrary(t, screen, "", gloves)

	screen.Do(sheet.startBuying)
	_, dropped := buyTestDrop(t, screen, sheet, library, gloves)
	c.True(dropped)
	c.Equal("Pay automatically: Silver Eagle ×1, Copper Penny ×10 ($30)", readBuyPanel(t, screen, sheet).summary)
	screen.Do(func() { sheet.buyPanel.finalize.Click() })
	c.Equal([]string{"Silver Eagle ×6", "Copper Penny ×1", "Leather Gloves ×2"}, buyTestCarried(screen, sheet),
		"the money is spent and the gloves merge into those carried")
	var exchanged fxp.Int
	screen.Do(func() {
		if record := sheet.Entity().ShoppingRecord; len(record) == 1 {
			exchanged = record[0].Exchanged
		}
	})
	c.Equal(fxp.FromInteger(30), exchanged)

	screen.Do(func() {
		sheet.UndoManager().Undo()
		sheet.startBuying()
	})
	_, dropped = buyTestDrop(t, screen, sheet, library, gloves)
	c.True(dropped)
	screen.Do(func() {
		sheet.buyPanel.merge.State = uncheck.Off
		sheet.buyPanel.finalize.Click()
	})
	c.Equal([]string{"Silver Eagle ×6", "Copper Penny ×1", "Leather Gloves ×1", "Leather Gloves ×1"},
		buyTestCarried(screen, sheet), "without merging, the gloves bought are carried separately")
}

// TestBuyPanelManualPayment checks that the money chosen by hand pays for the purchase in place of the automatic
// choice, starting from it, that a choice that no longer covers the price, or whose money is no longer carried, stops
// the purchase, and that the automatic choice can be had back.
func TestBuyPanelManualPayment(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	longbow := buyTestItem(nil, "Longbow", "45", 1)
	library := buyTestLibrary(t, screen, "", longbow)
	screen.Do(sheet.startBuying)
	_, dropped := buyTestDrop(t, screen, sheet, library, longbow)
	c.True(dropped)

	var ids []tid.TID
	screen.Do(func() {
		for _, one := range sheet.Entity().MoneyItems() {
			ids = append(ids, one.TID)
		}
	})
	if len(ids) != 2 {
		t.Fatal("the sheet must carry two kinds of money")
	}
	eagleID, pennyID := ids[0], ids[1]

	type answer struct {
		chosen       map[tid.TID]int
		useAutomatic bool
		ok           bool
	}
	var next answer
	var gotOwed fxp.Int
	var gotWhat string
	var gotCurrent map[tid.TID]int
	uxtest.SwapForTest(t, &chooseManualPayment,
		func(_ *Sheet, owed fxp.Int, what string, current map[tid.TID]int) (map[tid.TID]int, bool, bool) {
			gotOwed, gotWhat, gotCurrent = owed, what, current
			return next.chosen, next.useAutomatic, next.ok
		})
	choose := func(a answer) buyTestState {
		t.Helper()
		next = a
		screen.Do(func() { sheet.buyPanel.manualButton.Click() })
		return readBuyPanel(t, screen, sheet)
	}

	state := choose(answer{chosen: map[tid.TID]int{eagleID: 3}, ok: true})
	c.Equal(fxp.FromInteger(45), gotOwed, "the dialog is told what is owed")
	c.Equal("Longbow ×1", gotWhat, "the dialog is told what is being bought")
	c.Equal(map[tid.TID]int{eagleID: 2, pennyID: 5}, gotCurrent, "the dialog starts from the automatic payment")
	c.Equal("Pay manually: Silver Eagle ×3 ($60)", state.summary)
	c.Equal("$15 of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		state.notice)
	c.True(state.finalizeOn)

	state = choose(answer{ok: false})
	c.Equal("Pay manually: Silver Eagle ×3 ($60)", state.summary, "canceling keeps the choice")
	c.Equal(map[tid.TID]int{eagleID: 3}, gotCurrent, "the dialog starts from the manual choice")

	state = choose(answer{chosen: map[tid.TID]int{eagleID: 2, pennyID: 10}, ok: true})
	c.Equal("Pay manually: Silver Eagle ×2, Copper Penny ×10 ($50); change: Copper Penny ×5 ($5)", state.summary,
		"change comes from the money paid when the sheet has no sell money")
	c.False(state.noticeShown)

	state = choose(answer{chosen: map[tid.TID]int{pennyID: 5}, ok: true})
	c.Equal("The money chosen is short by $40: $45 needed, $5 chosen.", state.summary)
	c.False(state.finalizeOn, "a manual choice that doesn't cover the price can't be finalized")

	state = choose(answer{chosen: map[tid.TID]int{eagleID: 3}, ok: true})
	c.True(state.finalizeOn)
	screen.Do(func() {
		entity := sheet.Entity()
		entity.CarriedEquipment[0].Quantity = fxp.Two
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	state = readBuyPanel(t, screen, sheet)
	c.Equal("The money chosen is short by $45: $45 needed, $0 chosen.", state.summary,
		"a choice of more money than is carried is dropped")
	c.False(state.finalizeOn)

	state = choose(answer{useAutomatic: true, ok: true})
	c.Equal("Pay automatically: Silver Eagle ×2, Copper Penny ×5 ($45)", state.summary,
		"the automatic choice can be had back")
	c.True(state.finalizeOn)
}

// TestManualPaymentDialog drives the dialog that chooses the money to spend: it lists the money carried with the
// quantities it starts from, keeps OK disabled while the money chosen is short, notes whether the payment is short,
// exact or due change, with what the change will be, and returns the choice, the request to pay automatically again,
// or nothing when canceled.
func TestManualPaymentDialog(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return append(buyTestCoins(entity, 7, 11), buyTestItem(entity, "Rope", "5", 1))
	})
	var ids []tid.TID
	screen.Do(func() {
		for _, one := range sheet.Entity().MoneyItems() {
			ids = append(ids, one.TID)
		}
	})
	if len(ids) != 2 {
		t.Fatal("the sheet must carry two kinds of money")
	}
	eagleID, pennyID := ids[0], ids[1]

	var chosen map[tid.TID]int
	var useAutomatic, ok, returned bool
	open := func(current map[tid.TID]int) (*unison.Dialog, []*IntegerField) {
		t.Helper()
		returned = false
		c.True(screen.Post(func() {
			chosen, useAutomatic, ok = chooseManualPayment(sheet, fxp.FromInteger(45), "Longbow ×1", current)
			returned = true
		}))
		screen.Sync()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		var fields []*IntegerField
		screen.Do(func() { fields = uxtest.PanelsOfType[*IntegerField](dialogWnd.Content()) })
		if len(fields) != 2 {
			t.Fatalf("the dialog must have a field for each kind of money carried, not %d", len(fields))
		}
		return dialog, fields
	}
	type dialogState struct {
		labels   []string
		okOn     bool
		spend    []string
		note     string
		noteIcon *unison.SVG
		noteWarn bool
	}
	read := func(dialog *unison.Dialog, fields []*IntegerField) dialogState {
		t.Helper()
		var s dialogState
		screen.Do(func() {
			content := dialog.Window().Content()
			s.labels = uxtest.LabelTexts(content)
			s.okOn = dialog.Button(unison.ModalResponseOK).Enabled()
			for _, f := range fields {
				s.spend = append(s.spend, f.Text())
			}
			if notes := uxtest.PanelsOfType[*manualPaymentNote](content); len(notes) == 1 {
				s.note = notes[0].text.Title()
				if icon, isSVG := notes[0].icon.Drawable.(*unison.DrawableSVG); isSVG {
					s.noteIcon = icon.SVG
				}
				s.noteWarn = notes[0].warning
			}
		})
		if s.noteIcon == nil {
			t.Fatal("the dialog must have a note with an icon")
		}
		return s
	}
	click := func(dialog *unison.Dialog, code int) {
		t.Helper()
		var b *unison.Button
		screen.Do(func() { b = dialog.Button(code) })
		screen.Click(screen.PanelCenter(b))
		if !returned {
			t.Fatal("the dialog must close")
		}
	}

	dialog, fields := open(map[tid.TID]int{eagleID: 2, pennyID: 50})
	uxtest.CaptureScreen(t, c, screen, "manual_payment_dialog")
	state := read(dialog, fields)
	c.Equal([]string{"2", "11"}, state.spend, "the dialog starts from the quantities given, up to what is carried")
	for _, want := range []string{
		"Money Item", "Carried", "Value Each", "Spend", "Silver Eagle", "7", "$20", "Copper Penny", "11", "$1",
		"Money-tagged items carried by this character.\nOwed: $45 for Longbow ×1.", "Owed: $45", "Spending: $51",
		"Change due: $6",
	} {
		c.True(slices.Contains(state.labels, want), "the dialog must show %q; it shows %v", want, state.labels)
	}
	c.True(state.okOn)
	c.Equal("Change will be returned as Copper Penny ×6.", state.note,
		"without sell money, the change comes from the money paid")
	c.True(state.noteIcon == svg.Info, "change that will be returned is noted")
	c.False(state.noteWarn)
	var typed, refused bool
	screen.Do(func() {
		fields[1].SetText("1")
		fields[1].SetSelectionToEnd()
		refused = !fields[1].RuneTypedCallback('2')
		typed = fields[1].RuneTypedCallback('1')
	})
	c.True(refused, "more of a money item than is carried can't be typed")
	c.True(typed)
	c.Equal("11", read(dialog, fields).spend[1])
	screen.Do(func() { fields[1].SetText("0") })
	state = read(dialog, fields)
	c.True(slices.Contains(state.labels, "Spending: $40"), "the dialog shows %v", state.labels)
	c.True(slices.Contains(state.labels, "Change due: $0"), "the dialog shows %v", state.labels)
	uxtest.CaptureScreen(t, c, screen, "manual_payment_short")
	c.Equal("Short by $5.", state.note)
	c.True(state.noteIcon == svg.Not, "a shortfall shows the not symbol")
	c.True(state.noteWarn)
	c.False(state.okOn, "OK is disabled while the money chosen is short")
	screen.Do(func() { fields[1].SetText("5") })
	state = read(dialog, fields)
	c.True(slices.Contains(state.labels, "Spending: $45"), "the dialog shows %v", state.labels)
	uxtest.CaptureScreen(t, c, screen, "manual_payment_exact")
	c.Equal("Covered exactly. Money you do not spend stays on the sheet.", state.note)
	c.True(state.noteIcon == unison.CheckmarkSVG, "an exact payment shows a checkmark")
	c.False(state.noteWarn)
	c.True(state.okOn)
	click(dialog, unison.ModalResponseOK)
	c.True(ok)
	c.False(useAutomatic)
	c.Equal(map[tid.TID]int{eagleID: 2, pennyID: 5}, chosen)

	// The note says what the change will be, or how much of it neither the sell money nor the money paid can return.
	dialog, fields = open(map[tid.TID]int{eagleID: 3})
	state = read(dialog, fields)
	c.Equal("$15 of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		state.note)
	c.True(state.noteIcon == unison.TriangleExclamationSVG, "change that can't be returned is warned about")
	c.True(state.noteWarn)
	c.True(state.okOn, "an overpayment can still be made")
	click(dialog, unison.ModalResponseCancel)
	sources := buyTestMoneyLibrary(t, c)
	missing := gurps.Source{LibraryFile: sources[0].LibraryFile, TID: "missing"}
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = []gurps.Source{missing} })
	dialog, fields = open(map[tid.TID]int{eagleID: 3})
	state = read(dialog, fields)
	c.Equal("$15 of change can't be returned from the money paid, and none of the sheet's sell money could be found in the libraries.",
		state.note, "sell money that can't be found is told apart from none at all")
	c.True(state.noteWarn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sources[1:] })
	dialog, fields = open(map[tid.TID]int{eagleID: 3})
	state = read(dialog, fields)
	uxtest.CaptureScreen(t, c, screen, "manual_payment_change")
	c.Equal("Change will be returned as Copper Penny ×15.", state.note, "the sell money returns the change")
	c.True(state.noteIcon == svg.Info)
	c.False(state.noteWarn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sources[:1] })
	dialog, fields = open(map[tid.TID]int{eagleID: 4})
	state = read(dialog, fields)
	c.Equal("Change will be returned as Silver Eagle ×1 ($20). $15 of change can't be returned with the sheet's sell money or the money paid.",
		state.note)
	c.True(state.noteIcon == unison.TriangleExclamationSVG)
	c.True(state.noteWarn)
	screen.Do(func() { fields[1].SetText("10") })
	state = read(dialog, fields)
	c.Equal("Change will be returned as Silver Eagle ×2, Copper Penny ×5.", state.note,
		"the money paid is drawn on when the sell money falls short")
	c.True(state.noteIcon == svg.Info)
	c.False(state.noteWarn)
	c.True(state.okOn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = nil })

	dialog, _ = open(nil)
	click(dialog, unison.ModalResponseUserBase)
	c.True(ok)
	c.True(useAutomatic, "Use Automatic asks for the automatic choice")
	c.Nil(chosen)

	dialog, _ = open(nil)
	click(dialog, unison.ModalResponseCancel)
	c.False(ok, "canceling chooses nothing")
}

// buyTestModifierLibrary opens an equipment modifier library holding the modifiers and returns its table.
func buyTestModifierLibrary(t *testing.T, screen *unison.HeadlessScreen, modifiers ...*gurps.EquipmentModifier) *unison.Table[*Node[*gurps.EquipmentModifier]] {
	t.Helper()
	var table *unison.Table[*Node[*gurps.EquipmentModifier]]
	screen.Do(func() {
		d := NewEquipmentModifierTableDockable(filepath.Join(t.TempDir(), "Mods"+gurps.EquipmentModifiersExt), modifiers)
		DisplayNewDockable(d)
		table = d.table
	})
	if table == nil {
		t.Fatal("the equipment modifier list could not be opened")
	}
	return table
}

// TestBuyPanelAttachesAndRemovesModifiers drops an equipment modifier from a library onto an item waiting to be bought
// and takes it off again: the modifier goes only to the item it is dropped on, the price follows it, the row under the
// item shows it with a button that removes it, and a plain modifier asks nothing.
func TestBuyPanelAttachesAndRemovesModifiers(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	library := buyTestLibrary(t, screen, "", sword)
	screen.Do(func() { sheet.startBuying() })
	if op, dropped := buyTestDrop(t, screen, sheet, library, sword); op == drag.None || !dropped {
		t.Fatal("the sword must be dropped into the buy panel")
	}
	fine := gurps.NewEquipmentModifier(nil, nil, false)
	fine.Name = "Fine"
	fine.CostAmount = "+50%"
	modifiers := buyTestModifierLibrary(t, screen, fine)
	var rows []*Node[*gurps.EquipmentModifier]
	screen.Do(func() { rows = modifiers.RootRows() })
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modifiers,
		Rows:  rows,
	}))
	di := &fakeDragInfo{types: []string{equipmentModifierDragKey.UTI}}

	var overHint, overRow drag.Op
	var targeted, dropped bool
	screen.Do(func() {
		p := sheet.buyPanel
		p.ValidateLayout()
		item := p.items[0]
		hint := p.PointFromRoot(p.hint.PointToRoot(p.hint.ContentRect(true).Center()))
		row := p.PointFromRoot(item.name.PointToRoot(item.name.ContentRect(true).Center()))
		c.True(p.CanAcceptDropCallback(di), "the panel must take equipment modifier drags")
		overHint = p.DragUpdatedCallback(di, hint, 0)
		overRow = p.DragUpdatedCallback(di, row, 0)
		targeted = p.modifierTarget == item
		dropped = p.DropCallback(di, row, 0)
	})
	c.Equal(drag.None, overHint, "modifiers dragged over no item go nowhere")
	c.Equal(drag.Copy, overRow, "modifiers dragged over an item are offered to it")
	c.True(targeted, "the item under the modifiers is the target")
	c.True(dropped, "the drop must be taken")

	var modifierCount, rowCount int
	var enabled bool
	var value fxp.Int
	var description string
	var remove *unison.Button
	screen.Do(func() {
		item := sheet.buyPanel.items[0]
		modifierCount = len(item.eqp.Modifiers)
		if modifierCount == 1 {
			enabled = item.eqp.Modifiers[0].Enabled()
		}
		value = item.eqp.AdjustedValue()
		rowCount = len(item.modifierRows)
		if rowCount == 1 {
			if label, ok := item.modifierRows[0].widgets[1].(*unison.Label); ok {
				description = label.Text.String()
			}
			if button, ok := item.modifierRows[0].widgets[2].(*unison.Button); ok {
				remove = button
			}
		}
	})
	c.Equal(1, modifierCount, "the item must have the one modifier")
	c.True(enabled, "a modifier dropped on an item is enabled")
	c.Equal(fxp.FromInteger(150), value, "the modifier must raise the item's value")
	c.Equal(1, rowCount, "the modifier must have a row under the item")
	c.Equal(fine.FullDescription(), description)
	c.Contains(readBuyPanel(t, screen, sheet).rows[0], "$150", "the price must follow the modifier")
	if remove == nil {
		t.Fatal("the modifier's row must have a remove button")
	}

	screen.Do(func() { remove.Click() })
	screen.Do(func() {
		item := sheet.buyPanel.items[0]
		modifierCount = len(item.eqp.Modifiers)
		rowCount = len(item.modifierRows)
		value = item.eqp.AdjustedValue()
	})
	c.Equal(0, modifierCount, "removing takes the modifier off the item")
	c.Equal(0, rowCount, "removing takes the modifier's row away")
	c.Equal(fxp.FromInteger(100), value)
	c.Contains(readBuyPanel(t, screen, sheet).rows[0], "$100")
}

// startRowDrag presses on the table's row and moves the pointer far enough to start dragging the row, so that unison
// and the lists see the drag session a real drag makes, and returns a function that cancels the drag and releases the
// button. The table's dockable is activated first, since the row must be in view to be pressed.
func startRowDrag[T gurps.Node[T]](t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, table *unison.Table[*Node[T]], row int) (cancel func()) {
	t.Helper()
	screen.Do(func() {
		if d := unison.AncestorOrSelf[unison.Dockable](table); d != nil {
			ActivateDockable(d)
		}
	})
	from := cellScreenPoint(t, screen, wnd, table, row, 0)
	screen.MouseDown(from, unison.ButtonLeft, 0)
	screen.MouseMove(from.Add(geom.NewPoint(30, 30)), 0)
	return func() {
		screen.KeyPress(unison.KeyEscape, 0)
		screen.MouseUp(from.Add(geom.NewPoint(30, 30)), unison.ButtonLeft, 0)
	}
}

// TestBuyPanelClaimsEquipmentDroppedElsewhereOnTheSheet checks that, while the buy panel is up, equipment dragged from
// elsewhere onto the sheet's other equipment list is left to the sheet, which reroutes it to the buy panel, while the
// sheet's own equipment is still taken by the list, so that its rows can be moved about.
func TestBuyPanelClaimsEquipmentDroppedElsewhereOnTheSheet(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var own *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		own = buyTestItem(entity, "Tent", "50", 1)
		entity.OtherEquipment = []*gurps.Equipment{own}
		sheet.Rebuild(true)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	library := buyTestLibrary(t, screen, "", sword)
	// A list answers for a drag only during a drag session, which unison tracks for itself, so the rows are really
	// dragged, and the drag data the test leaves behind is cleared afterwards.
	uxtest.SwapForTest(t, &draggedTableData, nil)
	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	accepts := func() bool {
		var accepted bool
		screen.Do(func() { accepted = sheet.OtherEquipment.Table.CanAcceptDropCallback(di) })
		return accepted
	}

	cancel := startRowDrag(t, screen, wnd, library, 0)
	c.True(accepts(), "while not buying, the list takes library equipment")
	screen.Do(sheet.startBuying)
	c.False(accepts(), "other equipment leaves library equipment to the buy panel while buying")
	cancel()

	cancel = startRowDrag(t, screen, wnd, sheet.OtherEquipment.Table, 0)
	c.True(accepts(), "other equipment still takes the sheet's own equipment while buying")
	cancel()
	var otherAfterDrags int
	screen.Do(func() { otherAfterDrags = len(sheet.Entity().OtherEquipment) })
	c.Equal(1, otherAfterDrags, "the canceled drags must change nothing")

	if op, dropped := buyTestDrop(t, screen, sheet, library, sword); op == drag.None || !dropped {
		t.Fatal("the sword must be dropped into the buy panel")
	}
	var bought, other int
	screen.Do(func() {
		bought = len(sheet.buyPanel.items)
		other = len(sheet.Entity().OtherEquipment)
	})
	c.Equal(1, bought, "the drop must land in the buy panel")
	c.Equal(1, other, "the other equipment must be untouched")
}

// TestBuyPanelChangeFollowsEditsToTheMoneyPaid checks that change drawn from the money paid is worked out again when
// that money is edited without its value changing, such as by being renamed, so that the change carried is a copy of
// the money as it now is and merges with it.
func TestBuyPanelChangeFollowsEditsToTheMoneyPaid(t *testing.T) {
	c := check.New(t)
	var ten, penny *gurps.Equipment
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		ten = buyTestItem(entity, "Ten", "10", 1, gurps.MoneyTag)
		penny = buyTestItem(entity, "Copper Penny", "1", 5, gurps.MoneyTag)
		return []*gurps.Equipment{ten, penny}
	})
	dagger := buyTestItem(nil, "Dagger", "10", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = nil
		sheet.startBuying()
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	screen.Do(func() {
		sheet.buyPanel.manual = map[tid.TID]int{ten.TID: 1, penny.TID: 2}
		sheet.buyPanel.refresh()
	})
	c.Equal("Pay manually: Ten ×1, Copper Penny ×2 ($12); change: Copper Penny ×2 ($2)",
		readBuyPanel(t, screen, sheet).summary)

	screen.Do(func() {
		penny.Name = "Penny"
		sheet.Rebuild(true)
	})
	c.Equal("Pay manually: Ten ×1, Penny ×2 ($12); change: Penny ×2 ($2)", readBuyPanel(t, screen, sheet).summary,
		"the change is a copy of the money as it now is")
	screen.Do(func() { sheet.buyPanel.finalize.Click() })
	c.Equal([]string{"Penny ×5", "Dagger ×1"}, buyTestCarried(screen, sheet),
		"the change merges with the money it was copied from")
}

// buyTestDropModifier drags the modifier table's rows onto the buy panel's item and drops them there, returning whether
// the drop was taken.
func buyTestDropModifier(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet, modifiers *unison.Table[*Node[*gurps.EquipmentModifier]], item *buyItem) bool {
	t.Helper()
	var rows []*Node[*gurps.EquipmentModifier]
	screen.Do(func() { rows = modifiers.RootRows() })
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modifiers,
		Rows:  rows,
	}))
	di := &fakeDragInfo{types: []string{equipmentModifierDragKey.UTI}}
	var dropped bool
	screen.Do(func() {
		p := sheet.buyPanel
		p.ValidateLayout()
		row := p.PointFromRoot(item.name.PointToRoot(item.name.ContentRect(true).Center()))
		p.DragUpdatedCallback(di, row, 0)
		dropped = p.DropCallback(di, row, 0)
	})
	return dropped
}

// TestBuyPanelKeepsTheFocusWhenRowsChange checks that the keyboard focus is not left on a button the panel has
// discarded: removing a row with its own button moves the focus to the row that takes its place, or to the percent
// field once no row is left; a modifier's button keeps the focus when the rows are built anew for another item; and
// removing a modifier with its button moves the focus to its item's row.
func TestBuyPanelKeepsTheFocusWhenRowsChange(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 70, 11)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	rope := buyTestItem(nil, "Rope", "5", 1)
	library := buyTestLibrary(t, screen, "", sword, dagger, rope)
	fine := gurps.NewEquipmentModifier(nil, nil, false)
	fine.Name = "Fine"
	modifiers := buyTestModifierLibrary(t, screen, fine)
	screen.Do(sheet.startBuying)
	if _, dropped := buyTestDrop(t, screen, sheet, library, sword, dagger); !dropped {
		t.Fatal("the sword and dagger must be dropped into the buy panel")
	}
	focusIs := func(what string, expected func(p *buyPanel) unison.Paneler) {
		t.Helper()
		var ok bool
		screen.Do(func() {
			target := expected(sheet.buyPanel)
			ok = target != nil && wnd.CurrentFocus() == target.AsPanel()
		})
		c.True(ok, "the focus must be on %s", what)
	}

	screen.Do(func() {
		remove := sheet.buyPanel.items[1].remove
		remove.RequestFocus()
		remove.Click()
	})
	focusIs("the first row's quantity field after removing the last row", func(p *buyPanel) unison.Paneler {
		if len(p.items) != 1 {
			return nil
		}
		return p.items[0].field
	})

	if !buyTestDropModifier(t, screen, sheet, modifiers, sheet.buyPanel.items[0]) {
		t.Fatal("the modifier must be dropped onto the sword")
	}
	modifierButton := func(p *buyPanel) unison.Paneler {
		if len(p.items) != 0 && len(p.items[0].modifierRows) == 1 {
			return p.items[0].modifierRows[0].widgets[2]
		}
		return nil
	}
	screen.Do(func() { modifierButton(sheet.buyPanel).AsPanel().RequestFocus() })
	if _, dropped := buyTestDrop(t, screen, sheet, library, rope); !dropped {
		t.Fatal("the rope must be dropped into the buy panel")
	}
	focusIs("the modifier's button, built anew, after another item is dropped", modifierButton)

	screen.Do(func() {
		if button, ok := modifierButton(sheet.buyPanel).(*unison.Button); ok {
			button.Click()
		}
	})
	focusIs("the item's quantity field after removing its modifier", func(p *buyPanel) unison.Paneler {
		if len(p.items) != 2 || len(p.items[0].modifierRows) != 0 {
			return nil
		}
		return p.items[0].field
	})

	for range 2 {
		screen.Do(func() {
			remove := sheet.buyPanel.items[0].remove
			remove.RequestFocus()
			remove.Click()
		})
	}
	focusIs("the percent field once no row is left", func(p *buyPanel) unison.Paneler {
		if len(p.items) != 0 {
			return nil
		}
		return p.percentField
	})
}

// rerouteHighlightDrawn reports whether the sheet's drop rerouting draws its highlight (see installDropRerouting): the
// sheet's DrawOverCallback is run on a canvas of its own, which starts out clear, and the result is searched for any
// pixel it painted.
func rerouteHighlightDrawn(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) bool {
	t.Helper()
	var drawn bool
	screen.Do(func() {
		root := sheet.AsPanel()
		size := root.ContentRect(true).Size
		img, err := unison.NewImageFromDrawing(int(size.Width), int(size.Height), 72, func(gc *unison.Canvas) {
			root.DrawOverCallback(gc, geom.Rect{Size: size})
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer img.Dispose()
		pixels, err := img.ToNRGBA()
		if err != nil {
			t.Error(err)
			return
		}
		for i := 3; i < len(pixels.Pix); i += 4 {
			if pixels.Pix[i] != 0 {
				drawn = true
				return
			}
		}
	})
	return drawn
}

// TestSheetDropReroutingHighlightsOnlyAWillingTarget checks that the sheet highlights the panel it reroutes a drag to
// only while that panel will take the drop: the buy panel is highlighted for equipment it would buy, not for an
// equipment drag with no rows behind it, which it declines, and not once the drag has left.
func TestSheetDropReroutingHighlightsOnlyAWillingTarget(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	screen.Do(sheet.startBuying)
	library := buyTestLibrary(t, screen, "", buyTestItem(nil, "Shortsword", "100", 1))
	var fromLibrary *unison.TableDragData[*Node[*gurps.Equipment]]
	screen.Do(func() {
		fromLibrary = &unison.TableDragData[*Node[*gurps.Equipment]]{Table: library, Rows: library.RootRows()}
	})
	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	enter := func(data *unison.TableDragData[*Node[*gurps.Equipment]]) drag.Op {
		uxtest.SwapForTest(t, &draggedTableData, any(data))
		var op drag.Op
		screen.Do(func() { op = sheet.DragEnteredCallback(di, geom.NewPoint(10, 10), 0) })
		return op
	}

	c.Equal(drag.Copy, enter(fromLibrary), "the buy panel takes library equipment")
	c.True(rerouteHighlightDrawn(t, screen, sheet), "the buy panel is highlighted for equipment it will take")
	screen.Do(sheet.DragExitedCallback)
	c.False(rerouteHighlightDrawn(t, screen, sheet), "nothing is highlighted once the drag has left")

	c.Equal(drag.None, enter(nil), "the buy panel declines an equipment drag with no rows behind it")
	c.False(rerouteHighlightDrawn(t, screen, sheet), "the buy panel is not highlighted for a drag it declines")
	screen.Do(sheet.DragExitedCallback)
}

// TestBuyPanelNoticeTellsMissingSellMoneyApart checks that the notice about change that can't be returned says the
// sheet's sell money couldn't be found when it has some but none of it is in the libraries, rather than that it has
// none and should add some.
func TestBuyPanelNoticeTellsMissingSellMoneyApart(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	sources := buyTestMoneyLibrary(t, c)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = []gurps.Source{{LibraryFile: sources[0].LibraryFile, TID: "missing"}}
		sheet.startBuying()
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	state := readBuyPanel(t, screen, sheet)
	c.True(state.noticeShown, "change that can't be returned is warned about")
	c.Equal("$5 of change can't be returned from the money paid, and none of the sheet's sell money could be found in the libraries.",
		state.notice)
}

// TestBuyPanelTakesTheFocusFromTheListItReplaces checks that starting to buy while the keyboard focus is in the carried
// equipment list, which leaves the page, moves the focus into the buy panel that takes its place rather than to the top
// of the sheet, and that a focus elsewhere on the sheet is left where it is.
func TestBuyPanelTakesTheFocusFromTheListItReplaces(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var inPanel, onPercent bool
	screen.Do(func() {
		sheet.CarriedEquipment.Table.RequestFocus()
		sheet.startBuying()
		focus := wnd.CurrentFocus()
		inPanel = focus != nil && unison.AncestorIsOrSelf(focus, sheet.buyPanel)
		onPercent = focus == sheet.buyPanel.percentField.AsPanel()
	})
	c.True(inPanel, "the focus must move into the buy panel")
	c.True(onPercent, "the focus must land on the first field of the buy panel")

	var stays bool
	screen.Do(func() {
		sheet.stopBuying()
		sheet.OtherEquipment.Table.RequestFocus()
		sheet.startBuying()
		stays = wnd.CurrentFocus() == sheet.OtherEquipment.Table.AsPanel()
	})
	c.True(stays, "a focus elsewhere on the sheet stays where it is")
}

// TestBuyPanelDropHighlightUsesTheDropColor checks that the buy panel tints its drop area, while equipment is dragged
// over it, in the color the lists tint their drop targets with.
func TestBuyPanelDropHighlightUsesTheDropColor(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var center color.NRGBA
	expected := unison.ThemeWarning.GetColor()
	screen.Do(func() {
		sheet.startBuying()
		p := sheet.buyPanel
		p.ValidateLayout()
		p.dragOver = true
		area := p.dropArea
		size := area.ContentRect(true).Size
		img, err := unison.NewImageFromDrawing(int(size.Width), int(size.Height), 72, func(gc *unison.Canvas) {
			area.DrawCallback(gc, geom.Rect{Size: size})
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer img.Dispose()
		pixels, err := img.ToNRGBA()
		if err != nil {
			t.Error(err)
			return
		}
		center = pixels.NRGBAAt(pixels.Rect.Dx()/2, pixels.Rect.Dy()/2)
	})
	near := func(a uint8, b int) bool { return max(int(a), b)-min(int(a), b) <= 2 }
	c.True(near(center.R, expected.Red()) && near(center.G, expected.Green()) && near(center.B, expected.Blue()),
		"the drop area must be tinted with the drop color %v, not %v", expected, center)
	c.True(center.A > 0 && center.A < 255, "the tint must be translucent, not %d", center.A)
}

// TestBuyPanelNameFollowsTheItemsWhileFocused checks that the suggested transaction name follows the items until the
// user edits it even while the name field has the keyboard focus, since the buttons that remove items don't take the
// focus from it, and that the purchase is recorded under the name shown.
func TestBuyPanelNameFollowsTheItemsWhileFocused(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 70, 11)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", sword, dagger)
	screen.Do(sheet.startBuying)
	if _, dropped := buyTestDrop(t, screen, sheet, library, sword, dagger); !dropped {
		t.Fatal("the sword and dagger must be dropped into the buy panel")
	}
	c.Equal("Bought Shortsword, Dagger", readBuyPanel(t, screen, sheet).name)
	screen.Do(func() {
		sheet.buyPanel.nameField.RequestFocus()
		sheet.buyPanel.items[1].remove.Click()
	})
	c.Equal("Bought Shortsword", readBuyPanel(t, screen, sheet).name,
		"the name follows the items although the field has the focus")

	screen.Do(func() { sheet.buyPanel.finalize.Click() })
	var record []*gurps.ShoppingRecordEntry
	screen.Do(func() { record = gurps.CloneShoppingRecordList(sheet.Entity().ShoppingRecord) })
	c.Equal(1, len(record))
	if len(record) == 1 {
		c.Equal("Bought Shortsword", record[0].Name, "the purchase is recorded under the name shown")
		c.Equal(1, len(record[0].Items))
	}
}

// TestBuyPanelFollowsChangesToTheSellMoneyLibrary checks that the change follows the libraries while the panel is up: a
// sell money source whose library file doesn't exist yet can't return change, and once the file is written the next
// sync returns change from it, as the entity's own lookup does.
func TestBuyPanelFollowsChangesToTheSellMoneyLibrary(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	user := gurps.GlobalSettings().Libraries.User()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Late Money" + gurps.EquipmentExt}
	path := filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))
	penny := buyTestItem(nil, "Copper Penny", "1", 1, gurps.MoneyTag)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = []gurps.Source{{LibraryFile: libFile, TID: penny.TID}}
		sheet.startBuying()
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	state := readBuyPanel(t, screen, sheet)
	c.Equal("$5 of change can't be returned from the money paid, and none of the sheet's sell money could be found in the libraries.",
		state.notice, "the library file doesn't exist yet")

	c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{penny}, path))
	screen.Do(func() { sheet.Rebuild(true) })
	state = readBuyPanel(t, screen, sheet)
	c.Equal("Pay automatically: Silver Eagle ×1 ($20); change: Copper Penny ×5 ($5)", state.summary,
		"the change comes from the library file once it exists")
	c.False(state.noticeShown)
}
