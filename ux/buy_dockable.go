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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	// maxBuyQuantity is the most of one item the buy dockable buys at once.
	maxBuyQuantity = 9999
	// buyRowColumns is the number of columns of a row of items to buy: quantity, name, price of each, total and remove.
	buyRowColumns = 5
	// buyHintRows is the number of item rows the dockable asks room for however few it lists (see sizeHint).
	buyHintRows = 12
)

var (
	_ unison.Dockable            = &buyDockable{}
	_ unison.TabCloser           = &buyDockable{}
	_ GroupedCloser              = &buyDockable{}
	_ SheetSourceUser            = &buyDockable{}
	_ unison.UndoManagerProvider = &buyDockable{}
	_ Syncer                     = &buyDockable{}
)

// buyDockable is the dockable in which a sheet's character buys equipment (see showBuyDockable). It lists the equipment
// dropped on it, or sent to it with the Buy command (see buySelectedEquipment), from any equipment list, the sheet's
// own included, with the quantity of each to buy, and warns when the carried money (see gurps.Entity.MoneyItems) doesn't
// cover the total. The money paid with, and whether what is bought merges with matching carried items, are settled when
// the purchase is finalized (see confirmPurchase). Nothing on the sheet changes until then: the dockable's fields
// neither record undo nor mark the sheet as modified. The sheet stays editable meanwhile, so the payment is worked out
// again from the money carried on every change to the sheet (see SheetChanged) and checked once more as the purchase is
// finalized.
type buyDockable struct {
	unison.Panel
	sheet        *Sheet
	targetMgr    *TargetMgr
	scroll       *unison.ScrollPanel
	dropArea     *unison.Panel
	hint         *TextLabel
	header       *unison.Panel
	footer       *unison.Panel
	percentField *DecimalField
	summary      *TextLabel
	setSummary   func(text string, warn bool)
	nameField    *StringField
	finalize     *unison.Button
	items        []*buyItem
	percent      fxp.Int
	name         string
	title        string
	// merge is whether what is bought is to merge with matching carried items, as last settled when finalizing (see
	// confirmPurchase), which the next attempt starts from.
	merge  bool
	auto   buyAutoPayment
	change buyChange
	plan   buyPlan
	// percentEdited and nameEdited are set once the user changes the percentage or the transaction name, which until
	// then follow the sheet's buy percentage and the items being bought.
	percentEdited bool
	nameEdited    bool
	dragOver      bool
	// settled is set once the purchase has been made or discarded, so that closing the dockable no longer asks about it
	// and the rebuild reporting a purchase no longer syncs the dockable.
	settled bool
	// modified is whether the dockable last reported a purchase under way (see Modified), so that its tab is brought up
	// to date when that changes.
	modified bool
	// modifierTarget is the item that the equipment modifiers being dragged would be attached to, while they are over
	// it.
	modifierTarget *buyItem
}

// buyItem is a row of the buy dockable: a copy of the equipment to buy, owned by the sheet's entity but in none of its
// lists, whose Quantity is the quantity to buy, along with the widgets that show it.
type buyItem struct {
	eqp    *gurps.Equipment
	field  *DecimalField
	name   *unison.Label
	each   *unison.Label
	total  *unison.Label
	remove *unison.Button
	// modifierRows holds the widgets of the row shown under the item for each of its modifiers, which are built anew
	// whenever the rows are laid out (see buyDockable.layoutRows), so that they always match the modifiers the item has.
	modifierRows []buyModifierRow
}

// buyModifierRow is the row shown under an item of the buy dockable for one of its modifiers: its description and the
// button that takes it back off the item.
type buyModifierRow struct {
	modifier *gurps.EquipmentModifier
	widgets  []unison.Paneler
}

// buyPlan is how the purchase stands: its total price, the carried money that pays for it and the change returned.
type buyPlan struct {
	total      fxp.Int
	paid       fxp.Int
	changePaid fxp.Int
	payments   []gurps.Payout
	change     []gurps.Payout
	// sellMoney holds the denominations of the sheet's sell money found in the libraries (see
	// gurps.Entity.SellMoneyItems), which the change is returned in.
	sellMoney []*gurps.Equipment
	// short is set when the money paid doesn't cover the total.
	short bool
}

// buyMoneyState is what the automatic payment depends on for one carried money item.
type buyMoneyState struct {
	id       tid.TID
	value    fxp.Int
	quantity fxp.Int
}

// buyMoneySpend is a quantity of a carried money item to pay with.
type buyMoneySpend struct {
	id       tid.TID
	quantity int
}

// buyAutoPayment remembers the last automatic payment worked out, since the dockable is synced with every change to
// the sheet and gurps.PaymentFor builds a table as large as the total to do it.
type buyAutoPayment struct {
	money []buyMoneyState
	spend []buyMoneySpend
	total fxp.Int
	paid  fxp.Int
	valid bool
}

// buyChange remembers the last change worked out, for the same reason as buyAutoPayment. The money paid with can be
// drawn on for change, so given is its state and how much of it was paid, and sellMoney is the state of the sell money
// denominations it was worked out from.
type buyChange struct {
	given     []buyGivenMoney
	sellMoney []buySellMoneyState
	change    []gurps.Payout
	target    fxp.Int
	paid      fxp.Int
	valid     bool
}

// buySellMoneyState is what the change depends on for one sell money denomination: the source it was found through,
// since each lookup makes a fresh copy, and the copy itself, which the change is made of and which the next lookup's
// copy must match in everything gurps.Equipment.MergeMatches compares, as the library may have changed in between.
type buySellMoneyState struct {
	source gurps.Source
	money  *gurps.Equipment
}

// newBuySellMoneyState returns the state of the sell money denomination.
func newBuySellMoneyState(money *gurps.Equipment) buySellMoneyState {
	return buySellMoneyState{source: money.Source, money: money}
}

// matches returns true if the denomination is as it stood when the change was worked out.
func (s buySellMoneyState) matches(money *gurps.Equipment) bool {
	return s.source == money.Source && s.money.MergeMatches(money)
}

// buyGivenMoney is what the change depends on for one carried money item paid with: a copy of it as it stood, since
// the change may be made up of copies of it and depends on everything about it that gurps.Equipment.MergeMatches
// compares, along with its value and the quantity paid, which is the most of it the change may return.
type buyGivenMoney struct {
	id       tid.TID
	value    fxp.Int
	money    *gurps.Equipment
	quantity int
}

