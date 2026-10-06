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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/gcs/v5/ux/colors"
	"github.com/richardwilkes/gcs/v5/ux/fonts"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
)

// titledPagePanelInsets are the insets between the border of a sheet page block and its content.
var titledPagePanelInsets = geom.Insets{Top: 1, Left: 2, Bottom: 1, Right: 2}

// newTitledBlockBorder returns a sheet page block's titled border, without the standard insets. When there is a title,
// the border leaves the title strip to the block's heading (see addBlockHeading).
func newTitledBlockBorder(title string) *TitledBorder {
	return &TitledBorder{Title: title, HeadingInContent: title != ""}
}

// initTitledPagePanel sets up a panel as a sheet page block with the standard insets inside a titled border. A screen
// reader cannot read the border, so a non-empty title also becomes the block's accessible name and a heading (see
// addBlockHeading). When banded is true, the block's rows, each a label and field pair, are drawn with alternating
// backgrounds. See initPagePanel for the rest.
func initTitledPagePanel(p unison.Paneler, title string, columns int, banded bool, tint *unison.ThemeColor) (*unison.FlexLayout, *unison.FlexLayoutData) {
	border := newTitledBlockBorder(title)
	layout, layoutData := initPagePanel(p, unison.NewCompoundBorder(border, unison.NewEmptyBorder(titledPagePanelInsets)),
		columns, tint)
	panel := p.AsPanel()
	if title != "" {
		panel.Accessibility.Name = title
		addBlockHeading(panel, border, columns)
	}
	if banded {
		installBandedBackground(panel)
	}
	return layout, layoutData
}

// addBlockHeading adds the heading a screen reader reads a block's title as, as the block's first child, and returns
// it. The block's accessible name is spoken only as the focus enters the block, while a heading is found by heading
// navigation and can be a tab stop (see unison.SetFocusForReading). The heading draws nothing. The border leaves the
// title strip to the content (see TitledBorder.HeadingInContent), where the heading is laid out as a first row spanning
// the columns, so the content does not move; blockLayout then fits it to the painted strip. The rows start at index one
// (see blockRows), and a block that rebuilds them must use removeBlockRows. The block's layout must be the FlexLayout
// initPagePanel installs or a blockLayout from an earlier heading.
func addBlockHeading(panel *unison.Panel, border *TitledBorder, columns int) *unison.Panel {
	heading := newBlockHeading(border, columns)
	switch layout := panel.Layout().(type) {
	case *blockLayout:
		layout.border = border
		layout.heading = heading
	case *unison.FlexLayout:
		panel.SetLayout(&blockLayout{FlexLayout: layout, border: border, heading: heading})
	default:
		errs.Log(errs.Newf("a block heading needs a FlexLayout, not a %T", layout))
	}
	panel.AddChildAtIndex(heading, 0)
	return heading
}

// syncBlockHeading adds, renames or removes a block's heading to match its border's title, for a block whose title can
// change or be empty, and returns the heading, or nil when there is none. An untitled block's border keeps the title
// strip.
func syncBlockHeading(panel *unison.Panel, border *TitledBorder, columns int) *unison.Panel {
	heading := blockHeading(panel)
	border.HeadingInContent = border.Title != ""
	switch {
	case border.Title == "":
		if heading != nil {
			panel.RemoveChild(heading)
		}
		return nil
	case heading == nil:
		return addBlockHeading(panel, border, columns)
	default:
		heading.Accessibility.Name = border.Title
		return heading
	}
}

func blockHeading(panel *unison.Panel) *unison.Panel {
	if children := panel.Children(); len(children) > 0 && children[0].Accessibility.Role == role.Heading {
		return children[0]
	}
	return nil
}

func blockRowsStart(panel *unison.Panel) int {
	if blockHeading(panel) != nil {
		return 1
	}
	return 0
}

// blockRows returns a block's children after its heading. The slice is the block's own, so the block's children must
// not be added or removed while it is walked.
func blockRows(panel *unison.Panel) []*unison.Panel {
	return panel.Children()[blockRowsStart(panel):]
}

// blockLayout lays a titled block out with its FlexLayout, then puts the heading over the whole strip the border paints
// the title in. The FlexLayout's first row reserves the strip's height, but it lies inside the border's insets and is
// only as wide as the columns, which need not fill the block (the basic damage block centers them), and a focus ring
// around the heading should outline the whole title. Embedding the FlexLayout keeps a block's adjustments to the one
// initPagePanel returned in effect.
type blockLayout struct {
	*unison.FlexLayout
	border  *TitledBorder
	heading *unison.Panel
}

// PerformLayout implements unison.Layout.
func (l *blockLayout) PerformLayout(target *unison.Panel) {
	l.FlexLayout.PerformLayout(target)
	if l.heading != nil && l.heading.Parent() == target {
		l.heading.SetFrameRect(l.border.TitleStrip(target.FrameRect().Size))
	}
}

