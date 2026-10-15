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
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
)

// tradeTestMoney is an equipment library file, in the user library, of the money the trading tests pay out in: a Silver
// Eagle worth $20 and a Copper Penny worth $1, both tagged as money, along with a Purse, which is a container.
type tradeTestMoney struct {
	path    string
	libFile gurps.LibraryFile
	eagle   *gurps.Equipment
	penny   *gurps.Equipment
	purse   *gurps.Equipment
}

// newTradeTestMoney writes the money library file into the user library the global settings hold, which the test must
// already have replaced (see uxtest.UseTestLibraries).
func newTradeTestMoney(t *testing.T, c check.Checker) *tradeTestMoney {
	t.Helper()
	user := gurps.GlobalSettings().Libraries.User()
	name := "Money" + gurps.EquipmentExt
	m := &tradeTestMoney{
		path:    filepath.Join(user.Path(false), name),
		libFile: gurps.LibraryFile{Library: user.Key(), Path: name},
		eagle:   newTradeTestItem(nil, "Silver Eagle", "20", fxp.One, gurps.MoneyTag),
		penny:   newTradeTestItem(nil, "Copper Penny", "1", fxp.One, gurps.MoneyTag),
		purse:   gurps.NewEquipment(nil, nil, true),
	}
	m.purse.Name = "Purse"
	c.NoError(os.MkdirAll(filepath.Dir(m.path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{m.eagle, m.penny, m.purse}, m.path))
	return m
}

// source returns the source of the item in the money library file.
func (m *tradeTestMoney) source(eqp *gurps.Equipment) gurps.Source {
	return gurps.Source{LibraryFile: m.libFile, TID: eqp.TID}
}

// newTradeTestItem returns a piece of equipment owned by owner with the given name, value, quantity and tags.
func newTradeTestItem(owner gurps.DataOwner, name, value string, quantity fxp.Int, tags ...string) *gurps.Equipment {
	eqp := gurps.NewEquipment(owner, nil, false)
	eqp.Name = name
	eqp.BaseValue = value
	eqp.Quantity = quantity
	eqp.Tags = tags
	return eqp
}

// equipmentRowsFor returns the table's rows holding the equipment, in the order given, failing the test if any is not
// among the table's top-level rows.
func equipmentRowsFor(t *testing.T, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) []*Node[*gurps.Equipment] {
	t.Helper()
	rows := make([]*Node[*gurps.Equipment], 0, len(items))
	for _, eqp := range items {
		found := false
		for _, row := range table.RootRows() {
			if row.Data() == eqp {
				rows = append(rows, row)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("the table has no row for %s", eqp.Name)
		}
	}
	return rows
}

// dropOnSellMoney drags the rows of the equipment from the table over the dockable's sell money list and drops them,
// checking that the list offers to take the drag exactly when the drop adds something, and returning what the drop
// reported.
func dropOnSellMoney(t *testing.T, c check.Checker, d *sheetSettingsDockable, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) bool {
	t.Helper()
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.Equipment]]{
		Table: table,
		Rows:  equipmentRowsFor(t, table, items...),
	}))
	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	p := d.sellMoneyPanel
	c.True(p.CanAcceptDropCallback(di), "the sell money list must accept equipment")
	op := p.DragEnteredCallback(di, geom.Point{}, 0)
	c.Equal(op, p.DragUpdatedCallback(di, geom.Point{}, 0), "moving within offers what entering did")
	c.Equal(op == drag.Copy, d.sellMoneyDropping, "the list is highlighted only while a drag it can take is over it")
	dropped := p.DropCallback(di, geom.Point{}, 0)
	c.Equal(dropped, op == drag.Copy, "a copy is offered exactly when the drop adds something")
	c.False(d.sellMoneyDropping, "the highlight goes once the drop is made")
	return dropped
}

// sellMoneyRemoveButton returns the button that removes the named sell money item, failing the test if there is none.
func sellMoneyRemoveButton(t *testing.T, d *sheetSettingsDockable, name string) *unison.Button {
	t.Helper()
	for _, b := range uxtest.PanelsOfType[*unison.Button](d.sellMoneyPanel) {
		if b.Accessibility.Name == "Remove "+name {
			return b
		}
	}
	t.Fatalf("the sell money list has no button to remove %s", name)
	return nil
}

// newTradingSettingsTest replaces the libraries with a set holding the money library file, which it returns along with
// a library dockable showing that file, and keeps drag feedback from trying to draw.
func newTradingSettingsTest(t *testing.T, c check.Checker) (*tradeTestMoney, *TableDockable[*gurps.Equipment]) {
	t.Helper()
	RegisterKnownFileTypes()
	uxtest.UseTestLibraries(t, c)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	money := newTradeTestMoney(t, c)
	return money, NewEquipmentTableDockable(money.path, []*gurps.Equipment{money.eagle, money.penny, money.purse})
}