// newBuyGivenMoney returns the state of the money paid with.
func newBuyGivenMoney(payment gurps.Payout) buyGivenMoney {
	return buyGivenMoney{
		id:       payment.Money.TID,
		value:    payment.Money.AdjustedValue(),
		money:    payment.Money.Clone(gurps.LibraryFile{}, nil, nil, gurps.Duplicate),
		quantity: payment.Quantity,
	}
}

// matches returns true if the payment is as it stood when the change was worked out.
func (g buyGivenMoney) matches(payment gurps.Payout) bool {
	return g.id == payment.Money.TID && g.quantity == payment.Quantity && g.value == payment.Money.AdjustedValue() &&
		g.money.MergeMatches(payment.Money)
}

// showBuyDockable brings up the dockable in which the sheet's character buys equipment, opening it if the sheet has
// none yet, and returns it.
func showBuyDockable(sheet *Sheet) *buyDockable {
	if d := buyDockableFor(sheet); d != nil {
		ActivateDockable(d)
		return d
	}
	d := newBuyDockable(sheet)
	PlaceInDock(d, dgroup.Editors, false)
	d.percentField.RequestFocus()
	return d
}

// buyDockableFor returns the open dockable in which the sheet's character buys equipment, or nil if there is none.
func buyDockableFor(sheet *Sheet) *buyDockable {
	for _, one := range AllDockables() {
		if d, ok := one.AsPanel().Self.(*buyDockable); ok && d.sheet == sheet {
			return d
		}
	}
	return nil
}

// installBuyEquipmentHandler has the panel buy the equipment selected in its table for a character sheet (see
// buySelectedEquipment).
func installBuyEquipmentHandler(panel unison.Paneler, table *unison.Table[*Node[*gurps.Equipment]]) {
	panel.AsPanel().InstallCmdHandlers(BuySelectedEquipmentItemID,
		func(_ any) bool { return table.HasSelection() && len(OpenSheets(nil)) != 0 },
		func(_ any) { buySelectedEquipment(table) })
}

// buySelectedEquipment buys the equipment selected in the table for a character sheet: the one the table belongs to,
// or else the one the user picks of those open, which is asked only when there is more than one. The equipment joins
// the sheet's buy dockable, brought up for it, as it would were it dropped there.
func buySelectedEquipment(table *unison.Table[*Node[*gurps.Equipment]]) {
	rows := table.SelectedRows(true)
	if len(rows) == 0 {
		return
	}
	list := make([]*gurps.Equipment, 0, len(rows))
	for _, row := range rows {
		list = append(list, row.Data())
	}
	sheet := table.Ancestor[*Sheet]()
	if sheet == nil {
		var ok bool
		if sheet, ok = promptForSingleDestination(promptOperation{
			name:        i18n.Text("Buy"),
			description: i18n.Text("Buying %s", shortNames(describeRows(list))...),
		}, OpenSheets(nil)); !ok {
			return
		}
	}
	showBuyDockable(sheet).add(table, list)
}

func newBuyDockable(sheet *Sheet) *buyDockable {
	d := &buyDockable{sheet: sheet, merge: true}
	d.Self = d
	d.targetMgr = NewTargetMgr(d)
	d.SetLayout(&unison.FlexLayout{Columns: 1})
	d.Accessibility.Name = i18n.Text("Buy Equipment")
	d.percent = d.sheetPercent()
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 2 * unison.StdVSpacing,
	})
	insets := unison.StdInsets()
	insets.Top *= 2
	insets.Bottom *= 2
	content.SetBorder(unison.NewEmptyBorder(insets))
	content.AddChild(d.createHeader())
	content.AddChild(d.createDropArea())
	content.AddChild(d.createFooter())
	d.scroll = unison.NewScrollPanel()
	d.scroll.SetContent(content, behavior.HintedFill, behavior.Fill)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: d.sizeHint(),
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		HGrab:    true,
		VGrab:    true,
	})
	d.AddChild(d.scroll)
	d.installDropSupport()
	d.Sync()
	return d
}

// sizeHint returns the room the dockable asks for however little it lists: rows of typical width for buyHintRows items,
// or the header or footer row shown on one line if that is wider, between the header and the footer. A window packed
// around the dockable would otherwise close up around the hint and those rows, since an empty drop area asks for no
// room of its own.
func (d *buyDockable) sizeHint() geom.Size {
	font := unison.DefaultLabelTheme.Font
	measure := func(samples ...string) float32 {
		var width float32
		for _, sample := range samples {
			width += unison.NewText(sample, &unison.TextDecoration{Font: font}).Width() + 2*unison.StdHSpacing
		}
		return width
	}
	rowHeight := font.LineHeight() + unison.StdVSpacing
	// The remove button takes a row's height; the drop area's and the content's borders flank the row.
	row := measure("9999", "Leather Armor (Fine, Ornate, Waterproofed)", i18n.Text("$%s each", "1,234.56"),
		i18n.Text("Total $%s", "12,345.67")) + rowHeight + 6*unison.StdHSpacing
	header := measure(i18n.Text("Transaction"), "Bought Leather Armor, Longbow, Rope", i18n.Text("Percent of Full Price"),
		"100")
	footer := measure(i18n.Text("Not enough money carried: $%s needed, $%s carried.", "12,345.67", "1,234.56"),
		i18n.Text("Finalize Purchase…"))
	_, headerSize, _ := d.header.Sizes(geom.Size{})
	_, footerSize, _ := d.footer.Sizes(geom.Size{})
	// The hint takes a row of its own above the items.
	return geom.NewSize(xmath.Ceil(max(row, header, footer)),
		xmath.Ceil(rowHeight*(buyHintRows+1)+headerSize.Height+footerSize.Height))
}

