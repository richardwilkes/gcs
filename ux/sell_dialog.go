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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/role"
)

const (
	// maxSaleNamesLength is the most characters of item names the suggested transaction name of a sale holds.
	maxSaleNamesLength = 100
	// maxSaleContentsLines is the most items the tooltip beside a container for sale lists before counting the rest.
	maxSaleContentsLines = 30
	// saleContentsIndent is how far each level of what a container for sale holds is indented in its tooltip.
	saleContentsIndent = unison.StdHSpacing * 2
	// saleColumnCount is the number of columns in the list of items for sale: the item, the quantity to sell, the
	// quantity carried, the full price of one and what the quantity sells for.
	saleColumnCount = 5
	// saleNameWidthPrototype is the text the transaction name field is made wide enough to show. The field would
	// otherwise be as wide as the suggested name, which runs to maxSaleNamesLength characters.
	saleNameWidthPrototype = "Sold Boots, Crossbow, Hip Quiver, Leather Gloves"
)

// sellSelectedEquipment asks how much of the equipment selected in the sheet's carried equipment table to sell, and at
// what price, then sells it. It is a variable so that tests can substitute it.
var sellSelectedEquipment = func(sheet *Sheet, table *unison.Table[*Node[*gurps.Equipment]]) {
	showSellDialog(sheet, selectedSaleRows(table))
}

// selectedSaleRows returns the equipment of the table's selected rows that can be sold, leaving out any within a
// selected container, since selling a container sells what it holds, and any that is disabled or held by a disabled
// container, which counts as if it weren't on the sheet.
func selectedSaleRows(table *unison.Table[*Node[*gurps.Equipment]]) []*gurps.Equipment {
	rows := table.SelectedRows(true)
	list := make([]*gurps.Equipment, 0, len(rows))
	for _, row := range rows {
		if eqp := row.Data(); eqp.Enabled() {
			list = append(list, eqp)
		}
	}
	return list
}

type sellDialog struct {
	sheet  *Sheet
	dialog *unison.Dialog
	rows   []*sellRow
	// money holds the denominations the sale pays out in. It is resolved from the sheet's sell money just once, as
	// each lookup may load a library file.
	money   []*gurps.Equipment
	percent fxp.Int
	name    string
	// nameEdited is set once the user changes the transaction name, which until then names the items being sold.
	nameEdited bool
	merge      bool
	// payouts were last computed for total, and paid is their value. PayoutFor can be costly, so they are computed
	// again only when the total changes.
	total        fxp.Int
	paid         fxp.Int
	payouts      []gurps.Payout
	computed     bool
	list         *tradeList
	youGetHeader *unison.Label
	totalLabel   *unison.Label
	// summary holds the payout and, while there is something to warn of, the warning.
	summary     *unison.Panel
	payoutLabel *TextLabel
	warning     *TextLabel
	nameField   *StringField
	// invalid holds the fields whose text isn't a valid value, which keeps the sale from being made (see gateOnValidity).
	invalid map[*DecimalField]bool
}

type sellRow struct {
	eqp      *gurps.Equipment
	most     fxp.Int
	quantity fxp.Int
	youGet   *unison.Label
}

// saleTerms is what the sale of an item rests on: how much of it there is to sell, what each is worth and whether it is
// enabled, since a disabled item isn't sold (see gurps.Entity.CommitSale). Saving an editor can change any of them.
type saleTerms struct {
	most    fxp.Int
	value   fxp.Int
	enabled bool
}

// showSellDialog asks how much of each piece of the sheet's carried equipment to sell, all of it to start with, and at
// what percentage of its full price, then sells it as a single undoable edit, paid in the sheet's sell money. Nothing
// is changed unless the sale is accepted. Where an editor saved as the sale closes it changes what is for sale, the sale
// is shown again as it now stands rather than made on terms the user didn't see (see sell).
func showSellDialog(sheet *Sheet, list []*gurps.Equipment) {
	d := newSellDialog(sheet, list)
	for d.run() {
		if d.sell() || !d.renew() {
			return
		}
	}
}

