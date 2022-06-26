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

package measure

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// TrailingWeightUnitsFromString extracts a trailing WeightUnits from a string.
func TrailingWeightUnitsFromString(s string, defUnits WeightUnits) WeightUnits {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, one := range AllWeightUnits {
		if strings.HasSuffix(s, one.Key()) {
			return one
		}
	}
	return defUnits
}

// Format the weight for this WeightUnits.
func (enum WeightUnits) Format(weight Weight) string {
	switch enum {
	case Pound, PoundAlt:
		return fxp.Int(weight).String() + " " + enum.Key()
	case Ounce:
		return fxp.Int(weight).Mul(fxp.From(16)).String() + " " + enum.Key()
	case Ton, TonAlt:
		return fxp.Int(weight).Div(fxp.From(2000)).String() + " " + enum.Key()
	case Kilogram:
		return fxp.Int(weight).Div(fxp.From(2)).String() + " " + enum.Key()
	case Gram:
		return fxp.Int(weight).Mul(fxp.From(500)).String() + " " + enum.Key()
	default:
		return Pound.Format(weight)
	}
}

// ToPounds the weight for this WeightUnits.
func (enum WeightUnits) ToPounds(weight fxp.Int) fxp.Int {
	switch enum {
	case Pound, PoundAlt:
		return weight
	case Ounce:
		return weight.Div(fxp.From(16))
	case Ton, TonAlt:
		return weight.Mul(fxp.From(2000))
	case Kilogram:
		return weight.Mul(fxp.From(2))
	case Gram:
		return weight.Div(fxp.From(500))
	default:
		return Pound.ToPounds(weight)
	}
}
