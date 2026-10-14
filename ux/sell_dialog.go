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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fonts"
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
	"github.com/richardwilkes/unison/enums/behavior"
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
	total       fxp.Int
	paid        fxp.Int
	payouts     []gurps.Payout
	computed    bool
	scroll      *unison.ScrollPanel
	totalLabel  *unison.Label
	payoutLabel *unison.Label
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

// showSellDialog asks how much of each piece of the sheet's carried equipment to sell, all of it to start with, and at
// what percentage of its full price, then sells it as a single undoable edit, paid in the sheet's sell money. Nothing
// is changed unless the sale is accepted.
func showSellDialog(sheet *Sheet, list []*gurps.Equipment) {
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
	content := d.newContent()
	d.update() // Before the dialog is made, so that it is sized for the text.
	var err error
	if d.dialog, err = unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfoWithTitle(i18n.Text("Sell"))},
		unison.FloatingWindowOption()); err != nil {
		errs.Log(err)
		return
	}
	wnd := d.dialog.Window()
	wnd.SetTitle(i18n.Text("Sell Equipment"))
	holdMinSizeOnDisplay(wnd)
	d.update() // The fields couldn't set the button's state while they were being created, so do it now.
	if d.dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	items := make([]gurps.SaleItem, 0, len(d.rows))
	// An editor doesn't follow changes to its item, so one left open on an item the sale changes would put its stale
	// quantity back when applied. The editors open on everything the sale changes are closed first, as they are for a
	// deletion, and the sale is abandoned if one can't be closed: an item sold in full is removed, along with what it
	// holds, an item sold in part has its quantity reduced, as does each container the sale splits apart to do that,
	// and the carried money a payout merges into gains the payout.
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
		for _, payout := range d.payouts {
			if target := entity.MergeTarget(payout.Money, removed); target != nil {
				changed[target.TID] = true
			}
		}
	}
	maps.Copy(changed, removed)
	if !CloseID(changed) {
		return
	}
	name := strings.TrimSpace(d.name)
	if name == "" {
		name = d.defaultName()
	}
	recordTrade(sheet, i18n.Text("Sell Equipment"), func() {
		entity.CommitSale(items, d.payouts, d.merge, name, jio.Now())
	})
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
	content.AddChild(NewFieldLeadingLabel(i18n.Text("Percent of full price"), false))
	percent := NewDecimalField(nil, "", "", func() fxp.Int { return d.percent },
		func(v fxp.Int) {
			d.percent = v
			d.update()
		}, gurps.MinTradePercent, gurps.MaxTradePercent, false, false)
	percent.SetMarksModified(false)
	d.gateOnValidity(percent)
	content.AddChild(percent)

	d.scroll = unison.NewScrollPanel()
	d.scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	d.scroll.SetContent(d.newList(), behavior.Fill, behavior.Fill)
	d.scroll.BackgroundInk = unison.ThemeSurface
	d.scroll.SetLayout(&sellListLayout{Layout: d.scroll})
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	content.AddChild(d.scroll)

	d.totalLabel = unison.NewLabel()
	d.totalLabel.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.End})
	content.AddChild(d.totalLabel)
	d.payoutLabel = unison.NewLabel()
	d.payoutLabel.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.payoutLabel)
	d.warning = NewWrappingLabel()
	// The warning wraps to the list's width rather than widening the dialog.
	d.warning.SetSizer(func(size geom.Size) (minSize, prefSize, maxSize geom.Size) {
		if size.Width <= 0 {
			_, pref, _ := d.scroll.Sizes(geom.Size{})
			size.Width = pref.Width
		}
		return d.warning.sizes(size)
	})
	d.warning.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill})
	content.AddChild(d.warning)

	content.AddChild(NewFieldLeadingLabel(i18n.Text("Transaction name"), false))
	d.nameField = NewStringField(nil, "", "", func() string { return d.name }, func(s string) {
		if s != d.name {
			d.name = s
			d.nameEdited = true
		}
	})
	content.AddChild(d.nameField)
	merge := unison.NewCheckBox()
	merge.SetTitle(i18n.Text("Merge with matching money already carried"))
	merge.State = check.On
	merge.Tooltip = newWrappedTooltip(i18n.Text("Add the money received to the quantity of any carried item that matches it exactly, rather than carrying it as a new item"))
	merge.ClickCallback = func() { d.merge = merge.State == check.On }
	merge.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(merge)
	return content
}