// newSellDialog returns the dialog for selling the equipment, all of each item to start with, at the sheet's sell
// percentage.
func newSellDialog(sheet *Sheet, list []*gurps.Equipment) *sellDialog {
	entity := sheet.Entity()
	d := &sellDialog{
		sheet:   sheet,
		money:   entity.SellMoneyItems(),
		invalid: make(map[*DecimalField]bool),
		// The field clamps what it is given to its range and reports the result through its setter, which updates
		// labels that don't exist yet, so the percentage is kept within the range before the field sees it.
		percent: min(max(entity.SheetSettings.SellPercent, gurps.MinTradePercent), gurps.MaxTradePercent),
		merge:   true,
	}
	for _, eqp := range list {
		most := eqp.TradeQuantity()
		d.rows = append(d.rows, &sellRow{eqp: eqp, most: most, quantity: most})
	}
	d.name = d.defaultName()
	return d
}

// run shows the dialog, built anew from the rows, and returns true if the sale was accepted.
func (d *sellDialog) run() bool {
	d.dialog = nil // update must not reach the dialog of an earlier run while the content is built.
	content := d.newContent()
	d.update() // Before the dialog is made, so that it is sized for the text.
	var err error
	if d.dialog, err = unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfoWithTitle(i18n.Text("Sell"))},
		unison.FloatingWindowOption()); err != nil {
		errs.Log(err)
		return false
	}
	wnd := d.dialog.Window()
	wnd.SetTitle(i18n.Text("Sell Equipment"))
	holdMinSizeOnDisplay(wnd)
	d.update() // The fields couldn't set the button's state while they were being created, so do it now.
	return d.dialog.RunModal() == unison.ModalResponseOK
}

// sell makes the accepted sale and returns true, or returns true without selling anything when an editor refused to
// close. It returns false, again without selling anything, when closing the editors changed the terms of the sale, which
// is then to be shown again (see renew).
//
// An editor doesn't follow changes to its item, so one left open on an item the sale changes would put its stale
// quantity back when applied. The editors open on everything the sale changes are closed first, as they are for a
// deletion: an item sold in full is removed, along with what it holds, an item sold in part has its quantity reduced,
// as does each container the sale splits apart to do that, and the carried money a payout merges into gains the payout.
// An editor saved at its prompt applies its changes, which may alter what is for sale or what it is worth, while the
// payout was worked out for the sale as accepted, so the sale is only made on the terms it was accepted on.
func (d *sellDialog) sell() bool {
	entity := d.sheet.Entity()
	items := make([]gurps.SaleItem, 0, len(d.rows))
	removed := make(map[tid.TID]bool)
	changed := make(map[tid.TID]bool)
	for _, row := range d.rows {
		if row.quantity > 0 {
			items = append(items, gurps.SaleItem{Equipment: row.eqp, Quantity: row.quantity})
			if row.quantity >= row.most {
				gurps.Traverse(func(eqp *gurps.Equipment) bool {
					removed[eqp.TID] = true
					return false
				}, false, false, row.eqp)
			} else {
				changed[row.eqp.TID] = true
			}
		}
	}
	for _, eqp := range entity.SaleSplits(items) {
		changed[eqp.TID] = true
	}
	if d.merge {
		for _, target := range entity.PayoutMergeTargets(d.payouts, removed) {
			if target != nil {
				changed[target.TID] = true
			}
		}
	}
	maps.Copy(changed, removed)
	terms := d.terms()
	if !CloseID(changed) {
		return true
	}
	if !slices.Equal(terms, d.terms()) {
		return false
	}
	name := strings.TrimSpace(d.name)
	if name == "" {
		name = d.defaultName()
	}
	recordTrade(d.sheet, i18n.Text("Sell Equipment"), func() {
		entity.CommitSale(items, d.payouts, d.merge, name, jio.Now())
	})
	return true
}

// terms returns the terms the sale of each row's item rests on, as they stand now.
func (d *sellDialog) terms() []saleTerms {
	terms := make([]saleTerms, len(d.rows))
	for i, row := range d.rows {
		terms[i] = saleTerms{
			most:    row.eqp.TradeQuantity(),
			value:   row.eqp.ExtendedValueOfJustOne(),
			enabled: row.eqp.Enabled(),
		}
	}
	return terms
}

// renew brings the rows up to date with their items as they now stand, for the dialog to be run again: an item that is
// no longer enabled is dropped, and the quantity of each is kept to what a sale of it now takes (see
// gurps.Equipment.TradeTake). Returns false if nothing is left to sell.
func (d *sellDialog) renew() bool {
	rows := d.rows[:0]
	for _, row := range d.rows {
		if !row.eqp.Enabled() {
			continue
		}
		row.most = row.eqp.TradeQuantity()
		row.quantity = row.eqp.TradeTake(row.quantity)
		rows = append(rows, row)
	}
	d.rows = rows
	clear(d.invalid)
	d.computed = false
	return len(rows) != 0
}

