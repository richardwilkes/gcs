// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package frequency

import (
	"fmt"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible Frequency values.
const (
	None     = Roll(0)
	FR6      = Roll(6)
	FR9      = Roll(9)
	FR12     = Roll(12)
	FR15     = Roll(15)
	Constant = Roll(18)
)

// Rolls is the complete set of Roll values.
var Rolls = []Roll{
	None,
	FR6,
	FR9,
	FR12,
	FR15,
	Constant,
}

// Roll holds the information about a frequency of appearance roll, from B36.
type Roll byte

// EnsureValid ensures this is of a known value.
func (r Roll) EnsureValid() Roll {
	if slices.Contains(Rolls, r) {
		return r
	}
	return Rolls[0]
}

// String implements fmt.Stringer.
func (r Roll) String() string {
	var text string
	switch r {
	case FR6:
		text = i18n.Text("Quite rarely")
	case FR9:
		text = i18n.Text("Fairly often")
	case FR12:
		text = i18n.Text("Quite often")
	case FR15:
		text = i18n.Text("Almost all the time")
	case Constant:
		return i18n.Text("Constantly")
	default:
		return i18n.Text("None required")
	}
	return fmt.Sprintf("%d or less (%s)", r, text)
}

// ShortString returns a short description of the frequency.
func (r Roll) ShortString() string {
	switch r {
	case Constant:
		return i18n.Text("No FR")
	case FR6, FR9, FR12, FR15:
		return fmt.Sprintf(i18n.Text("FR%d"), r)
	default:
		return ""
	}
}

// Multiplier returns the cost multiplier.
func (r Roll) Multiplier() fxp.Int {
	switch r {
	case None:
		return fxp.One
	case FR6:
		return fxp.Half
	case FR9:
		return fxp.One
	case FR12:
		return fxp.Two
	case FR15:
		return fxp.Three
	case Constant:
		return fxp.Four
	default:
		return None.Multiplier()
	}
}
