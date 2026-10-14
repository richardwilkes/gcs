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

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// paymentColumnCount is the number of columns in the list of money to pay with: the money item, the quantity carried,
// the value of one, the quantity to spend and what that comes to.
const paymentColumnCount = 5

// chooseManualPayment asks the user which of the money the sheet's character carries (see gurps.Entity.MoneyItems) to
// pay owed with, for what, which names the items being bought, starting from the quantities in current, by money item
// ID. It returns the quantities chosen, by ID, or useAutomatic set when the user asked to have the money chosen
// automatically again, with ok false when the user canceled. It is a variable so that tests can substitute it.
var chooseManualPayment = func(sheet *Sheet, owed fxp.Int, what string, current map[tid.TID]int) (chosen map[tid.TID]int, useAutomatic, ok bool) {
	d := newManualPaymentDialog(sheet, owed, what, current)
	if d == nil {
		return nil, false, false
	}
	switch d.dialog.RunModal() {
	case unison.ModalResponseOK:
		return d.chosen(), false, true
	case unison.ModalResponseUserBase:
		return nil, true, true
	default:
		return nil, false, false
	}
}

// manualPaymentDialog is the dialog chooseManualPayment puts up.
type manualPaymentDialog struct {
	dialog    *unison.Dialog
	entity    *gurps.Entity
	rows      []*manualPaymentRow
	owed      fxp.Int
	sellMoney []*gurps.Equipment
	change    manualPaymentChange
	spending  *unison.Label
	changeDue *unison.Label
	note      *manualPaymentNote
}

// manualPaymentRow is one carried money item and the quantity of it to spend.
type manualPaymentRow struct {
	money    *gurps.Equipment
	value    fxp.Int
	spend    int
	spending *unison.Label
}

// manualPaymentChange is the change worked out for an overpayment, kept for as long as the amount and the money paid
// with, and how much of it, stay the same, since the search is costly. The money items are the dialog's own rows, so
// they can be compared as they are.
type manualPaymentChange struct {
	over    fxp.Int
	given   []gurps.Payout
	payouts []gurps.Payout
	paid    fxp.Int
	valid   bool
}

func newManualPaymentDialog(sheet *Sheet, owed fxp.Int, what string, current map[tid.TID]int) *manualPaymentDialog {
	entity := sheet.Entity()
	d := &manualPaymentDialog{entity: entity, owed: owed, sellMoney: entity.SellMoneyItems()}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	list := newTradeList(paymentColumnCount)
	list.addHeader(i18n.Text("Money Item"), 1, align.Start)
	list.addHeader(i18n.Text("Carried"), 1, align.End)
	list.addHeader(i18n.Text("Value Each"), 1, align.End)
	list.addHeader(i18n.Text("Spend"), 1, align.Start)
	list.addHeader(i18n.Text("Spending"), 1, align.End)
	money := entity.MoneyItems()
	for _, one := range money {
		row := &manualPaymentRow{money: one, value: one.AdjustedValue()}
		carried := one.TradeQuantity()
		most := max(carried.Floor().AsInteger[int](), 0)
		row.spend = min(max(current[one.TID], 0), most)
		d.rows = append(d.rows, row)
		name := one.String()
		list.rows.AddChild(newManualPaymentText(name, align.Start))
		list.rows.AddChild(newManualPaymentText(carried.Comma(), align.End))
		list.rows.AddChild(newManualPaymentText(i18n.Text("$%s", d.format(row.value)), align.End))
		field := NewIntegerField(nil, "", i18n.Text("Spend"),
			func() int { return row.spend },
			func(v int) {
				row.spend = v
				d.refresh()
			}, 0, most, false, false)
		field.withoutUndo()
		field.SetMarksModified(false)
		field.refuseAboveMax()
		field.Accessibility.Name = i18n.Text("%s to spend", name)
		field.Tooltip = newWrappedTooltip(i18n.Text("The number of %s to spend, up to the %d carried", name, most))
		field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle})
		list.rows.AddChild(field)
		row.spending = newManualPaymentText("", align.End)
		list.rows.AddChild(row.spending)
	}
	if len(money) == 0 {
		label := newManualPaymentText(i18n.Text("This character carries no money-tagged items."), align.Start)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: paymentColumnCount})
		list.rows.AddChild(label)
	}
	list.addTotal().SetTitle(i18n.Text("Owed: $%s", d.format(owed)))
	d.spending = list.addTotal()
	d.changeDue = list.addTotal()
	d.changeDue.Font = unison.EmphasizedSystemFont

	intro := NewWrappingLabel()
	intro.SetTitle(i18n.Text("Money-tagged items carried by this character.\nOwed: $%s for %s.", d.format(owed), what))
	wrapToWidthOf(intro.AsPanel(), list.box)
	intro.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	panel.AddChild(intro)
	panel.AddChild(list.box)
	d.note = newManualPaymentNote()
	wrapToWidthOf(d.note.AsPanel(), list.box)
	d.note.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	panel.AddChild(d.note)

	var err error
	if d.dialog, err = newPromptDialog(promptOperation{name: i18n.Text("Choose Money to Spend")}, nil, nil, panel,
		&unison.DialogButtonInfo{Title: i18n.Text("Use Automatic"), ResponseCode: unison.ModalResponseUserBase},
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo(),
	); err != nil {
		errs.Log(err)
		return nil
	}
	holdMinSizeOnDisplay(d.dialog.Window())
	automaticTip := i18n.Text("Go back to choosing the money to spend automatically")
	d.dialog.Button(unison.ModalResponseUserBase).Tooltip = newWrappedTooltip(automaticTip)
	// The fields couldn't reach the buttons while they were being created, so the state is brought up to date now.
	d.refresh()
	return d
}

