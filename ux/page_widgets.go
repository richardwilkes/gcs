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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// titledPagePanelInsets are the insets between the border of a sheet page block and its content.
var titledPagePanelInsets = geom.Insets{Top: 1, Left: 2, Bottom: 1, Right: 2}

// newTitledPageBorder returns the border of a titled sheet page block: the titled border with the standard insets
// inside it.
func newTitledPageBorder(title string) unison.Border {
	return unison.NewCompoundBorder(&TitledBorder{Title: title}, unison.NewEmptyBorder(titledPagePanelInsets))
}

// initTitledPagePanel sets up a panel as a sheet page block with the given title. See initPagePanel for the rest.
func initTitledPagePanel(p unison.Paneler, title string, columns int, banded bool, tint *unison.ThemeColor) (*unison.FlexLayout, *unison.FlexLayoutData) {
	return initPagePanel(p, newTitledPageBorder(title), columns, banded, tint)
}

// initPagePanel sets up a panel as a block of a sheet page: it becomes its own Self and is given the border, a grid of
// the given number of columns with the standard page spacing, and layout data that fills its cell. When banded is
// true, its rows, each of which pairs a label with a field, are drawn with alternating backgrounds. A tint that is not
// nil is installed as the block's tint. The layout and layout data are returned so that a block can adjust them,
// such as to grab extra space or to center its rows.
func initPagePanel(p unison.Paneler, border unison.Border, columns int, banded bool, tint *unison.ThemeColor) (*unison.FlexLayout, *unison.FlexLayoutData) {
	panel := p.AsPanel()
	panel.Self = p
	layout := &unison.FlexLayout{
		Columns:  columns,
		HSpacing: 4,
	}
	panel.SetLayout(layout)
	layoutData := &unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
	}
	panel.SetLayoutData(layoutData)
	panel.SetBorder(border)
	if banded {
		panel.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) { drawBandedBackground(panel, gc, rect, 0, 2, nil) }
	}
	if tint != nil {
		InstallTintFunc(panel, tint)
	}
	return layout, layoutData
}

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
	label.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
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

	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: 1})) // To match field underline spacing
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
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: 1})) // To match field underline spacing
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
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: 1})) // To match field underline spacing
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
	unison.InstallFocusBorders(
		field, field,
		unison.NewLineBorder(unison.ThemeFocus, geom.Size{}, geom.Insets{Bottom: 1}, false),
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Bottom: 1}, false),
	)
}

// NewHeightPageField creates a new height entry field for a sheet page.
func NewHeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Length, set func(fxp.Length), minValue, maxValue fxp.Length, noMinWidth bool) *LengthField {
	return newUnitsPageField(NewLengthField(targetMgr, targetKey, undoTitle, entity, get, set, minValue, maxValue, noMinWidth),
		func(v fxp.Length) string { return gurps.SheetSettingsFor(entity).FormatHeight(v) })
}

// NewWeightPageField creates a new weight entry field for a sheet page.
func NewWeightPageField(targetMgr *TargetMgr, targetKey, undoTitle string, entity *gurps.Entity, get func() fxp.Weight, set func(fxp.Weight), minValue, maxValue fxp.Weight, noMinWidth bool) *WeightField {
	return newUnitsPageField(NewWeightField(targetMgr, targetKey, undoTitle, entity, get, set, minValue, maxValue, noMinWidth),
		func(v fxp.Weight) string { return gurps.SheetSettingsFor(entity).FormatBodyWeight(v) })
}

// newUnitsPageField dresses a units field for a sheet page and installs the rendering to show while the field is not
// being edited.
func newUnitsPageField[T ~int64](field *NumericField[T], displayFormat func(T) string) *NumericField[T] {
	field.DisplayFormat = displayFormat
	installPageFieldFontAndFocusBorders(field.Field)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	// The field was synced with the exact text when it was created, before the display format was installed, so sync it
	// again to pick that up.
	field.Sync()
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
		field.SetMinimumTextWidthUsing(minValue.Floor().String(), maxValue.Floor().String())
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}
