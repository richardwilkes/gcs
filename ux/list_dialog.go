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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// showListQuestionDialog puts up a modal question dialog whose content is list, wrapped in a bordered scroll panel
// that takes whatever room the dialog offers, beneath a label showing header and then any extraHeaders, in order. The
// dialog is titled for the operation, whose description, if it has one, is shown above the header (see
// promptOperation). It reports whether the dialog was accepted.
func showListQuestionDialog(op promptOperation, header string, list unison.Paneler, extraHeaders ...*unison.Label) bool {
	panel, _ := newListQuestionPanel(op, header, list, extraHeaders...)
	return runPromptDialog(op, panel, unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfo()) == unison.ModalResponseOK
}

// newListQuestionPanel builds the content panel showListQuestionDialog shows: the operation's description, when it has
// one, then the header labels, stacked over the list's scroll panel in a single column, with the scroll panel the only
// child that grows to fill the dialog.
func newListQuestionPanel(op promptOperation, header string, list unison.Paneler, extraHeaders ...*unison.Label) (panel *unison.Panel, scroll *unison.ScrollPanel) {
	scroll = unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel = unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	if opLabel := newOperationLabel(op); opLabel != nil {
		panel.AddChild(opLabel)
	}
	label := unison.NewLabel()
	label.SetTitle(header)
	panel.AddChild(label)
	// The list sits inside a scroll panel, so it does not follow the header label as a sibling would.
	if lp := list.AsPanel(); lp.Accessibility.LabeledBy == nil {
		lp.Accessibility.LabeledBy = label
	}
	for _, extra := range extraHeaders {
		panel.AddChild(extra)
	}
	panel.AddChild(scroll)
	return panel, scroll
}

// holdMinSizeOnDisplay keeps the window at least as large as its content's minimum size, as a window is by default,
// save where its display has no room for that, where it is held to no more than the display allows.
func holdMinSizeOnDisplay(wnd *unison.Window) {
	wnd.MinMaxContentSizeCallback = func() (minimum, maximum geom.Size) {
		minimum, _, maximum = wnd.Content().Sizes(geom.Size{})
		if d := wnd.Display(); d != nil {
			frame, r := wnd.FrameRect(), wnd.ContentRect()
			minimum.Width = min(minimum.Width, d.Usable.Width-(frame.Width-r.Width))
			minimum.Height = min(minimum.Height, d.Usable.Height-(frame.Height-r.Height))
		}
		return minimum, maximum
	}
}

// listMinRows is how many rows of options the lists of the template picker and the modifier prompt have room for,
// however few they hold.
const listMinRows = 10

// listMinSize returns the least room a list of options is given inside its scroll panel's border: enough for
// listMinRows of the template picker's rows, and a width to go with it.
func listMinSize() geom.Size {
	font := unison.DefaultCheckBoxTheme.Font
	row := max(pickerCheckBoxSize().Height, pickerDisclosureSize().Height, font.LineHeight())
	_, button, _ := NewSVGButtonForFont(svg.Edit, font, -2).Sizes(geom.Size{})
	row = max(row, button.Height)
	// The list's border is StdHSpacing all around.
	return geom.NewSize(xmath.Ceil(font.SimpleWidth("n")*80),
		xmath.Ceil(row*listMinRows+unison.StdVSpacing*(listMinRows-1))).
		Add(geom.NewUniformInsets(unison.StdHSpacing).Size())
}

// minSizeLayout is a layout that asks for at least a minimum size for what it lays out, beyond the border, and prefers
// no less than that either.
type minSizeLayout struct {
	unison.Layout
	minimum geom.Size
}

func (l *minSizeLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	minSize, prefSize, maxSize = l.Layout.LayoutSizes(target, hint)
	minimum := l.minimum
	if border := target.Border(); border != nil {
		minimum = minimum.Add(border.Insets().Size())
	}
	minSize = geom.NewSize(max(minSize.Width, minimum.Width), max(minSize.Height, minimum.Height))
	prefSize = geom.NewSize(max(prefSize.Width, minSize.Width), max(prefSize.Height, minSize.Height))
	maxSize = geom.NewSize(max(maxSize.Width, prefSize.Width), max(maxSize.Height, prefSize.Height))
	return minSize, prefSize, maxSize
}

// setListMinSize gives the scroll panel holding a list of options its least room (see listMinSize), so the list has room
// for a number of rows however few it holds. The dialog it is in should hold that size (see holdMinSizeOnDisplay).
func setListMinSize(scroll *unison.ScrollPanel) {
	scroll.SetLayout(&minSizeLayout{Layout: scroll, minimum: listMinSize()})
}
