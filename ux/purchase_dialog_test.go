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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// TestPurchaseDialog drives the dialog that finalizes a purchase: it lists the money carried, by increasing value, with
// the quantities the automatic payment spends, keeps Finalize disabled while the money chosen is short, notes whether
// the payment is short, exact or due change, with what the change will be, puts the automatic payment back on request,
// leaves out money whose checkbox is turned off, and returns the money chosen and whether to merge, or nothing when
// canceled.
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
	// uses are the checkboxes of the dialog last opened that say whether to pay with each kind of money, in the order
	// the money is listed.
	var uses []*unison.CheckBox
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
		var names, useNames, useTitles []string
		uses = nil
		screen.Do(func() {
			fields = uxtest.PanelsOfType[*IntegerField](dialogWnd.Content())
			for _, f := range fields {
				names = append(names, f.Accessibility.Name)
			}
			for _, box := range uxtest.PanelsOfType[*unison.CheckBox](dialogWnd.Content()) {
				if strings.HasPrefix(box.Accessibility.Name, "Pay with ") {
					uses = append(uses, box)
					useNames = append(useNames, box.Accessibility.Name)
					useTitles = append(useTitles, box.Text.String())
				}
			}
		})
		if len(fields) != 2 {
			t.Fatalf("the dialog must have a field for each kind of money carried, not %d", len(fields))
		}
		c.Equal([]string{"Copper Penny to spend", "Silver Eagle to spend"}, names,
			"the money is listed by increasing value")
		c.Equal([]string{"Pay with Copper Penny", "Pay with Silver Eagle"}, useNames,
			"each kind of money has a checkbox saying whether to pay with it")
		c.Equal([]string{"Copper Penny", "Silver Eagle"}, useTitles, "the checkbox names the money")
		return dialog, fields
	}
	controls := func(dialog *unison.Dialog) (automatic *unison.Button, merge *unison.CheckBox) {
		content := dialog.Window().Content()
		for _, b := range uxtest.PanelsOfType[*unison.Button](content) {
			if b.Accessibility.Name == "Choose Automatically" {
				automatic = b
			}
		}
		for _, box := range uxtest.PanelsOfType[*unison.CheckBox](content) {
			if box.Accessibility.Name == "Merge with matching items" {
				merge = box
			}
		}
		return automatic, merge
	}
	type dialogState struct {
		labels      []string
		okOn        bool
		automaticOn bool
		merge       uncheck.Enum
		use         []uncheck.Enum
		spend       []string
		spendOn     []bool
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
			for _, box := range uses {
				s.use = append(s.use, box.State)
			}
			for _, f := range fields {
				s.spend = append(s.spend, f.Text())
				s.spendOn = append(s.spendOn, f.Enabled())
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
	c.Equal([]uncheck.Enum{uncheck.On, uncheck.On}, state.use, "every kind of money is paid with to begin with")
	c.Equal([]bool{true, true}, state.spendOn)
	contains(state, "Money Item", "Carried", "Value Each", "Spend", "7", "$20", "11", "$1",
		"Buying Longbow ×1 for $45.",
		"The checked money-tagged items carried pay for it, chosen automatically unless you change the quantities to spend.",
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
	sellMoney := buyTestSellMoney()
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sellMoney[1:] })
	dialog, fields = payWithEagles("3")
	state = read(dialog, fields)
	uxtest.CaptureScreen(t, c, screen, "purchase_change")
	c.Equal("Change will be returned as Copper Penny ×15.", state.note, "the sell money returns the change")
	c.True(state.noteIcon == svg.Info)
	c.False(state.noteWarn)
	click(dialog, unison.ModalResponseCancel)
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = sellMoney[:1] })
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

	// Turning off a kind of money spends none of it. While the payment is the automatic one, it is worked out again
	// from the money left; once it has been changed, the rest is left as it is.
	screen.Do(func() { sheet.Entity().SheetSettings.SellMoney = nil })
	dialog, fields = open()
	screen.Do(func() { focused = dialog.Window().CurrentFocus() == fields[penny].AsPanel() })
	c.True(focused, "the penny field holds the focus to begin with")
	screen.Click(screen.PanelCenter(uses[penny]))
	state = read(dialog, fields)
	uxtest.CaptureScreen(t, c, screen, "purchase_left_out")
	c.Equal([]uncheck.Enum{uncheck.Off, uncheck.On}, state.use)
	c.Equal([]string{"0", "3"}, state.spend, "the automatic payment is worked out again without the pennies")
	c.Equal([]bool{false, true}, state.spendOn, "money that is not paid with can't have a quantity")
	contains(state, "Spending: $60", "Change due: $15")
	c.False(state.automaticOn, "the payment is still the automatic one")
	screen.Do(func() { focused = dialog.Window().CurrentFocus() == uses[penny].AsPanel() })
	c.True(focused, "the focus leaves the field disabled while it had it for the checkbox")
	screen.Click(screen.PanelCenter(uses[eagle]))
	state = read(dialog, fields)
	c.Equal([]string{"0", "0"}, state.spend, "the pennies alone can't pay, so nothing is chosen")
	c.Equal("Short by $45.", state.note)
	c.False(state.okOn)
	c.False(state.automaticOn)
	screen.Click(screen.PanelCenter(uses[eagle]))
	state = read(dialog, fields)
	c.Equal([]string{"0", "3"}, state.spend, "the automatic payment is worked out again with the eagles back")
	screen.Click(screen.PanelCenter(uses[penny]))
	state = read(dialog, fields)
	c.Equal([]uncheck.Enum{uncheck.On, uncheck.On}, state.use)
	c.Equal([]string{"5", "2"}, state.spend, "the automatic payment is worked out again with the pennies back")
	c.Equal([]bool{true, true}, state.spendOn)
	screen.Do(func() { fields[eagle].SetText("1") })
	screen.Click(screen.PanelCenter(uses[penny]))
	state = read(dialog, fields)
	c.Equal([]string{"0", "1"}, state.spend, "a payment that was changed keeps the money still paid with")
	c.Equal([]bool{false, true}, state.spendOn)
	c.Equal("Short by $25.", state.note)
	c.True(state.automaticOn, "the automatic payment from the money left is offered")
	screen.Click(screen.PanelCenter(uses[penny]))
	state = read(dialog, fields)
	c.Equal([]string{"0", "1"}, state.spend, "money turned back on isn't spent until its quantity is changed")
	c.Equal([]bool{true, true}, state.spendOn)
	screen.Do(func() {
		automatic, _ := controls(dialog)
		automatic.Click()
	})
	state = read(dialog, fields)
	c.Equal([]string{"5", "2"}, state.spend)
	screen.Click(screen.PanelCenter(uses[penny]))
	state = read(dialog, fields)
	c.Equal([]string{"0", "3"}, state.spend)
	click(dialog, unison.ModalResponseOK)
	c.True(ok)
	c.Equal(map[tid.TID]int{eagleID: 3}, terms.chosen, "money turned off is left out of the terms")
}
