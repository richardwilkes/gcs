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

package feature

// AllWithoutContainedWeightType holds the possible Type values, minus the ContainedWeightReductionType.
var AllWithoutContainedWeightType []Type

func init() {
	for _, one := range AllType {
		if one != ContainedWeightReductionType {
			AllWithoutContainedWeightType = append(AllWithoutContainedWeightType, one)
		}
	}
}
