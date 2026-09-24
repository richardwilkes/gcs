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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestGroupContainersFirst verifies that a column's own comparison is handed the value a cell shows rather than the
// marker a container's sort text carries, and that containers still group ahead of everything else. Without this, a
// comparison that reads a number out of the text sees the marker instead and reports every container as equal.
func TestGroupContainersFirst(t *testing.T) {
	c := check.New(t)

	less := groupContainersFirst(gurps.PointsLessFromString)
	c.NotNil(less, "a column with a comparison of its own keeps one")
	c.Nil(groupContainersFirst(nil), "a column with no comparison of its own is left without one")

	c.True(less("9", "10"), "ordinary rows compare by value")
	c.True(less(containerMarker+"100", "9"), "containers group ahead of everything else")
	c.False(less("9", containerMarker+"100"), "containers group ahead of everything else, from either side")
	c.True(less(containerMarker+"9", containerMarker+"10"), "containers compare by value among themselves")
	c.True(less(containerMarker+"9~20", containerMarker+"10"), "a container showing a range compares by its low end")
}
