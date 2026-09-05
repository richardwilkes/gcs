// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "github.com/richardwilkes/toolbox/v2/xrand"

// WeightedAncestryOptions is a set of AncestryOptions that has a weight associated with it.
type WeightedAncestryOptions struct {
	Weight int              `json:"weight"`
	Value  *AncestryOptions `json:"value"`
	// KeyPrefix is a runtime-only key the editor uses to give each entry's widgets a stable identity. It is never
	// written to disk.
	KeyPrefix string `json:"-"`
}

// Clone returns a deep copy of this entry, or nil if this entry is nil.
func (o *WeightedAncestryOptions) Clone() *WeightedAncestryOptions {
	if o == nil {
		return nil
	}
	clone := *o
	clone.Value = o.Value.Clone()
	return &clone
}

// Valid returns true if this option has a valid weight and value. A file may omit the value entirely, so this must be
// checked before dereferencing Value.
func (o *WeightedAncestryOptions) Valid() bool {
	return o != nil && o.Weight > 0 && o.Value != nil
}

// ChooseWeightedAncestryOptions selects a string option from the available set.
func ChooseWeightedAncestryOptions(options []*WeightedAncestryOptions, omitter func(*AncestryOptions) bool) *AncestryOptions {
	total := 0
	for _, one := range options {
		if one.Valid() && (omitter == nil || !omitter(one.Value)) {
			total += one.Weight
		}
	}
	if total > 0 {
		choice := 1 + xrand.New().Intn(total)
		for _, one := range options {
			if one.Valid() && (omitter == nil || !omitter(one.Value)) {
				choice -= one.Weight
				if choice < 1 {
					return one.Value
				}
			}
		}
	}
	return nil
}
