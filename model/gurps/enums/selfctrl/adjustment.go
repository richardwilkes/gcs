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
func (enum Adjustment) Adjustment(cr Roll) int {
	if cr == NoCR {
		return 0
	}
	switch enum {
	case NoCRAdj:
		return 0
	case ActionPenalty:
		return cr.Penalty()
	case ReactionPenalty:
		return cr.Penalty()
	case FrightCheckPenalty:
		return cr.Penalty()
	case FrightCheckBonus:
		return -cr.Penalty()
	case MinorCostOfLivingIncrease:
		return -cr.Penalty() * 5
	case MajorCostOfLivingIncrease:
		return 10 * (1 << (-(cr.Penalty() + 1)))
	default:
		return NoCRAdj.Adjustment(cr)
	}
}

// Description returns a formatted description.
func (enum Adjustment) Description(cr Roll) string {
	switch {
	case cr == NoCR:
		return ""
	case enum == NoCRAdj:
		return enum.AltString()
	default:
		return fmt.Sprintf(enum.AltString(), enum.Adjustment(cr))
	}
}
