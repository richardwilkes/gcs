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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/weight"
)

type linkSpec struct {
	pageRef   string
	highlight string
}

// calculatorContent is the column of sections a calculator dockable is built from, and the helpers that add rows to
// it. The per-sheet Calculator and the standalone collision calculator both embed it, so the two lay their sections
// out the same way: a bold header, its controls indented beneath it, and the results set off by a divider.
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
func (c *calculatorContent) addFieldRow(field unison.Paneler, trailing string) *unison.Label {
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
	cb := unison.NewCheckBox()
	cb.SetTitle(title)
	cb.SetBorder(newSectionIndent())
	cb.ClickCallback = func() {
		*flag = cb.State == check.On
		changed()
	}
	c.content.AddChild(cb)
	return cb
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

// addPlainLabel adds a label with the given text to the parent and returns it.
func addPlainLabel(parent *unison.Panel, text string) *unison.Label {
	label := unison.NewLabel()
	label.SetTitle(text)
	parent.AddChild(label)
	return label
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
