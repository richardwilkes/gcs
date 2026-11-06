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
	"github.com/richardwilkes/unison/enums/mod"
)

// TestScaleChangesUndoOneAtATime verifies that Undo takes back each scale change made by a command on its own, and an
// uninterrupted run of Option+wheel steps as one, which any other change of scale, Undo included, ends.
func TestScaleChangesUndoOneAtATime(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	initial := d.scale
	mgr := unison.UndoManagerFor(d.table)
	wheel := func(dy float32) {
		d.scroll.ContentView().MouseWheelCallback(geom.Point{}, geom.NewPoint(0, dy), mod.Option)
	}
	undo := func(want int) {
		t.Helper()
		c.True(mgr.CanUndo(), "Undo must be on before restoring %d%%", want)
		mgr.Undo()
		c.Equal(want, d.scale, "Undo must restore %d%%", want)
	}
	d.PerformCmd(nil, Scale200ItemID)
	d.PerformCmd(nil, Scale300ItemID)
	wheel(1)
	wheel(1)
	d.PerformCmd(nil, ScaleDownItemID)
	undo(320)
	wheel(1)
	c.Equal(330, d.scale, "the wheel must step up from 320%")
	for _, want := range []int{320, 300, 200, initial} {
		undo(want)
	}
	c.False(mgr.CanUndo(), "no scale change may be left to undo")
}

// TestScaleCommandsFollowFieldEnabledState verifies that the scale commands are offered only while the scale field is
// enabled, since the field puts back its old value when one acts while it is disabled.
func TestScaleCommandsFollowFieldEnabledState(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	d.PerformCmd(nil, Scale200ItemID)
	ids := []int{ScaleDefaultItemID, ScaleUpItemID, ScaleDownItemID, Scale100ItemID}
	for _, id := range ids {
		c.True(d.CanPerformCmd(nil, id), "command %d must be offered while the scale field is enabled", id)
	}
	d.scaleField.SetEnabled(false)
	for _, id := range ids {
		c.False(d.CanPerformCmd(nil, id), "command %d must not be offered while the scale field is disabled", id)
	}
}