func (d *buyDockable) createDropArea() *unison.Panel {
	d.dropArea = unison.NewPanel()
	d.dropArea.SetLayout(&unison.FlexLayout{
		Columns:  buyRowColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.dropArea.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.dropArea.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(2 * unison.StdHSpacing)))
	d.dropArea.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		// The rows sit on the background a table gives its rows, so that the banding stands out as it does there.
		r := d.dropArea.ContentRect(true)
		gc.DrawRect(r, unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
		for i, item := range d.items {
			if i&1 == 1 {
				band := d.itemBounds(item)
				gc.DrawRect(band, unison.ThemeBanding.Paint(gc, band, paintstyle.Fill))
			}
		}
		var ink unison.Ink = unison.ThemeSurfaceEdge
		if d.dragOver {
			// The drop target is tinted as the lists tint theirs (see InstallTableDropSupport).
			ink = unison.ThemeWarning
			paint := ink.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
		r = r.Inset(geom.NewUniformInsets(0.5))
		paint := ink.Paint(gc, r, paintstyle.Stroke)
		paint.SetPathEffect(unison.DashEffect())
		gc.DrawRect(r, paint)
	}
	hint := NewWrappingLabel()
	fd := hint.Font.Descriptor()
	fd.Size--
	hint.Font = fd.Font()
	hint.setText(i18n.Text("Drop equipment here from any equipment list, or choose Buy from an equipment list's menu. Drop equipment modifiers onto an item to adjust it."),
		unison.ThemeOnBelowSurface)
	hint.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  buyRowColumns,
		HAlign: align.Fill,
		HGrab:  true,
	})
	hint.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: 2 * unison.StdVSpacing}))
	d.dropArea.AddChild(hint)
	d.hint = hint
	d.dropArea.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
		// The item the modifiers being dragged would go to is highlighted, as the rows of a list are (see
		// InstallTableDropSupport).
		if d.modifierTarget != nil {
			r := d.itemBounds(d.modifierTarget)
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}
	return d.dropArea
}

// createHeader returns the row above the drop area: the transaction name, which the shopping record gives the purchase,
// with the percentage of full price to pay at the right.
func (d *buyDockable) createHeader() *unison.Panel {
	d.header = newBuyRow(4)
	label := NewFieldLeadingLabel(i18n.Text("Transaction"), false)
	d.header.AddChild(label)
	d.nameField = NewStringField(d.targetMgr, "buy.name", i18n.Text("Transaction Name"),
		func() string { return d.name },
		func(s string) {
			if s != d.name {
				d.name = s
				d.nameEdited = true
			}
		})
	d.nameField.withoutUndo()
	d.nameField.SetMarksModified(false)
	d.nameField.Accessibility.Name = i18n.Text("Transaction name")
	d.nameField.Accessibility.LabeledBy = label
	d.nameField.Tooltip = newWrappedTooltip(i18n.Text("The name the shopping record gives the purchase"))
	d.nameField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	d.header.AddChild(d.nameField)
	label = NewFieldInteriorLeadingLabel(i18n.Text("Percent of Full Price"), false)
	d.header.AddChild(label)
	d.percentField = NewDecimalField(d.targetMgr, "buy.percent", i18n.Text("Percent of Full Price"),
		func() fxp.Int { return d.percent },
		func(v fxp.Int) {
			if v != d.percent {
				d.percent = v
				d.percentEdited = true
				d.refresh()
			}
		}, gurps.MinTradePercent, gurps.MaxTradePercent, false, false)
	d.percentField.withoutUndo()
	d.percentField.SetMarksModified(false)
	d.percentField.Accessibility.Name = i18n.Text("Percent of full price")
	d.percentField.Tooltip = newWrappedTooltip(i18n.Text("The percentage of each item's full price to pay"))
	d.percentField.Accessibility.LabeledBy = label
	d.header.AddChild(d.percentField)
	return d.header
}

// createFooter returns the row below the drop area: the summary of the purchase (see syncSummary) and the button that
// finalizes it.
func (d *buyDockable) createFooter() *unison.Panel {
	d.footer = newBuyRow(2)
	d.summary, d.setSummary = newBuyNoticeLabel()
	d.summary.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	d.footer.AddChild(d.summary)
	// Finalizing closes the dockable, which is left until the click that asked for it is over.
	d.finalize = newBuyButton(i18n.Text("Finalize Purchase…"),
		i18n.Text("Settle the money to pay with, then pay for the items and add them to the carried equipment"),
		func() { unison.InvokeTask(func() { d.finalizePurchase() }) })
	d.footer.AddChild(d.finalize)
	return d.footer
}

// newBuyRow returns a panel laying out the given number of columns in a row above or below the buy dockable's drop area.
func newBuyRow(columns int) *unison.Panel {
	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
	})
	row.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	return row
}

// newBuyButton returns a button for the buy dockable, named and described for assistive technologies.
func newBuyButton(title, tooltip string, onClick func()) *unison.Button {
	b := unison.NewButton()
	b.SetTitle(title)
	b.Accessibility.Name = title
	b.Tooltip = newWrappedTooltip(tooltip)
	b.ClickCallback = onClick
	b.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	return b
}

// newBuyNoticeLabel returns a wrapping label for the buy dockable and the function that sets its text, which is set
// out as a warning, in a box of the theme's warning color, when warn is true.
func newBuyNoticeLabel() (label *TextLabel, set func(text string, warn bool)) {
	label = NewWrappingLabel()
	warning := false
	label.DrawCallback = func(gc *unison.Canvas, dirty geom.Rect) {
		if warning {
			r := label.ContentRect(true)
			gc.DrawRoundedRect(r, geom.NewUniformSize(4), unison.ThemeWarning.Paint(gc, r, paintstyle.Fill))
		}
		label.draw(gc, dirty)
	}
	return label, func(text string, warn bool) {
		if text == label.Title() && warn == warning {
			return
		}
		warning = warn
		if warn {
			label.SetBorder(unison.NewEmptyBorder(geom.NewSymmetricInsets(unison.StdHSpacing, unison.StdVSpacing)))
			label.setText(text, unison.ThemeOnWarning)
		} else {
			label.SetBorder(nil)
			label.setText(text, unison.ThemeOnSurface)
		}
	}
}

