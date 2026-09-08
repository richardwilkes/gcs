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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/pathop"
	"github.com/richardwilkes/unison/enums/weight"
)

// calculatorFieldPrototype is the widest text every numeric field in a calculator is sized to hold. The fields sit in
// rows of their own, so left to themselves they would each take a width from their own range; sizing them all to the
// same text lines them up down the column.
const calculatorFieldPrototype = "-99,999.99"

type linkSpec struct {
	pageRef   string
	highlight string
}

// calculatorContent is the column of sections a calculator is built from, and the helpers that add rows to it. Every
// calculator the Calculator dockable shows embeds it, so they all lay their sections out the same way: a bold header,
// its controls indented beneath it, and the results set off in a box of their own.
type calculatorContent struct {
	content    *unison.Panel
	resultsBox *unison.Panel
	flush      bool // Whether rows sit flush with the content's edge rather than indented beneath a header.
}

// initCalculatorContent creates the content panel and gives it the margin and the column layout that the sections are
// added into. The Calculator's slot gives the content the full width of the view, so that the results box can span it.
func (c *calculatorContent) initCalculatorContent() {
	c.content = unison.NewPanel()
	c.content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	c.content.SetLayout(&columnLayout{VSpacing: unison.StdVSpacing})
}

// columnLayout stacks its children top to bottom, each as wide as the column, with VSpacing between them. The column
// takes whatever width it is given, so that a results box spans the view and the notes wrap to it, but never reports
// a width narrower than its widest child needs, so that the scroll panel around a calculator scrolls sideways rather
// than clipping rows. A single-column FlexLayout would not do: a child that grabs the width sets no floor on how far
// the column can shrink, and one that does not cannot be stretched to fill it.
type columnLayout struct {
	VSpacing float32
}

// LayoutSizes implements unison.Layout.
func (l *columnLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	var insets geom.Insets
	if b := target.Border(); b != nil {
		insets = b.Insets()
	}
	children := target.Children()
	// The floor is the widest any child needs when it is not offered a width to fit. A child that grabs the width, as
	// a row of wrapping notes does, takes whatever it is given and sets no floor, just as it would in a FlexLayout.
	var floor float32
	for _, child := range children {
		if data, ok := child.LayoutData().(*unison.FlexLayoutData); ok && data.HGrab {
			continue
		}
		childMin, _, _ := child.Sizes(geom.Size{})
		floor = max(floor, childMin.Width)
	}
	width := max(hint.Width-insets.Width(), floor)
	var height float32
	for i, child := range children {
		if i > 0 {
			height += l.VSpacing
		}
		_, childPref, _ := child.Sizes(geom.NewSize(width, 0))
		height += childPref.Height
	}
	prefSize = geom.NewSize(width, height).Add(insets.Size())
	minSize = geom.NewSize(floor, height).Add(insets.Size())
	return minSize, prefSize, unison.MaxSize(prefSize)
}

// PerformLayout implements unison.Layout.
func (l *columnLayout) PerformLayout(target *unison.Panel) {
	rect := target.ContentRect(false)
	y := rect.Y
	for i, child := range target.Children() {
		if i > 0 {
			y += l.VSpacing
		}
		_, childPref, _ := child.Sizes(geom.NewSize(rect.Width, 0))
		// A child wider than the column is laid out at its own width, so that it is cut off rather than crushed.
		child.SetFrameRect(geom.NewRect(rect.X, y, max(rect.Width, childPref.Width), childPref.Height))
		y += childPref.Height
	}
}

// newSectionIndent returns the border that sets a section's controls in from its header.
func newSectionIndent() unison.Border {
	return unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2})
}

// addRow adds a row of controls with the given number of columns to the content, indented beneath its section's
// header unless the content is flush, and returns it.
func (c *calculatorContent) addRow(columns int) *unison.Panel {
	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	if !c.flush {
		row.SetBorder(newSectionIndent())
	}
	c.content.AddChild(row)
	return row
}

// addFieldRow adds a row holding the field followed by a label with the given text, and returns the label so that a
// caller passing no text can fill it in later.
func (c *calculatorContent) addFieldRow(field unison.Paneler, trailing string) *textLabel {
	row := c.addRow(2)
	row.AddChild(field)
	return addPlainLabel(row, trailing)
}