func (d *sellDialog) newContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	percentLabel := NewFieldLeadingLabel(i18n.Text("Percent of full price"), false)
	content.AddChild(percentLabel)
	content.AddChild(d.newPercentField(percentLabel))
	content.AddChild(d.newList())

	d.summary = unison.NewPanel()
	d.summary.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	d.summary.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill})
	content.AddChild(d.summary)
	d.payoutLabel = d.newListWidthLabel()
	d.summary.AddChild(d.payoutLabel)
	d.warning = d.newListWidthLabel() // Added to the summary by update while there is something to warn of.

	content.AddChild(NewFieldLeadingLabel(i18n.Text("Transaction name"), false))
	content.AddChild(d.newNameField())
	content.AddChild(unison.NewPanel()) // Keeps the checkbox in the field column, under the name.
	merge := unison.NewCheckBox()
	merge.SetTitle(i18n.Text("Merge with matching money already carried"))
	merge.State = check.On
	merge.Tooltip = newWrappedTooltip(i18n.Text("Add the money received to the quantity of any carried item that matches it exactly, rather than carrying it as a new item"))
	merge.ClickCallback = func() { d.merge = merge.State == check.On }
	content.AddChild(merge)
	return content
}

// newPercentField returns the percentage field, named by its label, with its unit beside it.
func (d *sellDialog) newPercentField(label *unison.Label) *unison.Panel {
	percent := NewDecimalField(nil, "", "", func() fxp.Int { return d.percent },
		func(v fxp.Int) {
			d.percent = v
			d.update()
		}, gurps.MinTradePercent, gurps.MaxTradePercent, false, false)
	percent.SetMarksModified(false)
	d.gateOnValidity(percent)
	return WrapWithSpan(1, labelControl(percent, label), NewFieldTrailingHint(percent, i18n.Text("%"), false))
}

// newListWidthLabel returns a label, for a line below the list of items for sale, that wraps to the list's width rather
// than widening the dialog.
func (d *sellDialog) newListWidthLabel() *TextLabel {
	label := NewWrappingLabel()
	wrapToWidthOf(label.AsPanel(), d.list.box)
	label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	return label
}

// newNameField returns the transaction name field. The name it suggests is shown in a lighter ink until the user
// edits it, since until then it follows the quantities.
func (d *sellDialog) newNameField() *StringField {
	var ink unison.Ink
	d.nameField = NewStringField(nil, "", "", func() string { return d.name }, func(s string) {
		if s != d.name {
			d.name = s
			if !d.nameEdited {
				d.nameEdited = true
				d.nameField.OnEditableInk = ink
				d.nameField.MarkForRedraw()
			}
		}
	})
	ink = d.nameField.OnEditableInk
	d.nameField.OnEditableInk = &unison.ColorFilteredInk{OriginalInk: ink, ColorFilter: unison.Alpha50Filter()}
	d.nameField.Tooltip = newWrappedTooltip(i18n.Text("The name suggested follows what is being sold until you change it"))
	// The field is made wide enough for the prototype, not for the suggested name, and grows with the dialog.
	d.nameField.SetMinimumTextWidthUsing(saleNameWidthPrototype)
	d.nameField.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		minSize, _, maxSize = d.nameField.DefaultSizes(hint)
		return minSize, minSize, maxSize
	})
	return d.nameField
}

