/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Adjustment returns the adjustment amount.
func (enum SelfControlRollAdj) Adjustment(cr SelfControlRoll) int {
	if cr == NoCR {
		return 0
	}
	switch enum {
	case NoCRAdj:
		return 0
	case ActionPenalty:
		return cr.Index() - len(AllSelfControlRolls)
	case ReactionPenalty:
		return cr.Index() - len(AllSelfControlRolls)
	case FrightCheckPenalty:
		return cr.Index() - len(AllSelfControlRolls)
	case FrightCheckBonus:
		return len(AllSelfControlRolls) - cr.Index()
	case MinorCostOfLivingIncrease:
		return 5 * (len(AllSelfControlRolls) - cr.Index())
	case MajorCostOfLivingIncrease:
		return 10 * (1 << (len(AllSelfControlRolls) - (cr.Index() + 1)))
	default:
		return NoCRAdj.Adjustment(cr)
	}
}

// Description returns a formatted description.
func (enum SelfControlRollAdj) Description(cr SelfControlRoll) string {
	switch {
	case cr == NoCR:
		return ""
	case enum == NoCRAdj:
		return enum.AltString()
	default:
		return fmt.Sprintf(enum.AltString(), enum.Adjustment(cr))
	}
}

// Features returns the set of features to apply.
func (enum SelfControlRollAdj) Features(cr SelfControlRoll) Features {
	if enum.EnsureValid() != MajorCostOfLivingIncrease {
		return nil
	}
	f := NewSkillBonus()
	f.NameCriteria.Qualifier = "Merchant"
	f.Amount = fxp.From(cr.Index() - len(AllSelfControlRolls))
	return Features{f}
}