// TestSheetSettingsTradingPercentFields checks that the buy and sell percent fields show the settings, write their own
// setting when edited and tell the open sheets of it, and that sync() shows replaced settings without telling the
// sheets anything, since the change that replaced them does that.
func TestSheetSettingsTradingPercentFields(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	sellBefore := s.SellPercent
	c.Equal(s.BuyPercent, d.buyPercentField.CurrentValue(), "the buy percent field shows the setting")
	c.Equal(s.SellPercent, d.sellPercentField.CurrentValue(), "the sell percent field shows the setting")

	d.buyPercentField.SetText("75")
	c.Equal(75, s.BuyPercent, "the buy percent field writes the buy percent")
	c.Equal(sellBefore, s.SellPercent, "the buy percent field leaves the sell percent alone")
	c.Equal([]bool{false}, recorder.updates, "a new buy percent is reported, without a full rebuild")

	d.sellPercentField.SetText("40")
	c.Equal(40, s.SellPercent, "the sell percent field writes the sell percent")
	c.Equal(75, s.BuyPercent, "the sell percent field leaves the buy percent alone")
	c.Equal([]bool{false, false}, recorder.updates, "a new sell percent is reported, without a full rebuild")

	replacement := s.Clone(owner.entity)
	replacement.BuyPercent = 120
	replacement.SellPercent = 30
	owner.entity.SheetSettings = replacement
	replacement.SetOwningEntity(owner.entity)
	d.sync()
	c.Equal(120, d.buyPercentField.CurrentValue(), "sync shows the replacement's buy percent")
	c.Equal(30, d.sellPercentField.CurrentValue(), "sync shows the replacement's sell percent")
	c.Equal(120, replacement.BuyPercent, "sync leaves the buy percent as it was")
	c.Equal(30, replacement.SellPercent, "sync leaves the sell percent as it was")
	c.Equal(2, len(recorder.updates), "sync tells the sheets nothing")
}

