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

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// WeightMultiplier returns the weight multiplier associated with the Encumbrance level.
func (enum Encumbrance) WeightMultiplier() fxp.Int {
	switch enum {
	case NoEncumbrance:
		return fxp.One
	case LightEncumbrance:
		return fxp.Two
	case MediumEncumbrance:
		return fxp.Three
	case HeavyEncumbrance:
		return fxp.Six
	case ExtraHeavyEncumbrance:
		return fxp.Ten
	default:
		return NoEncumbrance.WeightMultiplier()
	}
}

// Penalty returns the penalty associated with the Encumbrance level.
func (enum Encumbrance) Penalty() fxp.Int {
	switch enum {
	case NoEncumbrance:
		return 0
	case LightEncumbrance:
		return -fxp.One
	case MediumEncumbrance:
		return -fxp.Two
	case HeavyEncumbrance:
		return -fxp.Three
	case ExtraHeavyEncumbrance:
		return -fxp.Four
	default:
		return NoEncumbrance.Penalty()
	}
}
