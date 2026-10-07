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

func TestTextMatchesListSingleQualifier(t *testing.T) {
	c := check.New(t)
	c.True(text(criteria.IsText, "Sword").MatchesList(nil, "Melee", "Sword"))
	c.False(text(criteria.IsText, "Sword").MatchesList(nil, "Melee", "Axe"))
	c.True(text(criteria.ContainsText, "wor").MatchesList(nil, "Sword"))
	c.False(text(criteria.ContainsText, "wor").MatchesList(nil, "Axe"))
}

func TestTextMatchesListCommaSeparatedQualifier(t *testing.T) {
	c := check.New(t)
	// "is" against a comma-separated list matches when any one tag matches any one qualifier.
	crit := text(criteria.IsText, "Sword, Axe, Polearm")
	c.True(crit.MatchesList(nil, "Melee", "Sword"))
	c.True(crit.MatchesList(nil, "Melee", "Axe"))
	c.True(crit.MatchesList(nil, "Polearm"))
	c.False(crit.MatchesList(nil, "Melee", "Bow"))
	c.False(crit.MatchesList(nil, "Ranged"))

	// A tag matching several of the listed qualifiers still only reports a single match (the boolean return),
	// which is the core of issue #1008 — one feature, one bonus.
	c.True(text(criteria.ContainsText, "Sword, Melee").MatchesList(nil, "Melee Sword"))
}

func TestTextMatchesListCommaSeparatedNegated(t *testing.T) {
	c := check.New(t)
	// "is not" against a comma-separated list matches only when no tag matches any qualifier.
	crit := text(criteria.IsNotText, "Sword, Axe")
	c.True(crit.MatchesList(nil, "Melee", "Bow"))
	c.False(crit.MatchesList(nil, "Melee", "Sword"))
	c.False(crit.MatchesList(nil, "Axe"))

	notContains := text(criteria.DoesNotContainText, "Sword, Axe")
	c.True(notContains.MatchesList(nil, "Bow", "Ranged"))
	c.False(notContains.MatchesList(nil, "Broadsword"))
}

func TestTextMatchesListWhitespaceAndEmptyEntries(t *testing.T) {
	c := check.New(t)
	// Surrounding whitespace is trimmed and empty entries (from stray commas) are ignored.
	crit := text(criteria.IsText, " Sword ,, Axe , ")
	c.True(crit.MatchesList(nil, "Sword"))
	c.True(crit.MatchesList(nil, "Axe"))
	c.False(crit.MatchesList(nil, "Polearm"))
}

func TestTextMatchesListEmptyValues(t *testing.T) {
	c := check.New(t)
	// "any" matches regardless of the values.
	c.True(text(criteria.AnyText, "").MatchesList(nil))
	c.True(text(criteria.AnyText, "").MatchesList(nil, "Sword"))
	// With no values present, a positive "is" cannot match, but a negative "is not" does.
	c.False(text(criteria.IsText, "Sword").MatchesList(nil))
	c.True(text(criteria.IsNotText, "Sword").MatchesList(nil))
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
