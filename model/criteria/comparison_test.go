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

// TestComparisonKeysMatchOnDiskForm pins the serialized form of the comparison enums, since data files written before
// they became generated enums must still load, and files written now must still be read by older releases.
func TestComparisonKeysMatchOnDiskForm(t *testing.T) {
	c := check.New(t)
	c.Equal(map[criteria.NumericComparison]string{
		criteria.AnyNumber:       "",
		criteria.EqualsNumber:    "is",
		criteria.NotEqualsNumber: "is_not",
		criteria.AtLeastNumber:   "at_least",
		criteria.AtMostNumber:    "at_most",
	}, keysOf(criteria.NumericComparisons))
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

// TestComparisonJSONRoundTrip verifies that the comparison is written as its key, that "is anything" is omitted
// entirely, and that a key no release has ever written loads as "is anything" rather than failing.
func TestComparisonJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	data, err := jio.Marshal(&criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Five})
	c.NoError(err)
	c.Equal(`{"compare":"at_least","qualifier":5}`, string(data))
	data, err = jio.Marshal(&criteria.Text{Compare: criteria.AnyText, Qualifier: "x"})
	c.NoError(err)
	c.Equal(`{"qualifier":"x"}`, string(data))

	var num criteria.Number
	c.NoError(jio.Unmarshal([]byte(`{"compare":"at_most","qualifier":3}`), &num))
	c.Equal(criteria.AtMostNumber, num.Compare)
	c.Equal(fxp.Three, num.Qualifier)
	c.NoError(jio.Unmarshal([]byte(`{"compare":"bogus","qualifier":3}`), &num))
	c.Equal(criteria.AnyNumber, num.Compare)

	var text criteria.Text
	c.NoError(jio.Unmarshal([]byte(`{"compare":"does_not_end_with","qualifier":"y"}`), &text))
	c.Equal(criteria.DoesNotEndWithText, text.Compare)
	c.NoError(jio.Unmarshal([]byte(`{"compare":"bogus","qualifier":"y"}`), &text))
	c.Equal(criteria.AnyText, text.Compare)
}

// TestStringComparisonAltStringOnlyDiffersForNotTypes verifies the alternate (plural) string is only distinct for the
// "not" comparisons, which is what DescribeWithPrefix and PrefixedStringComparisonChoices rely on.
func TestStringComparisonAltStringOnlyDiffersForNotTypes(t *testing.T) {
	c := check.New(t)
	for _, one := range criteria.StringComparisons {
		if one.IsNotType() {
			c.NotEqual(one.String(), one.AltString(), one.Key())
		} else {
			c.Equal(one.String(), one.AltString(), one.Key())
		}
	}
}

func TestStringComparisonClauses(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		compare      criteria.StringComparison
		plural, list string
		positive     criteria.StringComparison
	}{
		{criteria.AnyText, "that are anything", "that are anything", criteria.AnyText},
		{criteria.IsText, "that are", "where at least one is", criteria.IsText},
		{criteria.IsNotText, "that are not", "where none is", criteria.IsText},
		{criteria.ContainsText, "that contain", "where at least one contains", criteria.ContainsText},
		{criteria.DoesNotContainText, "that do not contain", "where none contains", criteria.ContainsText},
		{criteria.StartsWithText, "that start with", "where at least one starts with", criteria.StartsWithText},
		{criteria.DoesNotStartWithText, "that do not start with", "where none starts with", criteria.StartsWithText},
		{criteria.EndsWithText, "that end with", "where at least one ends with", criteria.EndsWithText},
		{criteria.DoesNotEndWithText, "that do not end with", "where none ends with", criteria.EndsWithText},
		{criteria.StringComparison(200), "that are anything", "that are anything", criteria.AnyText},
	} {
		c.Equal(one.plural, one.compare.PluralClause(), "plural clause of %d", one.compare)
		c.Equal(one.list, one.compare.ListClause(), "list clause of %d", one.compare)
		c.Equal(one.positive, one.compare.Positive(), "positive of %d", one.compare)
	}
	c.Equal(9, len(criteria.StringComparisons), "every comparison is covered above")
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
