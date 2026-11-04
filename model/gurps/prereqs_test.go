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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLegacyTraitPrereqKeyLoads verifies that "advantage_prereq", the key GCS wrote before traits were renamed, still
// resolves to a TraitPrereq rather than an UnknownPrereq, which is never satisfied.
func TestLegacyTraitPrereqKeyLoads(t *testing.T) {
	c := check.New(t)

	const data = `[
	{
		"type": "advantage_prereq",
		"has": true,
		"name": {
			"compare": "is",
			"qualifier": "Magery"
		},
		"level": {
			"compare": "at_least",
			"qualifier": 1
		}
	}
]`
	var prereqs gurps.Prereqs
	c.NoError(jio.Unmarshal([]byte(data), &prereqs), "a legacy prereq key must load")
	c.Equal(1, len(prereqs), "the prereq should be present")

	p, ok := prereqs[0].(*gurps.TraitPrereq)
	if !ok {
		c.True(false, "a legacy advantage_prereq must load as a TraitPrereq, not a %T", prereqs[0])
		return
	}
	c.Equal(prereq.Trait, p.PrereqType(), "the prereq should report itself as a trait prereq")
	c.True(p.Has, "the rest of the prereq should load normally")
	c.Equal("Magery", p.NameCriteria.Qualifier, "the name criteria should load normally")
	c.Equal(criteria.IsText, p.NameCriteria.Compare, "the name comparison should load normally")
	c.Equal(criteria.AtLeastNumber, p.LevelCriteria.Compare, "the level comparison should load normally")

	// Saving migrates the key to the current one rather than preserving the retired spelling.
	out, err := jio.Marshal(prereqs)
	c.NoError(err, "the prereqs should marshal")
	c.Contains(string(out), `"type":"trait_prereq"`, "saving should write the current key")
	c.NotContains(string(out), "advantage_prereq", "saving should not retain the retired key")
}
