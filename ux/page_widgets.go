/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
)

var nonEditableFieldColor = unison.NewDynamicColor(func() unison.Color {
	return unison.OnEditableColor.GetColor().SetAlphaIntensity(0.6)
})

// NewPageHeader creates a new center-aligned header for a sheet page.
func NewPageHeader(title string, hSpan int) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = model.OnHeaderColor
	label.Text = title
	label.Font = model.PageLabelPrimaryFont
	label.HAlign = unison.MiddleAlignment
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, model.HeaderColor.Paint(gc, rect, unison.Fill))
		label.DefaultDraw(gc, rect)
	}
	return label
}

// NewPageInternalHeader creates a new center-aligned internal header for a sheet page.
func NewPageInternalHeader(title string, span int) unison.Paneler {
	layoutData := &unison.FlexLayoutData{
		HSpan:  span,
		HAlign: unison.FillAlignment,
	}
	border := unison.NewEmptyBorder(unison.NewVerticalInsets(2))
	if title == "" {
		sep := unison.NewSeparator()
		sep.SetBorder(border)
		sep.SetLayoutData(layoutData)
		return sep
	}
	label := unison.NewLabel()
	label.Text = title
	label.Font = model.PageLabelSecondaryFont
	label.HAlign = unison.MiddleAlignment
	label.OnBackgroundInk = unison.OnContentColor
	label.SetLayoutData(layoutData)
	label.SetBorder(border)
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		_, pref, _ := label.Sizes(unison.Size{})
		paint := unison.DividerColor.Paint(gc, rect, unison.Stroke)
		paint.SetStrokeWidth(1)
		half := (rect.Width - pref.Width) / 2
		gc.DrawLine(rect.X, rect.CenterY(), rect.X+half-2, rect.CenterY(), paint)
		gc.DrawLine(2+rect.Right()-half, rect.CenterY(), rect.Right(), rect.CenterY(), paint)
		label.DefaultDraw(gc, rect)
	}
	return label
}

// NewPageLabel creates a new start-aligned field label for a sheet page.
func NewPageLabel(title string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = unison.OnContentColor
	label.Text = title
	label.Font = model.PageLabelPrimaryFont
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: 1})) // To match field underline spacing
	return label
}

// NewPageLabelEnd creates a new end-aligned field label for a sheet page.
func NewPageLabelEnd(title string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = unison.OnContentColor
	label.Text = title
	label.Font = model.PageLabelPrimaryFont
	label.HAlign = unison.EndAlignment
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: 1})) // To match field underline spacing
	return label
}

// NewPageLabelCenter creates a new center-aligned field label for a sheet page.
func NewPageLabelCenter(title string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = unison.OnContentColor
	label.Text = title
	label.Font = model.PageLabelPrimaryFont
	label.HAlign = unison.MiddleAlignment
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: 1})) // To match field underline spacing
	return label
}

// NewPageLabelWithRandomizer creates a new end-aligned field label for a sheet page that includes a randomization
// button.
func NewPageLabelWithRandomizer(title, tooltip string, clickCallback func()) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	b := unison.NewButton()
	b.ButtonTheme = unison.DefaultSVGButtonTheme
	b.DrawableOnlyVMargin = 1
	b.DrawableOnlyHMargin = 1
	b.HideBase = true
	b.SetFocusable(false)
	baseline := model.PageLabelPrimaryFont.Baseline()
	size := unison.NewSize(baseline, baseline)
	b.Drawable = &unison.DrawableSVG{
		SVG:  svg.Randomize,
		Size: *size.GrowToInteger(),
	}
	if tooltip != "" {
		b.Tooltip = unison.NewTooltipWithText(tooltip)
	}
	b.ClickCallback = clickCallback
	b.SetLayoutData(&unison.FlexLayoutData{HGrab: true})
	wrapper.AddChild(b)
	wrapper.AddChild(NewPageLabelEnd(title))
	return wrapper
}

// NewStringPageField creates a new text entry field for a sheet page.
func NewStringPageField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	field := NewStringField(targetMgr, targetKey, undoTitle, get, set)
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})
	return field
}

// NewStringPageFieldNoGrab creates a new text entry field for a sheet page, but with HGrab set to false.
func NewStringPageFieldNoGrab(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	field := NewStringField(targetMgr, targetKey, undoTitle, get, set)
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return field
}

// NewHeightPageField creates a new height entry field for a sheet page.
func NewHeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *model.Entity, get func() model.Length, set func(model.Length), min, max model.Length, noMinWidth bool) *LengthField {
	field := NewLengthField(targetMgr, targetKey, undoTitle, entity, get, set, min, max, noMinWidth)
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return field
}

// NewWeightPageField creates a new weight entry field for a sheet page.
func NewWeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *model.Entity, get func() model.Weight, set func(model.Weight), min, max model.Weight, noMinWidth bool) *WeightField {
	field := NewWeightField(targetMgr, targetKey, undoTitle, entity, get, set, min, max, noMinWidth)
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return field
}

// NewIntegerPageField creates a new integer entry field for a sheet page.
func NewIntegerPageField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() int, set func(int), min, max int, showSign, noMinWidth bool) *IntegerField {
	field := NewIntegerField(targetMgr, targetKey, undoTitle, get, set, min, max, showSign, noMinWidth)
	field.HAlign = unison.EndAlignment
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return field
}

// NewDecimalPageField creates a new numeric text entry field for a sheet page.
func NewDecimalPageField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() fxp.Int, set func(fxp.Int), min, max fxp.Int, noMinWidth bool) *DecimalField {
	field := NewDecimalField(targetMgr, targetKey, undoTitle, get, set, min, max, false, noMinWidth)
	field.HAlign = unison.EndAlignment
	field.Font = model.PageFieldPrimaryFont
	field.FocusedBorder = unison.NewLineBorder(unison.AccentColor, 0, unison.Insets{Bottom: 1}, false)
	field.UnfocusedBorder = unison.NewLineBorder(unison.ControlEdgeColor, 0, unison.Insets{Bottom: 1}, false)
	field.SetBorder(field.UnfocusedBorder)
	if !noMinWidth && min != fxp.Min && max != fxp.Max {
		// Override to ignore fractional values
		field.SetMinimumTextWidthUsing(min.Trunc().String(), max.Trunc().String())
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return field
}