// newList returns the panel listing the items for sale under a row of column headers. Each row holds the item's name,
// the quantity to sell and how many are carried, the full price of one and what the quantity sells for.
func (d *sellDialog) newList() *unison.Panel {
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	for i, header := range []string{
		i18n.Text("Item"),
		i18n.Text("Quantity"),
		i18n.Text("Full Price Each"),
		i18n.Text("You Get"),
	} {
		label := unison.NewLabel()
		label.Font = fonts.FieldSecondary
		label.SetTitle(header)
		data := &unison.FlexLayoutData{HAlign: align.End}
		switch i {
		case 0:
			data.HAlign = align.Start
		case 1:
			data.HSpan = 2
			data.HAlign = align.Start
		}
		label.SetLayoutData(data)
		list.AddChild(label)
	}
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
		quantity := NewDecimalField(nil, "", "", func() fxp.Int { return row.quantity },
			func(v fxp.Int) {
				row.quantity = v
				d.update()
			}, 0, row.most, false, false)
		quantity.SetMarksModified(false)
		quantity.refuseAboveMax()
		quantity.Accessibility.Name = i18n.Text("%s quantity to sell", name)
		d.gateOnValidity(quantity)
		// A group is sold in full (see gurps.Entity.CommitSale).
		quantity.SetEnabled(!row.eqp.IsGroup() && row.most > fxp.One)
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
	return list
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
	var total fxp.Int
	selling := false
	for _, row := range d.rows {
		price := gurps.TradePrice(row.eqp, row.quantity, d.percent)
		total += price
		if row.quantity > 0 {
			selling = true
		}
		setSellLabelTitle(row.youGet, dollars(price))
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
	setSellLabelTitle(d.totalLabel, i18n.Text("Target total: %s", dollars(total)))
	received := i18n.Text("nothing")
	if len(d.payouts) != 0 {
		parts := make([]string, 0, len(d.payouts))
		for _, one := range d.payouts {
			parts = append(parts, i18n.Text("%s ×%d", one.Money.String(), one.Quantity))
		}
		received = strings.Join(parts, ", ")
	}
	setSellLabelTitle(d.payoutLabel, i18n.Text("You receive: %s (%s)", received, dollars(d.paid)))
	var warning string
	switch {
	case len(settings.SellMoney) == 0:
		warning = i18n.Text("No sell money is configured. Add money items to the Trading section of this sheet's settings.")
	case len(d.money) == 0:
		warning = i18n.Text("None of the configured sell money could be found in the libraries.")
	case d.paid < total:
		warning = i18n.Text("%s can't be paid with the configured sell money. Receiving %s, the closest amount not over the target.",
			dollars(total-d.paid), dollars(d.paid))
	}
	if warning != d.warning.Title() {
		d.warning.setText(warning, unison.ThemeWarning)
	}
	if d.dialog != nil {
		d.dialog.Button(unison.ModalResponseOK).SetEnabled(selling && len(d.money) != 0 && !d.anyInvalid())
		growToFitContent(d.dialog.Window())
	}
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

// setSellLabelTitle replaces the label's text, if it differs, and has it laid out again for its new width.
func setSellLabelTitle(label *unison.Label, text string) {
	if label.String() != text {
		label.SetTitle(text)
		label.MarkForLayoutRecursivelyUpward()
		label.MarkForRedraw()
	}
}

// growToFitContent enlarges the window, should its content now want more room in either direction than it has, as far
// as its display allows.
func growToFitContent(wnd *unison.Window) {
	_, pref, _ := wnd.Content().Sizes(geom.Size{})
	r := wnd.ContentRect()
	if display := wnd.Display(); display != nil {
		frame := wnd.FrameRect()
		pref.Width = min(pref.Width, display.Usable.Width-(frame.Width-r.Width))
		pref.Height = min(pref.Height, display.Usable.Height-(frame.Height-r.Height))
	}
	if pref.Width > r.Width || pref.Height > r.Height {
		r.Width = max(r.Width, pref.Width)
		r.Height = max(r.Height, pref.Height)
		wnd.SetContentRect(r)
		wnd.EnsureOnDisplay()
	}
}

// sellListLayout sizes the scroll panel holding the items for sale to show them all, up to the height of a list of
// options (see listMinSize), beyond which it scrolls.
type sellListLayout struct {
	unison.Layout
}

func (l *sellListLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	_, prefSize, maxSize = l.Layout.LayoutSizes(target, hint)
	most := listMinSize(0).Height
	if border := target.Border(); border != nil {
		most += border.Insets().Height()
	}
	prefSize.Height = min(prefSize.Height, most)
	maxSize = geom.NewSize(max(maxSize.Width, prefSize.Width), max(maxSize.Height, prefSize.Height))
	return prefSize, prefSize, maxSize
}
