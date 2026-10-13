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

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	// maxBuyQuantity is the most of one item the buy panel buys at once.
	maxBuyQuantity = 9999
	// maxBuyPercent is the highest percentage of the full price the buy panel accepts.
	maxBuyPercent = 1000
	// buyRowColumns is the number of columns of a row of items to buy: quantity, name, price of each, total and remove.
	buyRowColumns = 5
)

var _ Syncer = &buyPanel{}

// buyPanel stands in for a sheet's carried equipment list while the character buys equipment (see Sheet.startBuying).
// It lists the equipment dropped on it from any list but the sheet's own, with the quantity of each to buy, and pays
// for it with the carried money (see gurps.Entity.MoneyItems), chosen automatically or by the user, returning any
// change in the sheet's sell money. Nothing on the sheet changes until the purchase is finalized: the panel's fields
// neither record undo nor mark the sheet as modified.
type buyPanel struct {
	unison.Panel
	sheet        *Sheet
	back         *unison.Button
	dropArea     *unison.Panel
	hint         *TextLabel
	footer       *unison.Panel
	percentField *DecimalField
	summary      *TextLabel
	setSummary   func(text string, warn bool)
	notice       *TextLabel
	setNotice    func(text string, warn bool)
	manualButton *unison.Button
	nameField    *StringField
	merge        *unison.CheckBox
	finalize     *unison.Button
	items        []*buyItem
	percent      fxp.Int
	name         string
	// manual holds the quantity of each carried money item, by ID, the user chose to pay with, or is nil to pay
	// automatically. Money items are kept by ID rather than by pointer, since undoing an edit replaces them.
	manual map[tid.TID]int
	auto   buyAutoPayment
	change buyChange
	// sellSources is the sheet's sell money as it stood when sellMoney was resolved from it.
	sellSources  []gurps.Source
	sellMoney    []*gurps.Equipment
	sellResolved bool
	plan         buyPlan
	// percentEdited and nameEdited are set once the user changes the percentage or the transaction name, which until
	// then follow the sheet's buy percentage and the items being bought.
	percentEdited bool
	nameEdited    bool
	dragOver      bool
	// modifierTarget is the item that the equipment modifiers being dragged would be attached to, while they are over
	// it.
	modifierTarget *buyItem
	// held keeps the panel at least as tall as the carried equipment list it stands in for; see holdHeight.
	held *heldHeightLayout
}

// heldHeightLayout lays a panel out with the layout it wraps, but never lets the panel ask for less than its height, so
// that a panel standing in for another on the page doesn't shift what lies below it until it needs more room.
type heldHeightLayout struct {
	unison.Layout
	height float32
}

// LayoutSizes implements unison.Layout.
func (l *heldHeightLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	minSize, prefSize, maxSize = l.Layout.LayoutSizes(target, hint)
	minSize.Height = max(minSize.Height, l.height)
	prefSize.Height = max(prefSize.Height, l.height)
	maxSize.Height = max(maxSize.Height, prefSize.Height)
	return minSize, prefSize, maxSize
}

// buyItem is a row of the buy panel: a copy of the equipment to buy, owned by the sheet's entity but in none of its
// lists, whose Quantity is the quantity to buy, along with the widgets that show it.
type buyItem struct {
	eqp    *gurps.Equipment
	field  *DecimalField
	name   *unison.Label
	each   *unison.Label
	total  *unison.Label
	remove *unison.Button
	// modifierRows holds the widgets of the row shown under the item for each of its modifiers, which are built anew
	// whenever the rows are laid out (see buyPanel.layoutRows), so that they always match the modifiers the item has.
	modifierRows []buyModifierRow
}

// buyModifierRow is the row shown under an item of the buy panel for one of its modifiers: its description and the
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

// buyAutoPayment remembers the last automatic payment worked out, since the panel is synced with every change to the
// sheet and gurps.PaymentFor builds a table as large as the total to do it.
type buyAutoPayment struct {
	money []buyMoneyState
	spend []buyMoneySpend
	total fxp.Int
	paid  fxp.Int
	valid bool
}

// buyChange remembers the last change worked out, for the same reason as buyAutoPayment. The money paid with can be
// drawn on for change, so given is its state, without the quantities, which don't matter to it.
type buyChange struct {
	given  []buyMoneyState
	change []gurps.Payout
	target fxp.Int
	paid   fxp.Int
	valid  bool
}

