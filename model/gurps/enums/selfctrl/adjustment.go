// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package selfctrl

import "fmt"

// Adjustment returns the adjustment amount.
func (enum Adjustment) Adjustment(sc Roll) int {
	if sc == None {
		return 0
	}
	switch enum {
	case NoAdjustment:
		return 0
	case ActionPenalty:
		return sc.Penalty()
	case ReactionPenalty:
		return sc.Penalty()
	case FrightCheckPenalty:
		return sc.Penalty()
	case FrightCheckBonus:
		return -sc.Penalty()
	case MinorCostOfLivingIncrease:
		return -sc.Penalty() * 5
	case MajorCostOfLivingIncrease:
		return 10 * (1 << (-(sc.Penalty() + 1)))
	default:
		return NoAdjustment.Adjustment(sc)
	}
}

// Description returns a formatted description.
func (enum Adjustment) Description(sc Roll) string {
	switch {
	case sc == None:
		return ""
	case enum == NoAdjustment:
		return enum.AltString()
	default:
		return fmt.Sprintf(enum.AltString(), enum.Adjustment(sc))
	}
}
