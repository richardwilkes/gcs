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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const searchFieldClientDataKey = "is_search_field"

// toolbarSearchWidthSample is the text whose width, in a toolbar search field's font, is the least room the field keeps
// for its text.
const toolbarSearchWidthSample = "MMMMMMMMMM"

// NewSearchField creates a new search widget.
func NewSearchField(watermark string, modifiedCallback func(before, after *unison.FieldState)) *unison.Field {
	f := unison.NewField()
	if watermark != "" {
		f.Watermark = watermark
		f.Tooltip = newWrappedTooltip(watermark)
		// The watermark is only a placeholder to a screen reader, and the field sits among toolbar buttons with no
		// label to name it, so the watermark serves as its name too.
		f.Accessibility.Name = watermark
	}
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	b := NewSVGButtonForFont(unison.CircledXSVG, unison.DefaultButtonTheme.Font, -2)
	b.OnSelectionInk = f.OnEditableInk
	b.SetFocusable(false)
	b.SetEnabled(false)
	b.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return unison.ArrowCursor() }
	b.Tooltip = newWrappedTooltip(i18n.Text("Clear"))
	b.ClickCallback = func() { f.SetText("") }
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		b.SetEnabled(after.Text != "")
		modifiedCallback(before, after)
	}
	f.InstallAccessoryPanel(b)
	f.ClientData()[searchFieldClientDataKey] = true
	return f
}

// NewToolbarSearchField creates a search widget for a toolbar, where it grabs the width the toolbar has to spare. A
// column that grabs in a FlexLayout shrinks no further than an explicit MinSize, ignoring the field's own minimum, so
// that minimum is copied there to keep the field from shrinking out of sight.
func NewToolbarSearchField(watermark string, modifiedCallback func(before, after *unison.FieldState)) *unison.Field {
	f := NewSearchField(watermark, modifiedCallback)
	f.SetMinimumTextWidthUsing(toolbarSearchWidthSample)
	if data, ok := f.LayoutData().(*unison.FlexLayoutData); ok {
		minSize, _, _ := f.Sizes(geom.Size{})
		data.MinSize.Width = minSize.Width
	}
	return f
}
