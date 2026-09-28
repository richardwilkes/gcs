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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/side"
)

// The picker processing is held in a variable so that tests, which have no way to respond to the dialogs it presents, can
// substitute their own.
var promptForPickers = processPickers

// processPickers presents the template picker dialog for each row of the parts that has one, replacing the rows with
// the resulting choices. It returns false if the user canceled one of them, in which case the parts must be discarded.
// The operation describes what the dialogs are part of and may be empty (see newOperationLabel).
func processPickers(op promptOperation, parts *applyParts) bool {
	return parts.all(func(part applyPartOps) bool { return part.resolvePickers(op) })
}

func processPickerRows[T gurps.Node[T]](op promptOperation, rows []T) (revised []T, abort bool) {
	for _, one := range rows {
		result, cancel := processPickerRow(op, one)
		if cancel {
			return nil, true
		}
		revised = append(revised, result...)
	}
	return revised, false
}

func processPickerRow[T gurps.Node[T]](op promptOperation, row T) (revised []T, abort bool) {
	if !row.Container() {
		return []T{row}, false
	}
	children := row.NodeChildren()
	tpp, ok := any(row).(gurps.TemplatePickerProvider)
	var tp *gurps.TemplatePicker
	if ok {
		_, tp = tpp.TemplatePickerData()
	}
	if !ok || tp.IsZero() {
		rowChildren := make([]T, 0, len(children))
		for _, child := range children {
			var result []T
			result, abort = processPickerRow(op, child)
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
		row.SetChildren(rowChildren)
		SetParents(rowChildren, row)
		return []T{row}, false
	}

	headers := pickerRowDetailHeaders(row)
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  2 + len(headers),
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	if len(headers) != 0 {
		list.AddChild(unison.NewPanel())
		for _, header := range headers {
			label := unison.NewLabel()
			label.Font = fonts.FieldSecondary
			label.SetTitle(header)
			label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
			list.AddChild(label)
		}
		list.AddChild(unison.NewPanel())
	}

	progress := unison.NewLabel()
	progressBackground := pickerMatchStateColor(tp.Qualifier.Matches(0))
	progress.SetBorder(unison.NewCompoundBorder(
		unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}),
		unison.NewEmptyBorder(geom.NewHorizontalInsets(unison.StdHSpacing)),
	))
	progress.Side = side.Right
	progress.OnBackgroundInk = progressBackground.On()
	progress.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		if tp.Type == picker.NotApplicable {
			return
		}
		r := progress.ContentRect(true)
		r.Y += unison.StdVSpacing * 2
		r.Height -= unison.StdVSpacing * 2
		gc.DrawRoundedRect(r, geom.NewUniformSize(8), progressBackground.Paint(gc, r, paintstyle.Fill))
		progress.DefaultDraw(gc, r)
	}
	boxes := make([]*unison.CheckBox, 0, len(children))
	var dialog *unison.Dialog
	callback := func() {
		// A picked row that presents choices of its own has no single cost yet, so the running total can be a range.
		// The picker is satisfied while some way of making those remaining choices would satisfy it.
		total := gurps.NumericRangeOf(0)
		for i, box := range boxes {
			if box.State == check.On {
				switch tp.Type {
				case picker.NotApplicable:
				case picker.Count:
					total = total.Add(gurps.NumericRangeOf(fxp.One))
				case picker.Points:
					total = total.Add(pointsRangeFor(children[i]))
				case picker.Value, picker.Weight:
					total = total.Add(pickerMeasureRange(children[i], tp.Type))
				}
			}
		}
		matches := total.CanSatisfy(tp.Qualifier)
		dialog.Button(unison.ModalResponseOK).SetEnabled(matches)
		if tp.Type != picker.NotApplicable {
			var img *unison.SVG
			if matches {
				img = unison.CheckmarkSVG
			} else {
				img = svg.Not
			}
			size := max(progress.Font.Baseline()-2, 6)
			progress.Drawable = &unison.DrawableSVG{
				SVG:  img,
				Size: geom.NewSize(size, size),
			}
			progressBackground = pickerMatchStateColor(matches)
			progress.OnBackgroundInk = progressBackground.On()
			progress.SetTitle(formatPickerTotal(row, tp.Type, total))
			progress.MarkForLayoutRecursivelyUpward()
			progress.MarkForRedraw()
		}
	}
	for _, child := range children {
		boxes = addPickerRow(op, list, child, tp.Type, callback, boxes)
	}

	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HSpan:  2,
		HGrab:  true,
		VGrab:  true,
	})

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	if opLabel := newOperationLabel(op); opLabel != nil {
		opLabel.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		panel.AddChild(opLabel)
	}
	label := unison.NewLabel()
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	label.SetTitle(row.String())
	panel.AddChild(label)
	// A choice nested within another is put to the user only once its enclosing choice has been answered, so the
	// containers above it are named to tie it back to the answer that brought it up.
	if location := rowLocation(row); location != "" {
		label = newTruncatedLabel(location, 80, fonts.FieldSecondary)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		panel.AddChild(label)
	}
	if notesCapable, hasNotes := any(row).(interface{ Notes() string }); hasNotes {
		if notes := notesCapable.Notes(); notes != "" {
			label = unison.NewLabel()
			label.Font = fonts.FieldSecondary
			label.SetTitle(notes)
			label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
			panel.AddChild(label)
		}
	}
	label = unison.NewLabel()
	label.SetTitle(tp.StringWithUnits(pickerWeightUnits(row)))
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}))
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	panel.AddChild(label)
	progress.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	panel.AddChild(progress)
	panel.AddChild(scroll)

	var err error
	dialog, err = newPromptDialog(op.at(promptstep.Choice), nil, nil, panel,
		unison.NewCancelButtonInfo(),
		&unison.DialogButtonInfo{
			Title:        i18n.Text("Override"),
			ResponseCode: unison.ModalResponseUserBase,
		},
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return nil, true
	}
	overrideTip := i18n.Text("Accept the checked options whether or not they satisfy the choice")
	dialog.Button(unison.ModalResponseUserBase).Tooltip = newWrappedTooltip(overrideTip)
	callback()
	if dialog.RunModal() == unison.ModalResponseCancel {
		return nil, true
	}

	rowChildren := make([]T, 0, len(children))
	for i, box := range boxes {
		if box.State == check.On {
			var result []T
			result, abort = processPickerRow(op, children[i])
			if abort {
				return nil, true
			}
			rowChildren = append(rowChildren, result...)
		}
	}
	SetParents(rowChildren, row.Parent())
	return rowChildren, false
}

