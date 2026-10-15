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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
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

func newTraitGrant(item *Trait, level fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantTrait)
	g.Trait = item
	g.Level = level
	g.Merge = merge
	return g
}

func newSkillGrant(item *Skill, tier fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantSkill)
	g.Skill = item
	g.Level = tier
	g.Merge = merge
	return g
}

func newSpellGrant(item *Spell, tier fxp.Int, merge grantmerge.Mode) *Grant {
	g := NewGrant(feature.GrantSpell)
	g.Spell = item
	g.Level = tier
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

func TestGrantedSkillTier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
	needsBrawling := addTraitWithFeatures(e, "Needs Brawling")
	needsBrawling.Prereq = NewPrereqList()
	addSkillPrereq(needsBrawling.Prereq, "Brawling", fxp.One)
	e.Recalculate()
	granted := grantedOf(e.Skills)
	c.Equal(1, len(granted))
	brawling := granted[0]
	c.Equal(fxp.Int(0), brawling.Points, "a granted skill has no points of its own")
	c.Equal(fxp.Four, brawling.AdjustedPoints(nil), "tier 3 stands for 4 points")
	c.Equal(fxp.FromInteger(12), brawling.LevelData.Level, "an Easy skill at 4 points is DX+2")
	c.Equal(fxp.Int(0), e.PointsBreakdown().Skills, "a granted skill costs nothing")
	c.Equal("", needsBrawling.UnsatisfiedReason, "a granted skill counts as having points")
}

func TestGrantBelowFirstTierPutsNothingOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Curse", newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergeLevels))
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Skills)))
}

func TestGrantSkillMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		merge  grantmerge.Mode
		tier   fxp.Int
		points fxp.Int
		level  fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, tier: fxp.Three, points: fxp.One, level: fxp.Ten},
		{name: "merge levels", merge: grantmerge.MergeLevels, tier: fxp.Three, points: fxp.One, level: fxp.FromInteger(13)},
		{name: "merge points", merge: grantmerge.MergePoints, tier: fxp.Three, points: fxp.Five, level: fxp.FromInteger(12)},
		{name: "keep highest", merge: grantmerge.KeepHighest, tier: fxp.Three, points: fxp.One, level: fxp.FromInteger(12)},
		{name: "keep highest lower", merge: grantmerge.KeepHighest, tier: fxp.One, points: fxp.One, level: fxp.Ten},
		{name: "curse levels", merge: grantmerge.MergeLevels, tier: -fxp.Two, points: fxp.One, level: fxp.Eight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			brawling := addSkill(e, "Brawling", fxp.One)
			addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), tc.tier, tc.merge))
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Skills)), "a grant whose skill is already present adds no row")
			c.Equal(tc.points, brawling.AdjustedPoints(nil))
			c.Equal(tc.level, brawling.LevelData.Level)
			c.Equal(fxp.One, e.PointsBreakdown().Skills, "merged points cost nothing")
		})
	}
}

func TestGrantStateCarriesOverToNextRecalculation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels),
		newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two, grantmerge.Ignore))
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), brawling.LevelData.Level)
	c.Equal(1, e.recalculate(maxRecalculationPassesInAll), "an unchanged sheet settles in one pass")
	c.Equal(fxp.FromInteger(13), brawling.LevelData.Level)

	training.Features = nil
	e.Recalculate()
	c.Equal(fxp.Ten, brawling.LevelData.Level, "a grant that is gone takes back what it merged")
}

func TestGrantCurseRemovesTopTiers(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.Five)
	addTraitWithFeatures(e, "Curse", newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints))
	e.Recalculate()
	c.Equal(fxp.Three, brawling.AdjustedPoints(nil), "tier 3 drops to tier 2, keeping the point beyond the tier")

	brawling.Points = fxp.One
	e.Recalculate()
	c.Equal(fxp.Int(0), brawling.AdjustedPoints(nil), "never below zero")
	c.Equal(fxp.Min, brawling.LevelData.Level, "with no points left, the skill has no level")
}

func TestGrantTraitMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		merge    grantmerge.Mode
		perLevel fxp.Int
		have     fxp.Int
		grant    fxp.Int
		level    fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.One},
		{name: "merge levels", merge: grantmerge.MergeLevels, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.Three},
		{name: "curse", merge: grantmerge.MergeLevels, perLevel: fxp.One, have: fxp.Three, grant: -fxp.One, level: fxp.Two},
		{name: "keep highest", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.One, grant: fxp.Three, level: fxp.Three},
		{name: "keep highest lower", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.Three},
		{name: "keep lowest disadvantage", merge: grantmerge.KeepHighest, perLevel: -fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.One},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			bad := addTrait(e, newGrantLeveledTrait("Bad Sight", tc.perLevel))
			bad.Levels = tc.have
			addTraitWithFeatures(e, "Racial", newTraitGrant(newGrantLeveledTrait("Bad Sight", tc.perLevel), tc.grant, tc.merge))
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Traits)), "a grant whose trait is already present adds no row")
			c.Equal(tc.level, bad.CurrentLevel())
			c.Equal(tc.have.Mul(tc.perLevel), bad.AdjustedPoints(nil), "merged levels cost nothing")
		})
	}
}

func TestGrantDisabledTraitCountsAsAbsent(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTrait(e, newGrantLeveledTrait("Night Vision", fxp.One)).SetDisabled(true)
	addTraitWithFeatures(e, "Racial", newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two,
		grantmerge.MergeLevels))
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
	grant.Level = fxp.Three
	e.Recalculate()
	after := grantedOf(e.Traits)[0]
	c.False(before == after)
	c.Equal(before.TID, after.TID, "the refreshed item keeps its ID")
	c.True(after.SwitchedOn, "and its switch")
	c.Equal(fxp.Three, after.CurrentLevel())
}

func TestGrantSurvivesSaveAndLoad(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Racial", newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two,
		grantmerge.Ignore))
	e.Recalculate()
	before := grantedOf(e.Traits)[0]
	data, err := json.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(json.Unmarshal(data, &loaded))
	loaded.Recalculate()
	granted := grantedOf(loaded.Traits)
	c.Equal(1, len(granted))
	c.Equal(before.TID, granted[0].TID, "the loaded item is kept, not made again")
	c.Equal(fxp.Two, granted[0].CurrentLevel())
	c.Equal(len(e.Traits), len(loaded.Traits))
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
}

func TestGrantGivesWayToTheSameItemAddedLater(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.MergePoints))
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
	e.Recalculate()
	summary := e.grantSummary(training)
	c.Contains(summary, "Grants Night Vision 2")
	c.Contains(summary, "Does not grant Acute Vision 1, since it is already present")
	c.Contains(e.grantSummary(goggles), "Would grant Dark Vision 1, but the grant is not in effect")
	c.Equal("", e.grantSummary(have), "an item carrying no grants has no summary")
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
			c.Equal(0, len(grantedOf(e.Traits)), "a grant whose carrier's prerequisites are unmet puts nothing on the sheet")
			c.Contains(e.grantSummary(carrier), "Does not grant Night Vision 1")

			addSkill(e, "Brawling", fxp.One)
			e.Recalculate()
			c.Equal(1, len(grantedOf(e.Traits)), "the grant puts its trait on the sheet once they are met")
			c.False(grantedOf(e.Traits)[0].grantNotInEffect())

			e.Skills = nil
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Traits)), "the trait goes again once they are no longer met")
		})
	}
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
	c.Equal(0, len(grantedOf(e.Traits)))

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
	c := check.New(t)
	e := NewEntity()
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.MergeLevels))
	training.Prereq = NewPrereqList()
	addSkillPrereq(training.Prereq, "Brawling", fxp.One)
	addSkill(e, "Brawling", fxp.One)
	e.Recalculate()
	c.Equal("", training.UnsatisfiedReason, "the skill bought meets the prerequisite")
	c.Equal(0, len(grantedOf(e.Skills)), "the grant merges into the skill bought")

	e.Skills = nil
	e.Recalculate()
	c.NotEqual("", training.UnsatisfiedReason, "the skill it grants can't meet its own prerequisite")
	c.Equal(0, len(grantedOf(e.Skills)), "so it grants nothing")
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
	merge := newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.Two, grantmerge.MergeLevels)
	c.Equal("Adds 2 levels to Acute Vision, already present", merge.mergeSummary(merge.itemLabel()))
	merge.Level = -fxp.One
	c.Equal("Removes 1 level from Acute Vision, already present", merge.mergeSummary(merge.itemLabel()))
	merge.Level = fxp.FromStringForced("1.5")
	c.Equal("Adds 1.5 levels to Acute Vision, already present", merge.mergeSummary(merge.itemLabel()))
	c.Equal("Grants the trait Acute Vision at level 1.5, or adds +1.5 levels to it when already present",
		merge.Describe(nil, nil, plain))
	merge.Level = fxp.Two
	merge.Merge = grantmerge.Ignore
	c.Equal("Does not grant Acute Vision 2, since it is already present", merge.mergeSummary(merge.itemLabel()))
}

