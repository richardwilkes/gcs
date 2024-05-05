/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var nonEditableFieldColor = unison.NewDynamicColor(func() unison.Color {
	return unison.PrimaryTheme.OnSurface.GetColor().SetAlphaIntensity(0.6)
})

// NewPageHeader creates a new center-aligned header for a sheet page.
func NewPageHeader(title string, hSpan int) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = gurps.OnHeaderColor
	label.Text = title
	label.Font = gurps.PageLabelPrimaryFont
	label.HAlign = align.Middle
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, gurps.HeaderColor.Paint(gc, rect, paintstyle.Fill))
		label.DefaultDraw(gc, rect)
	}
	return label
}

// NewPageInternalHeader creates a new center-aligned internal header for a sheet page.
func NewPageInternalHeader(title string, span int) unison.Paneler {
	layoutData := &unison.FlexLayoutData{
		HSpan:  span,
		HAlign: align.Fill,
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
	label.Font = gurps.PageLabelSecondaryFont
	label.HAlign = align.Middle
	label.OnBackgroundInk = &unison.PrimaryTheme.OnSurface
	label.SetLayoutData(layoutData)
	label.SetBorder(border)
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		_, pref, _ := label.Sizes(unison.Size{})
		paint := unison.PrimaryTheme.Outline.Paint(gc, rect, paintstyle.Stroke)
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
	label.OnBackgroundInk = &unison.PrimaryTheme.OnSurface
	label.Text = title
	label.Font = gurps.PageLabelPrimaryFont
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: 1})) // To match field underline spacing
	return label
}

// NewPageLabelEnd creates a new end-aligned field label for a sheet page.
func NewPageLabelEnd(title string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = &unison.PrimaryTheme.OnSurface
	label.Text = title
	label.Font = gurps.PageLabelPrimaryFont
	label.HAlign = align.End
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: 1})) // To match field underline spacing
	return label
}

// NewPageLabelCenter creates a new center-aligned field label for a sheet page.
func NewPageLabelCenter(title string) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = &unison.PrimaryTheme.OnSurface
	label.Text = title
	label.Font = gurps.PageLabelPrimaryFont
	label.HAlign = align.Middle
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
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
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	b := NewSVGButtonForFont(svg.Randomize, gurps.PageLabelPrimaryFont, -2)
	b.SetFocusable(false)
	if tooltip != "" {
		b.Tooltip = newWrappedTooltip(tooltip)
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
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	return field
}

func installPageFieldFontAndFocusBorders(field *unison.Field) {
	field.Font = gurps.PageFieldPrimaryFont
	unison.InstallFocusBorders(field, field,
		unison.NewLineBorder(&unison.PrimaryTheme.Primary, 0, unison.Insets{Bottom: 1}, false),
		unison.NewLineBorder(&unison.PrimaryTheme.Outline, 0, unison.Insets{Bottom: 1}, false),
	)
}

// NewStringPageFieldNoGrab creates a new text entry field for a sheet page, but with HGrab set to false.
func NewStringPageFieldNoGrab(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	field := NewStringField(targetMgr, targetKey, undoTitle, get, set)
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}

// NewHeightPageField creates a new height entry field for a sheet page.
func NewHeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Length, set func(fxp.Length), minValue, maxValue fxp.Length, noMinWidth bool) *LengthField {
	field := NewLengthField(targetMgr, targetKey, undoTitle, entity, get, set, minValue, maxValue, noMinWidth)
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}

// NewWeightPageField creates a new weight entry field for a sheet page.
func NewWeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Weight, set func(fxp.Weight), minValue, maxValue fxp.Weight, noMinWidth bool) *WeightField {
	field := NewWeightField(targetMgr, targetKey, undoTitle, entity, get, set, minValue, maxValue, noMinWidth)
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}

// NewIntegerPageField creates a new integer entry field for a sheet page.
func NewIntegerPageField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() int, set func(int), minValue, maxValue int, showSign, noMinWidth bool) *IntegerField {
	field := NewIntegerField(targetMgr, targetKey, undoTitle, get, set, minValue, maxValue, showSign, noMinWidth)
	field.HAlign = align.End
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}

// NewDecimalPageField creates a new numeric text entry field for a sheet page.
func NewDecimalPageField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() fxp.Int, set func(fxp.Int), minValue, maxValue fxp.Int, noMinWidth bool) *DecimalField {
	field := NewDecimalField(targetMgr, targetKey, undoTitle, get, set, minValue, maxValue, false, noMinWidth)
	field.HAlign = align.End
	installPageFieldFontAndFocusBorders(field.Field)
	if !noMinWidth && minValue != fxp.Min && maxValue != fxp.Max {
		// Override to ignore fractional values
		field.SetMinimumTextWidthUsing(minValue.Trunc().String(), maxValue.Trunc().String())
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}
