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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const searchFieldClientDataKey = "is_search_field"

// NewSearchField creates a new search widget.
func NewSearchField(watermark string, modifiedCallback func(before, after *unison.FieldState)) *unison.Field {
	f := unison.NewField()
	if watermark != "" {
		f.Watermark = watermark
		f.Tooltip = newWrappedTooltip(watermark)
	}
	f.SetLayout(&searchLayout{
		field: f,
		flex: unison.FlexLayout{
			Columns: 1,
			HAlign:  align.End,
			VAlign:  align.Middle,
		},
	})
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	b := NewSVGButtonForFont(unison.CircledXSVG, unison.DefaultButtonTheme.Font, -2)
	b.OnSelectionInk = f.OnEditableInk
	b.SetFocusable(false)
	b.SetEnabled(false)
	b.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor { return unison.ArrowCursor() }
	b.Tooltip = newWrappedTooltip(i18n.Text("Clear"))
	b.ClickCallback = func() { f.SetText("") }
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		b.SetEnabled(after.Text != "")
		modifiedCallback(before, after)
	}
	f.AddChild(b)
	f.ClientData()[searchFieldClientDataKey] = true
	return f
}

type searchLayout struct {
	field *unison.Field
	flex  unison.FlexLayout
}

func (s *searchLayout) LayoutSizes(_ *unison.Panel, hint unison.Size) (minSize, prefSize, maxSize unison.Size) {
	return s.field.DefaultSizes(hint)
}

func (s *searchLayout) PerformLayout(target *unison.Panel) {
	s.flex.PerformLayout(target)
}
