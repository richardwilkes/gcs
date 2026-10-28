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
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// newGrantLeveledTrait returns a trait outside any sheet with the given name, costing perLevel points a level.
func newGrantLeveledTrait(name string, perLevel fxp.Int) *Trait {
	t := NewTrait(nil, nil, false)
	t.Name = name
	t.CanLevel = true
	t.PointsPerLevel = perLevel
	t.Levels = fxp.One
	return t
}

// newEasySkill returns an Easy DX skill outside any sheet with the given name and points.
func newEasySkill(name string, points fxp.Int) *Skill {
	s := NewSkill(nil, nil, false)
	s.Name = name
	s.Difficulty.Difficulty = difficulty.Easy
	s.Points = points
	return s
}

// newTraitGrant returns a grant that sets the trait's level.
func newTraitGrant(item *Trait, level fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantTrait)
	g.Trait = item
	g.Amount = level
	g.Merge = merge
	return g
}

// addingLevels makes the trait grant add its levels to the trait rather than set the trait's level.
func addingLevels(g *Grant) *Grant {
	g.Adjust = grantadjust.AddLevels
	return g
}

// newSkillGrant returns a grant that adds the levels to the skill.
func newSkillGrant(item *Skill, levels fxp.Int, merge grantmerge.Mode) *Grant {
	return newSkillGrantOf(item, grantadjust.AddLevels, levels, merge)
}

func newSkillGrantOf(item *Skill, kind grantadjust.Kind, amount fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantSkill)
	g.Skill = item
	g.Adjust = kind
	g.Amount = amount
	g.Merge = merge
	return g
}

// newSpellGrant returns a grant that adds the levels to the spell.
func newSpellGrant(item *Spell, levels fxp.Int, merge grantmerge.Mode) *Grant {
	return newSpellGrantOf(item, grantadjust.AddLevels, levels, merge)
}

func newSpellGrantOf(item *Spell, kind grantadjust.Kind, amount fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantSpell)
	g.Spell = item
	g.Adjust = kind
	g.Amount = amount
	g.Merge = merge
	return g
}

// newHardSpell returns a Hard IQ spell outside any sheet with the given name and points.
func newHardSpell(name string, points fxp.Int) *Spell {
	s := NewSpell(nil, nil, false)
	s.Name = name
	s.Points = points
	return s
}

// traitPrereqOf returns the list's first prerequisite, which must be a trait prerequisite.
func traitPrereqOf(c check.Checker, list *PrereqList) *TraitPrereq {
	c.Helper()
	p, ok := list.Prereqs[0].(*TraitPrereq)
	if !ok {
		c.Fatalf("the first prerequisite is a %T, not a trait prerequisite", list.Prereqs[0])
	}
	return p
}

// grantedOf returns the items at the top of the list a grant put there.
func grantedOf[T grantNode[T]](list []T) []T {
	var result []T
	for _, one := range list {
		if IsGranted(one) {
			result = append(result, one)
		}
	}
	return result
}

// addTrait puts the trait on the sheet.
func addTrait(e *Entity, t *Trait) *Trait {
	t.SetDataOwner(e)
	e.Traits = append(e.Traits, t)
	return t
}

// addSkill puts an Easy DX skill with the given name and points on the sheet.
func addSkill(e *Entity, name string, points fxp.Int) *Skill {
	s := newEasySkill(name, points)
	s.SetDataOwner(e)
	e.Skills = append(e.Skills, s)
	return s
}

// newHardSkill returns a Hard IQ skill outside any sheet with the given name and points.
func newHardSkill(name string, points fxp.Int) *Skill {
	s := NewSkill(nil, nil, false)
	s.Name = name
	s.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Hard}
	s.Points = points
	return s
}

// addHardSkill puts a Hard IQ skill with the given name and points on the sheet.
func addHardSkill(e *Entity, name string, points fxp.Int) *Skill {
	s := newHardSkill(name, points)
	s.SetDataOwner(e)
	e.Skills = append(e.Skills, s)
	return s
}

// newTechniqueOf returns an Average technique outside any sheet with the given name and points, resting on the skill
// named base at +0, with no limit.
func newTechniqueOf(base, name string, points fxp.Int) *Skill {
	s := NewTechnique(nil, nil, base)
	s.Name = name
	s.Points = points
	return s
}

// newKicking returns Kicking, a Hard technique outside any sheet with the given points, resting on Karate at -2 and
// going no higher than Karate.
func newKicking(points fxp.Int) *Skill {
	s := newTechniqueOf("Karate", "Kicking", points)
	s.Difficulty.Difficulty = difficulty.Hard
	s.TechniqueDefault.Modifier = -fxp.Two
	limit := fxp.Int(0)
	s.TechniqueLimitModifier = &limit
	return s
}

// newDisarming returns Disarming, a Hard technique outside any sheet with the given points, resting on Karate at +0 and
// going up to 5 above it.
func newDisarming(points fxp.Int) *Skill {
	s := newTechniqueOf("Karate", "Disarming", points)
	s.Difficulty.Difficulty = difficulty.Hard
	limit := fxp.Five
	s.TechniqueLimitModifier = &limit
	return s
}

// newBless returns Bless, a Hard Ritual Magic spell outside any sheet with the given points and 3 prerequisites, which
// rests on the skill Ritual Magic at -3, going no higher than it.
func newBless(points fxp.Int) *Spell {
	s := NewRitualMagicSpell(nil, nil, false)
	s.Name = "Bless"
	s.PrereqCount = 3
	s.Points = points
	return s
}

func TestGrantPutsTraitOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	grant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
	racial := addTraitWithFeatures(e, "Racial", grant)
	racial.BasePoints = fxp.Five
	e.Recalculate()
	granted := grantedOf(e.Traits)
	c.Equal(1, len(granted))
	nightVision := granted[0]
	c.Equal("Night Vision", nightVision.Name)
	c.Equal(racial.TID, nightVision.Granted.Grantor)
	c.Equal(grant.Key(), nightVision.Granted.Key)
	c.NotEqual(grant.Key(), nightVision.TID, "the item on the sheet gets its own ID")
	c.True(nightVision.Enabled())
	c.Equal(fxp.Two, nightVision.CurrentLevel())
	c.Equal(fxp.Int(0), nightVision.AdjustedPoints(nil), "a granted trait costs nothing")
	c.Equal(NumericRangeOf(0), nightVision.PointsRange(nil))
	c.Equal(fxp.Two, nightVision.ClassifyingPoints(), "an export still files it by what it would cost")
	c.Equal(fxp.Five, e.PointsBreakdown().Advantages)
	needsPoints := addTraitWithFeatures(e, "Needs Night Vision")
	needsPoints.Prereq = newPrereqListForTrait("Night Vision", true)
	traitPrereqOf(c, needsPoints.Prereq).PointsCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Two}
	e.Recalculate()
	c.Equal("", needsPoints.UnsatisfiedReason, "a prerequisite's points criterion sees what the trait would cost too")

	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)), "recalculating doesn't add another")
	c.True(nightVision == grantedOf(e.Traits)[0], "an unchanged grant leaves its item alone")
}

func TestGrantNotInEffectKeepsItsItemDisabled(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	grant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
	racial := addTraitWithFeatures(e, "Racial", grant)
	e.Recalculate()
	nightVision := grantedOf(e.Traits)[0]
	c.True(nightVision.Enabled())

	grant.Switchable = true
	e.Recalculate()
	nightVision = grantedOf(e.Traits)[0]
	c.Equal(1, len(grantedOf(e.Traits)), "the item stays while its grant is not in effect")
	c.True(nightVision.grantNotInEffect(), "a switchable grant needs its item switched on")
	c.False(nightVision.Enabled())
	c.False(nightVision.IsDisabled(), "the user's own flag is left alone")
	c.Equal(fxp.Int(0), nightVision.CurrentLevel())

	racial.SwitchedOn = true
	e.Recalculate()
	c.False(nightVision.grantNotInEffect())

	racial.Features = nil
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "the item goes with its grant")
}

func TestGrantFromEquipmentNeedsItEquipped(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	grant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)
	goggles := addCarriedEquipmentWithFeatures(e, "Goggles", grant)
	goggles.Equipped = true
	e.Recalculate()
	nightVision := grantedOf(e.Traits)[0]
	c.False(nightVision.grantNotInEffect())

	goggles.Equipped = false
	e.Recalculate()
	c.True(nightVision.grantNotInEffect())
}

func TestGrantFromOtherEquipmentPutsNothingOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	grant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)
	goggles := addCarriedEquipmentWithFeatures(e, "Goggles", grant)
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)))

	e.CarriedEquipment = nil
	e.OtherEquipment = append(e.OtherEquipment, goggles)
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "equipment the character doesn't have on them grants nothing")

	e.OtherEquipment = nil
	e.CarriedEquipment = append(e.CarriedEquipment, goggles)
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)), "carrying it again grants the trait again")
}

func TestGrantedSkillHasThePointsItsGrantGives(t *testing.T) {
	for _, tc := range []struct {
		name   string
		adjust grantadjust.Kind
		amount fxp.Int
		// defaulted, when set, has the skill default to Karate-2 at 12, which stands for 4 points.
		defaulted bool
		points    fxp.Int
		level     fxp.Int
	}{
		{name: "levels added", adjust: grantadjust.AddLevels, amount: fxp.Three, points: fxp.Four, level: fxp.Twelve},
		{name: "level set", adjust: grantadjust.SetLevel, amount: fxp.Twelve, points: fxp.Four, level: fxp.Twelve},
		{
			// The 8 points that buy DX+3 count those the default stands for.
			name: "level set above the skill's default", adjust: grantadjust.SetLevel, amount: fxp.Thirteen,
			defaulted: true, points: fxp.Four, level: fxp.Thirteen,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			item := newEasySkill("Brawling", fxp.One)
			var bought fxp.Int
			if tc.defaulted {
				bought = addSkill(e, "Karate", fxp.Twelve).Points
				item.Defaults = []*SkillDefault{newSkillDefaultTo("Karate", "", true, -fxp.Two)}
			}
			addTraitWithFeatures(e, "Training", newSkillGrantOf(item, tc.adjust, tc.amount, grantmerge.Ignore))
			needsBrawling := addTraitWithFeatures(e, "Needs Brawling")
			needsBrawling.Prereq = NewPrereqList()
			addSkillPrereq(needsBrawling.Prereq, "Brawling", fxp.One)
			e.Recalculate()
			granted := grantedOf(e.Skills)
			c.Equal(1, len(granted))
			brawling := granted[0]
			c.Equal(fxp.Int(0), brawling.Points, "a granted skill has no points of its own")
			c.Equal(tc.points, brawling.AdjustedPoints(nil))
			c.Equal(tc.level, brawling.LevelData.Level)
			c.Equal(bought, e.PointsBreakdown().Skills, "a granted skill costs nothing")
			c.Equal("0", ResolveScript(e, deferredNewScriptSkill(brawling), "self.points"), "nor does it to a script")
			c.Equal("", needsBrawling.UnsatisfiedReason, "a granted skill counts as having points")
		})
	}
}

func TestGrantGivingNothingPutsNothingOnSheet(t *testing.T) {
	wildcard := newEasySkill("Brawling", 0)
	wildcard.Difficulty.Difficulty = difficulty.Wildcard
	for _, tc := range []struct {
		name  string
		grant *Grant
	}{
		{name: "levels removed", grant: newSkillGrant(newEasySkill("Brawling", 0), -fxp.One, grantmerge.Ignore)},
		{name: "no levels", grant: newSkillGrant(newEasySkill("Brawling", 0), 0, grantmerge.Ignore)},
		{
			name:  "points removed",
			grant: newSkillGrantOf(newEasySkill("Brawling", 0), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
		},
		{
			name:  "points too few to buy a level",
			grant: newSkillGrantOf(newEasySkill("Brawling", 0), grantadjust.AddPoints, fxp.Half, grantmerge.Ignore),
		},
		{
			name:  "points too few to buy a wildcard skill's first level",
			grant: newSkillGrantOf(wildcard, grantadjust.AddPoints, fxp.Two, grantmerge.Ignore),
		},
		{
			name:  "level set below 1",
			grant: newSkillGrantOf(newEasySkill("Brawling", 0), grantadjust.SetLevel, 0, grantmerge.KeepHighest),
		},
		{name: "trait below level 1", grant: addingLevels(newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One),
			fxp.Half, grantmerge.Ignore))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			addTraitWithFeatures(e, "Curse", tc.grant)
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Skills))+len(grantedOf(e.Traits)))
		})
	}
}

func TestGrantedRowIsTheKindOfItemItsGrantGives(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	pack := NewTrait(nil, nil, true)
	pack.Name = "Pack"
	child := newGrantLeveledTrait("Child", fxp.Five)
	child.SetParent(pack)
	pack.Children = []*Trait{child}
	addTraitWithFeatures(e, "Racial", newTraitGrant(pack, fxp.One, grantmerge.Ignore))
	for range 3 {
		e.Recalculate()
	}
	c.Equal(0, len(grantedOf(e.Traits)), "a container is not granted, as the picker leaves containers out")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Advantages)

	stale := pack.Clone(LibraryFile{}, e, nil, Duplicate)
	stale.Granted = &GrantedBy{Grantor: e.Traits[0].TID, Key: pack.TID, Hash: "stale"}
	e.Traits = append(e.Traits, stale)
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "a granted container row from a file goes as any row no grant keeps does")

	// A file can give the row a grant put on the sheet an ID of another kind, keeping the grant's link whether or not
	// its hash still matches, which turns a trait row into a container of other traits, a skill row into a technique or
	// back, or a spell row into ritual magic or back. The grant makes its item again in that row's place, leaving any
	// children behind, as it does when only the hash no longer matches.
	for _, tc := range []struct {
		name  string
		grant func() *Grant
		kind  byte
	}{
		{
			name: "trait as a container",
			grant: func() *Grant {
				return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
			},
			kind: kinds.TraitContainer,
		},
		{
			name: "skill as a technique",
			grant: func() *Grant {
				return newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore)
			},
			kind: kinds.Technique,
		},
		{
			name: "technique as a skill",
			grant: func() *Grant {
				return newSkillGrant(newTechniqueOf("Brawling", "Disarming", fxp.One), fxp.Three, grantmerge.Ignore)
			},
			kind: kinds.Skill,
		},
		{
			name: "spell as ritual magic",
			grant: func() *Grant {
				return newSpellGrant(newHardSpell("Light", fxp.One), fxp.Three, grantmerge.Ignore)
			},
			kind: kinds.RitualMagicSpell,
		},
		{
			name: "ritual magic spell as a regular one",
			grant: func() *Grant {
				ritual := NewRitualMagicSpell(nil, nil, false)
				ritual.Name = "Bless"
				return newSpellGrant(ritual, fxp.Three, grantmerge.Ignore)
			},
			kind: kinds.Spell,
		},
	} {
		for _, hash := range []string{"matching", "stale"} {
			what := tc.name + ", " + hash + " hash"
			e = NewEntity()
			g := tc.grant()
			addTraitWithFeatures(e, "Carrier", g)
			switch g.Type {
			case feature.GrantTrait:
				checkGrantRemakesRowOfOtherKind(c, e, &e.Traits, tc.kind, hash, what)
			case feature.GrantSkill:
				checkGrantRemakesRowOfOtherKind(c, e, &e.Skills, tc.kind, hash, what)
			case feature.GrantSpell:
				checkGrantRemakesRowOfOtherKind(c, e, &e.Spells, tc.kind, hash, what)
			}
		}
	}

	// The grant's copy of its item can also have an ID of a kind its type doesn't have, which the item the grant makes
	// from it doesn't take, so the row is not made again on that account.
	e = NewEntity()
	foreign := newGrantLeveledTrait("Night Vision", fxp.One)
	foreign.TID = tid.MustNewTID(kinds.Skill)
	addTraitWithFeatures(e, "Carrier", newTraitGrant(foreign, fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	row := grantedOf(e.Traits)[0]
	e.Recalculate()
	c.True(row == grantedOf(e.Traits)[0], "an unchanged grant leaves its item alone, whatever kind its copy's ID has")
}

// checkGrantRemakesRowOfOtherKind gives the row a grant put at the top of the list an ID of the kind, along with a
// child when that makes it a container, and the hash unless it is "matching", as a file can, and checks that the grant
// then makes its item again in the row's place, as it was before.
func checkGrantRemakesRowOfOtherKind[T grantNode[T]](c check.Checker, e *Entity, list *[]T, kind byte, hash,
	what string,
) {
	c.Helper()
	level := func(row T) fxp.Int {
		switch n := any(row).(type) {
		case *Trait:
			return n.Levels
		case *Skill:
			return n.LevelData.Level
		case *Spell:
			return n.LevelData.Level
		}
		return 0
	}
	e.Recalculate()
	granted := grantedOf(*list)
	if len(granted) != 1 {
		c.Fatalf("expected the grant to put one row on the sheet, %s, got %d", what, len(granted))
	}
	row := granted[0]
	at := slices.Index(*list, row)
	want := Hash64(row)
	wantLevel := level(row)
	switch n := any(row).(type) {
	case *Trait:
		n.TID = tid.MustNewTID(kind)
		if n.Container() {
			kid := newGrantLeveledTrait("Child", fxp.Five)
			kid.SetDataOwner(e)
			kid.SetParent(n)
			n.Children = []*Trait{kid}
		}
	case *Skill:
		n.TID = tid.MustNewTID(kind)
	case *Spell:
		n.TID = tid.MustNewTID(kind)
	}
	if hash != "matching" {
		(*grantLink(row)).Hash = hash
	}
	e.Recalculate()
	granted = grantedOf(*list)
	if len(granted) != 1 {
		c.Fatalf("expected one granted row, %s, got %d", what, len(granted))
	}
	c.Equal(want, Hash64(granted[0]), "the row is the item the grant gives again, %s", what)
	c.False(granted[0].HasChildren(), what)
	c.Equal(wantLevel, level(granted[0]), what)
	c.Equal(at, slices.Index(*list, granted[0]), "in its place, %s", what)
}

func TestGrantSkillMergeModes(t *testing.T) {
	typeKey := nameableKeyOf("@Type@")
	// boughtWith returns a setup that gives the skill bought the points.
	boughtWith := func(points fxp.Int) func(*Entity, *Skill, *Grant, *Trait) {
		return func(_ *Entity, bought *Skill, _ *Grant, _ *Trait) { bought.Points = points }
	}
	// defaultingToKarate gives the skill bought no points and, along with the grant's copy, a default to Karate-2, at
	// 12 with Karate at 14, which stands for 4 points.
	defaultingToKarate := func(e *Entity, bought *Skill, g *Grant, _ *Trait) {
		bought.Points = 0
		for _, sk := range []*Skill{bought, g.Skill} {
			sk.Defaults = []*SkillDefault{newSkillDefaultTo("Karate", "", true, -fxp.Two)}
		}
		addSkill(e, "Karate", fxp.Twelve)
	}
	for _, tc := range []struct {
		name   string
		merge  grantmerge.Mode
		adjust grantadjust.Kind
		amount fxp.Int
		// setup, when set, adjusts the skill bought, with 1 point, the grant and its carrier before the
		// recalculation.
		setup  func(e *Entity, bought *Skill, g *Grant, carrier *Trait)
		points fxp.Int
		level  fxp.Int
	}{
		{
			name: "ignore a level set", merge: grantmerge.Ignore, adjust: grantadjust.SetLevel, amount: fxp.Twelve,
			points: fxp.One, level: fxp.Ten,
		},
		{name: "add levels", merge: grantmerge.Ignore, amount: fxp.Three, points: fxp.One, level: fxp.FromInteger(13)},
		{
			name: "add points", merge: grantmerge.Ignore, adjust: grantadjust.AddPoints, amount: fxp.Four,
			points: fxp.Five, level: fxp.FromInteger(12),
		},
		{
			name: "keep highest of a level set", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel,
			amount: fxp.Twelve, points: fxp.One, level: fxp.Twelve,
		},
		{
			name: "keep highest of a level set lower", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel,
			amount: fxp.Eleven, setup: boughtWith(fxp.Four), points: fxp.Four, level: fxp.Twelve,
		},
		{
			// The level set is that before bonuses, which the level kept gets, as the skill's own does.
			name: "keep highest of a level set with placeholder and bonus", merge: grantmerge.KeepHighest,
			adjust: grantadjust.SetLevel, amount: fxp.Thirteen,
			setup: func(e *Entity, bought *Skill, g *Grant, carrier *Trait) {
				bought.Specialization = "@Type@"
				bought.Replacements = map[string]string{typeKey: "Pistol"}
				g.Skill.Specialization = "@Type@"
				carrier.Replacements = map[string]string{typeKey: "Pistol"}
				bonus := newSkillBonusTo("Brawling", fxp.Two)
				bonus.SpecializationCriteria.Compare = criteria.IsText
				bonus.SpecializationCriteria.Qualifier = "Pistol"
				addTraitWithFeatures(e, "Aim", bonus)
			}, points: fxp.One, level: fxp.FromInteger(15),
		},
		{
			// An optional specialization makes the Average skill on the sheet Easy, where 4 points buy level 12, but
			// leaves the grant's copy Average, where they would take 8, which buy level 13 on the sheet.
			name: "keep highest of a level set as the skill on the sheet prices it", merge: grantmerge.KeepHighest,
			adjust: grantadjust.SetLevel, amount: fxp.Twelve,
			setup: func(_ *Entity, bought *Skill, g *Grant, _ *Trait) {
				bought.Difficulty.Difficulty = difficulty.Average
				bought.OptionalSpecialization = "Boxing"
				g.Skill.Difficulty.Difficulty = difficulty.Average
			}, points: fxp.One, level: fxp.Twelve,
		},
		{
			name: "keep highest of a level set with a missing-equipment penalty", merge: grantmerge.KeepHighest,
			adjust: grantadjust.SetLevel, amount: fxp.Twelve,
			setup: func(_ *Entity, bought *Skill, g *Grant, _ *Trait) {
				for _, sk := range []*Skill{bought, g.Skill} {
					sk.Prereq = NewPrereqList()
					p := NewEquippedEquipmentPrereq()
					p.Parent = sk.Prereq
					p.NameCriteria.Qualifier = "Lockpicks"
					sk.Prereq.Prereqs = append(sk.Prereq.Prereqs, p)
				}
			}, points: fxp.One, level: fxp.Seven,
		},
		{
			name: "keep lowest of a level set", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			amount: fxp.Eleven, setup: boughtWith(fxp.Eight), points: fxp.Eight, level: fxp.Eleven,
		},
		{
			name: "keep lowest of a level set higher", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			amount: fxp.Twelve, points: fxp.One, level: fxp.Ten,
		},
		{
			// The level set is compared as it is, not as the 8 points that buy it would buy along with the 4 the
			// default to Karate-2 stands for.
			name: "keep highest of a level set above the skill's default", merge: grantmerge.KeepHighest,
			adjust: grantadjust.SetLevel, amount: fxp.Thirteen, setup: defaultingToKarate, points: 0,
			level: fxp.Thirteen,
		},
		{
			name: "keep lowest of a level set below the skill's default", merge: grantmerge.KeepLowest,
			adjust: grantadjust.SetLevel, amount: fxp.Eleven, setup: defaultingToKarate, points: 0, level: fxp.Eleven,
		},
		{name: "curse levels", merge: grantmerge.Ignore, amount: -fxp.Two, points: fxp.One, level: fxp.Eight},
		{
			// The level goes no lower than 0.
			name: "curse more levels than it has", merge: grantmerge.Ignore, amount: -fxp.Twelve, points: fxp.One,
			level: 0,
		},
		{
			// The points come off as they are, not as the points of that many levels, which would take 4.
			name: "curse points", merge: grantmerge.Ignore, adjust: grantadjust.AddPoints, amount: -fxp.Three,
			setup: boughtWith(fxp.Eight), points: fxp.Five, level: fxp.Twelve,
		},
		{
			// Never below zero, and with no points left, the skill has no level.
			name: "curse more points than it has", merge: grantmerge.Ignore, adjust: grantadjust.AddPoints,
			amount: -fxp.Two, points: 0, level: fxp.Min,
		},
		// A skill with no points and no default has no level of its own to add to or compare with, so it gets the
		// points the grant would give it standing alone, unless the grant is a level set under Ignore.
		{
			name: "add levels to a skill with no level", merge: grantmerge.Ignore, amount: fxp.Two,
			setup: boughtWith(0), points: fxp.Two, level: fxp.Eleven,
		},
		{
			name: "curse levels of a skill with no level", merge: grantmerge.Ignore, amount: -fxp.Two,
			setup: boughtWith(0), points: 0, level: fxp.Min,
		},
		{
			name: "ignore a level set for a skill with no level", merge: grantmerge.Ignore, adjust: grantadjust.SetLevel,
			amount: fxp.Twelve, setup: boughtWith(0), points: 0, level: fxp.Min,
		},
		{
			name: "keep highest of a level set for a skill with no level", merge: grantmerge.KeepHighest,
			adjust: grantadjust.SetLevel, amount: fxp.Twelve, setup: boughtWith(0), points: fxp.Four, level: fxp.Twelve,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			brawling := addSkill(e, "Brawling", fxp.One)
			g := newSkillGrantOf(newEasySkill("Brawling", fxp.One), tc.adjust, tc.amount, tc.merge)
			training := addTraitWithFeatures(e, "Training", g)
			if tc.setup != nil {
				tc.setup(e, brawling, g, training)
			}
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Skills)), "a grant whose skill is already present adds no row")
			c.Equal(tc.points, brawling.AdjustedPoints(nil))
			c.Equal(tc.level, brawling.LevelData.Level)
			var bought fxp.Int
			for _, sk := range e.Skills {
				bought += sk.Points
			}
			c.Equal(bought, e.PointsBreakdown().Skills, "merged points cost nothing")
		})
	}
}

