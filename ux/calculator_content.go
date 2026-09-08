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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
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

// calculatorContent is the column of sections a calculator dockable is built from, and the helpers that add rows to
// it. The per-sheet Calculator and the standalone calculators all embed it, so they lay their sections out the same
// way: a bold header, its controls indented beneath it, and the results set off by a divider.
type calculatorContent struct {
	content *unison.Panel
}

// initCalculatorContent creates the content panel and gives it the margin and single-column layout that the sections
// are added into.
func (c *calculatorContent) initCalculatorContent() {
	c.content = unison.NewPanel()
	c.content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	c.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
}

// newSectionIndent returns the border that sets a section's controls in from its header.
func newSectionIndent() unison.Border {
	return unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2})
}

// addRow adds a row of controls with the given number of columns to the content, indented beneath its section's
// header, and returns it.
func (c *calculatorContent) addRow(columns int) *unison.Panel {
	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	row.SetBorder(newSectionIndent())
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

// addResultRow adds a two-column row for a section's results, set off from the inputs above it by a divider, and
// returns it.
func (c *calculatorContent) addResultRow() *unison.Panel {
	row := c.addRow(2)
	divider := unison.NewSeparator()
	divider.SetBorder(unison.NewEmptyBorder(geom.NewVerticalInsets(unison.StdVSpacing * 2)))
	divider.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	row.AddChild(divider)
	return row
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
// unison's FlexLayout gives a hidden child a cell just the same, so a hidden group would leave a gap its own size.
func newRowGroup() *unison.Panel {
	group := unison.NewPanel()
	group.SetLayout(&unison.FlexLayout{Columns: 1, HSpacing: unison.StdHSpacing, VSpacing: unison.StdVSpacing})
	group.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
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

// newSubheader returns a bold label naming one of the parts a calculator's content is divided into.
func newSubheader(text string) *unison.Label {
	label := unison.NewLabel()
	label.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			desc := unison.LabelFont.Descriptor()
			desc.Weight = weight.Bold
			return desc
		},
	}
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
