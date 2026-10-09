// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"image"
	"testing"

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func TestStaticTextShowsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	calc := openCalculator(t, screen)
	explosion := calc.explosion
	selectCalculatorTab(t, screen, calc, explosion)
	setFocusForReading(true)
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	ringInWindow := func() (ring geom.Rect, ok bool) {
		screen.Do(func() {
			var clip geom.Rect
			if ring, clip, ok = ux.StaticTextFocusRing(wnd, wnd.Content()); ok {
				c.Equal(ring, ring.Intersect(clip), "all of the ring can be seen")
				ring = wnd.Content().RectToRoot(ring)
			}
		})
		return ring, ok
	}
	capture := func() *image.NRGBA {
		screen.Sync()
		img := screen.CaptureWindow(wnd)
		if img == nil {
			t.Fatal("the window could not be captured")
		}
		return img
	}
	changedWithin := func(before, after *image.NRGBA, r geom.Rect) (count int) {
		scale := float32(after.Bounds().Dx()) / wnd.ContentRect().Width
		area := image.Rect(int(r.X*scale), int(r.Y*scale), int(r.Right()*scale), int(r.Bottom()*scale)).
			Intersect(after.Bounds()).Intersect(before.Bounds())
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				if before.NRGBAAt(x, y) != after.NRGBAAt(x, y) {
					count++
				}
			}
		}
		return count
	}

	screen.Do(explosion.hotFragmentsBox.RequestFocus)
	c.Equal(explosion.hotFragmentsBox.AsPanel(), focus())
	_, ok := ringInWindow()
	c.False(ok, "a checkbox shows the focus itself")
	before := capture()

	screen.KeyPress(unison.KeyTab, mod.None)
	target, isLabel := focus().Self.(*unison.Label)
	if !isLabel {
		t.Fatalf("Tab from Hot fragments must go to the Target subheader, not to a %T", focus().Self)
	}
	c.Equal("Target", target.String())
	ring, ok := ringInWindow()
	c.True(ok, "static text holding the focus is given a ring")
	var frame geom.Rect
	screen.Do(func() { frame = target.RectToRoot(target.ContentRect(true)) })
	c.Equal(ring, ring.Intersect(frame), "the ring %v lies within the label %v", ring, frame)
	c.True(ring.Width < frame.Width/2, "the ring goes around the text, not the width of the column: %v in %v", ring, frame)
	c.True(ring.Width > 0 && ring.Height > 0)
	after := capture()
	c.True(changedWithin(before, after, ring) > 0, "the ring is drawn")

	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(explosion.target.popup.AsPanel(), focus(), "Tab goes on to the target's Source")
	_, ok = ringInWindow()
	c.False(ok, "a popup shows the focus itself")
	gone := capture()
	c.True(changedWithin(after, gone, ring) > 0, "the ring is gone once the focus has moved on")

	var header *ux.TextLabel
	screen.Do(func() {
		if children := explosion.content.Children(); len(children) > 0 {
			if label, isTextLabel := children[0].Self.(*ux.TextLabel); isTextLabel {
				header = label
			}
		}
	})
	if header == nil {
		t.Fatal("the calculator must start with its header")
	}
	screen.Do(header.RequestFocus)
	c.Equal(header.AsPanel(), focus())
	_, ok = ringInWindow()
	c.False(ok, "a label holding links shows the focus itself")

	setFocusForReading(false)
	screen.Do(explosion.hotFragmentsBox.RequestFocus)
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(explosion.target.popup.AsPanel(), focus(), "Tab goes from Hot fragments to the target's Source")
}
