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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNumericComparisonKeysMatchOnDiskForm pins the serialized form of the numeric comparison enum, since data files
// written before it became a generated enum must still load, and files written now must still be read by older
// releases.
func TestNumericComparisonKeysMatchOnDiskForm(t *testing.T) {
	c := check.New(t)
	c.Equal(map[criteria.NumericComparison]string{
		criteria.AnyNumber:       "",
		criteria.EqualsNumber:    "is",
		criteria.NotEqualsNumber: "is_not",
		criteria.AtLeastNumber:   "at_least",
		criteria.AtMostNumber:    "at_most",
	}, keysOf(criteria.NumericComparisons))
}

func keysOf[T interface {
	comparable
	Key() string
}](values []T) map[T]string {
	m := make(map[T]string, len(values))
	for _, v := range values {
		m[v] = v.Key()
	}
	return m
}

// TestNumericComparisonJSONRoundTrip verifies that the numeric comparison is written as its key, and that a key no
// release has ever written loads as "is anything" rather than failing.
func TestNumericComparisonJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	data, err := jio.Marshal(&criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Five})
	c.NoError(err)
	c.Equal(`{"compare":"at_least","qualifier":5}`, string(data))

	var num criteria.Number
	c.NoError(jio.Unmarshal([]byte(`{"compare":"at_most","qualifier":3}`), &num))
	c.Equal(criteria.AtMostNumber, num.Compare)
	c.Equal(fxp.Three, num.Qualifier)
	c.NoError(jio.Unmarshal([]byte(`{"compare":"bogus","qualifier":3}`), &num))
	c.Equal(criteria.AnyNumber, num.Compare)
}

func TestNumericComparisonPluralClause(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		compare criteria.NumericComparison
		want    string
	}{
		{criteria.AnyNumber, "that are anything"},
		{criteria.EqualsNumber, "that are"},
		{criteria.NotEqualsNumber, "that are not"},
		{criteria.AtLeastNumber, "that are at least"},
		{criteria.AtMostNumber, "that are at most"},
		{criteria.NumericComparison(200), "that are anything"},
	} {
		c.Equal(one.want, one.compare.PluralClause(), "plural clause of %d", one.compare)
	}
	c.Equal(5, len(criteria.NumericComparisons), "every comparison is covered above")
}
