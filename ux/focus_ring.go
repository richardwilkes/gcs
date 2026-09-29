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
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/pathop"
	"github.com/richardwilkes/unison/enums/role"
)

// installStaticTextFocusRing draws a ring around static text (a label, read-only field or heading) that holds the
// window's keyboard focus, since such text draws nothing to show it. Static text takes the focus only while a screen
// reader is served (see unison.SetFocusForReading), but a sighted person may be watching, and on macOS unison stays
// active after an assistive technology other than VoiceOver or Switch Control quits.
//
// The ring is drawn over the window's content, so install it after the content has been set.
func installStaticTextFocusRing(wnd *unison.Window) {
	content := wnd.Content()
	prior := content.DrawOverCallback
	content.DrawOverCallback = func(gc *unison.Canvas, rect geom.Rect) {
		if prior != nil {
			prior(gc, rect)
		}
		if ring, clip, ok := staticTextFocusRing(wnd, content); ok {
			gc.Save()
			gc.ClipRect(clip, pathop.Intersect, false)
			paint := unison.ThemeFocus.Paint(gc, ring, paintstyle.Stroke)
			paint.SetStrokeWidth(1)
			gc.DrawRect(ring.Inset(geom.NewUniformInsets(0.5)), paint)
			gc.Restore()
		}
	}
}

// staticTextFocusRing returns, in content coordinates, the ring around the static text holding the focused window's
// keyboard focus and the part of it that the text's ancestors leave visible, so a ring around text partly scrolled out
// of view stops at the view's edge. ok is false when no such text in content holds the focus or none of it is visible.
func staticTextFocusRing(wnd *unison.Window, content *unison.Panel) (ring, clip geom.Rect, ok bool) {
	focus := wnd.Focus()
	if focus == nil || !wnd.Focused() || !showsNoFocus(focus) {
		return ring, clip, false
	}
	bounds := focus.ContentRect(true)
	ring = focus.RectToRoot(staticTextBounds(focus, bounds))
	clip = focus.RectToRoot(bounds)
	inContent := false
	for ancestor := focus.Parent(); ancestor != nil; ancestor = ancestor.Parent() {
		clip = clip.Intersect(ancestor.RectToRoot(ancestor.ContentRect(true)))
		if ancestor == content {
			inContent = true
			break
		}
	}
	if !inContent || clip.Empty() {
		return ring, clip, false
	}
	return content.RectFromRoot(ring), content.RectFromRoot(clip), true
}

// staticTextBounds returns where the ring goes within a panel's bounds. A label is often wider than its text, so its
// ring goes around the text it draws, with a little room to either side where it fits. Other panels are ringed whole.
func staticTextBounds(p *unison.Panel, bounds geom.Rect) geom.Rect {
	label, ok := p.Self.(*unison.Label)
	if !ok {
		return bounds
	}
	_, size, _ := label.Sizes(geom.Size{})
	size.Width += focusRingRoom * 2
	rect := bounds
	if size.Width < bounds.Width {
		rect.Width = size.Width
		switch label.HAlign {
		case align.Middle:
			rect.X += (bounds.Width - size.Width) / 2
		case align.End:
			rect.X += bounds.Width - size.Width
		default:
			rect.X -= focusRingRoom
		}
	}
	if size.Height < bounds.Height {
		rect.Height = size.Height
		switch label.VAlign {
		case align.Middle:
			rect.Y += (bounds.Height - size.Height) / 2
		case align.End:
			rect.Y += bounds.Height - size.Height
		default:
		}
	}
	return rect.Intersect(bounds)
}

// showsNoFocus reports whether a panel is static text that draws nothing to show it holds the keyboard focus: a
// unison.Label left in the auto role, or any panel in the label or heading role. A textLabel draws its own ring.
func showsNoFocus(p *unison.Panel) bool {
	if _, ok := p.Self.(*textLabel); ok {
		return false
	}
	switch p.Accessibility.Role {
	case role.Label, role.Heading:
		return true
	case role.Auto:
		_, ok := p.Self.(*unison.Label)
		return ok
	default:
		return false
	}
}
