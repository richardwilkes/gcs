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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
)

// openSellTestSheet opens a new character sheet carrying the given equipment, which is made with the entity as its
// owner, showing equipment values to two decimal places, padded, and paying sales out in the sell money given.
func openSellTestSheet(t *testing.T, screen *unison.HeadlessScreen, sellMoney []gurps.Source, carried func(entity *gurps.Entity) []*gurps.Equipment) *Sheet {
	t.Helper()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		entity.CarriedEquipment = carried(entity)
		entity.SheetSettings.SellMoney = sellMoney
		entity.SheetSettings.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.undoMgr.Clear()
	})
	return sheet
}

// carriedSummary describes the sheet's top-level carried equipment as the name and quantity of each item, in order.
func carriedSummary(sheet *Sheet) []string {
	list := make([]string, 0, len(sheet.Entity().CarriedEquipment))
	for _, eqp := range sheet.Entity().CarriedEquipment {
		list = append(list, eqp.Name+" ×"+eqp.Quantity.String())
	}
	return list
}

// sellTestDialog is the sell dialog a test has opened, with the controls it drives.
type sellTestDialog struct {
	t          *testing.T
	screen     *unison.HeadlessScreen
	wnd        *unison.Window
	dialog     *unison.Dialog
	percent    *DecimalField
	quantities []*DecimalField
	name       *StringField
	merge      *unison.CheckBox
	sell       *unison.Button
	cancel     *unison.Button
	done       *bool
}

// openSellTestDialog selects the equipment in the sheet's carried equipment list and opens the sell dialog for it as
// the list's Sell command does. The dialog runs a nested modal loop, so it is opened from a posted task, which is done
// once the dialog has been dismissed and its sale, if any, made. The dialog must have a quantity field for each item
// selected that can be sold (see selectedSaleRows).
func openSellTestDialog(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, sheet *Sheet, items ...*gurps.Equipment) *sellTestDialog {
	t.Helper()
	done := false
	s := &sellTestDialog{t: t, screen: screen, done: &done}
	var table *unison.Table[*Node[*gurps.Equipment]]
	var sellable int
	screen.Do(func() {
		table = sheet.CarriedEquipment.Table
		selection := make(map[tid.TID]bool, len(items))
		for _, eqp := range items {
			selection[eqp.TID] = true
		}
		table.SetSelectionMap(selection)
		sellable = len(selectedSaleRows(table))
	})
	if !screen.Post(func() {
		sellSelectedEquipment(sheet, table)
		done = true
	}) {
		t.Fatal("the sell command could not be posted")
	}
	screen.Sync()
	s.wnd, s.dialog = modalDialog(t, screen, wnd)
	s.sell = dialogButton(t, screen, s.dialog, unison.ModalResponseOK)
	s.cancel = dialogButton(t, screen, s.dialog, unison.ModalResponseCancel)
	screen.Do(func() {
		content := s.wnd.Content()
		if fields := uxtest.PanelsOfType[*DecimalField](content); len(fields) != 0 {
			s.percent = fields[0]
			s.quantities = fields[1:]
		}
		s.name, _ = uxtest.FirstPanelOfType[*StringField](content)
		s.merge, _ = uxtest.FirstPanelOfType[*unison.CheckBox](content)
	})
	if s.percent == nil || s.name == nil || s.merge == nil {
		t.Fatal("the sell dialog must have a percent field, a transaction name field and a merge checkbox")
	}
	if len(s.quantities) != sellable {
		t.Fatalf("the sell dialog must have a quantity field for each of the %d items that can be sold; it has %d",
			sellable, len(s.quantities))
	}
	return s
}

// texts returns the text of every label in the dialog, in order.
func (s *sellTestDialog) texts() []string {
	var texts []string
	s.screen.Do(func() { texts = uxtest.LabelTexts(s.wnd.Content()) })
	return texts
}

// warning returns the text of the dialog's warning, which is empty while there is nothing to warn of. The warning is
// then not in the dialog at all; when it is, it follows the payout, the first of the dialog's wrapping labels.
func (s *sellTestDialog) warning() string {
	var text string
	s.screen.Do(func() {
		if labels := uxtest.PanelsOfType[*TextLabel](s.wnd.Content()); len(labels) > 1 {
			text = labels[1].Title()
		}
	})
	return text
}

func (s *sellTestDialog) sellEnabled() bool {
	var enabled bool
	s.screen.Do(func() { enabled = s.sell.Enabled() })
	return enabled
}

