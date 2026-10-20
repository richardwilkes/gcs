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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A matcher, and matching through the criteria, must give the same answer as the comparison of the value against the
// qualifier with its replacements applied, for every comparison, whether or not the qualifier carries a nameable marker
// and whether or not a replacement exists for it.
func TestTextMatcherMatchesText(t *testing.T) {
	c := check.New(t)
	comparisons := []criteria.StringComparison{
		criteria.AnyText, criteria.IsText, criteria.IsNotText, criteria.ContainsText, criteria.DoesNotContainText,
		criteria.StartsWithText, criteria.DoesNotStartWithText, criteria.EndsWithText, criteria.DoesNotEndWithText,
	}
	qualifiers := []string{"Broadsword", "@Skill@", "Short@Skill@", "", "  Broadsword "}
	replacements := []map[string]string{nil, {"Skill": "Broadsword"}, {"Skill": ""}, {"Other": "Broadsword"}}
	values := []string{"Broadsword", "broadsword ", "Shortsword", "ShortBroadsword", "", "@Skill@"}
	for _, compare := range comparisons {
		for _, qualifier := range qualifiers {
			text := criteria.Text{Compare: compare, Qualifier: qualifier}
			for _, r := range replacements {
				m := text.Matcher(r)
				for _, value := range values {
					want := compare.Matches(nameable.Apply(qualifier, r), value)
					c.Equal(want, m.Matches(value), "matcher: %v %q with %v against %q", compare, qualifier, r, value)
					c.Equal(want, text.Matches(r, value), "criteria: %v %q with %v against %q", compare, qualifier, r,
						value)
				}
			}
		}
	}
}