// resultsBoxMargin is the space between a calculator's inputs and the box its results are shown in.
const resultsBoxMargin = unison.StdVSpacing * 2

// addResultsBox adds the box a calculator's results are shown in, which sets them off from the inputs above it: a
// rounded box on a surface of its own, with a line around it and a "Results" header band across its top in the header
// colors. It returns the body of the box as content for the rows of results to be added to, flush with the body's
// padding so that they line up with the indented rows above the box, and remembers the box in resultsBox.
func (c *calculatorContent) addResultsBox() *calculatorContent {
	box := unison.NewPanel()
	box.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: resultsBoxMargin}))
	box.SetLayout(&columnLayout{})
	c.content.AddChild(box)
	c.resultsBox = box

	// The text takes its ink from the label as the title is set, so the ink comes first.
	header := unison.NewLabel()
	header.Font = subheaderFont
	header.HAlign = align.Middle
	header.OnBackgroundInk = colors.OnHeader
	header.SetTitle(i18n.Text("Results"))
	header.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
	box.AddChild(header)

	body := unison.NewPanel()
	body.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	body.SetLayout(&columnLayout{VSpacing: unison.StdVSpacing})
	box.AddChild(body)

	box.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := box.ContentRect(true)
		r.Y += resultsBoxMargin
		r.Height -= resultsBoxMargin
		radius := geom.NewUniformSize(8)
		gc.DrawRoundedRect(r, radius, unison.ThemeAboveSurface.Paint(gc, r, paintstyle.Fill))
		// The header's band runs the full width of the box, tucked under its rounded top corners, and the line along
		// its bottom does the same. The label draws only its text, so the band shows through behind it.
		band := r
		band.Height = header.FrameRect().Bottom() - r.Y
		gc.Save()
		clip := unison.NewPath()
		clip.RoundedRect(r, radius)
		gc.ClipPath(clip, pathop.Intersect, true)
		gc.DrawRect(band, colors.Header.Paint(gc, band, paintstyle.Fill))
		gc.Restore()
		edge := unison.ThemeSurfaceEdge.Paint(gc, r, paintstyle.Stroke)
		edge.SetStrokeWidth(1)
		gc.DrawLine(geom.NewPoint(band.X, band.Bottom()-0.5), geom.NewPoint(band.Right(), band.Bottom()-0.5), edge)
		gc.DrawRoundedRect(r.Inset(geom.NewUniformInsets(0.5)), geom.NewUniformSize(7.5), edge)
	}
	return &calculatorContent{content: body, flush: true}
}

// addResultRow adds the results box and a two-column row inside it for a calculator's results, and returns the row.
func (c *calculatorContent) addResultRow() *unison.Panel {
	return c.addResultsBox().addRow(2)
}

// addCheckBox adds an indented checkbox with the given title to the content. Clicking it stores whether it is now
// checked in *flag, then runs changed.
func (c *calculatorContent) addCheckBox(title string, flag *bool, changed func()) *unison.CheckBox {
	cb := newCheckBox(title, flag, changed)
	cb.SetBorder(newSectionIndent())
	c.content.AddChild(cb)
	return cb
}

// newCheckBox returns a checkbox with the given title that stores whether it is checked in *flag, then runs changed,
// for a caller that places it in a row of its own.
func newCheckBox(title string, flag *bool, changed func()) *unison.CheckBox {
	cb := unison.NewCheckBox()
	cb.SetTitle(title)
	cb.ClickCallback = func() {
		*flag = cb.State == check.On
		changed()
	}
	return cb
}

// newRowGroup returns a panel to hold a group of rows that come and go together as the choices change, or a slot that
// holds whichever of those groups is in use. A group is added to and removed from its slot rather than hidden, since
// a layout gives a hidden child its place just the same, so a hidden group would leave a gap its own size.
func newRowGroup() *unison.Panel {
	group := unison.NewPanel()
	group.SetLayout(&columnLayout{VSpacing: unison.StdVSpacing})
	return group
}

