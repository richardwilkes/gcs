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
	"fmt"
	"log/slog"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// entity.wealthCarried and entity.wealthNotCarried sum the extended value of the carried and other equipment lists,
// respectively, matching Entity.WealthCarried and Entity.WealthNotCarried.
func TestScriptEntityWealth(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	carried := NewEquipment(e, nil, false)
	carried.Name = "Sword"
	carried.BaseValue = "10"
	carried.Quantity = fxp.Two
	e.CarriedEquipment = []*Equipment{carried}

	other := NewEquipment(e, nil, false)
	other.Name = "Spare Armor"
	other.BaseValue = "50"
	other.Quantity = fxp.One
	e.OtherEquipment = []*Equipment{other}
	e.Recalculate()

	resolve := func(expr string) string { return ResolveScript(e, ScriptSelfProvider{}, expr) }
	c.Equal("20", resolve("entity.wealthCarried.toString()"))
	c.Equal("50", resolve("entity.wealthNotCarried.toString()"))
}

// TestScriptEntityExistsAndMissingAttribute verifies that a script can tell whether it has an entity, and that
// entity.attribute answers undefined for an attribute the entity does not have.
func TestScriptEntityExistsAndMissingAttribute(t *testing.T) {
	c := check.New(t)
	c.Equal("false", ResolveScript(nil, ScriptSelfProvider{}, "entity.exists"))
	e := NewEntity()
	c.Equal("true", ResolveScript(e, ScriptSelfProvider{}, "entity.exists"))
	c.Equal("undefined", ResolveScript(e, ScriptSelfProvider{}, `typeof entity.attribute("nope")`))
	c.Equal("object", ResolveScript(e, ScriptSelfProvider{}, `typeof entity.attribute("st")`))
}

// TestScriptEntityProfile verifies that the profile fields a script reads come from the entity's profile, and that
// reading any of them, but nothing else, is recorded, since the sheet re-derives its rows after a profile change only
// when a script has read the profile.
func TestScriptEntityProfile(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	p := &e.Profile
	p.PlayerName, p.Name, p.Title, p.Organization, p.Religion = "Player", "Name", "Title", "Organization", "Religion"
	p.TechLevel, p.Gender, p.Age, p.Birthday = "3", "Gender", "Age", "Birthday"
	p.Eyes, p.Hair, p.Skin, p.Handedness = "Eyes", "Hair", "Skin", "Handedness"
	p.Height = fxp.Length(fxp.FromInteger(70))
	p.Weight = fxp.Weight(fxp.FromInteger(150))
	for _, tc := range []struct {
		field string
		want  string
	}{
		{field: "playerName", want: "Player"},
		{field: "name", want: "Name"},
		{field: "title", want: "Title"},
		{field: "organization", want: "Organization"},
		{field: "religion", want: "Religion"},
		{field: "techLevel", want: "3"},
		{field: "gender", want: "Gender"},
		{field: "age", want: "Age"},
		{field: "birthday", want: "Birthday"},
		{field: "eyes", want: "Eyes"},
		{field: "hair", want: "Hair"},
		{field: "skin", want: "Skin"},
		{field: "handedness", want: "Handedness"},
		{field: "heightInInches", want: "70"},
		{field: "weightInPounds", want: "150"},
	} {
		e.ClearProfileReadByScript()
		c.Equal(tc.want, ResolveScript(e, ScriptSelfProvider{}, "entity."+tc.field), tc.field)
		c.True(e.ProfileReadByScript(), "reading %s is recorded", tc.field)
	}
	e.ClearProfileReadByScript()
	c.Equal("0", ResolveScript(e, ScriptSelfProvider{}, "entity.sizeModifier"))
	c.False(e.ProfileReadByScript(), "reading anything else is not recorded")
}

