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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// NewToolbarSeparator creates a new vertical separator for the toolbar.
func NewToolbarSeparator() *unison.Separator {
	spacer := unison.NewSeparator()
	spacer.LineInk = unison.ThemeSurfaceEdge
	spacer.Vertical = true
	spacer.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(unison.StdHSpacing)))
	spacer.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Fill})
	spacer.SetSizer(func(hint unison.Size) (minSize, prefSize, maxSize unison.Size) {
		minSize, prefSize, maxSize = spacer.DefaultSizes(hint)
		baseline := unison.DefaultButtonTheme.Font.Baseline()
		minSize.Height = baseline
		prefSize.Height = baseline
		maxSize.Height = baseline
		return
	})
	spacer.DrawCallback = func(canvas *unison.Canvas, _ unison.Rect) {
		rect := spacer.ContentRect(false)
		paint := spacer.LineInk.Paint(canvas, rect, paintstyle.Stroke)
		paint.SetPathEffect(unison.NewDashPathEffect([]float32{2, 2}, 0))
		canvas.DrawLine(rect.X, rect.Y, rect.X, rect.Bottom(), paint)
	}
	return spacer
}
