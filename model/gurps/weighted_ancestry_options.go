// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// WeightedAncestryOptions is a string that has a weight associated with it.
type WeightedAncestryOptions struct {
	Weight int              `json:"weight"`
	Value  *AncestryOptions `json:"value"`
}

// Valid returns true if this option has a valid weight.
func (o *WeightedAncestryOptions) Valid() bool {
	return o.Weight > 0
}

// ChooseWeightedAncestryOptions selects a string option from the available set.
func ChooseWeightedAncestryOptions(options []*WeightedAncestryOptions, omitter func(*AncestryOptions) bool) *AncestryOptions {
	total := 0
	for _, one := range options {
		if omitter == nil || !omitter(one.Value) {
			total += one.Weight
		}
	}
	if total > 0 {
		choice := 1 + rand.NewCryptoRand().Intn(total)
		for _, one := range options {
			if omitter == nil || !omitter(one.Value) {
				choice -= one.Weight
				if choice < 1 {
					return one.Value
				}
			}
		}
	}
	return nil
}