func pickerMatchStateColor(matches bool) unison.Color {
	if matches {
		return unison.Green
	}
	return unison.ThemeError.GetColor()
}

func addPickerRow[T gurps.Node[T]](op promptOperation, parent *unison.Panel, row T, pt picker.Type, callback func(), boxes []*unison.CheckBox) []*unison.CheckBox {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	// The wrapper fills its column so that the page reference it ends with lines up along the right edge.
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	parent.AddChild(wrapper)
	checkBox := unison.NewCheckBox()
	updatePickerCheckBoxTitle(checkBox, row, pt)
	checkBox.ClickCallback = callback
	wrapper.AddChild(checkBox)
	boxes = append(boxes, checkBox)
	var onClick func()
	var editTooltip string
	var details []*unison.Label
	pageRef := ""
	pageRefHighlight := ""
	switch actual := any(row).(type) {
	case *gurps.Trait:
		if actual.IsLeveled() {
			onClick = func() { pickerRowLevelEditor(op, actual, checkBox, pt, callback) }
			editTooltip = i18n.Text("Edit level")
		}
		pageRef = actual.PageRef
		pageRefHighlight = actual.PageRefHighlight
	case *gurps.Skill:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(op, actual, checkBox, pt, callback) }
			editTooltip = i18n.Text("Edit points")
		}
		pageRef = actual.PageRef
		pageRefHighlight = actual.PageRefHighlight
	case *gurps.Spell:
		if !actual.Container() {
			onClick = func() { pickerRowPointEditor(op, actual, checkBox, pt, callback) }
			editTooltip = i18n.Text("Edit points")
		}
		pageRef = actual.PageRef
		pageRefHighlight = actual.PageRefHighlight
	case *gurps.Equipment:
		// A choice made by value or weight may take more than one of an option, so its quantity may be set while
		// picking. A group has no quantity of its own to set.
		if (pt == picker.Value || pt == picker.Weight) && !actual.IsGroup() {
			onClick = func() { pickerRowQuantityEditor(op, actual, &details, callback) }
			editTooltip = i18n.Text("Edit quantity")
		}
		pageRef = actual.PageRef
		pageRefHighlight = actual.PageRefHighlight
	}
	if pageRef != "" {
		if pageRefs := ExtractPageReferences(pageRef); len(pageRefs) > 0 {
			var tooltip string
			var icon *unison.DrawableSVG
			title, img := convertLinksForPageRef(pageRefs[0])
			if img != nil {
				title = ""
				height := unison.DefaultLinkTheme.Font.Baseline()
				icon = &unison.DrawableSVG{
					SVG:  img,
					Size: geom.NewSize(height, height).Ceil(),
				}
				tooltip = pageRefs[0]
			}
			link := unison.NewLink(title, tooltip, "", &unison.DefaultLinkTheme, func(_ unison.Paneler, _ string) {
				OpenPageReference(pageRefs[0], pageRefHighlight, nil)
			})
			link.VAlign = align.Start
			link.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.End,
				HGrab:  true,
			})
			if icon != nil {
				link.Drawable = icon
			}
			if tooltip != "" {
				link.Tooltip = newWrappedTooltip(tooltip)
			}
			wrapper.AddChild(link)
		}
	}
	rowDetails := pickerRowDetails(row)
	details = make([]*unison.Label, 0, len(rowDetails))
	for _, detail := range rowDetails {
		label := unison.NewLabel()
		label.SetTitle(detail)
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
		parent.AddChild(label)
		details = append(details, label)
	}
	if onClick == nil {
		label := unison.NewLabel()
		label.SetTitle(" ")
		parent.AddChild(label)
	} else {
		button := NewSVGButtonForFont(svg.Edit, checkBox.Font, -2)
		button.Tooltip = newWrappedTooltip(editTooltip)
		button.ClickCallback = onClick
		parent.AddChild(button)
	}
	return boxes
}

