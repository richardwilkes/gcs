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
// once the dialog has been dismissed and its sale, if any, made.
func openSellTestDialog(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, sheet *Sheet, items ...*gurps.Equipment) *sellTestDialog {
	t.Helper()
	done := false
	s := &sellTestDialog{t: t, screen: screen, done: &done}
	var table *unison.Table[*Node[*gurps.Equipment]]
	screen.Do(func() {
		table = sheet.CarriedEquipment.Table
		selection := make(map[tid.TID]bool, len(items))
		for _, eqp := range items {
			selection[eqp.TID] = true
		}
		table.SetSelectionMap(selection)
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
	if len(s.quantities) != len(items) {
		t.Fatalf("the sell dialog must have a quantity field for each of the %d items; it has %d", len(items),
			len(s.quantities))
	}
	return s
}

// texts returns the text of every label in the dialog, in order.
func (s *sellTestDialog) texts() []string {
	var texts []string
	s.screen.Do(func() { texts = uxtest.LabelTexts(s.wnd.Content()) })
	return texts
}

// warning returns the text of the dialog's warning, which is empty while there is nothing to warn of.
func (s *sellTestDialog) warning() string {
	var text string
	s.screen.Do(func() {
		if label, ok := uxtest.FirstPanelOfType[*TextLabel](s.wnd.Content()); ok {
			text = label.Title()
		}
	})
	return text
}

func (s *sellTestDialog) sellEnabled() bool {
	var enabled bool
	s.screen.Do(func() { enabled = s.sell.Enabled() })
	return enabled
}

// replaceText types text into the field in place of what it holds.
func (s *sellTestDialog) replaceText(field unison.Paneler, text string) {
	s.screen.Click(s.screen.PanelCenter(field))
	s.screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
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
		c.False(s.quantities[0].Enabled(), "a quantity that can only be one can't be changed")
		c.Equal("2", s.quantities[1].Text(), "all of the knives are to be sold")
		c.True(s.quantities[1].Enabled(), "a quantity of more than one can be changed")
		c.Equal("Small Knife quantity to sell", s.quantities[1].Accessibility.Name)
		c.Equal("Sold Shortsword, Small Knife", s.name.Text(), "the transaction is named for what is sold")
		c.Equal(uncheck.On, s.merge.State, "money is merged by default")
	})
	texts := s.texts()
	for _, want := range []string{
		"of 1", "of 2", "$400.00", "$30.00", "$200.00", "Target total: $230.00",
		"You receive: Silver Eagle ×11, Copper Penny ×10 ($230.00)",
	} {
		c.True(slices.Contains(texts, want), "the dialog must show %q; it shows %v", want, texts)
	}
	c.Equal("", s.warning(), "an exact payout needs no warning")
	c.True(s.sellEnabled(), "the sale can be made")

	s.replaceText(s.percent, "52")
	texts = s.texts()
	for _, want := range []string{
		"$208.00", "$31.20", "Target total: $239.20", "You receive: Silver Eagle ×11, Copper Penny ×19 ($239.00)",
	} {
		c.True(slices.Contains(texts, want), "at 52%%, the dialog must show %q; it shows %v", want, texts)
	}
	c.Equal("$0.20 can't be paid with the configured sell money. Receiving $239.00, the closest amount not over the target.",
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
