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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestStringComparisonKeysMatchOnDiskForm pins the serialized form of the string comparison enum, since data files
// written before it became a generated enum must still load, and files written now must still be read by older
// releases.
func TestStringComparisonKeysMatchOnDiskForm(t *testing.T) {
	c := check.New(t)
	c.Equal(map[criteria.StringComparison]string{
		criteria.AnyText:              "",
		criteria.IsText:               "is",
		criteria.IsNotText:            "is_not",
		criteria.ContainsText:         "contains",
		criteria.DoesNotContainText:   "does_not_contain",
		criteria.StartsWithText:       "starts_with",
		criteria.DoesNotStartWithText: "does_not_start_with",
		criteria.EndsWithText:         "ends_with",
		criteria.DoesNotEndWithText:   "does_not_end_with",
	}, keysOf(criteria.StringComparisons))
}

// TestStringComparisonJSONRoundTrip verifies that "is anything" is omitted entirely when the string comparison is
// written, that the comparison is read from its key, and that a key no release has ever written loads as "is anything"
// rather than failing.
func TestStringComparisonJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	data, err := jio.Marshal(&criteria.Text{Compare: criteria.AnyText, Qualifier: "x"})
	c.NoError(err)
	c.Equal(`{"qualifier":"x"}`, string(data))

	var text criteria.Text
	c.NoError(jio.Unmarshal([]byte(`{"compare":"does_not_end_with","qualifier":"y"}`), &text))
	c.Equal(criteria.DoesNotEndWithText, text.Compare)
	c.NoError(jio.Unmarshal([]byte(`{"compare":"bogus","qualifier":"y"}`), &text))
	c.Equal(criteria.AnyText, text.Compare)
}

func TestStringComparisonClauses(t *testing.T) {
	c := check.New(t)
	type clauses struct {
		plural, list string
		positive     criteria.StringComparison
	}
	invalid := criteria.StringComparison(200)
	want := map[criteria.StringComparison]clauses{
		criteria.AnyText:              {"that are anything", "that are anything", criteria.AnyText},
		criteria.IsText:               {"that are", "where at least one is", criteria.IsText},
		criteria.IsNotText:            {"that are not", "where none is", criteria.IsText},
		criteria.ContainsText:         {"that contain", "where at least one contains", criteria.ContainsText},
		criteria.DoesNotContainText:   {"that do not contain", "where none contains", criteria.ContainsText},
		criteria.StartsWithText:       {"that start with", "where at least one starts with", criteria.StartsWithText},
		criteria.DoesNotStartWithText: {"that do not start with", "where none starts with", criteria.StartsWithText},
		criteria.EndsWithText:         {"that end with", "where at least one ends with", criteria.EndsWithText},
		criteria.DoesNotEndWithText:   {"that do not end with", "where none ends with", criteria.EndsWithText},
		invalid:                       {"that are anything", "that are anything", criteria.AnyText},
	}
	// A comparison missing from the table has empty clauses to match, so it fails here.
	for _, compare := range slices.Concat(criteria.StringComparisons, []criteria.StringComparison{invalid}) {
		one := want[compare]
		c.Equal(one.plural, compare.PluralClause(), "plural clause of %d", compare)
		c.Equal(one.list, compare.ListClause(), "list clause of %d", compare)
		c.Equal(one.positive, compare.Positive(), "positive of %d", compare)
	}
}