func newBuyPanel(sheet *Sheet) *buyPanel {
	p := &buyPanel{sheet: sheet, held: &heldHeightLayout{Layout: &unison.FlexLayout{Columns: 1}}}
	p.Self = p
	p.SetLayout(p.held)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(colors.Header, geom.Size{}, geom.NewUniformInsets(1), false))
	p.Accessibility.Name = i18n.Text("Buy Equipment")
	p.percent = p.sheetPercent()
	p.AddChild(p.createHeader())
	p.AddChild(p.createBody())
	p.installDropSupport()
	p.Sync()
	return p
}

// holdHeight keeps the panel at least as tall as height, the height of the carried equipment list it is taking the
// place of, so that the page below it doesn't jump when it does.
func (p *buyPanel) holdHeight(height float32) {
	p.held.height = height
	p.MarkForLayoutAndRedraw()
}

func (p *buyPanel) createHeader() *unison.Panel {
	header := unison.NewPanel()
	header.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	header.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	header.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 1, Left: 4, Bottom: 1, Right: 1}))
	header.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, colors.Header.Paint(gc, rect, paintstyle.Fill))
	}

	title := unison.NewLabel()
	title.Text = unison.NewSmallCapsText(i18n.Text("Buy Equipment"), &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: colors.OnHeader,
	})
	title.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
		HGrab:  true,
	})
	header.AddChild(title)

	// Going back takes this panel off the page, which is left until the click that asked for it is over.
	p.back = newBuyPanelButton(i18n.Text("Back to Carried Equipment"),
		i18n.Text("Put the carried equipment back without buying anything"),
		func() { unison.InvokeTask(func() { p.leave(p.sheet.stopBuying) }) })
	header.AddChild(p.back)
	return header
}

func (p *buyPanel) createBody() *unison.Panel {
	body := unison.NewPanel()
	body.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	body.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	body.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(4)))
	body.AddChild(p.createDropArea())
	body.AddChild(p.createFooter())
	return body
}

func (p *buyPanel) createDropArea() *unison.Panel {
	p.dropArea = unison.NewPanel()
	p.dropArea.SetLayout(&unison.FlexLayout{
		Columns:  buyRowColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.dropArea.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	p.dropArea.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(2 * unison.StdHSpacing)))
	p.dropArea.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := p.dropArea.ContentRect(true)
		var ink unison.Ink = unison.ThemeSurfaceEdge
		if p.dragOver {
			ink = unison.ThemeFocus
			paint := ink.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
		r = r.Inset(geom.NewUniformInsets(0.5))
		paint := ink.Paint(gc, r, paintstyle.Stroke)
		paint.SetPathEffect(unison.DashEffect())
		gc.DrawRect(r, paint)
	}
	// The hint wraps, since the panel is no wider than the sheet's list and the hint is longer than a line of it.
	hint := NewWrappingLabel()
	hint.Font = fonts.PageFieldSecondary
	hint.setText(i18n.Text("Drop equipment here from any equipment list other than this sheet's own. Drop equipment modifiers onto an item to adjust it."),
		unison.ThemeOnSurface)
	hint.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  buyRowColumns,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.dropArea.AddChild(hint)
	p.hint = hint
	p.dropArea.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
		// The item the modifiers being dragged would go to is highlighted, as the rows of a list are (see
		// InstallTableDropSupport).
		if p.modifierTarget != nil {
			r := p.itemBounds(p.modifierTarget)
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}
	return p.dropArea
}