func TestGrantMergesIntoGrantedSkillAsIntoBought(t *testing.T) {
	for _, tc := range []struct {
		name string
		// planterPoints is set when the grant that puts the skill on the sheet, when it isn't bought, gives it the 4
		// points that buy level 12 by adding them; otherwise it sets that level under Ignore, which ranks it ahead of
		// every grant below for putting the skill there (see comparePlanting).
		planterPoints bool
		merge         grantmerge.Mode
		adjust        grantadjust.Kind
		amount        fxp.Int
		level         fxp.Int
	}{
		{name: "add levels", merge: grantmerge.Ignore, amount: fxp.Two, level: fxp.FromInteger(14)},
		{
			name: "keep highest of a level set", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel,
			amount: fxp.Thirteen, level: fxp.Thirteen,
		},
		{
			name: "keep lowest of a level set", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			amount: fxp.Eleven, level: fxp.Eleven,
		},
		{name: "curse levels", merge: grantmerge.Ignore, amount: -fxp.One, level: fxp.FromInteger(11)},
		{
			name: "curse points", merge: grantmerge.Ignore, adjust: grantadjust.AddPoints, amount: -fxp.One,
			level: fxp.FromInteger(11),
		},
		{
			name: "curse levels into a skill points put there", planterPoints: true, merge: grantmerge.Ignore,
			amount: -fxp.One, level: fxp.FromInteger(11),
		},
	} {
		for _, bought := range []bool{false, true} {
			if bought && tc.planterPoints {
				// A bought skill has no planter, so the run would repeat that of the case with the same merge.
				continue
			}
			t.Run(fmt.Sprintf("%s bought=%t", tc.name, bought), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				var brawling *Skill
				if bought {
					brawling = addSkill(e, "Brawling", fxp.Four)
				} else {
					adjust, amount := grantadjust.SetLevel, fxp.Twelve
					if tc.planterPoints {
						adjust, amount = grantadjust.AddPoints, fxp.Four
					}
					addTraitWithFeatures(e, "Racial",
						newSkillGrantOf(newEasySkill("Brawling", fxp.One), adjust, amount, grantmerge.Ignore))
				}
				addTraitWithFeatures(e, "Training",
					newSkillGrantOf(newEasySkill("Brawling", fxp.One), tc.adjust, tc.amount, tc.merge))
				e.Recalculate()
				if !bought {
					brawling = grantedOf(e.Skills)[0]
				}
				c.Equal(tc.level, brawling.LevelData.Level,
					"a merge into a granted skill reads the level the points its grant gives it buy")
			})
		}
	}
}

// TestGrantSkillAdjustmentKinds checks what each kind of adjustment gives an IQ/Hard skill at IQ 10, standing alone and
// merged into the skill bought with the points of its first level, at IQ-2, or for a wildcard skill IQ-3, or for the
// Hard technique Disarming on Karate 10, up to 5 above it, Karate+1. The grants are Keep Highest, which only a level
// set heeds.
func TestGrantSkillAdjustmentKinds(t *testing.T) {
	for _, tc := range []struct {
		name      string
		adjust    grantadjust.Kind
		amount    fxp.Int
		wildcard  bool
		technique bool
		// alone is the points and level the skill has when the grant puts it on the sheet, and merged when the grant
		// merges into the skill bought.
		alonePoints, aloneLevel, mergedPoints, mergedLevel fxp.Int
	}{
		{
			name: "1 level", adjust: grantadjust.AddLevels, amount: fxp.One,
			alonePoints: fxp.One, aloneLevel: fxp.Eight, mergedPoints: fxp.One, mergedLevel: fxp.Nine,
		},
		{
			name: "2 levels", adjust: grantadjust.AddLevels, amount: fxp.Two,
			alonePoints: fxp.Two, aloneLevel: fxp.Nine, mergedPoints: fxp.One, mergedLevel: fxp.Ten,
		},
		{
			name: "1 point", adjust: grantadjust.AddPoints, amount: fxp.One,
			alonePoints: fxp.One, aloneLevel: fxp.Eight, mergedPoints: fxp.Two, mergedLevel: fxp.Nine,
		},
		{
			name: "2 points", adjust: grantadjust.AddPoints, amount: fxp.Two,
			alonePoints: fxp.Two, aloneLevel: fxp.Nine, mergedPoints: fxp.Three, mergedLevel: fxp.Nine,
		},
		{
			// 3 points buy what 2 do, while the 4 points of the third level buy IQ+0.
			name: "3 points", adjust: grantadjust.AddPoints, amount: fxp.Three,
			alonePoints: fxp.Three, aloneLevel: fxp.Nine, mergedPoints: fxp.Four, mergedLevel: fxp.Ten,
		},
		{
			// A level set gives a skill already present no points, but keeps it at that level when higher.
			name: "level 10", adjust: grantadjust.SetLevel, amount: fxp.Ten,
			alonePoints: fxp.Four, aloneLevel: fxp.Ten, mergedPoints: fxp.One, mergedLevel: fxp.Ten,
		},
		{
			// A wildcard skill costs three times the points a level, which prices the levels added, while the
			// points added are given as they are.
			name: "wildcard 1 level", adjust: grantadjust.AddLevels, amount: fxp.One, wildcard: true,
			alonePoints: fxp.Three, aloneLevel: fxp.Seven, mergedPoints: fxp.Three, mergedLevel: fxp.Eight,
		},
		{
			name: "wildcard 6 points", adjust: grantadjust.AddPoints, amount: fxp.Six, wildcard: true,
			alonePoints: fxp.Six, aloneLevel: fxp.Eight, mergedPoints: fxp.Nine, mergedLevel: fxp.Eight,
		},
		{
			// A technique's levels count from its default, a point each, plus 1 for a Hard one.
			name: "technique 1 level", adjust: grantadjust.AddLevels, amount: fxp.One, technique: true,
			alonePoints: fxp.Two, aloneLevel: fxp.Eleven, mergedPoints: fxp.Two, mergedLevel: fxp.Twelve,
		},
		{
			name: "technique 2 points", adjust: grantadjust.AddPoints, amount: fxp.Two, technique: true,
			alonePoints: fxp.Two, aloneLevel: fxp.Eleven, mergedPoints: fxp.Four, mergedLevel: fxp.Thirteen,
		},
		{
			// A fraction of a point buys a technique nothing, as it buys a skill nothing.
			name: "technique 3.5 points", adjust: grantadjust.AddPoints, amount: fxp.ThreeAndAHalf, technique: true,
			alonePoints: fxp.ThreeAndAHalf, aloneLevel: fxp.Twelve, mergedPoints: fxp.FromStringForced("5.5"),
			mergedLevel: fxp.FromInteger(14),
		},
		{
			// A level set is priced from the skill the technique rests on.
			name: "technique level 12", adjust: grantadjust.SetLevel, amount: fxp.Twelve, technique: true,
			alonePoints: fxp.Three, aloneLevel: fxp.Twelve, mergedPoints: fxp.Two, mergedLevel: fxp.Twelve,
		},
		{
			// Neither the points nor the levels added take a technique above its limit.
			name: "technique 6 levels", adjust: grantadjust.AddLevels, amount: fxp.Six, technique: true,
			alonePoints: fxp.Seven, aloneLevel: fxp.Fifteen, mergedPoints: fxp.Two, mergedLevel: fxp.Fifteen,
		},
		{
			// A level set above the limit is priced at the limit.
			name: "technique level 16", adjust: grantadjust.SetLevel, amount: fxp.Sixteen, technique: true,
			alonePoints: fxp.Six, aloneLevel: fxp.Fifteen, mergedPoints: fxp.Two, mergedLevel: fxp.Fifteen,
		},
	} {
		for _, merged := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s merged=%t", tc.name, merged), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				item := newHardSkill("Alchemy", 0)
				first := fxp.One
				switch {
				case tc.wildcard:
					item.Difficulty.Difficulty = difficulty.Wildcard
					first = fxp.Three
				case tc.technique:
					addHardSkill(e, "Karate", fxp.Four)
					item = newDisarming(0)
					first = fxp.Two
				}
				var skill *Skill
				switch {
				case merged && tc.technique:
					skill = newDisarming(first)
					skill.SetDataOwner(e)
					e.Skills = append(e.Skills, skill)
				case merged:
					skill = addHardSkill(e, "Alchemy", first)
					skill.Difficulty.Difficulty = item.Difficulty.Difficulty
				}
				addTraitWithFeatures(e, "Training", newSkillGrantOf(item, tc.adjust, tc.amount, grantmerge.KeepHighest))
				e.Recalculate()
				points, level := tc.alonePoints, tc.aloneLevel
				granted := grantedOf(e.Skills)
				if merged {
					c.Equal(0, len(granted), "the grant merges into the skill bought")
					points, level = tc.mergedPoints, tc.mergedLevel
				} else {
					if len(granted) != 1 {
						c.Fatalf("expected the grant to put the skill on the sheet, got %d granted", len(granted))
					}
					skill = granted[0]
				}
				c.Equal(points, skill.AdjustedPoints(nil))
				c.Equal(level, skill.LevelData.Level)
			})
		}
	}
}

// TestGrantSetLevelBelowWhatAPointBuys checks that a level set below what 1 point buys still puts its row on the sheet,
// since the rows are decided by the sheet's data alone, but leaves it out of effect until the point buys it.
func TestGrantSetLevelBelowWhatAPointBuys(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	alchemy := newHardSkill("Alchemy", 0)
	alchemy.Features = Features{newSkillBonusTo("Brawling", fxp.Two)}
	training := addTraitWithFeatures(e, "Training",
		newSkillGrantOf(alchemy, grantadjust.SetLevel, fxp.Seven, grantmerge.Ignore))
	brawling := addSkill(e, "Brawling", fxp.One)
	e.Recalculate()
	c.False(e.Unsettled())
	granted := grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected the row on the sheet, got %d granted", len(granted))
	}
	row := granted[0]
	c.False(row.Enabled(), "at IQ 10, the first point buys IQ-2, or 8, so the grant gives the row nothing")
	c.Equal(fxp.Int(0), row.AdjustedPoints(nil))
	c.Equal(fxp.Ten, brawling.LevelData.Level, "nor do its features take effect")
	var data CellData
	row.CellData(SkillDescriptionColumn, &data)
	c.True(data.Dim)
	c.True(data.Disabled, "the row is dimmed and struck through, as for a grant not in effect")
	c.Equal("Granted by Training\nThe grant gives it nothing, since the level it sets is below what 1 point buys.",
		data.GrantedBy)
	c.Equal("Does not grant Alchemy at skill level 7, since that is below what 1 point buys", e.grantSummary(training))
	e.Recalculate()
	c.False(e.DerivedStateChanged(), "a recalculation that changes nothing leaves the state as it found it")

	e.Attributes.Set[IntelligenceID].Adjustment = -fxp.One
	e.Recalculate()
	c.True(row == grantedOf(e.Skills)[0], "at IQ 9, the first point buys level 7, and the same row comes into effect")
	c.True(row.Enabled())
	c.Equal(fxp.One, row.AdjustedPoints(nil))
	c.Equal(fxp.Seven, row.LevelData.Level)
	c.Equal(fxp.Twelve, brawling.LevelData.Level, "along with its features")
	c.Equal("Grants Alchemy at skill level 7", e.grantSummary(training))

	// Nor does such a grant stand in for one not in effect, whether or not it is in effect itself. The goggles' level
	// set, under Ignore, ranks ahead of those for putting the row on the sheet.
	e = NewEntity()
	addCarriedEquipmentWithFeatures(e, "Goggles",
		newSkillGrantOf(newHardSkill("Alchemy", 0), grantadjust.SetLevel, fxp.Ten, grantmerge.Ignore)).Equipped = false
	dabbler := addTraitWithFeatures(e, "Dabbler",
		newSkillGrantOf(newHardSkill("Alchemy", 0), grantadjust.SetLevel, fxp.Seven, grantmerge.KeepHighest))
	kit := addCarriedEquipmentWithFeatures(e, "Kit",
		newSkillGrantOf(newHardSkill("Alchemy", 0), grantadjust.SetLevel, fxp.Seven, grantmerge.KeepHighest))
	kit.Equipped = false
	e.Recalculate()
	granted = grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected the row the goggles put on the sheet, got %d granted", len(granted))
	}
	c.False(granted[0].Enabled(), "the row stays out of effect with the goggles' grant")
	const unchanged = "Does not change Alchemy, already present, since skill level 7 is below what 1 point buys"
	c.Equal(unchanged, e.grantSummary(dabbler))
	c.Equal(unchanged, e.grantSummary(kit))

	// A grant that puts the row on the sheet says it gives it nothing whether or not it is in effect.
	e = NewEntity()
	kit = addCarriedEquipmentWithFeatures(e, "Kit",
		newSkillGrantOf(newHardSkill("Alchemy", 0), grantadjust.SetLevel, fxp.Seven, grantmerge.Ignore))
	kit.Equipped = false
	e.Recalculate()
	c.Equal("Does not grant Alchemy at skill level 7, since that is below what 1 point buys", e.grantSummary(kit))

	// A level set gives nothing either when the sheet lacks what the item's level rests on, or when it is below what
	// the points of the item's first level buy, which the texts say: 3 points buy a wildcard skill IQ-3, 7 at IQ 10,
	// and 2 points buy the Hard technique Kicking 1 level above its default, Karate-2, 8 with Karate at 10.
	sane := func() *Skill {
		s := newHardSkill("Alchemy", 0)
		s.Difficulty.Attribute = "sanity"
		return s
	}
	wildcard := func() *Skill {
		s := newHardSkill("Alchemy", 0)
		s.Difficulty.Difficulty = difficulty.Wildcard
		return s
	}
	kicking := func() *Skill { return newKicking(0) }
	saneKicking := func() *Skill {
		s := newKicking(0)
		s.TechniqueDefault = &SkillDefault{DefaultType: "sanity", Modifier: -fxp.Two}
		return s
	}
	for _, tc := range []struct {
		item   func() *Skill
		karate bool
		level  fxp.Int
		// grantedBy is the row's text, and planted and unchanged are the summaries of the grant that put it on the
		// sheet and of one merging into it.
		grantedBy, planted, unchanged string
	}{
		{
			item: sane, level: fxp.Twelve,
			grantedBy: "Granted by Training\nThe grant gives it nothing, since the sheet lacks the attribute its level " +
				"rests on.",
			planted:   "Does not grant Alchemy at skill level 12, since the sheet lacks the attribute its level rests on",
			unchanged: "Does not change Alchemy, already present, since the sheet lacks the attribute its level rests on",
		},
		{
			item: wildcard, level: fxp.Six,
			grantedBy: "Granted by Training\nThe grant gives it nothing, since the level it sets is below what 3 points " +
				"buy.",
			planted:   "Does not grant Alchemy at skill level 6, since that is below what 3 points buy",
			unchanged: "Does not change Alchemy, already present, since skill level 6 is below what 3 points buy",
		},
		{
			item: kicking, karate: true, level: fxp.Eight,
			grantedBy: "Granted by Training\nThe grant gives it nothing, since the level it sets is below what 2 points " +
				"buy.",
			planted:   "Does not grant Kicking at skill level 8, since that is below what 2 points buy",
			unchanged: "Does not change Kicking, already present, since skill level 8 is below what 2 points buy",
		},
		{
			item: kicking, level: fxp.Twelve,
			grantedBy: "Granted by Training\nThe grant gives it nothing, since the sheet lacks the skill its level rests " +
				"on.",
			planted:   "Does not grant Kicking at skill level 12, since the sheet lacks the skill its level rests on",
			unchanged: "Does not change Kicking, already present, since the sheet lacks the skill its level rests on",
		},
		{
			// A technique may rest on an attribute rather than a skill.
			item: saneKicking, level: fxp.Twelve,
			grantedBy: "Granted by Training\nThe grant gives it nothing, since the sheet lacks the attribute its level " +
				"rests on.",
			planted:   "Does not grant Kicking at skill level 12, since the sheet lacks the attribute its level rests on",
			unchanged: "Does not change Kicking, already present, since the sheet lacks the attribute its level rests on",
		},
	} {
		e = NewEntity()
		if tc.karate {
			addHardSkill(e, "Karate", fxp.Four)
		}
		training = addTraitWithFeatures(e, "Training",
			newSkillGrantOf(tc.item(), grantadjust.SetLevel, tc.level, grantmerge.Ignore))
		dabbler = addTraitWithFeatures(e, "Dabbler",
			newSkillGrantOf(tc.item(), grantadjust.SetLevel, tc.level, grantmerge.KeepHighest))
		e.Recalculate()
		granted = grantedOf(e.Skills)
		if len(granted) != 1 {
			c.Fatalf("expected the row on the sheet, %s, got %d granted", tc.planted, len(granted))
		}
		granted[0].CellData(SkillDescriptionColumn, &data)
		c.Equal(tc.grantedBy, data.GrantedBy)
		c.Equal(tc.planted, e.grantSummary(training))
		c.Equal(tc.unchanged, e.grantSummary(dabbler))
	}
}

func TestGrantStateAppliesPointsExcept(t *testing.T) {
	c := check.New(t)
	a, b := NewTrait(nil, nil, false), NewTrait(nil, nil, false)
	a.Name, b.Name = "A", "B"
	var state grantState
	state.addPoints(grantSource{owner: a}, fxp.Four, nil)
	state.addPoints(grantSource{owner: b}, fxp.One, nil)
	var tooltip xbytes.InsertBuffer
	c.Equal(fxp.Two, state.applyPoints(a, fxp.One, &tooltip))
	c.Equal("\nB [+1]", tooltip.String(), "the tooltip lists only the points that were added")
	tooltip.Reset()
	c.Equal(fxp.Six, state.applyPoints(nil, fxp.One, &tooltip))
	c.Equal("\nA [+4]\nB [+1]", tooltip.String())
}

func TestGrantStateAppliesLevelNotesInOrder(t *testing.T) {
	c := check.New(t)
	a, b, keeper := NewTrait(nil, nil, false), NewTrait(nil, nil, false), NewTrait(nil, nil, false)
	a.Name, b.Name, keeper.Name = "A", "B", "Keeper"
	var state grantState
	state.keepLevel(grantSource{owner: a}, fxp.Five, false)
	state.addLevels(grantSource{owner: b}, fxp.One)
	state.keepLevel(grantSource{owner: keeper}, fxp.Four, true)
	var tooltip xbytes.InsertBuffer
	c.Equal(fxp.Four, state.applyLevelNotes(nil, fxp.Three, &tooltip, nil))
	c.Equal("\nB [+1]\nA [+1]\nKeeper [-1]", tooltip.String(),
		"the levels added come first, and then the levels kept, in the order the grants were merged")
	c.Equal(fxp.Five, state.applyLevelNotes(keeper, fxp.Three, nil, nil),
		"what a Keep Lowest grant caps is left out for the prerequisites of what carries it")
}

func TestGrantStateHashCoversWhatTheGrantsGave(t *testing.T) {
	c := check.New(t)
	newState := func() *grantState {
		var state grantState
		state.keepLevel(grantSource{owner: NewTrait(nil, nil, false)}, fxp.Five, false)
		return &state
	}
	base := Hash64(hashOf(newState().hash))
	lower := newState()
	lower.levelNotes[0].lower = true
	c.NotEqual(base, Hash64(hashOf(lower.hash)), "a level kept as the lower of the two is another state")
	unreached := newState()
	unreached.unreached = true
	c.NotEqual(base, Hash64(hashOf(unreached.hash)), "as is a grant giving its item nothing")
	unresolved := newState()
	unresolved.unreached, unresolved.unresolved = true, true
	c.NotEqual(Hash64(hashOf(unreached.hash)), Hash64(hashOf(unresolved.hash)), "and why it gives nothing")
}

func TestGrantStateCarriesOverToNextRecalculation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore),
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), brawling.LevelData.Level)
	c.Equal(1, e.recalculate(maxRecalculationPassesInAll), "an unchanged sheet settles in one pass")
	c.Equal(fxp.FromInteger(13), brawling.LevelData.Level)

	training.Features = nil
	e.Recalculate()
	c.Equal(fxp.Ten, brawling.LevelData.Level, "a grant that is gone takes back what it merged")
}

func TestGrantTraitMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name string
		// adds is set for a grant that adds its levels rather than setting the trait's level as merge says.
		adds     bool
		merge    grantmerge.Mode
		perLevel fxp.Int
		have     fxp.Int
		// bonus is a trait bonus to the trait's level from elsewhere on the sheet.
		bonus fxp.Int
		grant fxp.Int
		level fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.One},
		{name: "add levels", adds: true, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.Three},
		{name: "curse", adds: true, perLevel: fxp.One, have: fxp.Three, grant: -fxp.One, level: fxp.Two},
		{name: "curse past level 0", adds: true, perLevel: fxp.One, have: fxp.Three, grant: -fxp.Five, level: 0},
		{name: "keep highest", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.One, grant: fxp.Three, level: fxp.Three},
		{name: "keep highest lower", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.Three},
		{name: "keep highest with bonus", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.One, bonus: fxp.Two, grant: fxp.Four, level: fxp.Six},
		{name: "keep lowest", merge: grantmerge.KeepLowest, perLevel: fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.One},
		{name: "keep lowest higher", merge: grantmerge.KeepLowest, perLevel: fxp.One, have: fxp.One, grant: fxp.Three, level: fxp.One},
		// Keep Highest and Keep Lowest compare the levels as they are, even for a disadvantage.
		{name: "keep highest disadvantage", merge: grantmerge.KeepHighest, perLevel: -fxp.One, have: fxp.One, grant: fxp.Three, level: fxp.Three},
		{name: "keep lowest disadvantage", merge: grantmerge.KeepLowest, perLevel: -fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.One},
		// Below level 1 a grant gives nothing, as it says of itself; only a negative level takes away.
		{name: "add half a level", adds: true, perLevel: fxp.One, have: fxp.Two, grant: fxp.Half, level: fxp.Two},
		{name: "keep lowest at a negative level", merge: grantmerge.KeepLowest, perLevel: fxp.One, have: fxp.Three, grant: -fxp.One, level: fxp.Three},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			bad := addTrait(e, newGrantLeveledTrait("Bad Sight", tc.perLevel))
			bad.Levels = tc.have
			if tc.bonus != 0 {
				bonus := NewTraitBonus()
				bonus.NameCriteria.Qualifier = "Bad Sight"
				bonus.Amount = tc.bonus
				addTraitWithFeatures(e, "Boost", bonus)
			}
			g := newTraitGrant(newGrantLeveledTrait("Bad Sight", tc.perLevel), tc.grant, tc.merge)
			if tc.adds {
				g = addingLevels(g)
			}
			addTraitWithFeatures(e, "Racial", g)
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Traits)), "a grant whose trait is already present adds no row")
			c.Equal(tc.level, bad.CurrentLevel())
			c.Equal(tc.have.Mul(tc.perLevel), bad.AdjustedPoints(nil), "merged levels cost nothing")
		})
	}
}

func TestGrantTraitWithModifierIsAnotherItem(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	dr := addTrait(e, newGrantLeveledTrait("Damage Resistance", fxp.Five))
	dr.Levels = fxp.Three
	toughSkin := newGrantLeveledTrait("Damage Resistance", fxp.Five)
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "Tough Skin"
	toughSkin.Modifiers = []*TraitModifier{mod}
	addTraitWithFeatures(e, "Racial", addingLevels(newTraitGrant(toughSkin, fxp.Two, grantmerge.Ignore)))
	e.Recalculate()
	granted := grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected the trait with a modifier to be another item than the plain trait, got %d granted", len(granted))
	}
	c.Equal("Tough Skin", granted[0].Modifiers[0].Name)
	c.Equal(fxp.Three, dr.CurrentLevel(), "nothing is merged into the plain trait")

	drMod := NewTraitModifier(e, nil, false)
	drMod.Name = "Tough Skin"
	drMod.Disabled = true
	dr.Modifiers = []*TraitModifier{drMod}
	AttachModifiers(dr, dr.Modifiers)
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)), "a disabled modifier doesn't make it the same item")

	drMod.Disabled = false
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "with the same modifier in effect, the grant merges")
	c.Equal(fxp.Five, dr.CurrentLevel())
}

