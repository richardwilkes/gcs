/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/gurps"
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
		return gurps.PageLabelPrimaryFont
	}
	return t.Font
}

// Insets implements unison.Border
func (t *TitledBorder) Insets() unison.Insets {
	return unison.Insets{
		Top:    t.font().LineHeight() + 2,
		Left:   1,
		Bottom: 1,
		Right:  1,
	}
}

// Draw implements unison.Border
func (t *TitledBorder) Draw(gc *unison.Canvas, rect unison.Rect) {
	clip := rect.Inset(t.Insets())
	path := unison.NewPath()
	path.SetFillType(filltype.EvenOdd)
	path.Rect(rect)
	path.Rect(clip)
	gc.DrawPath(path, unison.ThemeAboveSurface.Paint(gc, rect, paintstyle.Fill))
	text := unison.NewText(t.Title, &unison.TextDecoration{
		Font:       t.font(),
		Foreground: unison.ThemeOnSurface,
	})
	text.Draw(gc, rect.X+(rect.Width-text.Width())/2, rect.Y+1+text.Baseline())
}
