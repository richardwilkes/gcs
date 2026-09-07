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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/behavior"
)

// newScrollPanDragFixture builds a 100x100 scroll panel over 500x400 content, laid out so the scroll bars have a real
// range, and returns the pan state wired to it.
func newScrollPanDragFixture() (*scrollPanDrag, *unison.ScrollPanel, *unison.Panel) {
	content := unison.NewPanel()
	content.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
		size := geom.NewSize(500, 400)
		return size, size, size
	})
	scroll := unison.NewScrollPanel()
	scroll.SetContent(content, behavior.Fill, behavior.Fill)
	scroll.SetFrameRect(geom.NewRect(0, 0, 100, 100))
	scroll.ValidateLayout()
	return &scrollPanDrag{scroll: scroll, content: content}, scroll, content
}

func scrollPosition(scroll *unison.ScrollPanel) geom.Point {
	h, v := scroll.Position()
	return geom.NewPoint(h, v)
}

// TestScrollPanDragFollowsThePointer verifies that dragging the content moves the scroll position by the same amount
// the pointer moved, in the opposite direction, so the content stays under the pointer.
func TestScrollPanDragFollowsThePointer(t *testing.T) {
	c := check.New(t)
	pan, scroll, content := newScrollPanDragFixture()
	c.False(pan.active)
	c.Equal(geom.Point{}, scrollPosition(scroll))

	pan.begin(geom.NewPoint(50, 50))
	c.True(pan.active)
	// The pointer moves 20 left and 30 up, so the content scrolls 20 right and 30 down.
	pan.drag(geom.NewPoint(30, 20))
	c.Equal(geom.NewPoint(20, 30), scrollPosition(scroll))
	c.Equal(geom.NewPoint(-20, -30), content.FrameRect().Point, "the content is repositioned to match")

	// The pointer holds still. In content-local terms it now sits 20 further right and 30 further down, since the
	// content moved underneath it, and that must not be mistaken for further movement.
	pan.drag(geom.NewPoint(50, 50))
	c.Equal(geom.NewPoint(20, 30), scrollPosition(scroll))

	// Dragging back past where the drag began is clamped at the start of the content.
	pan.drag(geom.NewPoint(200, 200))
	c.Equal(geom.Point{}, scrollPosition(scroll))

	pan.end()
	c.False(pan.active)
}

// TestScrollPanDragStartsFromTheCurrentPosition verifies that a drag begun while already scrolled offsets from that
// position rather than from the origin, and that scrolling is clamped at the end of the content.
func TestScrollPanDragStartsFromTheCurrentPosition(t *testing.T) {
	c := check.New(t)
	pan, scroll, _ := newScrollPanDragFixture()
	scroll.SetPosition(100, 50)
	c.Equal(geom.NewPoint(100, 50), scrollPosition(scroll))

	pan.begin(geom.NewPoint(10, 10))
	pan.drag(geom.NewPoint(0, 5))
	c.Equal(geom.NewPoint(110, 55), scrollPosition(scroll))

	// The furthest the 100x100 view can scroll over 500x400 content is 400x300.
	pan.drag(geom.NewPoint(-1000, -1000))
	c.Equal(geom.NewPoint(400, 300), scrollPosition(scroll))
}

// TestScrollPanDragCursor verifies the move cursor is shown only while a drag is in progress.
func TestScrollPanDragCursor(t *testing.T) {
	c := check.New(t)
	pan, _, _ := newScrollPanDragFixture()
	c.Equal(unison.ArrowCursor(), pan.cursor())
	pan.begin(geom.Point{})
	c.Equal(unison.MoveCursor(), pan.cursor())
	pan.end()
	c.Equal(unison.ArrowCursor(), pan.cursor())
}

// TestScrollPanDragInstall verifies that install wires the pan state to the scroll panel and content, and that the
// callbacks it sets on the content drive the drag: a mouse-down begins it, a mouse-drag pans, and a mouse-up ends it.
func TestScrollPanDragInstall(t *testing.T) {
	c := check.New(t)
	_, scroll, content := newScrollPanDragFixture()
	var pan scrollPanDrag
	pan.install(scroll, content, content)
	c.Equal(scroll, pan.scroll)
	c.Equal(content, pan.content)
	c.NotNil(content.MouseDownCallback)
	c.NotNil(content.MouseDragCallback)
	c.NotNil(content.MouseUpCallback)
	c.NotNil(content.UpdateCursorCallback)

	c.Equal(unison.ArrowCursor(), content.UpdateCursorCallback(geom.Point{}))
	c.True(content.MouseDownCallback(geom.NewPoint(50, 50), 1, 1, 0))
	c.True(pan.active)
	c.Equal(unison.MoveCursor(), content.UpdateCursorCallback(geom.Point{}))
	c.True(content.MouseDragCallback(geom.NewPoint(30, 20), 1, 0))
	c.Equal(geom.NewPoint(20, 30), scrollPosition(scroll))
	c.True(content.MouseUpCallback(geom.NewPoint(30, 20), 1, 0))
	c.False(pan.active)
	c.Equal(unison.ArrowCursor(), content.UpdateCursorCallback(geom.Point{}))
}

// TestDockablesPanOnDrag verifies that dragging the content of the image and markdown dockables pans their scroll
// panels. The markdown dockable once installed the mouse handlers without wiring the pan state to its scroll panel,
// so a mouse-down on it dereferenced a nil scroll panel and panicked.
func TestDockablesPanOnDrag(t *testing.T) {
	c := check.New(t)
	svgPath := filepath.Join(t.TempDir(), "image.svg")
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 16" width="480" height="320">` +
		`<rect x="0" y="0" width="24" height="16"/></svg>`
	c.NoError(os.WriteFile(svgPath, []byte(svg), 0o600))
	image, err := NewImageDockable(svgPath)
	c.NoError(err)
	imageDockable, ok := image.(*ImageDockable)
	c.True(ok)
	markdown := NewMarkdownDockableWithContent("notes", strings.Repeat("Some text.\n\n", 100), false, false)
	markdownDockable, ok := markdown.(*MarkdownDockable)
	c.True(ok)
	for _, tc := range []struct {
		name    string
		scroll  *unison.ScrollPanel
		content *unison.Panel
	}{
		{name: "image", scroll: imageDockable.scroll, content: imageDockable.drawablePanel},
		{name: "markdown", scroll: markdownDockable.scroller, content: markdownDockable.markdown.AsPanel()},
	} {
		tc.scroll.SetFrameRect(geom.NewRect(0, 0, 100, 100))
		tc.scroll.ValidateLayout()
		c.Equal(geom.Point{}, scrollPosition(tc.scroll), tc.name)
		c.True(tc.content.MouseDownCallback(geom.NewPoint(50, 50), 1, 1, 0), tc.name)
		c.True(tc.content.MouseDragCallback(geom.NewPoint(50, 20), 1, 0), tc.name)
		c.Equal(geom.NewPoint(0, 30), scrollPosition(tc.scroll), "%s: the content should pan with the pointer", tc.name)
		c.True(tc.content.MouseUpCallback(geom.NewPoint(50, 20), 1, 0), tc.name)
	}
}