func (p *buyPanel) createFooter() *unison.Panel {
	p.footer = unison.NewPanel()
	p.footer.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	p.footer.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	payment := newBuyPanelRow(4)
	label := NewPageLabel(i18n.Text("Percent of Full Price"))
	label.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	payment.AddChild(label)
	p.percentField = NewDecimalPageField(p.sheet.targetMgr, "buy.percent", i18n.Text("Percent of Full Price"),
		func() fxp.Int { return p.percent },
		func(v fxp.Int) {
			if v != p.percent {
				p.percent = v
				p.percentEdited = true
				p.refresh()
			}
		}, fxp.One, fxp.FromInteger(maxBuyPercent), false)
	p.percentField.withoutUndo()
	p.percentField.SetMarksModified(false)
	p.percentField.Accessibility.Name = i18n.Text("Percent of full price")
	p.percentField.Tooltip = newWrappedTooltip(i18n.Text("The percentage of each item's full price to pay"))
	p.percentField.Accessibility.LabeledBy = label
	payment.AddChild(p.percentField)
	p.summary, p.setSummary = newBuyNoticeLabel()
	p.summary.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	payment.AddChild(p.summary)
	p.manualButton = newBuyPanelButton(i18n.Text("Choose Manually…"),
		i18n.Text("Choose the carried money to pay with"), p.choosePaymentManually)
	payment.AddChild(p.manualButton)
	p.footer.AddChild(payment)

	p.notice, p.setNotice = newBuyNoticeLabel()
	p.notice.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	finish := newBuyPanelRow(4)
	label = NewPageLabel(i18n.Text("Transaction"))
	label.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	finish.AddChild(label)
	p.nameField = NewStringPageField(p.sheet.targetMgr, "buy.name", i18n.Text("Transaction Name"),
		func() string { return p.name },
		func(s string) {
			if s != p.name {
				p.name = s
				p.nameEdited = true
			}
		})
	p.nameField.withoutUndo()
	p.nameField.SetMarksModified(false)
	p.nameField.Accessibility.Name = i18n.Text("Transaction name")
	p.nameField.Accessibility.LabeledBy = label
	p.nameField.Tooltip = newWrappedTooltip(i18n.Text("The name the shopping record gives the purchase"))
	finish.AddChild(p.nameField)
	p.merge = unison.NewCheckBox()
	p.merge.Font = fonts.PageFieldPrimary
	p.merge.SetTitle(i18n.Text("Merge with matching items"))
	p.merge.State = check.On
	p.merge.Accessibility.Name = i18n.Text("Merge with matching items")
	p.merge.Tooltip = newWrappedTooltip(i18n.Text("Add the quantity of each item bought, and of the change, to a carried item that matches it, rather than carrying a separate copy"))
	p.merge.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	finish.AddChild(p.merge)
	// Finalizing takes this panel off the page, which is left until the click that asked for it is over.
	p.finalize = newBuyPanelButton(i18n.Text("Finalize Purchase"),
		i18n.Text("Pay for the items and add them to the carried equipment"),
		func() { unison.InvokeTask(func() { p.leave(p.finalizePurchase) }) })
	finish.AddChild(p.finalize)
	p.footer.AddChild(finish)
	return p.footer
}

