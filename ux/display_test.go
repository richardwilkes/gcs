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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestPlaceWindowOver verifies the placement the about box, progress window and undocked dockable windows all share:
// centered horizontally over the frame, one third of the way down it, aligned to whole pixels, and clamped onto the
// display when the frame would otherwise push it off.
func TestPlaceWindowOver(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	screen.Do(func() {
		wnd, err := unison.NewWindow("placement", unison.NotResizableWindowOption())
		c.NoError(err)
		defer wnd.Dispose()
		content := wnd.Content()
		content.SetLayout(&unison.FlexLayout{Columns: 1})
		fixed := unison.NewPanel()
		fixed.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
			size := geom.NewSize(200, 100)
			return size, size, size
		})
		content.AddChild(fixed)
		wnd.Pack()
		size := wnd.FrameRect().Size
		c.Equal(geom.NewSize(200, 100), size, "packing must size the window to its content")

		frame := geom.NewRect(100.5, 50.5, 600, 400)
		placeWindowOver(wnd, frame)
		want := geom.NewRect(frame.X+(frame.Width-size.Width)/2, frame.Y+(frame.Height-size.Height)/3, size.Width,
			size.Height).Align()
		c.Equal(want, wnd.FrameRect(), "the window must sit centered and one third down over the frame")

		display := unison.BestDisplayForRect(want)
		c.NotNil(display)
		offscreen := geom.NewRect(display.Usable.Right()+1000, display.Usable.Bottom()+1000, 600, 400)
		placeWindowOver(wnd, offscreen)
		got := wnd.FrameRect()
		c.True(got.In(display.Usable), "a frame off the display must be clamped back onto it: %v not in %v", got,
			display.Usable)
		c.Equal(size, got.Size, "clamping must not resize the window")
	})
}
