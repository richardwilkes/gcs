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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
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
		{
			"features that differ from that in an amount",
			`{"state":"Tired","value":"0","features":[{"type":"attribute_bonus","attribute":"dodge","multiply":true,"amount":0.25}]}`,
			[]string{"×0.25 to Dodge"},
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

// TestPoolThresholdLegacyValue verifies how a threshold saved before its value was a script gets one: a value given
// outright is kept, an expression is converted to a script, and without either, the multiplier, divisor and addition
// become the expression round($self*multiplier/divisor+addition), which is just the addition for a multiplier of 0 and
// needs no rounding for one of 1 or -1 with nothing to divide by.
func TestPoolThresholdLegacyValue(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		data string
		want string
	}{
		{`{"value":"7","expression":"9","multiplier":2}`, "7"},
		{`{"expression":"$self/2"}`, ExprToScript("$self/2")},
		{`{"addition":5}`, ExprToScript("5")},
		{`{"multiplier":-1}`, ExprToScript("-$self")},
		{`{"multiplier":-1,"divisor":1,"addition":2}`, ExprToScript("-$self+2")},
		{`{"multiplier":1,"divisor":1}`, ExprToScript("$self")},
		{`{"multiplier":1,"addition":-1}`, ExprToScript("$self-1")},
		{`{"multiplier":1,"divisor":3,"addition":-1}`, ExprToScript("round($self/3-1)")},
		{`{"multiplier":-1,"divisor":2}`, ExprToScript("round(-$self/2)")},
		{`{"multiplier":2,"divisor":3}`, ExprToScript("round($self*2/3)")},
		{`{"multiplier":-2}`, ExprToScript("round(-$self*2)")},
	} {
		var threshold PoolThreshold
		c.NoError(jio.Unmarshal([]byte(one.data), &threshold), one.data)
		c.Equal(one.want, threshold.Value, one.data)
	}
}

// TestThresholdIgnores verifies which of a pool threshold's features do nothing: one of a type a threshold may not
// hold, an amount per level, though not a factor marked as one, and a bonus to this armor or this weapon.
func TestThresholdIgnores(t *testing.T) {
	c := check.New(t)
	perLevel := func(f Feature) Feature {
		switch one := f.(type) {
		case *AttributeBonus:
			one.PerLevel = true
		case *ConditionalModifierBonus:
			one.PerLevel = true
		case *DRBonus:
			one.PerLevel = true
		case *SkillBonus:
			one.PerLevel = true
		case *SkillPointBonus:
			one.PerLevel = true
		case *SpellBonus:
			one.PerLevel = true
		case *SpellPointBonus:
			one.PerLevel = true
		case *TraitBonus:
			one.PerLevel = true
		case *WeaponBonus:
			one.PerLevel = true
		}
		return f
	}
	thisArmor := NewDRBonus()
	thisArmor.Locations = nil
	thisWeaponSkill := NewSkillBonus()
	thisWeaponSkill.SelectionType = skillsel.ThisWeapon
	thisWeapon := NewWeaponBonus(feature.WeaponBonus)
	thisWeapon.SelectionType = wsel.ThisWeapon
	byName := NewWeaponBonus(feature.WeaponBonus)
	byName.SelectionType = wsel.WithName
	defs := FactoryAttributeDefs()
	for _, one := range []struct {
		name    string
		feature Feature
		ignored bool
	}{
		{"a type a threshold may not hold", NewCostReduction(StrengthID), true},
		{"a factor marked as per level", perLevel(newMultiplier(DexterityID, fxp.Half)), false},
		{"a conditional modifier", NewConditionalModifierBonus(), false},
		{"a conditional modifier per level", perLevel(NewConditionalModifierBonus()), true},
		{"a DR bonus to a location", NewDRBonus(), false},
		{"a DR bonus per level", perLevel(NewDRBonus()), true},
		{"a DR bonus to this armor", thisArmor, true},
		{"a skill bonus to this weapon", thisWeaponSkill, true},
		{"a skill bonus per level", perLevel(NewSkillBonus()), true},
		{"a skill point bonus per level", perLevel(NewSkillPointBonus()), true},
		{"a spell bonus per level", perLevel(NewSpellBonus()), true},
		{"a spell point bonus per level", perLevel(NewSpellPointBonus()), true},
		{"a trait bonus per level", perLevel(NewTraitBonus()), true},
		{"a weapon bonus by name", byName, false},
		{"a weapon bonus per level", perLevel(NewWeaponBonus(feature.WeaponBonus)), true},
		{"a weapon bonus to this weapon", thisWeapon, true},
	} {
		c.Equal(one.ignored, ThresholdIgnores(defs, one.feature), one.name)
	}
}

// TestDescribeThresholdFeature verifies that a threshold's switchable feature is described as one that takes effect
// whatever the switch, since a threshold has none, without the feature itself being changed.
func TestDescribeThresholdFeature(t *testing.T) {
	c := check.New(t)
	em := func(s string) string { return s }
	bonus := newAttrBonus(DexterityID, -fxp.One)
	plain := bonus.Describe(nil, nil, em)
	bonus.Switchable = true
	c.NotEqual(plain, bonus.Describe(nil, nil, em), "precondition: a switchable feature says so")
	c.Equal(plain, DescribeThresholdFeature(bonus, nil, nil, em))
	c.True(bonus.IsSwitchable(), "the feature itself stays switchable")
}