// replaceText types text into the field in place of what it holds, or clears the field when text is empty.
func (s *sellTestDialog) replaceText(field unison.Paneler, text string) {
	s.screen.Click(s.screen.PanelCenter(field))
	s.screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	if text == "" {
		s.screen.KeyPress(unison.KeyBackspace, 0)
		return
	}
	s.screen.Type(text)
}

// dismiss clicks the button and checks that the dialog has gone and the task that opened it has finished.
func (s *sellTestDialog) dismiss(c check.Checker, button *unison.Button) {
	s.screen.Click(s.screen.PanelCenter(button))
	var windows int
	s.screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(*s.done, "the sell command has finished")
}

// TestSellDialogSellsAndUndoes drives the sell dialog from opening to the sale: it opens with all of each item to be
// sold at the sheet's sell percentage, shows what each brings in, the total and the money paid out, follows a change of
// percentage with a warning of what can't be paid, and on Sell removes what was sold, merges the money into the
// matching money already carried or carries it anew, and records the sale, all as one edit that Undo takes back and
// Redo makes again.
func TestSellDialogSellsAndUndoes(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var sword, knife *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.eagle), money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			sword = newTradeTestItem(entity, "Shortsword", "400", fxp.One)
			knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			eagles := money.eagle.Clone(money.libFile, entity, nil, gurps.Reference)
			eagles.Quantity = fxp.FromInteger(7)
			return []*gurps.Equipment{sword, knife, eagles}
		})
	before := carriedSummary(sheet)
	c.Equal([]string{"Shortsword ×1", "Small Knife ×2", "Silver Eagle ×7"}, before)

	s := openSellTestDialog(t, screen, wnd, sheet, sword, knife)
	screen.Do(func() {
		c.Equal("Sell Equipment", s.wnd.Title(), "the dialog is titled for what it does")
		c.Equal("50", s.percent.Text(), "the percentage starts at the sheet's sell percent")
		c.Equal("1", s.quantities[0].Text(), "all of the sword is to be sold")
		c.True(s.quantities[0].Enabled(), "a quantity of one can be changed, to leave the item out")
		c.Equal("2", s.quantities[1].Text(), "all of the knives are to be sold")
		c.True(s.quantities[1].Enabled(), "a quantity of more than one can be changed")
		c.Equal("Small Knife quantity to sell", s.quantities[1].Accessibility.Name)
		c.Equal("Sold Shortsword, Small Knife", s.name.Text(), "the transaction is named for what is sold")
		c.Equal(uncheck.On, s.merge.State, "money is merged by default")
	})
	texts := s.texts()
	for _, want := range []string{
		"of 1", "of 2", "$400.00", "$30.00", "$200.00", "Total: $230.00",
		"You receive: Silver Eagle ×11, Copper Penny ×10 ($230.00)",
	} {
		c.True(slices.Contains(texts, want), "the dialog must show %q; it shows %v", want, texts)
	}
	c.Equal("", s.warning(), "an exact payout needs no warning")
	c.True(s.sellEnabled(), "the sale can be made")

	s.replaceText(s.percent, "52")
	texts = s.texts()
	for _, want := range []string{
		"$208.00", "$31.20", "Total: $239.20", "You receive: Silver Eagle ×11, Copper Penny ×19 ($239.00)",
	} {
		c.True(slices.Contains(texts, want), "at 52%%, the dialog must show %q; it shows %v", want, texts)
	}
	c.Equal("$0.20 can't be paid with the configured sell money. Receiving $239.00, the closest amount not over the total.",
		s.warning(), "what can't be paid is explained")
	c.True(s.sellEnabled(), "a sale that can't be paid exactly can still be made")
	uxtest.CaptureScreen(t, c, screen, "sell_dialog")
	c.Equal(before, carriedSummary(sheet), "nothing is sold before the sale is accepted")

	s.replaceText(s.name, "Market day")
	s.dismiss(c, s.sell)
	screen.Do(func() {
		entity := sheet.Entity()
		c.Equal([]string{"Silver Eagle ×18", "Copper Penny ×19"}, carriedSummary(sheet),
			"the sold items are gone, the eagles merged into those carried and the pennies carried anew")
		c.Equal(1, len(entity.ShoppingRecord), "the sale is recorded")
		if len(entity.ShoppingRecord) == 1 {
			entry := entity.ShoppingRecord[0]
			c.Equal(gurps.ShoppingSold, entry.Kind)
			c.Equal("Market day", entry.Name)
			c.Equal([]gurps.ShoppingRecordItem{
				{Name: "Shortsword", Quantity: fxp.One, Value: fxp.FromInteger(400)},
				{Name: "Small Knife", Quantity: fxp.Two, Value: fxp.FromInteger(30)},
			}, entry.Items, "the record lists what was sold")
			c.Equal([]gurps.ShoppingRecordItem{
				{Name: "Silver Eagle", Quantity: fxp.FromInteger(11), Value: fxp.FromInteger(20)},
				{Name: "Copper Penny", Quantity: fxp.FromInteger(19), Value: fxp.One},
			}, entry.Money, "the record lists the money received")
			c.Equal(fxp.FromInteger(239), entry.Exchanged, "the record holds the amount received")
		}
		c.True(sheet.undoMgr.CanUndo(), "the sale can be undone")
		c.Equal("Undo Sell Equipment", sheet.undoMgr.UndoTitle())
		c.Equal(2, len(sheet.CarriedEquipment.Table.RootRows()), "the list shows what is now carried")
	})

	screen.Do(sheet.undoMgr.Undo)
	screen.Do(func() {
		c.Equal(before, carriedSummary(sheet), "undo puts back what was carried")
		c.Equal(0, len(sheet.Entity().ShoppingRecord), "undo takes the sale off the record")
		c.Equal(3, len(sheet.CarriedEquipment.Table.RootRows()), "the list shows what was carried")
	})
	screen.Do(sheet.undoMgr.Redo)
	screen.Do(func() {
		c.Equal([]string{"Silver Eagle ×18", "Copper Penny ×19"}, carriedSummary(sheet), "redo sells again")
		c.Equal(1, len(sheet.Entity().ShoppingRecord), "redo records the sale again")
		if len(sheet.Entity().ShoppingRecord) == 1 {
			c.Equal("Market day", sheet.Entity().ShoppingRecord[0].Name)
		}
	})
}

