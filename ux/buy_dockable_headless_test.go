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
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
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
	uxtest.SwapForTest(t, &confirmPurchase, acceptPurchaseAsPlanned)
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

// acceptPurchaseAsPlanned stands in for the dialog that finalizes a purchase (see confirmPurchaseWithDialog), accepting
// the automatic payment and the dockable's merge setting without asking.
func acceptPurchaseAsPlanned(d *buyDockable, plan *buyPlan) (purchaseTerms, bool) {
	return purchaseTerms{chosen: paymentQuantities(plan.payments), merge: d.merge}, true
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

// buyTestState is what a test reads from the buy dockable.
type buyTestState struct {
	summary    string
	name       string
	rows       []string
	finalizeOn bool
}

// readBuyDockable reads the state of the sheet's buy dockable, failing the test if it has none.
func readBuyDockable(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) buyTestState {
	t.Helper()
	var state buyTestState
	found := false
	screen.Do(func() {
		p := buyDockableFor(sheet)
		if p == nil {
			return
		}
		found = true
		state.summary = p.summary.Title()
		state.name = p.nameField.Text()
		state.finalizeOn = p.finalize.Enabled()
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

// buyTestSetQuantity types text into the quantity field of the buy dockable's row at index.
func buyTestSetQuantity(screen *unison.HeadlessScreen, sheet *Sheet, index int, text string) {
	screen.Do(func() { buyDockableFor(sheet).items[index].field.SetText(text) })
}

// closeBuyDockable closes the sheet's buy dockable, if it has one, abandoning the purchase.
func closeBuyDockable(sheet *Sheet) {
	if d := buyDockableFor(sheet); d != nil {
		d.discardAndClose()
	}
}

// TestBuyDockableDropAndAutomaticPayment checks that equipment dropped on the sheet while buying lands in the buy dockable
// with a quantity of one and its price, that the carried money pays for it automatically, that the payment follows the
// quantities and the percentage, that a purchase the money can't cover can't be finalized, that a group's quantity
// can't be changed, and that a row can be removed again, all without touching the sheet or its undo.
func TestBuyDockableDropAndAutomaticPayment(t *testing.T) {
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
		showBuyDockable(sheet)
	})
	state := readBuyDockable(t, screen, sheet)
	c.Equal("Nothing to buy yet.", state.summary)
	c.False(state.finalizeOn, "nothing can be bought before something is dropped")

	op, dropped := buyTestDrop(t, screen, sheet, library, longbow)
	c.Equal(drag.Copy, op, "a drop from an equipment list is offered as a copy")
	c.True(dropped, "the drop is taken")
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45"}, state.rows)
	c.Equal("Total: $45", state.summary)
	c.Equal("Bought Longbow", state.name, "the transaction is named after what is bought")
	c.True(state.finalizeOn, "the purchase can be finalized")

	buyTestSetQuantity(screen, sheet, 0, "2")
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"2 | Longbow | $45 each | Total $90"}, state.rows)
	c.Equal("Total: $90", state.summary)

	buyTestSetQuantity(screen, sheet, 0, "4")
	state = readBuyDockable(t, screen, sheet)
	c.Equal("Not enough money carried: $180 needed, $151 carried.", state.summary)
	c.False(state.finalizeOn, "a purchase the money carried can't cover can't be finalized")

	buyTestSetQuantity(screen, sheet, 0, "1")
	screen.Do(func() { buyDockableFor(sheet).percentField.SetText("200") })
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $90 each | Total $90"}, state.rows, "the prices follow the percentage")
	screen.Do(func() { buyDockableFor(sheet).percentField.SetText("100") })

	op, dropped = buyTestDrop(t, screen, sheet, library, kit)
	c.Equal(drag.Copy, op)
	c.True(dropped)
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45", "1 | Kit | $10 each | Total $10"}, state.rows,
		"a group is bought once, its price that of what it holds")
	c.Equal("Total: $55", state.summary)
	c.Equal("Bought Longbow, Kit", state.name)
	var groupFieldEnabled bool
	screen.Do(func() { groupFieldEnabled = buyDockableFor(sheet).items[1].field.Enabled() })
	c.False(groupFieldEnabled, "a group's quantity can't be changed")

	screen.Do(func() {
		buyDockableFor(sheet).nameField.SetText("Archery")
		buyDockableFor(sheet).items[1].remove.Click()
	})
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Longbow | $45 each | Total $45"}, state.rows, "the row is removed")
	c.Equal("Archery", state.name, "a name the user typed is kept")
	screen.Do(func() { buyDockableFor(sheet).items[0].remove.Click() })
	state = readBuyDockable(t, screen, sheet)
	c.Nil(state.rows)
	c.Equal("Nothing to buy yet.", state.summary)
	c.False(state.finalizeOn)

	screen.Do(func() { modified = sheet.Modified() })
	c.False(modified, "nothing done in the buy dockable marks the sheet as modified")
	c.False(sheet.UndoManager().CanUndo(), "nothing done in the buy dockable is recorded for undo")
	c.Equal([]string{"Silver Eagle ×7", "Copper Penny ×11"}, buyTestCarried(screen, sheet))
}

// TestBuyDockableBuysTheSheetsOwnEquipment checks that equipment dragged from the sheet's own other equipment list is
// bought like any other, both when it reaches the buy dockable through the sheet's rerouting of drops and when it reaches
// the dockable directly, and that finalizing carries a copy while the original stays in the other equipment list.
func TestBuyDockableBuysTheSheetsOwnEquipment(t *testing.T) {
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
		showBuyDockable(sheet)
	})
	op, dropped := buyTestDrop(t, screen, sheet, sheet.OtherEquipment.Table, rope)
	c.Equal(drag.Copy, op, "the sheet's own equipment is offered as a copy")
	c.True(dropped, "the drop is taken")
	state := readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Rope | $5 each | Total $5"}, state.rows)
	c.Equal("Total: $5", state.summary)

	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	var canAccept, droppedDirectly bool
	var updated drag.Op
	screen.Do(func() {
		p := buyDockableFor(sheet).AsPanel()
		canAccept = p.CanAcceptDropCallback(di)
		updated = p.DragUpdatedCallback(di, geom.Point{}, 0)
		droppedDirectly = p.DropCallback(di, geom.Point{}, 0)
	})
	c.True(canAccept, "the dockable accepts the sheet's own equipment")
	c.Equal(drag.Copy, updated, "moving over the dockable offers to take it")
	c.True(droppedDirectly, "dropping it on the dockable adds it")
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Rope | $5 each | Total $5", "1 | Rope | $5 each | Total $5"}, state.rows)
	c.Equal("Total: $10", state.summary)

	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
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

// TestBuyDockableReturnsChange checks that an overpayment warns that the change can't be returned while the sheet has no
// sell money, that the change is worked out from the sell money once the sheet has some, and that finalizing spends
// the money, carries the item and the change, records the purchase, closes the dockable with the focus on the carried
// equipment list and can be undone and redone as one edit.
func TestBuyDockableReturnsChange(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	sources := buyTestMoneyLibrary(t, c)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = nil
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	state := readBuyDockable(t, screen, sheet)
	c.Equal("Total: $15", state.summary)
	c.True(state.finalizeOn, "an overpayment can still be finalized")
	var change string
	screen.Do(func() { change = describePayouts(buyDockableFor(sheet).planPayment(nil).change) })
	c.Equal("", change, "without sell money, no change can be returned from an eagle")

	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = sources
		sheet.Rebuild(true)
		change = describePayouts(buyDockableFor(sheet).planPayment(nil).change)
	})
	c.Equal("Copper Penny ×5", change, "the dockable follows a change to the sell money")

	screen.Do(func() {
		buyDockableFor(sheet).finalize.RequestFocus()
		buyDockableFor(sheet).finalize.Click()
	})
	var buying, listFocused bool
	screen.Do(func() {
		buying = buyDockableFor(sheet) != nil
		listFocused = wnd.Focus() == sheet.CarriedEquipment.Table.AsPanel()
	})
	c.False(buying, "finalizing closes the buy dockable")
	c.True(listFocused, "the focus moves from the button that finalized to the carried equipment list")
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
	c.False(daggerEquipped, "the item bought is carried unequipped")
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

