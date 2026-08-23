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
	"hash"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/zeebo/xxh3"
)

func number(compare criteria.NumericComparison, qualifier fxp.Int) criteria.Number {
	var n criteria.Number
	n.Compare = compare
	n.Qualifier = qualifier
	return n
}

func hashOf(in interface{ Hash(h hash.Hash) }) uint64 {
	h := xxh3.New()
	in.Hash(h)
	return h.Sum64()
}

// TestNumberHashIgnoresQualifierWhenAny verifies that a number criteria whose comparison is "is anything" hashes the
// same no matter what stale qualifier it carries. Marshaling omits the criteria entirely when IsZero(), so two states
// that write identical JSON must not produce differing hashes -- otherwise the source-matching code reports an item as
// modified from its library source when its saved form is identical.
func TestNumberHashIgnoresQualifierWhenAny(t *testing.T) {
	c := check.New(t)
	none := hashOf(number(criteria.AnyNumber, 0))
	c.Equal(none, hashOf(number(criteria.AnyNumber, fxp.FromInteger(10))))
	c.Equal(none, hashOf(criteria.Number{}))

	// An invalid comparison is treated as "any" everywhere else, so it must hash as "any", too.
	c.Equal(none, hashOf(number(criteria.NumericComparison("bogus"), fxp.FromInteger(10))))

	// A real comparison still contributes both the comparison and its qualifier.
	ten := hashOf(number(criteria.AtLeastNumber, fxp.FromInteger(10)))
	c.NotEqual(none, ten)
	c.NotEqual(ten, hashOf(number(criteria.AtLeastNumber, fxp.FromInteger(11))))
	c.NotEqual(ten, hashOf(number(criteria.AtMostNumber, fxp.FromInteger(10))))
}

// TestNumberEqualJSONHashesEqually pins the invariant the hash guard exists for: whenever two criteria serialize
// identically, they hash identically.
func TestNumberEqualJSONHashesEqually(t *testing.T) {
	c := check.New(t)
	type holder struct {
		Criteria criteria.Number `json:"criteria,omitzero"`
	}
	stale, err := jio.Marshal(&holder{Criteria: number(criteria.AnyNumber, fxp.FromInteger(10))})
	c.NoError(err)
	clean, err := jio.Marshal(&holder{Criteria: number(criteria.AnyNumber, 0)})
	c.NoError(err)
	c.Equal(string(clean), string(stale))
	c.Equal(hashOf(number(criteria.AnyNumber, 0)), hashOf(number(criteria.AnyNumber, fxp.FromInteger(10))))
}