// removeBlockRows removes a block's children other than its heading. RemoveAllChildren would remove the heading too,
// and the first row would then be laid out under the title, since the border leaves the title strip to the content.
func removeBlockRows(panel *unison.Panel) {
	for _, child := range slices.Clone(blockRows(panel)) {
		panel.RemoveChild(child)
	}
}

// newBlockHeading returns the heading addBlockHeading adds. Its height is asked of the border at each layout, since a
// font change lays windows out again without rebuilding them.
func newBlockHeading(border *TitledBorder, columns int) *unison.Panel {
	heading := unison.NewPanel()
	heading.Accessibility.Role = role.Heading
	heading.Accessibility.Level = 1
	heading.Accessibility.Name = border.Title
	heading.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
		height := border.TitleHeight()
		return geom.NewSize(0, height), geom.NewSize(0, height), geom.NewSize(unison.DefaultMaxSize, height)
	})
	heading.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  columns,
		HAlign: align.Fill,
	})
	return heading
}

// installBandedBackground has the panel draw its rows, each two children wide, with alternating backgrounds. The rows
// start after the heading, which is looked for at each draw, since a heading may be added or removed later.
func installBandedBackground(panel *unison.Panel) {
	panel.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		drawBandedBackground(panel, gc, rect, blockRowsStart(panel), 2, nil)
	}
}

// initPagePanel sets up a panel as a block of a sheet page: it becomes its own Self and is given the border, a grid of
// the given number of columns with the standard page spacing, and layout data that fills its cell. A non-nil tint is
// installed as the block's tint. The layout and layout data are returned so that a block can adjust them.
func initPagePanel(p unison.Paneler, border unison.Border, columns int, tint *unison.ThemeColor) (*unison.FlexLayout, *unison.FlexLayoutData) {
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
// button. It returns the wrapper holding both, which is what gets added to the page, along with the label within it,
// which is what the field it names should set as its Accessibility.LabeledBy, since the wrapper is not itself a label.
// The button is left out of the Tab order unless the setting that puts static text there for screen readers is on (see
// tabStopForReading).
func NewPageLabelWithRandomizer(title, tooltip string, clickCallback func()) (wrapper *unison.Panel, label *unison.Label) {
	wrapper = unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	b := NewSVGButtonForFont(svg.Randomize, fonts.PageLabelPrimary, -2)
	tabStopForReading(b)
	if tooltip != "" {
		b.Tooltip = newWrappedTooltip(tooltip)
	}
	b.ClickCallback = clickCallback
	b.SetLayoutData(&unison.FlexLayoutData{HGrab: true})
	wrapper.AddChild(b)
	label = NewPageLabelEnd(title)
	wrapper.AddChild(label)
	return wrapper, label
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

// addLabeledStringPageField adds a label, built by newLabel (typically NewPageLabel or NewPageLabelEnd), and a text
// entry field for a sheet page to parent, in that order, and returns the field.
func addLabeledStringPageField(parent unison.Paneler, targetMgr *TargetMgr, targetKey, title string, newLabel func(string) *unison.Label, get func() string, set func(string)) *StringField {
	p := parent.AsPanel()
	p.AddChild(newLabel(title))
	field := NewStringPageField(targetMgr, targetKey, title, get, set)
	p.AddChild(field)
	return field
}

// addRandomizedStringPageField adds a label with a randomization button and a text entry field for a sheet page to
// parent, in that order, and returns the field. Clicking the button stores what random returns via set, then shows it
// in the field and marks the sheet modified.
func addRandomizedStringPageField(parent unison.Paneler, targetMgr *TargetMgr, targetKey, title, tooltip string, get func() string, set func(string), random func() string) *StringField {
	return addRandomizedPageField(parent, NewStringPageField(targetMgr, targetKey, title, get, set), title, tooltip,
		func() string {
			set(random())
			return get()
		})
}

// addRandomizedPageField adds a label with a randomization button and field to parent, in that order, and returns the
// field. Clicking the button calls randomize, which must store the new value and return the text to show for it, then
// shows that text in the field and marks the sheet modified. The field is flagged SkipDeepSync, since the values these
// fields hold have no bearing on the rest of the sheet.
func addRandomizedPageField[F SelectableTextField](parent unison.Paneler, field F, title, tooltip string, randomize func() string) F {
	p := parent.AsPanel()
	wrapper, label := NewPageLabelWithRandomizer(title, tooltip, func() { SetTextAndMarkModified(field, randomize()) })
	p.AddChild(wrapper)
	fp := field.AsPanel()
	fp.Accessibility.LabeledBy = label
	fp.ClientData()[SkipDeepSync] = true
	p.AddChild(field)
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
	// The field was synced with the exact text when it was created, before the display format was installed.
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
