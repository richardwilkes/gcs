// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package prereq_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestExtractKnownTypeAcceptsRetiredKeys verifies that ExtractKnownType recognizes the keys earlier versions of GCS
// wrote, not just the current ones. Missing them made every pre-rename file's trait prerequisites load as unknown --
// and thus permanently unsatisfied.
func TestExtractKnownTypeAcceptsRetiredKeys(t *testing.T) {
	c := check.New(t)

	// Whatever ExtractType accepts, ExtractKnownType must accept too, or a file that loads under one path breaks
	// under the other.
	for _, one := range prereq.Types {
		value, known := prereq.ExtractKnownType(one.Key())
		c.True(known, "%s should be recognized", one.Key())
		c.Equal(one, value, "%s should resolve to itself", one.Key())
		c.Equal(prereq.ExtractType(one.Key()), value, "%s should agree with ExtractType", one.Key())
	}

	for _, one := range []struct {
		key   string
		want  prereq.Type
		known bool
	}{
		// Case is ignored by ExtractType, so it must be ignored here as well.
		{key: "Advantage_Prereq", want: prereq.Trait, known: true},
		// Neither an unknown key nor the Unknown type's own key may report as known.
		{key: "future_prereq_from_a_newer_gcs", want: prereq.Unknown},
		{key: prereq.Unknown.Key(), want: prereq.Unknown},
	} {
		value, known := prereq.ExtractKnownType(one.key)
		c.Equal(one.known, known, "known for %q", one.key)
		c.Equal(one.want, value, "value for %q", one.key)
	}
}