// TestSellDialogGatesSale checks that Sell is unavailable, with the reason given, while the sheet has no sell money or
// none of it can be found in the libraries, and while nothing is to be sold, and that canceling the dialog changes
// nothing.
func TestSellDialogGatesSale(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var knife *gurps.Equipment
	sheet := openSellTestSheet(t, screen, nil, func(entity *gurps.Entity) []*gurps.Equipment {
		knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
		return []*gurps.Equipment{knife}
	})
	before := carriedSummary(sheet)
	unchanged := func(why string) {
		t.Helper()
		screen.Do(func() {
			c.Equal(before, carriedSummary(sheet), "%s: nothing is sold", why)
			c.Equal(0, len(sheet.Entity().ShoppingRecord), "%s: nothing is recorded", why)
			c.False(sheet.undoMgr.CanUndo(), "%s: there is nothing to undo", why)
		})
	}

	s := openSellTestDialog(t, screen, wnd, sheet, knife)
	c.False(s.sellEnabled(), "without sell money, nothing can be sold")
	c.Equal("No sell money is configured. Add money items to the Trading section of this sheet's settings.",
		s.warning(), "the dialog says how to add sell money")
	s.dismiss(c, s.cancel)
	unchanged("with no sell money")

	screen.Do(func() {
		sheet.Entity().SheetSettings.SellMoney = []gurps.Source{{LibraryFile: money.libFile, TID: "missing"}}
	})
	s = openSellTestDialog(t, screen, wnd, sheet, knife)
	c.False(s.sellEnabled(), "with sell money that can't be found, nothing can be sold")
	c.Equal("None of the configured sell money could be found in the libraries.", s.warning(),
		"the dialog says the sell money can't be found")
	s.dismiss(c, s.cancel)
	unchanged("with missing sell money")

	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = []gurps.Source{money.source(money.penny)} })
	s = openSellTestDialog(t, screen, wnd, sheet, knife)
	c.True(s.sellEnabled(), "with sell money, the knives can be sold")
	c.Equal("", s.warning())
	s.replaceText(s.quantities[0], "0")
	c.False(s.sellEnabled(), "with nothing to sell, nothing can be sold")
	s.replaceText(s.quantities[0], "1")
	c.True(s.sellEnabled(), "with something to sell again, it can be sold")
	c.True(slices.Contains(s.texts(), "You receive: Copper Penny ×15 ($15.00)"), "one knife brings in $15")
	s.dismiss(c, s.cancel)
	unchanged("when canceled")
}

// TestSellDialogClampsTheSheetsSellPercent checks that a sell percentage outside the range the dialog's field accepts,
// which the settings allow, opens the dialog at the nearest the field accepts rather than failing to open it.
func TestSellDialogClampsTheSheetsSellPercent(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var knife *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			return []*gurps.Equipment{knife}
		})
	for _, tt := range []struct {
		percent  string
		expected string
	}{
		{percent: "2000", expected: "1000"},
		{percent: "0.5", expected: "1"},
	} {
		screen.Do(func() { sheet.Entity().SheetSettings.SellPercent = fxp.FromStringForced(tt.percent) })
		s := openSellTestDialog(t, screen, wnd, sheet, knife)
		var text string
		screen.Do(func() { text = s.percent.Text() })
		c.Equal(tt.expected, text, "a sell percent of %s opens the dialog at %s", tt.percent, tt.expected)
		s.dismiss(c, s.cancel)
	}
}