// TitleIcon implements unison.Dockable.
func (d *buyDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.ShoppingCart,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (d *buyDockable) Title() string {
	return i18n.Text("Buy Equipment: %s", d.sheet.Entity().Profile.Name)
}

// Tooltip implements unison.Dockable.
func (d *buyDockable) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable. A purchase under way is what closing the dockable would lose, so it is what
// marks the dockable modified, as an editor's pending changes mark it.
func (d *buyDockable) Modified() bool {
	return !d.settled && len(d.items) != 0
}

// MayAttemptClose implements unison.TabCloser.
func (d *buyDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser. A purchase under way is asked about first, as an editor's pending changes
// are (see confirmClose).
func (d *buyDockable) AttemptClose() bool {
	if !d.confirmClose() {
		return false
	}
	return AttemptCloseForDockable(d)
}

// confirmClose asks what to do with a purchase under way: finalize it, discard it, or keep the dockable open, in which
// case false is returned. When the money carried doesn't cover it, so that finalizing is not an option, it asks whether
// to discard it instead. A purchase already made or discarded is not asked about.
func (d *buyDockable) confirmClose() bool {
	if !d.Modified() {
		return true
	}
	d.plan = d.planPayment(nil)
	if d.plan.short {
		if unison.QuestionDialog(i18n.Text("Discard the purchase for\n%s?", dockableTitle(d.sheet)),
			i18n.Text("The money carried doesn't cover it, so it can't be finalized.")) != unison.ModalResponseOK {
			return false
		}
		d.settled = true
		return true
	}
	switch unison.YesNoCancelDialog(i18n.Text("Finalize the purchase for\n%s?", dockableTitle(d.sheet)),
		i18n.Text("%s\n%s", d.describeItems(), d.summary.Title())) {
	case unison.ModalResponseDiscard:
		d.settled = true
		return true
	case unison.ModalResponseOK:
		return d.finalizePurchase()
	default:
		return false
	}
}

// discardAndClose closes the dockable, dropping the purchase under way without asking about it.
func (d *buyDockable) discardAndClose() {
	d.settled = true
	d.AttemptClose()
}

// CloseWithGroup implements GroupedCloser.
func (d *buyDockable) CloseWithGroup(other unison.Paneler) bool {
	return other == d.sheet
}

// UndoManager implements unison.UndoManagerProvider.
func (d *buyDockable) UndoManager() *unison.UndoManager {
	return unison.UndoManagerFor(d.sheet)
}

// SheetChanged implements SheetSourceUser. The payment follows the money carried, which any edit to the sheet may
// change, until the purchase has been made.
func (d *buyDockable) SheetChanged(sheet *Sheet) {
	if sheet == d.sheet && !d.settled {
		d.Sync()
	}
}

func (d *buyDockable) installDropSupport() {
	d.CanAcceptDropCallback = func(di drag.Info) bool {
		return d.dropData(di) != nil || d.modifierDropData(di) != nil
	}
	update := func(di drag.Info, where geom.Point, _ mod.Modifiers) drag.Op {
		if d.modifierDropData(di) != nil {
			// Modifiers go to the item they are over, and nowhere else.
			target := d.itemAt(where)
			d.setModifierTarget(target)
			d.setDragOver(false)
			if target != nil {
				return drag.Copy
			}
			return drag.None
		}
		d.setModifierTarget(nil)
		// The sheet reroutes drags over it here as well, so the drags this dockable can't take are declined here too,
		// not just by CanAcceptDropCallback.
		accept := d.dropData(di) != nil
		d.setDragOver(accept)
		if accept {
			return drag.Copy
		}
		return drag.None
	}
	d.DragEnteredCallback = update
	d.DragUpdatedCallback = update
	d.DragExitedCallback = func() {
		d.setDragOver(false)
		d.setModifierTarget(nil)
	}
	d.DropCallback = func(di drag.Info, where geom.Point, _ mod.Modifiers) bool {
		d.setDragOver(false)
		d.setModifierTarget(nil)
		if data := d.modifierDropData(di); data != nil {
			target := d.itemAt(where)
			return target != nil && d.attachModifiers(target, data)
		}
		data := d.dropData(di)
		return data != nil && d.addDropped(data)
	}
}

// claimsDrag returns true if the equipment being dragged comes from elsewhere than the sheet and would be bought were
// it dropped on this dockable, in which case the sheet and its lists leave the drag to it (see Sheet.keyToPanel and
// PageList.installBuyDragDeferral). The sheet's own equipment is left to its lists, so rows can be moved about while
// buying; dropped on the dockable, it is bought like any other.
func (d *buyDockable) claimsDrag(di drag.Info) bool {
	data := d.dropData(di)
	if data == nil {
		return false
	}
	entity := d.sheet.Entity()
	for _, row := range data.Rows {
		if gurps.EntityFromNode(row.Data()) == entity {
			return false
		}
	}
	return true
}

// modifierDropData returns the equipment modifier rows being dragged, or nil if the drag holds none.
func (d *buyDockable) modifierDropData(di drag.Info) *unison.TableDragData[*Node[*gurps.EquipmentModifier]] {
	if !di.HasDataType(equipmentModifierDragKey.UTI) {
		return nil
	}
	data, ok := draggedTableData.(*unison.TableDragData[*Node[*gurps.EquipmentModifier]])
	if !ok || data == nil || len(data.Rows) == 0 {
		return nil
	}
	return data
}

// itemAt returns the item under the point, which is in the dockable's coordinates, if it can take modifiers, or nil.
func (d *buyDockable) itemAt(where geom.Point) *buyItem {
	local := d.dropArea.PointFromRoot(d.PointToRoot(where))
	for _, item := range d.items {
		if gurps.CanTakeModifiers(item.eqp) && local.In(d.itemBounds(item)) {
			return item
		}
	}
	return nil
}

// itemBounds returns the rows of the item -- its own and those of its modifiers -- in the drop area's coordinates,
// across the whole of its width within the dashed edge and extended halfway to the rows above and below, so that the
// bounds of one item meet the next's.
func (d *buyDockable) itemBounds(item *buyItem) geom.Rect {
	var bounds geom.Rect
	for _, w := range item.allWidgets() {
		if bounds.Empty() {
			bounds = w.AsPanel().FrameRect()
		} else {
			bounds = bounds.Union(w.AsPanel().FrameRect())
		}
	}
	area := d.dropArea.ContentRect(true).Inset(geom.NewUniformInsets(1))
	bounds.X = area.X
	bounds.Width = area.Width
	bounds.Y -= unison.StdVSpacing / 2
	bounds.Height += unison.StdVSpacing
	return bounds
}

func (d *buyDockable) setModifierTarget(item *buyItem) {
	if d.modifierTarget != item {
		d.modifierTarget = item
		flushDragFeedback(d.dropArea)
	}
}

// attachModifiers attaches clones of the dragged modifiers to the item, asking the questions a drop onto the carried
// equipment would (see attachModifiers). Returns false if the user canceled.
func (d *buyDockable) attachModifiers(item *buyItem, data *unison.TableDragData[*Node[*gurps.EquipmentModifier]]) bool {
	modifiers := make([]*gurps.EquipmentModifier, 0, len(data.Rows))
	for _, row := range data.Rows {
		modifiers = append(modifiers, row.Data())
	}
	if !attachModifiers(i18n.Text("Drag"), d.sheet.Entity(), []*gurps.Equipment{item.eqp}, modifiers,
		libraryFileFromTable(data.Table)) {
		return false
	}
	d.layoutRows()
	d.Sync()
	return true
}

// removeModifier takes the modifier back off the item. The focus, if it was on the modifier's row, goes to the item's
// (see layoutRows).
func (d *buyDockable) removeModifier(item *buyItem, modifier *gurps.EquipmentModifier) {
	item.eqp.Modifiers = slices.DeleteFunc(item.eqp.Modifiers,
		func(one *gurps.EquipmentModifier) bool { return one == modifier })
	d.layoutRows()
	d.Sync()
}

// dropData returns the equipment rows being dragged, or nil if the drag holds none.
func (d *buyDockable) dropData(di drag.Info) *unison.TableDragData[*Node[*gurps.Equipment]] {
	if !di.HasDataType(equipmentDragKey.UTI) {
		return nil
	}
	data, ok := draggedTableData.(*unison.TableDragData[*Node[*gurps.Equipment]])
	if !ok || data == nil || len(data.Rows) == 0 {
		return nil
	}
	return data
}

func (d *buyDockable) setDragOver(over bool) {
	if d.dragOver != over {
		d.dragOver = over
		flushDragFeedback(d.dropArea)
	}
}

// addDropped adds copies of the dragged rows to the items to buy (see add). Returns false if the user canceled.
func (d *buyDockable) addDropped(data *unison.TableDragData[*Node[*gurps.Equipment]]) bool {
	list := make([]*gurps.Equipment, 0, len(data.Rows))
	for _, row := range data.Rows {
		list = append(list, row.Data())
	}
	return d.add(data.Table, list)
}

// add adds copies of the equipment, which the table shows, to the items to buy, first putting the questions a drop onto
// the carried equipment would to the user (see promptForTransfer). Returns false if the user canceled.
func (d *buyDockable) add(table *unison.Table[*Node[*gurps.Equipment]], list []*gurps.Equipment) bool {
	entity := d.sheet.Entity()
	carried := d.sheet.CarriedEquipment.Table
	from := libraryFileFromTable(table)
	part := applyPart[*gurps.Equipment]{table: carried, index: -1}
	for _, eqp := range list {
		part.rows = append(part.rows, eqp.Clone(from, entity, nil, gurps.Reference))
	}
	parts := newApplyParts(part)
	opts := applyOptionsFor(table, carried)
	op := promptOperation{
		name:        i18n.Text("Buy"),
		description: i18n.Text("Buying %s for %s", shortNames(describeRows(part.rows), dockableTitle(d.sheet))...),
	}
	if !promptForTransfer(parts, opts, op) {
		return false
	}
	// The template choices may have replaced the rows.
	rows := parts.equipment.rows
	if opts.clearTemplateOnly {
		parts.each(applyPartOps.clearTemplateOnly)
	}
	if opts.clearPreconfigured {
		gurps.ClearPreconfigured(rows...)
	}
	for _, eqp := range rows {
		// What is bought is enabled and equipped, as the purchase carries it (see gurps.Entity.CommitPurchase), so
		// that it is priced that way, even when copied from a sheet where it was disabled.
		gurps.Traverse(func(one *gurps.Equipment) bool {
			one.SetDisabled(false)
			one.Equipped = true
			return false
		}, false, false, eqp)
		if eqp.IsGroup() {
			eqp.Quantity = fxp.One
		} else {
			eqp.Quantity = min(max(eqp.Quantity, fxp.One), fxp.FromInteger(maxBuyQuantity))
		}
		d.addItem(eqp)
	}
	d.Sync()
	return true
}

func (d *buyDockable) addItem(eqp *gurps.Equipment) {
	item := &buyItem{eqp: eqp}
	name := eqp.String()
	item.field = NewDecimalField(d.targetMgr, "buy.qty."+string(eqp.TID), i18n.Text("Quantity to Buy"),
		item.quantity,
		func(v fxp.Int) {
			if !eqp.IsGroup() && v != eqp.Quantity {
				eqp.Quantity = v
				d.refresh()
			}
		}, fxp.One, fxp.FromInteger(maxBuyQuantity), false, false)
	item.field.withoutUndo()
	item.field.SetMarksModified(false)
	item.field.Accessibility.Name = i18n.Text("%s quantity to buy", name)
	if eqp.IsGroup() {
		item.field.SetEnabled(false)
		item.field.Tooltip = newWrappedTooltip(i18n.Text("A group is bought just once"))
	} else {
		item.field.Tooltip = newWrappedTooltip(i18n.Text("The quantity to buy"))
	}
	item.name = newBuyText(name)
	item.name.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	item.each = newBuyText("")
	item.each.HAlign = align.End
	item.each.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	item.total = newBuyText("")
	item.total.HAlign = align.End
	item.total.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	item.remove = unison.NewSVGButton(unison.TrashSVG)
	item.remove.Accessibility.Name = i18n.Text("Remove %s", name)
	item.remove.Tooltip = newWrappedTooltip(i18n.Text("Remove from the purchase"))
	// Removing the row takes the button out of the dockable, which is left until the click that asked for it is over.
	item.remove.ClickCallback = func() { unison.InvokeTask(func() { d.removeItem(item) }) }
	item.remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	d.items = append(d.items, item)
	d.layoutRows()
}

// removeItem takes the item off the purchase. The focus, if it was on the item's rows, goes to the row that takes its
// place, or to the last row, or to the percent field once no row is left, since the focus would otherwise be left on a
// panel that is no longer in the window.
func (d *buyDockable) removeItem(item *buyItem) {
	i := slices.Index(d.items, item)
	if i == -1 {
		return
	}
	focus := d.currentFocus()
	hadFocus := focus != nil && slices.ContainsFunc(item.allWidgets(), func(w unison.Paneler) bool {
		return w.AsPanel() == focus
	})
	d.items = slices.Delete(d.items, i, i+1)
	d.layoutRows()
	if hadFocus {
		var target *unison.Panel
		if len(d.items) != 0 {
			target = d.items[min(i, len(d.items)-1)].focusTarget()
		}
		if target == nil {
			target = d.percentField.AsPanel()
		}
		target.RequestFocus()
	}
	d.Sync()
}

// currentFocus returns the panel holding the keyboard focus in the dockable's window, or nil.
func (d *buyDockable) currentFocus() *unison.Panel {
	if wnd := d.Window(); wnd != nil {
		return wnd.CurrentFocus()
	}
	return nil
}

// layoutRows fills the drop area with the hint, then each item's row followed by a row for each of its modifiers,
// which are built afresh so that they match the modifiers the item now has. The focus, if it was on a modifier's row,
// goes to the row built anew for the modifier, or to its item's row when the modifier is gone (see removeModifier).
func (d *buyDockable) layoutRows() {
	focusedItem, focusedModifier, focusedWidget := d.focusedModifierRow()
	d.dropArea.RemoveAllChildren()
	d.dropArea.AddChild(d.hint)
	for _, item := range d.items {
		item.modifierRows = item.modifierRows[:0]
		for _, modifier := range item.eqp.Modifiers {
			item.modifierRows = append(item.modifierRows, d.newModifierRow(item, modifier))
		}
		for _, w := range item.allWidgets() {
			d.dropArea.AddChild(w)
		}
	}
	d.dropArea.MarkForLayoutRecursivelyUpward()
	d.dropArea.MarkForRedraw()
	if focusedItem != nil {
		target := focusedItem.focusTarget()
		for _, row := range focusedItem.modifierRows {
			if row.modifier == focusedModifier {
				target = row.widgets[focusedWidget].AsPanel()
				break
			}
		}
		if target != nil {
			target.RequestFocus()
		}
	}
}

// focusedModifierRow returns the item and modifier of the modifier row holding the keyboard focus, along with the index
// of the widget holding it, or a nil item when no modifier row holds it.
func (d *buyDockable) focusedModifierRow() (item *buyItem, modifier *gurps.EquipmentModifier, widget int) {
	focus := d.currentFocus()
	if focus == nil {
		return nil, nil, 0
	}
	for _, one := range d.items {
		for _, row := range one.modifierRows {
			for i, w := range row.widgets {
				if w.AsPanel() == focus {
					return one, row.modifier, i
				}
			}
		}
	}
	return nil, nil, 0
}

// newModifierRow returns the row shown under an item for one of its modifiers: its description, set in from the
// item's name, and a button that takes it back off the item.
func (d *buyDockable) newModifierRow(item *buyItem, modifier *gurps.EquipmentModifier) buyModifierRow {
	description := modifier.FullDescription()
	label := newBuyText(description)
	adjustForSmall(label, true)
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.DefaultLabelTheme.Font.LineHeight()}))
	if !modifier.Enabled() {
		label.OnBackgroundInk = dimmedTextColor
	}
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  buyRowColumns - 2,
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	remove := unison.NewSVGButton(unison.TrashSVG)
	remove.Accessibility.Name = i18n.Text("Remove %s from %s", description, item.eqp.String())
	remove.Tooltip = newWrappedTooltip(i18n.Text("Remove the modifier from the item"))
	// Removing the row takes the button out of the dockable, which is left until the click that asked for it is over.
	remove.ClickCallback = func() { unison.InvokeTask(func() { d.removeModifier(item, modifier) }) }
	remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	return buyModifierRow{modifier: modifier, widgets: []unison.Paneler{unison.NewPanel(), label, remove}}
}

