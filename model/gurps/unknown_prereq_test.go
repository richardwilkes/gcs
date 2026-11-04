// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestUnknownPrereqPreserved verifies the same round-trip guarantee for prerequisites. Because an unrecognized prereq
// was coerced into a prereq.List, it previously became an empty, always-satisfied list, silently dropping the
// constraint. Unlike an unknown feature, an unknown prereq must report itself as unsatisfied rather than being ignored.
func TestUnknownPrereqPreserved(t *testing.T) {
	c := check.New(t)

	const data = `[
	{
		"type": "attribute_prereq",
		"has": true,
		"which": "iq",
		"qualifier": {
			"compare": "at_least",
			"qualifier": 12
		}
	},
	{
		"type": "future_prereq_from_a_newer_gcs",
		"has": true,
		"details": {
			"whatever": "keep me"
		}
	}
]`
	var prereqs gurps.Prereqs
	c.NoError(jio.Unmarshal([]byte(data), &prereqs), "unknown prereq type must not prevent loading")
	c.Equal(2, len(prereqs), "both prereqs should be present")

	_, ok := prereqs[0].(*gurps.AttributePrereq)
	c.True(ok, "the known prereq should still load normally")

	unknown, ok := prereqs[1].(*gurps.UnknownPrereq)
	c.True(ok, "unrecognized prereq should load as an UnknownPrereq, not be coerced to another type")
	c.Equal("future_prereq_from_a_newer_gcs", unknown.Kind, "the original type should be retained")
	c.Equal(prereq.Unknown, unknown.PrereqType(), "an UnknownPrereq reports the Unknown type")

	// A prerequisite that can't be evaluated must fail closed, and must explain why.
	var tooltip xbytes.InsertBuffer
	c.False(unknown.Satisfied(nil, nil, &tooltip, "", nil), "an unknown prereq is never satisfied")
	c.Contains(tooltip.String(), "future_prereq_from_a_newer_gcs", "the tooltip should name the unknown type")

	out, err := jio.Marshal(prereqs)
	c.NoError(err, "prereqs should marshal")
	c.Equal(jsonArrayElement(c, data, 1), jsonArrayElement(c, string(out), 1),
		"saving must reproduce the unknown prereq exactly")

	clone, ok := unknown.Clone(nil).(*gurps.UnknownPrereq)
	c.True(ok, "a cloned UnknownPrereq is still an UnknownPrereq")
	c.Equal(unknown.Data.String(), clone.Data.String(), "the clone retains the original data")
}
