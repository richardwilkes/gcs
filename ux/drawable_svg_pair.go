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
)

// DrawableSVGPair draws two SVG's side-by-side.
type DrawableSVGPair struct {
	Left  *unison.SVG
	Right *unison.SVG
	Size  unison.Size
}

// LogicalSize implements the Drawable interface.
func (s *DrawableSVGPair) LogicalSize() unison.Size {
	return s.Size
}

// DrawInRect implements the Drawable interface.
func (s *DrawableSVGPair) DrawInRect(canvas *unison.Canvas, rect unison.Rect, _ *unison.SamplingOptions, paint *unison.Paint) {
	r := rect
	r.Width /= 2
	offset := s.Left.OffsetToCenterWithinScaledSize(r.Size)
	r.X += offset.X
	r.Y += offset.Y
	s.Left.DrawInRectPreservingAspectRatio(canvas, r, nil, paint)
	r.X = rect.X + r.Width
	r.Y = rect.Y
	offset = s.Right.OffsetToCenterWithinScaledSize(r.Size)
	r.X += offset.X
	r.Y += offset.Y
	s.Right.DrawInRectPreservingAspectRatio(canvas, r, nil, paint)
}