// TestSellDialogNamesWhatIsSold checks that the suggested transaction name follows the quantities, naming only the items
// of which something is to be sold, until the name is edited, after which it is left alone, and that a name left empty
// falls back to the suggestion.
func TestSellDialogNamesWhatIsSold(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var sword, knife *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			sword = newTradeTestItem(entity, "Shortsword", "400", fxp.One)
			knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			return []*gurps.Equipment{sword, knife}
		})
	name := func(s *sellTestDialog) string {
		var text string
		screen.Do(func() { text = s.name.Text() })
		return text
	}

	s := openSellTestDialog(t, screen, wnd, sheet, sword, knife)
	c.Equal("Sold Shortsword, Small Knife", name(s))
	s.replaceText(s.quantities[1], "0")
	c.Equal("Sold Shortsword", name(s), "an item of which nothing is sold leaves the name")
	s.replaceText(s.quantities[1], "1")
	c.Equal("Sold Shortsword, Small Knife", name(s), "and comes back once some of it is")
	s.replaceText(s.name, "Market day")
	s.replaceText(s.quantities[1], "0")
	c.Equal("Market day", name(s), "an edited name is left alone")
	s.replaceText(s.name, "")
	s.dismiss(c, s.sell)
	screen.Do(func() {
		record := sheet.Entity().ShoppingRecord
		c.Equal(1, len(record))
		if len(record) == 1 {
			c.Equal("Sold Shortsword", record[0].Name, "an empty name falls back to naming what was sold")
		}
	})
}

// TestSellDialogClosesEditorsOfWhatTheSaleChanges checks that a sale first closes the editors open on everything it
// changes, as deleting them does: the items sold in full, along with what a container sold holds, the items sold in
// part and the carried money the payouts merge into, since an editor left open would put its stale quantity back when
// applied. The editor of carried money the payouts don't touch stays open.
func TestSellDialogClosesEditorsOfWhatTheSaleChanges(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var sword, knives, backpack, rope, eagles, gems *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.eagle), money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			sword = newTradeTestItem(entity, "Shortsword", "400", fxp.One)
			knives = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			backpack = gurps.NewEquipment(entity, nil, true)
			backpack.Name = "Backpack"
			rope = newTradeTestItem(entity, "Rope", "5", fxp.One)
			rope.SetParent(backpack)
			backpack.Children = []*gurps.Equipment{rope}
			eagles = money.eagle.Clone(money.libFile, entity, nil, gurps.Reference)
			eagles.Quantity = fxp.FromInteger(7)
			gems = newTradeTestItem(entity, "Gem", "100", fxp.Three, gurps.MoneyTag)
			return []*gurps.Equipment{sword, knives, backpack, eagles, gems}
		})
	editorsOpenFor := func(items ...*gurps.Equipment) int {
		var count int
		screen.Do(func() {
			count = len(AllMatchingDockables(func(d unison.Dockable) bool {
				e, ok := d.(*editor[*gurps.Equipment, *gurps.EquipmentEditData])
				return ok && slices.Contains(items, e.target)
			}))
		})
		return count
	}
	screen.Do(func() {
		for _, eqp := range []*gurps.Equipment{sword, knives, rope, eagles, gems} {
			EditEquipment(sheet, eqp, true)
		}
	})
	c.Equal(5, editorsOpenFor(sword, knives, rope, eagles, gems), "the editors must be open")

	s := openSellTestDialog(t, screen, wnd, sheet, sword, knives, backpack)
	s.replaceText(s.quantities[1], "1")
	s.dismiss(c, s.sell)
	screen.Do(func() {
		c.Equal([]string{"Small Knife ×1", "Silver Eagle ×17", "Gem ×3", "Copper Penny ×17"}, carriedSummary(sheet),
			"the sword and the backpack are sold in full and one of the knives, and the eagles paid merge into those carried")
	})
	c.Equal(0, editorsOpenFor(sword, rope), "the editors of the sword and of the rope in the backpack are closed")
	c.Equal(0, editorsOpenFor(knives), "the editor of the knives, one of which is sold, is closed")
	c.Equal(0, editorsOpenFor(eagles), "the editor of the eagles the payout merges into is closed")
	c.Equal(1, editorsOpenFor(gems), "the editor of money the sale doesn't touch stays open")
}

