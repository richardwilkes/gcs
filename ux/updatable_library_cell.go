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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

type updatableLibraryCell struct {
	unison.Panel
	library *library.Library
	release *library.Release
	title   *unison.Label
	button  *unison.Button
}

func newUpdatableLibraryCell(lib *library.Library, title *unison.Label, rel *library.Release) *updatableLibraryCell {
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
	// Only the primary button works the button: unison hands a right-drag on the row back to the table, which forwards
	// it here like any other press, and the default handling would fire the click for any button released over it.
	c.button.MouseDownCallback = func(where geom.Point, btn, clickCount int, mods mod.Modifiers) bool {
		return btn == unison.ButtonLeft && c.button.DefaultMouseDown(where, btn, clickCount, mods)
	}
	c.button.MouseDragCallback = func(where geom.Point, btn int, mods mod.Modifiers) bool {
		return btn == unison.ButtonLeft && c.button.DefaultMouseDrag(where, btn, mods)
	}
	c.button.MouseUpCallback = func(where geom.Point, btn int, mods mod.Modifiers) bool {
		return btn == unison.ButtonLeft && c.button.DefaultMouseUp(where, btn, mods)
	}
	c.AddChild(c.button)
	return c
}

func (c *updatableLibraryCell) updateForeground(fg unison.Ink) {
	c.title.OnBackgroundInk = fg
	c.title.SetTitle(c.title.String())
}
