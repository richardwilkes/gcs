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
	"github.com/richardwilkes/unison/enums/align"
)

// newFixedSizePanel returns a panel whose preferred size is as given and whose minimum width is minWidth, placed in
// the middle of its row.
func newFixedSizePanel(minWidth, width, height float32) *unison.Panel {
	p := unison.NewPanel()
	p.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
		prefSize = geom.NewSize(width, height)
		return geom.NewSize(minWidth, height), prefSize, unison.MaxSize(prefSize)
	})
	p.SetLayoutData(align.Middle)
	return p
}

// TestFlowFillLayout lays four 20-point buttons and a field at least 50 points wide out at several widths: on one row
// while the field has its minimum left after the buttons, with the field taking whatever is left; with the field on a
// row of its own, taking the whole width, when less is left; and with the buttons themselves wrapping when they don't
// fit, the field joining the last of their rows when it can. The minimum size holds the widest child and every row the
// width calls for, and the preferred size at no width is that of a single row with the field at its preferred width.
func TestFlowFillLayout(t *testing.T) {
	c := check.New(t)
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(2)))
	for range 4 {
		toolbar.AddChild(newFixedSizePanel(20, 20, 20))
	}
	field := newFixedSizePanel(50, 80, 10)
	toolbar.AddChild(field)
	layout := &flowFillLayout{HSpacing: 4, VSpacing: 2}
	toolbar.SetLayout(layout)
	frames := func(width float32) []geom.Rect {
		toolbar.SetFrameRect(geom.NewRect(0, 0, width, 100))
		layout.PerformLayout(toolbar)
		rects := make([]geom.Rect, 0, 5)
		for _, child := range toolbar.Children() {
			rects = append(rects, child.FrameRect())
		}
		return rects
	}

	// 4 buttons at 20 plus 4 gaps of 4 take 96 of the width inside the 2-point border, so a width of 150 leaves 50 for
	// the field, which is its minimum.
	got := frames(150)
	c.Equal(geom.NewRect(2, 2, 20, 20), got[0])
	c.Equal(geom.NewRect(74, 2, 20, 20), got[3])
	c.Equal(geom.NewRect(98, 7, 50, 10), got[4], "the field takes what is left of the row, centered in its height")
	minSize, prefSize, _ := layout.LayoutSizes(toolbar, geom.NewSize(150, 0))
	c.Equal(geom.NewSize(50+4, 20+4), minSize, "the minimum is the widest child by the rows needed")
	c.Equal(geom.NewSize(150, 20+4), prefSize)

	// A point less leaves less than the field's minimum, so it goes on a row of its own, taking the whole width.
	got = frames(149)
	c.Equal(geom.NewRect(74, 2, 20, 20), got[3], "the buttons stay on the first row")
	c.Equal(geom.NewRect(2, 2+20+2, 145, 10), got[4], "the field takes a row of its own")
	minSize, prefSize, _ = layout.LayoutSizes(toolbar, geom.NewSize(149, 0))
	c.Equal(geom.NewSize(50+4, 20+2+10+4), minSize, "the minimum height holds both rows")
	c.Equal(geom.NewSize(149, 20+2+10+4), prefSize)

	// At 100, three buttons fit on the first row (20+4+20+4+20 = 68, and a fourth would end at 92 of the 96 inside the
	// border -- it fits), so all four stay on the first row and the field, with only 0 left, takes the second.
	got = frames(100)
	c.Equal(geom.NewRect(74, 2, 20, 20), got[3])
	c.Equal(geom.NewRect(2, 24, 96, 10), got[4])

	// At 90, only three buttons fit on the first row; the fourth starts the second, where 86-24 = 62 is left for the
	// field, which is more than its minimum, so the field joins it.
	got = frames(90)
	c.Equal(geom.NewRect(50, 2, 20, 20), got[2])
	c.Equal(geom.NewRect(2, 24, 20, 20), got[3], "the fourth button wraps")
	c.Equal(geom.NewRect(26, 29, 62, 10), got[4], "and the field joins it on the second row")

	// With no width to fit to, everything is on one row and the field is at its preferred width.
	_, prefSize, _ = layout.LayoutSizes(toolbar, geom.Size{})
	c.Equal(geom.NewSize(96+80+4, 20+4), prefSize)

	// A child with no alignment is placed at the top of its row; one that fills takes the row's height.
	field.SetLayoutData(nil)
	c.Equal(geom.NewRect(98, 2, 50, 10), frames(150)[4])
	field.SetLayoutData(align.Fill)
	c.Equal(geom.NewRect(98, 2, 50, 20), frames(150)[4])
	field.SetLayoutData(align.End)
	c.Equal(geom.NewRect(98, 12, 50, 10), frames(150)[4])

	// A child after Fill follows it on its row at its preferred width, and they move to a row of their own together
	// when less than the field's minimum would be left ahead of it. Here the field and a 30-point tag after it need 50
	// plus 4 plus 30 after the 96 the buttons take, so a width of 184 is the least that keeps everything on one row.
	field.SetLayoutData(align.Middle)
	tag := newFixedSizePanel(30, 30, 20)
	toolbar.AddChild(tag)
	layout.Fill = field
	got = frames(184)
	c.Equal(geom.NewRect(98, 7, 50, 10), got[4], "the field takes what is left ahead of the tag")
	c.Equal(geom.NewRect(152, 2, 30, 20), got[5], "which ends the row")
	minSize, prefSize, _ = layout.LayoutSizes(toolbar, geom.NewSize(184, 0))
	c.Equal(geom.NewSize(50+4+30+4, 20+4), minSize, "the minimum width holds the field and the tag together")
	c.Equal(geom.NewSize(184, 20+4), prefSize)
	got = frames(183)
	c.Equal(geom.NewRect(74, 2, 20, 20), got[3], "the buttons stay on the first row")
	c.Equal(geom.NewRect(2, 29, 145, 10), got[4], "the field moves to a row of its own, with the tag")
	c.Equal(geom.NewRect(151, 24, 30, 20), got[5])
	_, prefSize, _ = layout.LayoutSizes(toolbar, geom.Size{})
	c.Equal(geom.NewSize(96+80+4+30+4, 20+4), prefSize, "at no width, the row holds everything")
	layout.Fill = nil
	c.Equal(geom.NewRect(98, 7, 80, 10), frames(300)[4], "without Fill, the last child is given the width left")
	c.Equal(geom.NewRect(182, 2, 116, 20), frames(300)[5])

	// An empty toolbar has nothing but its border.
	toolbar.RemoveAllChildren()
	minSize, prefSize, _ = layout.LayoutSizes(toolbar, geom.NewSize(150, 0))
	c.Equal(geom.NewSize(4, 4), minSize)
	c.Equal(geom.NewSize(4, 4), prefSize)
	layout.PerformLayout(toolbar)
}