// TestSellDialogClosesTheEditorOfAContainerTheSaleSplits checks that selling part of what a container with several
// copies holds, in an amount that splits the container apart, closes the editor open on the container, whose quantity
// the split reduces, while the editor of a container whose copies share the sale evenly stays open.
func TestSellDialogClosesTheEditorOfAContainerTheSaleSplits(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var box, gem, crate, torches *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.eagle), money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			box = gurps.NewEquipment(entity, nil, true)
			box.Name = "Box"
			box.Quantity = fxp.Three
			gem = newTradeTestItem(entity, "Gem", "40", fxp.One)
			gem.SetParent(box)
			box.Children = []*gurps.Equipment{gem}
			crate = gurps.NewEquipment(entity, nil, true)
			crate.Name = "Crate"
			crate.Quantity = fxp.Two
			torches = newTradeTestItem(entity, "Torch", "2", fxp.Two)
			torches.SetParent(crate)
			crate.Children = []*gurps.Equipment{torches}
			return []*gurps.Equipment{box, crate}
		})
	editorsOpenFor := func(items ...*gurps.Equipment) int {
		var count int
		screen.Do(func() {
			count = len(AllMatchingDockables(func(d unison.Dockable) bool {
				e, ok := d.(*editor[*gurps.Equipment, *gurps.EquipmentEditData])
				return ok && slices.Contains(items, e.target)
			}))
		})
		return count
	}
	screen.Do(func() {
		for _, eqp := range []*gurps.Equipment{box, crate} {
			EditEquipment(sheet, eqp, true)
		}
	})
	c.Equal(2, editorsOpenFor(box, crate), "the editors must be open")

	s := openSellTestDialog(t, screen, wnd, sheet, gem, torches)
	s.replaceText(s.quantities[0], "1")
	s.replaceText(s.quantities[1], "2")
	s.dismiss(c, s.sell)
	screen.Do(func() {
		c.Equal([]string{"Box ×2", "Box ×1", "Crate ×2", "Silver Eagle ×1", "Copper Penny ×2"}, carriedSummary(sheet),
			"the gem comes out of one box, which is split off, and a torch out of each crate")
		c.Equal(fxp.One, torches.Quantity)
	})
	c.Equal(0, editorsOpenFor(box), "the editor of the box, whose quantity the split reduces, is closed")
	c.Equal(1, editorsOpenFor(crate), "the editor of the crate, which the sale doesn't split, stays open")
}