// TestBuyDockableFinalizeMerges checks that a purchase merges into a matching carried item when asked to and is carried
// separately otherwise.
func TestBuyDockableFinalizeMerges(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return append(buyTestCoins(entity, 7, 11), buyTestItem(entity, "Leather Gloves", "30", 1))
	})
	gloves := buyTestItem(nil, "Leather Gloves", "30", 1)
	library := buyTestLibrary(t, screen, "", gloves)

	screen.Do(func() { showBuyDockable(sheet) })
	_, dropped := buyTestDrop(t, screen, sheet, library, gloves)
	c.True(dropped)
	c.Equal("Total: $30", readBuyDockable(t, screen, sheet).summary)
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
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
		showBuyDockable(sheet)
	})
	_, dropped = buyTestDrop(t, screen, sheet, library, gloves)
	c.True(dropped)
	uxtest.SwapForTest(t, &confirmPurchase, func(_ *buyDockable, plan *buyPlan) (purchaseTerms, bool) {
		return purchaseTerms{chosen: paymentQuantities(plan.payments), merge: false}, true
	})
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Silver Eagle ×6", "Copper Penny ×1", "Leather Gloves ×1", "Leather Gloves ×1"},
		buyTestCarried(screen, sheet), "without merging, the gloves bought are carried separately")
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

// TestBuyDockableAttachesAndRemovesModifiers drops an equipment modifier from a library onto an item waiting to be bought
// and takes it off again: the modifier goes only to the item it is dropped on, the price follows it, the row under the
// item shows it with a button that removes it, and a plain modifier asks nothing.
func TestBuyDockableAttachesAndRemovesModifiers(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	library := buyTestLibrary(t, screen, "", sword)
	screen.Do(func() { showBuyDockable(sheet) })
	if op, dropped := buyTestDrop(t, screen, sheet, library, sword); op == drag.None || !dropped {
		t.Fatal("the sword must be dropped into the buy dockable")
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
		p := buyDockableFor(sheet)
		p.ValidateLayout()
		item := p.items[0]
		hint := p.PointFromRoot(p.hint.PointToRoot(p.hint.ContentRect(true).Center()))
		row := p.PointFromRoot(item.name.PointToRoot(item.name.ContentRect(true).Center()))
		c.True(p.CanAcceptDropCallback(di), "the dockable must take equipment modifier drags")
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
		item := buyDockableFor(sheet).items[0]
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
	c.Contains(readBuyDockable(t, screen, sheet).rows[0], "$150", "the price must follow the modifier")
	if remove == nil {
		t.Fatal("the modifier's row must have a remove button")
	}

	screen.Do(func() { remove.Click() })
	screen.Do(func() {
		item := buyDockableFor(sheet).items[0]
		modifierCount = len(item.eqp.Modifiers)
		rowCount = len(item.modifierRows)
		value = item.eqp.AdjustedValue()
	})
	c.Equal(0, modifierCount, "removing takes the modifier off the item")
	c.Equal(0, rowCount, "removing takes the modifier's row away")
	c.Equal(fxp.FromInteger(100), value)
	c.Contains(readBuyDockable(t, screen, sheet).rows[0], "$100")
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

// TestBuyDockableClaimsEquipmentDroppedElsewhereOnTheSheet checks that, while the buy dockable is up, equipment dragged from
// elsewhere onto the sheet's other equipment list is left to the sheet, which reroutes it to the buy dockable, while the
// sheet's own equipment is still taken by the list, so that its rows can be moved about.
func TestBuyDockableClaimsEquipmentDroppedElsewhereOnTheSheet(t *testing.T) {
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
	screen.Do(func() { showBuyDockable(sheet) })
	c.False(accepts(), "other equipment leaves library equipment to the buy dockable while buying")
	cancel()

	cancel = startRowDrag(t, screen, wnd, sheet.OtherEquipment.Table, 0)
	c.True(accepts(), "other equipment still takes the sheet's own equipment while buying")
	cancel()
	var otherAfterDrags int
	screen.Do(func() { otherAfterDrags = len(sheet.Entity().OtherEquipment) })
	c.Equal(1, otherAfterDrags, "the canceled drags must change nothing")

	if op, dropped := buyTestDrop(t, screen, sheet, library, sword); op == drag.None || !dropped {
		t.Fatal("the sword must be dropped into the buy dockable")
	}
	var bought, other int
	screen.Do(func() {
		bought = len(buyDockableFor(sheet).items)
		other = len(sheet.Entity().OtherEquipment)
	})
	c.Equal(1, bought, "the drop must land in the buy dockable")
	c.Equal(1, other, "the other equipment must be untouched")
}

// TestBuyDockableChangeFollowsEditsToTheMoneyPaid checks that change drawn from the money paid is worked out again when
// that money is edited without its value changing, such as by being renamed, so that the change carried is a copy of
// the money as it now is and merges with it.
func TestBuyDockableChangeFollowsEditsToTheMoneyPaid(t *testing.T) {
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
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	chosen := map[tid.TID]int{ten.TID: 1, penny.TID: 2}
	uxtest.SwapForTest(t, &confirmPurchase, func(_ *buyDockable, _ *buyPlan) (purchaseTerms, bool) {
		return purchaseTerms{chosen: chosen, merge: true}, true
	})
	// The change is worked out, and remembered, before the money is renamed, as it is while the dialog is up.
	var before, after string
	screen.Do(func() {
		before = describePayouts(buyDockableFor(sheet).planPayment(chosen).change)
		penny.Name = "Penny"
		sheet.Rebuild(true)
		after = describePayouts(buyDockableFor(sheet).planPayment(chosen).change)
	})
	c.Equal("Copper Penny ×2", before)
	c.Equal("Penny ×2", after, "the change is a copy of the money as it now is")
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Penny ×5", "Dagger ×1"}, buyTestCarried(screen, sheet),
		"the change merges with the money it was copied from")
}

// buyTestDropModifier drags the modifier table's rows onto the buy dockable's item and drops them there, returning whether
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
		p := buyDockableFor(sheet)
		p.ValidateLayout()
		row := p.PointFromRoot(item.name.PointToRoot(item.name.ContentRect(true).Center()))
		p.DragUpdatedCallback(di, row, 0)
		dropped = p.DropCallback(di, row, 0)
	})
	return dropped
}

