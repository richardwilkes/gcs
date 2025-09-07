// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
)

type updatableLibraryCell struct {
	unison.Panel
	library           *gurps.Library
	release           gurps.Release
	title             *unison.Label
	button            *unison.Button
	inButtonMouseDown bool
}

func newUpdatableLibraryCell(lib *gurps.Library, title *unison.Label, rel gurps.Release) *updatableLibraryCell {
	c := &updatableLibraryCell{
		library: lib,
		release: rel,
		title:   title,
	}
	c.Self = c
	c.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})

	c.AddChild(title)

	c.button = unison.NewButton()
	fd := c.button.Font.Descriptor()
	fd.Size = xmath.Round(fd.Size * 0.8)
	c.button.Font = fd.Font()
	version := filterVersion(rel.Version)
	if strings.HasPrefix(version, "v") {
		version = i18n.Text("Update to") + " " + version
	} else {
		version = i18n.Text("Update")
	}
	c.button.SetTitle(version)
	c.button.ClickCallback = func() { initiateLibraryUpdate(c.library, c.release) }
	c.AddChild(c.button)

	c.MouseDownCallback = c.mouseDown
	c.MouseDragCallback = c.mouseDrag
	c.MouseUpCallback = c.mouseUp
	return c
}

func (c *updatableLibraryCell) updateForeground(fg unison.Ink) {
	c.title.OnBackgroundInk = fg
	c.title.SetTitle(c.title.String())
}

func (c *updatableLibraryCell) mouseDown(where geom.Point, btn, clickCount int, mod unison.Modifiers) bool {
	if !where.In(c.button.FrameRect()) {
		return false
	}
	c.inButtonMouseDown = true
	return c.button.DefaultMouseDown(c.button.PointFromRoot(c.PointToRoot(where)), btn, clickCount, mod)
}

func (c *updatableLibraryCell) mouseDrag(where geom.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	return c.button.DefaultMouseDrag(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}

func (c *updatableLibraryCell) mouseUp(where geom.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	c.inButtonMouseDown = false
	return c.button.DefaultMouseUp(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}
