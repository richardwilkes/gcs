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
	"github.com/richardwilkes/unison/enums/mod"
)

// scrollPanDrag pans a scroll panel by dragging its content: the scroll position captured when the drag began is
// offset by how far the pointer has moved since, so the content follows the pointer. Points are tracked in root
// coordinates, since the content panel's own coordinate space shifts underneath the pointer as it scrolls.
type scrollPanDrag struct {
	scroll  *unison.ScrollPanel
	content *unison.Panel
	start   geom.Point
	origin  geom.Point
	active  bool
}

// install wires the pan state to scroll and content, and sets content's mouse and cursor callbacks so that dragging it
// pans the scroll panel. A mouse-down also gives focus to the given panel. Dockables whose content only pans should use
// this rather than installing the callbacks themselves, so that the wiring cannot be forgotten.
func (p *scrollPanDrag) install(scroll *unison.ScrollPanel, content *unison.Panel, focus unison.Paneler) {
	p.scroll = scroll
	p.content = content
	content.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return p.cursor() }
	content.MouseDownCallback = func(where geom.Point, _, _ int, _ mod.Modifiers) bool {
		p.begin(where)
		focus.AsPanel().RequestFocus()
		content.UpdateCursorNow()
		return true
	}
	content.MouseDragCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		p.drag(where)
		return true
	}
	content.MouseUpCallback = func(_ geom.Point, _ int, _ mod.Modifiers) bool {
		p.end()
		content.UpdateCursorNow()
		return true
	}
}

// begin starts a drag at the given content-local point.
func (p *scrollPanDrag) begin(where geom.Point) {
	p.start = p.content.PointToRoot(where)
	p.origin.X, p.origin.Y = p.scroll.Position()
	p.active = true
}

// drag scrolls so the content keeps pace with the pointer, now at the given content-local point.
func (p *scrollPanDrag) drag(where geom.Point) {
	pt := p.start.Sub(p.content.PointToRoot(where)).Add(p.origin)
	p.scroll.SetPosition(pt.X, pt.Y)
}

// end finishes the drag.
func (p *scrollPanDrag) end() {
	p.active = false
}

// cursor returns the cursor to show over the content: a move cursor while dragging, the arrow otherwise.
func (p *scrollPanDrag) cursor() *unison.Cursor {
	if p.active {
		return unison.MoveCursor()
	}
	return unison.ArrowCursor()
}
