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

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestLibraryUpdateButtonIgnoresNonPrimaryPresses verifies that a library's Update button starts an update only for a
// click of the primary button: a right-drag unison hands back to the table reaches the button like any other press.
func TestLibraryUpdateButtonIgnoresNonPrimaryPresses(t *testing.T) {
	c := check.New(t)
	cell := newUpdatableLibraryCell(&library.Library{}, unison.NewLabel(), &library.Release{Version: "v1.2.3"})
	clicks := 0
	cell.button.ClickCallback = func() { clicks++ }
	cell.button.SetFrameRect(geom.NewRect(0, 0, 100, 40))
	inside := geom.NewPoint(50, 20)
	c.True(inside.In(cell.button.ContentRect(false)), "the test needs a point within the button")

	for _, button := range []int{unison.ButtonRight, unison.ButtonMiddle} {
		c.False(cell.button.MouseDownCallback(inside, button, 1, mod.None),
			"a non-primary press must be left for the table")
		c.False(cell.button.Pressed, "a non-primary press must not show the button pressed")
		c.False(cell.button.MouseDragCallback(inside, button, mod.None),
			"a drag from a press the button didn't take must be left for the table as well")
		c.False(cell.button.Pressed, "nor may the drag show the button pressed")
		c.False(cell.button.MouseUpCallback(inside, button, mod.None),
			"the release of a press the button didn't take must be left for the table as well")
		c.Equal(0, clicks, "a non-primary release over the button must not start an update")
	}

	c.True(cell.button.MouseDownCallback(inside, unison.ButtonLeft, 1, mod.None), "a primary press must be taken")
	c.True(cell.button.Pressed, "and show the button pressed")
	c.True(cell.button.MouseDragCallback(inside, unison.ButtonLeft, mod.None), "along with its drags")
	c.True(cell.button.MouseUpCallback(inside, unison.ButtonLeft, mod.None), "and its release")
	c.False(cell.button.Pressed, "which must show the button released")
	c.Equal(1, clicks, "a primary click must start the update")
}
