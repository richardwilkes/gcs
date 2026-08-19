// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// TestForSortDistinguishesSwitchStates verifies that the three states a switch column can be in sort as three distinct
// groups. The cells are deliberately drawn as three different things -- a checkmark for on, a dash for off, and a blank
// cell for a row with nothing to switch -- and a row with nothing to switch never gets a switch cell at all, so it
// reaches ForSort as the zero value, i.e. an empty text cell. The off state used to share the toggle's empty result,
// which interleaved those rows with the ones that have no switch whenever the column was sorted.
func TestForSortDistinguishesSwitchStates(t *testing.T) {
	c := check.New(t)
	on := (&CellData{Type: cell.Switch, Checked: true}).ForSort()
	off := (&CellData{Type: cell.Switch}).ForSort()
	nothingToSwitch := (&CellData{}).ForSort()

	c.NotEqual(on, off, "an on switch must not sort as an off switch")
	c.NotEqual(off, nothingToSwitch, "an off switch must not sort as a row with nothing to switch")
	c.NotEqual(on, nothingToSwitch, "an on switch must not sort as a row with nothing to switch")
	c.Equal("", nothingToSwitch, "a row with nothing to switch must sort as nothing")
	c.NotEqual("", off, "an off switch must sort as something")

	// The table sorts with xstrings.NaturalLess, so check the grouping the chosen values actually produce.
	c.True(xstrings.NaturalLess(nothingToSwitch, off, true), "rows with nothing to switch sort ahead of off switches")
	c.True(xstrings.NaturalLess(off, on, true), "off switches sort ahead of on switches")

	// A search of the sheet matches against the same strings, so the off state must not be something that could be
	// typed into the search field while looking for something else.
	c.NotContains(off, "-", "an off switch must not be matched by a search for an ASCII hyphen")

	// Toggles remain a two-state cell: they are present on every row of their column, so there is no third state to
	// distinguish and the off state stays empty.
	c.Equal(on, (&CellData{Type: cell.Toggle, Checked: true}).ForSort(), "an on toggle sorts as an on switch does")
	c.Equal("", (&CellData{Type: cell.Toggle}).ForSort(), "an off toggle sorts as nothing")
}