// fillSlot makes the slot hold exactly the given panels, in order.
func fillSlot(slot *unison.Panel, panels ...*unison.Panel) {
	current := slot.Children()
	same := len(current) == len(panels)
	if same {
		for i, p := range panels {
			if current[i] != p {
				same = false
				break
			}
		}
	}
	if same {
		return
	}
	slot.RemoveAllChildren()
	for _, p := range panels {
		slot.AddChild(p)
	}
}

// subheaderFont is the bold label font the parts a calculator's content is divided into are named in.
var subheaderFont = &unison.DynamicFont{
	Resolver: func() unison.FontDescriptor {
		desc := unison.LabelFont.Descriptor()
		desc.Weight = weight.Bold
		return desc
	},
}

// newSubheader returns a bold label naming one of the parts a calculator's content is divided into.
func newSubheader(text string) *unison.Label {
	label := unison.NewLabel()
	label.Font = subheaderFont
	label.SetTitle(text)
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}))
	return label
}

// addSubheader adds a subheader with the given text to the content and returns it.
func (c *calculatorContent) addSubheader(text string) *unison.Label {
	label := newSubheader(text)
	c.content.AddChild(label)
	return label
}

// sameWidth sizes the field to calculatorFieldPrototype and returns it.
func sameWidth[T xmath.Integer | xmath.Float](field *NumericField[T]) *NumericField[T] {
	field.SetMinimumTextWidthUsing(calculatorFieldPrototype)
	return field
}

// addIndexPopup adds a popup offering the items to the parent, with the one at *index selected, and returns it.
// Choosing an item stores its position in *index, then runs changed.
func addIndexPopup[T comparable](parent *unison.Panel, items []T, index *int, changed func()) *unison.PopupMenu[T] {
	popup := unison.NewPopupMenu[T]()
	popup.AddItem(items...)
	popup.SelectIndex(*index)
	popup.SelectionChangedCallback = func(_ *unison.PopupMenu[T]) {
		*index = popup.SelectedIndex()
		changed()
	}
	parent.AddChild(popup)
	return popup
}

// addPlainLabel adds a single-line label with the given text to the parent and returns it. Any page reference in the
// text, such as "(BX400)", is a link that opens the page.
func addPlainLabel(parent *unison.Panel, text string) *textLabel {
	label := newSingleLineLabel()
	label.linkPageRefs(openPageRefLink)
	label.SetTitle(text)
	parent.AddChild(label)
	return label
}

// openPageRefLink opens the page a link in a label refers to.
func openPageRefLink(ref string) {
	OpenPageReference(ref, "", nil)
}

// addResultLabel adds a bold label for showing a result to the parent and returns it.
func addResultLabel(parent *unison.Panel) *unison.Label {
	label := unison.NewLabel()
	label.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.DefaultLabelTheme.Font.Descriptor()
			desc.Weight = weight.Bold
			return desc
		},
	}
	parent.AddChild(label)
	return label
}

// addResult adds a labeled result to the parent.
func addResult(parent *unison.Panel, label, value string) {
	addPlainLabel(parent, label)
	addResultLabel(parent).SetTitle(value)
}

// newNoteRow returns a bulleted note that wraps to the width it is given, with its lines hanging under the first. The
// page references it cites, such as "(BX377)", are links that open the page, as the ones in a section header are.
func newNoteRow(note string) *unison.Panel {
	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	row.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	bullet := unison.NewLabel()
	bullet.SetTitle("•")
	bullet.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
	row.AddChild(bullet)
	text := newWrappingLabel()
	text.linkPageRefs(openPageRefLink)
	text.setText(note, unison.DefaultLabelTheme.OnBackgroundInk)
	text.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	row.AddChild(text)
	return row
}

// addNotes adds a bulleted note for each of the given texts, indented beneath the section they belong to.
func (c *calculatorContent) addNotes(notes ...string) {
	group := newRowGroup()
	group.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2, Left: unison.StdHSpacing * 2}))
	for _, note := range notes {
		group.AddChild(newNoteRow(note))
	}
	c.content.AddChild(group)
}

