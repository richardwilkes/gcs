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

package ancestry

import (
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// StringOption is a string that has a weight associated with it.
type StringOption struct {
	Weight int    `json:"weight"`
	Value  string `json:"value"`
}

// Valid returns true if this option has a valid weight.
func (o *StringOption) Valid() bool {
	return o.Weight > 0
}

// ChooseStringOption selects a string option from the available set.
func ChooseStringOption(options []*StringOption, not string) string {
	total := 0
	for _, one := range options {
		if one.Value != not {
			total += one.Weight
		}
	}
	if total > 0 {
		choice := 1 + rand.NewCryptoRand().Intn(total)
		for _, one := range options {
			if one.Value != not {
				choice -= one.Weight
				if choice < 1 {
					return one.Value
				}
			}
		}
	}
	return ""
}
