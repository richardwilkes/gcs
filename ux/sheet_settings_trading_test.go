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
	"image/color"
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
	uncheck "github.com/richardwilkes/unison/enums/check"
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

// dropOnSellMoney drags the rows of the equipment from the table over the dockable's sell money list and drops them
// (see dropRowsOnSellMoney), returning what the drop reported. It must run on the test goroutine, since it fails the
// test outright when the table lacks a row for an item.
func dropOnSellMoney(t *testing.T, c check.Checker, d *sheetSettingsDockable, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) bool {
	t.Helper()
	return dropRowsOnSellMoney(t, c, d, table, equipmentRowsFor(t, table, items...))
}

// dropRowsOnSellMoney drags the table's rows over the dockable's sell money list and drops them, checking that the list
// offers to take the drag exactly when the drop adds something, and returning what the drop reported. It may run on
// the UI goroutine of a headless workspace, since it only ever reports failures rather than ending the test.
func dropRowsOnSellMoney(t *testing.T, c check.Checker, d *sheetSettingsDockable, table *unison.Table[*Node[*gurps.Equipment]], rows []*Node[*gurps.Equipment]) bool {
	t.Helper()
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.Equipment]]{Table: table, Rows: rows}))
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

// sellMoneyNames returns the names of the settings' sell money, in the order held.
func sellMoneyNames(s *gurps.SheetSettings) []string {
	names := make([]string, 0, len(s.SellMoney))
	for _, eqp := range s.SellMoney {
		names = append(names, eqp.Name)
	}
	return names
}

