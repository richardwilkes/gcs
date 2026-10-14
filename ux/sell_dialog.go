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

	"github.com/richardwilkes/gcs/v5/model/colors"
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
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
)

const (
	// maxSaleNamesLength is the most characters of item names the suggested transaction name of a sale holds.
	maxSaleNamesLength = 100
	// maxSaleContentsLines is the most items the tooltip beside a container for sale lists before counting the rest.
	maxSaleContentsLines = 30
	// saleContentsIndent is how far each level of what a container for sale holds is indented in its tooltip.
	saleContentsIndent = unison.StdHSpacing * 2
	// maxSaleRowsShown is the most items the list of items for sale shows at once, beyond which it scrolls.
	maxSaleRowsShown = 10
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
	total    fxp.Int
	paid     fxp.Int
	payouts  []gurps.Payout
	computed bool
	// box holds the list of items for sale, under a header naming its columns and over a footer giving the total.
	box          *unison.Panel
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
	percentLabel := NewFieldLeadingLabel(i18n.Text("Percent of full price"), false)
	content.AddChild(percentLabel)
	content.AddChild(d.newPercentField(percentLabel))
	content.AddChild(d.newListBox())

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

// newListWidthLabel returns a label that wraps to the width of the list of items for sale rather than widening the
// dialog, for the lines below the list.
func (d *sellDialog) newListWidthLabel() *TextLabel {
	label := NewWrappingLabel()
	label.SetSizer(func(size geom.Size) (minSize, prefSize, maxSize geom.Size) {
		if size.Width <= 0 {
			_, pref, _ := d.box.Sizes(geom.Size{})
			size.Width = pref.Width
		}
		return label.sizes(size)
	})
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

// newListBox returns the box holding the list of items for sale: a header naming the columns, the rows, which scroll
// once there are more than maxSaleRowsShown of them, and a footer giving the total. Each is a panel of its own, laid
// out in the columns they share.
func (d *sellDialog) newListBox() *unison.Panel {
	d.box = unison.NewPanel()
	d.box.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	d.box.SetLayout(&unison.FlexLayout{Columns: 1})
	d.box.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	columns := &saleColumns{}
	header := d.newListHeader(columns)
	layout := &saleColumnsLayout{columns: columns}
	list := d.newList(layout)
	footer := d.newListFooter(columns)
	columns.panels = []*unison.Panel{header, list, footer}
	scroll := unison.NewScrollPanel()
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayout(&sellListLayout{Layout: scroll, list: list, rows: layout})
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.box.AddChild(header)
	d.box.AddChild(scroll)
	d.box.AddChild(footer)
	return d.box
}

// newSaleColumnsPanel returns a panel laid out by layout in the list's columns, inset from its sides as the header, the
// rows and the footer all are, so that their columns line up, and from its top and bottom by vertical. The right inset
// leaves room for the scroll bar over the rows, so that it never covers the last column.
func newSaleColumnsPanel(layout *saleColumnsLayout, vertical float32) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(layout)
	panel.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Top:    vertical,
		Left:   unison.StdHSpacing,
		Bottom: vertical,
		Right:  max(unison.StdHSpacing, unison.DefaultScrollBarTheme.MinimumThickness),
	}))
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return panel
}

// newListHeader returns the row of column headers, drawn as the sheet's tables draw theirs.
func (d *sellDialog) newListHeader(columns *saleColumns) *unison.Panel {
	header := newSaleColumnsPanel(&saleColumnsLayout{columns: columns}, unison.StdVSpacing)
	header.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, colors.Header.Paint(gc, rect, paintstyle.Fill))
	}
	add := func(title string, span int, hAlign align.Enum) *unison.Label {
		label := unison.NewLabel()
		label.OnBackgroundInk = colors.OnHeader
		label.SetTitle(title)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: span, HAlign: hAlign, VAlign: align.Middle})
		header.AddChild(label)
		return label
	}
	add(i18n.Text("Item"), 1, align.Start)
	add(i18n.Text("Quantity"), 2, align.Start)
	add(i18n.Text("Full Price Each"), 1, align.End)
	d.youGetHeader = add("", 1, align.End) // Titled by update, since it names the percentage.
	return header
}

// newListFooter returns the row under the list giving the total, right-aligned under what each item sells for.
func (d *sellDialog) newListFooter(columns *saleColumns) *unison.Panel {
	footer := newSaleColumnsPanel(&saleColumnsLayout{columns: columns}, unison.StdVSpacing)
	footer.SetBorder(unison.NewCompoundBorder(
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Top: 1}, false), footer.Border(),
	))
	d.totalLabel = unison.NewLabel()
	d.totalLabel.SetLayoutData(&unison.FlexLayoutData{HSpan: saleColumnCount, HAlign: align.End, VAlign: align.Middle})
	footer.AddChild(d.totalLabel)
	return footer
}

// newList returns the panel listing the items for sale, a row each, banded alternately. Each row holds the item's name,
// the quantity to sell and how many are carried, the full price of one and what the quantity sells for.
func (d *sellDialog) newList(layout *saleColumnsLayout) *unison.Panel {
	list := newSaleColumnsPanel(layout, unison.StdVSpacing/2)
	list.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		for i, band := range layout.bands {
			if i&1 == 1 {
				gc.DrawRect(band, unison.ThemeBanding.Paint(gc, band, paintstyle.Fill))
			}
		}
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
		// Any item can be left out of the sale by setting its quantity to zero. A container is sold by the copy, so its
		// field refuses fractions unless the quantity carried is itself a fraction, and a group is sold in full (see
		// gurps.Entity.CommitSale), so any other quantity of one is the whole of it, and the field is made to show that.
		var quantity *DecimalField
		quantity = NewDecimalField(nil, "", "", func() fxp.Int { return row.quantity },
			func(v fxp.Int) {
				if row.eqp.IsGroup() && v > 0 && v != row.most {
					row.quantity = row.most
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
	setSellLabelTitle(d.youGetHeader, i18n.Text("You Get (%s%%)", d.percent.String()))
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
	setSellLabelTitle(d.totalLabel, i18n.Text("Total: %s", dollars(total)))
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

// sellListLayout sizes the scroll panel holding the rows of items for sale to show them all, up to maxSaleRowsShown
// of them, beyond which it scrolls.
type sellListLayout struct {
	unison.Layout
	list *unison.Panel
	rows *saleColumnsLayout
}

func (l *sellListLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	_, prefSize, maxSize = l.Layout.LayoutSizes(target, hint)
	most := l.rows.heightShowing(l.list, maxSaleRowsShown)
	if border := target.Border(); border != nil {
		most += border.Insets().Height()
	}
	prefSize.Height = min(prefSize.Height, most)
	maxSize = geom.NewSize(max(maxSize.Width, prefSize.Width), max(maxSize.Height, prefSize.Height))
	return prefSize, prefSize, maxSize
}
