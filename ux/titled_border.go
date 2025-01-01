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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/filltype"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var _ unison.Border = &TitledBorder{}

// TitledBorder provides a titled line border.
type TitledBorder struct {
	Title string
	Font  unison.Font
}

func (t *TitledBorder) font() unison.Font {
	if t.Font == nil {
		return fonts.PageLabelPrimary
	}
	return t.Font
}

// Insets implements unison.Border
func (t *TitledBorder) Insets() unison.Insets {
	return unison.Insets{
		Top:    xmath.Ceil(t.font().LineHeight()) + 2,
		Left:   1,
		Bottom: 1,
		Right:  1,
	}
}

// Draw implements unison.Border
func (t *TitledBorder) Draw(gc *unison.Canvas, rect unison.Rect) {
	clip := rect.Inset(t.Insets())
	clip.Y += 0.5
	clip.Height -= 0.5
	path := unison.NewPath()
	path.SetFillType(filltype.EvenOdd)
	path.Rect(rect)
	path.Rect(clip)
	gc.DrawPath(path, colors.Header.Paint(gc, rect, paintstyle.Fill))
	text := unison.NewSmallCapsText(t.Title, &unison.TextDecoration{
		Font:            t.font(),
		OnBackgroundInk: colors.OnHeader,
	})
	text.Draw(gc, rect.X+(rect.Width-text.Width())/2, rect.Y+1+text.Baseline())
}
