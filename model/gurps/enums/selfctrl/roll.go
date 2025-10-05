// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package selfctrl

import (
	"fmt"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible Roll values.
const (
	None   = Roll(0)
	Always = Roll(1)
	CR6    = Roll(6)
	CR7    = Roll(7)
	CR8    = Roll(8)
	CR9    = Roll(9)
	CR10   = Roll(10)
	CR11   = Roll(11)
	CR12   = Roll(12)
	CR13   = Roll(13)
	CR14   = Roll(14)
	CR15   = Roll(15)
)

// Rolls is the complete set of Roll values.
var Rolls = []Roll{
	None,
	Always,
	CR6,
	CR7,
	CR8,
	CR9,
	CR10,
	CR11,
	CR12,
	CR13,
	CR14,
	CR15,
}

// Roll holds the information about a self-control roll, from B121 and Z60.
type Roll byte

// EnsureValid ensures this is of a known value.
func (r Roll) EnsureValid() Roll {
	if slices.Contains(Rolls, r) {
		return r
	}
	return None
}

// String implements fmt.Stringer.
func (r Roll) String() string {
	var text string
	switch {
	case r == None:
		return i18n.Text("None required")
	case r == Always:
		return i18n.Text("Never resist")
	case r < CR8:
		text = i18n.Text("Resist rarely")
	case r < CR11:
		text = i18n.Text("Resist fairly often")
	case r < CR14:
		text = i18n.Text("Resist quite often")
	case r <= CR15:
		text = i18n.Text("Resist almost all the time")
	default:
		return None.String()
	}
	var nonStandard string
	if r%3 != 0 {
		nonStandard = i18n.Text("; non-standard")
	}
	return fmt.Sprintf("%d or less (%s%s)", r, text, nonStandard)
}

// ShortString returns a short description of the frequency.
func (r Roll) ShortString() string {
	switch {
	case r == Always:
		return i18n.Text("No CR")
	case r >= CR6 && r <= CR15:
		return fmt.Sprintf(i18n.Text("CR%d"), r)
	default:
		return ""
	}
}

// Multiplier returns the cost multiplier.
func (r Roll) Multiplier() fxp.Int {
	switch r {
	case None:
		return fxp.One
	case Always:
		return fxp.TwoAndAHalf
	case CR6:
		return fxp.Two
	case CR7:
		return fxp.FromStringForced("1.83")
	case CR8:
		return fxp.FromStringForced("1.67")
	case CR9:
		return fxp.OneAndAHalf
	case CR10:
		return fxp.FromStringForced("1.33")
	case CR11:
		return fxp.FromStringForced("1.17")
	case CR12:
		return fxp.One
	case CR13:
		return fxp.FromStringForced("0.83")
	case CR14:
		return fxp.FromStringForced("0.67")
	case CR15:
		return fxp.Half
	default:
		return None.Multiplier()
	}
}

// Penalty returns the general penalty for this roll.
func (r Roll) Penalty() int {
	switch r {
	case None:
		return 0
	case CR14, CR15:
		return -1
	case CR11, CR12, CR13:
		return -2
	case CR8, CR9, CR10:
		return -3
	case CR6, CR7:
		return -4
	case Always:
		return -5 // No actual rules for this from Z60, so just extend the progression
	default:
		return 0
	}
}