// TestSheetSettingsSellMoneyDrop checks that dropping library equipment onto the sell money list adds a source for
// each item that can be money, once, and tells the open sheets, while containers, equipment that isn't tagged as money
// and equipment from outside the libraries are refused, with a drag of nothing but those declined outright; that each
// entry's button removes it; that an entry the library no longer has is still shown; and that sync() shows the list of
// replaced settings.
func TestSheetSettingsSellMoneyDrop(t *testing.T) {
	c := check.New(t)
	money, library := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	hint := []string{"Drag equipment tagged \"Money\" from an equipment library here."}
	c.Equal(hint, uxtest.LabelTexts(d.sellMoneyPanel), "an empty list explains how to add to it")

	other := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	c.False(d.sellMoneyPanel.CanAcceptDropCallback(other), "the list takes nothing but equipment")
	c.Equal(drag.None, d.sellMoneyPanel.DragUpdatedCallback(other, geom.Point{}, 0), "nor offers to")

	bead := newTradeTestItem(nil, "Trade Bead", "2", fxp.One)
	beadPath := filepath.Join(filepath.Dir(money.path), "Beads"+gurps.EquipmentExt)
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{bead}, beadPath))
	beads := NewEquipmentTableDockable(beadPath, []*gurps.Equipment{bead})
	c.False(dropOnSellMoney(t, c, d, beads.table, bead), "equipment that isn't tagged as money is refused")
	c.Equal(0, len(s.SellMoney))

	c.True(dropOnSellMoney(t, c, d, library.table, money.eagle), "a money item from a library is added")
	c.Equal([]gurps.Source{money.source(money.eagle)}, s.SellMoney, "the item is kept by its library source")
	c.Equal([]bool{false}, recorder.updates, "the addition is reported, without a full rebuild")
	c.Equal([]string{"Silver Eagle"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item")

	c.False(dropOnSellMoney(t, c, d, library.table, money.eagle), "an item already listed is not added again")
	c.False(dropOnSellMoney(t, c, d, library.table, money.purse), "a container is refused")
	c.Equal(1, len(s.SellMoney), "neither changes the list")
	c.Equal(1, len(recorder.updates), "nor is reported")

	c.True(dropOnSellMoney(t, c, d, library.table, money.purse, money.penny),
		"of several rows, those that can be money are added")
	c.Equal([]gurps.Source{money.source(money.eagle), money.source(money.penny)}, s.SellMoney,
		"the money item is added after the one already there")
	c.Equal([]string{"Silver Eagle", "Copper Penny"}, uxtest.LabelTexts(d.sellMoneyPanel))

	elsewhere := newTradeTestItem(nil, "Gold Crown", "100", fxp.One, gurps.MoneyTag)
	outside := NewEquipmentTableDockable(filepath.Join(t.TempDir(), "Coins"+gurps.EquipmentExt),
		[]*gurps.Equipment{elsewhere})
	c.False(dropOnSellMoney(t, c, d, outside.table, elsewhere), "equipment from outside the libraries is refused")
	c.Equal(2, len(s.SellMoney))

	sellMoneyRemoveButton(t, d, "Silver Eagle").ClickCallback()
	c.Equal([]gurps.Source{money.source(money.penny)}, s.SellMoney, "the button removes its own entry")
	c.Equal(3, len(recorder.updates), "the removal is reported")
	c.Equal([]string{"Copper Penny"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list no longer shows it")

	missing := gurps.Source{LibraryFile: money.libFile, TID: "missing"}
	s.SellMoney = append(s.SellMoney, missing)
	d.rebuildSellMoney()
	gone := "Money" + gurps.EquipmentExt + " (missing)"
	c.Equal([]string{"Copper Penny", gone + " — not found"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"an entry the library lacks is shown by its file and ID")
	sellMoneyRemoveButton(t, d, gone).ClickCallback()
	c.Equal([]gurps.Source{money.source(money.penny)}, s.SellMoney, "an entry the library lacks can be removed")

	d.reset()
	c.Equal(0, len(owner.entity.SheetSettings.SellMoney), "reset restores the defaults, which have no sell money")
	c.Equal(hint, uxtest.LabelTexts(d.sellMoneyPanel), "sync shows the list of the reset settings")
}

// TestSheetSettingsSellMoneyWithoutOwner checks that the defaults' sell money list, which no sheet owns, takes drops
// and removals too.
func TestSheetSettingsSellMoneyWithoutOwner(t *testing.T) {
	c := check.New(t)
	money, library := newTradingSettingsTest(t, c)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))
	d, recorder := newTestSheetSettingsDockable(t, nil)

	c.True(dropOnSellMoney(t, c, d, library.table, money.penny), "a money item from a library is added")
	c.Equal([]gurps.Source{money.source(money.penny)}, global.Sheet.SellMoney, "the defaults hold its source")
	c.Equal([]*gurps.Entity{nil}, recorder.entities, "the change is reported as one to the defaults")
	c.Equal([]string{"Copper Penny"}, uxtest.LabelTexts(d.sellMoneyPanel))

	sellMoneyRemoveButton(t, d, "Copper Penny").ClickCallback()
	c.Equal(0, len(global.Sheet.SellMoney), "the button removes the entry from the defaults")
	c.Equal(2, len(recorder.updates), "the removal is reported")
}

// TestSheetSettingsSellMoneyDropKeepsTheItemDragged checks that an item dragged from a library that was itself copied
// into that library from another, and so carries that other library's source, is added as the item dragged rather than
// as the item it was copied from.
func TestSheetSettingsSellMoneyDropKeepsTheItemDragged(t *testing.T) {
	c := check.New(t)
	money, _ := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, _ := newTestSheetSettingsDockable(t, owner)
	user := gurps.GlobalSettings().Libraries.User()
	name := "Local Coins" + gurps.EquipmentExt
	path := filepath.Join(user.Path(false), name)
	local := money.eagle.Clone(money.libFile, nil, nil, gurps.Reference)
	local.Name = "Local Eagle"
	local.BaseValue = "25"
	c.Equal(money.source(money.eagle), local.Source, "the copy keeps the source it was copied from")
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{local}, path))
	library := NewEquipmentTableDockable(path, []*gurps.Equipment{local})

	c.True(dropOnSellMoney(t, c, d, library.table, local), "the copied item is added")
	c.Equal([]gurps.Source{{Library: user.Key(), Path: name, TID: local.TID}},
		owner.entity.SheetSettings.SellMoney, "the item dragged is what is kept, not the item it was copied from")
	c.Equal([]string{"Local Eagle"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item dragged")
}

// TestSheetSettingsSellMoneyRemoveKeepsTheFocus checks that removing a sell money entry with its button while the button
// holds the keyboard focus moves the focus to the button that takes its place, or to the sell percent field once no
// entry is left, rather than leaving it on the button the rebuilt list no longer holds.
func TestSheetSettingsSellMoneyRemoveKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.eagle), money.source(money.penny)},
		func(_ *gurps.Entity) []*gurps.Equipment { return nil })
	var d *sheetSettingsDockable
	screen.Do(func() {
		ShowSheetSettings(sheet)
		for _, one := range AllDockables() {
			if s, ok := one.AsPanel().Self.(*sheetSettingsDockable); ok && s.owner == sheet {
				d = s
			}
		}
	})
	if d == nil {
		t.Fatal("the sheet's settings must be open")
	}
	removeButton := func(name string) *unison.Button {
		var button *unison.Button
		screen.Do(func() {
			for _, b := range uxtest.PanelsOfType[*unison.Button](d.sellMoneyPanel) {
				if b.Accessibility.Name == "Remove "+name {
					button = b
				}
			}
		})
		return button
	}
	press := func(name string) {
		t.Helper()
		button := removeButton(name)
		if button == nil {
			t.Fatalf("the sell money list has no button to remove %s", name)
		}
		screen.Do(func() {
			button.RequestFocus()
			button.Click()
		})
	}

	press("Silver Eagle")
	var focus *unison.Panel
	screen.Do(func() { focus = wnd.CurrentFocus() })
	c.Equal([]string{"Copper Penny"}, uxtest.LabelTexts(d.sellMoneyPanel), "the entry is removed")
	c.True(focus != nil && focus == removeButton("Copper Penny").AsPanel(),
		"the focus must move to the button that takes the removed one's place")

	press("Copper Penny")
	screen.Do(func() { focus = wnd.CurrentFocus() })
	c.True(focus == d.sellPercentField.AsPanel(), "the focus must move to the sell percent field once no entry is left")
}
