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
	"cmp"
	"maps"
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
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// paymentColumnCount is the number of columns in the list of money to pay with: the money item, the quantity carried,
// the value of one, the quantity to spend and what that comes to.
const paymentColumnCount = 5

// purchaseTerms is what the user settles when finalizing a purchase: the quantity of each carried money item to pay
// with, by ID, and whether to merge what the purchase brings with matching carried items.
type purchaseTerms struct {
	chosen map[tid.TID]int
	merge  bool
}

// confirmPurchase asks how the dockable's purchase is to be made (see confirmPurchaseWithDialog). It is a variable so
// that tests can substitute it.
var confirmPurchase = confirmPurchaseWithDialog

// confirmPurchaseWithDialog puts up the dialog that finalizes the dockable's purchase. It lists the carried money, by
// increasing value, with the quantity of each that the automatic payment in plan spends, which the user may change, or
// go back to, and asks whether to merge what is bought with matching carried items. It returns the terms chosen, with
// ok false when the user canceled.
func confirmPurchaseWithDialog(d *buyDockable, plan *buyPlan) (terms purchaseTerms, ok bool) {
	dlg := newPurchaseDialog(d, plan)
	if dlg == nil || dlg.dialog.RunModal() != unison.ModalResponseOK {
		return purchaseTerms{}, false
	}
	return purchaseTerms{chosen: dlg.chosen(), merge: dlg.merge.State == check.On}, true
}

// paymentQuantities returns the quantity of each money item the payouts spend, by ID.
func paymentQuantities(payouts []gurps.Payout) map[tid.TID]int {
	m := make(map[tid.TID]int, len(payouts))
	for _, one := range payouts {
		if one.Quantity > 0 {
			m[one.Money.TID] = one.Quantity
		}
	}
	return m
}

// purchaseDialog is the dialog confirmPurchaseWithDialog puts up.
type purchaseDialog struct {
	dialog    *unison.Dialog
	entity    *gurps.Entity
	rows      []*purchaseMoneyRow
	owed      fxp.Int
	sellMoney []*gurps.Equipment
	change    purchaseChange
	// auto is the quantity of each money item the automatic payment spends, by ID, which the rows start from and the
	// automatic button puts back.
	auto      map[tid.TID]int
	automatic *unison.Button
	spending  *unison.Label
	changeDue *unison.Label
	note      *purchaseNote
	merge     *unison.CheckBox
}

// purchaseMoneyRow is one carried money item and the quantity of it to spend.
type purchaseMoneyRow struct {
	money    *gurps.Equipment
	value    fxp.Int
	spend    int
	field    *IntegerField
	spending *unison.Label
}

// purchaseChange is the change worked out for an overpayment, kept for as long as the amount and the money paid with,
// and how much of it, stay the same, since the search is costly. The money items are the dialog's own rows, so they can
// be compared as they are.
type purchaseChange struct {
	over    fxp.Int
	given   []gurps.Payout
	payouts []gurps.Payout
	paid    fxp.Int
	valid   bool
}

func newPurchaseDialog(buy *buyDockable, plan *buyPlan) *purchaseDialog {
	entity := buy.sheet.Entity()
	d := &purchaseDialog{
		entity:    entity,
		owed:      plan.total,
		sellMoney: entity.SellMoneyItems(),
		auto:      paymentQuantities(plan.payments),
	}
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
	for _, one := range entity.MoneyItems() {
		d.rows = append(d.rows, &purchaseMoneyRow{money: one, value: one.AdjustedValue()})
	}
	// The money is listed by increasing value, as the sheet settings list the sell money, those of equal value in the
	// order they are carried.
	slices.SortStableFunc(d.rows, func(a, b *purchaseMoneyRow) int { return cmp.Compare(a.value, b.value) })
	for _, row := range d.rows {
		one := row.money
		carried := one.TradeQuantity()
		most := max(carried.Floor().AsInteger[int](), 0)
		row.spend = min(max(d.auto[one.TID], 0), most)
		name := one.String()
		list.rows.AddChild(newPurchaseText(name, align.Start))
		list.rows.AddChild(newPurchaseText(carried.Comma(), align.End))
		list.rows.AddChild(newPurchaseText(i18n.Text("$%s", d.format(row.value)), align.End))
		row.field = NewIntegerField(nil, "", i18n.Text("Spend"),
			func() int { return row.spend },
			func(v int) {
				row.spend = v
				d.refresh()
			}, 0, most, false, false)
		row.field.withoutUndo()
		row.field.SetMarksModified(false)
		row.field.refuseAboveMax()
		row.field.Accessibility.Name = i18n.Text("%s to spend", name)
		row.field.Tooltip = newWrappedTooltip(i18n.Text("The number of %s to spend, up to the %d carried", name, most))
		row.field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle})
		list.rows.AddChild(row.field)
		row.spending = newPurchaseText("", align.End)
		list.rows.AddChild(row.spending)
	}
	if len(d.rows) == 0 {
		label := newPurchaseText(i18n.Text("This character carries no money-tagged items."), align.Start)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: paymentColumnCount})
		list.rows.AddChild(label)
	}
	list.addTotal().SetTitle(i18n.Text("Owed: $%s", d.format(plan.total)))
	d.spending = list.addTotal()
	d.changeDue = list.addTotal()
	d.changeDue.Font = unison.EmphasizedSystemFont

	intro := NewWrappingLabel()
	intro.SetTitle(i18n.Text("Buying %s for $%s.", buy.describeItems(), d.format(plan.total)))
	wrapToWidthOf(intro.AsPanel(), list.box)
	intro.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	intro.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: unison.StdVSpacing}))
	panel.AddChild(intro)
	note := NewWrappingLabel()
	fd := note.Font.Descriptor()
	fd.Size--
	note.Font = fd.Font()
	note.SetTitle(i18n.Text("The money-tagged items carried pay for it, chosen automatically unless you change the quantities to spend."))
	wrapToWidthOf(note.AsPanel(), list.box)
	note.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	panel.AddChild(note)
	panel.AddChild(list.box)
	d.note = newPurchaseNote()
	wrapToWidthOf(d.note.AsPanel(), list.box)
	d.note.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	panel.AddChild(d.note)
	panel.AddChild(d.createOptions(buy.merge))

	var err error
	if d.dialog, err = newPromptDialog(
		promptOperation{name: i18n.Text("Finalize Purchase")}, nil, nil, panel,
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Finalize")),
	); err != nil {
		errs.Log(err)
		return nil
	}
	holdMinSizeOnDisplay(d.dialog.Window())
	// The fields couldn't reach the buttons while they were being created, so the state is brought up to date now.
	d.refresh()
	return d
}

