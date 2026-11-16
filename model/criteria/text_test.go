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
	"github.com/richardwilkes/toolbox/v2/check"
)

func text(compare criteria.StringComparison, qualifier string) criteria.Text {
	var t criteria.Text
	t.Compare = compare
	t.Qualifier = qualifier
	return t
}

// TestTextMatches verifies that "is" and "is not" ignore space at either end of the qualifier and the value while the
// other comparisons keep it, and that the qualifier's replacements are applied before it is compared.
func TestTextMatches(t *testing.T) {
	c := check.New(t)
	skill := map[string]string{"Skill": "Broadsword"}
	for _, tc := range []struct {
		compare      criteria.StringComparison
		qualifier    string
		replacements map[string]string
		value        string
		want         bool
	}{
		{criteria.IsText, " ", nil, "", true},
		{criteria.IsText, " ", nil, "Guns", false},
		{criteria.IsText, "Guns ", nil, "Guns", true},
		{criteria.IsText, "Guns ", nil, " guns", true},
		{criteria.IsText, "Pistol", nil, "Pistol ", true},
		{criteria.IsNotText, " ", nil, "", false},
		{criteria.IsNotText, " ", nil, "Guns", true},
		{criteria.IsNotText, "Guns ", nil, "Guns", false},
		{criteria.StartsWithText, "Seek ", nil, "Seek Water", true},
		{criteria.StartsWithText, "Seek ", nil, "Seeker", false},
		{criteria.EndsWithText, " Dull", nil, "Very Dull", true},
		{criteria.EndsWithText, " Dull", nil, "Dull", false},
		{criteria.ContainsText, " ", nil, "Fast Draw", true},
		{criteria.ContainsText, " ", nil, "Guns", false},
		{criteria.DoesNotContainText, " ", nil, "Guns", true},
		{criteria.IsText, "@Skill@", skill, "Broadsword", true},
	} {
		crit := text(tc.compare, tc.qualifier)
		c.Equal(tc.want, crit.Matches(tc.replacements, tc.value), "criteria: %v %q with %v against %q", tc.compare,
			tc.qualifier, tc.replacements, tc.value)
		c.Equal(tc.want, crit.Matcher(tc.replacements).Matches(tc.value), "matcher: %v %q with %v against %q",
			tc.compare, tc.qualifier, tc.replacements, tc.value)
	}
}

// TestTextMatchesList verifies that MatchesList applies the qualifier's replacements and splits it on commas, trimming
// each entry and ignoring empty ones, that a positive comparison matches when any value meets any entry while a
// negative one matches only when no value meets any entry, and that no values are judged as one empty value.
func TestTextMatchesList(t *testing.T) {
	c := check.New(t)
	skill := map[string]string{"Skill": "Sword"}
	for _, tc := range []struct {
		compare      criteria.StringComparison
		qualifier    string
		replacements map[string]string
		values       []string
		want         bool
	}{
		{criteria.IsText, "Sword", nil, []string{"Melee", "Sword"}, true},
		{criteria.IsText, "Sword", nil, []string{"Melee", "Axe"}, false},
		{criteria.ContainsText, "wor", nil, []string{"Sword"}, true},
		{criteria.ContainsText, "wor", nil, []string{"Axe"}, false},
		{criteria.IsText, "Sword, Axe, Polearm", nil, []string{"Melee", "Sword"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", nil, []string{"Melee", "Axe"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", nil, []string{"Polearm"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", nil, []string{"Melee", "Bow"}, false},
		// A tag matching several of the listed qualifiers still only reports a single match (the boolean return),
		// which is the core of issue #1008 — one feature, one bonus.
		{criteria.ContainsText, "Sword, Melee", nil, []string{"Melee Sword"}, true},
		{criteria.IsNotText, "Sword, Axe", nil, []string{"Melee", "Bow"}, true},
		{criteria.IsNotText, "Sword, Axe", nil, []string{"Melee", "Sword"}, false},
		{criteria.IsNotText, "Sword, Axe", nil, []string{"Axe"}, false},
		{criteria.DoesNotContainText, "Sword, Axe", nil, []string{"Bow", "Ranged"}, true},
		{criteria.DoesNotContainText, "Sword, Axe", nil, []string{"Broadsword"}, false},
		{criteria.DoesNotStartWithText, "Sword, Axe", nil, []string{"Axehead"}, false},
		{criteria.DoesNotEndWithText, "Sword, Axe", nil, []string{"Broadsword"}, false},
		{criteria.ContainsText, " Sword ,, Axe , ", nil, []string{"Sword"}, true},
		{criteria.ContainsText, " Sword ,, Axe , ", nil, []string{"Axe"}, true},
		{criteria.ContainsText, " Sword ,, Axe , ", nil, []string{"Polearm"}, false},
		// A qualifier with no entries is judged as one empty entry.
		{criteria.IsText, " , ", nil, []string{""}, true},
		{criteria.AnyText, "", nil, []string{"Sword"}, true},
		// No values are judged as one empty value.
		{criteria.IsText, "", nil, nil, true},
		{criteria.IsText, "Sword", nil, nil, false},
		{criteria.IsNotText, "Sword", nil, nil, true},
		{criteria.IsText, "@Skill@, Axe", skill, []string{"Sword"}, true},
	} {
		crit := text(tc.compare, tc.qualifier)
		c.Equal(tc.want, crit.MatchesList(tc.replacements, tc.values...), "criteria: %v %q with %v against %q",
			tc.compare, tc.qualifier, tc.replacements, tc.values)
		c.Equal(tc.want, crit.Matcher(tc.replacements).MatchesList(tc.values...), "matcher: %v %q with %v against %q",
			tc.compare, tc.qualifier, tc.replacements, tc.values)
	}
}

func TestTextMatchesEachKeepsCommas(t *testing.T) {
	c := check.New(t)
	// The qualifier is taken whole, so a comma is part of what is matched rather than a separator.
	c.True(text(criteria.IsText, "Melee, Ranged").MatchesEach(nil, "Thrown", "Melee, Ranged"))
	c.False(text(criteria.IsText, "Melee, Ranged").MatchesEach(nil, "Melee", "Ranged"))
	c.True(text(criteria.ContainsText, "e, R").MatchesEach(nil, "Melee, Ranged"))
	c.False(text(criteria.ContainsText, "Melee, Thrown").MatchesEach(nil, "Melee, Ranged"))
	// A negative comparison holds only when every value meets it.
	c.True(text(criteria.DoesNotContainText, "Melee, Ranged").MatchesEach(nil, "Melee", "Ranged"))
	c.False(text(criteria.DoesNotContainText, "Melee, Ranged").MatchesEach(nil, "Melee", "Melee, Ranged"))
}

func TestTextDescribeTrimsOnlyIs(t *testing.T) {
	c := check.New(t)
	c.Equal(`is "Guns"`, text(criteria.IsText, " Guns ").String(nil))
	c.Equal(`is not ""`, text(criteria.IsNotText, " ").String(nil))
	c.Equal(`starts with "Seek "`, text(criteria.StartsWithText, "Seek ").String(nil))
}
