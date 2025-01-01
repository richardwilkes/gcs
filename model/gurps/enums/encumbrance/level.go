// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package encumbrance

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// WeightMultiplier returns the weight multiplier associated with the encumbrance level.
func (enum Level) WeightMultiplier() fxp.Int {
	switch enum {
	case No:
		return fxp.One
	case Light:
		return fxp.Two
	case Medium:
		return fxp.Three
	case Heavy:
		return fxp.Six
	case ExtraHeavy:
		return fxp.Ten
	default:
		return No.WeightMultiplier()
	}
}

// Penalty returns the penalty associated with the encumbrance level.
func (enum Level) Penalty() fxp.Int {
	switch enum {
	case No:
		return 0
	case Light:
		return -fxp.One
	case Medium:
		return -fxp.Two
	case Heavy:
		return -fxp.Three
	case ExtraHeavy:
		return -fxp.Four
	default:
		return No.Penalty()
	}
}