// pickerRowDetailHeaders returns the headings of the columns of details shown for each option of the picker container,
// if its options have any. Only equipment has them: an option's quantity, and the value and weight of all of it.
func pickerRowDetailHeaders[T gurps.Node[T]](container T) []string {
	if _, ok := any(container).(*gurps.Equipment); ok {
		return []string{i18n.Text("Qty"), i18n.Text("Value"), i18n.Text("Weight")}
	}
	return nil
}

// pickerRowDetails returns the details shown for an option in the columns pickerRowDetailHeaders names. The value and
// weight are ranges when the option presents a choice of its own.
func pickerRowDetails[T gurps.Node[T]](row T) []string {
	eqp, ok := any(row).(*gurps.Equipment)
	if !ok {
		return nil
	}
	defUnits := pickerWeightUnits(eqp)
	quantity := ""
	if !eqp.IsGroup() {
		quantity = eqp.Quantity.Comma()
	}
	return []string{
		quantity,
		"$" + gurps.FormatValueRange(eqp.ExtendedValueRange(), fxp.Int.Comma),
		gurps.FormatWeightRange(eqp.ExtendedWeightRange(defUnits), defUnits.Format),
	}
}

// pickerMeasureRange returns what an option counts toward a choice made by value or weight: the range of its extended
// value or weight, which takes its quantity into account.
func pickerMeasureRange[T gurps.Node[T]](row T, pt picker.Type) gurps.NumericRange {
	eqp, ok := any(row).(*gurps.Equipment)
	if !ok || xreflect.IsNil(eqp) {
		return gurps.NumericRangeOf(0)
	}
	if pt == picker.Weight {
		return eqp.ExtendedWeightRange(pickerWeightUnits(eqp))
	}
	return eqp.ExtendedValueRange()
}

// pickerWeightUnits returns the units the picker dialog shows weights in: those of the sheet the row being picked from
// is headed for, since it is already owned by that sheet while the dialog is shown, or the default ones otherwise.
func pickerWeightUnits[T gurps.Node[T]](row T) fxp.WeightUnit {
	var entity *gurps.Entity
	if !xreflect.IsNil(row) {
		entity = gurps.EntityFromNode(row)
	}
	return gurps.SheetSettingsFor(entity).DefaultWeightUnits
}

// formatPickerTotal renders the running total of the options picked, as a value or a weight when the choice is made by
// one.
func formatPickerTotal[T gurps.Node[T]](row T, pt picker.Type, total gurps.NumericRange) string {
	switch pt {
	case picker.Value:
		return "$" + gurps.FormatValueRange(total, fxp.Int.Comma)
	case picker.Weight:
		return gurps.FormatWeightRange(total, pickerWeightUnits(row).Format)
	default:
		return total.Comma()
	}
}

// pickerRowQuantityEditor asks for a new quantity of an option of a choice made by value or weight, updating the
// option's details and the running total to match.
func pickerRowQuantityEditor(op promptOperation, eqp *gurps.Equipment, details *[]*unison.Label, callback func()) {
	quantity := eqp.Quantity
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Quantity"), eqp.String()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return quantity },
		func(value fxp.Int) { quantity = value }, fxp.One, fxp.Max-1, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Quantity), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	setPickerRowQuantity(eqp, quantity, *details)
	callback()
}