func TestGrantMatchesItemCarryingGrants(t *testing.T) {
	for _, tc := range []struct {
		name string
		// carrying gives the library's Magery the grant it carries, on the trait itself or on a modifier of it.
		carrying func(magery *Trait, g *Grant)
	}{
		{name: "on the trait", carrying: func(magery *Trait, g *Grant) {
			magery.Features = append(magery.Features, g)
		}},
		{name: "on a modifier", carrying: func(magery *Trait, g *Grant) {
			mod := NewTraitModifier(nil, nil, false)
			mod.Name = "Dark Aspected"
			mod.Features = Features{g}
			magery.Modifiers = []*TraitModifier{mod}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			newMagery := func() *Trait {
				magery := newGrantLeveledTrait("Magery", fxp.Ten)
				magery.Features = Features{newSkillBonusTo("Thaumatology", fxp.One)}
				tc.carrying(magery, newSpellGrant(newHardSpell("Light", fxp.One), fxp.One, grantmerge.Ignore))
				return magery
			}
			magery := addTrait(e, newMagery())
			// The grant's copy carries no grants, as SetItem leaves them out.
			g := addingLevels(newTraitGrant(newMagery(), fxp.One, grantmerge.Ignore))
			DropGrants(g.Trait)
			addTraitWithFeatures(e, "Elf", g)
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Traits)), "the grant merges into the Magery bought, grants and all")
			c.Equal(fxp.Two, magery.CurrentLevel())
		})
	}
}

func TestGrantDisabledTraitCountsAsAbsent(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTrait(e, newGrantLeveledTrait("Night Vision", fxp.One)).SetDisabled(true)
	addTraitWithFeatures(e, "Racial",
		addingLevels(newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)))
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)))
}

func TestGrantChangeRefreshesItem(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	grant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
	addTraitWithFeatures(e, "Racial", grant)
	e.Recalculate()
	before := grantedOf(e.Traits)[0]
	before.SwitchedOn = true
	grant.Amount = fxp.Three
	e.Recalculate()
	after := grantedOf(e.Traits)[0]
	c.False(before == after)
	c.Equal(before.TID, after.TID, "the refreshed item keeps its ID")
	c.True(after.SwitchedOn, "and its switch")
	c.Equal(fxp.Three, after.CurrentLevel())

	after.SetDisabled(true)
	grant.Amount = fxp.Four
	e.Recalculate()
	refreshed := grantedOf(e.Traits)[0]
	c.False(after == refreshed)
	c.True(refreshed.IsDisabled(), "the refreshed item stays disabled when the user disabled it")
	c.Equal(fxp.Four, refreshed.Levels)
}

// TestGrantRecordsClosedStatesOnlyForRowsItKeeps checks that the items a grant makes from its copy of an item whose
// notes are collapsed, for which a clone records that state (see PropagateNodeNoteClosedState), record nothing in the
// settings when they are thrown away: when the grant merges into the item already on the sheet, when its row is made
// afresh in place of the old one, and when an editor previews edits to the item it merges into. The row a grant puts on
// the sheet keeps the state.
func TestGrantRecordsClosedStatesOnlyForRowsItKeeps(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings()
	count := func() int {
		closedLock.Lock()
		defer closedLock.Unlock()
		return len(settings.Closed)
	}
	e := NewEntity()
	guns := addSkill(e, "Guns", fxp.One)
	merging := newEasySkill("Guns", fxp.One)
	planting := newEasySkill("Brawling", fxp.One)
	keys := []string{"N:" + string(merging.TID), "N:" + string(planting.TID)}
	for _, key := range keys {
		SetClosedState(key, true)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			SetClosedState(key, false)
		}
	})
	grant := newSkillGrant(planting, fxp.One, grantmerge.Ignore)
	addTraitWithFeatures(e, "Racial", newSkillGrant(merging, fxp.One, grantmerge.Ignore), grant)
	before := count()
	e.Recalculate()
	rows := grantedOf(e.Skills)
	if len(rows) != 1 {
		c.Fatalf("expected one granted row, got %d", len(rows))
	}
	rowKey := "N:" + string(rows[0].TID)
	t.Cleanup(func() { SetClosedState(rowKey, false) })
	c.True(IsClosed(rowKey), "the row a grant puts on the sheet keeps its copy's collapsed notes")
	c.Equal(before+1, count(), "and is the only item made that records it")
	e.Recalculate()
	c.Equal(before+1, count(), "a recalculation records nothing more")
	grant.Amount = fxp.Two
	e.Recalculate()
	c.Equal(rowKey, "N:"+string(grantedOf(e.Skills)[0].TID), "the row made afresh keeps the old row's ID")
	c.Equal(before+1, count(), "and records nothing under its own")
	var data SkillEditData
	data.CopyFrom(guns)
	data.Specialization = "Pistol"
	guns.PointsWithGrants(&data, guns.Points)
	c.Equal(before+1, count(), "previewing edits to the item a grant merges into records nothing")
}

// TestGrantsSharingAKeyEachKeepTheirRow checks that when a carrier holds two grants of one item ID, as a hand-edited
// sheet can, and each has put a row on the sheet, a change to one grant refreshes its own row and leaves the other
// grant its row, rather than the other grant claiming the row the refresh took out of the list.
func TestGrantsSharingAKeyEachKeepTheirRow(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	cheap := newGrantLeveledTrait("Night Vision", fxp.One)
	dear := newGrantLeveledTrait("Night Vision", fxp.Two)
	dear.TID = cheap.TID
	first := newTraitGrant(cheap, fxp.One, grantmerge.Ignore)
	second := newTraitGrant(dear, fxp.One, grantmerge.Ignore)
	addTraitWithFeatures(e, "Racial", first, second)
	e.Recalculate()
	granted := grantedOf(e.Traits)
	if len(granted) != 2 {
		c.Fatalf("expected a row for each grant, got %d", len(granted))
	}
	firstRow, secondRow := granted[0], granted[1]
	first.Amount = fxp.Two
	e.Recalculate()
	granted = grantedOf(e.Traits)
	c.Equal(2, len(granted), "each grant keeps a row of its own")
	if len(granted) != 2 {
		return
	}
	c.Equal(firstRow.TID, granted[0].TID, "the changed grant's row is refreshed in place")
	c.Equal(fxp.Two, granted[0].CurrentLevel())
	c.True(granted[1] == secondRow, "the other grant keeps its own row")
}

func TestGrantSurvivesSaveAndLoad(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Profile.TechLevel = "8"
	guns := newEasySkill("Guns", fxp.One)
	guns.TechLevel = new(string)
	skillGrant := newSkillGrantOf(guns, grantadjust.SetLevel, fxp.Eleven, grantmerge.KeepLowest)
	traitGrant := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
	addTraitWithFeatures(e, "Racial", traitGrant, skillGrant)
	e.Recalculate()
	beforeTrait, beforeSkill := grantedOf(e.Traits)[0], grantedOf(e.Skills)[0]
	data, err := json.Marshal(e)
	c.NoError(err)
	saved, err := json.Marshal(skillGrant)
	c.NoError(err)
	c.Contains(string(saved), `"amount":11,"adjust":"set_level","merge":"keep_lowest",`)
	c.NotContains(string(saved), `"level":`, "a grant saves its amount, not a level")
	saved, err = json.Marshal(traitGrant)
	c.NoError(err)
	c.Contains(string(saved), `"amount":2,"adjust":"set_level",`,
		"a trait grant saves that it sets the level, since a kind left out loads as adding levels")
	var unknown Grant
	c.NoError(json.Unmarshal([]byte(strings.Replace(string(saved), `"set_level"`, `"set_levels"`, 1)), &unknown))
	c.Equal(grantadjust.SetLevel, unknown.EffectiveAdjust(), "a kind unknown to the trait grant falls back")
	var loaded Entity
	c.NoError(json.Unmarshal(data, &loaded))
	traits, skills := grantedOf(loaded.Traits), grantedOf(loaded.Skills)
	if len(traits) != 1 || len(skills) != 1 {
		c.Fatalf("expected one granted trait and one granted skill, got %d and %d", len(traits), len(skills))
	}
	// Loading recalculates, and a row made again keeps its ID, so only the link's hash would tell.
	c.Equal(beforeTrait.TID, traits[0].TID)
	c.Equal(*beforeTrait.Granted, *traits[0].Granted, "the loaded trait is the one saved, not made again")
	c.Equal(beforeSkill.TID, skills[0].TID)
	c.Equal(*beforeSkill.Granted, *skills[0].Granted, "the loaded skill is the one saved, empty tech level and all")
	c.Equal("8", skills[0].TL())
	loaded.Recalculate()
	c.True(traits[0] == grantedOf(loaded.Traits)[0], "the loaded item is kept, not made again, which keeps its ID anyway")
	c.True(skills[0] == grantedOf(loaded.Skills)[0])
	c.Equal(fxp.Two, traits[0].CurrentLevel())
	c.Equal(fxp.Eleven, skills[0].LevelData.Level)
	c.Equal(len(e.Traits), len(loaded.Traits))
}

func TestGrantedSkillFollowsTheSheetsTechLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Profile.TechLevel = "3"
	guns := newEasySkill("Guns", fxp.One)
	guns.TechLevel = new(string)
	addTraitWithFeatures(e, "Training", newSkillGrant(guns, fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	granted := grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected one granted skill, got %d", len(granted))
	}
	row := granted[0]
	c.Equal("3", row.TL())
	e.Profile.TechLevel = "8"
	e.Recalculate()
	granted = grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected one granted skill, got %d", len(granted))
	}
	c.Equal("8", granted[0].TL(), "a grant leaving the tech level empty gives the sheet's")
	c.Equal(row.TID, granted[0].TID)

	bought := addSkill(e, "Guns", fxp.One)
	bought.TechLevel = new("8")
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Skills)), "a copy bought at the sheet's tech level absorbs the grant")
}

func TestGrantedLinkStaysOnlyOnItsSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
	e.Recalculate()
	brawling := grantedOf(e.Skills)[0]
	c.NotNil(brawling.Clone(LibraryFile{}, e, nil, Copy).Granted, "a copy for the same sheet keeps the link")
	other := NewEntity()
	copied := brawling.Clone(LibraryFile{}, other, nil, Reference)
	c.Nil(copied.Granted, "a copy for another sheet has no grant to answer to there")
	c.Equal(fxp.Four, copied.Points, "so it is bought with the points its grant gave it")
	template := brawling.Clone(LibraryFile{}, nil, nil, Reference)
	c.Nil(template.Granted, "nor does one for a library or template")
	c.Equal(fxp.Four, template.Points)
	addTraitWithFeatures(e, "Magery", newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(fxp.Two, grantedOf(e.Spells)[0].Clone(LibraryFile{}, nil, nil, Reference).Points, "as is a spell's")
	amulet := addCarriedEquipmentWithFeatures(e, "Amulet",
		newSkillGrant(newEasySkill("Karate", fxp.One), fxp.Three, grantmerge.Ignore))
	amulet.Equipped = false
	e.Recalculate()
	karate := grantedOf(e.Skills)[1]
	c.True(karate.grantNotInEffect())
	c.Equal(fxp.Four, karate.Clone(LibraryFile{}, nil, nil, Reference).Points, "whether or not the grant is in effect")
	belt := addCarriedEquipmentWithFeatures(e, "Belt",
		newSkillGrant(newEasySkill("Karate", fxp.One), fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	c.True(GrantingItem(karate) == belt)
	c.Equal(fxp.Two, karate.Clone(LibraryFile{}, nil, nil, Reference).Points, "or of the grant standing in for it")
}

// newGrantNeedingBrawling returns a grant of the trait Night Vision, which needs the skill Brawling, checking that
// prerequisite when checkPrereqs is set.
func newGrantNeedingBrawling(checkPrereqs bool, merge grantmerge.Mode) *Grant {
	item := newGrantLeveledTrait("Night Vision", fxp.One)
	item.Prereq = NewPrereqList()
	addSkillPrereq(item.Prereq, "Brawling", fxp.One)
	g := newTraitGrant(item, fxp.Two, merge)
	g.CheckPrereqs = checkPrereqs
	return g
}

func TestGrantIgnoresPrereqsByDefault(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	addTraitWithFeatures(e, "Training", newGrantNeedingBrawling(false, grantmerge.Ignore))
	e.Recalculate()
	nightVision := grantedOf(e.Traits)[0]
	c.False(nightVision.grantNotInEffect())
	c.Equal("", nightVision.UnsatisfiedReason, "ignored prerequisites are not reported")
	c.False(nightVision.EffectivelyDisabled(), "ignored prerequisites don't disable the trait")

	e = NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newKicking(0), fxp.One, grantmerge.Ignore))
	e.Recalculate()
	granted := grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected the technique on the sheet, got %d granted", len(granted))
	}
	c.Equal("", granted[0].UnsatisfiedReason, "nor is a technique's need of the skill it rests on")
}

func TestGrantCheckingPrereqsLeavesThemToTheSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.EnforceTraitPrereqs = true
	addTraitWithFeatures(e, "Training", newGrantNeedingBrawling(true, grantmerge.Ignore))
	e.Recalculate()
	nightVision := grantedOf(e.Traits)[0]
	c.False(nightVision.grantNotInEffect(), "the grant is in effect whatever the prerequisites")
	c.NotEqual("", nightVision.UnsatisfiedReason, "the unmet prerequisites are reported")
	c.True(nightVision.EffectivelyDisabled(), "the sheet disables the trait as it would one bought")

	addSkill(e, "Brawling", fxp.One)
	e.Recalculate()
	c.Equal("", nightVision.UnsatisfiedReason, "the error goes once the prerequisites are met")
	c.False(nightVision.EffectivelyDisabled())

	e = NewEntity()
	kicking := newSkillGrant(newKicking(0), fxp.One, grantmerge.Ignore)
	kicking.CheckPrereqs = true
	addTraitWithFeatures(e, "Training", kicking)
	e.Recalculate()
	granted := grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected the technique on the sheet, got %d granted", len(granted))
	}
	c.Contains(granted[0].UnsatisfiedReason, "Requires a skill named Karate",
		"as is a technique's need of the skill it rests on")
}

func TestGrantGivesWayToTheSameItemAddedLater(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training",
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, fxp.One, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Skills)))

	addSkill(e, "Brawling", fxp.One)
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Skills)), "the granted row gives way to the skill the user added")
	brawling := e.Skills[len(e.Skills)-1]
	c.Equal(fxp.Two, brawling.AdjustedPoints(nil), "the grant merges into the skill the user added")
}

func TestGrantSummaryTellsWhatEachGrantDid(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	have := addTrait(e, newGrantLeveledTrait("Acute Vision", fxp.One))
	training := addTraitWithFeatures(e, "Training",
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore),
		newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.One, grantmerge.Ignore))
	goggles := addCarriedEquipmentWithFeatures(e, "Goggles",
		newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), fxp.One, grantmerge.Ignore))
	goggles.Equipped = false
	pack := NewTrait(nil, nil, true)
	pack.Name = "Pack"
	curse := addTraitWithFeatures(e, "Curse",
		newSkillGrant(newEasySkill("Karate", fxp.One), -fxp.One, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.SetLevel, -fxp.One, grantmerge.KeepHighest),
		addingLevels(newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), -fxp.One, grantmerge.Ignore)),
		newTraitGrant(newGrantLeveledTrait("Infravision", fxp.One), 0, grantmerge.Ignore),
		newTraitGrant(pack, fxp.One, grantmerge.Ignore))
	addSkill(e, "Brawling", fxp.One)
	visor := addCarriedEquipmentWithFeatures(e, "Visor",
		newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.One, grantmerge.Ignore),
		addingLevels(newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), -fxp.Two, grantmerge.Ignore)),
		addingLevels(newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.Two, grantmerge.Ignore)),
		newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.Three, grantmerge.KeepHighest),
		newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.Two, grantmerge.KeepLowest),
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.Two, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, fxp.Two, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve, grantmerge.KeepHighest),
		newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), fxp.One, grantmerge.Ignore))
	visor.Equipped = false
	e.Recalculate()
	summary := e.grantSummary(training)
	c.Contains(summary, "Grants Night Vision 2")
	c.Contains(summary, "Does not grant Acute Vision 1, since it is already present")
	c.Contains(e.grantSummary(goggles), "Would grant Dark Vision 1, but the grant is not in effect")
	c.Equal("", e.grantSummary(have), "an item carrying no grants has no summary")
	c.Equal(`Removes nothing from Karate, since it is not present
Removes nothing from Karate, since it is not present
Does not grant Karate at skill level -1, since a skill level below 1 grants nothing
Removes 1 level from Acute Vision, already present
Does not grant Infravision 0, since a level below 1 grants nothing
Does not grant Pack, since a container can't be granted`, e.grantSummary(curse),
		"a grant that does nothing has a line saying why, in the order the item carries its grants")
	c.Equal(`Does not grant Acute Vision 1, since it is already present
Would remove 2 levels from Acute Vision, already present, but the grant is not in effect
Would add 2 levels to Acute Vision, already present, but the grant is not in effect
Would keep the higher level of Acute Vision, as it is or at level 3, but the grant is not in effect
Would keep the lower level of Acute Vision, as it is or at level 2, but the grant is not in effect
Would remove 2 points from Brawling, already present, but the grant is not in effect
Would add 2 points to Brawling, already present, but the grant is not in effect
Would keep the higher level of Brawling, as it is or at skill level 12, but the grant is not in effect
Would grant Dark Vision 1, but the grant is not in effect`, e.grantSummary(visor),
		"a grant not in effect says what it would do, which for one that would stand in is grant the item")

	e = NewEntity()
	addSkill(e, "Karate", fxp.One)
	shades := addCarriedEquipmentWithFeatures(e, "Shades",
		newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), fxp.Two, grantmerge.Ignore))
	shades.Equipped = false
	cursed := addTraitWithFeatures(e, "Cursed",
		newSkillGrant(newEasySkill("Karate", fxp.One), -fxp.One, grantmerge.Ignore),
		newSkillGrant(newEasySkill("Karate", fxp.One), fxp.Two, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.SetLevel, fxp.Twelve, grantmerge.KeepHighest),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.SetLevel, fxp.Eleven, grantmerge.KeepLowest),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.AddPoints, -fxp.Two, grantmerge.Ignore),
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.AddPoints, fxp.Two, grantmerge.Ignore),
		newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), fxp.One, grantmerge.Ignore))
	cursed.Prereq = NewPrereqList()
	addSkillPrereq(cursed.Prereq, "Brawling", fxp.One)
	e.Recalculate()
	c.Equal(`Would remove 1 level from Karate, already present, but the prerequisites of what carries the grant are not met
Would add 2 levels to Karate, already present, but the prerequisites of what carries the grant are not met
Would keep the higher level of Karate, as it is or at skill level 12, but the prerequisites of what carries the grant are not met
Would keep the lower level of Karate, as it is or at skill level 11, but the prerequisites of what carries the grant are not met
Would remove 2 points from Karate, already present, but the prerequisites of what carries the grant are not met
Would add 2 points to Karate, already present, but the prerequisites of what carries the grant are not met
Does not grant Dark Vision 1, since the prerequisites of what carries the grant are not met`, e.grantSummary(cursed),
		"a grant whose carrier's prerequisites are unmet says what it would do, as one not in effect does")

	e = NewEntity()
	darkVision := func(level fxp.Int) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), level, grantmerge.Ignore)
	}
	goggles = addCarriedEquipmentWithFeatures(e, "Goggles", darkVision(fxp.Three))
	goggles.Equipped = false
	sight := addTraitWithFeatures(e, "Sight", darkVision(fxp.One))
	hawk := addTraitWithFeatures(e, "Hawk", darkVision(fxp.Two))
	hawk.Prereq = NewPrereqList()
	addSkillPrereq(hawk.Prereq, "Brawling", fxp.One)
	visor = addCarriedEquipmentWithFeatures(e, "Visor", darkVision(fxp.Two))
	visor.Equipped = false
	monocle := addCarriedEquipmentWithFeatures(e, "Monocle", darkVision(fxp.One))
	monocle.Equipped = false
	e.Recalculate()
	c.Equal("Grants Dark Vision 1, standing in for a grant of it that is not in effect", e.grantSummary(sight))
	c.Equal("Would grant Dark Vision 2, but the grant is not in effect", e.grantSummary(visor),
		"a grant not in effect that ranks above the one standing in would take its place")
	c.Equal("Does not grant Dark Vision 2, since the prerequisites of what carries the grant are not met",
		e.grantSummary(hawk), "as would one whose carrier's prerequisites are unmet")
	c.Equal("Does not grant Dark Vision 1, since it is already present", e.grantSummary(monocle),
		"while one ranking below it would not")

	e = NewEntity()
	karate := addSkill(e, "Karate", 0)
	jinxed := addTraitWithFeatures(e, "Jinxed",
		newSkillGrant(newEasySkill("Karate", fxp.One), -fxp.One, grantmerge.Ignore))
	charm := addCarriedEquipmentWithFeatures(e, "Charm",
		newSkillGrant(newEasySkill("Karate", fxp.One), -fxp.Two, grantmerge.Ignore))
	charm.Equipped = false
	hexed := addTraitWithFeatures(e, "Hexed",
		newSkillGrantOf(newEasySkill("Karate", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(fxp.Min, karate.LevelData.Level, "the test requires Karate to have no level")
	c.Equal("Removes nothing from Karate, since it has no level of its own", e.grantSummary(jinxed),
		"a grant taking levels from a skill with none removes nothing")
	c.Equal("Removes nothing from Karate, since it has no level of its own", e.grantSummary(charm),
		"whether or not it is in effect")
	c.Equal("Removes nothing from Karate, since it has no level of its own", e.grantSummary(hexed),
		"as does one taking points")
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Karate", fxp.One), fxp.Three, grantmerge.Ignore))
	e.Recalculate()
	c.Equal("Removes 1 level from Karate, already present", e.grantSummary(jinxed),
		"but takes from the level another grant gives it")
	c.Equal("Removes 1 point from Karate, already present", e.grantSummary(hexed))
	c.Equal(fxp.Ten, karate.LevelData.Level, "the 4 points of DX+2 less 1 buy DX+1, less 1 level")

	// A level set below what 1 point buys gives a skill already present nothing, whatever the merge mode: an Easy DX
	// skill's first point buys DX+0.
	e = NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	trainee := addTraitWithFeatures(e, "Trainee",
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Nine, grantmerge.KeepHighest),
		newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Nine, grantmerge.KeepLowest))
	e.Recalculate()
	c.Equal(`Does not change Brawling, already present, since skill level 9 is below what 1 point buys
Does not change Brawling, already present, since skill level 9 is below what 1 point buys`, e.grantSummary(trainee))
	c.Equal(fxp.One, brawling.AdjustedPoints(nil))
	c.Equal(fxp.Ten, brawling.LevelData.Level)
}

func TestGrantNeedsItsCarriersPrereqsMet(t *testing.T) {
	for _, tc := range []struct {
		name  string
		carry func(e *Entity, g *Grant) (any, **PrereqList)
	}{
		{name: "trait", carry: func(e *Entity, g *Grant) (any, **PrereqList) {
			t := addTraitWithFeatures(e, "Training", g)
			return t, &t.Prereq
		}},
		{name: "equipment", carry: func(e *Entity, g *Grant) (any, **PrereqList) {
			eqp := addCarriedEquipmentWithFeatures(e, "Goggles", g)
			return eqp, &eqp.Prereq
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			carrier, prereq := tc.carry(e, newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One,
				grantmerge.Ignore))
			*prereq = NewPrereqList()
			addSkillPrereq(*prereq, "Brawling", fxp.One)
			e.Recalculate()
			granted := grantedOf(e.Traits)
			if len(granted) != 1 {
				c.Fatalf("expected the trait on the sheet, not in effect, got %d granted", len(granted))
			}
			nightVision := granted[0]
			c.True(nightVision.grantNotInEffect(), "a grant whose carrier's prerequisites are unmet is not in effect")
			c.Contains(e.grantSummary(carrier), "Does not grant Night Vision 1")

			addSkill(e, "Brawling", fxp.One)
			e.Recalculate()
			c.True(nightVision == grantedOf(e.Traits)[0], "the trait stays the row it was once they are met")
			c.False(nightVision.grantNotInEffect())

			e.Skills = nil
			e.Recalculate()
			c.True(nightVision == grantedOf(e.Traits)[0], "and once they are no longer met")
			c.True(nightVision.grantNotInEffect())
			c.False(e.Unsettled())

			e.Recalculate()
			c.False(e.DerivedStateChanged(), "a recalculation that changes nothing leaves the state as it found it")
		})
	}
}

func TestGrantFromTraitAboveItsMaximumLevelStaysInEffect(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	gifted := newGrantLeveledTrait("Gifted", fxp.One)
	gifted.MaxLevels = "3"
	gifted.Levels = fxp.Four
	gifted.Features = Features{
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore),
		newSkillBonusTo("Brawling", fxp.Two),
	}
	addTrait(e, gifted)
	e.Recalculate()
	c.NotEqual("", gifted.UnsatisfiedReason, "the level exceeds the maximum")
	c.Equal(fxp.FromInteger(12), brawling.LevelData.Level, "the trait's other features stay in effect")
	c.False(grantedOf(e.Traits)[0].grantNotInEffect(), "and so does its grant")
	c.Equal("Grants Night Vision 1", e.grantSummary(gifted))
}

