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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
)

// Possible Roll values.
const (
	NoCR   = Roll(0)
	CRNone = Roll(1)
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
	NoCR,
	CRNone,
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
func (s Roll) EnsureValid() Roll {
	for _, one := range Rolls {
		if one == s {
			return s
		}
	}
	return Rolls[0]
}

// String implements fmt.Stringer.
func (s Roll) String() string {
	switch s {
	case NoCR:
		return i18n.Text("None Required")
	case CRNone:
		return i18n.Text("None Allowed")
	case CR6:
		return i18n.Text("CR: 6 (Resist rarely)")
	case CR9:
		return i18n.Text("CR: 9 (Resist fairly often)")
	case CR12:
		return i18n.Text("CR: 12 (Resist quite often)")
	case CR15:
		return i18n.Text("CR: 15 (Resist almost all the time)")
	case CR7, CR8, CR10, CR11, CR13, CR14:
		return fmt.Sprintf(i18n.Text("CR: %d (non-standard)"), s)
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
	case CRNone:
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
		return NoCR.Multiplier()
	}
}

// Penalty returns the general penalty for this roll.
func (s Roll) Penalty() int {
	switch s {
	case NoCR:
		return 0
	case CR14, CR15:
		return -1
	case CR11, CR12, CR13:
		return -2
	case CR8, CR9, CR10:
		return -3
	case CR6, CR7:
		return -4
	case CRNone:
		return -5 // No actual rules for this from Z60, so just extend the progression
	default:
		return 0
	}
}
