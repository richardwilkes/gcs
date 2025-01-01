// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// NewPageHeader creates a new center-aligned header for a sheet page.
func NewPageHeader(title string, hSpan int) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = colors.OnHeader
	label.Text = unison.NewSmallCapsText(title, &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: colors.OnHeader,
	})
	label.HAlign = align.Middle
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hSpan,
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, colors.Header.Paint(gc, rect, paintstyle.Fill))
		label.DefaultDraw(gc, rect)
	}
	return label
}

// NewPageLabel creates a new start-aligned field label for a sheet page.
func NewPageLabel(title string) *unison.Label {
	return NewPageLabelWithInk(title, unison.ThemeOnSurface)
}

// NewPageLabelWithInk creates a new start-aligned field label for a sheet page with the given Ink.
func NewPageLabelWithInk(title string, ink unison.Ink) *unison.Label {
	label := unison.NewLabel()
	label.Text = unison.NewSmallCapsText(title, &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: ink,
	})
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
	label.Text = unison.NewSmallCapsText(title, &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: unison.ThemeOnSurface,
	})
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
	label.Text = unison.NewSmallCapsText(title, &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: unison.ThemeOnSurface,
	})
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
	b := NewSVGButtonForFont(svg.Randomize, fonts.PageLabelPrimary, -2)
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
	field.Font = fonts.PageFieldPrimary
	unison.InstallFocusBorders(field, field,
		unison.NewLineBorder(unison.ThemeFocus, 0, unison.Insets{Bottom: 1}, false),
		unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1}, false),
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