// TestFlowFillLayoutNarrowsWideChild lays a child ahead of Fill whose preferred width is more than the toolbar has out
// at the layout's own minimum width and at wider ones: it is narrowed to the width of its row, but never below its own
// minimum, after wrapping to a row of its own when it follows another child.
func TestFlowFillLayoutNarrowsWideChild(t *testing.T) {
	c := check.New(t)
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(2)))
	button := newFixedSizePanel(20, 20, 20)
	toolbar.AddChild(button)
	wide := newFixedSizePanel(30, 120, 20)
	toolbar.AddChild(wide)
	toolbar.AddChild(newFixedSizePanel(50, 80, 10))
	layout := &flowFillLayout{HSpacing: 4, VSpacing: 2}
	toolbar.SetLayout(layout)
	frames := func(width float32) []geom.Rect {
		toolbar.SetFrameRect(geom.NewRect(0, 0, width, 100))
		layout.PerformLayout(toolbar)
		rects := make([]geom.Rect, 0, 3)
		for _, child := range toolbar.Children() {
			rects = append(rects, child.FrameRect())
		}
		return rects
	}

	minSize, _, _ := layout.LayoutSizes(toolbar, geom.Size{})
	c.Equal(float32(50+4), minSize.Width)
	got := frames(minSize.Width)
	c.Equal(geom.NewRect(2, 2, 20, 20), got[0])
	c.Equal(geom.NewRect(2, 24, 50, 20), got[1], "the wide child wraps and is narrowed to the row")
	c.Equal(geom.NewRect(2, 46, 50, 10), got[2])

	got = frames(100)
	c.Equal(geom.NewRect(2, 24, 96, 20), got[1], "the wide child is narrowed to the row, not to its minimum")
	_, prefSize, _ := layout.LayoutSizes(toolbar, geom.NewSize(100, 0))
	c.Equal(geom.NewSize(100, 20+2+20+2+10+4), prefSize, "the preferred width does not exceed the width given")

	// Narrowing never goes below the child's minimum, even when the toolbar is given less than the layout's minimum.
	toolbar.RemoveChild(button)
	c.Equal(geom.NewRect(2, 2, 30, 20), frames(20)[0])
}
