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
	"math"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
)

const (
	sideBarInitialWidth = 200
	sideBarMaxFitWidth  = 300
	// sideBarMaxFraction is the most of its container's content width a side bar is laid out at, so that a narrow view
	// leaves the rest to its content.
	sideBarMaxFraction = 0.4
)

var _ unison.Layout = &sideBar{}

// sideBar is a scrolling panel that a view shows to the left of its content, such as a table of contents, with a
// divider between them that can be dragged to change its width. It is laid out at its preferred width, though at no
// more than sideBarMaxFraction of its container's content width.
type sideBar struct {
	scroll    *unison.ScrollPanel
	divider   *unison.Panel
	preferred float32 // set by fitTo and by dragging the divider
}

// newSideBar returns a side bar that scrolls the content. It is not shown until toggle is called.
func newSideBar(content unison.Paneler) *sideBar {
	s := &sideBar{
		scroll:    unison.NewScrollPanel(),
		divider:   unison.NewPanel(),
		preferred: sideBarInitialWidth,
	}
	s.scroll.SetContent(content, behavior.Fill, behavior.Fill)
	// The side bar stands in for the scroll panel's own layout, to report the width it is to be laid out at.
	s.scroll.SetLayout(s)
	s.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		VGrab:  true,
	})

	s.divider.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: geom.Size{Width: unison.DefaultDockTheme.DockDividerSize()},
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		VGrab:    true,
	})
	s.divider.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor {
		return unison.ResizeHorizontalCursor()
	}
	s.divider.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		unison.DefaultDockTheme.DrawHorizontalGripper(gc, s.divider.ContentRect(true))
	}
	var initialWidth float32
	var eventPosition float32
	s.divider.MouseDownCallback = func(where geom.Point, _, _ int, _ mod.Modifiers) bool {
		initialWidth = s.scroll.FrameRect().Width
		eventPosition = s.divider.Parent().PointFromRoot(s.divider.PointToRoot(where)).X
		return true
	}
	s.divider.MouseDragCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		pos := eventPosition - s.divider.Parent().PointFromRoot(s.divider.PointToRoot(where)).X
		old := s.preferred
		// Held to the most the side bar is laid out at, so that the divider stops where the pointer passes that and
		// the side bar keeps the width it was left showing when the view widens.
		s.preferred = max(min(initialWidth-pos, s.maxWidth()), 1)
		if old != s.preferred {
			s.divider.Parent().MarkForLayoutAndRedraw()
		}
		return true
	}
	return s
}

// toggle shows the side bar and its divider as the first two of the container's three columns when the container has
// one, and removes them, leaving it one, when it has three. The container must use a FlexLayout. When the focus is
// within the side bar as it is removed, the focus moves to document, the view's content.
func (s *sideBar) toggle(container *unison.Panel, document unison.Paneler) {
	if layout, ok := container.Layout().(*unison.FlexLayout); ok {
		if layout.Columns == 1 {
			layout.Columns = 3
			container.AddChildAtIndex(s.scroll, 0)
			container.AddChildAtIndex(s.divider, 1)
		} else {
			if focus := container.Window().CurrentFocus(); focus != nil &&
				s.scroll.HasInSelfOrDescendants(func(p *unison.Panel) bool { return p == focus }) {
				document.AsPanel().RequestFocus()
			}
			layout.Columns = 1
			s.divider.RemoveFromParent()
			s.scroll.RemoveFromParent()
		}
		container.MarkForLayoutAndRedraw()
	}
}

// fitTo raises the side bar's preferred width to the given width of its content, though to no more than
// sideBarMaxFitWidth, and never lowers it. The change takes effect at the next layout.
func (s *sideBar) fitTo(width float32) {
	s.preferred = max(min(width, sideBarMaxFitWidth), s.preferred)
}

// maxWidth returns the most the side bar is laid out at: sideBarMaxFraction of its container's content width, or no
// limit while it has no container or the container has yet to be given a width.
func (s *sideBar) maxWidth() float32 {
	if container := s.scroll.Parent(); container != nil {
		if width := container.ContentRect(false).Width; width > 0 {
			return xmath.Floor(width * sideBarMaxFraction)
		}
	}
	return math.MaxFloat32
}

// LayoutSizes implements unison.Layout for the side bar's scroll panel, giving the width the side bar is to be laid out
// at as its preferred width. The container's FlexLayout asks for it as the container is laid out, with the container's
// width already set.
func (s *sideBar) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	minSize, prefSize, maxSize = s.scroll.LayoutSizes(target, hint)
	prefSize.Width = min(s.preferred, s.maxWidth())
	return minSize, prefSize, maxSize
}

// PerformLayout implements unison.Layout for the side bar's scroll panel.
func (s *sideBar) PerformLayout(target *unison.Panel) {
	s.scroll.PerformLayout(target)
}