// addResultsSection adds the results box, then inside it a two-column panel for results that are rewritten as a whole
// and a group for the notes that go with them, and returns those.
func (c *calculatorContent) addResultsSection() (results, notes *unison.Panel) {
	box := c.addResultsBox()
	results = box.addRow(2)
	results.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing * 2, VSpacing: unison.StdVSpacing})
	return results, box.addNotesGroup()
}

// addNotesGroup adds a group for the notes that go with a calculator's results, set a little apart from what is above
// it, and returns it for setNotes to fill.
func (c *calculatorContent) addNotesGroup() *unison.Panel {
	notes := newRowGroup()
	notes.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}))
	c.content.AddChild(notes)
	return notes
}

// setNotes makes the notes panel hold a bulleted row for each of the notes, skipping empty ones.
func setNotes(panel *unison.Panel, notes []string) {
	panel.RemoveAllChildren()
	for _, note := range notes {
		if note != "" {
			panel.AddChild(newNoteRow(note))
		}
	}
}

// createHeader returns a section header holding the text, followed by the page references in parentheses, with the
// given amount of empty space above it.
func (c *calculatorContent) createHeader(text string, linkSpecs []linkSpec, topMargin float32) *unison.Panel {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1 + 2*len(linkSpecs)})
	if topMargin > 0 {
		wrapper.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: topMargin}))
	}

	first := unison.NewLabel()
	first.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Size += 2
			desc.Weight = weight.Bold
			return desc
		},
	}
	if len(linkSpecs) > 0 {
		first.SetTitle(text + " (")
		wrapper.AddChild(first)
	} else {
		first.SetTitle(text)
		wrapper.AddChild(first)
		return wrapper
	}

	linkTheme := unison.DefaultLinkTheme
	linkTheme.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Weight = weight.Bold
			return desc
		},
	}
	for index, linkSpec := range linkSpecs {
		link := unison.NewLink(linkSpec.pageRef, "", linkSpec.pageRef, &linkTheme, func(_ unison.Paneler, _ string) {
			OpenPageReference(linkSpec.pageRef, linkSpec.highlight, nil)
		})
		wrapper.AddChild(link)
		if index < len(linkSpecs)-1 {
			comma := unison.NewLabel()
			comma.Font = first.Font
			comma.SetTitle(", ")
			wrapper.AddChild(comma)
		}
	}

	last := unison.NewLabel()
	last.Font = first.Font
	last.SetTitle(")")
	wrapper.AddChild(last)
	return wrapper
}

// useMetersFor returns true if lengths for the given entity should be shown in metric units. A nil entity uses the
// global default sheet settings.
func useMetersFor(entity *gurps.Entity) bool {
	units := gurps.SheetSettingsFor(entity).DefaultLengthUnits
	return units == fxp.Centimeter || units == fxp.Meter || units == fxp.Kilometer
}

// lengthToText formats the given length, expressed in inches, using the length units the given entity prefers. A nil
// entity uses the global default sheet settings.
func lengthToText(entity *gurps.Entity, inches fxp.Int) string {
	var buffer strings.Builder
	if useMetersFor(entity) {
		meters := fxp.Meter.FromInches(inches).Mul(fxp.Hundred).Round().Div(fxp.Hundred)
		if meters == fxp.One {
			buffer.WriteString(i18n.Text("1 meter"))
		} else {
			fmt.Fprintf(&buffer, i18n.Text("%s meters"), meters.Comma())
		}
	} else {
		if inches >= fxp.ThirtySix {
			yards := inches.Div(fxp.ThirtySix).Floor()
			if yards == fxp.One {
				buffer.WriteString(i18n.Text("1 yard"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s yards"), yards.Comma())
			}
			inches -= yards.Mul(fxp.ThirtySix)
		}
		if inches >= fxp.Twelve {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			feet := inches.Div(fxp.Twelve).Floor()
			if feet == fxp.One {
				buffer.WriteString(i18n.Text("1 foot"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s feet"), feet.Comma())
			}
			inches -= feet.Mul(fxp.Twelve)
		}
		if inches > 0 || buffer.Len() == 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			if inches == fxp.One {
				buffer.WriteString(i18n.Text("1 inch"))
			} else {
				fmt.Fprintf(&buffer, i18n.Text("%s inches"), inches.Comma())
			}
		}
	}
	return buffer.String()
}
