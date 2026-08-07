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
	"encoding/json/v2"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLegacyTraitPrereqKeyLoads verifies that "advantage_prereq", the key GCS wrote before traits were renamed, still
// resolves to a TraitPrereq. Dispatching on the current keys alone turned every such prerequisite into an
// UnknownPrereq, which is never satisfied, so all pre-rename files loaded with their trait prerequisites broken.
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
	c.NoError(json.Unmarshal([]byte(data), &prereqs), "a legacy prereq key must load")
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
	out, err := json.Marshal(prereqs)
	c.NoError(err, "the prereqs should marshal")
	c.Contains(string(out), `"type":"trait_prereq"`, "saving should write the current key")
	c.NotContains(string(out), "advantage_prereq", "saving should not retain the retired key")
}

// TestLegacyTraitPrereqIsEvaluated goes past loading and confirms the migrated prerequisite actually participates in
// prereq evaluation: an UnknownPrereq always fails, so a trait that meets its own requirement would still have been
// flagged as unsatisfied.
func TestLegacyTraitPrereqIsEvaluated(t *testing.T) {
	c := check.New(t)

	const data = `{
	"version": 5,
	"rows": [
		{
			"id": "00000000-0000-0000-0000-000000000001",
			"name": "Magery",
			"base_points": 5
		},
		{
			"id": "00000000-0000-0000-0000-000000000002",
			"name": "Needs Magery",
			"base_points": 10,
			"prereqs": {
				"type": "prereq_list",
				"all": true,
				"prereqs": [
					{
						"type": "advantage_prereq",
						"has": true,
						"name": {
							"compare": "is",
							"qualifier": "Magery"
						}
					}
				]
			}
		}
	]
}`
	traits, err := gurps.NewTraitsFromFile(fstest.MapFS{"Test.adq": {Data: []byte(data)}}, "Test.adq")
	c.NoError(err, "the file should load")
	c.Equal(2, len(traits), "both traits should load")

	entity := gurps.NewEntity()
	entity.Traits = traits
	entity.Recalculate()

	c.Equal("", traits[1].UnsatisfiedReason,
		"a satisfied legacy trait prereq must not be reported as unsatisfied: %s", traits[1].UnsatisfiedReason)

	// The same prereq must still fail when the required trait isn't present, proving it is genuinely being evaluated
	// rather than being ignored.
	entity.Traits = traits[1:]
	entity.Recalculate()
	c.NotEqual("", traits[1].UnsatisfiedReason, "the legacy trait prereq should fail when the trait is absent")
}