// TestBuyDockableKeepsTheFocusWhenRowsChange checks that the keyboard focus is not left on a button the dockable has
// discarded: removing a row with its own button moves the focus to the row that takes its place, or to the percent
// field once no row is left; a modifier's button keeps the focus when the rows are built anew for another item; and
// removing a modifier with its button moves the focus to its item's row.
func TestBuyDockableKeepsTheFocusWhenRowsChange(t *testing.T) {
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
	screen.Do(func() { showBuyDockable(sheet) })
	if _, dropped := buyTestDrop(t, screen, sheet, library, sword, dagger); !dropped {
		t.Fatal("the sword and dagger must be dropped into the buy dockable")
	}
	focusIs := func(what string, expected func(p *buyDockable) unison.Paneler) {
		t.Helper()
		var ok bool
		screen.Do(func() {
			target := expected(buyDockableFor(sheet))
			ok = target != nil && wnd.CurrentFocus() == target.AsPanel()
		})
		c.True(ok, "the focus must be on %s", what)
	}

	screen.Do(func() {
		remove := buyDockableFor(sheet).items[1].remove
		remove.RequestFocus()
		remove.Click()
	})
	focusIs("the first row's quantity field after removing the last row", func(p *buyDockable) unison.Paneler {
		if len(p.items) != 1 {
			return nil
		}
		return p.items[0].field
	})

	if !buyTestDropModifier(t, screen, sheet, modifiers, buyDockableFor(sheet).items[0]) {
		t.Fatal("the modifier must be dropped onto the sword")
	}
	modifierButton := func(p *buyDockable) unison.Paneler {
		if len(p.items) != 0 && len(p.items[0].modifierRows) == 1 {
			return p.items[0].modifierRows[0].widgets[2]
		}
		return nil
	}
	screen.Do(func() { modifierButton(buyDockableFor(sheet)).AsPanel().RequestFocus() })
	if _, dropped := buyTestDrop(t, screen, sheet, library, rope); !dropped {
		t.Fatal("the rope must be dropped into the buy dockable")
	}
	focusIs("the modifier's button, built anew, after another item is dropped", modifierButton)

	screen.Do(func() {
		if button, ok := modifierButton(buyDockableFor(sheet)).(*unison.Button); ok {
			button.Click()
		}
	})
	focusIs("the item's quantity field after removing its modifier", func(p *buyDockable) unison.Paneler {
		if len(p.items) != 2 || len(p.items[0].modifierRows) != 0 {
			return nil
		}
		return p.items[0].field
	})

	for range 2 {
		screen.Do(func() {
			remove := buyDockableFor(sheet).items[0].remove
			remove.RequestFocus()
			remove.Click()
		})
	}
	focusIs("the percent field once no row is left", func(p *buyDockable) unison.Paneler {
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

// TestSheetDropReroutingHighlightsOnlyAWillingTarget checks that the sheet highlights itself, as the whole of it stands
// for the buy dockable it reroutes a drag to, only while the dockable will take the drop: for equipment it would buy, not
// for an equipment drag with no rows behind it, which it declines, and not once the drag has left.
func TestSheetDropReroutingHighlightsOnlyAWillingTarget(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	screen.Do(func() { showBuyDockable(sheet) })
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

	c.Equal(drag.Copy, enter(fromLibrary), "the buy dockable takes library equipment")
	c.True(rerouteHighlightDrawn(t, screen, sheet), "the sheet is highlighted for equipment the buy dockable will take")
	screen.Do(sheet.DragExitedCallback)
	c.False(rerouteHighlightDrawn(t, screen, sheet), "nothing is highlighted once the drag has left")

	c.Equal(drag.None, enter(nil), "the buy dockable declines an equipment drag with no rows behind it")
	c.False(rerouteHighlightDrawn(t, screen, sheet), "the sheet is not highlighted for a drag the buy dockable declines")
	screen.Do(sheet.DragExitedCallback)
}

// TestBuyDockableOpensOncePerSheet checks that opening the buy dockable puts the keyboard focus on its first field, that
// asking for it again brings the one already open forward rather than opening another, that its title follows the
// character's name, and that it closes with its sheet.
func TestBuyDockableOpensOncePerSheet(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var first, again *buyDockable
	var onPercent bool
	var title, renamed string
	var count int
	screen.Do(func() {
		sheet.CarriedEquipment.Table.RequestFocus()
		first = showBuyDockable(sheet)
		onPercent = wnd.CurrentFocus() == first.percentField.AsPanel()
		title = first.Title()
		sheet.OtherEquipment.Table.RequestFocus()
		again = showBuyDockable(sheet)
		for _, d := range AllDockables() {
			if _, ok := d.AsPanel().Self.(*buyDockable); ok {
				count++
			}
		}
		sheet.Entity().Profile.Name = "Bob"
		sheet.Rebuild(true)
		renamed = first.Title()
	})
	c.True(onPercent, "the focus must land on the first field of the buy dockable")
	c.True(first == again, "asking to buy again must bring the open dockable forward")
	c.Equal(1, count, "a sheet must have one buy dockable at most")
	c.Equal("Buy Equipment: "+sheet.Entity().Profile.Name, renamed)
	c.NotEqual(title, renamed, "the title must follow the character's name")

	var open bool
	screen.Do(func() {
		sheet.markUnmodified()
		sheet.AttemptClose()
		open = buyDockableFor(sheet) != nil
	})
	c.False(open, "closing the sheet must close its buy dockable")
}

// TestBuyDockableDropHighlightUsesTheDropColor checks that the buy dockable tints its drop area, while equipment is dragged
// over it, in the color the lists tint their drop targets with, laid translucently over the background a table gives
// its rows, which the drop area shares so that its banding stands out as a table's does.
func TestBuyDockableDropHighlightUsesTheDropColor(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var center color.NRGBA
	background := unison.ThemeBelowSurface.GetColor()
	tint := unison.ThemeWarning.GetColor()
	blend := func(bg, fg int) int { return int(math.Round(float64(bg)*0.7 + float64(fg)*0.3)) }
	expected := color.NRGBA{
		R: uint8(blend(background.Red(), tint.Red())),
		G: uint8(blend(background.Green(), tint.Green())),
		B: uint8(blend(background.Blue(), tint.Blue())),
		A: 255,
	}
	screen.Do(func() {
		showBuyDockable(sheet)
		p := buyDockableFor(sheet)
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
	near := func(a, b uint8) bool { return max(a, b)-min(a, b) <= 3 }
	c.True(near(center.R, expected.R) && near(center.G, expected.G) && near(center.B, expected.B),
		"the drop area must be tinted with the drop color over its background, %v, not %v", expected, center)
	c.Equal(uint8(255), center.A, "the background under the tint must be opaque")
}

// TestBuyDockableNameFollowsTheItemsWhileFocused checks that the suggested transaction name follows the items until the
// user edits it even while the name field has the keyboard focus, since the buttons that remove items don't take the
// focus from it, and that the purchase is recorded under the name shown.
func TestBuyDockableNameFollowsTheItemsWhileFocused(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 70, 11)
	})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", sword, dagger)
	screen.Do(func() { showBuyDockable(sheet) })
	if _, dropped := buyTestDrop(t, screen, sheet, library, sword, dagger); !dropped {
		t.Fatal("the sword and dagger must be dropped into the buy dockable")
	}
	c.Equal("Bought Shortsword, Dagger", readBuyDockable(t, screen, sheet).name)
	screen.Do(func() {
		buyDockableFor(sheet).nameField.RequestFocus()
		buyDockableFor(sheet).items[1].remove.Click()
	})
	c.Equal("Bought Shortsword", readBuyDockable(t, screen, sheet).name,
		"the name follows the items although the field has the focus")

	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	var record []*gurps.ShoppingRecordEntry
	screen.Do(func() { record = gurps.CloneShoppingRecordList(sheet.Entity().ShoppingRecord) })
	c.Equal(1, len(record))
	if len(record) == 1 {
		c.Equal("Bought Shortsword", record[0].Name, "the purchase is recorded under the name shown")
		c.Equal(1, len(record[0].Items))
	}
}

// TestBuyDockableFollowsChangesToTheSellMoneyLibrary checks that the change follows the libraries while the dockable is up: a
// sell money source whose library file doesn't exist yet can't return change, and once the file is written the next
// sync returns change from it, as the entity's own lookup does.
func TestBuyDockableFollowsChangesToTheSellMoneyLibrary(t *testing.T) {
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
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	change := func() (found int, change string) {
		screen.Do(func() {
			plan := buyDockableFor(sheet).planPayment(nil)
			found, change = len(plan.sellMoney), describePayouts(plan.change)
		})
		return found, change
	}
	found, returned := change()
	c.Equal(0, found, "the library file doesn't exist yet")
	c.Equal("", returned, "so no change can be returned")

	c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{penny}, path))
	screen.Do(func() { sheet.Rebuild(true) })
	found, returned = change()
	c.Equal(1, found, "the sell money is found once the file exists")
	c.Equal("Copper Penny ×5", returned, "and the change comes from it")
}