func TestGrantStandsInForOneNotInEffect(t *testing.T) {
	nightVision := func(level fxp.Int) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), level, grantmerge.Ignore)
	}
	badTemper := func(cr selfctrl.Roll) *Grant {
		item := NewTrait(nil, nil, false)
		item.Name = "Bad Temper"
		item.BasePoints = -fxp.Ten
		item.SelfControl = cr
		return newTraitGrant(item, fxp.One, grantmerge.Ignore)
	}
	guns := func(optionalSpecialization string) *Grant {
		item := newEasySkill("Guns", 0)
		item.Difficulty.Difficulty = difficulty.Average
		item.OptionalSpecialization = optionalSpecialization
		return newSkillGrant(item, fxp.Three, grantmerge.Ignore)
	}
	unequipped := func(e *Entity, name string, g *Grant) *Equipment {
		eqp := addCarriedEquipmentWithFeatures(e, name, g)
		eqp.Equipped = false
		return eqp
	}
	// alchemy is a grant of an IQ/Hard skill, whose first point buys IQ-2, or 8, which a level set of 7 doesn't reach.
	alchemy := func(adjust grantadjust.Kind, amount fxp.Int) *Grant {
		return newSkillGrantOf(newHardSkill("Alchemy", 0), adjust, amount, grantmerge.Ignore)
	}
	// unreached puts on the sheet a carrier of a grant of Alchemy that sets a level 7 and one that adds two levels to
	// it, the latter first when reversed is set, and returns them.
	unreached := func(reversed bool) func(e *Entity) (any, any) {
		return func(e *Entity) (any, any) {
			setter := addTraitWithFeatures(e, "Setter", alchemy(grantadjust.SetLevel, fxp.Seven))
			adder := addTraitWithFeatures(e, "Adder", alchemy(grantadjust.AddLevels, fxp.Two))
			if reversed {
				slices.Reverse(e.Traits)
			}
			return setter, adder
		}
	}
	// alchemyAt checks that the row is Alchemy at level 9, which the 2 points of the second level buy, and that it
	// says it stands in for a grant giving it nothing.
	alchemyAt := func(c check.Checker, e *Entity) {
		row := grantedOf(e.Skills)[0]
		c.Equal(fxp.Nine, row.LevelData.Level)
		c.Equal(fxp.Two, row.AdjustedPoints(nil))
		var data CellData
		row.CellData(SkillDescriptionColumn, &data)
		c.Equal("Granted by Adder, standing in for Setter, whose grant gives it nothing", data.GrantedBy)
	}
	// nightVisionAt checks that the row is Night Vision at the level, costing what it would if bought at that level.
	nightVisionAt := func(level fxp.Int) func(check.Checker, *Entity) {
		return func(c check.Checker, e *Entity) {
			row := grantedOf(e.Traits)[0]
			c.Equal(level, row.CurrentLevel())
			c.Equal(level, row.ClassifyingPoints(), "the row costs what the level in effect would if bought")
		}
	}
	for _, tc := range []struct {
		name string
		// carry puts two carriers of a grant of the same item on the sheet and returns them, first the one whose grant
		// doesn't give the item.
		carry func(e *Entity) (first, second any)
		// first is what the first carrier's summary says of its grant, and granted is the item the second's grant gives,
		// as the summaries name it.
		first, granted string
		// check checks that the row is the item the second carrier's grant gives.
		check func(c check.Checker, e *Entity)
	}{
		{
			name: "unequipped before equipped", carry: func(e *Entity) (any, any) {
				return unequipped(e, "Spare Goggles", nightVision(fxp.One)),
					addCarriedEquipmentWithFeatures(e, "Worn Goggles", nightVision(fxp.One))
			}, first: "Would grant Night Vision 1, but the grant is not in effect", granted: "Night Vision 1",
			check: nightVisionAt(fxp.One),
		},
		{
			name: "unmet carrier before met", carry: func(e *Entity) (any, any) {
				a := addTraitWithFeatures(e, "A", nightVision(fxp.One))
				a.Prereq = NewPrereqList()
				addSkillPrereq(a.Prereq, "Brawling", fxp.One)
				return a, addTraitWithFeatures(e, "B", nightVision(fxp.One))
			}, first: "Does not grant Night Vision 1, since the prerequisites of what carries the grant are not met",
			granted: "Night Vision 1", check: nightVisionAt(fxp.One),
		},
		{
			name: "higher level not in effect", carry: func(e *Entity) (any, any) {
				return unequipped(e, "Spare Goggles", nightVision(fxp.Three)),
					addCarriedEquipmentWithFeatures(e, "Worn Goggles", nightVision(fxp.One))
			}, first: "Would grant Night Vision 3, but the grant is not in effect", granted: "Night Vision 1",
			check: nightVisionAt(fxp.One),
		},
		{
			name: "self-control roll worth more not in effect", carry: func(e *Entity) (any, any) {
				grumpy := badTemper(selfctrl.CR12)
				grumpy.Trait.Frequency = frequency.FR9
				grumpy.Trait.UserDesc = "Snaps at everyone"
				grumpy.Trait.VTTNotes = "Roll before speaking"
				return unequipped(e, "Cursed Ring", badTemper(selfctrl.CR6)), addTraitWithFeatures(e, "Grumpy", grumpy)
			}, first: "Would grant Bad Temper, but the grant is not in effect", granted: "Bad Temper",
			check: func(c check.Checker, e *Entity) {
				row := grantedOf(e.Traits)[0]
				c.Equal(selfctrl.CR12, row.SelfControl)
				c.Equal(frequency.FR9, row.Frequency)
				c.Equal("Snaps at everyone", row.UserDesc)
				c.Equal("Roll before speaking", row.VTTNotes)
				c.Equal(-fxp.Ten, row.ClassifyingPoints(), "the row costs what the trait with its rolls would if bought")
			},
		},
		{
			name: "optional specialization not in effect", carry: func(e *Entity) (any, any) {
				holstered := guns("Pistol")
				holstered.Skill.VTTNotes = "Quick draw"
				return unequipped(e, "Spare Holster", guns("")),
					addCarriedEquipmentWithFeatures(e, "Worn Holster", holstered)
			}, first: "Would grant Guns at level 3, but the grant is not in effect", granted: "Guns (Pistol) at level 3",
			check: func(c check.Checker, e *Entity) {
				row := grantedOf(e.Skills)[0]
				c.Equal("Pistol", row.OptionalSpecialization)
				c.Equal("Quick draw", row.VTTNotes)
				c.Equal(fxp.FromInteger(12), row.LevelData.Level,
					"an optional specialization makes the Average skill Easy, so 4 points give DX+2")
			},
		},
		{
			name: "level set below what 1 point buys", carry: unreached(false),
			first:   "Does not grant Alchemy at skill level 7, since that is below what 1 point buys",
			granted: "Alchemy at level 2, standing in for a grant of it that gives it nothing", check: alchemyAt,
		},
		{
			name: "level set below what 1 point buys, listed second", carry: unreached(true),
			first:   "Does not grant Alchemy at skill level 7, since that is below what 1 point buys",
			granted: "Alchemy at level 2, standing in for a grant of it that gives it nothing", check: alchemyAt,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			first, second := tc.carry(e)
			e.Recalculate()
			c.False(e.Unsettled())
			traits, skills := grantedOf(e.Traits), grantedOf(e.Skills)
			rows := make([]interface{ Enabled() bool }, 0, len(traits)+len(skills))
			for _, one := range traits {
				rows = append(rows, one)
			}
			for _, one := range skills {
				rows = append(rows, one)
			}
			if len(rows) != 1 {
				c.Fatalf("expected 1 granted row, got %d", len(rows))
			}
			c.True(rows[0].Enabled(), "the grant in effect stands in for the one that isn't")
			c.True(GrantingItem(rows[0]) == second)
			c.Contains(e.grantSummary(second), "Grants "+tc.granted)
			c.Equal(tc.first, e.grantSummary(first))
			tc.check(c, e)
		})
	}
}

func TestGrantStandInFollowsTheGrantInEffect(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newGrant := func(level fxp.Int) *Grant {
		dexterous := newGrantLeveledTrait("Dexterous", fxp.Ten)
		bonus := NewAttributeBonus(DexterityID)
		bonus.PerLevel = true
		dexterous.Features = Features{bonus}
		return newTraitGrant(dexterous, level, grantmerge.Ignore)
	}
	spare := addCarriedEquipmentWithFeatures(e, "Spare Gloves", newGrant(fxp.Three))
	worn := addCarriedEquipmentWithFeatures(e, "Worn Gloves", newGrant(fxp.One))
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), e.ResolveAttributeCurrent(DexterityID))

	spare.Equipped = false
	e.Recalculate()
	c.Equal(fxp.FromInteger(11), e.ResolveAttributeCurrent(DexterityID), "the stand-in's level takes effect at once")

	worn.Equipped = false
	e.Recalculate()
	c.Equal(fxp.Ten, e.ResolveAttributeCurrent(DexterityID))
	c.Equal(fxp.Three, grantedOf(e.Traits)[0].Levels, "with no grant in effect, the row is what its own grant gives")
}

func TestGrantedRowInDisabledContainerCountsAsAbsent(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newGrant := func() *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)
	}
	a := addTraitWithFeatures(e, "A", newGrant())
	e.Recalculate()
	group := NewTrait(e, nil, true)
	group.SetDisabled(true)
	row := grantedOf(e.Traits)[0]
	row.SetParent(group)
	group.Children = []*Trait{row}
	e.Traits = []*Trait{a, group}
	b := addTraitWithFeatures(e, "B", newGrant())
	e.Recalculate()
	c.False(e.Unsettled())
	c.True(len(group.Children) == 1 && group.Children[0] == row, "A's row stays where the user put it")
	c.True(GrantingItem(row) == a)
	granted := grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected B's grant to put a row of its own on the sheet, since A's counts as absent, got %d rows",
			len(granted))
	}
	c.True(GrantingItem(granted[0]) == b)
	c.True(granted[0].Enabled())
	c.Equal("Grants Night Vision 1", e.grantSummary(a))
	c.Equal("Grants Night Vision 1", e.grantSummary(b))

	higher := addTraitWithFeatures(e, "C", newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two,
		grantmerge.Ignore))
	e.Recalculate()
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row at the top, got %d", len(granted))
	}
	c.True(GrantingItem(granted[0]) == higher, "a higher grant takes over the row in effect, not one the user disabled")
	c.True(granted[0].Enabled())
	c.Equal(fxp.Two, granted[0].CurrentLevel())
	c.True(len(group.Children) == 1 && group.Children[0] == row,
		"A's row stays where the user put it rather than A merging into the higher grant's row")
	c.True(GrantingItem(row) == a)

	e.Traits = slices.DeleteFunc(e.Traits, func(one *Trait) bool { return one == higher })
	e.Recalculate()
	c.True(len(group.Children) == 1 && group.Children[0] == row, "A's row stays where the user put it")
	c.True(GrantingItem(row) == a)
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row at the top, got %d", len(granted))
	}
	c.True(GrantingItem(granted[0]) == b, "the row at the top is B's, A's being where the user put it")
}

func TestGrantCarrierJudgedAfresh(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
	training.Prereq = NewPrereqList()
	addSkillPrereq(training.Prereq, "Karate", fxp.One)
	addSkill(e, "Karate", fxp.One)
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), brawling.LevelData.Level)

	e.Skills = e.Skills[:1]
	e.Recalculate()
	c.Equal(fxp.Ten, brawling.LevelData.Level, "a carrier whose prerequisites are no longer met is judged afresh")
	c.False(e.Unsettled())
	c.True(e.DerivedStateChanged())

	e.Recalculate()
	c.False(e.DerivedStateChanged(), "a recalculation that changes nothing leaves the state as it found it")
}

func TestGrantCarrierVerdictDoesNotDependOnHistory(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	needsY := addTraitWithFeatures(e, "Needs Y",
		newTraitGrant(newGrantLeveledTrait("X", fxp.One), fxp.One, grantmerge.Ignore))
	needsY.Prereq = newPrereqListForTrait("Y", true)
	needsX := addTraitWithFeatures(e, "Needs X")
	needsX.Prereq = newPrereqListForTrait("X", true)
	e.Recalculate()
	c.True(grantedOf(e.Traits)[0].grantNotInEffect(), "the carrier's prerequisites are unmet")

	needsX.Features = Features{newTraitGrant(newGrantLeveledTrait("Y", fxp.One), fxp.One, grantmerge.Ignore)}
	e.Recalculate()
	data, err := json.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(json.Unmarshal(data, &loaded))
	loaded.Recalculate()
	c.Equal(len(grantedOf(loaded.Traits)), len(grantedOf(e.Traits)), "an edit ends where loading the result does")
	c.Equal(2, len(grantedOf(e.Traits)))
}

func TestGrantFromDisabledItemPutsNothingOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	training := addTraitWithFeatures(e, "Training",
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore))
	goggles := addCarriedEquipmentWithFeatures(e, "Goggles",
		newTraitGrant(newGrantLeveledTrait("Dark Vision", fxp.One), fxp.One, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(2, len(grantedOf(e.Traits)))

	training.SetDisabled(true)
	goggles.SetDisabled(true)
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "a disabled item grants nothing, not even an inactive row")

	training.SetDisabled(false)
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)))
}

