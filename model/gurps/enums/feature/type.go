// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package feature

// TypesWithoutContainedWeightReduction holds the possible Type values, minus the ContainedWeightReduction.
var TypesWithoutContainedWeightReduction []Type

func init() {
	TypesWithoutContainedWeightReduction = make([]Type, 0, len(Types)-1)
	for _, one := range Types {
		if one != ContainedWeightReduction {
			TypesWithoutContainedWeightReduction = append(TypesWithoutContainedWeightReduction, one)
		}
	}
}
