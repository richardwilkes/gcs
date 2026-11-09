// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// describeFeatures returns the plain description of each feature.
func describeFeatures(list Features) []string {
	result := make([]string, 0, len(list))
	for _, f := range list {
		result = append(result, f.Describe(nil, nil, func(s string) string { return s }))
	}
	return result
}

// TestPoolThresholdLegacyOpsBecomeFeatures verifies that the operations a threshold held before it held features are
// converted to the features that do the same, in a fixed order, matching keys ignoring case and dropping unknown ones,
// that features given outright are kept instead, and that the threshold is written back with the features along with
// the operations they are the conversion of in any order, after the features, for older versions to read, and with no
// operations when they aren't.
func TestPoolThresholdLegacyOpsBecomeFeatures(t *testing.T) {
	c := check.New(t)
	const move = "×0.5 to Move"
	const dodge = "×0.5 to Dodge"
	const st = "×0.5 to ST for lifting only"
	for _, one := range []struct {
		name string
		data string
		want []string
		ops  string
	}{
		{"halve move", `{"state":"Reeling","value":"0","ops":["halve_move"]}`, []string{move}, `["halve_move"]`},
		{"halve dodge", `{"state":"Reeling","value":"0","ops":["halve_dodge"]}`, []string{dodge}, `["halve_dodge"]`},
		{"halve st", `{"state":"Tired","value":"0","ops":["halve_st"]}`, []string{st}, `["halve_st"]`},
		{
			"all, in a fixed order whatever the data's",
			`{"state":"Tired","value":"0","ops":["halve_st","HALVE_MOVE","halve_dodge"]}`,
			[]string{move, dodge, st},
			`["halve_move","halve_dodge","halve_st"]`,
		},
		{"unknown dropped", `{"state":"Tired","value":"0","ops":["halve_all","halve_move"]}`, []string{move}, `["halve_move"]`},
		{
			"features win",
			`{"state":"Tired","value":"0","ops":["halve_move"],"features":[{"type":"attribute_bonus","attribute":"dx","amount":-1}]}`,
			[]string{"-1 to DX"},
			"",
		},
		{
			"features that are the conversion of operations",
			`{"state":"Tired","value":"0","features":[{"type":"attribute_bonus","attribute":"dodge","multiply":true,"amount":0.5}]}`,
			[]string{dodge},
			`["halve_dodge"]`,
		},
		{
			"features that are the conversion of operations, in another order",
			`{"state":"Tired","value":"0","features":[{"type":"attribute_bonus","attribute":"dodge","multiply":true,"amount":0.5},{"type":"attribute_bonus","attribute":"move","multiply":true,"amount":0.5}]}`,
			[]string{dodge, move},
			`["halve_move","halve_dodge"]`,
		},
		{
			"features that are more than that",
			`{"state":"Tired","value":"0","features":[{"type":"attribute_bonus","attribute":"dodge","multiply":true,"amount":0.5},{"type":"attribute_bonus","attribute":"dx","amount":-1}]}`,
			[]string{dodge, "-1 to DX"},
			"",
		},
		{"none", `{"state":"Rested","value":"$fp"}`, nil, ""},
	} {
		var threshold PoolThreshold
		c.NoError(jio.Unmarshal([]byte(one.data), &threshold), one.name)
		got := describeFeatures(threshold.Features)
		if len(one.want) == 0 {
			c.Equal(0, len(got), one.name)
		} else {
			c.Equal(one.want, got, one.name)
		}
		data, err := jio.Marshal(&threshold)
		c.NoError(err, one.name)
		c.Equal(len(one.want) != 0, strings.Contains(string(data), `"features"`), "%s: features are written back",
			one.name)
		if one.ops == "" {
			c.False(strings.Contains(string(data), `"ops"`), "%s: no ops are written back", one.name)
		} else {
			c.True(strings.Contains(string(data), `"ops":`+one.ops), "%s: the ops are written back: %s", one.name, data)
			c.True(bytes.Index(data, []byte(`"features"`)) < bytes.Index(data, []byte(`"ops"`)),
				"%s: the ops follow the features", one.name)
		}
	}
}

// TestPoolThresholdFactoryDataWritesLegacyOps verifies that the factory thresholds, whose features are the conversion
// of the operations older versions of GCS read, are written with those operations, so that a sheet saved by this
// version loads in an older one with what its thresholds do intact.
func TestPoolThresholdFactoryDataWritesLegacyOps(t *testing.T) {
	c := check.New(t)
	tired := poolThreshold(NewEntity(), FatiguePointsID, "Tired")
	c.NotNil(tired)
	data, err := jio.Marshal(tired)
	c.NoError(err)
	c.True(strings.Contains(string(data), `"ops":["halve_move","halve_dodge","halve_st"]`), "%s", data)
	var old struct {
		Ops []string `json:"ops"`
	}
	c.NoError(jio.Unmarshal(data, &old))
	c.Equal([]string{"halve_move", "halve_dodge", "halve_st"}, old.Ops, "an older version reads the operations")
	bonus, ok := tired.Features[0].(*AttributeBonus)
	c.True(ok)
	bonus.Amount = fxp.Quarter
	data, err = jio.Marshal(tired)
	c.NoError(err)
	c.False(strings.Contains(string(data), `"ops"`), "changed features write no operations: %s", data)
}