func TestGrantTiers(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		tier     fxp.Int
		wildcard bool
		points   fxp.Int
	}{
		{tier: 0, points: 0},
		{tier: fxp.One, points: fxp.One},
		{tier: fxp.Two, points: fxp.Two},
		{tier: fxp.Three, points: fxp.Four},
		{tier: fxp.Four, points: fxp.Eight},
		{tier: fxp.Three, wildcard: true, points: fxp.FromInteger(12)},
	} {
		c.Equal(tc.points, grantTierPoints(tc.tier, tc.wildcard))
		c.Equal(tc.tier, grantTierOf(tc.points, tc.wildcard))
	}
}

func TestGrantSpellMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		merge grantmerge.Mode
		tier  fxp.Int
		level fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, tier: fxp.Three, level: fxp.Eight},
		{name: "merge levels", merge: grantmerge.MergeLevels, tier: fxp.Two, level: fxp.Ten},
		{name: "keep highest", merge: grantmerge.KeepHighest, tier: fxp.Three, level: fxp.Ten},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			addTraitWithFeatures(e, "Magery", newSpellGrant(newHardSpell("Light", fxp.One), tc.tier, tc.merge))
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Spells)))
			c.Equal(tc.level, light.LevelData.Level)
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
}

func TestGrantKeepHighestComparesWithOtherMerges(t *testing.T) {
	for _, keepFirst := range []bool{false, true} {
		c := check.New(t)
		e := NewEntity()
		bad := addTrait(e, newGrantLeveledTrait("Bad Sight", fxp.One))
		bad.Levels = fxp.Three
		grants := []*Trait{
			addTraitWithFeatures(e, "Adds", newTraitGrant(newGrantLeveledTrait("Bad Sight", fxp.One), fxp.One,
				grantmerge.MergeLevels)),
			addTraitWithFeatures(e, "Keeps", newTraitGrant(newGrantLeveledTrait("Bad Sight", fxp.One), fxp.Five,
				grantmerge.KeepHighest)),
		}
		if keepFirst {
			e.Traits = []*Trait{bad, grants[1], grants[0]}
		}
		e.Recalculate()
		c.Equal(fxp.Five, bad.CurrentLevel(), "whichever comes first")
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
	g.Merge = grantmerge.MergePoints
	fit := newGrantLeveledTrait("Fit", fxp.One)
	fit.Levels = fxp.Three
	g.SetItem(fit, LibraryFile{})
	c.Equal(fxp.Three, g.Level, "a leveled trait is granted at its level")
	c.Equal(grantmerge.Ignore, g.Merge, "a merge that means nothing for a trait becomes Ignore")
	c.NotEqual(fit.TID, g.Key(), "the grant holds its own copy")

	g = NewGrant(feature.GrantSpell)
	g.SetItem(nil, LibraryFile{})
	c.NotNil(g.Spell, "nothing makes a blank item of the grant's kind")
	c.NotEqual("", string(g.Key()))
}

func TestGrantPlaceholdersAnsweredByItsCarrier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	keyOf := func(marker string) string {
		for k := range nameable.Extract(nil, nil, marker) {
			return k
		}
		return ""
	}
	typeKey, ammoKey := keyOf("@Type@"), keyOf("@Ammo@")
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

	e.Recalculate()
	granted := grantedOf(e.Skills)
	c.Equal(1, len(granted))
	c.Equal("Pistol", granted[0].SpecializationWithReplacements())
	c.Equal("9mm", granted[0].LocalNotesWithReplacements())

	id := granted[0].TID
	training.Replacements[typeKey] = "Rifle"
	e.Recalculate()
	granted = grantedOf(e.Skills)
	c.Equal("Rifle", granted[0].SpecializationWithReplacements(), "a new answer remakes the item")
	c.Equal(id, granted[0].TID)
}