// newBuyText returns a label showing text in a row of the buy dockable.
func newBuyText(text string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = unison.ThemeOnBelowSurface
	label.SetTitle(text)
	return label
}

func (item *buyItem) widgets() []unison.Paneler {
	return []unison.Paneler{item.field, item.name, item.each, item.total, item.remove}
}

// focusTarget returns the first of the item's own widgets that can take the keyboard focus: its quantity field, or its
// remove button for a group, whose field is disabled. Nil is returned when the item isn't in a window.
func (item *buyItem) focusTarget() *unison.Panel {
	for _, w := range item.widgets() {
		if panel := w.AsPanel(); panel.Focusable() && panel.Window() != nil {
			return panel
		}
	}
	return nil
}

// allWidgets returns the item's widgets followed by those of its modifier rows.
func (item *buyItem) allWidgets() []unison.Paneler {
	all := item.widgets()
	for _, row := range item.modifierRows {
		all = append(all, row.widgets...)
	}
	return all
}

// quantity returns the quantity to buy: the equipment's own, or one for a group.
func (item *buyItem) quantity() fxp.Int {
	if item.eqp.IsGroup() {
		return fxp.One
	}
	return item.eqp.Quantity
}

// Sync implements Syncer. The percentage and the transaction name follow the sheet's buy percentage and the items until
// the user changes them, even while their fields have the focus, since the buttons that change the items don't take
// it, the payment follows the money carried and the title follows the character's name.
func (d *buyDockable) Sync() {
	if d.percentEdited {
		d.percentField.Sync()
	} else {
		d.percent = d.sheetPercent()
		d.percentField.syncFromValue()
	}
	if d.nameEdited {
		d.nameField.Sync()
	} else {
		d.name = d.defaultName()
		d.nameField.syncFromValue()
	}
	d.refresh()
	if title, modified := d.Title(), d.Modified(); title != d.title || modified != d.modified {
		d.title = title
		d.modified = modified
		UpdateTitleForDockable(d)
	}
}