// newManualPaymentText returns a label showing text, aligned within its cell as given.
func newManualPaymentText(text string, hAlign align.Enum) *unison.Label {
	label := unison.NewLabel()
	label.SetTitle(text)
	label.SetLayoutData(&unison.FlexLayoutData{HAlign: hAlign, VAlign: align.Middle})
	return label
}

// refresh brings the totals, the note and the OK button up to date with the quantities chosen.
func (d *manualPaymentDialog) refresh() {
	var spending fxp.Int
	for _, row := range d.rows {
		amount := row.value.Mul(fxp.FromInteger(row.spend))
		setLabelTitle(row.spending, i18n.Text("$%s", d.format(amount)))
		spending += amount
	}
	setLabelTitle(d.spending, i18n.Text("Spending: $%s", d.format(spending)))
	over := max(spending-d.owed, 0)
	setLabelTitle(d.changeDue, i18n.Text("Change due: $%s", d.format(over)))
	short := spending < d.owed
	switch {
	case short:
		d.note.set(i18n.Text("Short by $%s.", d.format(d.owed-spending)), svg.Not, unison.ThemeOnWarning, true)
	case over == 0:
		d.note.set(i18n.Text("Covered exactly. Money you do not spend stays on the sheet."), unison.CheckmarkSVG,
			colors.Success, false)
	default:
		d.describeChange(over)
	}
	if d.dialog != nil {
		d.dialog.Button(unison.ModalResponseOK).SetEnabled(!short)
		growToFitContent(d.dialog.Window())
	}
}

// describeChange has the note say what the change for an overpayment of over will be, or warn when all of it can't be
// returned.
func (d *manualPaymentDialog) describeChange(over fxp.Int) {
	change, paid := d.changeFor(over)
	if unreturned := over - paid; unreturned > 0 {
		text := unreturnedChangeNotice(d.format(unreturned), len(d.entity.SheetSettings.SellMoney), len(d.sellMoney))
		if paid > 0 {
			text = i18n.Text("Change will be returned as %s ($%s). ", describePayouts(change), d.format(paid)) + text
		}
		d.note.set(text, unison.TriangleExclamationSVG, unison.ThemeOnWarning, true)
		return
	}
	d.note.set(i18n.Text("Change will be returned as %s.", describePayouts(change)), svg.Info, unison.ThemeOnSurface,
		false)
}

// changeFor returns the change for an overpayment of over, and what it comes to (see gurps.Entity.ChangeFor).
func (d *manualPaymentDialog) changeFor(over fxp.Int) (change []gurps.Payout, paid fxp.Int) {
	payments := d.payments()
	if !d.change.valid || d.change.over != over || !slices.Equal(d.change.given, payments) {
		payouts, found := d.entity.ChangeFor(over, d.sellMoney, payments)
		d.change = manualPaymentChange{over: over, given: payments, payouts: payouts, paid: found, valid: true}
	}
	return d.change.payouts, d.change.paid
}

// payments returns the money chosen to spend, leaving out items with none.
func (d *manualPaymentDialog) payments() []gurps.Payout {
	payments := make([]gurps.Payout, 0, len(d.rows))
	for _, row := range d.rows {
		if row.spend > 0 {
			payments = append(payments, gurps.Payout{Money: row.money, Quantity: row.spend})
		}
	}
	return payments
}

// chosen returns the quantity of each money item to spend, by ID, leaving out those with none.
func (d *manualPaymentDialog) chosen() map[tid.TID]int {
	m := make(map[tid.TID]int, len(d.rows))
	for _, row := range d.rows {
		if row.spend > 0 {
			m[row.money.TID] = row.spend
		}
	}
	return m
}

func (d *manualPaymentDialog) format(value fxp.Int) string {
	return d.entity.SheetSettings.FormatEquipmentValue(value)
}

// manualPaymentNote is the note beside the totals: an icon beside wrapping text, in a box of the theme's warning color
// when it warns.
type manualPaymentNote struct {
	unison.Panel
	icon    *unison.Label
	text    *TextLabel
	warning bool
}

func newManualPaymentNote() *manualPaymentNote {
	n := &manualPaymentNote{icon: unison.NewLabel(), text: NewWrappingLabel()}
	n.Self = n
	n.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	n.AddChild(n.icon)
	n.text.Font = unison.SystemFont
	n.text.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Start, HGrab: true})
	n.AddChild(n.text)
	n.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		if n.warning {
			r := n.ContentRect(true)
			gc.DrawRoundedRect(r, geom.NewUniformSize(4), unison.ThemeWarning.Paint(gc, r, paintstyle.Fill))
		}
	}
	return n
}

// set shows the text beside the icon, drawn in ink. When warn is true, both sit in a box of the theme's warning color,
// so the ink should then be the theme's color for what is on a warning.
func (n *manualPaymentNote) set(text string, icon *unison.SVG, ink unison.Ink, warn bool) {
	size := xmath.Ceil(n.text.Font.Baseline())
	n.icon.Drawable = &unison.DrawableSVG{SVG: icon, Size: geom.NewSize(size, size)}
	n.icon.OnBackgroundInk = ink
	putOnLine(n.icon.AsPanel(), n.text.Font.LineHeight(), size)
	n.warning = warn
	textInk := unison.Ink(unison.ThemeOnSurface)
	if warn {
		n.SetBorder(unison.NewEmptyBorder(geom.NewSymmetricInsets(unison.StdHSpacing, unison.StdVSpacing)))
		textInk = unison.ThemeOnWarning
	} else {
		n.SetBorder(nil)
	}
	n.text.setText(text, textInk)
	n.MarkForLayoutRecursivelyUpward()
	n.MarkForRedraw()
}