// TestScriptEntityCurrentEncumbrance verifies that entity.currentEncumbrance gives the encumbrance level, counting the
// weight that is ignored for skills only when not asked for the level for skills, or the move factor of that level: 1
// less a fifth per level.
func TestScriptEntityCurrentEncumbrance(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	load := addCarriedEquipmentWithFeatures(e, "Load")
	load.BaseWeight = "50 lb"
	load.WeightIgnoredForSkills = true
	e.Recalculate()
	c.Equal(fxp.FromInteger(20), fxp.Int(e.BasicLift()), "precondition: 50 lb is a medium load for ST 10")
	for _, tc := range []struct {
		forSkills  bool
		moveFactor bool
		want       string
	}{
		{forSkills: false, moveFactor: false, want: "2"},
		{forSkills: false, moveFactor: true, want: "0.6"},
		{forSkills: true, moveFactor: false, want: "0"},
		{forSkills: true, moveFactor: true, want: "1"},
	} {
		c.Equal(tc.want, ResolveScript(e, ScriptSelfProvider{},
			fmt.Sprintf("entity.currentEncumbrance(%v, %v)", tc.forSkills, tc.moveFactor)),
			"forSkills %v, moveFactor %v", tc.forSkills, tc.moveFactor)
	}
}

// TestScriptRandomWeightFollowsBuild verifies that entity.randomWeightInPounds scales the B18 weight for an ST, 145 lb
// give or take 30 at ST 10 and 235 lb give or take 48 at ST 16, by a shift over 3: the 3 of an average build, less 1
// for Skinny or plus 1, 2 or 3 for Overweight, Fat or Very Fat, whose names match without regard to case, plus the
// shift it is given. Without an entity there is no build to read. Each row passes the shift that cancels the build's,
// so it weighs exactly 0 whatever the roll.
func TestScriptRandomWeightFollowsBuild(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		trait    string
		shift    int
		noEntity bool
	}{
		{shift: -3},
		{shift: -3, noEntity: true},
		{trait: "Skinny", shift: -2},
		{trait: "Overweight", shift: -4},
		{trait: "fat", shift: -5},
		{trait: "Very Fat", shift: -6},
	} {
		var e *Entity
		if !tc.noEntity {
			e = NewEntity()
			if tc.trait != "" {
				addTraitWithFeatures(e, tc.trait)
			}
			e.Recalculate()
		}
		c.Equal("0", ResolveScript(e, ScriptSelfProvider{}, fmt.Sprintf("entity.randomWeightInPounds(10, %d)",
			tc.shift)), "trait %q, shift %d, no entity %v", tc.trait, tc.shift, tc.noEntity)
	}
	e := NewEntity()
	e.Recalculate()
	for _, tc := range []struct {
		st     int
		lo, hi int
	}{
		{st: 10, lo: 115, hi: 175},
		{st: 16, lo: 187, hi: 283},
	} {
		weight := ResolveToNumber(e, ScriptSelfProvider{}, fmt.Sprintf("entity.randomWeightInPounds(%d, 0)", tc.st))
		c.True(weight >= fxp.FromInteger(tc.lo) && weight <= fxp.FromInteger(tc.hi),
			"an average build at ST %d weighs %v", tc.st, weight)
	}
}

// TestScriptEntitySkillLevelExcludesSkillBeingResolved verifies that entity.skillLevel answers 0 for a skill whose
// level is already being resolved, rather than resolving it again.
func TestScriptEntitySkillLevelExcludesSkillBeingResolved(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Name = "Guns"
	sk.Specialization = "Rifle"
	sk.Points = fxp.One
	e.Skills = []*Skill{sk}
	e.Recalculate()
	const script = `entity.skillLevel("Guns", "Rifle", false)`
	c.Equal(fmt.Sprint(sk.LevelData.Level.AsInteger[int]()), ResolveScript(e, ScriptSelfProvider{}, script))
	e.DiscardCaches()
	countLogs(t, slog.LevelError) // The attempt is logged as an error.
	e.registerSkillLevelResolutionExclusion("Guns", "Rifle", "")
	defer e.unregisterSkillLevelResolutionExclusion("Guns", "Rifle", "")
	c.Equal("0", ResolveScript(e, ScriptSelfProvider{}, script))
}