// createOptions returns the row below the note: the choice of whether to merge what is bought with matching carried
// items, which starts out as given, and the button that puts the automatic payment back.
func (d *purchaseDialog) createOptions(merge bool) *unison.Panel {
	options := unison.NewPanel()
	options.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	options.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	d.merge = unison.NewCheckBox()
	d.merge.SetTitle(i18n.Text("Merge with matching items"))
	d.merge.State = check.FromBool(merge)
	d.merge.Accessibility.Name = i18n.Text("Merge with matching items")
	d.merge.Tooltip = newWrappedTooltip(i18n.Text("Add the quantity of each item bought, and of the change, to a carried item that matches it, rather than carrying a separate copy"))
	d.merge.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Start, VAlign: align.Middle, HGrab: true})
	options.AddChild(d.merge)
	d.automatic = unison.NewButton()
	d.automatic.SetTitle(i18n.Text("Choose Automatically"))
	d.automatic.Accessibility.Name = i18n.Text("Choose Automatically")
	d.automatic.Tooltip = newWrappedTooltip(i18n.Text("Go back to the money chosen automatically, which makes the smallest total that covers what is owed"))
	d.automatic.ClickCallback = func() {
		for _, row := range d.rows {
			row.spend = d.auto[row.money.TID]
			// The button doesn't take the focus, so the field that has it is replaced too.
			row.field.syncFromValue()
		}
		d.refresh()
	}
	d.automatic.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle})
	options.AddChild(d.automatic)
	return options
}

// newPurchaseText returns a label showing text, aligned within its cell as given.
func newPurchaseText(text string, hAlign align.Enum) *unison.Label {
	label := unison.NewLabel()
	label.SetTitle(text)
	label.SetLayoutData(&unison.FlexLayoutData{HAlign: hAlign, VAlign: align.Middle})
	return label
}

// refresh brings the totals, the note and the buttons up to date with the quantities chosen.
func (d *purchaseDialog) refresh() {
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
	d.automatic.SetEnabled(!maps.Equal(d.chosen(), d.auto))
	if d.dialog != nil {
		d.dialog.Button(unison.ModalResponseOK).SetEnabled(!short)
		growToFitContent(d.dialog.Window())
	}
}

// describeChange has the note say what the change for an overpayment of over will be, or warn when all of it can't be
// returned.
func (d *purchaseDialog) describeChange(over fxp.Int) {
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
func (d *purchaseDialog) changeFor(over fxp.Int) (change []gurps.Payout, paid fxp.Int) {
	payments := d.payments()
	if !d.change.valid || d.change.over != over || !slices.Equal(d.change.given, payments) {
		payouts, found := d.entity.ChangeFor(over, d.sellMoney, payments)
		d.change = purchaseChange{over: over, given: payments, payouts: payouts, paid: found, valid: true}
	}
	return d.change.payouts, d.change.paid
}

// payments returns the money chosen to spend, leaving out items with none.
func (d *purchaseDialog) payments() []gurps.Payout {
	payments := make([]gurps.Payout, 0, len(d.rows))
	for _, row := range d.rows {
		if row.spend > 0 {
			payments = append(payments, gurps.Payout{Money: row.money, Quantity: row.spend})
		}
	}
	return payments
}

// chosen returns the quantity of each money item to spend, by ID, leaving out those with none.
func (d *purchaseDialog) chosen() map[tid.TID]int {
	return paymentQuantities(d.payments())
}

func (d *purchaseDialog) format(value fxp.Int) string {
	return d.entity.SheetSettings.FormatEquipmentValue(value)
}

// purchaseNote is the note beside the totals: an icon beside wrapping text, in a box of the theme's warning color when
// it warns.
type purchaseNote struct {
	unison.Panel
	icon    *unison.Label
	text    *TextLabel
	warning bool
}

func newPurchaseNote() *purchaseNote {
	n := &purchaseNote{icon: unison.NewLabel(), text: NewWrappingLabel()}
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
func (n *purchaseNote) set(text string, icon *unison.SVG, ink unison.Ink, warn bool) {
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