// TestBuyDockableEnablesWhatItBuys checks that equipment copied into the buy dockable from a sheet where it was disabled, or
// where a container held something disabled, is enabled on the way in, so that it is priced in full and carried
// enabled, and that the original is left as it was.
func TestBuyDockableEnablesWhatItBuys(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 0, 20)
	})
	var rope, kit, bandage *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		rope = buyTestItem(entity, "Rope", "5", 1)
		rope.SetDisabled(true)
		kit = gurps.NewEquipment(entity, nil, true)
		kit.Name = "Kit"
		kit.BaseValue = "10"
		bandage = gurps.NewEquipment(entity, kit, false)
		bandage.Name = "Bandage"
		bandage.BaseValue = "2"
		bandage.SetDisabled(true)
		kit.Children = []*gurps.Equipment{bandage}
		entity.OtherEquipment = []*gurps.Equipment{rope, kit}
		entity.Recalculate()
		sheet.Rebuild(true)
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, sheet.OtherEquipment.Table, rope, kit)
	c.True(dropped, "the drop is taken")
	state := readBuyDockable(t, screen, sheet)
	c.Equal([]string{"1 | Rope | $5 each | Total $5", "1 | Kit | $12 each | Total $12"}, state.rows,
		"the disabled rope is priced, as is the kit with its disabled bandage")
	c.Equal("Total: $17", state.summary)

	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Silver Eagle ×0", "Copper Penny ×3", "Rope ×1", "Kit ×1"}, buyTestCarried(screen, sheet))
	screen.Do(func() {
		carried := sheet.Entity().CarriedEquipment
		gurps.Traverse(func(eqp *gurps.Equipment) bool {
			c.True(eqp.Enabled(), "%s is carried enabled", eqp.Name)
			c.False(eqp.Equipped, "%s is carried unequipped", eqp.Name)
			return false
		}, false, false, carried[2:]...)
		c.Equal(1, len(carried[3].Children), "the kit is carried with its bandage")
		if len(carried[3].Children) == 1 {
			c.Equal("Bandage", carried[3].Children[0].Name)
		}
		c.True(rope.IsDisabled(), "the original rope stays disabled")
		c.True(bandage.IsDisabled(), "the original bandage stays disabled")
	})
}

// TestBuyDockableTotalSaturatesRatherThanWrapping checks that a total made of prices too large for fxp to hold saturates at
// its maximum, so that the money carried is seen to fall short and the purchase can't be finalized, rather than wrapping
// around to a negative total that nothing pays for and that lets everything be carried for free.
func TestBuyDockableTotalSaturatesRatherThanWrapping(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	hoards := make([]*gurps.Equipment, 103)
	for i := range hoards {
		hoards[i] = buyTestItem(nil, fmt.Sprintf("Hoard %d", i+1), "9000000000000", 1)
	}
	library := buyTestLibrary(t, screen, "", hoards...)
	screen.Do(func() { showBuyDockable(sheet) })
	_, dropped := buyTestDrop(t, screen, sheet, library, hoards...)
	c.True(dropped)
	state := readBuyDockable(t, screen, sheet)
	c.Equal("Not enough money carried: $922,337,203,685,477.5807 needed, $151 carried.", state.summary,
		"the total is held at fxp's maximum")
	c.False(state.finalizeOn, "the purchase can't be finalized")
}

// TestBuyDockableChangeFollowsEditsToTheSellMoneyLibrary checks that the change follows a change to the sell money in the
// library that leaves its value alone, such as renaming the coin: the dockable looks the sell money up each time, so the
// change it shows and hands over must be made of the copies that lookup returns, not of those an earlier lookup did.
func TestBuyDockableChangeFollowsEditsToTheSellMoneyLibrary(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	user := gurps.GlobalSettings().Libraries.User()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Renamed Money" + gurps.EquipmentExt}
	path := filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))
	penny := buyTestItem(nil, "Copper Penny", "1", 1, gurps.MoneyTag)
	c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{penny}, path))
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = []gurps.Source{{LibraryFile: libFile, TID: penny.TID}}
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	var before, after string
	screen.Do(func() { before = describePayouts(buyDockableFor(sheet).planPayment(nil).change) })
	c.Equal("Copper Penny ×5", before)

	penny.Name = "Penny"
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{penny}, path))
	screen.Do(func() {
		sheet.Rebuild(true)
		after = describePayouts(buyDockableFor(sheet).planPayment(nil).change)
	})
	c.Equal("Penny ×5", after, "the change is made of the coin as the library now has it")
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Silver Eagle ×6", "Dagger ×1", "Penny ×5"}, buyTestCarried(screen, sheet))
}

// buyTestEditorsOpenFor returns how many equipment editors are open on the items.
func buyTestEditorsOpenFor(screen *unison.HeadlessScreen, items ...*gurps.Equipment) int {
	var count int
	screen.Do(func() {
		count = len(AllMatchingDockables(func(d unison.Dockable) bool {
			e, ok := d.(*editor[*gurps.Equipment, *gurps.EquipmentEditData])
			return ok && slices.Contains(items, e.target)
		}))
	})
	return count
}