func TestGrantDoesNotMeetItsCarriersPrereqs(t *testing.T) {
	t.Run("item it puts on the sheet", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		training := addTraitWithFeatures(e, "Training",
			newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore))
		training.Prereq = NewPrereqList()
		addSkillPrereq(training.Prereq, "Brawling", fxp.One)
		addSkill(e, "Brawling", fxp.One)
		e.Recalculate()
		c.Equal("", training.UnsatisfiedReason, "the skill bought meets the prerequisite")
		c.Equal(0, len(grantedOf(e.Skills)), "the grant merges into the skill bought")

		e.Skills = nil
		e.Recalculate()
		c.NotEqual("", training.UnsatisfiedReason, "the skill it grants can't meet its own prerequisite")
		c.True(grantedOf(e.Skills)[0].grantNotInEffect(), "so the skill it grants is not in effect")
	})

	// Nor is the item seen by a script run for the prerequisites, whatever the script reads it through.
	brawlingGrant := func() *Grant {
		return newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore)
	}
	nightVisionGrant := func() *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)
	}
	lightGrant := func() *Grant { return newSpellGrant(newHardSpell("Light", fxp.One), fxp.One, grantmerge.Ignore) }
	clawsGrant := func() *Grant {
		claws := newGrantLeveledTrait("Claws", fxp.One)
		slash := NewWeapon(claws, true)
		slash.Usage = "Slash"
		claws.Weapons = []*Weapon{slash}
		return newTraitGrant(claws, fxp.One, grantmerge.Ignore)
	}
	// grantedNotInEffect returns true if a grant put exactly one item on the sheet and it is not in effect.
	grantedNotInEffect := func(e *Entity) bool {
		return len(grantedOf(e.Traits))+len(grantedOf(e.Skills))+len(grantedOf(e.Spells)) == 1 &&
			(slices.ContainsFunc(grantedOf(e.Traits), (*Trait).grantNotInEffect) ||
				slices.ContainsFunc(grantedOf(e.Skills), (*Skill).grantNotInEffect) ||
				slices.ContainsFunc(grantedOf(e.Spells), (*Spell).grantNotInEffect))
	}
	for _, tc := range []struct {
		name   string
		grant  func() *Grant
		script string
	}{
		{name: "skill level", grant: brawlingGrant, script: `entity.skillLevel("Brawling", "", false) > 0`},
		{name: "found skill", grant: brawlingGrant, script: `entity.findSkills("Brawling").length > 0`},
		{name: "listed skill", grant: brawlingGrant, script: `entity.skills.some(s => s.name == "Brawling")`},
		{name: "has trait", grant: nightVisionGrant, script: `entity.hasTrait("Night Vision")`},
		{name: "trait level", grant: nightVisionGrant, script: `entity.traitLevel("Night Vision") > 0`},
		{name: "found trait", grant: nightVisionGrant, script: `entity.findTraits("Night Vision").length > 0`},
		{name: "listed trait", grant: nightVisionGrant, script: `entity.traits.some(t => t.name == "Night Vision")`},
		{name: "found spell", grant: lightGrant, script: `entity.findSpells("Light").length > 0`},
		{name: "listed spell", grant: lightGrant, script: `entity.spells.some(s => s.name == "Light")`},
		{name: "weapon damage", grant: clawsGrant, script: `entity.weaponDamage("Claws", "Slash") !== undefined`},
		{name: "found weapon", grant: clawsGrant, script: `entity.findWeapons(true, "Claws", "Slash").length > 0`},
		{name: "listed weapon", grant: clawsGrant, script: `entity.weapons.some(w => w.name == "Claws")`},
	} {
		t.Run("item it puts on the sheet seen by a script through its "+tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			training := addTraitWithFeatures(e, "Training", tc.grant())
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, tc.script)
			e.Recalculate()
			c.False(e.Unsettled())
			c.NotEqual("", training.UnsatisfiedReason, "the item it grants can't meet its own prerequisite")
			c.True(grantedNotInEffect(e), "so the item it grants is not in effect")
		})
	}

	// The item as bought can, in which case the grant merges into it, and the level a script read for the
	// prerequisites, without the merge, is not what the sheet keeps. A script prerequisite run on its own, as the
	// editor's preview runs it, reads the level without the merge too.
	t.Run("merge once the prerequisite is met without it", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			prereq func(list *PrereqList)
		}{
			{name: "skill prerequisite", prereq: func(list *PrereqList) {
				addSkillPrereq(list, "Brawling", fxp.FromInteger(12))
			}},
			{name: "script reading the skill level", prereq: func(list *PrereqList) {
				addScriptPrereq(list, `entity.skillLevel("Brawling", "", false) >= 12`)
			}},
			{name: "script reading the skill level as at most 13, judged by the editor's preview as by the sheet", prereq: func(list *PrereqList) {
				addScriptPrereq(list, `entity.skillLevel("Brawling", "", false) <= 13`)
			}},
			{name: "script reading the skill found", prereq: func(list *PrereqList) {
				addScriptPrereq(list, `entity.findSkills("Brawling").some(s => s.level >= 12)`)
			}},
		} {
			c := check.New(t)
			e := NewEntity()
			brawling := addSkill(e, "Brawling", fxp.Four) // DX+2, so 12 on its own
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			tc.prereq(training.Prereq)
			e.Recalculate()
			c.False(e.Unsettled(), tc.name)
			c.Equal("", training.UnsatisfiedReason, "%s: the skill as bought meets the prerequisite", tc.name)
			c.Equal(fxp.FromInteger(15), brawling.LevelData.Level,
				"%s: so the grant merges into it, and the sheet keeps the merged level", tc.name)
			for _, one := range training.Prereq.Prereqs {
				if script, ok := one.(*ScriptPrereq); ok {
					result, _ := script.Evaluate(e, training)
					c.Equal(CheckMet, result, "%s: the script run on its own agrees with the sheet", tc.name)
				}
			}
		}
	})

	// Nor can a curse it carries leave its prerequisite unmet: a default reads the skill the curse takes from at the
	// level the prerequisites see, without the curse, though the level the skill recorded is lower, whichever skill the
	// default weighs first.
	t.Run("curse once the prerequisite is met without it", func(t *testing.T) {
		guns := func(tl string, points fxp.Int) *Skill {
			s := newEasySkill("Guns", points)
			s.Specialization = "Pistol"
			s.TechLevel = new(tl)
			return s
		}
		for _, tl8First := range []bool{false, true} {
			c := check.New(t)
			e := NewEntity()
			tl7, tl8 := guns("7", fxp.Four), guns("8", fxp.Eight) // 12 and 13
			tl7.SetDataOwner(e)
			tl8.SetDataOwner(e)
			e.Skills = []*Skill{tl7, tl8}
			if tl8First {
				e.Skills = []*Skill{tl8, tl7}
			}
			knife := addSkill(e, "Knife", fxp.One)
			knife.Defaults = []*SkillDefault{newSkillDefaultTo("Guns", "Pistol", false, 0)}
			cursed := addTraitWithFeatures(e, "Cursed",
				newSkillGrantOf(guns("8", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore))
			cursed.Prereq = newPrereqListRequiringSkill("Knife", fxp.Thirteen)
			e.Recalculate()
			c.False(e.Unsettled(), "TL8 first %v", tl8First)
			c.Equal("", cursed.UnsatisfiedReason, "TL8 first %v: Knife is 13 from Guns (Pistol)/TL8 without the curse",
				tl8First)
			c.Equal(fxp.Twelve, tl8.LevelData.Level, "TL8 first %v: so the curse takes a level from it", tl8First)
		}
	})

	// Nor can a Keep Lowest grant it carries, which caps the skill below the level the prerequisite needs.
	t.Run("keep lowest once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		brawling := addSkill(e, "Brawling", fxp.Four) // DX+2, so 12 on its own
		clumsy := addTraitWithFeatures(e, "Clumsy",
			newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Ten, grantmerge.KeepLowest))
		clumsy.Prereq = newPrereqListRequiringSkill("Brawling", fxp.Twelve)
		for _, when := range []string{"", ", recalculated again"} {
			e.Recalculate()
			c.False(e.Unsettled(), when)
			c.Equal("", clumsy.UnsatisfiedReason, "the skill as bought meets the prerequisite"+when)
			c.Equal(fxp.Ten, brawling.LevelData.Level, "so the grant keeps it at level 10"+when)
		}
	})

	// A default resting on the skill it puts on the sheet gives way to the skill's next best, as it would on the sheet
	// without the grant, when the skill is read through a technique based on it: Knife Trick is 13 through Knife's
	// default from Shortsword, and no more.
	t.Run("other default once the one in use rests on the skill it puts on the sheet", func(t *testing.T) {
		for _, tc := range []struct {
			trick fxp.Int
			met   bool
			knife fxp.Int
		}{
			{trick: fxp.Thirteen, met: true, knife: fxp.FromInteger(19)},
			{trick: fxp.FromInteger(14), met: false, knife: fxp.Twelve},
		} {
			c := check.New(t)
			e := NewEntity()
			addSkill(e, "Shortsword", fxp.Eight) // DX+3, so 13
			knife := addSkill(e, "Knife", fxp.One)
			knife.Defaults = []*SkillDefault{
				newSkillDefaultTo("Broadsword", "", true, 0),
				newSkillDefaultTo("Shortsword", "", true, -fxp.One),
			}
			trick := NewTechnique(e, nil, "Knife")
			trick.Name = "Knife Trick"
			e.Skills = append(e.Skills, trick)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Ten, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Knife Trick", tc.trick)
			e.Recalculate()
			c.False(e.Unsettled(), "Knife Trick at least %v", tc.trick)
			c.Equal(tc.met, training.UnsatisfiedReason == "", "Knife Trick at least %v: met", tc.trick)
			c.Equal(tc.knife, knife.LevelData.Level,
				"Knife Trick at least %v: Knife defaults from the Broadsword it puts on the sheet only with the grant in "+
					"effect", tc.trick)
		}
	})

	// Nor is an attribute as the prerequisites saw it, without the bonus from what the grant puts on the sheet, what
	// the sheet then shows, though the script giving the attribute its base ran for them.
	t.Run("attribute bonus once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		smart := newGrantLeveledTrait("Smart", fxp.One)
		bonus := NewAttributeBonus(IntelligenceID)
		bonus.Amount = fxp.Two
		smart.Features = Features{bonus}
		mentor := addTraitWithFeatures(e, "Mentor", newTraitGrant(smart, fxp.One, grantmerge.Ignore))
		mentor.Prereq = NewPrereqList()
		p := NewAttributePrereq(e)
		p.Parent = mentor.Prereq
		p.Which = "will"
		p.QualifierCriteria.Qualifier = fxp.Ten
		mentor.Prereq.Prereqs = append(mentor.Prereq.Prereqs, p)
		e.Recalculate()
		c.Equal("", mentor.UnsatisfiedReason, "Will is 10 without the grant")
		c.Equal(fxp.Twelve, e.ResolveAttributeCurrent("will"),
			"so the grant is in effect, and Will is derived from IQ with the bonus")
	})

	// Nor are the thresholds of a pool, which its maximum gives, taken as the sheet has them while the pool is compared
	// with them as the prerequisites see it: FP 5 of 10 is Tiring, which leaves Dodge alone, though below the Tired
	// threshold of FP 20.
	t.Run("pool threshold once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		e.Attributes.Set[FatiguePointsID].Damage = fxp.Five
		vigor := newGrantLeveledTrait("Vigor", fxp.One)
		bonus := NewAttributeBonus(FatiguePointsID)
		bonus.Amount = fxp.Ten
		vigor.Features = Features{bonus}
		training := addTraitWithFeatures(e, "Training", newTraitGrant(vigor, fxp.One, grantmerge.Ignore))
		training.Prereq = NewPrereqList()
		p := NewAttributePrereq(e)
		p.Parent = training.Prereq
		p.Which = DodgeID
		p.QualifierCriteria.Qualifier = fxp.Eight
		training.Prereq.Prereqs = append(training.Prereq.Prereqs, p)
		e.Recalculate()
		c.False(e.Unsettled())
		c.Equal("", training.UnsatisfiedReason, "Dodge is 8 without the grant")
		c.Equal(fxp.FromInteger(15), e.ResolveAttributeCurrent(FatiguePointsID), "so the grant is in effect")
	})

	// Each carrier's prerequisites see the pool's threshold as they see the pool: FP 3 of 10 is Tired, which halves
	// Dodge 8 to 4, but the FP a grant of the first carrier adds lifts the second out of it.
	t.Run("pool threshold as each carrier's prerequisites see it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		e.Attributes.Set[FatiguePointsID].Damage = fxp.Seven
		dodgeAt := func(carrier *Trait, compare criteria.NumericComparison, qualifier fxp.Int) {
			carrier.Prereq = NewPrereqList()
			p := NewAttributePrereq(e)
			p.Parent = carrier.Prereq
			p.Which = DodgeID
			p.QualifierCriteria = criteria.Number{Compare: compare, Qualifier: qualifier}
			carrier.Prereq.Prereqs = append(carrier.Prereq.Prereqs, p)
		}
		vigor := newGrantLeveledTrait("Vigor", fxp.One)
		bonus := NewAttributeBonus(FatiguePointsID)
		bonus.Amount = fxp.Ten
		vigor.Features = Features{bonus}
		tired := addTraitWithFeatures(e, "Tired Training", newTraitGrant(vigor, fxp.One, grantmerge.Ignore))
		dodgeAt(tired, criteria.AtMostNumber, fxp.Four)
		rested := addTraitWithFeatures(e, "Rested Training", nightVisionGrant())
		dodgeAt(rested, criteria.AtLeastNumber, fxp.Eight)
		e.Recalculate()
		c.False(e.Unsettled())
		c.Equal("", tired.UnsatisfiedReason, "Dodge is 4 without the first carrier's grant")
		c.Equal("", rested.UnsatisfiedReason, "Dodge is 8 with it, without the second carrier's grant")
	})

	// Nor is a weapon as a script run for the prerequisites resolves it, without the skill the grant puts on the sheet,
	// what the sheet then shows, though the script runs on its own after the recalculation, as the editor's preview runs
	// it; and the script sees the weapon so even once the sheet has resolved it.
	t.Run("weapon once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		e.Traits = nil // Drop the natural attacks
		arts := addTraitWithFeatures(e, "Martial Arts")
		strike := NewWeapon(arts, true)
		strike.Usage = "Strike"
		strike.Defaults = []*SkillDefault{newSkillDefaultTo("Karate", "", true, 0)}
		arts.Weapons = []*Weapon{strike}
		training := addTraitWithFeatures(e, "Karate Training",
			newSkillGrant(newEasySkill("Karate", fxp.One), fxp.Three, grantmerge.Ignore)) // DX+2, so 12
		training.Prereq = NewPrereqList()
		script := addScriptPrereq(training.Prereq, `entity.weapons.every(w => w.level < 12)`)
		e.Recalculate()
		c.Equal("", training.UnsatisfiedReason, "the weapon has no skill to use without the grant")
		for _, resolvedFirst := range []bool{false, true} {
			e.DiscardCaches()
			if resolvedFirst {
				c.Equal(fxp.Twelve, strike.SkillLevel(nil), "the sheet resolves the weapon with the skill granted")
			}
			result, _ := script.Evaluate(e, training)
			c.Equal(CheckMet, result, "resolved first %v: the script run on its own agrees with the sheet", resolvedFirst)
			c.Equal(fxp.Twelve, strike.SkillLevel(nil), "resolved first %v: the sheet still shows the weapon with the "+
				"skill granted", resolvedFirst)
		}
	})

	// Nor is a weapon's damage as a script run for the prerequisites resolves it, without the weapon bonus or selector
	// override from what the grant puts on the sheet, what the sheet then shows, though the script runs on its own after
	// the recalculation.
	t.Run("weapon bonus or override once the prerequisite is met without it", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			feature func() Feature
			damage  string
		}{
			{name: "weapon bonus selecting by name", feature: func() Feature {
				bonus := NewWeaponBonus(feature.WeaponBonus)
				bonus.SelectionType = wsel.WithName
				bonus.NameCriteria.Qualifier = "Natural Attacks"
				bonus.Amount = fxp.Five
				return bonus
			}, damage: "1d+2 cr"},
			{name: "weapon bonus selecting by required skill", feature: func() Feature {
				bonus := NewWeaponBonus(feature.WeaponBonus)
				bonus.NameCriteria.Compare = criteria.AnyText
				bonus.Amount = fxp.Five
				return bonus
			}, damage: "1d+2 cr"},
			{name: "selector override", feature: func() Feature {
				override := NewSelectorOverride(selector.WeaponDamageStrengthBasis)
				override.SelectionType = wsel.WithName
				override.NameCriteria.Qualifier = "Natural Attacks"
				override.Value = stdmg.Swing.Key()
				return override
			}, damage: "1d-1 cr"},
		} {
			c := check.New(t)
			e := NewEntity()
			attacks := NewNaturalAttacks(e, nil)
			e.Traits = []*Trait{attacks}
			punch := attacks.Weapons[slices.IndexFunc(attacks.Weapons, func(w *Weapon) bool { return w.Usage == "Punch" })]
			mighty := newGrantLeveledTrait("Mighty", fxp.One)
			mighty.Features = Features{tc.feature()}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(mighty, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			script := addScriptPrereq(training.Prereq, `entity.weaponDamage("Natural Attacks", "Punch") == "1d-3 cr"`)
			e.Recalculate()
			c.False(e.Unsettled(), tc.name)
			c.Equal("", training.UnsatisfiedReason, "%s: the punch does 1d-3 cr without the grant", tc.name)
			c.Equal(tc.damage, punch.Damage.ResolvedDamage(nil), "%s: so the grant is in effect, and it counts", tc.name)
			e.DiscardCaches()
			result, _ := script.Evaluate(e, training)
			c.Equal(CheckMet, result, "%s: the script run on its own agrees with the sheet", tc.name)
			c.Equal(tc.damage, punch.Damage.ResolvedDamage(nil), "%s: the sheet still shows the punch with it", tc.name)
		}
	})

	// Nor does a selector override of a trait's self-control roll or adjustment that what the grant puts on the sheet
	// carries.
	t.Run("trait selector override once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		badTemper := addTraitWithFeatures(e, "Bad Temper")
		badTemper.BasePoints = -fxp.Ten
		badTemper.SelfControl = selfctrl.CR12
		calm := NewTrait(nil, nil, false)
		calm.Name = "Calm"
		override := NewSelectorOverride(selector.TraitSelfControlRoll)
		override.NameCriteria.Qualifier = "Bad Temper"
		override.Value = selfControlRollStateKey(selfctrl.CR15)
		calm.Features = Features{override}
		training := addTraitWithFeatures(e, "Training", newTraitGrant(calm, fxp.One, grantmerge.Ignore))
		training.Prereq = NewPrereqList()
		script := addScriptPrereq(training.Prereq,
			`entity.traits.some(t => t.name == "Bad Temper" && t.selfControl == 12)`)
		e.Recalculate()
		c.False(e.Unsettled())
		c.Equal("", training.UnsatisfiedReason, "Bad Temper is at CR12 without the grant")
		c.Equal(selfctrl.CR15, badTemper.ResolvedSelfControl(nil), "so the grant is in effect, and its override counts")
		e.DiscardCaches()
		result, _ := script.Evaluate(e, training)
		c.Equal(CheckMet, result, "the script run on its own agrees with the sheet")

		e = NewEntity()
		miserliness := addTraitWithFeatures(e, "Miserliness")
		miserliness.BasePoints = -fxp.Ten
		miserliness.SelfControl = selfctrl.CR6
		merchant := addSkill(e, "Merchant", fxp.One)
		frugal := NewTrait(nil, nil, false)
		frugal.Name = "Frugal"
		frugal.Features = Features{newTraitSelectorOverride(selector.TraitSelfControlAdjustment, "Miserliness",
			selfctrl.MajorCostOfLivingIncrease.Key())}
		haggler := addTraitWithFeatures(e, "Haggler", newTraitGrant(frugal, fxp.One, grantmerge.Ignore))
		haggler.Prereq = newPrereqListRequiringSkill("Merchant", fxp.Eight)
		e.Recalculate()
		c.False(e.Unsettled(), "Haggler")
		c.Equal("", haggler.UnsatisfiedReason, "Merchant is 10 without the grant")
		c.Equal(fxp.Six, merchant.LevelData.Level,
			"so the grant is in effect, and its override gives Merchant the penalty of Miserliness at CR6")
	})

	// Nor are the Basic Lift, the carried weight and the encumbrance a script run for the prerequisites reads what the
	// sheet keeps: ST is derived from IQ here and the pack weighs three times the ST, so the script reads them without
	// the bonus to IQ from what the grant puts on the sheet, at ST 10: 20 lb, 30 lb and a Light load.
	t.Run("Basic Lift, carried weight and encumbrance once the prerequisite is met without it", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		e.SheetSettings.Attributes.Set[StrengthID].Base = "$iq"
		addCarriedEquipmentWithFeatures(e, "Pack").BaseWeight = "$st * 3"
		smart := newGrantLeveledTrait("Smart", fxp.One)
		bonus := NewAttributeBonus(IntelligenceID)
		bonus.Amount = fxp.Five
		smart.Features = Features{bonus}
		training := addTraitWithFeatures(e, "Training", newTraitGrant(smart, fxp.One, grantmerge.Ignore))
		training.Prereq = NewPrereqList()
		script := addScriptPrereq(training.Prereq, `entity.encumbrance.basicLift < 45 && `+
			`entity.encumbrance.weightCarried < 45 && entity.encumbrance.level == 1`)
		e.Recalculate()
		c.Equal("", training.UnsatisfiedReason)
		lift := fxp.WeightFromInteger(45, fxp.Pound)
		checkSheet := func(when string) {
			c.Equal(lift, e.BasicLift(), "%s: ST is 15, from the IQ with the bonus", when)
			c.Equal(lift, e.WeightCarried(false), "%s: so the pack weighs 45 lb", when)
			c.Equal(encumbrance.No, e.EncumbranceLevel(false), "%s: which is no load", when)
		}
		checkSheet("recalculated")
		e.DiscardCaches()
		result, _ := script.Evaluate(e, training)
		c.Equal(CheckMet, result)
		checkSheet("after the script is run on its own, as the editor's preview runs it")
	})

	// Nor does a level another carrier's grant sets on an item rest, for the prerequisites, on what the carrier's own
	// grants give the attribute or skill the item rests on, since the other grant gives the level set whether or not
	// the carrier's grants are in effect: IQ 12 or Karate 12 with them, 10 without.
	t.Run("level set by another carrier on what the carrier's grants raise", func(t *testing.T) {
		smart := func(iq fxp.Int) func() *Grant {
			return func() *Grant {
				trait := newGrantLeveledTrait("Smart", fxp.One)
				bonus := NewAttributeBonus(IntelligenceID)
				bonus.Amount = iq
				trait.Features = Features{bonus}
				return newTraitGrant(trait, fxp.One, grantmerge.Ignore)
			}
		}
		karate := func() *Grant { return newSkillGrant(newHardSkill("Karate", fxp.Four), fxp.Two, grantmerge.Ignore) }
		for _, tc := range []struct {
			name  string
			raise func() *Grant
			item  *Skill
			// bought, when set, puts the item on the sheet at 1 point, which the level set, Keep Highest, merges into.
			bought bool
		}{
			{name: "skill resting on an attribute", raise: smart(fxp.Two), item: newHardSkill("Alchemy", 0)},
			{name: "technique resting on a skill", raise: karate, item: newDisarming(0)},
			{name: "level set kept", raise: smart(fxp.Two), item: newHardSkill("Alchemy", 0), bought: true},
			{
				// At IQ 15, 1 point buys 13, so that the level set gives nothing with the carrier's grants.
				name: "level set kept, giving nothing with the carrier's grants", raise: smart(fxp.Five),
				item: newHardSkill("Alchemy", 0), bought: true,
			},
		} {
			c := check.New(t)
			e := NewEntity()
			addHardSkill(e, "Karate", fxp.Four)
			merge := grantmerge.Ignore
			if tc.bought {
				addHardSkill(e, tc.item.Name, fxp.One)
				merge = grantmerge.KeepHighest
			}
			x := addTraitWithFeatures(e, "X", tc.raise())
			x.Prereq = NewPrereqList()
			addSkillPrereq(x.Prereq, tc.item.Name, fxp.Twelve)
			addTraitWithFeatures(e, "Y", newSkillGrantOf(tc.item, grantadjust.SetLevel, fxp.Twelve, merge))
			e.Recalculate()
			c.False(e.Unsettled(), tc.name)
			c.Equal("", x.UnsatisfiedReason, tc.name)
		}
	})

	t.Run("item another grant puts on the sheet too", func(t *testing.T) {
		newGrant := func() *Grant {
			return newTraitGrant(newGrantLeveledTrait("X", fxp.One), fxp.One, grantmerge.Ignore)
		}
		for _, carrierFirst := range []bool{true, false} {
			c := check.New(t)
			e := NewEntity()
			a := addTraitWithFeatures(e, "A", newGrant())
			a.Prereq = newPrereqListForTrait("X", true)
			b := addTraitWithFeatures(e, "B", newGrant())
			if !carrierFirst {
				e.Traits = []*Trait{b, a}
			}
			e.Recalculate()
			c.False(e.Unsettled())
			c.Equal("", a.UnsatisfiedReason, "the other grant's X meets A's prerequisite whichever is listed first")
			granted := grantedOf(e.Traits)
			if len(granted) != 1 {
				c.Fatalf("expected one X row, got %d", len(granted))
			}
			c.True(granted[0].Enabled())
		}
	})

	// Nor does an item another grant puts there too while the carrier's own grant gives the item its details, as the
	// grant that put it there or one standing in for that grant: the item is judged as the other grant would make it.
	t.Run("item another grant puts on the sheet too, as that grant makes it", func(t *testing.T) {
		nightVision := func(level fxp.Int) *Grant {
			return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), level, grantmerge.Ignore)
		}
		// keenNightVision is nightVision at 10 points a level with a +10% modifier costed at the trait's levels.
		keenNightVision := func(level fxp.Int) *Grant {
			g := newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.Ten), level, grantmerge.Ignore)
			mod := NewTraitModifier(nil, nil, false)
			mod.Name = "Keen"
			mod.CostAdj = "+10%"
			mod.UseLevelFromTrait = true
			g.Trait.Modifiers = []*TraitModifier{mod}
			return g
		}
		// brawling and light set the level of an Easy DX skill and a Hard IQ spell under Ignore, as nightVision sets a
		// trait's, so that none of these merges into the row another grant of the item puts on the sheet.
		brawling := func(level fxp.Int) *Grant {
			return newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, level, grantmerge.Ignore)
		}
		brawlingIn := func(level fxp.Int, optionalSpecialization string) *Grant {
			g := brawling(level)
			g.Skill.OptionalSpecialization = optionalSpecialization
			return g
		}
		light := func(level fxp.Int) *Grant {
			return newSpellGrantOf(newHardSpell("Light", fxp.One), grantadjust.SetLevel, level, grantmerge.Ignore)
		}
		// knuckled gives the grant's skill or spell a prerequisite of brass knuckles equipped, which nothing meets, and
		// has the grant check it when checked is set.
		knuckled := func(g *Grant, checked bool) *Grant {
			g.CheckPrereqs = checked
			list := NewPrereqList()
			p := NewEquippedEquipmentPrereq()
			p.Parent = list
			p.NameCriteria.Qualifier = "Brass Knuckles"
			list.Prereqs = append(list.Prereqs, p)
			if g.Skill != nil {
				g.Skill.Prereq = list
			} else {
				g.Spell.Prereq = list
			}
			return g
		}
		// mage puts on the sheet a trait carrying the grant whose prerequisite is Light at 8 or more.
		mage := func(e *Entity, g *Grant) *Trait {
			carrier := addTraitWithFeatures(e, "Mage", g)
			carrier.Prereq = NewPrereqList()
			p := NewSpellPrereq()
			p.Parent = carrier.Prereq
			p.SamePowerSource = false
			p.QualifierCriteria.Qualifier = "Light"
			p.LevelCriteria.Compare = criteria.AtLeastNumber
			p.LevelCriteria.Qualifier = fxp.Eight
			carrier.Prereq.Prereqs = append(carrier.Prereq.Prereqs, p)
			return carrier
		}
		// trait returns an Ignore grant of a trait with the name, which set adjusts.
		trait := func(name string, set func(item *Trait)) *Grant {
			item := NewTrait(nil, nil, false)
			item.Name = name
			set(item)
			return newTraitGrant(item, fxp.One, grantmerge.Ignore)
		}
		badTemper := func(cr selfctrl.Roll) *Grant {
			return trait("Bad Temper", func(item *Trait) {
				item.BasePoints = -fxp.Ten
				item.SelfControl = cr
			})
		}
		ally := func(freq frequency.Roll) *Grant {
			return trait("Ally", func(item *Trait) {
				item.BasePoints = fxp.Five
				item.Frequency = freq
			})
		}
		contact := func(desc string) *Grant {
			return trait("Contact", func(item *Trait) { item.UserDesc = desc })
		}
		miserliness := func(cr selfctrl.Roll) *Grant {
			return trait("Miserliness", func(item *Trait) {
				item.BasePoints = -fxp.Ten
				item.SelfControl = cr
				item.SelfControlAdj = selfctrl.MajorCostOfLivingIncrease
			})
		}
		// agile is a grant at the level of a leveled trait giving +2 DX whose prerequisite nothing meets, which the
		// grant checks when checked is set.
		agile := func(level fxp.Int, checked bool) *Grant {
			item := newGrantLeveledTrait("Agile", fxp.One)
			bonus := NewAttributeBonus(DexterityID)
			bonus.Amount = fxp.Two
			item.Features = Features{bonus}
			item.Prereq = newPrereqListForTrait("Elf", true)
			g := newTraitGrant(item, level, grantmerge.Ignore)
			g.CheckPrereqs = checked
			return g
		}
		// cappedAgile is agile with a maximum level of 1, given by a grant that treats its prerequisite as met.
		cappedAgile := func(level fxp.Int) *Grant {
			g := agile(level, false)
			g.Trait.MaxLevels = "1"
			return g
		}
		// scripted puts on the sheet a trait with the name, carrying the grant, whose prerequisite is the script.
		scripted := func(e *Entity, name string, g *Grant, script string) *Trait {
			carrier := addTraitWithFeatures(e, name, g)
			carrier.Prereq = NewPrereqList()
			addScriptPrereq(carrier.Prereq, script)
			return carrier
		}
		// goggles puts on the sheet worn goggles carrying the second grant and, when the first is not nil, unequipped
		// ones carrying that.
		goggles := func(e *Entity, spare, worn *Grant) {
			if spare != nil {
				addCarriedEquipmentWithFeatures(e, "Spare Goggles", spare).Equipped = false
			}
			addCarriedEquipmentWithFeatures(e, "Worn Goggles", worn)
		}
		seer := func(c check.Checker, e *Entity, level, atLeast fxp.Int) *Trait {
			carrier := addTraitWithFeatures(e, "Seer", nightVision(level))
			carrier.Prereq = newPrereqListForTrait("Night Vision", true)
			traitPrereqOf(c, carrier.Prereq).LevelCriteria.Qualifier = atLeast
			return carrier
		}
		nightVisionLevel := func(e *Entity) any { return grantedOf(e.Traits)[0].CurrentLevel().AsInteger[int]() }
		for _, tc := range []struct {
			name  string
			carry func(c check.Checker, e *Entity) *Trait
			met   bool
			// row returns the detail of the granted row, once settled, that tells which grant gave it its details.
			row  func(e *Entity) any
			want any
		}{
			{name: "standing in for a trait", carry: func(c check.Checker, e *Entity) *Trait {
				goggles(e, nightVision(fxp.Three), nightVision(fxp.One))
				return seer(c, e, fxp.Two, fxp.Two)
			}, row: nightVisionLevel, want: 1},
			{name: "putting a trait there", carry: func(c check.Checker, e *Entity) *Trait {
				goggles(e, nil, nightVision(fxp.One))
				return seer(c, e, fxp.Three, fxp.Two)
			}, row: nightVisionLevel, want: 1},
			{
				name: "standing in for a trait the other grant's level meets",
				carry: func(c check.Checker, e *Entity) *Trait {
					goggles(e, nightVision(fxp.Three), nightVision(fxp.One))
					return seer(c, e, fxp.Two, fxp.One)
				}, met: true, row: nightVisionLevel, want: 2,
			},
			{
				name: "putting a trait there whose modifier takes its level from it",
				carry: func(c check.Checker, e *Entity) *Trait {
					goggles(e, nil, keenNightVision(fxp.One))
					carrier := addTraitWithFeatures(e, "Seer", keenNightVision(fxp.Three))
					carrier.Prereq = newPrereqListForTrait("Night Vision", true)
					p := traitPrereqOf(c, carrier.Prereq)
					p.PointsCriteria.Compare = criteria.AtLeastNumber
					p.PointsCriteria.Qualifier = fxp.FromInteger(13)
					return carrier
				}, row: nightVisionLevel, want: 1,
			},
			{
				name: "standing in for a trait whose points a prerequisite reads",
				carry: func(c check.Checker, e *Entity) *Trait {
					goggles(e, nightVision(fxp.Three), nightVision(fxp.One))
					carrier := seer(c, e, fxp.Two, 0)
					p := traitPrereqOf(c, carrier.Prereq)
					p.PointsCriteria.Compare = criteria.AtLeastNumber
					p.PointsCriteria.Qualifier = fxp.Two
					return carrier
				}, row: nightVisionLevel, want: 1,
			},
			{name: "standing in for a trait read by a script", carry: func(_ check.Checker, e *Entity) *Trait {
				goggles(e, badTemper(selfctrl.CR6), badTemper(selfctrl.CR12))
				return scripted(e, "Hothead", badTemper(selfctrl.CR9),
					`entity.traits.some(t => t.name == "Bad Temper" && t.selfControl <= 9)`)
			}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].SelfControl.Number() }, want: 12},
			{
				name: "standing in for a trait whose frequency a script reads",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, ally(frequency.FR15), ally(frequency.FR9))
					return scripted(e, "Leader", ally(frequency.FR12),
						`entity.traits.some(t => t.name == "Ally" && t.frequency >= 12)`)
				}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Frequency }, want: frequency.FR9,
			},
			{
				name: "standing in for a trait whose notes a script reads",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, contact("Army"), contact("Police"))
					return scripted(e, "Sailor", contact("Navy"),
						`entity.traits.some(t => t.name == "Contact" && t.notes.includes("Navy"))`)
				}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].UserDesc }, want: "Police",
			},
			{
				name: "putting a trait there whose self-control roll penalizes a skill a prerequisite reads",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, miserliness(selfctrl.CR12))
					addSkill(e, "Merchant", fxp.One)
					haggler := addTraitWithFeatures(e, "Haggler", miserliness(selfctrl.CR6))
					haggler.Prereq = newPrereqListRequiringSkill("Merchant", fxp.Eight)
					return haggler
				}, met: true, row: func(e *Entity) any { return grantedOf(e.Traits)[0].SelfControl.Number() }, want: 6,
			},
			{
				name: "putting a trait there whose prerequisites only the other grant checks",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					goggles(e, nil, agile(fxp.One, true))
					addSkill(e, "Brawling", fxp.One)
					training := addTraitWithFeatures(e, "Training", agile(fxp.One, false))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
					return training
				}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Enabled() }, want: false,
			},
			{
				name: "putting a trait there whose prerequisites the other grant bypasses",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					goggles(e, nil, agile(fxp.One, false))
					addSkill(e, "Brawling", fxp.One)
					training := addTraitWithFeatures(e, "Training", agile(fxp.Two, true))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
					return training
				}, met: true, want: 10,
				row: func(e *Entity) any { return e.ResolveAttributeCurrent(DexterityID).AsInteger[int]() },
			},
			{
				name: "putting a trait there whose prerequisites the other grant bypasses, which a prerequisite names",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					goggles(e, nil, agile(fxp.One, false))
					training := addTraitWithFeatures(e, "Training", agile(fxp.Two, true))
					training.Prereq = newPrereqListForTrait("Agile", true)
					return training
				}, met: true, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Enabled() }, want: false,
			},
			{
				name: "putting a trait there above its maximum level, which the other grant's level is not",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					goggles(e, nil, cappedAgile(fxp.One))
					addSkill(e, "Brawling", fxp.One)
					training := addTraitWithFeatures(e, "Training", cappedAgile(fxp.Two))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
					return training
				}, met: true, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Enabled() }, want: false,
			},
			{
				name: "putting a trait there above its maximum level, as the other grant's level is too",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					goggles(e, nil, cappedAgile(fxp.Two))
					training := addTraitWithFeatures(e, "Training", cappedAgile(fxp.Three))
					training.Prereq = newPrereqListForTrait("Agile", true)
					return training
				}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Enabled() }, want: false,
			},
			{
				name: "putting a trait there within its maximum level, which the level the other grant keeps is not",
				carry: func(_ check.Checker, e *Entity) *Trait {
					e.SheetSettings.EnforceTraitPrereqs = true
					kept := cappedAgile(fxp.Two)
					kept.Merge = grantmerge.KeepHighest
					goggles(e, nil, kept)
					addSkill(e, "Brawling", fxp.One)
					training := addTraitWithFeatures(e, "Training", cappedAgile(fxp.One))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
					return training
				}, row: func(e *Entity) any { return grantedOf(e.Traits)[0].Enabled() }, want: false,
			},
			{
				name: "standing in for a skill the other grant's level meets",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, brawling(fxp.Thirteen), brawling(fxp.Ten))
					training := addTraitWithFeatures(e, "Training", brawling(fxp.Twelve))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.Ten)
					return training
				}, met: true, want: 12,
				row: func(e *Entity) any { return grantedOf(e.Skills)[0].LevelData.Level.AsInteger[int]() },
			},
			{
				name: "standing in for a skill whose optional specialization a prerequisite reads",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, brawling(fxp.Thirteen), brawlingIn(fxp.Ten, "Wrestling"))
					training := addTraitWithFeatures(e, "Training", brawlingIn(fxp.Twelve, "Boxing"))
					training.Prereq = NewPrereqList()
					p := addSkillPrereq(training.Prereq, "Brawling", 0)
					p.OptionalSpecializationCriteria.Compare = criteria.IsText
					p.OptionalSpecializationCriteria.Qualifier = "Boxing"
					return training
				}, row: func(e *Entity) any { return grantedOf(e.Skills)[0].OptionalSpecialization }, want: "Wrestling",
			},
			{
				name: "putting a skill there whose missing-equipment penalty a prerequisite reads",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, knuckled(brawlingIn(fxp.Ten, "Wrestling"), true))
					training := addTraitWithFeatures(e, "Training", knuckled(brawlingIn(fxp.Twelve, "Boxing"), true))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.Eight)
					return training
				}, row: func(e *Entity) any { return grantedOf(e.Skills)[0].OptionalSpecialization }, want: "Wrestling",
			},
			{
				name: "putting a skill there whose missing-equipment penalty the other grant bypasses",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, knuckled(brawling(fxp.Ten), false))
					training := addTraitWithFeatures(e, "Training", knuckled(brawling(fxp.Twelve), true))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.Eight)
					return training
				}, met: true, want: 7,
				row: func(e *Entity) any { return grantedOf(e.Skills)[0].LevelData.Level.AsInteger[int]() },
			},
			{
				name: "putting a skill there whose missing-equipment penalty only the other grant takes",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, knuckled(brawling(fxp.Ten), true))
					training := addTraitWithFeatures(e, "Training", knuckled(brawling(fxp.Twelve), false))
					training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.Eight)
					return training
				}, want: 5,
				row: func(e *Entity) any { return grantedOf(e.Skills)[0].LevelData.Level.AsInteger[int]() },
			},
			{
				name: "putting a spell there whose missing-equipment penalty the other grant bypasses",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, knuckled(light(fxp.Eight), false))
					return mage(e, knuckled(light(fxp.Ten), true))
				}, met: true, want: 5,
				row: func(e *Entity) any { return grantedOf(e.Spells)[0].LevelData.Level.AsInteger[int]() },
			},
			{
				name: "putting a spell there whose missing-equipment penalty only the other grant takes",
				carry: func(_ check.Checker, e *Entity) *Trait {
					goggles(e, nil, knuckled(light(fxp.Eight), true))
					return mage(e, knuckled(light(fxp.Ten), false))
				}, want: 3,
				row: func(e *Entity) any { return grantedOf(e.Spells)[0].LevelData.Level.AsInteger[int]() },
			},
		} {
			c := check.New(t)
			e := NewEntity()
			carrier := tc.carry(c, e)
			// A recalculation begins from the state the last one left, which must not change where it ends.
			for _, when := range []string{"", ", recalculated again"} {
				e.Recalculate()
				c.False(e.Unsettled(), tc.name+when)
				c.Equal(tc.met, carrier.UnsatisfiedReason == "", "%s%s: %s", tc.name, when, carrier.UnsatisfiedReason)
				c.Equal(tc.want, tc.row(e), tc.name+when)
			}
		}
	})

	// Nor does the skill a technique or ritual magic spell is based on, when the item's own grant puts it there.
	for _, tc := range []struct {
		name string
		// carry puts on the sheet the technique or ritual magic spell, carrying a grant of the skill it is based on,
		// and returns it along with the reason it records.
		carry func(e *Entity) (carrier any, reason func() string, granted func() bool)
	}{
		{name: "technique's base skill", carry: func(e *Entity) (any, func() string, func() bool) {
			disarming := NewTechnique(e, nil, "Brawling")
			disarming.Name = "Disarming"
			disarming.Features = Features{newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore)}
			e.Skills = append(e.Skills, disarming)
			return disarming, func() string { return disarming.UnsatisfiedReason },
				func() bool { return len(grantedOf(e.Skills)) == 1 && grantedOf(e.Skills)[0].grantNotInEffect() }
		}},
		{name: "ritual magic spell's ritual skill", carry: func(e *Entity) (any, func() string, func() bool) {
			rite := NewRitualMagicSpell(e, nil, false)
			rite.Name = "Rite"
			rite.College = []string{"Fire"}
			ritual := newEasySkill("Ritual Magic", fxp.One)
			ritual.Specialization = "Fire"
			rite.Features = Features{newSkillGrant(ritual, fxp.Three, grantmerge.Ignore)}
			e.Spells = append(e.Spells, rite)
			return rite, func() string { return rite.UnsatisfiedReason },
				func() bool { return len(grantedOf(e.Skills)) == 1 && grantedOf(e.Skills)[0].grantNotInEffect() }
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			_, reason, granted := tc.carry(e)
			e.Recalculate()
			c.False(e.Unsettled())
			c.NotEqual("", reason(), "the skill it grants can't be the one it is based on")
			c.True(granted(), "so the skill it grants is not in effect")
		})
	}

	t.Run("merge from a skill judged before the levels", func(t *testing.T) {
		c := check.New(t)
		e := NewEntity()
		brawling := addSkill(e, "Brawling", fxp.One)
		karate := addSkill(e, "Karate", fxp.One)
		karate.Features = Features{newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Two, grantmerge.Ignore)}
		karate.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
		e.Recalculate()
		c.NotEqual("", karate.UnsatisfiedReason, "what the grant merges in can't meet its carrier's prerequisite")
		c.Equal(fxp.Ten, brawling.LevelData.Level, "so nothing is merged")
		c.False(e.Unsettled())
	})

	// Nor can what it merges into an item already present.
	for _, tc := range []struct {
		name string
		// carry puts on the sheet an item the user bought and the carrier of a grant that merges into it enough to
		// meet the carrier's prerequisite, which the item as bought does not, and returns both along with what reads
		// the item's level.
		carry func(c check.Checker, e *Entity) (carrier *Trait, level func() fxp.Int)
		level fxp.Int
	}{
		{name: "levels into a trait", carry: func(c check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			gifted := addTraitWithFeatures(e, "Gifted",
				addingLevels(newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.Ignore)))
			gifted.Prereq = newPrereqListForTrait("Magery", true)
			traitPrereqOf(c, gifted.Prereq).LevelCriteria.Qualifier = fxp.Two
			return gifted, magery.CurrentLevel
		}, level: fxp.One},
		{name: "levels into a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a skill read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.skillLevel("Brawling", "", false) >= 12`)
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a skill found by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.findSkills("Brawling").some(s => s.level >= 12)`)
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a trait read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			gifted := addTraitWithFeatures(e, "Gifted",
				addingLevels(newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.Ignore)))
			gifted.Prereq = NewPrereqList()
			addScriptPrereq(gifted.Prereq, `entity.traitLevel("Magery") >= 2`)
			return gifted, magery.CurrentLevel
		}, level: fxp.One},
		{name: "levels into a spell found by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			magery := addTraitWithFeatures(e, "Magery",
				newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.Ignore))
			magery.Prereq = NewPrereqList()
			addScriptPrereq(magery.Prereq, `entity.findSpells("Light").some(s => s.level >= 10)`)
			return magery, func() fxp.Int { return light.LevelData.Level }
		}, level: fxp.Eight},
		{name: "points into a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, fxp.Four, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "level set kept for a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve,
					grantmerge.KeepHighest))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "bonus from a granted trait to a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			boost := newGrantLeveledTrait("Boost", fxp.One)
			boost.Features = Features{newSkillBonusTo("Brawling", fxp.Two)}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(boost, fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "override from a granted trait taking a Merchant penalty off a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			miserliness := addTraitWithFeatures(e, "Miserliness")
			miserliness.BasePoints = -fxp.Ten
			miserliness.SelfControl = selfctrl.CR6
			miserliness.SelfControlAdj = selfctrl.MajorCostOfLivingIncrease
			merchant := addSkill(e, "Merchant", fxp.One) // 6 with the penalty of Miserliness at CR6
			frugal := NewTrait(nil, nil, false)
			frugal.Name = "Frugal"
			frugal.Features = Features{newTraitSelectorOverride(selector.TraitSelfControlAdjustment, "Miserliness",
				selfctrl.NoAdjustment.Key())}
			haggler := addTraitWithFeatures(e, "Haggler", newTraitGrant(frugal, fxp.One, grantmerge.Ignore))
			haggler.Prereq = newPrereqListRequiringSkill("Merchant", fxp.Eight)
			return haggler, func() fxp.Int { return merchant.LevelData.Level }
		}, level: fxp.Six},
		{name: "bonus from a granted trait to an attribute", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Amount = fxp.Two
			strong.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			p := NewAttributePrereq(e)
			p.Parent = training.Prereq
			p.QualifierCriteria.Qualifier = fxp.FromInteger(12)
			training.Prereq.Prereqs = append(training.Prereq.Prereqs, p)
			return training, func() fxp.Int { return e.ResolveAttributeCurrent(StrengthID) }
		}, level: fxp.Ten},
		{name: "bonus from a granted trait to the size modifier read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			big := newGrantLeveledTrait("Big", fxp.One)
			big.Features = Features{NewAttributeBonus(SizeModifierID)}
			giant := addTraitWithFeatures(e, "Giant", newTraitGrant(big, fxp.One, grantmerge.Ignore))
			giant.Prereq = NewPrereqList()
			addScriptPrereq(giant.Prereq, `entity.sizeModifier >= 1 ? "" : "Too small"`)
			return giant, func() fxp.Int { return fxp.FromInteger(e.Profile.AdjustedSizeModifier()) }
		}, level: 0},
		{name: "bonus from a granted trait to the size modifier read by an attribute prerequisite", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			big := newGrantLeveledTrait("Big", fxp.One)
			big.Features = Features{NewAttributeBonus(SizeModifierID)}
			giant := addTraitWithFeatures(e, "Giant", newTraitGrant(big, fxp.One, grantmerge.Ignore))
			giant.Prereq = NewPrereqList()
			p := NewAttributePrereq(e)
			p.Parent = giant.Prereq
			p.Has = false // Met by a size modifier above 0
			p.Which = SizeModifierID
			p.QualifierCriteria = criteria.Number{Compare: criteria.AtMostNumber}
			giant.Prereq.Prereqs = append(giant.Prereq.Prereqs, p)
			return giant, func() fxp.Int { return fxp.FromInteger(e.Profile.AdjustedSizeModifier()) }
		}, level: 0},
		{name: "bonus from a granted trait to the size modifier read by a script as the points spent on attributes", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			e.Attributes.Set[StrengthID].Adjustment = fxp.Ten // ST 20 costs 100 points, 90 at SM +1
			big := newGrantLeveledTrait("Big", fxp.One)
			big.Features = Features{NewAttributeBonus(SizeModifierID)}
			giant := addTraitWithFeatures(e, "Giant", newTraitGrant(big, fxp.One, grantmerge.Ignore))
			giant.Prereq = NewPrereqList()
			addScriptPrereq(giant.Prereq, `entity.points.attributes < 100`)
			return giant, func() fxp.Int { return e.Attributes.Cost(StrengthID) }
		}, level: fxp.Hundred},
		{name: "cost reduction from a granted trait read by a script as the points spent on attributes", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			e.Attributes.Set[StrengthID].Adjustment = fxp.Ten // ST 20 costs 100 points, 50 with the reduction
			cheap := newGrantLeveledTrait("Cheap", fxp.One)
			reduction := NewCostReduction(StrengthID)
			reduction.Percentage = fxp.Fifty
			cheap.Features = Features{reduction}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(cheap, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.points.attributes < 100`)
			return training, func() fxp.Int { return e.Attributes.Cost(StrengthID) }
		}, level: fxp.Hundred},
		{name: "bonus from a granted trait to ST read by a script as striking ST", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Amount = fxp.Five
			strong.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.strikingStrength >= 15 ? "" : "Too weak"`)
			return training, e.StrikingStrength
		}, level: fxp.Ten},
		{name: "lifting ST bonus from a granted trait read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Limitation = stlimit.LiftingOnly
			bonus.Amount = fxp.Five
			strong.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.liftingStrength >= 15`)
			return training, e.LiftingStrength
		}, level: fxp.Ten},
		{name: "ST bonus from a granted trait read by a script as a weapon's damage", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			fist := addTraitWithFeatures(e, "Fist")
			punch := NewWeapon(fist, true) // Thrust crushing, so 1d-2 cr at ST 10
			punch.Usage = "Punch"
			fist.Weapons = []*Weapon{punch}
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Amount = fxp.Five
			strong.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.weaponDamage("Fist", "Punch") == "1d-2 cr" ? "Too weak" : ""`)
			return training, e.StrikingStrength
		}, level: fxp.Ten},
		{name: "ST bonus from a granted trait read by a script as a thrown weapon's range", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			rock := addTraitWithFeatures(e, "Rock")
			throw := NewWeapon(rock, false)
			throw.Usage = "Throw"
			throw.Range = attack.ParseRange("x1/x1.5") // 10/15 at ST 10
			rock.Weapons = []*Weapon{throw}
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Amount = fxp.Five
			strong.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq,
				`entity.findWeapons(false, "Rock", "Throw").some(w => w.range == "10/15") ? "Too weak" : ""`)
			return training, e.ThrowingStrength
		}, level: fxp.Ten},
		{name: "IQ bonus from a granted trait read by a script as an IQ-based weapon's damage", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			mind := addTraitWithFeatures(e, "Mind")
			blast := NewWeapon(mind, true)
			blast.Usage = "Blast"
			blast.Damage.StrengthType = stdmg.IQThrust // 1d-2 cr at IQ 10
			mind.Weapons = []*Weapon{blast}
			smart := newGrantLeveledTrait("Smart", fxp.One)
			bonus := NewAttributeBonus(IntelligenceID)
			bonus.Amount = fxp.Five
			smart.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(smart, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.weaponDamage("Mind", "Blast") == "1d-2 cr" ? "Too dull" : ""`)
			return training, func() fxp.Int { return e.ResolveAttributeCurrent(IntelligenceID) }
		}, level: fxp.Ten},
		{name: "levels into a skill read by a script as a weapon's level", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			sword := addTraitWithFeatures(e, "Sword")
			swing := NewWeapon(sword, true)
			swing.Usage = "Swing"
			swing.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, 0)}
			sword.Weapons = []*Weapon{swing}
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.weapons.some(w => w.level >= 14)`)
			return training, func() fxp.Int { return swing.SkillLevel(nil) }
		}, level: fxp.Ten},
		{name: "levels into a skill read by a script as a weapon's parry and block", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			sword := addTraitWithFeatures(e, "Sword")
			swing := NewWeapon(sword, true)
			swing.Usage = "Swing"
			swing.Parry.CanParry = true
			swing.Block.CanBlock = true
			swing.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, 0)}
			sword.Weapons = []*Weapon{swing}
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq,
				`entity.weapons.some(w => w.kind == "melee" && (parseInt(w.parry) >= 10 || parseInt(w.block) >= 10))`)
			return training, func() fxp.Int { return swing.ResolvedParry(nil).Modifier + swing.ResolvedBlock(nil).Modifier }
		}, level: fxp.FromInteger(16)},
		{name: "levels into a skill read by a script through a weapon bonus selecting by relative skill level", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One) // DX+0
			bonus := NewWeaponBonus(feature.WeaponBonus)
			bonus.NameCriteria.Qualifier = "Brawling"
			bonus.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
			bonus.Amount = fxp.Five
			fist := addTraitWithFeatures(e, "Fist", bonus)
			punch := NewWeapon(fist, true) // Thrust crushing, so 1d-2 cr at ST 10
			punch.Usage = "Punch"
			punch.Defaults = []*SkillDefault{newSkillDefaultTo("Brawling", "", true, 0)}
			fist.Weapons = []*Weapon{punch}
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Two, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.weaponDamage("Fist", "Punch") == "1d-2 cr" ? "Too weak" : ""`)
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "ST bonus from a granted trait read by a script as the encumbrance", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			strong := newGrantLeveledTrait("Strong", fxp.One)
			bonus := NewAttributeBonus(StrengthID)
			bonus.Amount = fxp.Five
			strong.Features = Features{bonus}
			addCarriedEquipmentWithFeatures(e, "Pack").BaseWeight = "30 lb" // Light at ST 10, but not at ST 15
			// A skill with an encumbrance penalty looks up the encumbrance before the traits are judged.
			addSkill(e, "Climbing", fxp.One).EncumbrancePenaltyMultiplier = fxp.One
			training := addTraitWithFeatures(e, "Training", newTraitGrant(strong, fxp.One, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.currentEncumbrance(true, false) == 0 ? "" : "Too burdened"`)
			return training, func() fxp.Int { return fxp.FromInteger(int(e.EncumbranceLevel(true))) }
		}, level: fxp.One},
		{name: "bonus from a granted trait to the attribute another is derived from", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			smart := newGrantLeveledTrait("Smart", fxp.One)
			bonus := NewAttributeBonus(IntelligenceID)
			bonus.Amount = fxp.Two
			smart.Features = Features{bonus}
			mentor := addTraitWithFeatures(e, "Mentor", newTraitGrant(smart, fxp.One, grantmerge.Ignore))
			mentor.Prereq = NewPrereqList()
			p := NewAttributePrereq(e)
			p.Parent = mentor.Prereq
			p.Which = "will"
			p.QualifierCriteria.Qualifier = fxp.FromInteger(12)
			mentor.Prereq.Prereqs = append(mentor.Prereq.Prereqs, p)
			return mentor, func() fxp.Int { return e.ResolveAttributeCurrent("will") }
		}, level: fxp.Ten},
		{name: "per-level bonus from the levels merged into a trait", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			newMagery := func() *Trait {
				magery := newGrantLeveledTrait("Magery", fxp.Ten)
				bonus := NewAttributeBonus(IntelligenceID)
				bonus.PerLevel = true
				magery.Features = Features{bonus}
				return magery
			}
			addTrait(e, newMagery())
			gifted := addTraitWithFeatures(e, "Gifted",
				addingLevels(newTraitGrant(newMagery(), fxp.Two, grantmerge.Ignore)))
			gifted.Prereq = NewPrereqList()
			p := NewAttributePrereq(e)
			p.Parent = gifted.Prereq
			p.Which = IntelligenceID
			p.QualifierCriteria.Qualifier = fxp.FromInteger(13)
			gifted.Prereq.Prereqs = append(gifted.Prereq.Prereqs, p)
			return gifted, func() fxp.Int { return e.ResolveAttributeCurrent(IntelligenceID) }
		}, level: fxp.FromInteger(11)},
		{name: "technique based on the skill levels are merged into", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Brawling", fxp.One)
			disarming := NewTechnique(e, nil, "Brawling")
			disarming.Name = "Disarming"
			e.Skills = append(e.Skills, disarming)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Disarming", fxp.FromInteger(14))
			return training, func() fxp.Int { return disarming.LevelData.Level }
		}, level: fxp.FromInteger(11)},
		{name: "technique based on the parry of the skill levels are merged into", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			crossParry := NewTechnique(e, nil, "Broadsword")
			crossParry.Name = "Cross Parry"
			crossParry.TechniqueDefault.DefaultType = ParryID
			e.Skills = append(e.Skills, crossParry)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Cross Parry", fxp.Ten)
			return training, func() fxp.Int { return crossParry.LevelData.Level }
		}, level: fxp.Nine},
		{name: "parry bonus from a granted trait to a technique based on the parry", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			crossParry := NewTechnique(e, nil, "Broadsword")
			crossParry.Name = "Cross Parry"
			crossParry.TechniqueDefault.DefaultType = ParryID
			e.Skills = append(e.Skills, crossParry)
			deft := newGrantLeveledTrait("Deft", fxp.One)
			bonus := NewAttributeBonus(ParryID)
			bonus.Amount = fxp.Two
			deft.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(deft, fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Cross Parry", fxp.Ten)
			return training, func() fxp.Int { return crossParry.LevelData.Level }
		}, level: fxp.Nine},
		{name: "bonuses from a granted trait to the Dodge and Basic Speed read by a technique based on the Dodge", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			roll := NewTechnique(e, nil, "")
			roll.Name = "Roll"
			roll.TechniqueDefault = &SkillDefault{DefaultType: DodgeID} // Dodge 8, so 9 at 1 point
			e.Skills = append(e.Skills, roll)
			nimble := newGrantLeveledTrait("Nimble", fxp.One)
			// +1 to each, so Dodge 10 and Roll 11
			nimble.Features = Features{NewAttributeBonus(DodgeID), NewAttributeBonus(BasicSpeedID)}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(nimble, fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Roll", fxp.Ten)
			return training, func() fxp.Int { return roll.LevelData.Level }
		}, level: fxp.Nine},
		{name: "parry bonus from a granted trait to a skill defaulting from the parry", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			feint := addSkill(e, "Feint", fxp.One)
			parry := newSkillDefaultTo("Broadsword", "", true, 0)
			parry.DefaultType = ParryID
			feint.Defaults = []*SkillDefault{parry}
			deft := newGrantLeveledTrait("Deft", fxp.One)
			bonus := NewAttributeBonus(ParryID)
			bonus.Amount = fxp.Five
			deft.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(deft, fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Feint", fxp.Eleven)
			return training, func() fxp.Int { return feint.LevelData.Level }
		}, level: fxp.Ten},
		{name: "technique based on a skill defaulting from the skill levels are merged into", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			addSkill(e, "Broadsword", fxp.One)
			shortsword := addSkill(e, "Shortsword", fxp.One)
			shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
			disarming := NewTechnique(e, nil, "Shortsword")
			disarming.Name = "Disarming"
			e.Skills = append(e.Skills, disarming)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Disarming", fxp.FromInteger(14))
			return training, func() fxp.Int { return disarming.LevelData.Level }
		}, level: fxp.FromInteger(11)},
		{name: "bonus from a granted trait to a skill's attribute", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			agile := newGrantLeveledTrait("Agile", fxp.One)
			bonus := NewAttributeBonus(DexterityID)
			bonus.Amount = fxp.Two
			agile.Features = Features{bonus}
			training := addTraitWithFeatures(e, "Training", newTraitGrant(agile, fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "bonus from a granted trait to a spell's attribute", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			bright := newGrantLeveledTrait("Bright", fxp.One)
			bonus := NewAttributeBonus(IntelligenceID)
			bonus.Amount = fxp.Two
			bright.Features = Features{bonus}
			magery := addTraitWithFeatures(e, "Magery", newTraitGrant(bright, fxp.One, grantmerge.Ignore))
			magery.Prereq = NewPrereqList()
			p := NewSpellPrereq()
			p.Parent = magery.Prereq
			p.SamePowerSource = false
			p.QualifierCriteria.Qualifier = "Light"
			p.LevelCriteria.Compare = criteria.AtLeastNumber
			p.LevelCriteria.Qualifier = fxp.Ten
			magery.Prereq.Prereqs = append(magery.Prereq.Prereqs, p)
			return magery, func() fxp.Int { return light.LevelData.Level }
		}, level: fxp.Eight},
		{name: "default from a granted skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			shortsword := addSkill(e, "Shortsword", fxp.One)
			shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Five, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Shortsword", fxp.FromInteger(12))
			return training, func() fxp.Int { return shortsword.LevelData.Level }
		}, level: fxp.Ten},
		{name: "skills defaulting from each other, read after a third defaulting from them", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			xs := addSkill(e, "Xs", fxp.Four)
			xs.Defaults = []*SkillDefault{newSkillDefaultTo("Gs", "", true, -fxp.Three), newSkillDefaultTo("Ys", "", true, 0)}
			ys := addSkill(e, "Ys", fxp.Four)
			ys.Defaults = []*SkillDefault{newSkillDefaultTo("Gs", "", true, 0), newSkillDefaultTo("Xs", "", true, 0)}
			ws := addSkill(e, "Ws", fxp.One)
			ws.Defaults = []*SkillDefault{newSkillDefaultTo("Xs", "", true, 0)}
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Gs", fxp.One), fxp.Twenty, grantmerge.Ignore))
			training.Prereq = NewPrereqList()
			addSkillPrereq(training.Prereq, "Ws", fxp.One)
			addSkillPrereq(training.Prereq, "Ys", fxp.FromInteger(14)) // 13 at most without the grant
			return training, func() fxp.Int { return ys.LevelData.Level }
		}, level: fxp.Twelve},
		{name: "levels into a spell", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			magery := addTraitWithFeatures(e, "Magery",
				newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.Ignore))
			magery.Prereq = NewPrereqList()
			p := NewSpellPrereq()
			p.Parent = magery.Prereq
			p.SamePowerSource = false
			p.QualifierCriteria.Qualifier = "Light"
			p.LevelCriteria.Compare = criteria.AtLeastNumber
			p.LevelCriteria.Qualifier = fxp.Ten
			magery.Prereq.Prereqs = append(magery.Prereq.Prereqs, p)
			return magery, func() fxp.Int { return light.LevelData.Level }
		}, level: fxp.Eight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			carrier, level := tc.carry(c, e)
			e.Recalculate()
			c.NotEqual("", carrier.UnsatisfiedReason, "what the grant merges in can't meet its carrier's prerequisite")
			c.Equal(tc.level, level(), "so nothing is merged")
			c.False(e.Unsettled())
		})
	}
}

func TestGrantSyncsItsItemWithSource(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, TraitsExt)
	libFile := LibraryFile{Library: "Test Library", Path: "Grants" + TraitsExt}
	path := filepath.Join(useTestLibrary(t, libFile.Library), libFile.Path)
	lib := newGrantLeveledTrait("Night Vision", fxp.One)
	modTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	save := func() {
		c.NoError(SaveTraits([]*Trait{lib}, path))
		c.NoError(os.Chtimes(path, modTime, modTime))
	}
	save()
	g := newTraitGrant(lib.Clone(libFile, nil, nil, Reference), fxp.One, grantmerge.Ignore)
	key := g.Key()
	c.True(g.ItemHasSource())

	lib.Name = "Dark Vision"
	modTime = modTime.Add(time.Minute)
	save()
	g.SyncItemWithSource()
	c.Equal("Dark Vision", g.Trait.Name, "the grant's copy takes the library's changes")
	c.Equal(key, g.Key(), "and keeps the ID that ties it to the item on the sheet")

	lib.Features = Features{newTraitGrant(newGrantLeveledTrait("Nested", fxp.One), fxp.One, grantmerge.Ignore)}
	modTime = modTime.Add(time.Minute)
	save()
	g.SyncItemWithSource()
	c.Equal(0, len(g.Trait.Features), "a grant the library's item gained is dropped, as a granted item grants nothing")

	c.False(newTraitGrant(newGrantLeveledTrait("Blank", fxp.One), fxp.One, grantmerge.Ignore).ItemHasSource(),
		"a copy not from a library has nothing to sync with")
}

func TestGrantHashCoversWholeGrant(t *testing.T) {
	c := check.New(t)
	skill := newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore)
	before := Hash64(skill)
	skill.Amount = fxp.Two
	c.NotEqual(before, Hash64(skill), "a change to the amount changes the grant")
	before = Hash64(skill)
	skill.Adjust = grantadjust.SetLevel
	c.NotEqual(before, Hash64(skill), "as does one to the kind of adjustment")
	before = Hash64(skill)
	skill.Merge = grantmerge.KeepLowest
	c.NotEqual(before, Hash64(skill), "or the merge mode")

	g := newTraitGrant(newGrantLeveledTrait("Bad Temper", fxp.One), fxp.One, grantmerge.Ignore)
	g.Trait.SelfControl = selfctrl.CR12
	before = Hash64(g)
	g.Trait.SelfControl = selfctrl.CR9
	c.NotEqual(before, Hash64(g), "a change to the item's self-control roll changes the grant")
	before = Hash64(g)
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "Tough Skin"
	g.Trait.Modifiers = []*TraitModifier{mod}
	c.NotEqual(before, Hash64(g), "as does a modifier it gains")
	before = Hash64(g)
	g.Trait.UserDesc = "Flies into a rage"
	c.NotEqual(before, Hash64(g), "or a change to its user description")
}

func TestGrantedSpellRitualsWhileNotInEffect(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.ShowSpellAdj = true
	goggles := addCarriedEquipmentWithFeatures(e, "Amulet",
		newSpellGrant(newHardSpell("Light", fxp.One), fxp.Three, grantmerge.Ignore))
	e.Recalculate()
	light := grantedOf(e.Spells)[0]
	c.Equal("**Ritual:** speak quietly and make a gesture", light.Rituals(), "a Hard spell at 4 points is IQ+0")

	goggles.Equipped = false
	e.Recalculate()
	c.True(light.grantNotInEffect())
	c.Equal("**Ritual:** speak quietly and make a gesture", light.Rituals(),
		"not in effect, the rituals are those of the level the grant would give")
}

func TestGrantedItemShowsNoLibrarySource(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training",
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore))
	e.Recalculate()
	var data CellData
	grantedOf(e.Traits)[0].CellData(TraitLibSrcColumn, &data)
	c.Equal("", data.Primary, "a granted item's grant is its source")
	c.Equal("", data.Tooltip)
}

func TestGrantWording(t *testing.T) {
	c := check.New(t)
	plain := func(s string) string { return s }
	c.Equal("Grants a skill, not yet chosen", NewGrant(feature.GrantSkill).Describe(nil, nil, plain))
	acuteVision := func(amount fxp.Int, merge grantmerge.Mode) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), amount, merge)
	}
	brawling := func(adjust grantadjust.Kind, amount fxp.Int, merge grantmerge.Mode) *Grant {
		return newSkillGrantOf(newEasySkill("Brawling", fxp.One), adjust, amount, merge)
	}
	wildcard := brawling(grantadjust.AddPoints, fxp.Two, grantmerge.Ignore)
	wildcard.Skill.Difficulty.Difficulty = difficulty.Wildcard
	checked := brawling(grantadjust.SetLevel, fxp.Ten, grantmerge.Ignore)
	checked.CheckPrereqs = true
	for _, tc := range []struct {
		grant *Grant
		// label is how a summary names what the grant gives, describe is the grant's description, merged is what it
		// says of an item already present, and idle, when not empty, what it says of an item absent that it can't put
		// on the sheet.
		label, describe, merged, idle string
	}{
		{
			grant: acuteVision(fxp.Two, grantmerge.Ignore), label: "Acute Vision 2",
			describe: "Grants the trait Acute Vision at level 2, unless already present",
			merged:   "Does not grant Acute Vision 2, since it is already present",
		},
		{
			grant: addingLevels(acuteVision(fxp.FromStringForced("1.5"), grantmerge.Ignore)), label: "Acute Vision 1.5",
			describe: "Grants the trait Acute Vision at level 1.5, or adds 1.5 levels to it when already present",
			merged:   "Adds 1.5 levels to Acute Vision, already present",
		},
		{
			grant:    addingLevels(acuteVision(-fxp.One, grantmerge.Ignore)),
			describe: "Removes 1 level from the trait Acute Vision when already present",
			merged:   "Removes 1 level from Acute Vision, already present",
			idle:     "Removes nothing from Acute Vision, since it is not present",
		},
		{
			grant:    addingLevels(acuteVision(fxp.Half, grantmerge.Ignore)),
			describe: "Would grant the trait Acute Vision at level 0.5, but a level below 1 grants nothing",
			merged:   "Does not change Acute Vision, already present, since a level below 1 grants nothing",
			idle:     "Does not grant Acute Vision 0.5, since a level below 1 grants nothing",
		},
		{
			grant:    acuteVision(fxp.Two, grantmerge.KeepHighest),
			describe: "Grants the trait Acute Vision at level 2, or keeps the higher of the two levels when already present",
			merged:   "Keeps the higher level of Acute Vision, as it is or at level 2",
		},
		{
			grant:    acuteVision(fxp.Two, grantmerge.KeepLowest),
			describe: "Grants the trait Acute Vision at level 2, or keeps the lower of the two levels when already present",
			merged:   "Keeps the lower level of Acute Vision, as it is or at level 2",
		},
		{
			grant: brawling(grantadjust.AddLevels, fxp.Two, grantmerge.Ignore), label: "Brawling at level 2",
			describe: "Grants the skill Brawling at level 2, or adds 2 levels to it when already present",
			merged:   "Adds 2 levels to Brawling, already present",
		},
		{
			grant:    brawling(grantadjust.AddLevels, -fxp.One, grantmerge.Ignore),
			describe: "Removes 1 level from the skill Brawling when already present",
			merged:   "Removes 1 level from Brawling, already present",
			idle:     "Removes nothing from Brawling, since it is not present",
		},
		{
			grant: brawling(grantadjust.AddPoints, fxp.One, grantmerge.Ignore), label: "Brawling with 1 point",
			describe: "Grants the skill Brawling with 1 point, or adds 1 point to it when already present",
			merged:   "Adds 1 point to Brawling, already present",
		},
		{
			grant:    brawling(grantadjust.AddPoints, -fxp.Two, grantmerge.Ignore),
			describe: "Removes 2 points from the skill Brawling when already present",
			merged:   "Removes 2 points from Brawling, already present",
			idle:     "Removes nothing from Brawling, since it is not present",
		},
		{
			// Points too few to buy a level still add to a skill already present.
			grant:    brawling(grantadjust.AddPoints, fxp.Half, grantmerge.Ignore),
			describe: "Adds 0.5 points to the skill Brawling when already present, though too few to grant it",
			merged:   "Adds 0.5 points to Brawling, already present",
			idle:     "Does not grant Brawling with 0.5 points, since points too few to buy a level grant nothing",
		},
		{
			grant:    wildcard,
			describe: "Adds 2 points to the skill Brawling when already present, though too few to grant it",
			merged:   "Adds 2 points to Brawling, already present",
		},
		{
			grant: newSkillGrant(newDisarming(0), fxp.Two, grantmerge.Ignore), label: "Disarming at level 2",
			describe: "Grants the technique Disarming at level 2, or adds 2 levels to it when already present",
			merged:   "Adds 2 levels to Disarming, already present",
		},
		{
			// 1 point buys a Hard technique nothing above its default.
			grant:    newSkillGrantOf(newDisarming(0), grantadjust.AddPoints, fxp.One, grantmerge.Ignore),
			describe: "Adds 1 point to the technique Disarming when already present, though too few to grant it",
			merged:   "Adds 1 point to Disarming, already present",
			idle:     "Does not grant Disarming with 1 point, since points too few to buy a level grant nothing",
		},
		{
			grant:    brawling(grantadjust.AddPoints, 0, grantmerge.Ignore),
			describe: "Would grant the skill Brawling with 0 points, but points too few to buy a level grant nothing",
			merged:   "Does not change Brawling, already present, since points too few to buy a level grant nothing",
			idle:     "Does not grant Brawling with 0 points, since points too few to buy a level grant nothing",
		},
		{
			grant: brawling(grantadjust.SetLevel, fxp.Ten, grantmerge.KeepHighest), label: "Brawling at skill level 10",
			describe: "Grants the skill Brawling at skill level 10, or keeps the higher of the two levels when already " +
				"present",
			merged: "Keeps the higher level of Brawling, as it is or at skill level 10",
		},
		{
			grant: checked,
			describe: "Grants the skill Brawling at skill level 10, unless already present, checking its " +
				"prerequisites",
			merged: "Does not grant Brawling at skill level 10, since it is already present",
		},
		{
			grant: brawling(grantadjust.SetLevel, fxp.Ten, grantmerge.KeepLowest),
			describe: "Grants the skill Brawling at skill level 10, or keeps the lower of the two levels when already " +
				"present",
			merged: "Keeps the lower level of Brawling, as it is or at skill level 10",
		},
		{
			// A negative level set takes nothing away.
			grant:    brawling(grantadjust.SetLevel, -fxp.Three, grantmerge.KeepHighest),
			describe: "Would grant the skill Brawling at skill level -3, but a skill level below 1 grants nothing",
			merged:   "Does not change Brawling, already present, since a skill level below 1 grants nothing",
			idle:     "Does not grant Brawling at skill level -3, since a skill level below 1 grants nothing",
		},
	} {
		g := tc.grant
		c.Equal(tc.describe, g.Describe(nil, nil, plain))
		if tc.label != "" {
			c.Equal(tc.label, g.itemLabel(nil))
		}
		c.Equal(tc.merged, g.mergeSummary(g.itemLabel(nil), nil, grantActive), tc.describe)
		if tc.idle != "" {
			c.Equal(tc.idle, g.idleSummary(nil), tc.describe)
		}
	}
}

func TestGrantTiers(t *testing.T) {
	c := check.New(t)
	skill := newHardSkill("Alchemy", 0)
	wildcard := newHardSkill("Alchemy", 0)
	wildcard.Difficulty.Difficulty = difficulty.Wildcard
	// A technique's levels count from its default, at a point each, plus 1 for a Hard one, as do a Ritual Magic
	// spell's.
	average := newTechniqueOf("Karate", "Kicking", 0)
	hard := newKicking(0)
	ritual := newBless(0)
	for _, tc := range []struct {
		name   string
		item   grantLeveler
		tier   fxp.Int
		points fxp.Int
	}{
		{name: "skill tier 0", item: skill, tier: 0, points: 0},
		{name: "skill tier 1", item: skill, tier: fxp.One, points: fxp.One},
		{name: "skill tier 2", item: skill, tier: fxp.Two, points: fxp.Two},
		{name: "skill tier 3", item: skill, tier: fxp.Three, points: fxp.Four},
		{name: "skill tier 4", item: skill, tier: fxp.Four, points: fxp.Eight},
		{name: "wildcard tier 3", item: wildcard, tier: fxp.Three, points: fxp.FromInteger(12)},
		{name: "Average technique tier 3", item: average, tier: fxp.Three, points: fxp.Three},
		{name: "Hard technique tier 1", item: hard, tier: fxp.One, points: fxp.Two},
		{name: "Hard technique tier 4", item: hard, tier: fxp.Four, points: fxp.Five},
		{name: "Hard Ritual Magic spell tier 2", item: ritual, tier: fxp.Two, points: fxp.Three},
	} {
		c.Equal(tc.points, tc.item.pointsForLevels(tc.tier), tc.name)
	}
}

func TestGrantSetLevelPoints(t *testing.T) {
	hard := newHardSkill("Alchemy", 0)
	wildcard := newHardSkill("Alchemy", 0)
	wildcard.Difficulty.Difficulty = difficulty.Wildcard
	specialized := newHardSkill("Alchemy", 0)
	specialized.OptionalSpecialization = "Potions"
	unresolved := newHardSkill("Alchemy", 0)
	unresolved.Difficulty.Attribute = "missing"
	spell := newHardSpell("Light", 0)
	kicking := newKicking(0)
	averageKicking := newKicking(0)
	averageKicking.Difficulty.Difficulty = difficulty.Average
	sanity := newKicking(0)
	sanity.TechniqueDefault = &SkillDefault{DefaultType: "sanity", Modifier: -fxp.Two}
	bless := newBless(0)
	// Light has no prerequisites, so that it rests on Ritual Magic at +0, which it goes no higher than, and capped
	// Kicking goes no higher than Karate-3, below its default.
	light := newBless(0)
	light.Name = "Light"
	light.PrereqCount = 0
	capped := newKicking(0)
	limit := -fxp.Three
	capped.TechniqueLimitModifier = &limit
	for _, tc := range []struct {
		name string
		item grantLeveler
		// halfStat, when set, has the sheet use half-stat defaults at IQ 14, which halves IQ to 12 for a skill; IQ is
		// 10 otherwise, where Karate and Ritual Magic, both IQ/Hard at 4 points, are at 10.
		halfStat bool
		level    fxp.Int
		points   fxp.Int
	}{
		{name: "below what 1 point buys", item: hard, level: fxp.Seven, points: 0},
		{name: "1 point", item: hard, level: fxp.Eight, points: fxp.One},
		{name: "2 points", item: hard, level: fxp.Nine, points: fxp.Two},
		{name: "4 points", item: hard, level: fxp.Ten, points: fxp.Four},
		{name: "8 points", item: hard, level: fxp.Eleven, points: fxp.Eight},
		{name: "12 points", item: hard, level: fxp.Twelve, points: fxp.Twelve},
		{name: "wildcard", item: wildcard, level: fxp.Eight, points: fxp.Six},
		{
			name: "optional specialization, which makes the skill Average", item: specialized, level: fxp.Ten,
			points: fxp.Two,
		},
		{name: "attribute missing", item: unresolved, level: fxp.Ten, points: 0},
		{name: "spell", item: spell, level: fxp.Ten, points: fxp.Four},
		{name: "skill at half-stat", item: hard, halfStat: true, level: fxp.Twelve, points: fxp.Four},
		{
			name: "spell at half-stat, which spells don't use", item: spell, halfStat: true, level: fxp.Twelve,
			points: fxp.One,
		},
		// A technique's level is priced from the skill it rests on, at a point a level above its default, plus 1 for a
		// Hard one, and no higher than its limit.
		{name: "Hard technique", item: kicking, level: fxp.Nine, points: fxp.Two},
		{name: "Average technique", item: averageKicking, level: fxp.Nine, points: fxp.One},
		{name: "technique above its limit", item: kicking, level: fxp.Twelve, points: fxp.Three},
		{name: "technique resting on a missing attribute", item: sanity, level: fxp.Nine, points: 0},
		// A Ritual Magic spell's is priced from its ritual skill in the same way.
		{name: "ritual magic", item: bless, level: fxp.Nine, points: fxp.Three},
		{name: "ritual magic above its ritual skill", item: bless, level: fxp.Twelve, points: fxp.Four},
		// One whose limit is at or below its default is priced at the points of its first level, which buy the limit,
		// for a level at or above the limit.
		{name: "ritual magic at its default", item: light, level: fxp.Ten, points: fxp.Two},
		{name: "ritual magic above its default", item: light, level: fxp.Twelve, points: fxp.Two},
		{name: "ritual magic below its default", item: light, level: fxp.Nine, points: 0},
		{name: "technique at a limit below its default", item: capped, level: fxp.Seven, points: fxp.Two},
		{name: "technique above a limit below its default", item: capped, level: fxp.Twelve, points: fxp.Two},
		{name: "technique below a limit below its default", item: capped, level: fxp.Six, points: 0},
	} {
		c := check.New(t)
		e := NewEntity()
		if tc.halfStat {
			e.SheetSettings.UseHalfStatDefaults = true
			e.Attributes.Set[IntelligenceID].Adjustment = fxp.Four
		}
		addHardSkill(e, "Karate", fxp.Four)
		addHardSkill(e, "Ritual Magic", fxp.Four)
		e.Recalculate()
		c.Equal(tc.points, tc.item.pointsForBaseLevel(e, tc.level), tc.name)
	}
}

func TestGrantSpellMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		merge  grantmerge.Mode
		adjust grantadjust.Kind
		// points is what the spell bought has; with none, it has no level. ritual, when set, makes the spell Bless,
		// which rests on Ritual Magic at 10, 7 with no points, rather than Light.
		points fxp.Int
		amount fxp.Int
		level  fxp.Int
		ritual bool
	}{
		{
			name: "ignore a level set", merge: grantmerge.Ignore, adjust: grantadjust.SetLevel, points: fxp.One,
			amount: fxp.Ten, level: fxp.Eight,
		},
		{name: "add levels", merge: grantmerge.Ignore, points: fxp.One, amount: fxp.Two, level: fxp.Ten},
		{name: "add levels without level", merge: grantmerge.Ignore, amount: fxp.Two, level: fxp.Nine},
		{
			name: "keep highest of a level set", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel, points: fxp.One,
			amount: fxp.Nine, level: fxp.Nine,
		},
		{
			name: "keep highest of a level set without level", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel,
			amount: fxp.Ten, level: fxp.Ten,
		},
		{
			name: "keep lowest of a level set", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			points: fxp.Four, amount: fxp.Nine, level: fxp.Nine,
		},
		{
			name: "keep lowest of a level set higher", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			points: fxp.One, amount: fxp.Ten, level: fxp.Eight,
		},
		{
			name: "keep lowest of a level set without level", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			amount: fxp.Ten, level: fxp.Ten,
		},
		{
			name: "keep highest of a ritual spell's level set", merge: grantmerge.KeepHighest, adjust: grantadjust.SetLevel,
			points: fxp.Two, amount: fxp.Nine, level: fxp.Nine, ritual: true,
		},
		{
			name: "keep lowest of a ritual spell's level set", merge: grantmerge.KeepLowest, adjust: grantadjust.SetLevel,
			points: fxp.Three, amount: fxp.Eight, level: fxp.Eight, ritual: true,
		},
		{
			// The levels added go no higher than the ritual skill.
			name: "add levels to a ritual spell", merge: grantmerge.Ignore, points: fxp.Two, amount: fxp.Four,
			level: fxp.Ten, ritual: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			spell, item := newHardSpell("Light", tc.points), newHardSpell("Light", fxp.One)
			if tc.ritual {
				addHardSkill(e, "Ritual Magic", fxp.Four)
				spell, item = newBless(tc.points), newBless(0)
			}
			spell.SetDataOwner(e)
			e.Spells = append(e.Spells, spell)
			addTraitWithFeatures(e, "Magery", newSpellGrantOf(item, tc.adjust, tc.amount, tc.merge))
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Spells)))
			c.Equal(tc.level, spell.LevelData.Level)
		})
	}
}

func TestGrantPutsSpellOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Magery", newSpellGrant(newHardSpell("Light", fxp.One), fxp.Three, grantmerge.Ignore))
	e.Recalculate()
	list := grantedOf(e.Spells)
	c.Equal(1, len(list))
	c.Equal(fxp.Ten, list[0].LevelData.Level, "a Hard spell at 4 points is IQ+0")
	c.Equal(NumericRangeOf(0), list[0].PointsRange(nil))
	c.Equal(fxp.Int(0), e.PointsBreakdown().Spells)
	c.Equal("0", ResolveScript(e, deferredNewScriptSpell(list[0]), "self.points"), "a granted spell costs nothing to a script")
}

func TestGrantKeepComparesWithOtherMerges(t *testing.T) {
	badSight := func(level fxp.Int, merge grantmerge.Mode) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Bad Sight", fxp.One), level, merge)
	}
	for _, tc := range []struct {
		name string
		// grants merge into Bad Sight bought at level 3.
		grants []*Grant
		level  fxp.Int
	}{
		{name: "keep highest after a merge", grants: []*Grant{
			addingLevels(badSight(fxp.One, grantmerge.Ignore)), badSight(fxp.Five, grantmerge.KeepHighest),
		}, level: fxp.Five},
		{name: "keep lowest after a merge", grants: []*Grant{
			addingLevels(badSight(fxp.Two, grantmerge.Ignore)), badSight(fxp.Four, grantmerge.KeepLowest),
		}, level: fxp.Four},
		// A Keep Lowest grant caps what a Keep Highest one raised.
		{name: "keep lowest after keep highest", grants: []*Grant{
			badSight(fxp.Four, grantmerge.KeepLowest), badSight(fxp.Five, grantmerge.KeepHighest),
		}, level: fxp.Four},
	} {
		for _, reversed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s reversed=%t", tc.name, reversed), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				bad := addTrait(e, newGrantLeveledTrait("Bad Sight", fxp.One))
				bad.Levels = fxp.Three
				for i, g := range tc.grants {
					addTraitWithFeatures(e, fmt.Sprintf("Carrier %d", i), g.Clone())
				}
				if reversed {
					slices.Reverse(e.Traits[1:])
				}
				e.Recalculate()
				c.Equal(tc.level, bad.CurrentLevel(), "whichever comes first")
			})
		}
	}
}

func TestGrantKeepHighestHoldsWithoutTheOtherMerges(t *testing.T) {
	for _, tc := range []struct {
		name string
		// carry puts on the sheet an item the user bought, a carrier whose grant merges a level into it and requires
		// it at the level a Keep Highest grant carried by another trait holds it at, and that other trait, and returns
		// the first carrier along with what reads the item's level.
		carry func(c check.Checker, e *Entity) (carrier *Trait, level func() fxp.Int)
		level fxp.Int
	}{
		{name: "trait", carry: func(c check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			training := addTraitWithFeatures(e, "Training",
				addingLevels(newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.Ignore)))
			training.Prereq = newPrereqListForTrait("Magery", true)
			traitPrereqOf(c, training.Prereq).LevelCriteria.Qualifier = fxp.Three
			addTraitWithFeatures(e, "Gifted",
				newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.Three, grantmerge.KeepHighest))
			return training, magery.CurrentLevel
		}, level: fxp.Three},
		{name: "skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			addTraitWithFeatures(e, "Gifted", newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel,
				fxp.Twelve, grantmerge.KeepHighest))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.FromInteger(12)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			carrier, level := tc.carry(c, e)
			e.Recalculate()
			c.False(e.Unsettled())
			c.Equal("", carrier.UnsatisfiedReason, "the level kept is there without the carrier's own merge")
			c.Equal(tc.level, level())
		})
	}
}

func TestGrantMergesDoNotDependOnCarrierOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		// bought is the points spent on Brawling, and grants the two grants of it that merge into it.
		bought fxp.Int
		grants func() [2]*Grant
		points fxp.Int
		level  fxp.Int
	}{
		{name: "points added and a curse", bought: fxp.One, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, fxp.Four, grantmerge.Ignore),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
			}
		}, points: fxp.Four, level: fxp.FromInteger(12)},
		{name: "levels and points into a skill with no level", bought: 0, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, fxp.One, grantmerge.Ignore),
			}
		}, points: fxp.Five, level: fxp.FromInteger(12)},
		// The first grant to give a skill with no level of its own its points gives it its level, and the other then
		// merges into that.
		{name: "two keep lowest into a skill with no level", bought: 0, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Ten,
					grantmerge.KeepLowest),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Ten,
					grantmerge.KeepLowest),
			}
		}, points: fxp.One, level: fxp.Ten},
		{name: "keep highest and keep lowest into a skill with no level", bought: 0, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve,
					grantmerge.KeepHighest),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve,
					grantmerge.KeepLowest),
			}
		}, points: fxp.Four, level: fxp.Twelve},
		// A Keep Highest grant keeps its level when the points taken away leave the skill with no level.
		{name: "keep highest and points taken away", bought: fxp.One, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve,
					grantmerge.KeepHighest),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
			}
		}, points: 0, level: fxp.Twelve},
		{name: "two curses", bought: fxp.Eight, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
			}
		}, points: fxp.Six, level: fxp.Twelve},
		{name: "curses of different sizes", bought: fxp.Eight, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.One, grantmerge.Ignore),
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.AddPoints, -fxp.Two, grantmerge.Ignore),
			}
		}, points: fxp.Five, level: fxp.Twelve},
	} {
		for _, reversed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s reversed=%t", tc.name, reversed), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				brawling := addSkill(e, "Brawling", tc.bought)
				grants := tc.grants()
				if reversed {
					grants[0], grants[1] = grants[1], grants[0]
				}
				addTraitWithFeatures(e, "First", grants[0])
				addTraitWithFeatures(e, "Second", grants[1])
				e.Recalculate()
				c.Equal(tc.points, brawling.AdjustedPoints(nil))
				c.Equal(tc.level, brawling.LevelData.Level)
				if tc.level != fxp.Min {
					c.Equal(tc.level-fxp.Ten, brawling.LevelData.RelativeLevel, "the level is relative to DX 10")
				}
			})
		}
	}
}

func TestGrantPlanterDoesNotDependOnCarrierOrder(t *testing.T) {
	nightVision := func(level fxp.Int, merge grantmerge.Mode) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), level, merge)
	}
	// alchemy is a grant of an IQ/Hard skill, whose first point buys IQ-2, or 8.
	alchemy := func(adjust grantadjust.Kind, amount fxp.Int, merge grantmerge.Mode) *Grant {
		return newSkillGrantOf(newHardSkill("Alchemy", 0), adjust, amount, merge)
	}
	for _, tc := range []struct {
		name string
		// grants are the grants of the item, a trait or skill, none of which finds it bought. unmet, when not empty, is
		// the name of the carrier whose prerequisites are unmet, so that its grant is not in effect. selfControl, when
		// set, is the self-control roll of the row.
		grants      []*Grant
		unmet       string
		level       fxp.Int
		selfControl selfctrl.Roll
	}{
		{name: "ignore and levels added", grants: []*Grant{
			nightVision(fxp.Two, grantmerge.Ignore), addingLevels(nightVision(fxp.One, grantmerge.Ignore)),
		}, level: fxp.Three},
		{name: "two ignores", grants: []*Grant{
			nightVision(fxp.One, grantmerge.Ignore), nightVision(fxp.Two, grantmerge.Ignore),
		}, level: fxp.Two},
		{name: "ignore and keep highest", grants: []*Grant{
			nightVision(fxp.Three, grantmerge.Ignore), nightVision(fxp.Two, grantmerge.KeepHighest),
		}, level: fxp.Three},
		{name: "keep highest and levels added", grants: []*Grant{
			nightVision(fxp.One, grantmerge.KeepHighest), addingLevels(nightVision(fxp.Two, grantmerge.Ignore)),
		}, level: fxp.Three},
		{name: "keep highest and keep lowest", grants: []*Grant{
			nightVision(fxp.Two, grantmerge.KeepLowest), nightVision(fxp.Three, grantmerge.KeepHighest),
		}, level: fxp.Two},
		// A grant that keeps a level still keeps it once it has put the item there, as it does when it merges into
		// the row another grant put there, so a cap holds the levels added to it and a floor the levels taken away.
		{name: "keep lowest and levels added", grants: []*Grant{
			nightVision(fxp.Two, grantmerge.KeepLowest), addingLevels(nightVision(fxp.One, grantmerge.Ignore)),
		}, level: fxp.Two},
		{name: "keep highest and levels taken away", grants: []*Grant{
			nightVision(fxp.Three, grantmerge.KeepHighest), addingLevels(nightVision(-fxp.Two, grantmerge.Ignore)),
		}, level: fxp.Three},
		{name: "level set and levels added", grants: []*Grant{
			alchemy(grantadjust.AddLevels, fxp.One, grantmerge.Ignore),
			alchemy(grantadjust.SetLevel, fxp.Ten, grantmerge.Ignore),
		}, level: fxp.Eleven},
		// Levels added rank above points added, which then add to the 1 point the level gives: 5 buy IQ+0.
		{name: "levels and points added", grants: []*Grant{
			alchemy(grantadjust.AddPoints, fxp.Four, grantmerge.Ignore),
			alchemy(grantadjust.AddLevels, fxp.One, grantmerge.Ignore),
		}, level: fxp.Ten},
		{name: "levels set to keep highest and keep lowest", grants: []*Grant{
			alchemy(grantadjust.SetLevel, fxp.Ten, grantmerge.KeepLowest),
			alchemy(grantadjust.SetLevel, fxp.Eleven, grantmerge.KeepHighest),
		}, level: fxp.Ten},
		// A level set ranks above points added, whatever the level those buy, and the points then add to the 4 the
		// level set gives: 16 buy IQ+3.
		{name: "points and a level set", grants: []*Grant{
			alchemy(grantadjust.AddPoints, fxp.Twelve, grantmerge.Ignore),
			alchemy(grantadjust.SetLevel, fxp.Ten, grantmerge.Ignore),
		}, level: fxp.Thirteen},
		{name: "stand-in for an ignore not in effect", grants: []*Grant{
			nightVision(fxp.Three, grantmerge.Ignore), nightVision(fxp.Two, grantmerge.KeepHighest),
			addingLevels(nightVision(fxp.One, grantmerge.Ignore)),
		}, unmet: "Carrier 0", level: fxp.Three},
		{name: "checking the item's prerequisites and not", grants: func() []*Grant {
			needy := func(checkPrereqs bool) *Grant {
				g := nightVision(fxp.One, grantmerge.Ignore)
				g.Trait.Prereq = newPrereqListRequiringSkill("Brawling", fxp.One)
				g.CheckPrereqs = checkPrereqs
				return g
			}
			return []*Grant{needy(true), needy(false)}
		}(), level: fxp.One},
		{name: "self-control rolls the item's hash leaves out", grants: func() []*Grant {
			badTemper := func(cr selfctrl.Roll) *Grant {
				g := nightVision(fxp.One, grantmerge.Ignore)
				g.Trait.Name = "Bad Temper"
				g.Trait.SelfControl = cr
				return g
			}
			return []*Grant{badTemper(selfctrl.CR12), badTemper(selfctrl.CR6)}
		}(), level: fxp.One, selfControl: selfctrl.CR6},
	} {
		for _, reversed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s reversed=%t", tc.name, reversed), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				for i, g := range tc.grants {
					name := fmt.Sprintf("Carrier %d", i)
					carrier := addTraitWithFeatures(e, name, g.Clone())
					if name == tc.unmet {
						carrier.Prereq = newPrereqListRequiringSkill("Brawling", fxp.One)
					}
				}
				if reversed {
					slices.Reverse(e.Traits)
				}
				e.Recalculate()
				c.False(e.Unsettled())
				traits, skills := grantedOf(e.Traits), grantedOf(e.Skills)
				if len(traits)+len(skills) != 1 {
					c.Fatalf("expected one granted row, got %d", len(traits)+len(skills))
				}
				if len(skills) == 1 {
					c.True(skills[0].Enabled())
					c.Equal(tc.level, skills[0].LevelData.Level)
					return
				}
				c.True(traits[0].Enabled())
				c.Equal(tc.level, traits[0].CurrentLevel())
				c.Equal("", traits[0].UnsatisfiedReason, "a grant that treats the item's prerequisites as met wins")
				if tc.selfControl != selfctrl.None {
					c.Equal(tc.selfControl, traits[0].SelfControl, "the grant of the item worth more puts it on the sheet")
				}
			})
		}
	}
}

func TestGrantedRowStaysWithTheGrantHoldingIt(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	nightVision := func() *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore)
	}
	aGrant, bGrant := nightVision(), nightVision()
	a := addTraitWithFeatures(e, "A", aGrant)
	b := addTraitWithFeatures(e, "B", bGrant)
	e.Recalculate()
	granted := grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row, got %d", len(granted))
	}
	row := granted[0]
	e.Traits = []*Trait{b, a, row}
	e.Recalculate()
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row, got %d", len(granted))
	}
	c.True(granted[0] == row, "reordering the carriers of equal grants leaves the row with the grant that put it there")

	group := NewTrait(e, nil, true)
	group.Name = "Senses"
	group.Children = []*Trait{row}
	row.SetParent(group)
	e.Traits = []*Trait{b, a, group}
	a.Disabled = true
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "disabling the carrier whose grant put the row there adds no row at the top")
	granted = grantedOf(group.Children)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row in the container, got %d", len(granted))
	}
	c.Equal(row.TID, granted[0].TID, "the other carrier's grant takes over the row where the user put it")
	row = granted[0]
	a.Disabled = false
	e.Recalculate()
	granted = grantedOf(group.Children)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row in the container, got %d", len(granted))
	}
	c.True(granted[0] == row, "enabling the other carrier again leaves the row with the grant that put it there")

	row.SwitchedOn = true
	aGrant.Amount = fxp.Three
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "a grant that comes to sort first adds no row at the top")
	granted = grantedOf(group.Children)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row in the container, got %d", len(granted))
	}
	c.Equal(row.TID, granted[0].TID, "it takes over the row where the user put it")
	c.True(granted[0].SwitchedOn, "along with its switch")
	c.True(GrantingItem(granted[0]) == a)
	c.Equal(fxp.Three, granted[0].CurrentLevel())

	aGrant.Trait.Name = "Dark Vision"
	e.Recalculate()
	granted = grantedOf(group.Children)
	if len(granted) != 1 {
		c.Fatalf("expected one Dark Vision row in the container, got %d", len(granted))
	}
	c.Equal("Dark Vision", granted[0].Name)
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected the other grant to put Night Vision back at the top, not to take over the row the "+
			"renamed grant has already refreshed, got %d rows", len(granted))
	}
	c.Equal("Night Vision", granted[0].Name)
	c.True(GrantingItem(granted[0]) == b)
	c.Equal(fxp.Two, granted[0].CurrentLevel())

	aGrant.Trait.Name = "Night Vision"
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "renamed back, the grant sorting first keeps the row where the user put it")
	row = grantedOf(group.Children)[0]
	aGrant.Amount = fxp.One
	aGrant.Trait.PageRef = "B71"
	bGrant.Trait.PageRef = "B71"
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)),
		"a grant that comes to sort first as the item of both changes adds no row at the top")
	granted = grantedOf(group.Children)
	if len(granted) != 1 {
		c.Fatalf("expected one Night Vision row in the container, got %d", len(granted))
	}
	c.Equal(row.TID, granted[0].TID, "its row takes the place of the one the user put there")
	c.True(granted[0].SwitchedOn, "along with its switch")
	c.True(GrantingItem(granted[0]) == b)

	granted[0].SetDisabled(true)
	cGrant := nightVision()
	cGrant.Trait.PageRef = "B71"
	third := addTraitWithFeatures(e, "C", cGrant)
	e.Recalculate()
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected C to put a row at the top, B's being disabled, got %d rows", len(granted))
	}
	enabled := granted[0]
	c.True(GrantingItem(enabled) == third)
	e.Traits = slices.DeleteFunc(e.Traits, func(one *Trait) bool { return one == b || one == third })
	e.Recalculate()
	c.Equal(0, len(grantedOf(group.Children)), "A, merging into C's row, does not take over the row the user disabled")
	granted = grantedOf(e.Traits)
	if len(granted) != 1 {
		c.Fatalf("expected A to take over C's row at the top, got %d rows there", len(granted))
	}
	c.Equal(enabled.TID, granted[0].TID)
	c.True(GrantingItem(granted[0]) == a)
	c.True(granted[0].Enabled(), "so A's grant still gives its item")
}

func TestGrantedRowsFollowTheirGrantsInSheetOrder(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Racial",
		newTraitGrant(newGrantLeveledTrait("Acute Hearing", fxp.Two), fxp.One, grantmerge.Ignore),
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Five, grantmerge.Ignore),
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.Ignore),
		newSkillGrant(newEasySkill("Karate", fxp.One), fxp.One, grantmerge.Ignore))
	addTraitWithFeatures(e, "Training",
		newTraitGrant(newGrantLeveledTrait("Fit", fxp.Five), fxp.Ten, grantmerge.Ignore))
	e.Recalculate()
	grantedTraits, grantedSkills := grantedOf(e.Traits), grantedOf(e.Skills)
	traits := make([]string, 0, len(grantedTraits))
	for _, one := range grantedTraits {
		traits = append(traits, one.Name)
	}
	skills := make([]string, 0, len(grantedSkills))
	for _, one := range grantedSkills {
		skills = append(skills, one.Name)
	}
	c.Equal([]string{"Acute Hearing", "Night Vision", "Fit"}, traits, "carriers in sheet order, each in its own order")
	c.Equal([]string{"Brawling", "Karate"}, skills, "whatever the grants' merge modes")
}

func TestGrantedSkillSatisfiesTechnique(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	kicking := newKicking(0)
	kicking.TechniqueDefault.Name.Qualifier = "Brawling"
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore),
		newSkillGrantOf(kicking, grantadjust.SetLevel, fxp.Eleven, grantmerge.Ignore))
	disarming := NewTechnique(e, nil, "Brawling")
	disarming.Name = "Disarming"
	e.Skills = append(e.Skills, disarming)
	e.Recalculate()
	c.False(e.Unsettled())
	c.Equal(fxp.FromInteger(13), disarming.LevelData.Level, "the technique takes its level from the granted skill")
	c.Equal("", disarming.UnsatisfiedReason, "which the points its grant gives it satisfy, as they give it its level")
	var granted *Skill
	for _, one := range grantedOf(e.Skills) {
		if one.IsTechnique() {
			granted = one
		}
	}
	if granted == nil {
		c.Fatal("expected the grant to put Kicking on the sheet")
	}
	c.Equal(fxp.Eleven, granted.LevelData.Level,
		"a level set on a granted technique is priced from what the grants give the skill it rests on, Brawling 12")
}

func TestGrantPricesTechniqueFromTheTechniqueItRestsOn(t *testing.T) {
	// Finisher rests on the technique Arm Lock, which rests on Karate at 10 and gets 3 levels from a grant, so that a
	// level set on Finisher is priced from Arm Lock with them, whether Arm Lock is the sheet's own, at 11 with its point,
	// or the grant's, though a level set comes ahead of levels added in plan order.
	for _, tc := range []struct {
		name    string
		bought  bool
		armLock fxp.Int
		points  fxp.Int
	}{
		{name: "bought", bought: true, armLock: fxp.FromInteger(14), points: fxp.One},
		{name: "granted", armLock: fxp.Thirteen, points: fxp.Two},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			addHardSkill(e, "Karate", fxp.Four)
			if tc.bought {
				armLock := newTechniqueOf("Karate", "Arm Lock", fxp.One)
				armLock.SetDataOwner(e)
				e.Skills = append(e.Skills, armLock)
			}
			addTraitWithFeatures(e, "Training",
				newSkillGrant(newTechniqueOf("Karate", "Arm Lock", 0), fxp.Three, grantmerge.Ignore),
				newSkillGrantOf(newTechniqueOf("Arm Lock", "Finisher", 0), grantadjust.SetLevel, fxp.Fifteen,
					grantmerge.Ignore))
			e.Recalculate()
			c.False(e.Unsettled())
			levels := make(map[string]fxp.Int)
			var finisher *Skill
			for _, one := range e.Skills {
				levels[one.Name] = one.LevelData.Level
				if one.Name == "Finisher" {
					finisher = one
				}
			}
			if finisher == nil {
				c.Fatal("expected the grant to put Finisher on the sheet")
			}
			c.Equal(tc.armLock, levels["Arm Lock"])
			c.Equal(tc.points, finisher.AdjustedPoints(nil))
			c.Equal(fxp.Fifteen, finisher.LevelData.Level)
		})
	}
}

func TestGrantedSkillInContainerCostsNothing(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
	e.Recalculate()
	brawling := grantedOf(e.Skills)[0]
	group := NewSkill(e, nil, true)
	group.Children = []*Skill{brawling}
	brawling.SetParent(group)
	e.Skills = []*Skill{group}
	e.Recalculate()
	c.Equal(NumericRangeOf(0), group.PointsRange(nil), "the container shows nothing for what its granted child stands for")
	c.Equal(fxp.FromInteger(12), brawling.LevelData.Level)
}

func TestGrantFromModifierNamesIt(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	training := addTraitWithFeatures(e, "Training")
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Focused"
	mod.Features = Features{newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)}
	training.Modifiers = []*TraitModifier{mod}
	e.Recalculate()
	c.Equal(1, len(grantedOf(e.Traits)))
	c.Equal("Grants Night Vision 1 (from Focused)", e.grantSummary(training))

	mod.Disabled = true
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Traits)), "a disabled modifier grants nothing")
}

func TestGrantSetItem(t *testing.T) {
	c := check.New(t)
	g := NewGrant(feature.GrantTrait)
	g.Adjust = grantadjust.AddPoints
	g.Merge = grantmerge.KeepLowest
	fit := newGrantLeveledTrait("Fit", fxp.One)
	fit.Levels = fxp.Three
	g.SetItem(fit, LibraryFile{})
	c.Equal(fxp.Three, g.Amount, "a leveled trait is granted at its level")
	c.Equal(grantadjust.SetLevel, g.Adjust, "a trait's points follow from its level, so a grant of points sets the level")
	c.Equal(grantmerge.KeepLowest, g.Merge, "every merge mode means something for a leveled trait's level set")
	c.NotEqual(fit.TID, g.Key(), "the grant holds its own copy")
	plain := NewTrait(nil, nil, false)
	plain.Name = "Fearless"
	g.SetItem(plain, LibraryFile{})
	c.Equal(grantmerge.Ignore, g.Merge, "only Ignore means something for a trait without levels")
	g = addingLevels(newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), -fxp.Two, grantmerge.Ignore))
	g.SetItem(newGrantLeveledTrait("Acute Vision", fxp.Two), LibraryFile{})
	c.Equal(-fxp.Two, g.Amount, "levels added are not replaced by the trait's level")

	g = NewGrant(feature.GrantSpell)
	g.Adjust = grantadjust.SetLevel
	g.Amount = fxp.Twelve
	g.SetItem(newHardSpell("Light", fxp.One), LibraryFile{})
	c.Equal(grantadjust.SetLevel, g.Adjust, "a spell can have its level set")
	c.Equal(fxp.Twelve, g.Amount)
	g.Merge = grantmerge.KeepHighest
	g.SetItem(newBless(0), LibraryFile{})
	c.Equal(grantadjust.SetLevel, g.Adjust, "as can a Ritual Magic spell")
	c.Equal(fxp.Twelve, g.Amount)
	c.Equal(grantmerge.KeepHighest, g.Merge)

	g = NewGrant(feature.GrantSpell)
	g.SetItem(nil, LibraryFile{})
	c.NotNil(g.Spell, "nothing makes a blank item of the grant's kind")
	c.NotEqual("", string(g.Key()))

	g = NewGrant(feature.GrantTrait)
	granting := newGrantLeveledTrait("Granting", fxp.One)
	granting.Features = Features{NewGrant(feature.GrantSkill)}
	mod := NewTraitModifier(nil, nil, false)
	mod.Features = Features{NewGrant(feature.GrantSpell), newSkillBonusTo("Brawling", fxp.One)}
	granting.Modifiers = []*TraitModifier{mod}
	g.SetItem(granting, LibraryFile{})
	c.Equal(0, len(g.Trait.Features), "the copy carries no grants, as a sync leaves it")
	c.Equal(1, len(g.Trait.Modifiers[0].Features), "nor do its modifiers")
}

func TestGrantRenewKeyKeepsItsItemsKind(t *testing.T) {
	c := check.New(t)
	for _, g := range []*Grant{
		newSkillGrant(newKicking(0), fxp.One, grantmerge.Ignore),
		newSpellGrant(newBless(0), fxp.One, grantmerge.Ignore),
	} {
		key := g.Key()
		g.RenewKey()
		c.NotEqual(key, g.Key(), g.ItemName())
		c.Equal(tidKind(key), tidKind(g.Key()), "a duplicate of the grant gives the same kind of item, %s", g.ItemName())
	}
}

// nameableKeyOf returns the key of the placeholder marker, such as "@Type@".
func nameableKeyOf(marker string) string {
	for k := range nameable.Extract(nil, nil, marker) {
		return k
	}
	return ""
}

func TestGrantPlaceholdersAnsweredByItsCarrier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	typeKey, ammoKey := nameableKeyOf("@Type@"), nameableKeyOf("@Ammo@")
	guns := newEasySkill("Guns", fxp.One)
	guns.Specialization = "@Type@"
	guns.LocalNotes = "@Ammo@"
	guns.Replacements = map[string]string{ammoKey: "9mm"}
	training := addTraitWithFeatures(e, "Training", newSkillGrant(guns, fxp.One, grantmerge.Ignore))
	m := make(map[string]string)
	training.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{typeKey: nameable.Unset}, m, "the carrier answers what the item leaves unanswered")
	m[typeKey] = "Pistol"
	training.ApplyNameableKeys(m)
	c.Equal(map[string]string{typeKey: "Pistol"}, training.Replacements)
	g, ok := training.Features[0].(*Grant)
	c.True(ok)
	c.Equal("Grants the skill Guns (Pistol) at level 1, or adds 1 level to it when already present",
		g.Describe(e, training.Replacements, func(s string) string { return s }),
		"the description answers the placeholders as the carrier does")

	e.Recalculate()
	granted := grantedOf(e.Skills)
	c.Equal(1, len(granted))
	c.Equal("Pistol", granted[0].SpecializationWithReplacements())
	c.Equal("9mm", granted[0].LocalNotesWithReplacements())
	c.Equal("Grants Guns (Pistol) at level 1", e.grantSummary(training), "as does the carrier's summary")

	id := granted[0].TID
	training.Replacements[typeKey] = "Rifle"
	e.Recalculate()
	granted = grantedOf(e.Skills)
	c.Equal("Rifle", granted[0].SpecializationWithReplacements(), "a new answer remakes the item")
	c.Equal(id, granted[0].TID)

	// A grant carried by a modifier sees its trait's answers through the modifier's nameable links, as a bonus does.
	kindKey := nameableKeyOf("@Kind@")
	e = NewEntity()
	training = addTraitWithFeatures(e, "Training @Kind@")
	training.Replacements = map[string]string{kindKey: "Pistol"}
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Focused"
	linked := newEasySkill("Guns", fxp.One)
	linked.Specialization = "@Type@"
	mod.Features = Features{newSkillGrant(linked, fxp.One, grantmerge.Ignore)}
	training.Modifiers = []*TraitModifier{mod}
	AttachModifiers(training, training.Modifiers)
	training.NameableLinks = map[tid.TID]map[string]string{mod.TID: {typeKey: kindKey}}
	m = make(map[string]string)
	training.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{kindKey: "Pistol"}, m, "the linked placeholder is asked about as the trait's")
	e.Recalculate()
	granted = grantedOf(e.Skills)
	c.Equal(1, len(granted))
	c.Equal("Pistol", granted[0].SpecializationWithReplacements())

	// A level set on a technique is priced from the skill the carrier's answers have it rest on, Guns (Pistol) at 10.
	e = NewEntity()
	addSkill(e, "Guns", fxp.One).Specialization = "Pistol"
	quickShot := newTechniqueOf("Guns", "Quick-Shot", 0)
	quickShot.TechniqueDefault.Specialization.Compare = criteria.IsText
	quickShot.TechniqueDefault.Specialization.Qualifier = "@Type@"
	training = addTraitWithFeatures(e, "Training",
		newSkillGrantOf(quickShot, grantadjust.SetLevel, fxp.Twelve, grantmerge.Ignore))
	training.Replacements = map[string]string{typeKey: "Pistol"}
	e.Recalculate()
	granted = grantedOf(e.Skills)
	if len(granted) != 1 {
		c.Fatalf("expected the technique on the sheet, got %d granted", len(granted))
	}
	c.Equal(fxp.Two, granted[0].AdjustedPoints(nil))
	c.Equal(fxp.Twelve, granted[0].LevelData.Level)
}