// TestSheetSettingsSellMoneyDrop checks that dropping equipment onto the sell money list adds a copy of each item that
// can be money, once, wherever it comes from, and tells the open sheets, while containers and equipment that isn't
// tagged as money are refused, with a drag of nothing but those declined outright; that each entry's button removes it;
// and that sync() shows the list of replaced settings.
func TestSheetSettingsSellMoneyDrop(t *testing.T) {
	c := check.New(t)
	money, library := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	hint := []string{"Drag equipment tagged \"Money\" from any equipment list here."}
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
	c.Equal([]string{"Silver Eagle"}, sellMoneyNames(s))
	c.Equal([]bool{false}, recorder.updates, "the addition is reported, without a full rebuild")
	c.Equal([]string{"Silver Eagle", "$20"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item and its value")
	if len(s.SellMoney) == 1 {
		eagle := s.SellMoney[0]
		c.True(eagle != money.eagle, "a copy is kept, not the library's item")
		c.NotEqual(money.eagle.TID, eagle.TID)
		c.Equal(money.source(money.eagle), eagle.Source, "the copy points back at the library's item")
		c.Nil(eagle.DataOwner(), "the copy has no owner")
	}

	c.False(dropOnSellMoney(t, c, d, library.table, money.eagle), "an item already listed is not added again")
	c.False(dropOnSellMoney(t, c, d, library.table, money.purse), "a container is refused")
	c.Equal(1, len(s.SellMoney), "neither changes the list")
	c.Equal(1, len(recorder.updates), "nor is reported")

	c.True(dropOnSellMoney(t, c, d, library.table, money.purse, money.penny),
		"of several rows, those that can be money are added")
	c.Equal([]string{"Silver Eagle", "Copper Penny"}, sellMoneyNames(s),
		"the money item is added after the one already there")
	c.Equal([]string{"Copper Penny", "$1", "Silver Eagle", "$20"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list shows the items by increasing value")

	elsewhere := newTradeTestItem(nil, "Gold Crown", "100", fxp.One, gurps.MoneyTag)
	outside := NewEquipmentTableDockable(filepath.Join(t.TempDir(), "Coins"+gurps.EquipmentExt),
		[]*gurps.Equipment{elsewhere})
	c.True(dropOnSellMoney(t, c, d, outside.table, elsewhere), "equipment from outside the libraries is added too")
	c.Equal([]string{"Silver Eagle", "Copper Penny", "Gold Crown"}, sellMoneyNames(s))
	if len(s.SellMoney) == 3 {
		c.True(s.SellMoney[2].Source.IsZero(), "a copy of an item from no library, with no Source of its own, has none")
	}
	c.Equal([]string{"Copper Penny", "$1", "Silver Eagle", "$20", "Gold Crown", "$100"},
		uxtest.LabelTexts(d.sellMoneyPanel))

	sellMoneyRemoveButton(t, d, "Silver Eagle").ClickCallback()
	c.Equal([]string{"Copper Penny", "Gold Crown"}, sellMoneyNames(s), "the button removes its own entry")
	c.Equal(4, len(recorder.updates), "the removal is reported")
	c.Equal([]string{"Copper Penny", "$1", "Gold Crown", "$100"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list no longer shows it")

	d.reset()
	c.Equal(0, len(owner.entity.SheetSettings.SellMoney), "reset restores the defaults, which have no sell money")
	c.Equal(hint, uxtest.LabelTexts(d.sellMoneyPanel), "sync shows the list of the reset settings")
}

// TestSheetSettingsSellMoneyFollowsTheValueFormat checks that the sell money list shows the values in the equipment
// value format of the settings and follows a change to that format made in the same dockable.
func TestSheetSettingsSellMoneyFollowsTheValueFormat(t *testing.T) {
	c := check.New(t)
	money, _ := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, _ := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	s.EquipmentValueFormat = fxp.NumberFormat{}
	halfCrown := newTradeTestItem(nil, "Half Crown", "2.5", fxp.One, gurps.MoneyTag)
	path := filepath.Join(filepath.Dir(money.path), "Crowns"+gurps.EquipmentExt)
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{halfCrown}, path))
	crowns := NewEquipmentTableDockable(path, []*gurps.Equipment{halfCrown})
	c.True(dropOnSellMoney(t, c, d, crowns.table, halfCrown))
	c.Equal([]string{"Half Crown", "$2.5"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the value is shown with as many decimal places as it has")

	row := d.numberFormats[3]
	row.popup.Select(fxp.ZeroPlaces)
	c.Equal(fxp.ZeroPlaces, s.EquipmentValueFormat.Places, "the equipment value popup must be the fourth row")
	c.Equal([]string{"Half Crown", "$3"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list follows a change to the decimal places")

	row.popup.Select(fxp.TwoPlaces)
	row.pad.State = uncheck.On
	row.pad.ClickCallback()
	c.Equal([]string{"Half Crown", "$2.50"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list follows a change to the padding")
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
	c.Equal([]string{"Copper Penny"}, sellMoneyNames(global.Sheet), "the defaults hold a copy of it")
	c.Equal([]*gurps.Entity{nil}, recorder.entities, "the change is reported as one to the defaults")
	c.Equal([]string{"Copper Penny", "$1"}, uxtest.LabelTexts(d.sellMoneyPanel))

	sellMoneyRemoveButton(t, d, "Copper Penny").ClickCallback()
	c.Equal(0, len(global.Sheet.SellMoney), "the button removes the entry from the defaults")
	c.Equal(2, len(recorder.updates), "the removal is reported")
}

// TestSheetSettingsSellMoneyDropCopiesTheItemDragged checks that an item dragged from a library that was itself copied
// into that library from another, and so carries that other library's source, is added as the item dragged, with its
// own name and value, keeping the source it carries as any copy of it does (see gurps.Reference).
func TestSheetSettingsSellMoneyDropCopiesTheItemDragged(t *testing.T) {
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
	s := owner.entity.SheetSettings
	c.Equal([]string{"Local Eagle"}, sellMoneyNames(s), "the item dragged is what is copied, not the item it was copied from")
	if len(s.SellMoney) == 1 {
		c.Equal("25", s.SellMoney[0].BaseValue)
		c.NotEqual(local.TID, s.SellMoney[0].TID)
		c.Equal(money.source(money.eagle), s.SellMoney[0].Source, "the copy keeps the source the item carries")
	}
	c.Equal([]string{"Local Eagle", "$25"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item dragged")
}

// TestSheetSettingsSellMoneyDropFromASheet checks that equipment dragged from a sheet's carried equipment list, which
// no library holds, is added as a copy owned by no one with a quantity of one, leaving the item carried as it was.
func TestSheetSettingsSellMoneyDropFromASheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	var bar *gurps.Equipment
	sheet := openSellTestSheet(t, screen, nil, func(entity *gurps.Entity) []*gurps.Equipment {
		bar = newTradeTestItem(entity, "Trade Bar", "50", fxp.Five, gurps.MoneyTag)
		return []*gurps.Equipment{bar}
	})
	d := openSheetSettingsFor(t, screen, sheet)
	// The row is looked up here and the test failed on this goroutine, as testing requires, rather than by
	// dropOnSellMoney on the UI goroutine, which would leave the headless screen without its event loop.
	var rows []*Node[*gurps.Equipment]
	screen.Do(func() {
		for _, row := range sheet.CarriedEquipment.Table.RootRows() {
			if row.Data() == bar {
				rows = append(rows, row)
			}
		}
	})
	if len(rows) != 1 {
		t.Fatal("the sheet's carried equipment list must have a row for the bar")
	}
	var dropped bool
	var names, texts []string
	var entry *gurps.Equipment
	screen.Do(func() {
		dropped = dropRowsOnSellMoney(t, c, d, sheet.CarriedEquipment.Table, rows)
		s := sheet.Entity().SheetSettings
		names = sellMoneyNames(s)
		if len(s.SellMoney) == 1 {
			entry = s.SellMoney[0]
		}
		texts = uxtest.LabelTexts(d.sellMoneyPanel)
	})
	c.True(dropped, "money carried on a sheet can be dropped on its own settings")
	c.Equal([]string{"Trade Bar"}, names)
	if entry != nil {
		c.True(entry != bar, "a copy is kept, not the carried item")
		c.NotEqual(bar.TID, entry.TID)
		c.Equal(fxp.One, entry.Quantity, "the copy holds one, however many are carried")
		c.Nil(entry.DataOwner(), "the copy has no owner")
		c.True(entry.Source.IsZero(), "a copy of an item from no library, with no Source of its own, has none")
	}
	c.Equal(fxp.Five, bar.Quantity, "the carried item is untouched")
	c.Equal([]string{"Trade Bar", "$50.00"}, texts, "the list shows the copy, formatted as the sheet formats values")
}

// openSheetSettingsFor opens the sheet's settings in the headless workspace and returns their dockable, failing the test
// if it can't be found.
func openSheetSettingsFor(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) *sheetSettingsDockable {
	t.Helper()
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
	return d
}

// TestSheetSettingsSellMoneyBanding checks that the sell money list of a sheet whose settings show two decimal places
// shows the values so, by increasing value, and that it draws the rows on the background a table gives its rows, with
// every other row banded.
func TestSheetSettingsSellMoneyBanding(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	sheet := openSellTestSheet(t, screen, []*gurps.Equipment{money.eagle, money.penny},
		func(_ *gurps.Entity) []*gurps.Equipment { return nil })
	d := openSheetSettingsFor(t, screen, sheet)
	var texts []string
	var rows int
	var samples []color.NRGBA
	screen.Do(func() {
		p := d.sellMoneyPanel
		texts = uxtest.LabelTexts(p)
		rows = len(d.sellMoneyRows)
		d.ValidateLayout()
		area := p.ContentRect(true)
		img, err := unison.NewImageFromDrawing(int(area.Width), int(area.Height), 72, func(gc *unison.Canvas) {
			p.DrawCallback(gc, area)
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
		for _, row := range d.sellMoneyRows {
			center := sellMoneyRowBounds(row, area).Center()
			samples = append(samples, pixels.NRGBAAt(int(center.X), int(center.Y)))
		}
	})
	c.Equal([]string{"Copper Penny", "$1.00", "Silver Eagle", "$20.00"}, texts,
		"the list shows the items by increasing value, formatted as the sheet formats equipment values")
	c.Equal(2, rows, "the list keeps the widgets of each row")
	nrgba := func(ink unison.ColorProvider) color.NRGBA {
		clr := ink.GetColor()
		return color.NRGBA{R: uint8(clr.Red()), G: uint8(clr.Green()), B: uint8(clr.Blue()), A: 255}
	}
	c.Equal([]color.NRGBA{nrgba(unison.ThemeBelowSurface), nrgba(unison.ThemeBanding)}, samples,
		"the first row sits on the background a table gives its rows and the second is banded")
}

// TestSheetSettingsSellMoneyRemoveKeepsTheFocus checks that removing a sell money entry with its button while the button
// holds the keyboard focus moves the focus to the button that takes its place, or to the sell percent field once no
// entry is left, rather than leaving it on the button the rebuilt list no longer holds.
func TestSheetSettingsSellMoneyRemoveKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	sheet := openSellTestSheet(t, screen, []*gurps.Equipment{money.eagle, money.penny},
		func(_ *gurps.Entity) []*gurps.Equipment { return nil })
	d := openSheetSettingsFor(t, screen, sheet)
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
	c.Equal([]string{"Copper Penny", "$1.00"}, uxtest.LabelTexts(d.sellMoneyPanel), "the entry is removed")
	c.True(focus != nil && focus == removeButton("Copper Penny").AsPanel(),
		"the focus must move to the button that takes the removed one's place")

	press("Copper Penny")
	screen.Do(func() { focus = wnd.CurrentFocus() })
	c.True(focus == d.sellPercentField.AsPanel(), "the focus must move to the sell percent field once no entry is left")
}