// TestBuyDockableFinalizeClosesTheEditorsOfWhatItChanges checks that finalizing closes the editors open on everything the
// purchase changes, as a sale does, since an editor doesn't follow changes to its item and would put its stale quantity
// back when applied: the carried money spent, a container the payment splits apart, and the carried items the purchase
// and the change merge into. The editor of an item the purchase leaves alone stays open.
func TestBuyDockableFinalizeClosesTheEditorsOfWhatItChanges(t *testing.T) {
	c := check.New(t)
	var eagles, pile, purse, pursePennies, gloves, rope *gurps.Equipment
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		eagles = buyTestItem(entity, "Silver Eagle", "20", 7, gurps.MoneyTag)
		pile = buyTestItem(entity, "Copper Penny", "1", 11, gurps.MoneyTag)
		purse = gurps.NewEquipment(entity, nil, true)
		purse.Name = "Purse"
		purse.Quantity = fxp.Two
		pursePennies = buyTestItem(entity, "Copper Penny", "1", 3, gurps.MoneyTag)
		pursePennies.SetParent(purse)
		purse.Children = []*gurps.Equipment{pursePennies}
		gloves = buyTestItem(entity, "Leather Gloves", "30", 1)
		rope = buyTestItem(entity, "Rope", "5", 1)
		return []*gurps.Equipment{eagles, pile, purse, gloves, rope}
	})
	sources := buyTestMoneyLibrary(t, c)
	library := buyTestLibrary(t, screen, "", buyTestItem(nil, "Leather Gloves", "30", 1))
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = sources
		for _, eqp := range []*gurps.Equipment{eagles, pile, purse, pursePennies, gloves, rope} {
			EditEquipment(sheet, eqp, true)
		}
		showBuyDockable(sheet)
	})
	c.Equal(6, buyTestEditorsOpenFor(screen, eagles, pile, purse, pursePennies, gloves, rope), "the editors must be open")
	var rows []*Node[*gurps.Equipment]
	screen.Do(func() { rows = library.RootRows() })
	_, dropped := buyTestDrop(t, screen, sheet, library, rows[0].Data())
	c.True(dropped)
	chosen := map[tid.TID]int{eagles.TID: 2, pursePennies.TID: 1}
	uxtest.SwapForTest(t, &confirmPurchase, func(_ *buyDockable, _ *buyPlan) (purchaseTerms, bool) {
		return purchaseTerms{chosen: chosen, merge: true}, true
	})
	var change string
	screen.Do(func() { change = describePayouts(buyDockableFor(sheet).planPayment(chosen).change) })
	c.Equal("Copper Penny ×11", change, "the change for the money chosen comes back in pennies")
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Silver Eagle ×5", "Copper Penny ×22", "Purse ×1", "Purse ×1", "Leather Gloves ×2", "Rope ×1"},
		buyTestCarried(screen, sheet), "the eagles and a purse's penny are spent, splitting the purse, and the gloves "+
			"and the change merge into those carried")
	c.Equal(0, buyTestEditorsOpenFor(screen, eagles, pursePennies), "the editors of the money spent are closed")
	c.Equal(0, buyTestEditorsOpenFor(screen, purse), "the editor of the purse the payment splits apart is closed")
	c.Equal(0, buyTestEditorsOpenFor(screen, gloves), "the editor of the gloves the purchase merges into is closed")
	c.Equal(0, buyTestEditorsOpenFor(screen, pile), "the editor of the pennies the change merges into is closed")
	c.Equal(1, buyTestEditorsOpenFor(screen, rope), "the editor of an item the purchase leaves alone stays open")
}

// TestBuyDockableDoesNotFinalizeOnAPaymentASavedEditorChanged checks that when an editor open on the money paid with is
// saved at the prompt put up as finalizing closes it, and saving changes the payment, the purchase isn't made on the
// terms the user accepted: the dockable, brought up to date by the save, stays for the user to look over.
func TestBuyDockableDoesNotFinalizeOnAPaymentASavedEditorChanged(t *testing.T) {
	c := check.New(t)
	var eagles *gurps.Equipment
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		eagles = buyTestItem(entity, "Silver Eagle", "20", 7, gurps.MoneyTag)
		return []*gurps.Equipment{eagles}
	})
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		e := EditEquipment(sheet, eagles, true)
		e.editorData.Quantity = 0
		e.MarkModified(nil)
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	c.Equal("Total: $15", readBuyDockable(t, screen, sheet).summary)
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	screen.Sync() // Finalizing is posted as a task, which puts up the editor's prompt.
	_, prompt := modalDialog(t, screen, wnd)
	screen.Click(screen.PanelCenter(dialogButton(t, screen, prompt, unison.ModalResponseOK)))
	c.Equal(0, buyTestEditorsOpenFor(screen, eagles), "the editor is saved and closed")
	c.Equal([]string{"Silver Eagle ×0"}, buyTestCarried(screen, sheet), "the eagles are as saved and nothing is bought")
	state := readBuyDockable(t, screen, sheet)
	c.Equal("Not enough money carried: $15 needed, $0 carried.", state.summary, "the dockable shows the payment as it now stands")
	c.False(state.finalizeOn)
	var records int
	screen.Do(func() { records = len(sheet.Entity().ShoppingRecord) })
	c.Equal(0, records)
}

// TestBuyDockableFinalizingLeavesAFocusElsewhereAlone checks that finalizing a purchase while the keyboard focus is
// outside the buy dockable, as when an assistive technology presses the button, closes the dockable without moving the
// focus to the carried equipment list.
func TestBuyDockableFinalizingLeavesAFocusElsewhereAlone(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() { showBuyDockable(sheet) })
	if _, dropped := buyTestDrop(t, screen, sheet, library, dagger); !dropped {
		t.Fatal("the dagger must be dropped into the buy dockable")
	}
	var open, stays bool
	screen.Do(func() {
		library.RequestFocus()
		buyDockableFor(sheet).finalizePurchase()
		open = buyDockableFor(sheet) != nil
		stays = wnd.CurrentFocus() == library.AsPanel()
	})
	c.False(open, "finalizing must close the dockable")
	c.True(stays, "a focus outside the dockable must stay where it is")
	c.Equal([]string{"Silver Eagle ×6", "Copper Penny ×11", "Dagger ×1"}, buyTestCarried(screen, sheet))
}

// TestBuyDockableFinalizesWithUnequippedSellMoney checks that a purchase returning change in a library coin saved without
// being equipped, as the Dungeon Fantasy RPG coins are, is finalized: predicting where the change merges must not change
// the coin, or the payment would no longer look like the one the user accepted.
func TestBuyDockableFinalizesWithUnequippedSellMoney(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 0)[:1]
	})
	user := gurps.GlobalSettings().Libraries.User()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Unequipped Change" + gurps.EquipmentExt}
	path := filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))
	penny := buyTestItem(nil, "Copper Penny", "1", 1, gurps.MoneyTag)
	penny.Equipped = false
	c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{penny}, path))
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = []gurps.Source{{LibraryFile: libFile, TID: penny.TID}}
		showBuyDockable(sheet)
	})
	_, dropped := buyTestDrop(t, screen, sheet, library, dagger)
	c.True(dropped)
	c.Equal("Total: $15", readBuyDockable(t, screen, sheet).summary)
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	var buying bool
	screen.Do(func() { buying = (buyDockableFor(sheet) != nil) })
	c.False(buying, "the purchase is finalized")
	c.Equal([]string{"Silver Eagle ×6", "Dagger ×1", "Copper Penny ×5"}, buyTestCarried(screen, sheet))
}

// TestBuyDockableFollowsEditsToTheCarriedEquipment checks that the carried equipment list stays editable while the buy
// dockable is open and that the payment follows what is done to it: deleting the money paid with through the list
// leaves the purchase short and unable to be finalized, and undoing the deletion restores the payment.
func TestBuyDockableFollowsEditsToTheCarriedEquipment(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return append(buyTestCoins(entity, 2, 5), buyTestItem(entity, "Tent", "50", 1))
	})
	longbow := buyTestItem(nil, "Longbow", "45", 1)
	library := buyTestLibrary(t, screen, "", longbow)
	screen.Do(func() { showBuyDockable(sheet) })
	if _, dropped := buyTestDrop(t, screen, sheet, library, longbow); !dropped {
		t.Fatal("the longbow must be dropped into the buy dockable")
	}
	total := "Total: $45"
	state := readBuyDockable(t, screen, sheet)
	c.Equal(total, state.summary)
	c.True(state.finalizeOn)

	table := sheet.CarriedEquipment.Table
	var canDelete bool
	screen.Do(func() {
		table.SelectByIndex(0)
		canDelete = table.CanPerformCmd(table, unison.DeleteItemID)
		table.PerformCmd(table, unison.DeleteItemID)
	})
	c.True(canDelete, "the carried equipment list must stay editable while buying")
	c.Equal([]string{"Copper Penny ×5", "Tent ×1"}, buyTestCarried(screen, sheet), "the money must be deleted")
	state = readBuyDockable(t, screen, sheet)
	c.Equal("Not enough money carried: $45 needed, $5 carried.", state.summary, "the payment must follow the deletion")
	c.False(state.finalizeOn, "the purchase can't be finalized once the money is gone")

	screen.Do(func() { sheet.UndoManager().Undo() })
	c.Equal([]string{"Silver Eagle ×2", "Copper Penny ×5", "Tent ×1"}, buyTestCarried(screen, sheet))
	state = readBuyDockable(t, screen, sheet)
	c.Equal(total, state.summary, "undoing the deletion must restore the payment")
	c.True(state.finalizeOn)
}