// TestPoolThresholdLegacyOpsMatchFactoryThresholds verifies that the factory thresholds and the conversion of the
// operations they used to hold agree, so that a sheet saved before thresholds held features loads as the factory data
// has them.
func TestPoolThresholdLegacyOpsMatchFactoryThresholds(t *testing.T) {
	c := check.New(t)
	factory := FactoryAttributeDefs()
	for _, one := range []struct {
		attrID string
		data   string
	}{
		{
			FatiguePointsID,
			`{"state":"Tired","value":"Math.ceil($fp / 3) - 1","explanation":"Move, Dodge and ST are halved (B426)","ops":["halve_move","halve_dodge","halve_st"]}`,
		},
		{
			HitPointsID,
			`{"state":"Reeling","value":"Math.ceil($hp / 3) - 1","explanation":"Move and Dodge are halved (B419)","ops":["halve_move","halve_dodge"]}`,
		},
	} {
		var threshold PoolThreshold
		c.NoError(jio.Unmarshal([]byte(one.data), &threshold))
		var want *PoolThreshold
		for _, t := range factory.Set[one.attrID].Thresholds {
			if t.State == threshold.State {
				want = t
			}
		}
		c.NotNil(want, threshold.State)
		c.Equal(Hash64(want), Hash64(&threshold), threshold.State)
	}
}

// TestPoolThresholdID verifies that a threshold has an ID of its kind, given to one loaded without a valid one and
// kept by one loaded with it, that a clone keeps it, and that the hash leaves it out.
func TestPoolThresholdID(t *testing.T) {
	c := check.New(t)
	fresh := NewPoolThreshold()
	c.True(tid.IsKindAndValid(fresh.TID, kinds.PoolThreshold), "a new threshold has an ID")
	var loaded PoolThreshold
	c.NoError(jio.Unmarshal([]byte(`{"state":"Tired","value":"0"}`), &loaded))
	c.True(tid.IsKindAndValid(loaded.TID, kinds.PoolThreshold), "one loaded without an ID is given one")
	c.NotEqual(fresh.TID, loaded.TID)
	c.NoError(jio.Unmarshal([]byte(`{"id":"tired","state":"Tired","value":"0"}`), &loaded))
	c.True(tid.IsKindAndValid(loaded.TID, kinds.PoolThreshold), "one loaded with an invalid ID is given a new one")
	c.NoError(jio.Unmarshal([]byte(`{"id":"`+string(fresh.TID)+`","state":"Tired","value":"0"}`), &loaded))
	c.Equal(fresh.TID, loaded.TID, "one loaded with an ID keeps it")
	c.Equal(fresh.TID, loaded.Clone().TID, "as does its clone")
	c.Equal(Hash64(fresh), Hash64(NewPoolThreshold()), "the ID is no part of the hash")
	data, err := jio.Marshal(&loaded)
	c.NoError(err)
	c.True(strings.HasPrefix(string(data), `{"id":"`+string(fresh.TID)+`"`), "the ID is written first: %s", data)
}

// TestPoolThresholdCloneCopiesFeatures verifies that a clone carries copies of the features, so that a change to the
// clone's leaves the original alone, which the hash reflects.
func TestPoolThresholdCloneCopiesFeatures(t *testing.T) {
	c := check.New(t)
	original := &PoolThreshold{
		State:    "Tired",
		Value:    "0",
		Features: Features{newMultiplier(DodgeID, fxp.Half)},
	}
	hash := Hash64(original)
	clone := original.Clone()
	c.Equal(hash, Hash64(clone), "the clone hashes as the original")
	bonus, ok := clone.Features[0].(*AttributeBonus)
	c.True(ok)
	bonus.Amount = fxp.One
	c.Equal(hash, Hash64(original), "a change to the clone's feature leaves the original alone")
	c.NotEqual(hash, Hash64(clone), "the clone's hash reflects its feature")
}

// TestPoolThresholdKeepsUnknownMembers verifies that a threshold, and its clone, write back the members this release
// does not know, but none of the legacy members it reads, other than the operations it writes for older versions.
func TestPoolThresholdKeepsUnknownMembers(t *testing.T) {
	c := check.New(t)
	var threshold PoolThreshold
	c.NoError(jio.Unmarshal([]byte(`{"state":"Tired","expression":"$self","ops":["halve_move"],"zzz_future":{"a":1}}`),
		&threshold))
	for _, one := range []struct {
		name      string
		threshold *PoolThreshold
	}{
		{name: "the threshold", threshold: &threshold},
		{name: "its clone", threshold: threshold.Clone()},
	} {
		data, err := jio.Marshal(one.threshold)
		c.NoError(err, one.name)
		c.Equal(`{"a":1}`, jsonMember(t, data, "zzz_future"), one.name)
		c.Equal(`["halve_move"]`, jsonMember(t, data, "ops"), one.name)
		c.Equal("", jsonMember(t, data, "expression"), one.name)
	}
}
