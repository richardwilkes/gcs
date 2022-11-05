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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

type updatableLibraryCell struct {
	unison.Panel
	library           *library.Library
	release           library.Release
	title             *unison.Label
	button            *unison.Button
	inButtonMouseDown bool
	inPanel           bool
	overButton        bool
}

func newUpdatableLibraryCell(lib *library.Library, title *unison.Label, rel *library.Release) *updatableLibraryCell {
	c := &updatableLibraryCell{
		library: lib,
		release: *rel,
		title:   title,
	}
	c.Self = &c
	c.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})

	c.AddChild(title)

	c.button = unison.NewButton()
	c.button.Text = fmt.Sprintf("Update to v%s", filterVersion(rel.Version))
	fd := c.button.Font.Descriptor()
	fd.Size = xmath.Round(fd.Size * 0.8)
	c.button.Font = fd.Font()
	c.button.ClickCallback = func() { initiateLibraryUpdate(c.library, c.release) }
	c.AddChild(c.button)

	c.MouseDownCallback = c.mouseDown
	c.MouseDragCallback = c.mouseDrag
	c.MouseUpCallback = c.mouseUp
	c.MouseEnterCallback = c.mouseEnter
	c.MouseMoveCallback = c.mouseMove
	c.MouseExitCallback = c.mouseExit
	return c
}

func (c *updatableLibraryCell) updateForeground(fg unison.Ink) {
	c.title.OnBackgroundInk = fg
}

func (c *updatableLibraryCell) mouseDown(where unison.Point, btn, clickCount int, mod unison.Modifiers) bool {
	if !c.button.FrameRect().ContainsPoint(where) {
		return false
	}
	c.inButtonMouseDown = true
	return c.button.DefaultMouseDown(c.button.PointFromRoot(c.PointToRoot(where)), btn, clickCount, mod)
}

func (c *updatableLibraryCell) mouseDrag(where unison.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	return c.button.DefaultMouseDrag(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}

func (c *updatableLibraryCell) mouseUp(where unison.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	c.inButtonMouseDown = false
	return c.button.DefaultMouseUp(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}

func (c *updatableLibraryCell) mouseEnter(where unison.Point, mod unison.Modifiers) bool {
	c.inPanel = true
	if !c.button.FrameRect().ContainsPoint(where) {
		return false
	}
	c.overButton = true
	return c.button.DefaultMouseEnter(c.button.PointFromRoot(c.PointToRoot(where)), mod)
}

func (c *updatableLibraryCell) mouseMove(where unison.Point, mod unison.Modifiers) bool {
	if c.inPanel {
		over := c.button.FrameRect().ContainsPoint(where)
		if over != c.overButton {
			if over {
				c.overButton = true
				return c.button.DefaultMouseEnter(c.button.PointFromRoot(c.PointToRoot(where)), mod)
			}
			c.overButton = false
			return c.button.DefaultMouseExit()
		}
	}
	return false
}

func (c *updatableLibraryCell) mouseExit() bool {
	c.inPanel = false
	if !c.overButton {
		return false
	}
	c.overButton = false
	return c.button.DefaultMouseExit()
}