// TestBuyMoreOfUnequippedCarriedEquipment checks the purchase of more of something carried unequipped, such as bolts in
// a quiver, from the sheet's own carried list: what is bought merges into it, since the equipped state plays no part in
// matching, and leaves it unequipped.
func TestBuyMoreOfUnequippedCarriedEquipment(t *testing.T) {
	c := check.New(t)
	var bolts *gurps.Equipment
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		bolts = buyTestItem(entity, "Crossbow Bolt", "2", 20)
		bolts.Equipped = false
		return append(buyTestCoins(entity, 7, 11), bolts)
	})
	screen.Do(func() {
		table := sheet.CarriedEquipment.Table
		table.SelectByIndex(2)
		table.PerformCmd(table, BuySelectedEquipmentItemID)
	})
	state := readBuyDockable(t, screen, sheet)
	c.Equal([]string{"20 | Crossbow Bolt | $2 each | Total $40"}, state.rows)
	screen.Do(func() { buyDockableFor(sheet).finalize.Click() })
	c.Equal([]string{"Silver Eagle ×5", "Copper Penny ×11", "Crossbow Bolt ×40"}, buyTestCarried(screen, sheet),
		"the bolts bought merge into the unequipped bolts carried")
	var equipped bool
	screen.Do(func() { equipped = bolts.Equipped })
	c.False(equipped, "the bolts stay unequipped")
}

// TestBuySelectedEquipmentCommand checks the way to buy without dragging: an equipment list's menu offers Buy, which
// sends the selected rows to the buy dockable of the one open sheet, opening it, or of the sheet the list is on, and
// asks which sheet to buy for when several are open. The command is disabled while no sheet is open.
func TestBuySelectedEquipmentCommand(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	forbidModifierPrompts(t)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	sword := buyTestItem(nil, "Shortsword", "100", 1)
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", sword, dagger)
	var enabledWithoutSheet bool
	screen.Do(func() {
		library.SelectAll()
		enabledWithoutSheet = library.CanPerformCmd(library, BuySelectedEquipmentItemID)
	})
	c.False(enabledWithoutSheet, "Buy must be disabled while no sheet is open")

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		entity.CarriedEquipment = buyTestCoins(entity, 7, 11)
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	screen.Click(cellScreenPoint(t, screen, wnd, library, 0, gurps.EquipmentDescriptionColumn))
	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, "Buy…"), "an equipment library's menu must offer Buy; it offers %v", titles)
	closeContextMenu(c, screen, wnd)
	screen.Do(func() { library.PerformCmd(library, BuySelectedEquipmentItemID) })
	swordRow := "1 | Shortsword | $100 each | Total $100"
	state := readBuyDockable(t, screen, sheet)
	c.Equal([]string{swordRow}, state.rows, "Buy must open the one sheet's buy dockable with the selection")

	// From the sheet's own list, the sheet buys for itself.
	screen.Do(func() {
		entity := sheet.Entity()
		entity.OtherEquipment = []*gurps.Equipment{buyTestItem(entity, "Rope", "5", 1)}
		sheet.Rebuild(true)
		other := sheet.OtherEquipment.Table
		other.SelectByIndex(0)
		other.PerformCmd(other, BuySelectedEquipmentItemID)
	})
	ropeRow := "1 | Rope | $5 each | Total $5"
	state = readBuyDockable(t, screen, sheet)
	c.Equal([]string{swordRow, ropeRow}, state.rows, "a sheet's own list must buy for that sheet")

	// With another sheet open, a library asks which sheet to buy for. Canceling buys for none of them; accepting buys
	// for the one chosen.
	other, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	buyDagger := func(response int) {
		t.Helper()
		screen.Do(func() { library.SelectByIndex(1) })
		if !screen.Post(func() { library.PerformCmd(library, BuySelectedEquipmentItemID) }) {
			t.Fatal("Buy could not be performed")
		}
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Do(func() { dialog.StopModal(response) })
	}
	daggerRows := func() (count int) {
		screen.Do(func() {
			for _, s := range []*Sheet{sheet, other} {
				if d := buyDockableFor(s); d != nil {
					for _, item := range d.items {
						if item.eqp.Name == "Dagger" {
							count++
						}
					}
				}
			}
		})
		return count
	}
	buyDagger(unison.ModalResponseCancel)
	c.Equal(0, daggerRows(), "canceling the choice of sheet must buy nothing")
	buyDagger(unison.ModalResponseOK)
	c.Equal(1, daggerRows(), "accepting the choice of sheet must buy for that one sheet")
}

// TestBuyDockableAsksBeforeDiscardingAPurchase checks that closing the dockable with a purchase under way asks, as
// closing an editor with pending changes does: Cancel keeps the dockable open, No discards the purchase and Yes
// finalizes it; that closing the sheet asks the same way; that a purchase the money carried doesn't cover is only
// offered for discarding; and that the dockable is marked modified only while a purchase is under way.
func TestBuyDockableAsksBeforeDiscardingAPurchase(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	dagger := buyTestItem(nil, "Dagger", "15", 1)
	library := buyTestLibrary(t, screen, "", dagger)
	before := []string{"Silver Eagle ×7", "Copper Penny ×11"}
	var d *buyDockable
	modified := func() bool {
		var result bool
		screen.Do(func() { result = d.Modified() })
		return result
	}
	open := func() bool {
		var result bool
		screen.Do(func() { result = buyDockableFor(sheet) == d })
		return result
	}
	startBuying := func() {
		t.Helper()
		screen.Do(func() { d = showBuyDockable(sheet) })
		c.False(modified(), "an empty purchase must not mark the dockable modified")
		if _, dropped := buyTestDrop(t, screen, sheet, library, dagger); !dropped {
			t.Fatal("the dagger must be dropped into the buy dockable")
		}
		c.True(modified(), "a purchase under way must mark the dockable modified")
	}
	// closeWith posts the close, which runs a modal prompt, answers the prompt with the response and reports whether the
	// prompt offered to discard, as the Yes/No/Cancel prompt does and the plain question does not.
	closeWith := func(attempt func(), response int) (offeredNo bool) {
		t.Helper()
		if !screen.Post(attempt) {
			t.Fatal("the close could not be posted")
		}
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Do(func() {
			offeredNo = dialog.Button(unison.ModalResponseDiscard) != nil
			dialog.StopModal(response)
		})
		return offeredNo
	}

	startBuying()
	c.True(closeWith(func() { d.AttemptClose() }, unison.ModalResponseCancel), "closing must offer to discard")
	c.True(open(), "Cancel must keep the dockable open")
	c.True(closeWith(func() { sheet.AttemptClose() }, unison.ModalResponseCancel),
		"closing the sheet must ask about the purchase")
	c.True(open(), "Cancel must keep the dockable open when the sheet is closing")
	var sheetOpen bool
	screen.Do(func() { sheetOpen = slices.Contains(OpenSheets(nil), sheet) })
	c.True(sheetOpen, "Cancel must keep the sheet open")
	c.True(closeWith(func() { d.AttemptClose() }, unison.ModalResponseDiscard))
	c.False(open(), "No must close the dockable")
	c.Equal(before, buyTestCarried(screen, sheet), "No must discard the purchase")

	startBuying()
	c.True(closeWith(func() { d.AttemptClose() }, unison.ModalResponseOK))
	c.False(open(), "Yes must close the dockable")
	c.Equal([]string{"Silver Eagle ×6", "Copper Penny ×11", "Dagger ×1"}, buyTestCarried(screen, sheet),
		"Yes must finalize the purchase")

	screen.Do(func() {
		sheet.Entity().CarriedEquipment = nil
		sheet.Rebuild(true)
	})
	startBuying()
	c.False(closeWith(func() { d.AttemptClose() }, unison.ModalResponseCancel),
		"a purchase the money can't cover must only be offered for discarding")
	c.True(open(), "Cancel must keep the dockable open")
	c.False(closeWith(func() { d.AttemptClose() }, unison.ModalResponseOK))
	c.False(open(), "OK must discard the purchase and close the dockable")
	c.Nil(buyTestCarried(screen, sheet), "nothing must be bought")
}

