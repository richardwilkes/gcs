/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"time"

	"github.com/richardwilkes/unison"
)

// NewLink creates a new clickable link.
func NewLink(title string, callback func()) *unison.Label {
	label := unison.NewLabel()
	label.Text = title
	label.Underline = true
	over := false
	pressed := false
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		if over {
			if pressed {
				label.OnBackgroundInk = unison.LinkPressedColor
			} else {
				label.OnBackgroundInk = unison.LinkRolloverColor
			}
		} else {
			label.OnBackgroundInk = unison.LinkColor
		}
		label.DefaultDraw(gc, rect)
	}
	label.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor {
		return unison.PointingCursor()
	}
	label.MouseEnterCallback = func(where unison.Point, _ unison.Modifiers) bool {
		over = true
		label.MarkForRedraw()
		return true
	}
	label.MouseMoveCallback = func(where unison.Point, _ unison.Modifiers) bool {
		if over != label.ContentRect(true).ContainsPoint(where) {
			over = !over
			label.MarkForRedraw()
		}
		return true
	}
	label.MouseExitCallback = func() bool {
		over = false
		label.MarkForRedraw()
		return true
	}
	label.MouseDownCallback = func(where unison.Point, _, _ int, _ unison.Modifiers) bool {
		pressed = label.ContentRect(true).ContainsPoint(where)
		label.MarkForRedraw()
		return true
	}
	label.MouseDragCallback = func(where unison.Point, _ int, _ unison.Modifiers) bool {
		in := label.ContentRect(true).ContainsPoint(where)
		if pressed != in {
			pressed = in
			label.MarkForRedraw()
		}
		return true
	}
	label.MouseUpCallback = func(where unison.Point, _ int, _ unison.Modifiers) bool {
		if over = label.ContentRect(true).ContainsPoint(where); over {
			unison.InvokeTaskAfter(callback, time.Millisecond)
		}
		pressed = false
		label.MarkForRedraw()
		return true
	}
	return label
}
