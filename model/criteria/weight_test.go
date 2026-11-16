// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

func weight(compare criteria.NumericComparison, qualifier fxp.Weight) criteria.Weight {
	var w criteria.Weight
	w.Compare = compare
	w.Qualifier = qualifier
	return w
}

// TestWeightDescribe verifies that the description of a weight criteria carries the qualifier in the units asked for,
// since the same number means something different in each unit, and that String and AltString describe it in pounds.
func TestWeightDescribe(t *testing.T) {
	c := check.New(t)
	fivePounds := fxp.WeightFromInteger(5, fxp.Pound)
	for _, tc := range []struct {
		compare       criteria.NumericComparison
		units         fxp.WeightUnit
		describe, alt string
	}{
		{criteria.AtMostNumber, fxp.Pound, "is at most 5 lb", "at most 5 lb"},
		{criteria.AtLeastNumber, fxp.Pound, "is at least 5 lb", "at least 5 lb"},
		{criteria.EqualsNumber, fxp.Pound, "is 5 lb", "5 lb"},
		{criteria.NotEqualsNumber, fxp.Pound, "is not 5 lb", "not 5 lb"},
		{criteria.AtMostNumber, fxp.Kilogram, "is at most 2.5 kg", "at most 2.5 kg"},
		// "is anything" has no qualifier to describe, so no units appear.
		{criteria.AnyNumber, fxp.Pound, "is anything", "anything"},
		// An invalid comparison is treated as "any", just as it is when matching.
		{criteria.LastNumericComparison + 1, fxp.Pound, "is anything", "anything"},
	} {
		crit := weight(tc.compare, fivePounds)
		c.Equal(tc.describe, crit.Describe(tc.units), "%d in %v", tc.compare, tc.units)
		c.Equal(tc.alt, crit.AltDescribe(tc.units), "alt of %d in %v", tc.compare, tc.units)
		if tc.units == fxp.Pound {
			c.Equal(tc.describe, crit.String(), "string of %d", tc.compare)
			c.Equal(tc.alt, crit.AltString(), "alt string of %d", tc.compare)
		}
	}
}

// TestWeightHashIgnoresQualifierWhenAny verifies that a weight criteria whose comparison is "is anything" hashes the
// same no matter what stale qualifier it carries, since marshaling omits the criteria entirely in that state.
func TestWeightHashIgnoresQualifierWhenAny(t *testing.T) {
	c := check.New(t)
	fivePounds := fxp.WeightFromInteger(5, fxp.Pound)
	none := hashOf(weight(criteria.AnyNumber, 0))
	c.Equal(none, hashOf(weight(criteria.AnyNumber, fivePounds)))
	c.Equal(none, hashOf(criteria.Weight{}))

	// An invalid comparison is treated as "any" everywhere else, so it must hash as "any", too.
	c.Equal(none, hashOf(weight(criteria.LastNumericComparison+1, fivePounds)))

	// A real comparison still contributes both the comparison and its qualifier.
	five := hashOf(weight(criteria.AtMostNumber, fivePounds))
	c.NotEqual(none, five)
	c.NotEqual(five, hashOf(weight(criteria.AtMostNumber, fxp.WeightFromInteger(6, fxp.Pound))))
	c.NotEqual(five, hashOf(weight(criteria.AtLeastNumber, fivePounds)))
}