// refresh brings the prices, the payment and the buttons up to date with the items and the money carried.
func (d *buyDockable) refresh() {
	for _, item := range d.items {
		item.each.SetTitle(i18n.Text("$%s each", d.formatValue(gurps.TradePrice(item.eqp, fxp.One, d.percent))))
		item.total.SetTitle(i18n.Text("Total $%s",
			d.formatValue(gurps.TradePrice(item.eqp, item.quantity(), d.percent))))
	}
	d.plan = d.planPayment(nil)
	d.syncSummary()
	d.finalize.SetEnabled(len(d.items) != 0 && !d.plan.short)
	d.MarkForLayoutRecursively()
	d.MarkForRedraw()
}

// syncSummary has the summary show the total price of the items, or warn that the money carried doesn't cover it. How
// the total is paid is settled when the purchase is finalized (see confirmPurchase), so nothing of that is shown here.
func (d *buyDockable) syncSummary() {
	switch {
	case len(d.items) == 0:
		d.setSummary(i18n.Text("Nothing to buy yet."), false)
	case d.plan.short:
		d.setSummary(i18n.Text("Not enough money carried: $%s needed, $%s carried.", d.formatValue(d.plan.total),
			d.formatValue(d.carriedMoney())), true)
	default:
		d.setSummary(i18n.Text("Total: $%s", d.formatValue(d.plan.total)), false)
	}
}

