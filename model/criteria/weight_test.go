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

// TestWeightDescriptionHasUnits verifies that the description of a weight criteria carries the qualifier's units. A
// weight without them is ambiguous, since the same number means something different in each unit.
func TestWeightDescriptionHasUnits(t *testing.T) {
	c := check.New(t)
	fivePounds := fxp.WeightFromInteger(5, fxp.Pound)
	c.Equal("is at most 5 lb", weight(criteria.AtMostNumber, fivePounds).String())
	c.Equal("is at least 5 lb", weight(criteria.AtLeastNumber, fivePounds).String())
	c.Equal("is 5 lb", weight(criteria.EqualsNumber, fivePounds).String())
	c.Equal("is not 5 lb", weight(criteria.NotEqualsNumber, fivePounds).String())

	// "is anything" has no qualifier to describe, so no units appear.
	c.Equal("is anything", weight(criteria.AnyNumber, fivePounds).String())

	// An invalid comparison is treated as "any", just as it is when matching.
	c.Equal("is anything", weight(criteria.NumericComparison("bogus"), fivePounds).String())
}

// TestWeightDescriptionUsesRequestedUnits verifies that a weight criteria can be described in the units the user has
// chosen, so that the description matches what the editor field for the same value shows.
func TestWeightDescriptionUsesRequestedUnits(t *testing.T) {
	c := check.New(t)
	crit := weight(criteria.AtMostNumber, fxp.WeightFromInteger(5, fxp.Pound))
	c.Equal("is at most 5 lb", crit.Describe(fxp.Pound))
	c.Equal("is at most 2.5 kg", crit.Describe(fxp.Kilogram))
	c.Equal("is at most 80 oz", crit.Describe(fxp.Ounce))

	// The units are those asked for, not those the qualifier was created with.
	metric := weight(criteria.AtLeastNumber, fxp.WeightFromInteger(2, fxp.Kilogram))
	c.Equal("is at least 2 kg", metric.Describe(fxp.Kilogram))
	c.Equal("is at least 4 lb", metric.Describe(fxp.Pound))
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
	c.Equal(none, hashOf(weight(criteria.NumericComparison("bogus"), fivePounds)))

	// A real comparison still contributes both the comparison and its qualifier.
	five := hashOf(weight(criteria.AtMostNumber, fivePounds))
	c.NotEqual(none, five)
	c.NotEqual(five, hashOf(weight(criteria.AtMostNumber, fxp.WeightFromInteger(6, fxp.Pound))))
	c.NotEqual(five, hashOf(weight(criteria.AtLeastNumber, fivePounds)))
}
