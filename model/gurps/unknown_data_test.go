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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// compactJSON strips insignificant whitespace so that two JSON texts can be compared without regard to formatting.
func compactJSON(c check.Checker, in string) string {
	v := jsontext.Value(in)
	c.NoError(v.Compact(), "should be able to compact %s", in)
	return v.String()
}

// jsonArrayElement extracts a single element from a JSON array as raw text.
func jsonArrayElement(c check.Checker, in string, index int) string {
	var elements []jsontext.Value
	c.NoError(json.Unmarshal([]byte(in), &elements), "should be able to split %s", in)
	c.True(index < len(elements), "element %d should exist in %s", index, in)
	return compactJSON(c, elements[index].String())
}

// TestUnknownFeaturePreserved verifies that a feature whose type this build doesn't recognize doesn't prevent the data
// from loading, isn't silently reinterpreted as some other feature, and is written back out unchanged. The enum's
// UnmarshalText maps anything it doesn't recognize onto the first value, so before this was addressed such a feature
// became an AttributeBonus with an empty attribute and was permanently rewritten as "attribute_bonus" on save.
func TestUnknownFeaturePreserved(t *testing.T) {
	c := check.New(t)

	const data = `[
	{
		"type": "attribute_bonus",
		"attribute": "st",
		"amount": 2
	},
	{
		"type": "future_bonus_from_a_newer_gcs",
		"amount": 3,
		"nested": {
			"list": [1, 2, 3],
			"text": "keep me"
		}
	}
]`
	var features gurps.Features
	c.NoError(json.Unmarshal([]byte(data), &features), "unknown feature type must not prevent loading")
	c.Equal(2, len(features), "both features should be present")

	bonus, ok := features[0].(*gurps.AttributeBonus)
	c.True(ok, "first feature should be an AttributeBonus")
	c.Equal("st", bonus.Attribute, "the known feature should still load normally")

	unknown, ok := features[1].(*gurps.UnknownFeature)
	c.True(ok, "unrecognized feature should load as an UnknownFeature, not be coerced to another type")
	c.Equal("future_bonus_from_a_newer_gcs", unknown.Kind, "the original type should be retained")
	c.Equal(feature.Unknown, unknown.FeatureType(), "an UnknownFeature reports the Unknown type")

	out, err := json.Marshal(features)
	c.NoError(err, "features should marshal")
	c.Equal(jsonArrayElement(c, data, 1), jsonArrayElement(c, string(out), 1),
		"saving must reproduce the unknown feature exactly")

	// A clone must carry the data along, since editors work on clones.
	clone, ok := features.Clone()[1].(*gurps.UnknownFeature)
	c.True(ok, "a cloned UnknownFeature is still an UnknownFeature")
	c.Equal(unknown.Kind, clone.Kind, "the clone retains the original type")
	c.Equal(unknown.Data.String(), clone.Data.String(), "the clone retains the original data")
}

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
	c.NoError(json.Unmarshal([]byte(data), &prereqs), "unknown prereq type must not prevent loading")
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

	out, err := json.Marshal(prereqs)
	c.NoError(err, "prereqs should marshal")
	c.Equal(jsonArrayElement(c, data, 1), jsonArrayElement(c, string(out), 1),
		"saving must reproduce the unknown prereq exactly")

	clone, ok := unknown.Clone(nil).(*gurps.UnknownPrereq)
	c.True(ok, "a cloned UnknownPrereq is still an UnknownPrereq")
	c.Equal(unknown.Data.String(), clone.Data.String(), "the clone retains the original data")
}

// TestUnknownDataSurvivesFileRoundTrip loads a trait file holding both an unknown feature and an unknown prerequisite,
// saves it back out, and confirms the unrecognized data is still intact and still unrecognized -- i.e. that opening a
// file written by a newer version of GCS and saving it doesn't quietly discard what that version added.
func TestUnknownDataSurvivesFileRoundTrip(t *testing.T) {
	c := check.New(t)

	const data = `{
	"version": 5,
	"rows": [
		{
			"name": "Test Trait",
			"base_points": 10,
			"features": [
				{
					"type": "future_bonus_from_a_newer_gcs",
					"amount": 3,
					"nested": {
						"text": "keep me"
					}
				}
			],
			"prereqs": {
				"type": "prereq_list",
				"all": true,
				"prereqs": [
					{
						"type": "future_prereq_from_a_newer_gcs",
						"details": "keep me too"
					}
				]
			}
		}
	]
}`
	fileSystem := fstest.MapFS{"Test.adq": {Data: []byte(data)}}
	traits, err := gurps.NewTraitsFromFile(fileSystem, "Test.adq")
	c.NoError(err, "a file with unknown types must still load")
	c.Equal(1, len(traits), "the trait should be loaded")

	unknownFeature, ok := traits[0].Features[0].(*gurps.UnknownFeature)
	c.True(ok, "the unknown feature should be retained")
	c.Contains(unknownFeature.Data.String(), "keep me", "the unknown feature's data should be retained")

	unknownPrereq, ok := traits[0].Prereq.Prereqs[0].(*gurps.UnknownPrereq)
	c.True(ok, "the unknown prereq should be retained")
	c.Contains(unknownPrereq.Data.String(), "keep me too", "the unknown prereq's data should be retained")

	savePath := filepath.Join(t.TempDir(), "Test.adq")
	c.NoError(gurps.SaveTraits(traits, savePath), "the trait should save")
	saved, err := os.ReadFile(savePath)
	c.NoError(err, "the saved file should be readable")
	c.Contains(string(saved), "future_bonus_from_a_newer_gcs", "the unknown feature type should survive the save")
	c.Contains(string(saved), "future_prereq_from_a_newer_gcs", "the unknown prereq type should survive the save")
	c.NotContains(string(saved), "attribute_bonus", "the unknown feature must not be rewritten as another type")

	// Reload what was written and confirm it is still the same unrecognized data.
	reloaded, err := gurps.NewTraitsFromFile(os.DirFS(filepath.Dir(savePath)), "Test.adq")
	c.NoError(err, "the saved file should load again")
	reloadedFeature, ok := reloaded[0].Features[0].(*gurps.UnknownFeature)
	c.True(ok, "the reloaded feature should still be unknown")
	c.Equal(compactJSON(c, unknownFeature.Data.String()), compactJSON(c, reloadedFeature.Data.String()),
		"the unknown feature's data should be unchanged by the round trip")
	reloadedPrereq, ok := reloaded[0].Prereq.Prereqs[0].(*gurps.UnknownPrereq)
	c.True(ok, "the reloaded prereq should still be unknown")
	c.Equal(compactJSON(c, unknownPrereq.Data.String()), compactJSON(c, reloadedPrereq.Data.String()),
		"the unknown prereq's data should be unchanged by the round trip")
}