// newList returns the list of items for sale, under a header naming its columns and over a footer giving the total.
// Each row holds the item's name, the quantity to sell and how many are carried, the full price of one and what the
// quantity sells for.
func (d *sellDialog) newList() *unison.Panel {
	d.list = newTradeList(saleColumnCount)
	d.list.box.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.list.addHeader(i18n.Text("Item"), 1, align.Start)
	d.list.addHeader(i18n.Text("Quantity"), 2, align.Start)
	d.list.addHeader(i18n.Text("Full Price Each"), 1, align.End)
	d.youGetHeader = d.list.addHeader("", 1, align.End) // Titled by update, since it names the percentage.
	d.totalLabel = d.list.addTotal()
	list := d.list.rows
	settings := d.sheet.Entity().SheetSettings
	for _, row := range d.rows {
		name := row.eqp.String()
		label := newTruncatedLabel(name, maxContextLineLength, unison.DefaultLabelTheme.Font)
		label.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
		if row.eqp.HasChildren() {
			list.AddChild(withSaleContentsInfo(label, row))
		} else {
			list.AddChild(label)
		}
		// Any item can be left out of the sale by setting its quantity to zero. A container is sold by the copy, so its
		// field refuses fractions unless the quantity carried is itself a fraction. Any other quantity is kept to what
		// the sale takes (see gurps.Equipment.TradeTake): the whole of a group, and only what the copies of containers
		// that can't be split apart give up, so that the payout is for what is sold, and the field is made to show that.
		var quantity *DecimalField
		quantity = NewDecimalField(nil, "", "", func() fxp.Int { return row.quantity },
			func(v fxp.Int) {
				if take := row.eqp.TradeTake(v); take != v {
					row.quantity = take
					quantity.syncFromValue()
					return
				}
				row.quantity = v
				d.update()
			}, 0, row.most, false, false)
		quantity.SetMarksModified(false)
		quantity.refuseAboveMax()
		if row.eqp.Container() && row.most.Floor() == row.most {
			refuseFractions(quantity)
		}
		quantity.Accessibility.Name = i18n.Text("%s quantity to sell", name)
		d.gateOnValidity(quantity)
		list.AddChild(quantity)
		carried := unison.NewLabel()
		carried.SetTitle(i18n.Text("of %s", row.most.Comma()))
		carried.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
		list.AddChild(carried)
		each := unison.NewLabel()
		each.SetTitle(i18n.Text("$%s",
			settings.FormatEquipmentValue(gurps.TradePrice(row.eqp, fxp.One, fxp.Hundred))))
		each.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle})
		list.AddChild(each)
		row.youGet = unison.NewLabel()
		row.youGet.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle})
		list.AddChild(row.youGet)
	}
	return d.list.box
}

// withSaleContentsInfo returns a panel holding the label of a container for sale and, beside it, an information icon
// whose tooltip lists what the container holds, since that is sold along with it (see gurps.Entity.CommitSale).
func withSaleContentsInfo(label *unison.Label, row *sellRow) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	panel.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	panel.AddChild(label)
	info := NewInfoPop()
	info.Accessibility.Name = i18n.Text("What %s holds", row.eqp.String())
	info.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	info.Tooltip = newSaleContentsTooltip(row)
	info.TooltipImmediate = true
	panel.AddChild(info)
	return panel
}

// newSaleContentsTooltip returns a tooltip listing what the row's container holds, each item with its quantity, indented
// beneath the container holding it and struck through when disabled, as the equipment table shows it. A group has no
// quantity of its own. Past maxSaleContentsLines items, the rest are counted instead. Like unison.NewTooltipWithText,
// the tooltip is named with the whole of the text, so the labels say nothing of their own to an assistive technology.
func newSaleContentsTooltip(row *sellRow) *unison.Panel {
	tip := unison.NewTooltipBase()
	tip.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	var lines []string
	addLine := func(text string, depth int, disabled bool) {
		lines = append(lines, text)
		label := unison.NewLabel()
		label.LabelTheme = unison.DefaultTooltipTheme.Label
		label.StrikeThrough = disabled
		label.Accessibility.Role = role.None
		label.SetTitle(text)
		if depth > 0 {
			label.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: float32(depth) * saleContentsIndent}))
		}
		tip.AddChild(label)
	}
	if row.most == fxp.One {
		addLine(i18n.Text("Sold with it:"), 0, false)
	} else {
		addLine(i18n.Text("Sold with each one:"), 0, false)
	}
	listed, more := 0, 0
	var add func(list []*gurps.Equipment, depth int)
	add = func(list []*gurps.Equipment, depth int) {
		for _, child := range list {
			if listed < maxSaleContentsLines {
				text := xstrings.Truncate(child.String(), maxContextLineLength, true)
				if !child.IsGroup() {
					text = i18n.Text("%s ×%s", text, child.Quantity.Comma())
				}
				addLine(text, depth, !child.Enabled())
				listed++
			} else {
				more++
			}
			add(child.Children, depth+1)
		}
	}
	add(row.eqp.Children, 1)
	if more != 0 {
		addLine(i18n.Text("…and %d more", more), 1, false)
	}
	tip.Accessibility.Name = strings.Join(lines, "\n")
	return tip
}

