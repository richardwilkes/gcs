// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestParseRulesLookupData(t *testing.T) {
	c := check.New(t)

	t.Run("groups by book", func(_ *testing.T) {
		rules, err := parseRulesLookupData([]byte(`[
			{"Rule":"Dodge","Category":["Combat"],"Book":"B","Page":"374","Link":"B374"},
			{"Rule":"Parry","Category":["Combat"],"Book":"B","Page":"376","Link":"B376"},
			{"Rule":"Bless","Category":["Spells"],"Book":"M","Page":"37","Link":"M37"}
		]`))
		c.NoError(err)
		c.Equal(2, len(rules))
		c.Equal(2, len(rules["B"]))
		c.Equal(1, len(rules["M"]))
		c.Equal("Dodge", rules["B"][0].Rule)
		c.Equal("B374", rules["B"][0].Link)
		c.Equal([]string{"Spells"}, rules["M"][0].Category)
	})

	t.Run("strips a byte order mark", func(_ *testing.T) {
		rules, err := parseRulesLookupData(append([]byte("\xef\xbb\xbf"),
			[]byte(`[{"Rule":"Dodge","Book":"B","Page":"374","Link":"B374"}]`)...))
		c.NoError(err)
		c.Equal(1, len(rules["B"]))
	})

	t.Run("reports malformed data", func(_ *testing.T) {
		rules, err := parseRulesLookupData([]byte("not json"))
		c.HasError(err)
		c.Equal(0, len(rules), "a parse failure must not yield any rules")
	})
}