// TestSellDialogLeavesOutWhatIsSetToZero checks that any item is left out of the sale when its quantity is set to zero,
// one of which only one is carried and a group included; that an item can be sold in part, a fraction of it included,
// while a container, which is sold by the copy, refuses a fraction; and that a group, which is sold in full, takes any
// other quantity as the whole of it.
func TestSellDialogLeavesOutWhatIsSetToZero(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var sword, box, crate, tools *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			sword = newTradeTestItem(entity, "Shortsword", "400", fxp.One)
			box = gurps.NewEquipment(entity, nil, true)
			box.Name = "Box"
			box.BaseValue = "10"
			box.Quantity = fxp.Three
			crate = gurps.NewEquipment(entity, nil, true)
			crate.Name = "Crate"
			crate.Quantity = fxp.Three
			tools = gurps.NewEquipmentGroup(entity, crate)
			tools.Name = "Tools"
			hammer := newTradeTestItem(entity, "Hammer", "15", fxp.One)
			hammer.SetParent(tools)
			tools.Children = []*gurps.Equipment{hammer}
			crate.Children = []*gurps.Equipment{tools}
			return []*gurps.Equipment{sword, box, crate}
		})
	s := openSellTestDialog(t, screen, wnd, sheet, sword, box, tools)
	text := func(field *DecimalField) string {
		var text string
		screen.Do(func() { text = field.Text() })
		return text
	}
	total := func(want, why string) {
		t.Helper()
		texts := s.texts()
		c.True(slices.Contains(texts, "Total: "+want), "%s; the dialog shows %v", why, texts)
	}
	screen.Do(func() {
		c.Equal("1", s.quantities[0].Text())
		c.True(s.quantities[0].Enabled(), "a quantity of one can be changed")
		c.Equal("3", s.quantities[1].Text(), "every copy of the box is to be sold")
		c.Equal("3", s.quantities[2].Text(), "the group comes once for each copy of the crate")
		c.True(s.quantities[2].Enabled(), "a group's quantity can be changed")
	})
	total("$237.50", "the sword, the boxes and the tools in each crate are to be sold")
	s.replaceText(s.quantities[0], "0")
	total("$37.50", "the sword set to zero is left out")
	c.True(s.sellEnabled(), "the rest can still be sold")
	screen.Do(func() { c.Equal("Sold Box, Tools", s.name.Text(), "the name leaves out the sword") })
	s.replaceText(s.quantities[0], "0.5")
	c.Equal("0.5", text(s.quantities[0]), "an item can be sold in part, a fraction of it included")
	total("$137.50", "half the sword is to be sold")
	s.replaceText(s.quantities[1], "2.5")
	c.Equal("2", text(s.quantities[1]), "a container refuses the decimal point, and then a quantity above the copies carried")
	total("$132.50", "two of the boxes are to be sold")
	screen.Do(func() { unison.ClipboardSetText("1.5") })
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.KeyPress(unison.KeyV, mod.OSMenuCommand())
	c.Equal("1.5", text(s.quantities[1]), "a pasted fraction shows")
	c.False(s.sellEnabled(), "a container's pasted fraction isn't a valid quantity")
	s.replaceText(s.quantities[1], "2")
	c.True(s.sellEnabled(), "a whole number is valid again")
	s.replaceText(s.quantities[2], "1")
	c.Equal("3", text(s.quantities[2]), "any quantity of a group is the whole of it")
	total("$132.50", "the tools are still sold in full")
	s.replaceText(s.quantities[2], "0")
	total("$110.00", "the tools set to zero are left out")
	s.replaceText(s.quantities[1], "0")
	s.replaceText(s.quantities[0], "0")
	c.False(s.sellEnabled(), "with everything set to zero, nothing can be sold")
	total("$0.00", "nothing is to be sold")
	s.replaceText(s.quantities[0], "1")
	c.True(s.sellEnabled(), "the sword set back to one can be sold")
	s.dismiss(c, s.sell)
	screen.Do(func() {
		c.Equal([]string{"Box ×3", "Crate ×3", "Copper Penny ×200"}, carriedSummary(sheet),
			"the sword is sold and the boxes and tools kept")
		c.Equal([]*gurps.Equipment{tools}, crate.Children, "the tools stay in the crates")
	})
}

// TestSellDialogGatesSaleOnValidFields checks that Sell is unavailable while the percent field or a quantity field holds
// text that isn't a valid value, since the sale would otherwise go through at a value other than the one typed, and is
// available again once the text is valid.
func TestSellDialogGatesSaleOnValidFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var sword, knives *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			sword = newTradeTestItem(entity, "Shortsword", "400", fxp.One)
			knives = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			return []*gurps.Equipment{sword, knives}
		})
	s := openSellTestDialog(t, screen, wnd, sheet, sword, knives)
	c.True(s.sellEnabled())
	s.replaceText(s.percent, "")
	c.False(s.sellEnabled(), "an empty percent field holds no valid value")
	s.replaceText(s.percent, "5000")
	c.False(s.sellEnabled(), "a percent above the field's maximum isn't valid")
	s.replaceText(s.percent, "50")
	c.True(s.sellEnabled(), "a valid percent allows the sale again")
	s.replaceText(s.quantities[1], "")
	c.False(s.sellEnabled(), "an empty quantity field holds no valid value, even while another item is to be sold")
	s.replaceText(s.quantities[1], "2")
	c.True(s.sellEnabled())
	s.dismiss(c, s.cancel)
}

// TestSellDialogRefusesPastedQuantitiesAboveTheMaximum checks that a quantity field refuses text pasted with the
// keyboard shortcut that would put it above the quantity carried, as it does text typed or pasted by command: the
// dialog is modal and so has no menu bar to route the shortcut through, and the field pastes on it itself.
func TestSellDialogRefusesPastedQuantitiesAboveTheMaximum(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var knives *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			knives = newTradeTestItem(entity, "Small Knife", "30", fxp.FromInteger(11))
			return []*gurps.Equipment{knives}
		})
	s := openSellTestDialog(t, screen, wnd, sheet, knives)
	screen.Do(func() { unison.ClipboardSetText("500") })
	quantity := s.quantities[0]
	screen.Click(screen.PanelCenter(quantity))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.KeyPress(unison.KeyV, mod.OSMenuCommand())
	var text string
	screen.Do(func() { text = quantity.Text() })
	c.Equal("11", text, "a pasted quantity above the number carried is refused")
	screen.Do(func() { unison.ClipboardSetText("5") })
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.KeyPress(unison.KeyV, mod.OSMenuCommand())
	screen.Do(func() { text = quantity.Text() })
	c.Equal("5", text, "a pasted quantity within the number carried is taken")
	s.dismiss(c, s.cancel)
}