// defaultName returns the transaction name suggested for the sale, naming the items of which something is to be sold, or
// nothing when there are none.
func (d *sellDialog) defaultName() string {
	names := make([]string, 0, len(d.rows))
	for _, row := range d.rows {
		if row.quantity > 0 {
			names = append(names, row.eqp.String())
		}
	}
	if len(names) == 0 {
		return ""
	}
	return i18n.Text("Sold %s", xstrings.Truncate(strings.Join(names, ", "), maxSaleNamesLength, true))
}

// gateOnValidity keeps the sale from being made while the field's text isn't a valid value, since the field then reports
// the nearest value it allows rather than what was typed. The field validates after it has reported a change, so the
// dialog is brought up to date from the validation rather than from the value alone.
func (d *sellDialog) gateOnValidity(field *DecimalField) {
	validate := field.ValidateCallback
	field.ValidateCallback = func() bool {
		valid := validate()
		if d.invalid[field] == valid {
			d.invalid[field] = !valid
			d.update()
		}
		return valid
	}
}

// update brings what the dialog shows, and whether the sale can be made, up to date with the quantities and percentage
// entered. The suggested transaction name follows the quantities until the user edits it.
func (d *sellDialog) update() {
	settings := d.sheet.Entity().SheetSettings
	dollars := func(value fxp.Int) string { return i18n.Text("$%s", settings.FormatEquipmentValue(value)) }
	setLabelTitle(d.youGetHeader, i18n.Text("You Get (%s%%)", d.percent.String()))
	var total fxp.Int
	selling := false
	for _, row := range d.rows {
		price := gurps.TradePrice(row.eqp, row.quantity, d.percent)
		total = fxp.SaturatingAdd(total, price)
		if row.quantity > 0 {
			selling = true
		}
		setLabelTitle(row.youGet, dollars(price))
	}
	if !d.nameEdited {
		if name := d.defaultName(); name != d.name {
			d.name = name
			if d.nameField != nil {
				d.nameField.Sync()
			}
		}
	}
	if !d.computed || total != d.total {
		d.total = total
		d.payouts, d.paid = gurps.PayoutFor(total, d.money)
		d.computed = true
	}
	setLabelTitle(d.totalLabel, i18n.Text("Total: %s", dollars(total)))
	received := i18n.Text("nothing")
	if len(d.payouts) != 0 {
		parts := make([]string, 0, len(d.payouts))
		for _, one := range d.payouts {
			parts = append(parts, i18n.Text("%s ×%d", one.Money.String(), one.Quantity))
		}
		received = strings.Join(parts, ", ")
	}
	if text := i18n.Text("You receive: %s (%s)", received, dollars(d.paid)); text != d.payoutLabel.Title() {
		d.payoutLabel.SetTitle(text)
	}
	var warning string
	switch {
	case len(settings.SellMoney) == 0:
		warning = i18n.Text("No sell money is configured. Add money items to the Trading section of this sheet's settings.")
	case len(d.money) == 0:
		warning = i18n.Text("None of the configured sell money could be found in the libraries.")
	case d.paid < total:
		warning = i18n.Text("%s can't be paid with the configured sell money. Receiving %s, the closest amount not over the total.",
			dollars(total-d.paid), dollars(d.paid))
	}
	d.showWarning(warning)
	if d.dialog != nil {
		d.dialog.Button(unison.ModalResponseOK).SetEnabled(selling && len(d.money) != 0 && !d.anyInvalid())
		growToFitContent(d.dialog.Window())
	}
}

// showWarning puts the warning under the payout, or takes it away when there is nothing to warn of, so that it takes
// no room then.
func (d *sellDialog) showWarning(text string) {
	if text != d.warning.Title() {
		d.warning.setText(text, unison.ThemeWarning)
	}
	shown := d.warning.Parent() != nil
	switch {
	case text == "" && shown:
		d.warning.RemoveFromParent()
	case text != "" && !shown:
		d.summary.AddChild(d.warning)
	default:
		return
	}
	d.summary.MarkForLayoutRecursivelyUpward()
	d.summary.MarkForRedraw()
}

// anyInvalid returns true if any field's text isn't a valid value.
func (d *sellDialog) anyInvalid() bool {
	for _, invalid := range d.invalid {
		if invalid {
			return true
		}
	}
	return false
}