// TestBuyDockableWindowHasRoomForItems checks that a window packed around the buy dockable, as when the Editors group
// opens in a window, is big enough to list items rather than closing up around the empty drop area and the footer.
func TestBuyDockableWindowHasRoomForItems(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	var content, hint geom.Size
	var rowHeight float32
	screen.Do(func() {
		d := newBuyDockable(sheet)
		wnd, err := NewWindowForDockable(d, dgroup.Editors)
		if err != nil {
			t.Errorf("unable to open the buy dockable in a window of its own: %v", err)
			return
		}
		defer wnd.Dispose()
		content = wnd.ContentRect().Size
		hint = d.sizeHint()
		rowHeight = unison.DefaultLabelTheme.Font.LineHeight() + unison.StdVSpacing
	})
	c.True(content.Width >= hint.Width && content.Height >= hint.Height,
		"the window must be at least as big as the hint %v, not %v", hint, content)
	c.True(hint.Height > rowHeight*float32(buyHintRows), "the hint must have room for %d rows", buyHintRows)
}

// TestBuyDockableFinalizesOnTheTermsChosen checks that finalizing hands the dialog the automatic payment to start from
// and makes the purchase on the terms it returns: the money chosen pays in place of the automatic choice, with change
// drawn from it, terms that no longer cover the price make no purchase, the merge choice is kept for the next attempt,
// and canceling leaves the dockable as it was.
func TestBuyDockableFinalizesOnTheTermsChosen(t *testing.T) {
	c := check.New(t)
	screen, _, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return buyTestCoins(entity, 7, 11)
	})
	longbow := buyTestItem(nil, "Longbow", "45", 1)
	library := buyTestLibrary(t, screen, "", longbow)
	screen.Do(func() { showBuyDockable(sheet) })
	if _, dropped := buyTestDrop(t, screen, sheet, library, longbow); !dropped {
		t.Fatal("the longbow must be dropped into the buy dockable")
	}
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
	before := []string{"Silver Eagle ×7", "Copper Penny ×11"}

	var next purchaseTerms
	var nextOK, gotMerge bool
	var gotPlan buyPlan
	uxtest.SwapForTest(t, &confirmPurchase, func(d *buyDockable, plan *buyPlan) (purchaseTerms, bool) {
		gotPlan = *plan
		gotMerge = d.merge
		return next, nextOK
	})
	finalize := func(terms purchaseTerms, ok bool) (made, open bool) {
		t.Helper()
		next, nextOK = terms, ok
		screen.Do(func() {
			made = buyDockableFor(sheet).finalizePurchase()
			open = buyDockableFor(sheet) != nil
		})
		return made, open
	}

	made, open := finalize(purchaseTerms{}, false)
	c.False(made, "canceling makes no purchase")
	c.True(open, "canceling leaves the dockable open")
	c.Equal(fxp.FromInteger(45), gotPlan.total, "the dialog is told what is owed")
	c.Equal(map[tid.TID]int{eagleID: 2, pennyID: 5}, paymentQuantities(gotPlan.payments),
		"the dialog starts from the automatic payment")
	c.True(gotMerge, "merging is on to begin with")
	c.Equal(before, buyTestCarried(screen, sheet))

	made, open = finalize(purchaseTerms{chosen: map[tid.TID]int{pennyID: 5}, merge: false}, true)
	c.False(made, "terms that don't cover the price make no purchase")
	c.True(open)
	c.Equal(before, buyTestCarried(screen, sheet))
	finalize(purchaseTerms{}, false)
	c.False(gotMerge, "the merge choice is kept for the next attempt")

	made, open = finalize(purchaseTerms{chosen: map[tid.TID]int{eagleID: 3}, merge: false}, true)
	c.True(made, "the purchase is made on the terms chosen")
	c.False(open, "the dockable closes once the purchase is made")
	c.Equal([]string{"Silver Eagle ×4", "Copper Penny ×11", "Longbow ×1"}, buyTestCarried(screen, sheet),
		"the money chosen is spent in place of the automatic choice")
	var record []*gurps.ShoppingRecordEntry
	screen.Do(func() { record = gurps.CloneShoppingRecordList(sheet.Entity().ShoppingRecord) })
	c.Equal(1, len(record))
	if len(record) == 1 {
		c.Equal([]gurps.ShoppingRecordItem{{Name: "Silver Eagle", Quantity: -fxp.Three, Value: fxp.FromInteger(20)}},
			record[0].Money, "the record shows the money chosen")
		c.Equal(fxp.FromInteger(60), record[0].Exchanged, "change that can't be returned is still exchanged")
	}
}

