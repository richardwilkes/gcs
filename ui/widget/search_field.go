/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package widget

import (
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// NewSearchField creates a new search widget.
func NewSearchField() *unison.Field {
	f := unison.NewField()
	f.Watermark = i18n.Text("Search")
	f.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.EndAlignment,
		VAlign:  unison.MiddleAlignment,
	})
	b := unison.NewSVGButton(unison.CircledXSVG())
	b.OnSelectionInk = f.OnEditableInk
	b.SetEnabled(false)
	b.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor { return unison.ArrowCursor() }
	b.ClickCallback = func() { f.SetText("") }
	f.ModifiedCallback = func() { b.SetEnabled(f.Text() != "") }
	f.AddChild(b)
	return f
}
