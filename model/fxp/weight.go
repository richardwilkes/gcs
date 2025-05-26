// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"strings"

	"github.com/richardwilkes/json"
	"golang.org/x/exp/constraints"
)

// Weight contains a fixed-point value in pounds.
type Weight Int

// WeightFromInteger creates a new Weight.
func WeightFromInteger[T constraints.Integer](value T, unit WeightUnit) Weight {
	return Weight(unit.ToPounds(From(value)))
}

// WeightFromFixed creates a new Weight.
func WeightFromFixed(value Int, unit WeightUnit) Weight {
	return Weight(unit.ToPounds(value))
}

// WeightFromStringForced creates a new Weight. May have any of the known Weight suffixes or no notation at all, in
// which case defaultUnits is used.
func WeightFromStringForced(text string, defaultUnits WeightUnit) Weight {
	weight, err := WeightFromString(text, defaultUnits)
	if err != nil {
		return 0
	}
	return weight
}

// WeightFromString creates a new Weight. May have any of the known Weight suffixes or no notation at all, in which case
// defaultUnits is used.
func WeightFromString(text string, defaultUnits WeightUnit) (Weight, error) {
	text = strings.TrimLeft(strings.TrimSpace(text), "+")
	for _, unit := range WeightUnits {
		if strings.HasSuffix(text, unit.Key()) {
			value, err := FromString(strings.TrimSpace(strings.TrimSuffix(text, unit.Key())))
			if err != nil {
				return 0, err
			}
			return Weight(unit.ToPounds(value)), nil
		}
	}
	// No matches, so let's use our passed-in default units
	value, err := FromString(strings.TrimSpace(text))
	if err != nil {
		return 0, err
	}
	return Weight(defaultUnits.ToPounds(value)), nil
}

func (w Weight) String() string {
	return Pound.Format(w)
}

// MarshalJSON implements json.Marshaler.
func (w Weight) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weight) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	var err error
	*w, err = WeightFromString(s, Pound)
	return err
}

// WeightLessFromStringFunc returns a func to compare two strings as Weights.
func WeightLessFromStringFunc(units WeightUnit) func(a, b string) bool {
	return func(a, b string) bool { return WeightFromStringForced(a, units) < WeightFromStringForced(b, units) }
}
