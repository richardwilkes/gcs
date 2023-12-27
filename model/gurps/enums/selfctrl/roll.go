/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package selfctrl

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
)

// Possible Roll values.
const (
	NoCR = Roll(0)
	CR6  = Roll(6)
	CR9  = Roll(9)
	CR12 = Roll(12)
	CR15 = Roll(15)
)

// Rolls is the complete set of Roll values.
var Rolls = []Roll{
	NoCR,
	CR6,
	CR9,
	CR12,
	CR15,
}

// Roll holds the information about a self-control roll, from B121.
type Roll int

// EnsureValid ensures this is of a known value.
func (s Roll) EnsureValid() Roll {
	for _, one := range Rolls {
		if one == s {
			return s
		}
	}
	return Rolls[0]
}

// Index returns of the Roll within Rolls.
func (s Roll) Index() int {
	for i, one := range Rolls {
		if one == s {
			return i
		}
	}
	return 0
}

// String implements fmt.Stringer.
func (s Roll) String() string {
	switch s {
	case NoCR:
		return i18n.Text("None Required")
	case CR6:
		return i18n.Text("CR: 6 (Resist rarely)")
	case CR9:
		return i18n.Text("CR: 9 (Resist fairly often)")
	case CR12:
		return i18n.Text("CR: 12 (Resist quite often)")
	case CR15:
		return i18n.Text("CR: 15 (Resist almost all the time)")
	default:
		return NoCR.String()
	}
}

// DescriptionWithCost returns a formatted description that includes the cost multiplier.
func (s Roll) DescriptionWithCost() string {
	v := s.EnsureValid()
	if v == NoCR {
		return ""
	}
	return v.String() + ", x" + v.Multiplier().String()
}

// Multiplier returns the cost multiplier.
func (s Roll) Multiplier() fxp.Int {
	switch s {
	case NoCR:
		return fxp.One
	case CR6:
		return fxp.Two
	case CR9:
		return fxp.OneAndAHalf
	case CR12:
		return fxp.One
	case CR15:
		return fxp.Half
	default:
		return NoCR.Multiplier()
	}
}

// MinimumRoll returns the minimum roll to retain control.
func (s Roll) MinimumRoll() int {
	return int(s.EnsureValid())
}