// planPayment works out the total price of the items and how it is paid for: with the quantity of each carried money
// item given in chosen, by ID, or automatically when chosen is nil. The sheet's sell money is looked up in the libraries
// each time, since they may have changed; the lookup is cheap once the library files are loaded, as the entity's source
// matcher reloads one only when it has changed on disk.
func (d *buyDockable) planPayment(chosen map[tid.TID]int) buyPlan {
	var plan buyPlan
	if len(d.items) == 0 {
		return plan
	}
	for _, item := range d.items {
		plan.total = fxp.SaturatingAdd(plan.total, gurps.TradePrice(item.eqp, item.quantity(), d.percent))
	}
	entity := d.sheet.Entity()
	money := entity.MoneyItems()
	if chosen != nil {
		plan.payments, plan.paid = manualPayouts(chosen, money)
	} else {
		plan.payments, plan.paid = d.automaticPayment(plan.total, money)
	}
	plan.short = plan.paid < plan.total
	plan.sellMoney = entity.SellMoneyItems()
	if !plan.short && plan.paid > plan.total {
		plan.change, plan.changePaid = d.changeFor(plan.paid-plan.total, plan.payments, plan.sellMoney)
	}
	return plan
}

// automaticPayment returns the carried money that pays for total (see gurps.PaymentFor), and what it comes to. It is
// worked out again only when the total or the money carried has changed since the last time.
func (d *buyDockable) automaticPayment(total fxp.Int, money []*gurps.Equipment) (payments []gurps.Payout, paid fxp.Int) {
	state := make([]buyMoneyState, len(money))
	for i, one := range money {
		state[i] = buyMoneyState{id: one.TID, value: one.AdjustedValue(), quantity: one.TradeQuantity()}
	}
	if !d.auto.valid || d.auto.total != total || !slices.Equal(d.auto.money, state) {
		found, foundPaid := gurps.PaymentFor(total, money)
		spend := make([]buyMoneySpend, 0, len(found))
		for _, one := range found {
			spend = append(spend, buyMoneySpend{id: one.Money.TID, quantity: one.Quantity})
		}
		d.auto = buyAutoPayment{money: state, spend: spend, total: total, paid: foundPaid, valid: true}
	}
	byID := make(map[tid.TID]*gurps.Equipment, len(money))
	for _, one := range money {
		byID[one.TID] = one
	}
	for _, one := range d.auto.spend {
		if eqp := byID[one.id]; eqp != nil && one.quantity > 0 {
			payments = append(payments, gurps.Payout{Money: eqp, Quantity: one.quantity})
		}
	}
	return payments, d.auto.paid
}

// manualPayouts returns the payouts for the quantities chosen of each carried money item, by ID, and what they come
// to. A choice whose money is no longer carried, or no longer carried in that quantity, is left out.
func manualPayouts(chosen map[tid.TID]int, money []*gurps.Equipment) (payments []gurps.Payout, paid fxp.Int) {
	for _, one := range money {
		qty := chosen[one.TID]
		if qty <= 0 || fxp.FromInteger(qty) > one.TradeQuantity() {
			continue
		}
		payments = append(payments, gurps.Payout{Money: one, Quantity: qty})
		paid = fxp.SaturatingAdd(paid, one.AdjustedValue().Mul(fxp.FromInteger(qty)))
	}
	return payments, paid
}

// changeFor returns the change for an overpayment of target, returned in the sell money denominations (see
// gurps.Entity.ChangeFor), and what it comes to. It is worked out again only when the amount, the money paid with or
// the sell money has changed since the last time.
func (d *buyDockable) changeFor(target fxp.Int, payments []gurps.Payout, sellMoney []*gurps.Equipment) (change []gurps.Payout, paid fxp.Int) {
	if !d.change.valid || d.change.target != target ||
		!slices.EqualFunc(d.change.sellMoney, sellMoney, buySellMoneyState.matches) ||
		!slices.EqualFunc(d.change.given, payments, buyGivenMoney.matches) {
		denominations := make([]buySellMoneyState, 0, len(sellMoney))
		for _, one := range sellMoney {
			denominations = append(denominations, newBuySellMoneyState(one))
		}
		given := make([]buyGivenMoney, 0, len(payments))
		for _, one := range payments {
			given = append(given, newBuyGivenMoney(one))
		}
		found, foundPaid := d.sheet.Entity().ChangeFor(target, sellMoney, payments)
		d.change = buyChange{
			given:     given,
			sellMoney: denominations,
			change:    found,
			target:    target,
			paid:      foundPaid,
			valid:     true,
		}
	}
	return d.change.change, d.change.paid
}

// unreturnedChangeNotice warns that unreturned of the change can't be returned, given how many sell money items the
// sheet's settings list and how many of them were found in the libraries (see gurps.Entity.SellMoneyItems): it says how
// to get change when the sheet has no sell money, and that none of it could be found when it has some but none was.
func unreturnedChangeNotice(unreturned string, configured, found int) string {
	switch {
	case configured == 0:
		return i18n.Text("$%s of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
			unreturned)
	case found == 0:
		return i18n.Text("$%s of change can't be returned from the money paid, and none of the sheet's sell money could be found in the libraries.",
			unreturned)
	default:
		return i18n.Text("$%s of change can't be returned with the sheet's sell money or the money paid.", unreturned)
	}
}