// setPickerRowQuantity sets the quantity of an option of a choice, updating the details shown for it to match.
func setPickerRowQuantity(eqp *gurps.Equipment, quantity fxp.Int, details []*unison.Label) {
	eqp.Quantity = quantity
	for i, detail := range pickerRowDetails(eqp) {
		if i < len(details) {
			details[i].SetTitle(detail)
			details[i].MarkForLayoutRecursivelyUpward()
			details[i].MarkForRedraw()
		}
	}
}

func updatePickerCheckBoxTitle[T gurps.Node[T]](checkBox *unison.CheckBox, row T, pt picker.Type) {
	title := row.String()
	switch pt {
	case picker.Points:
		// A row that presents choices of its own is worth a range rather than a single cost, which is worth showing
		// even though picking it leads to another dialog: it is what the row will add to the total.
		points := pointsRangeFor(row)
		value, settled := points.Settled()
		if !settled || value != 0 {
			pointsLabel := i18n.Text("points")
			if settled && value == fxp.One {
				pointsLabel = i18n.Text("point")
			}
			title += fmt.Sprintf(" [%s %s]", points.Comma(), pointsLabel)
		}
	case picker.Count:
		// NOP
	default:
		// NOP
	}
	checkBox.SetTitle(title)
}

func pickerRowLevelEditor(op promptOperation, trait *gurps.Trait, checkBox *unison.CheckBox, pt picker.Type, callback func()) {
	levels := trait.Levels
	maximum := trait.ResolvedMaxLevels()
	fieldMax := fxp.MaxBasePoints
	if maximum > 0 {
		fieldMax = maximum
		levels = min(levels, maximum)
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	if maximum > 0 {
		label.SetTitle(fmt.Sprintf(i18n.Text("%s Level (max %s)"), trait.Kind(), maximum.String()))
	} else {
		label.SetTitle(fmt.Sprintf(i18n.Text("%s Level"), trait.Kind()))
	}
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return levels },
		func(value fxp.Int) { levels = value }, 0, fieldMax, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Level), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	trait.Levels = levels
	updatePickerCheckBoxTitle(checkBox, trait, pt)
	callback()
	checkBox.MarkForLayoutRecursivelyUpward()
	checkBox.MarkForRedraw()
}

type pickerRowPointEditorTypes[T gurps.Node[T]] interface {
	gurps.Node[T]
	gurps.RawPointsAdjuster
}

func pickerRowPointEditor[T pickerRowPointEditorTypes[T]](op promptOperation, node T, checkBox *unison.CheckBox, pt picker.Type, callback func()) {
	points := node.RawPoints()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(fmt.Sprintf(i18n.Text("%s Points"), node.Kind()))
	panel.AddChild(label)
	panel.AddChild(NewDecimalField(nil, "", "", func() fxp.Int { return points },
		func(value fxp.Int) { points = value }, 0, fxp.MaxBasePoints, false, false))
	dialog, err := newPromptDialog(op.at(promptstep.Points), nil, nil, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo())
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	node.SetRawPoints(points)
	updatePickerCheckBoxTitle(checkBox, node, pt)
	callback()
	checkBox.MarkForLayoutRecursivelyUpward()
	checkBox.MarkForRedraw()
}

// pointsRangeFor returns the span of costs a picker may end up counting a row as being worth. A skill or spell is
// counted by its raw points, inside a container as much as on its own, since a picker counts what is being bought
// rather than what the destination sheet's bonuses make of it. The rows are already owned by that sheet by the time
// the picker is shown, so its bonuses would otherwise be counted. A trait is counted by its adjusted points, which is
// the only cost a trait has. Either way a container accounts for any choices it presents -- including the exact ones,
// which are worth what they ask for rather than what their children add up to.
func pointsRangeFor[T gurps.Node[T]](child T) gurps.NumericRange {
	if xreflect.IsNil(child) {
		return gurps.NumericRangeOf(0)
	}
	// Covers skills and spells
	if rp, ok := any(child).(interface{ RawPointsRange() gurps.NumericRange }); ok {
		return rp.RawPointsRange()
	}
	// Covers traits
	if rp, ok := any(child).(interface {
		PointsRange(tooltip *xbytes.InsertBuffer) gurps.NumericRange
	}); ok {
		return rp.PointsRange(nil)
	}
	return gurps.NumericRangeOf(0)
}