// newBuyPanelRow returns a panel laying out the given number of columns in a row of the buy panel's footer.
func newBuyPanelRow(columns int) *unison.Panel {
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

// newBuyPanelButton returns a button for the buy panel, named and described for assistive technologies.
func newBuyPanelButton(title, tooltip string, onClick func()) *unison.Button {
	b := unison.NewButton()
	b.Font = fonts.PageLabelPrimary
	b.HMargin = 4
	b.VMargin = 0
	b.SetTitle(title)
	b.Accessibility.Name = title
	b.Tooltip = newWrappedTooltip(tooltip)
	b.ClickCallback = onClick
	b.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	return b
}

// newBuyNoticeLabel returns a wrapping label for the buy panel and the function that sets its text, which is set out
// as a warning, in a box of the theme's warning color, when warn is true.
func newBuyNoticeLabel() (label *TextLabel, set func(text string, warn bool)) {
	label = NewWrappingLabel()
	label.Font = fonts.PageFieldPrimary
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

func (p *buyPanel) installDropSupport() {
	p.CanAcceptDropCallback = func(di drag.Info) bool {
		return p.dropData(di) != nil || p.modifierDropData(di) != nil
	}
	update := func(di drag.Info, where geom.Point, _ mod.Modifiers) drag.Op {
		if p.modifierDropData(di) != nil {
			// Modifiers go to the item they are over, and nowhere else.
			target := p.itemAt(where)
			p.setModifierTarget(target)
			p.setDragOver(false)
			if target != nil {
				return drag.Copy
			}
			return drag.None
		}
		p.setModifierTarget(nil)
		// The window reroutes drags over the rest of the sheet here as well, so the drags this panel can't take are
		// declined here too, not just by CanAcceptDropCallback.
		accept := p.dropData(di) != nil
		p.setDragOver(accept)
		if accept {
			return drag.Copy
		}
		return drag.None
	}
	p.DragEnteredCallback = update
	p.DragUpdatedCallback = update
	p.DragExitedCallback = func() {
		p.setDragOver(false)
		p.setModifierTarget(nil)
	}
	p.DropCallback = func(di drag.Info, where geom.Point, _ mod.Modifiers) bool {
		p.setDragOver(false)
		p.setModifierTarget(nil)
		if data := p.modifierDropData(di); data != nil {
			target := p.itemAt(where)
			return target != nil && p.attachModifiers(target, data)
		}
		data := p.dropData(di)
		return data != nil && p.addDropped(data)
	}
}

// claimsDrag returns true if the equipment being dragged would be bought were it dropped on this panel, in which case
// the rest of the sheet leaves the drag to it (see Sheet.keyToPanel and NewOtherEquipmentPageList).
func (p *buyPanel) claimsDrag(di drag.Info) bool {
	return p.dropData(di) != nil
}

// modifierDropData returns the equipment modifier rows being dragged, or nil if the drag holds none.
func (p *buyPanel) modifierDropData(di drag.Info) *unison.TableDragData[*Node[*gurps.EquipmentModifier]] {
	if !di.HasDataType(equipmentModifierDragKey.UTI) {
		return nil
	}
	data, ok := draggedTableData.(*unison.TableDragData[*Node[*gurps.EquipmentModifier]])
	if !ok || data == nil || len(data.Rows) == 0 {
		return nil
	}
	return data
}

// itemAt returns the item under the point, which is in the panel's coordinates, if it can take modifiers, or nil.
func (p *buyPanel) itemAt(where geom.Point) *buyItem {
	local := p.dropArea.PointFromRoot(p.PointToRoot(where))
	for _, item := range p.items {
		if gurps.CanTakeModifiers(item.eqp) && local.In(p.itemBounds(item)) {
			return item
		}
	}
	return nil
}

// itemBounds returns the rows of the item -- its own and those of its modifiers -- across the whole width of the drop
// area, in the drop area's coordinates.
func (p *buyPanel) itemBounds(item *buyItem) geom.Rect {
	var bounds geom.Rect
	for _, w := range item.allWidgets() {
		if bounds.Empty() {
			bounds = w.AsPanel().FrameRect()
		} else {
			bounds = bounds.Union(w.AsPanel().FrameRect())
		}
	}
	area := p.dropArea.ContentRect(false)
	bounds.X = area.X
	bounds.Width = area.Width
	return bounds
}

func (p *buyPanel) setModifierTarget(item *buyItem) {
	if p.modifierTarget != item {
		p.modifierTarget = item
		flushDragFeedback(p.dropArea)
	}
}

// attachModifiers attaches clones of the dragged modifiers to the item, asking the questions a drop onto the carried
// equipment would (see attachModifiers). Returns false if the user canceled.
func (p *buyPanel) attachModifiers(item *buyItem, data *unison.TableDragData[*Node[*gurps.EquipmentModifier]]) bool {
	modifiers := make([]*gurps.EquipmentModifier, 0, len(data.Rows))
	for _, row := range data.Rows {
		modifiers = append(modifiers, row.Data())
	}
	if !attachModifiers(i18n.Text("Drag"), p.sheet.Entity(), []*gurps.Equipment{item.eqp}, modifiers,
		libraryFileFromTable(data.Table)) {
		return false
	}
	p.layoutRows()
	p.Sync()
	return true
}

// removeModifier takes the modifier back off the item.
func (p *buyPanel) removeModifier(item *buyItem, modifier *gurps.EquipmentModifier) {
	item.eqp.Modifiers = slices.DeleteFunc(item.eqp.Modifiers,
		func(one *gurps.EquipmentModifier) bool { return one == modifier })
	p.layoutRows()
	p.Sync()
}

// dropData returns the equipment rows being dragged, or nil if the drag holds none or any of them belongs to this
// sheet, whose own equipment it can't buy.
func (p *buyPanel) dropData(di drag.Info) *unison.TableDragData[*Node[*gurps.Equipment]] {
	if !di.HasDataType(equipmentDragKey.UTI) {
		return nil
	}
	data, ok := draggedTableData.(*unison.TableDragData[*Node[*gurps.Equipment]])
	if !ok || data == nil || len(data.Rows) == 0 {
		return nil
	}
	entity := p.sheet.Entity()
	for _, row := range data.Rows {
		if gurps.EntityFromNode(row.Data()) == entity {
			return nil
		}
	}
	return data
}

func (p *buyPanel) setDragOver(over bool) {
	if p.dragOver != over {
		p.dragOver = over
		flushDragFeedback(p.dropArea)
	}
}

// addDropped adds copies of the dragged rows to the items to buy, first putting the questions a drop onto the carried
// equipment would to the user (see promptForTransfer). Returns false if the user canceled.
func (p *buyPanel) addDropped(data *unison.TableDragData[*Node[*gurps.Equipment]]) bool {
	entity := p.sheet.Entity()
	carried := p.sheet.CarriedEquipment.Table
	from := libraryFileFromTable(data.Table)
	part := applyPart[*gurps.Equipment]{table: carried, index: -1}
	for _, row := range data.Rows {
		part.rows = append(part.rows, row.Data().Clone(from, entity, nil, gurps.Reference))
	}
	parts := newApplyParts(part)
	opts := applyOptionsFor(data.Table, carried)
	op := promptOperation{
		name:        i18n.Text("Buy"),
		description: i18n.Text("Buying %s for %s", shortNames(describeRows(part.rows), dockableTitle(p.sheet))...),
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
		gurps.Traverse(func(one *gurps.Equipment) bool {
			one.Equipped = true
			return false
		}, false, false, eqp)
		if eqp.IsGroup() {
			eqp.Quantity = fxp.One
		} else {
			eqp.Quantity = min(max(eqp.Quantity, fxp.One), fxp.FromInteger(maxBuyQuantity))
		}
		p.addItem(eqp)
	}
	p.Sync()
	return true
}

func (p *buyPanel) addItem(eqp *gurps.Equipment) {
	item := &buyItem{eqp: eqp}
	name := eqp.String()
	item.field = NewDecimalPageField(p.sheet.targetMgr, "buy.qty."+string(eqp.TID), i18n.Text("Quantity to Buy"),
		item.quantity,
		func(v fxp.Int) {
			if !eqp.IsGroup() && v != eqp.Quantity {
				eqp.Quantity = v
				p.refresh()
			}
		}, fxp.One, fxp.FromInteger(maxBuyQuantity), false)
	item.field.withoutUndo()
	item.field.SetMarksModified(false)
	item.field.Accessibility.Name = i18n.Text("%s quantity to buy", name)
	if eqp.IsGroup() {
		item.field.SetEnabled(false)
		item.field.Tooltip = newWrappedTooltip(i18n.Text("A group is bought just once"))
	} else {
		item.field.Tooltip = newWrappedTooltip(i18n.Text("The quantity to buy"))
	}
	item.name = newBuyPanelText(name)
	item.name.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	item.each = newBuyPanelText("")
	item.each.HAlign = align.End
	item.each.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	item.total = newBuyPanelText("")
	item.total.HAlign = align.End
	item.total.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	item.remove = NewSVGButtonForFont(unison.TrashSVG, fonts.PageLabelPrimary, -2)
	item.remove.Accessibility.Name = i18n.Text("Remove %s", name)
	item.remove.Tooltip = newWrappedTooltip(i18n.Text("Remove from the purchase"))
	// Removing the row takes the button off the page, which is left until the click that asked for it is over.
	item.remove.ClickCallback = func() { unison.InvokeTask(func() { p.removeItem(item) }) }
	item.remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	p.items = append(p.items, item)
	p.layoutRows()
}

func (p *buyPanel) removeItem(item *buyItem) {
	i := slices.Index(p.items, item)
	if i == -1 {
		return
	}
	p.items = slices.Delete(p.items, i, i+1)
	if len(p.items) == 0 {
		// A manual choice of money was made for items no longer being bought.
		p.manual = nil
	}
	p.layoutRows()
	p.Sync()
}

// layoutRows fills the drop area with the hint, then each item's row followed by a row for each of its modifiers,
// which are built afresh so that they match the modifiers the item now has.
func (p *buyPanel) layoutRows() {
	p.dropArea.RemoveAllChildren()
	p.dropArea.AddChild(p.hint)
	for _, item := range p.items {
		item.modifierRows = item.modifierRows[:0]
		for _, modifier := range item.eqp.Modifiers {
			item.modifierRows = append(item.modifierRows, p.newModifierRow(item, modifier))
		}
		for _, w := range item.allWidgets() {
			p.dropArea.AddChild(w)
		}
	}
	p.dropArea.MarkForLayoutRecursivelyUpward()
	p.dropArea.MarkForRedraw()
}

// newModifierRow returns the row shown under an item for one of its modifiers: its description, set in from the
// item's name, and a button that takes it back off the item.
func (p *buyPanel) newModifierRow(item *buyItem, modifier *gurps.EquipmentModifier) buyModifierRow {
	description := modifier.FullDescription()
	label := newBuyPanelText(description)
	label.Font = fonts.PageFieldSecondary
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: fonts.PageFieldPrimary.LineHeight()}))
	if !modifier.Enabled() {
		label.OnBackgroundInk = dimmedTextColor
	}
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  buyRowColumns - 2,
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	remove := NewSVGButtonForFont(unison.TrashSVG, fonts.PageLabelPrimary, -2)
	remove.Accessibility.Name = i18n.Text("Remove %s from %s", description, item.eqp.String())
	remove.Tooltip = newWrappedTooltip(i18n.Text("Remove the modifier from the item"))
	// Removing the row takes the button off the page, which is left until the click that asked for it is over.
	remove.ClickCallback = func() { unison.InvokeTask(func() { p.removeModifier(item, modifier) }) }
	remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	return buyModifierRow{modifier: modifier, widgets: []unison.Paneler{unison.NewPanel(), label, remove}}
}