// carriedMoney returns the value of the whole units of money carried, which is all a payment can use.
func (d *buyDockable) carriedMoney() fxp.Int {
	var total fxp.Int
	for _, one := range d.sheet.Entity().MoneyItems() {
		total = fxp.SaturatingAdd(total, one.AdjustedValue().Mul(one.TradeQuantity().Floor()))
	}
	return total
}

// describePayouts names each payout with its quantity, such as "Copper Penny ×5", separated by commas.
func describePayouts(payouts []gurps.Payout) string {
	parts := make([]string, 0, len(payouts))
	for _, one := range payouts {
		parts = append(parts, i18n.Text("%s ×%d", one.Money.String(), one.Quantity))
	}
	return strings.Join(parts, i18n.Text(", "))
}

// describeItems names the items with the quantity of each, such as "Longbow ×1".
func (d *buyDockable) describeItems() string {
	names := make([]string, 0, len(d.items))
	for _, item := range d.items {
		names = append(names, i18n.Text("%s ×%s", item.eqp.String(), item.quantity().String()))
	}
	return strings.Join(names, i18n.Text(", "))
}

func (d *buyDockable) defaultName() string {
	if len(d.items) == 0 {
		return ""
	}
	names := make([]string, 0, len(d.items))
	for _, item := range d.items {
		names = append(names, item.eqp.String())
	}
	return i18n.Text("Bought %s", tradeNames(names))
}

func (d *buyDockable) sheetPercent() fxp.Int {
	return min(max(d.sheet.Entity().SheetSettings.BuyPercent, gurps.MinTradePercent), gurps.MaxTradePercent)
}

func (d *buyDockable) formatValue(value fxp.Int) string {
	return d.sheet.Entity().SheetSettings.FormatEquipmentValue(value)
}

// finalizePurchase settles the money to pay with and whether to merge what is bought (see confirmPurchase), then pays
// for the items, adds them and any change to the carried equipment and records the purchase, as one undoable edit, and
// closes the dockable, returning to the sheet if the focus was within the dockable. It returns false if the purchase
// was not made, for any of the reasons below, in which case the dockable stays open.
//
// An editor doesn't follow changes to its item, so one left open on an item the purchase changes would put its stale
// quantity back when applied. The editors open on everything the purchase changes are closed first, as they are for a
// sale, and the purchase is abandoned if one can't be closed: the carried money spent, each container the payment
// splits apart to spend it, and the carried items the purchase and the change merge into. An editor saved at its prompt
// applies its changes, which may alter the money carried, so the purchase is made only if the payment stands as the
// user accepted it; otherwise the dockable, which the save brought up to date, stays for the user to look over.
func (d *buyDockable) finalizePurchase() bool {
	if d.settled || buyDockableFor(d.sheet) != d {
		return false
	}
	d.plan = d.planPayment(nil)
	if len(d.items) == 0 || d.plan.short {
		return false
	}
	terms, ok := confirmPurchase(d, &d.plan)
	if !ok || d.settled || buyDockableFor(d.sheet) != d {
		return false
	}
	d.merge = terms.merge
	plan := d.planPayment(terms.chosen)
	if plan.short {
		return false
	}
	items := make([]gurps.PurchaseItem, 0, len(d.items))
	for _, item := range d.items {
		items = append(items, gurps.PurchaseItem{Equipment: item.eqp, Quantity: item.quantity()})
	}
	merge := terms.merge
	entity := d.sheet.Entity()
	removed := make(map[tid.TID]bool)
	changed := make(map[tid.TID]bool)
	for _, payment := range plan.payments {
		changed[payment.Money.TID] = true
		if fxp.FromInteger(payment.Quantity) >= payment.Money.TradeQuantity() {
			removed[payment.Money.TID] = true
		}
	}
	for _, eqp := range entity.PaymentSplits(plan.payments) {
		changed[eqp.TID] = true
	}
	if merge {
		for _, item := range items {
			if target := entity.MergeTarget(item.Equipment, removed); target != nil {
				changed[target.TID] = true
			}
		}
		for _, target := range entity.PayoutMergeTargets(plan.change, removed) {
			if target != nil {
				changed[target.TID] = true
			}
		}
	}
	if !CloseID(changed) || d.settled || buyDockableFor(d.sheet) != d {
		return false
	}
	if again := d.planPayment(terms.chosen); !again.same(&plan) {
		return false
	}
	name := strings.TrimSpace(d.nameField.Text())
	if name == "" {
		name = d.defaultName()
	}
	hadFocus := false
	if focus := d.currentFocus(); focus != nil {
		hadFocus = unison.AncestorIsOrSelf(focus, d)
	}
	d.settled = true
	recordTrade(d.sheet, i18n.Text("Buy Equipment"), func() {
		entity.CommitPurchase(items, plan.payments, plan.change, merge, name, jio.Now())
	})
	d.AttemptClose()
	if hadFocus {
		d.returnToSheet()
	}
	return true
}

// returnToSheet brings the sheet forward with the keyboard focus on its carried equipment list, which now holds what
// was bought, or elsewhere on the sheet while the layout leaves the list off the page.
func (d *buyDockable) returnToSheet() {
	ActivateDockable(d.sheet)
	if table := d.sheet.CarriedEquipment.Table; unison.AncestorIs(table, d.sheet.page) {
		table.RequestFocusWithoutScroll()
	}
}

// same returns true if the other plan is this one: the same total, paid for with the same payments and returning the
// same change.
func (plan *buyPlan) same(other *buyPlan) bool {
	return plan.total == other.total && plan.paid == other.paid && plan.changePaid == other.changePaid &&
		plan.short == other.short && samePayouts(plan.payments, other.payments) && samePayouts(plan.change, other.change)
}

// samePayouts returns true if the payouts are the same money in the same quantities, in the same order. Money is the same
// when it is the one item, or when it matches in everything but identity (see gurps.Equipment.MergeMatches), since the
// sell money is copied afresh from the library each time it is looked up (see gurps.Entity.SellMoneyItems).
func samePayouts(a, b []gurps.Payout) bool {
	return slices.EqualFunc(a, b, func(x, y gurps.Payout) bool {
		return x.Quantity == y.Quantity && (x.Money == y.Money || x.Money.MergeMatches(y.Money))
	})
}
