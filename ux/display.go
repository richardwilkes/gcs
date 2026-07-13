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
	"github.com/richardwilkes/unison"
)

// unison.PrimaryDisplay() can return nil on some configurations (notably on Linux when no monitor is enumerated).
// Dereferencing it directly then crashes, and when that happens on an unrecovered background goroutine (e.g. the
// markdown image loader) the process terminates with nothing written to the log. The helpers below provide safe
// fallbacks so a missing primary display degrades gracefully instead of crashing.

// primaryDisplayScale returns the content scale of the primary display, or a 1:1 scale when no display is available.
func primaryDisplayScale() geom.Point {
	if d := unison.PrimaryDisplay(); d != nil {
		return d.Scale
	}
	return geom.NewPoint(1, 1)
}

// primaryDisplayUsableRect returns the usable area of the primary display, or a reasonable default when no display is
// available.
func primaryDisplayUsableRect() geom.Rect {
	if d := unison.PrimaryDisplay(); d != nil {
		return d.Usable
	}
	return geom.NewRect(0, 0, 1024, 768)
}

// windowPlacementFrame returns a rectangle within which to center a transient window: the active window's frame when
// there is one, or failing that, the frontmost window's frame, so that transient windows open on the same display the
// user is working on. Only when no window is available does it fall back to the primary display's usable area (with a
// safe fallback when no display is available).
func windowPlacementFrame() geom.Rect {
	if focused := unison.ActiveWindow(); focused != nil {
		return focused.FrameRect()
	}
	if frontmost := unison.FrontmostWindow(); frontmost != nil {
		return frontmost.FrameRect()
	}
	return primaryDisplayUsableRect()
}