// TestSellSkipsDisabledEquipment checks that disabled equipment, and equipment held by a disabled container, can't be
// sold: the Sell command is off while nothing else is selected, and the dialog leaves it out of a wider selection.
func TestSellSkipsDisabledEquipment(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var knife, dagger, sack, gem *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			dagger = newTradeTestItem(entity, "Dagger", "20", fxp.One)
			dagger.SetDisabled(true)
			sack = gurps.NewEquipment(entity, nil, true)
			sack.Name = "Sack"
			sack.SetDisabled(true)
			gem = gurps.NewEquipment(entity, sack, false)
			gem.Name = "Gem"
			gem.BaseValue = "100"
			sack.Children = []*gurps.Equipment{gem}
			return []*gurps.Equipment{knife, dagger, sack}
		})
	canSell := func(items ...*gurps.Equipment) bool {
		var enabled bool
		screen.Do(func() {
			table := sheet.CarriedEquipment.Table
			selection := make(map[tid.TID]bool, len(items))
			for _, eqp := range items {
				selection[eqp.TID] = true
			}
			table.SetSelectionMap(selection)
			enabled = table.CanPerformCmd(table, SellEquipmentItemID)
		})
		return enabled
	}
	c.True(canSell(knife), "the knives can be sold")
	c.False(canSell(dagger), "a disabled item can't be sold")
	c.False(canSell(sack), "nor can a disabled container")
	c.False(canSell(gem), "nor can an item in a disabled container")
	c.True(canSell(dagger, gem, knife), "the knives can be sold from a selection holding them")

	s := openSellTestDialog(t, screen, wnd, sheet, dagger, gem, knife)
	texts := s.texts()
	c.True(slices.Contains(texts, "Small Knife"), "the dialog lists the knives; it shows %v", texts)
	c.False(slices.Contains(texts, "Dagger"), "the dialog leaves out the disabled dagger")
	c.False(slices.Contains(texts, "Gem"), "the dialog leaves out the gem in the disabled sack")
	s.dismiss(c, s.sell)
	c.Equal([]string{"Dagger ×1", "Sack ×1", "Copper Penny ×30"}, carriedSummary(sheet),
		"only the knives are sold")
}

// TestSellDialogListsWhatAContainerHolds checks that a container with something in it has an information icon beside
// its name whose tooltip lists what it holds: each item with its quantity, indented beneath the container holding it, a
// group without a quantity and a disabled item struck through, with the rest counted past maxSaleContentsLines of them.
// An item that holds nothing has no icon.
func TestSellDialogListsWhatAContainerHolds(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var pack, sack, crate, knife, box *gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			container := func(parent *gurps.Equipment, name string, quantity fxp.Int, children ...*gurps.Equipment) *gurps.Equipment {
				eqp := gurps.NewEquipment(entity, parent, true)
				eqp.Name = name
				eqp.BaseValue = "10"
				eqp.Quantity = quantity
				eqp.Children = children
				for _, child := range children {
					child.SetParent(eqp)
				}
				return eqp
			}
			hammer := newTradeTestItem(entity, "Hammer", "15", fxp.One)
			hammer.SetDisabled(true)
			tools := container(nil, "Tools", fxp.One, hammer)
			tools.ContainerType = eqcontainer.Group
			pack = container(nil, "Backpack", fxp.Two, newTradeTestItem(entity, "Rope", "5", fxp.Three), tools)
			sack = container(nil, "Sack", fxp.One,
				container(nil, "Pouch", fxp.One, newTradeTestItem(entity, "Gem", "100", fxp.Five)))
			nails := make([]*gurps.Equipment, maxSaleContentsLines+2)
			for i := range nails {
				nails[i] = newTradeTestItem(entity, fmt.Sprintf("Nail %d", i+1), "1", fxp.One)
			}
			crate = container(nil, "Crate", fxp.One, nails...)
			knife = newTradeTestItem(entity, "Small Knife", "30", fxp.Two)
			box = container(nil, "Box", fxp.One)
			return []*gurps.Equipment{pack, sack, crate, knife, box}
		})
	s := openSellTestDialog(t, screen, wnd, sheet, pack, sack, crate, knife, box)
	type tip struct {
		texts  []string
		struck []string
		indent map[string]float32
	}
	var tips []tip
	screen.Do(func() {
		for _, label := range uxtest.PanelsOfType[*unison.Label](s.wnd.Content()) {
			if label.Drawable == nil {
				continue
			}
			one := tip{texts: uxtest.LabelTexts(label.Tooltip), indent: make(map[string]float32)}
			for _, line := range uxtest.PanelsOfType[*unison.Label](label.Tooltip) {
				if line.StrikeThrough {
					one.struck = append(one.struck, line.String())
				}
				if border := line.Border(); border != nil {
					one.indent[line.String()] = border.Insets().Left
				}
			}
			tips = append(tips, one)
		}
	})
	c.Equal(3, len(tips), "the backpack, sack and crate have an information icon apiece; the knife and empty box have none")
	if len(tips) != 3 {
		s.dismiss(c, s.cancel)
		return
	}
	c.Equal([]string{"Sold with each one:", "Rope ×3", "Tools", "Hammer ×1"}, tips[0].texts,
		"the tooltip for the two backpacks lists what each holds, the group without a quantity")
	c.Equal([]string{"Hammer ×1"}, tips[0].struck, "the disabled hammer is struck through")
	c.True(tips[0].indent["Hammer ×1"] > tips[0].indent["Tools"], "the hammer is indented beneath the tools")
	c.Equal(tips[0].indent["Rope ×3"], tips[0].indent["Tools"], "the rope and tools are indented alike")
	c.Equal([]string{"Sold with it:", "Pouch ×1", "Gem ×5"}, tips[1].texts,
		"the tooltip for the one sack lists what it holds")
	c.True(tips[1].indent["Gem ×5"] > tips[1].indent["Pouch ×1"], "the gem is indented beneath the pouch")
	c.Equal(maxSaleContentsLines+2, len(tips[2].texts), "the crate's tooltip lists the first nails and counts the rest")
	c.Equal("Nail 1 ×1", tips[2].texts[1])
	c.Equal(fmt.Sprintf("Nail %d ×1", maxSaleContentsLines), tips[2].texts[maxSaleContentsLines])
	c.Equal("…and 2 more", tips[2].texts[maxSaleContentsLines+1])
	s.dismiss(c, s.cancel)
}

