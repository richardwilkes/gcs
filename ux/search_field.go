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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// NewSearchField creates a new search widget.
func NewSearchField(watermark string, modifiedCallback func(before, after *unison.FieldState)) *unison.Field {
	f := unison.NewField()
	if watermark != "" {
		f.Watermark = watermark
		f.Tooltip = unison.NewTooltipWithText(watermark)
	}
	f.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.EndAlignment,
		VAlign:  unison.MiddleAlignment,
	})
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})
	b := NewSVGButtonForFont(unison.CircledXSVG, unison.DefaultSVGButtonTheme.Font, -2)
	b.OnSelectionInk = f.OnEditableInk
	b.SetFocusable(false)
	b.SetEnabled(false)
	b.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor { return unison.ArrowCursor() }
	b.Tooltip = unison.NewTooltipWithText(i18n.Text("Clear"))
	b.ClickCallback = func() { f.SetText("") }
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		b.SetEnabled(after.Text != "")
		modifiedCallback(before, after)
	}
	f.AddChild(b)
	return f
}
