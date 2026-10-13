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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// maxSaleNamesLength is the most characters of item names the suggested transaction name of a sale holds.
const maxSaleNamesLength = 100

// sellSelectedEquipment asks how much of the equipment selected in the sheet's carried equipment table to sell, and at
// what price, then sells it. It is a variable so that tests can substitute it.
var sellSelectedEquipment = func(sheet *Sheet, table *unison.Table[*Node[*gurps.Equipment]]) {
	showSellDialog(sheet, selectedSaleRows(table))
}

// selectedSaleRows returns the equipment of the table's selected rows, leaving out any within a selected container,
// since selling a container sells what it holds.
func selectedSaleRows(table *unison.Table[*Node[*gurps.Equipment]]) []*gurps.Equipment {
	rows := table.SelectedRows(true)
	list := make([]*gurps.Equipment, 0, len(rows))
	for _, row := range rows {
		list = append(list, row.Data())
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
	merge   bool
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
		percent: entity.SheetSettings.SellPercent,
		merge:   true,
	}
	names := make([]string, 0, len(list))
	for _, eqp := range list {
		most := eqp.TradeQuantity()
		d.rows = append(d.rows, &sellRow{eqp: eqp, most: most, quantity: most})
		names = append(names, eqp.String())
	}
	d.name = i18n.Text("Sold %s", xstrings.Truncate(strings.Join(names, ", "), maxSaleNamesLength, true))
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
	for _, row := range d.rows {
		if row.quantity > 0 {
			items = append(items, gurps.SaleItem{Equipment: row.eqp, Quantity: row.quantity})
		}
	}
	name := strings.TrimSpace(d.name)
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
		}, fxp.One, fxp.Thousand, false, false)
	percent.SetMarksModified(false)
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
	content.AddChild(NewStringField(nil, "", "", func() string { return d.name }, func(s string) { d.name = s }))
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
		list.AddChild(label)
		quantity := NewDecimalField(nil, "", "", func() fxp.Int { return row.quantity },
			func(v fxp.Int) {
				row.quantity = v
				d.update()
			}, 0, row.most, false, false)
		quantity.SetMarksModified(false)
		quantity.refuseAboveMax()
		quantity.Accessibility.Name = i18n.Text("%s quantity to sell", name)
		quantity.SetEnabled(row.most > fxp.One)
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

// update brings what the dialog shows, and whether the sale can be made, up to date with the quantities and percentage
// entered.
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
		d.dialog.Button(unison.ModalResponseOK).SetEnabled(selling && len(d.money) != 0)
		growToFitContent(d.dialog.Window())
	}
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
