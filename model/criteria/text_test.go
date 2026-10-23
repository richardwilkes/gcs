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

// TestTextMatchesListCommaSeparatedQualifier verifies that MatchesList splits the qualifier on commas, trimming each
// entry and ignoring empty ones, and that a positive comparison matches when any value meets any entry while a negative
// one matches only when no value meets any entry.
func TestTextMatchesListCommaSeparatedQualifier(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		compare   criteria.StringComparison
		qualifier string
		values    []string
		want      bool
	}{
		{criteria.IsText, "Sword", []string{"Melee", "Sword"}, true},
		{criteria.IsText, "Sword", []string{"Melee", "Axe"}, false},
		{criteria.ContainsText, "wor", []string{"Sword"}, true},
		{criteria.ContainsText, "wor", []string{"Axe"}, false},
		{criteria.IsText, "Sword, Axe, Polearm", []string{"Melee", "Sword"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", []string{"Melee", "Axe"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", []string{"Polearm"}, true},
		{criteria.IsText, "Sword, Axe, Polearm", []string{"Melee", "Bow"}, false},
		{criteria.IsText, "Sword, Axe, Polearm", []string{"Ranged"}, false},
		// A tag matching several of the listed qualifiers still only reports a single match (the boolean return),
		// which is the core of issue #1008 — one feature, one bonus.
		{criteria.ContainsText, "Sword, Melee", []string{"Melee Sword"}, true},
		{criteria.IsNotText, "Sword, Axe", []string{"Melee", "Bow"}, true},
		{criteria.IsNotText, "Sword, Axe", []string{"Melee", "Sword"}, false},
		{criteria.IsNotText, "Sword, Axe", []string{"Axe"}, false},
		{criteria.DoesNotContainText, "Sword, Axe", []string{"Bow", "Ranged"}, true},
		{criteria.DoesNotContainText, "Sword, Axe", []string{"Broadsword"}, false},
		{criteria.ContainsText, " Sword ,, Axe , ", []string{"Sword"}, true},
		{criteria.ContainsText, " Sword ,, Axe , ", []string{"Axe"}, true},
		{criteria.ContainsText, " Sword ,, Axe , ", []string{"Polearm"}, false},
	} {
		c.Equal(tc.want, text(tc.compare, tc.qualifier).MatchesList(nil, tc.values...), "%v %q against %q", tc.compare,
			tc.qualifier, tc.values)
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

func TestTextIsIgnoresSurroundingSpace(t *testing.T) {
	c := check.New(t)
	blank := text(criteria.IsText, " ")
	c.Equal(" ", blank.Qualifier)
	c.True(blank.Matches(nil, ""))
	c.False(blank.Matches(nil, "Guns"))
	padded := text(criteria.IsText, "Guns ")
	c.Equal("Guns ", padded.Qualifier)
	c.True(padded.Matches(nil, "Guns"))
	c.True(padded.Matches(nil, " guns"))
	plain := text(criteria.IsText, "Pistol")
	c.Equal("Pistol", plain.Qualifier)
	c.True(plain.Matches(nil, "Pistol "), "space around the data is ignored too")
	notBlank := text(criteria.IsNotText, " ")
	c.Equal(" ", notBlank.Qualifier)
	c.False(notBlank.Matches(nil, ""))
	c.True(notBlank.Matches(nil, "Guns"))
	notPadded := text(criteria.IsNotText, "Guns ")
	c.Equal("Guns ", notPadded.Qualifier)
	c.False(notPadded.Matches(nil, "Guns"))
}

func TestTextOtherComparisonsKeepSurroundingSpace(t *testing.T) {
	c := check.New(t)
	startsWith := text(criteria.StartsWithText, "Seek ")
	c.Equal("Seek ", startsWith.Qualifier)
	c.True(startsWith.Matches(nil, "Seek Water"))
	c.False(startsWith.Matches(nil, "Seeker"))
	endsWith := text(criteria.EndsWithText, " Dull")
	c.Equal(" Dull", endsWith.Qualifier)
	c.True(endsWith.Matches(nil, "Very Dull"))
	c.False(endsWith.Matches(nil, "Dull"))
	contains := text(criteria.ContainsText, " ")
	c.Equal(" ", contains.Qualifier)
	c.True(contains.Matches(nil, "Fast Draw"))
	c.False(contains.Matches(nil, "Guns"))
	notContains := text(criteria.DoesNotContainText, " ")
	c.Equal(" ", notContains.Qualifier)
	c.True(notContains.Matches(nil, "Guns"))
}

func TestTextDescribeTrimsOnlyIs(t *testing.T) {
	c := check.New(t)
	padded := text(criteria.IsText, " Guns ")
	c.Equal(" Guns ", padded.Qualifier)
	c.Equal(`is "Guns"`, padded.String(nil))
	notPadded := text(criteria.IsNotText, " ")
	c.Equal(" ", notPadded.Qualifier)
	c.Equal(`is not ""`, notPadded.String(nil))
	startsWith := text(criteria.StartsWithText, "Seek ")
	c.Equal("Seek ", startsWith.Qualifier)
	c.Equal(`starts with "Seek "`, startsWith.String(nil))
}
