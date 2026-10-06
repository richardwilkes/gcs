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
	"github.com/richardwilkes/gcs/v5/ux/colors"
	"github.com/richardwilkes/gcs/v5/ux/fonts"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/filltype"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var _ unison.Border = &TitledBorder{}

// TitledBorder provides a titled line border.
type TitledBorder struct {
	Title string
	Font  unison.Font
	// HeadingInContent leaves the title strip to the panel's content, for the heading addBlockHeading adds there. The
	// title is drawn the same either way.
	HeadingInContent bool
}

func (t *TitledBorder) font() unison.Font {
	if t.Font == nil {
		return fonts.PageLabelPrimary
	}
	return t.Font
}

// TitleHeight returns the height of the strip the title is drawn in, below the border's top line.
func (t *TitledBorder) TitleHeight() float32 {
	return xmath.Ceil(t.font().LineHeight()) + 1
}

// TitleStrip returns the strip the title is drawn in, for a border of the given size, in the coordinates of the panel
// the border is on.
func (t *TitledBorder) TitleStrip(size geom.Size) geom.Rect {
	return geom.NewRect(1, 1, size.Width-2, t.TitleHeight())
}

// titleInsets returns the insets including the title strip, which the border paints even when HeadingInContent is set.
func (t *TitledBorder) titleInsets() geom.Insets {
	return geom.Insets{
		Top:    t.TitleHeight() + 1,
		Left:   1,
		Bottom: 1,
		Right:  1,
	}
}

// Insets implements unison.Border.
func (t *TitledBorder) Insets() geom.Insets {
	insets := t.titleInsets()
	if t.HeadingInContent {
		insets.Top = 1
	}
	return insets
}

// Draw implements unison.Border.
func (t *TitledBorder) Draw(gc *unison.Canvas, rect geom.Rect) {
	clip := rect.Inset(t.titleInsets())
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
	text.Draw(gc, geom.NewPoint(rect.X+(rect.Width-text.Width())/2, rect.Y+1+text.Baseline()))
}