// newBuyPanelText returns a label showing text in the font of the page's fields.
func newBuyPanelText(text string) *unison.Label {
	label := unison.NewLabel()
	label.Font = fonts.PageFieldPrimary
	label.OnBackgroundInk = unison.ThemeOnSurface
	label.SetTitle(text)
	return label
}

func (item *buyItem) widgets() []unison.Paneler {
	return []unison.Paneler{item.field, item.name, item.each, item.total, item.remove}
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
// the user changes them, and the payment follows the money carried.
func (p *buyPanel) Sync() {
	if !p.percentEdited {
		p.percent = p.sheetPercent()
	}
	if !p.nameEdited {
		p.name = p.defaultName()
	}
	p.percentField.Sync()
	p.nameField.Sync()
	p.refresh()
}

// refresh brings the prices, the payment and the buttons up to date with the items and the money carried.
func (p *buyPanel) refresh() {
	for _, item := range p.items {
		item.each.SetTitle(i18n.Text("$%s each", p.formatValue(gurps.TradePrice(item.eqp, fxp.One, p.percent))))
		item.total.SetTitle(i18n.Text("Total $%s",
			p.formatValue(gurps.TradePrice(item.eqp, item.quantity(), p.percent))))
	}
	p.plan = p.planPayment()
	p.syncSummary()
	p.manualButton.SetEnabled(len(p.items) != 0)
	p.finalize.SetEnabled(len(p.items) != 0 && !p.plan.short)
	p.MarkForLayoutRecursively()
	MarkForLayoutWithinDockable(p)
	p.MarkForRedraw()
}

func (p *buyPanel) syncSummary() {
	plan := p.plan
	var summary, notice string
	warn := false
	switch {
	case len(p.items) == 0:
		summary = i18n.Text("Nothing to buy yet.")
	case plan.short && p.manual == nil:
		summary = i18n.Text("Not enough money carried: $%s needed, $%s carried.", p.formatValue(plan.total),
			p.formatValue(p.carriedMoney()))
		warn = true
	case plan.short:
		summary = i18n.Text("The money chosen is short by $%s: $%s needed, $%s chosen.",
			p.formatValue(plan.total-plan.paid), p.formatValue(plan.total), p.formatValue(plan.paid))
		warn = true
	case plan.paid == 0:
		summary = i18n.Text("Nothing to pay.")
	default:
		if p.manual == nil {
			summary = i18n.Text("Pay automatically: %s ($%s)", describePayouts(plan.payments),
				p.formatValue(plan.paid))
		} else {
			summary = i18n.Text("Pay manually: %s ($%s)", describePayouts(plan.payments), p.formatValue(plan.paid))
		}
		if plan.changePaid > 0 {
			summary += i18n.Text("; change: %s ($%s)", describePayouts(plan.change), p.formatValue(plan.changePaid))
		}
		if unreturned := plan.paid - plan.total - plan.changePaid; unreturned > 0 {
			notice = unreturnedChangeNotice(p.formatValue(unreturned), len(p.sellMoney) != 0)
		}
	}
	p.setSummary(summary, warn)
	p.setNotice(notice, notice != "")
	switch {
	case notice == "":
		p.notice.RemoveFromParent()
	case p.notice.Parent() == nil:
		p.footer.AddChildAtIndex(p.notice, 1)
	}
}

// planPayment works out the total price of the items and how it is paid for.
func (p *buyPanel) planPayment() buyPlan {
	var plan buyPlan
	if len(p.items) == 0 {
		return plan
	}
	for _, item := range p.items {
		plan.total += gurps.TradePrice(item.eqp, item.quantity(), p.percent)
	}
	money := p.sheet.Entity().MoneyItems()
	if p.manual != nil {
		plan.payments, plan.paid = manualPayouts(p.manual, money)
	} else {
		plan.payments, plan.paid = p.automaticPayment(plan.total, money)
	}
	plan.short = plan.paid < plan.total
	if !plan.short && plan.paid > plan.total {
		plan.change, plan.changePaid = p.changeFor(plan.paid-plan.total, plan.payments)
	}
	return plan
}

// automaticPayment returns the carried money that pays for total (see gurps.PaymentFor), and what it comes to. It is
// worked out again only when the total or the money carried has changed since the last time.
func (p *buyPanel) automaticPayment(total fxp.Int, money []*gurps.Equipment) (payments []gurps.Payout, paid fxp.Int) {
	state := make([]buyMoneyState, len(money))
	for i, one := range money {
		state[i] = buyMoneyState{id: one.TID, value: one.AdjustedValue(), quantity: one.Quantity}
	}
	if !p.auto.valid || p.auto.total != total || !slices.Equal(p.auto.money, state) {
		found, foundPaid := gurps.PaymentFor(total, money)
		spend := make([]buyMoneySpend, 0, len(found))
		for _, one := range found {
			spend = append(spend, buyMoneySpend{id: one.Money.TID, quantity: one.Quantity})
		}
		p.auto = buyAutoPayment{money: state, spend: spend, total: total, paid: foundPaid, valid: true}
	}
	byID := make(map[tid.TID]*gurps.Equipment, len(money))
	for _, one := range money {
		byID[one.TID] = one
	}
	for _, one := range p.auto.spend {
		if eqp := byID[one.id]; eqp != nil && one.quantity > 0 {
			payments = append(payments, gurps.Payout{Money: eqp, Quantity: one.quantity})
		}
	}
	return payments, p.auto.paid
}

// manualPayouts returns the payouts for the quantities chosen of each carried money item, by ID, and what they come
// to. A choice whose money is no longer carried, or no longer carried in that quantity, is left out.
func manualPayouts(chosen map[tid.TID]int, money []*gurps.Equipment) (payments []gurps.Payout, paid fxp.Int) {
	for _, one := range money {
		qty := chosen[one.TID]
		if qty <= 0 || fxp.FromInteger(qty) > one.Quantity {
			continue
		}
		payments = append(payments, gurps.Payout{Money: one, Quantity: qty})
		paid += one.AdjustedValue().Mul(fxp.FromInteger(qty))
	}
	return payments, paid
}

// changeFor returns the change for an overpayment of target, and what it comes to (see gurps.Entity.ChangeFor). It
// is worked out again only when the amount or the money paid with has changed since the last time.
func (p *buyPanel) changeFor(target fxp.Int, payments []gurps.Payout) (change []gurps.Payout, paid fxp.Int) {
	money := p.sellMoneyItems()
	given := make([]buyMoneyState, 0, len(payments))
	for _, one := range payments {
		given = append(given, buyMoneyState{id: one.Money.TID, value: one.Money.AdjustedValue()})
	}
	if !p.change.valid || p.change.target != target || !slices.Equal(p.change.given, given) {
		found, foundPaid := p.sheet.Entity().ChangeFor(target, money, payments)
		p.change = buyChange{given: given, change: found, target: target, paid: foundPaid, valid: true}
	}
	return p.change.change, p.change.paid
}

// unreturnedChangeNotice warns that unreturned of the change can't be returned, saying how to get change when the
// sheet has no sell money.
func unreturnedChangeNotice(unreturned string, hasSellMoney bool) string {
	if hasSellMoney {
		return i18n.Text("$%s of change can't be returned with the sheet's sell money or the money paid.", unreturned)
	}
	return i18n.Text("$%s of change can't be returned from the money paid, and the sheet has no sell money. Add some in Sheet Settings to receive change.",
		unreturned)
}

// sellMoneyItems returns the denominations of the sheet's sell money (see gurps.Entity.SellMoneyItems). They are looked
// up in the libraries only when first needed and again whenever the sheet's sell money changes.
func (p *buyPanel) sellMoneyItems() []*gurps.Equipment {
	entity := p.sheet.Entity()
	if !p.sellResolved || !slices.Equal(p.sellSources, entity.SheetSettings.SellMoney) {
		p.sellMoney = entity.SellMoneyItems()
		p.sellSources = slices.Clone(entity.SheetSettings.SellMoney)
		p.sellResolved = true
		p.change = buyChange{}
	}
	return p.sellMoney
}

// carriedMoney returns the value of the whole units of money carried, which is all a payment can use.
func (p *buyPanel) carriedMoney() fxp.Int {
	var total fxp.Int
	for _, one := range p.sheet.Entity().MoneyItems() {
		total += one.AdjustedValue().Mul(one.Quantity.Floor())
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
func (p *buyPanel) describeItems() string {
	names := make([]string, 0, len(p.items))
	for _, item := range p.items {
		names = append(names, i18n.Text("%s ×%s", item.eqp.String(), item.quantity().String()))
	}
	return strings.Join(names, i18n.Text(", "))
}

func (p *buyPanel) defaultName() string {
	if len(p.items) == 0 {
		return ""
	}
	names := make([]string, 0, len(p.items))
	for _, item := range p.items {
		names = append(names, item.eqp.String())
	}
	return i18n.Text("Bought %s", joinNames(names))
}

func (p *buyPanel) sheetPercent() fxp.Int {
	return min(max(p.sheet.Entity().SheetSettings.BuyPercent, fxp.One), fxp.FromInteger(maxBuyPercent))
}

func (p *buyPanel) formatValue(value fxp.Int) string {
	return p.sheet.Entity().SheetSettings.FormatEquipmentValue(value)
}

// choosePaymentManually lets the user choose the carried money to pay with, starting from the payment as it stands.
func (p *buyPanel) choosePaymentManually() {
	p.plan = p.planPayment()
	current := make(map[tid.TID]int, len(p.plan.payments))
	for _, one := range p.plan.payments {
		current[one.Money.TID] = one.Quantity
	}
	chosen, useAutomatic, ok := chooseManualPayment(p.sheet, p.plan.total, p.describeItems(), current)
	if !ok {
		return
	}
	switch {
	case useAutomatic:
		p.manual = nil
	case chosen == nil:
		p.manual = make(map[tid.TID]int)
	default:
		p.manual = chosen
	}
	p.refresh()
}

// leave runs fn, which may take the panel off the page, then gives the keyboard focus to the carried equipment list if
// the focus was within the panel and the panel is now off the page, as the focus would otherwise be left on a panel
// that is no longer in the window.
func (p *buyPanel) leave(fn func()) {
	var hadFocus bool
	if wnd := p.Window(); wnd != nil {
		if focus := wnd.Focus(); focus != nil {
			hadFocus = unison.AncestorIsOrSelf(focus, p)
		}
	}
	fn()
	if hadFocus && p.Parent() == nil {
		p.sheet.CarriedEquipment.Table.RequestFocusWithoutScroll()
	}
}

// finalizePurchase pays for the items, adds them and any change to the carried equipment and records the purchase, as
// one undoable edit, then puts the carried equipment list back in place of the panel.
func (p *buyPanel) finalizePurchase() {
	sheet := p.sheet
	if sheet.buyPanel != p {
		return
	}
	p.plan = p.planPayment()
	if len(p.items) == 0 || p.plan.short {
		return
	}
	plan := p.plan
	items := make([]gurps.PurchaseItem, 0, len(p.items))
	for _, item := range p.items {
		items = append(items, gurps.PurchaseItem{Equipment: item.eqp, Quantity: item.quantity()})
	}
	// The field's text is what the user sees, which differs from the name while the field has the focus and the items
	// have changed.
	name := strings.TrimSpace(p.nameField.Text())
	if name == "" {
		name = p.defaultName()
	}
	merge := p.merge.State == check.On
	entity := sheet.Entity()
	// The panel leaves the page in the rebuild that reports the purchase rather than in one of its own.
	p.RemoveFromParent()
	sheet.buyPanel = nil
	recordTrade(sheet, i18n.Text("Buy Equipment"), func() {
		entity.CommitPurchase(items, plan.payments, plan.change, merge, name, jio.Now())
	})
}