// TestSellDialogLinesUpTheList checks the look of the list of items for sale: with more items than it shows at once, it
// scrolls, cut between rows rather than through one, and its header and its total, which don't scroll, line up with the
// rows' columns. With nothing to warn of, no warning takes up room.
func TestSellDialogLinesUpTheList(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	var items []*gurps.Equipment
	sheet := openSellTestSheet(t, screen, []gurps.Source{money.source(money.penny)},
		func(entity *gurps.Entity) []*gurps.Equipment {
			for i := range maxSaleRowsShown + 5 {
				items = append(items, newTradeTestItem(entity, fmt.Sprintf("Item %d", i+1), "10", fxp.One))
			}
			return items
		})
	s := openSellTestDialog(t, screen, wnd, sheet, items...)
	uxtest.CaptureScreen(t, c, screen, "sell_dialog_long")
	screen.Do(func() {
		content := s.wnd.Content()
		scroll, ok := uxtest.FirstPanelOfType[*unison.ScrollPanel](content)
		if !ok {
			t.Fatal("the dialog must scroll its list")
		}
		view := scroll.ContentView().RectToRoot(scroll.ContentView().ContentRect(false))
		list := scroll.Content().AsPanel()
		c.True(list.FrameRect().Height > view.Height, "the list has more rows than it shows")
		shown := s.quantities[maxSaleRowsShown-1].RectToRoot(s.quantities[maxSaleRowsShown-1].ContentRect(true))
		hidden := s.quantities[maxSaleRowsShown].RectToRoot(s.quantities[maxSaleRowsShown].ContentRect(true))
		c.True(shown.Bottom() <= view.Bottom(), "the last row shown is shown whole: %v within %v", shown, view)
		c.True(hidden.Y >= view.Bottom(), "the first row past those shown is hidden whole: %v below %v", hidden, view)
		rights := make(map[string]float32)
		for _, label := range uxtest.PanelsOfType[*unison.Label](content) {
			if text := label.String(); text == "You Get (50%)" || text == "$5.00" || strings.HasPrefix(text, "Total: ") {
				rights[text] = label.RectToRoot(label.ContentRect(true)).Right()
			}
		}
		c.Equal(3, len(rights), "the header, a row and the footer are found: %v", rights)
		c.Equal(rights["You Get (50%)"], rights["$5.00"], "the header lines up with the rows")
		c.Equal(rights["You Get (50%)"], rights["Total: $75.00"], "the total lines up with the rows")
		c.Equal(1, len(uxtest.PanelsOfType[*TextLabel](content)), "with nothing to warn of, only the payout is shown")
	})
	s.dismiss(c, s.cancel)
}