// TestPurchaseDialog drives the dialog that finalizes a purchase: it lists the money carried, by increasing value, with
// the quantities the automatic payment spends, keeps Finalize disabled while the money chosen is short, notes whether the payment is
// short, exact or due change, with what the change will be, puts the automatic payment back on request, and returns the
// money chosen and whether to merge, or nothing when canceled.
func TestPurchaseDialog(t *testing.T) {
	c := check.New(t)
	screen, wnd, sheet := buyTestSheet(t, c, func(entity *gurps.Entity) []*gurps.Equipment {
		return append(buyTestCoins(entity, 7, 11), buyTestItem(entity, "Rope", "5", 1))
	})
	longbow := buyTestItem(nil, "Longbow", "45", 1)
	library := buyTestLibrary(t, screen, "", longbow)
	var buy *buyDockable
	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = nil
		buy = showBuyDockable(sheet)
	})
	if _, dropped := buyTestDrop(t, screen, sheet, library, longbow); !dropped {
		t.Fatal("the longbow must be dropped into the buy dockable")
	}
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
	// The dialog lists the money by increasing value, so the penny's field comes before the eagle's.
	const penny, eagle = 0, 1

	var terms purchaseTerms
	var ok, returned bool
	open := func() (*unison.Dialog, []*IntegerField) {
		t.Helper()
		returned = false
		c.True(screen.Post(func() {
			terms, ok = confirmPurchaseWithDialog(buy, &buy.plan)
			returned = true
		}))
		screen.Sync()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		var fields []*IntegerField
		var names []string
		screen.Do(func() {
			fields = uxtest.PanelsOfType[*IntegerField](dialogWnd.Content())
			for _, f := range fields {
				names = append(names, f.Accessibility.Name)
			}
		})
		if len(fields) != 2 {
			t.Fatalf("the dialog must have a field for each kind of money carried, not %d", len(fields))
		}
		c.Equal([]string{"Copper Penny to spend", "Silver Eagle to spend"}, names,
			"the money is listed by increasing value")
		return dialog, fields
	}
	controls := func(dialog *unison.Dialog) (automatic *unison.Button, merge *unison.CheckBox) {
		content := dialog.Window().Content()
		for _, b := range uxtest.PanelsOfType[*unison.Button](content) {
			if b.Accessibility.Name == "Choose Automatically" {
				automatic = b
			}
		}
		merge, _ = uxtest.FirstPanelOfType[*unison.CheckBox](content)
		return automatic, merge
	}
	type dialogState struct {
		labels      []string
		okOn        bool
		automaticOn bool
		merge       uncheck.Enum
		spend       []string
		note        string
		noteIcon    *unison.SVG
		noteWarn    bool
	}
	read := func(dialog *unison.Dialog, fields []*IntegerField) dialogState {
		t.Helper()
		var s dialogState
		screen.Do(func() {
			content := dialog.Window().Content()
			s.labels = uxtest.LabelTexts(content)
			s.okOn = dialog.Button(unison.ModalResponseOK).Enabled()
			automatic, merge := controls(dialog)
			if automatic == nil || merge == nil {
				return
			}
			s.automaticOn = automatic.Enabled()
			s.merge = merge.State
			for _, f := range fields {
				s.spend = append(s.spend, f.Text())
			}
			if notes := uxtest.PanelsOfType[*purchaseNote](content); len(notes) == 1 {
				s.note = notes[0].text.Title()
				if icon, isSVG := notes[0].icon.Drawable.(*unison.DrawableSVG); isSVG {
					s.noteIcon = icon.SVG
				}
				s.noteWarn = notes[0].warning
			}
		})
		if s.noteIcon == nil {
			t.Fatal("the dialog must have a note with an icon, a Choose Automatically button and a merge checkbox")
		}
		return s
	}
	click := func(dialog *unison.Dialog, code int) {
		t.Helper()
		screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, code)))
		if !returned {
			t.Fatal("the dialog must close")
		}
	}
	contains := func(state dialogState, wants ...string) {
		t.Helper()
		for _, want := range wants {
			c.True(slices.Contains(state.labels, want), "the dialog must show %q; it shows %v", want, state.labels)
		}
	}

	dialog, fields := open()
	uxtest.CaptureScreen(t, c, screen, "purchase_dialog")
	state := read(dialog, fields)
	c.Equal([]string{"5", "2"}, state.spend, "the dialog starts from the automatic payment")
	contains(state, "Money Item", "Carried", "Value Each", "Spend", "Silver Eagle", "7", "$20", "Copper Penny", "11",
		"$1", "Buying Longbow ×1 for $45.",
		"The money-tagged items carried pay for it, chosen automatically unless you change the quantities to spend.",
		"Owed: $45", "Spending: $45", "Change due: $0")
	c.True(state.okOn)
	c.False(state.automaticOn, "the automatic payment is what is chosen to begin with")
	c.Equal(uncheck.On, state.merge, "merging is on to begin with")
	c.Equal("Covered exactly. Money you do not spend stays on the sheet.", state.note)
	c.True(state.noteIcon == unison.CheckmarkSVG, "an exact payment shows a checkmark")
	c.False(state.noteWarn)

	screen.Do(func() { fields[penny].SetText("11") })
	state = read(dialog, fields)
	contains(state, "Spending: $51", "Change due: $6")
	c.Equal("Change will be returned as Copper Penny ×6.", state.note,
		"without sell money, the change comes from the money paid")
	c.True(state.noteIcon == svg.Info, "change that will be returned is noted")
	c.False(state.noteWarn)
	c.True(state.automaticOn, "a change to the money chosen offers the automatic payment back")
	var typed, refused bool
	screen.Do(func() {
		fields[penny].SetText("1")
		fields[penny].SetSelectionToEnd()
		refused = !fields[penny].RuneTypedCallback('2')
		typed = fields[penny].RuneTypedCallback('1')
	})
	c.True(refused, "more of a money item than is carried can't be typed")
	c.True(typed)
	c.Equal("11", read(dialog, fields).spend[penny])
	screen.Do(func() { fields[penny].SetText("0") })
	state = read(dialog, fields)
	contains(state, "Spending: $40", "Change due: $0")
	uxtest.CaptureScreen(t, c, screen, "purchase_short")
	c.Equal("Short by $5.", state.note)
	c.True(state.noteIcon == svg.Not, "a shortfall shows the not symbol")
	c.True(state.noteWarn)
	c.False(state.okOn, "Finalize is disabled while the money chosen is short")
	var focused bool
	screen.Do(func() {
		focused = dialog.Window().CurrentFocus() == fields[penny].AsPanel()
		automatic, merge := controls(dialog)
		automatic.Click()
		merge.Click()
	})
	c.True(focused, "the first field holds the focus, so the button must replace the text of a focused field")
	state = read(dialog, fields)
	c.Equal([]string{"5", "2"}, state.spend, "Choose Automatically puts the automatic payment back")
	c.False(state.automaticOn)
	c.True(state.okOn)
	c.Equal(uncheck.Off, state.merge)
	click(dialog, unison.ModalResponseOK)
	c.True(ok)
	c.Equal(map[tid.TID]int{eagleID: 2, pennyID: 5}, terms.chosen)
	c.False(terms.merge, "the merge choice is returned")

	// The note says what the change will be, or how much of it neither the sell money nor the money paid can return.
	payWithEagles := func(count string) (*unison.Dialog, []*IntegerField) {
		t.Helper()
		dialog, fields = open()
		screen.Do(func() {
			fields[eagle].SetText(count)
			fields[penny].SetText("0")
		})
		return dialog, fields
	}
	dialog, fields = payWithEagles("3")
	state = read(dialog, fields)
	uxtest.CaptureScreen(t, c, screen, "purchase_unreturned")
	c.Equal("$15 of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		state.note)
	c.True(state.noteIcon == unison.TriangleExclamationSVG, "change that can't be returned is warned about")
	c.True(state.noteWarn)
	c.True(state.okOn, "an overpayment can still be made")
	click(dialog, unison.ModalResponseCancel)
	c.False(ok, "canceling chooses nothing")
	sources := buyTestMoneyLibrary(t, c)
	missing := gurps.Source{LibraryFile: sources[0].LibraryFile, TID: "missing"}
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = []gurps.Source{missing} })
	dialog, fields = payWithEagles("3")
	state = read(dialog, fields)
	c.Equal("$15 of change can't be returned from the money paid, and none of the sheet's sell money could be found in the libraries.",
		state.note, "sell money that can't be found is told apart from none at all")
	c.True(state.noteWarn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sources[1:] })
	dialog, fields = payWithEagles("3")
	state = read(dialog, fields)
	uxtest.CaptureScreen(t, c, screen, "purchase_change")
	c.Equal("Change will be returned as Copper Penny ×15.", state.note, "the sell money returns the change")
	c.True(state.noteIcon == svg.Info)
	c.False(state.noteWarn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sources[:1] })
	dialog, fields = payWithEagles("4")
	state = read(dialog, fields)
	c.Equal("Change will be returned as Silver Eagle ×1 ($20). $15 of change can't be returned with the sheet's sell money or the money paid.",
		state.note)
	c.True(state.noteIcon == unison.TriangleExclamationSVG)
	c.True(state.noteWarn)
	screen.Do(func() { fields[penny].SetText("10") })
	state = read(dialog, fields)
	c.Equal("Change will be returned as Silver Eagle ×2, Copper Penny ×5.", state.note,
		"the money paid is drawn on when the sell money falls short")
	c.True(state.noteIcon == svg.Info)
	c.False(state.noteWarn)
	c.True(state.okOn)
	click(dialog, unison.ModalResponseCancel)
}
