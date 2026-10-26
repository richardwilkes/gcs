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
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
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
	c.Equal("0", ResolveScript(e, deferredNewScriptSkill(brawling), "self.points"), "nor does it to a script")
	c.Equal("", needsBrawling.UnsatisfiedReason, "a granted skill counts as having points")
}

func TestGrantBelowFirstTierPutsNothingOnSheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Curse", newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergeLevels))
	e.Recalculate()
	c.Equal(0, len(grantedOf(e.Skills)))
}

func TestGrantOfContainerPutsNothingOnSheet(t *testing.T) {
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
}

func TestGrantSkillMergeModes(t *testing.T) {
	typeKey := nameableKeyOf("@Type@")
	for _, tc := range []struct {
		name  string
		merge grantmerge.Mode
		tier  fxp.Int
		// setup, when set, adjusts the skill bought, the grant and its carrier before the recalculation.
		setup  func(e *Entity, bought *Skill, g *Grant, carrier *Trait)
		points fxp.Int
		level  fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, tier: fxp.Three, points: fxp.One, level: fxp.Ten},
		{name: "merge levels", merge: grantmerge.MergeLevels, tier: fxp.Three, points: fxp.One, level: fxp.FromInteger(13)},
		{name: "merge points", merge: grantmerge.MergePoints, tier: fxp.Three, points: fxp.Five, level: fxp.FromInteger(12)},
		{name: "keep highest", merge: grantmerge.KeepHighest, tier: fxp.Three, points: fxp.One, level: fxp.FromInteger(12)},
		{name: "keep highest lower", merge: grantmerge.KeepHighest, tier: fxp.One, points: fxp.One, level: fxp.Ten},
		{
			name: "keep highest with placeholder and bonus", merge: grantmerge.KeepHighest, tier: fxp.Four,
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
		{name: "curse levels", merge: grantmerge.MergeLevels, tier: -fxp.Two, points: fxp.One, level: fxp.Eight},
		{
			name: "keep highest with a missing-equipment penalty", merge: grantmerge.KeepHighest, tier: fxp.Three,
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			brawling := addSkill(e, "Brawling", fxp.One)
			g := newSkillGrant(newEasySkill("Brawling", fxp.One), tc.tier, tc.merge)
			training := addTraitWithFeatures(e, "Training", g)
			if tc.setup != nil {
				tc.setup(e, brawling, g, training)
			}
			e.Recalculate()
			c.Equal(0, len(grantedOf(e.Skills)), "a grant whose skill is already present adds no row")
			c.Equal(tc.points, brawling.AdjustedPoints(nil))
			c.Equal(tc.level, brawling.LevelData.Level)
			c.Equal(fxp.One, e.PointsBreakdown().Skills, "merged points cost nothing")
		})
	}
}

func TestGrantMergesIntoGrantedSkillAsIntoBought(t *testing.T) {
	for _, tc := range []struct {
		name  string
		merge grantmerge.Mode
		tier  fxp.Int
		level fxp.Int
	}{
		{name: "merge levels", merge: grantmerge.MergeLevels, tier: fxp.Two, level: fxp.FromInteger(14)},
		{name: "keep highest", merge: grantmerge.KeepHighest, tier: fxp.Four, level: fxp.FromInteger(13)},
		{name: "curse levels", merge: grantmerge.MergeLevels, tier: -fxp.One, level: fxp.FromInteger(11)},
		{name: "curse points", merge: grantmerge.MergePoints, tier: -fxp.One, level: fxp.FromInteger(11)},
	} {
		for _, bought := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s bought=%t", tc.name, bought), func(t *testing.T) {
				c := check.New(t)
				e := NewEntity()
				var brawling *Skill
				if bought {
					brawling = addSkill(e, "Brawling", fxp.Four)
				} else {
					addTraitWithFeatures(e, "Racial",
						newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
				}
				addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), tc.tier, tc.merge))
				e.Recalculate()
				if !bought {
					brawling = grantedOf(e.Skills)[0]
				}
				c.Equal(tc.level, brawling.LevelData.Level, "a merge into a granted skill reads the level its tier gives it")
			})
		}
	}
}

func TestGrantStateAppliesPointsExcept(t *testing.T) {
	c := check.New(t)
	a, b := NewTrait(nil, nil, false), NewTrait(nil, nil, false)
	a.Name, b.Name = "A", "B"
	var state grantState
	state.addPoints(grantSource{owner: a}, fxp.Four)
	state.addPoints(grantSource{owner: b}, fxp.One)
	var tooltip xbytes.InsertBuffer
	c.Equal(fxp.Two, state.applyPoints(a, fxp.One, &tooltip))
	c.Equal("\nB [+1]", tooltip.String(), "the tooltip lists only the points that were added")
	tooltip.Reset()
	c.Equal(fxp.Six, state.applyPoints(nil, fxp.One, &tooltip))
	c.Equal("\nA [+4]\nB [+1]", tooltip.String())
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
	for _, tc := range []struct {
		bought fxp.Int
		points fxp.Int
		level  fxp.Int
	}{
		{bought: fxp.Four, points: fxp.Two, level: fxp.FromInteger(11)},
		// The points beyond the tier go with it, or the level would not drop.
		{bought: fxp.Five, points: fxp.Two, level: fxp.FromInteger(11)},
		{bought: fxp.Three, points: fxp.One, level: fxp.Ten},
		{bought: fxp.Six, points: fxp.Two, level: fxp.FromInteger(11)},
		{bought: fxp.Seven, points: fxp.Two, level: fxp.FromInteger(11)},
		// Never below zero, and with no points left, the skill has no level.
		{bought: fxp.One, points: 0, level: fxp.Min},
	} {
		t.Run(tc.bought.String()+" points", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			brawling := addSkill(e, "Brawling", tc.bought)
			addTraitWithFeatures(e, "Curse",
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints))
			e.Recalculate()
			c.Equal(tc.points, brawling.AdjustedPoints(nil))
			c.Equal(tc.level, brawling.LevelData.Level)
		})
	}
}

func TestGrantTraitMergeModes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		merge    grantmerge.Mode
		perLevel fxp.Int
		have     fxp.Int
		// bonus is a trait bonus to the trait's level from elsewhere on the sheet.
		bonus fxp.Int
		grant fxp.Int
		level fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.One},
		{name: "merge levels", merge: grantmerge.MergeLevels, perLevel: fxp.One, have: fxp.One, grant: fxp.Two, level: fxp.Three},
		{name: "curse", merge: grantmerge.MergeLevels, perLevel: fxp.One, have: fxp.Three, grant: -fxp.One, level: fxp.Two},
		{name: "keep highest", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.One, grant: fxp.Three, level: fxp.Three},
		{name: "keep highest lower", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.Three},
		{name: "keep highest with bonus", merge: grantmerge.KeepHighest, perLevel: fxp.One, have: fxp.One, bonus: fxp.Two, grant: fxp.Four, level: fxp.Six},
		{name: "keep lowest disadvantage", merge: grantmerge.KeepHighest, perLevel: -fxp.One, have: fxp.Three, grant: fxp.One, level: fxp.One},
		// Below level 1 a grant gives nothing, as it says of itself; only a negative level takes away.
		{name: "merge half a level", merge: grantmerge.MergeLevels, perLevel: fxp.One, have: fxp.Two, grant: fxp.Half, level: fxp.Two},
		{name: "keep highest at level 0 of a disadvantage", merge: grantmerge.KeepHighest, perLevel: -fxp.One, have: fxp.Three, grant: 0, level: fxp.Three},
		// A disadvantage at level 0 costs nothing, but is still a disadvantage, so the lower level is kept.
		{name: "keep lowest for a disadvantage at level 0", merge: grantmerge.KeepHighest, perLevel: -fxp.One, have: 0, grant: fxp.Two, level: 0},
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
			addTraitWithFeatures(e, "Racial", newTraitGrant(newGrantLeveledTrait("Bad Sight", tc.perLevel), tc.grant, tc.merge))
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
	addTraitWithFeatures(e, "Racial", newTraitGrant(toughSkin, fxp.Two, grantmerge.MergeLevels))
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
			g := newTraitGrant(newMagery(), fxp.One, grantmerge.MergeLevels)
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
	first.Level = fxp.Two
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
	addTraitWithFeatures(e, "Racial", newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.Two,
		grantmerge.Ignore))
	e.Recalculate()
	before := grantedOf(e.Traits)[0]
	data, err := json.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(json.Unmarshal(data, &loaded))
	granted := grantedOf(loaded.Traits)
	c.Equal(1, len(granted))
	c.Equal(before.TID, granted[0].TID)
	loaded.Recalculate()
	c.True(granted[0] == grantedOf(loaded.Traits)[0], "the loaded item is kept, not made again, which keeps its ID anyway")
	c.Equal(fxp.Two, granted[0].CurrentLevel())
	c.Equal(len(e.Traits), len(loaded.Traits))
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
	c.Equal(fxp.Four, copied.Points, "so it is bought with the points its tier stood for")
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
	newGrant := func() *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore)
	}
	for _, tc := range []struct {
		name string
		// carry puts two carriers of a Night Vision grant on the sheet, the first of which doesn't give the trait.
		carry func(e *Entity) (first, second any)
		// first is what the first carrier's summary says of its grant.
		first string
	}{
		{name: "unequipped before equipped", carry: func(e *Entity) (any, any) {
			spare := addCarriedEquipmentWithFeatures(e, "Spare Goggles", newGrant())
			spare.Equipped = false
			return spare, addCarriedEquipmentWithFeatures(e, "Worn Goggles", newGrant())
		}, first: "Would grant Night Vision 1, but the grant is not in effect"},
		{name: "unmet carrier before met", carry: func(e *Entity) (any, any) {
			a := addTraitWithFeatures(e, "A", newGrant())
			a.Prereq = NewPrereqList()
			addSkillPrereq(a.Prereq, "Brawling", fxp.One)
			return a, addTraitWithFeatures(e, "B", newGrant())
		}, first: "Does not grant Night Vision 1, since the prerequisites of what carries the grant are not met"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			first, second := tc.carry(e)
			e.Recalculate()
			c.False(e.Unsettled())
			granted := grantedOf(e.Traits)
			if len(granted) != 1 {
				c.Fatalf("expected 1 granted trait, got %d", len(granted))
			}
			c.True(granted[0].Enabled(), "the grant in effect stands in for the one that isn't")
			c.Equal(fxp.One, granted[0].CurrentLevel())
			c.True(GrantingItem(granted[0]) == second)
			c.Contains(e.grantSummary(second), "Grants Night Vision 1")
			c.Equal(tc.first, e.grantSummary(first))
		})
	}
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
}

func TestGrantCarrierJudgedAfresh(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	brawling := addSkill(e, "Brawling", fxp.One)
	training := addTraitWithFeatures(e, "Training",
		newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
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
		c.True(grantedOf(e.Skills)[0].grantNotInEffect(), "so the skill it grants is not in effect")
	})

	// Nor is the item seen by a script run for the prerequisites, whatever the script reads it through.
	brawlingGrant := func() *Grant {
		return newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.MergeLevels)
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
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
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
		karate.Features = Features{newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Two, grantmerge.MergeLevels)}
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
				newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.MergeLevels))
			gifted.Prereq = newPrereqListForTrait("Magery", true)
			traitPrereqOf(c, gifted.Prereq).LevelCriteria.Qualifier = fxp.Two
			return gifted, magery.CurrentLevel
		}, level: fxp.One},
		{name: "levels into a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a skill read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.skillLevel("Brawling", "", false) >= 12`)
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a skill found by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
			training.Prereq = NewPrereqList()
			addScriptPrereq(training.Prereq, `entity.findSkills("Brawling").some(s => s.level >= 12)`)
			return training, func() fxp.Int { return brawling.LevelData.Level }
		}, level: fxp.Ten},
		{name: "levels into a trait read by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			magery := addTrait(e, newGrantLeveledTrait("Magery", fxp.Ten))
			gifted := addTraitWithFeatures(e, "Gifted",
				newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.MergeLevels))
			gifted.Prereq = NewPrereqList()
			addScriptPrereq(gifted.Prereq, `entity.traitLevel("Magery") >= 2`)
			return gifted, magery.CurrentLevel
		}, level: fxp.One},
		{name: "levels into a spell found by a script", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			magery := addTraitWithFeatures(e, "Magery",
				newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.MergeLevels))
			magery.Prereq = NewPrereqList()
			addScriptPrereq(magery.Prereq, `entity.findSpells("Light").some(s => s.level >= 10)`)
			return magery, func() fxp.Int { return light.LevelData.Level }
		}, level: fxp.Eight},
		{name: "points into a skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergePoints))
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
			gifted := addTraitWithFeatures(e, "Gifted", newTraitGrant(newMagery(), fxp.Two, grantmerge.MergeLevels))
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
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels))
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
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.MergeLevels))
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
				newSkillGrant(newEasySkill("Broadsword", fxp.One), fxp.Six, grantmerge.MergeLevels))
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
				newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.MergeLevels))
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

func TestGrantHashCoversWholeItem(t *testing.T) {
	c := check.New(t)
	g := newTraitGrant(newGrantLeveledTrait("Bad Temper", fxp.One), fxp.One, grantmerge.Ignore)
	g.Trait.SelfControl = selfctrl.CR12
	before := Hash64(g)
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
	merge := newTraitGrant(newGrantLeveledTrait("Acute Vision", fxp.One), fxp.Two, grantmerge.MergeLevels)
	c.Equal("Adds 2 levels to Acute Vision, already present", merge.mergeSummary(merge.itemLabel(nil), nil))
	merge.Level = -fxp.One
	c.Equal("Removes 1 level from Acute Vision, already present", merge.mergeSummary(merge.itemLabel(nil), nil))
	merge.Level = fxp.FromStringForced("1.5")
	c.Equal("Adds 1.5 levels to Acute Vision, already present", merge.mergeSummary(merge.itemLabel(nil), nil))
	c.Equal("Grants the trait Acute Vision at level 1.5, or adds +1.5 levels to it when already present",
		merge.Describe(nil, nil, plain))
	merge.Level = fxp.Two
	merge.Merge = grantmerge.Ignore
	c.Equal("Does not grant Acute Vision 2, since it is already present", merge.mergeSummary(merge.itemLabel(nil), nil))

	// A grant below level 1 grants nothing, and one with a negative level only takes away.
	curse := newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints)
	c.Equal("Removes the points of 1 level from the skill Brawling when already present", curse.Describe(nil, nil, plain))
	c.Equal("Removes the points of 1 level from Brawling, already present", curse.mergeSummary(curse.itemLabel(nil), nil))
	curse.Level = -fxp.Two
	c.Equal("Removes the points of 2 levels from the skill Brawling when already present", curse.Describe(nil, nil, plain))
	c.Equal("Removes the points of 2 levels from Brawling, already present", curse.mergeSummary(curse.itemLabel(nil), nil))
	curse.Merge = grantmerge.MergeLevels
	c.Equal("Removes 2 levels from the skill Brawling when already present", curse.Describe(nil, nil, plain))
	curse.Merge = grantmerge.Ignore
	c.Equal("Would grant the skill Brawling at level -2, but a level below 1 grants nothing", curse.Describe(nil, nil, plain))
	curse.Level = 0
	curse.Merge = grantmerge.MergeLevels
	c.Equal("Would grant the skill Brawling at level 0, but a level below 1 grants nothing", curse.Describe(nil, nil, plain))
	// Nor does such a grant merge anything into an item already present, which its summary says too.
	c.Equal("Does not change Brawling, already present, since a level below 1 grants nothing",
		curse.mergeSummary(curse.itemLabel(nil), nil))
	curse.Merge = grantmerge.MergePoints
	c.Equal("Does not change Brawling, already present, since a level below 1 grants nothing",
		curse.mergeSummary(curse.itemLabel(nil), nil))
	curse.Level = -fxp.One
	curse.Merge = grantmerge.KeepHighest
	c.Equal("Does not change Brawling, already present, since a level below 1 grants nothing",
		curse.mergeSummary(curse.itemLabel(nil), nil))
	merge.Level = fxp.Half
	merge.Merge = grantmerge.MergeLevels
	c.Equal("Does not change Acute Vision, already present, since a level below 1 grants nothing",
		merge.mergeSummary(merge.itemLabel(nil), nil))
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
		// points is what the spell bought has; with none, it has no level.
		points fxp.Int
		tier   fxp.Int
		level  fxp.Int
	}{
		{name: "ignore", merge: grantmerge.Ignore, points: fxp.One, tier: fxp.Three, level: fxp.Eight},
		{name: "merge levels", merge: grantmerge.MergeLevels, points: fxp.One, tier: fxp.Two, level: fxp.Ten},
		{name: "merge levels without level", merge: grantmerge.MergeLevels, tier: fxp.Two, level: fxp.Nine},
		{name: "keep highest", merge: grantmerge.KeepHighest, points: fxp.One, tier: fxp.Three, level: fxp.Ten},
		{name: "keep highest without level", merge: grantmerge.KeepHighest, tier: fxp.Three, level: fxp.Ten},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			light := newHardSpell("Light", tc.points)
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
	c.Equal("0", ResolveScript(e, deferredNewScriptSpell(list[0]), "self.points"), "a granted spell costs nothing to a script")
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
				newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.One, grantmerge.MergeLevels))
			training.Prereq = newPrereqListForTrait("Magery", true)
			traitPrereqOf(c, training.Prereq).LevelCriteria.Qualifier = fxp.Three
			addTraitWithFeatures(e, "Gifted",
				newTraitGrant(newGrantLeveledTrait("Magery", fxp.Ten), fxp.Three, grantmerge.KeepHighest))
			return training, magery.CurrentLevel
		}, level: fxp.Three},
		{name: "skill", carry: func(_ check.Checker, e *Entity) (*Trait, func() fxp.Int) {
			brawling := addSkill(e, "Brawling", fxp.One)
			training := addTraitWithFeatures(e, "Training",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.MergeLevels))
			training.Prereq = newPrereqListRequiringSkill("Brawling", fxp.FromInteger(12))
			addTraitWithFeatures(e, "Gifted",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.KeepHighest))
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
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergePoints),
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints),
			}
		}, points: fxp.Four, level: fxp.FromInteger(12)},
		{name: "levels and points into a skill with no level", bought: 0, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.MergeLevels),
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.One, grantmerge.MergePoints),
			}
		}, points: fxp.Five, level: fxp.FromInteger(12)},
		// Two curses take their tiers off together, rather than each taking its own off the points bought.
		{name: "two curses", bought: fxp.Eight, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints),
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints),
			}
		}, points: fxp.Two, level: fxp.FromInteger(11)},
		{name: "curses of different sizes", bought: fxp.Eight, grants: func() [2]*Grant {
			return [2]*Grant{
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.One, grantmerge.MergePoints),
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.Two, grantmerge.MergePoints),
			}
		}, points: fxp.One, level: fxp.Ten},
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
			})
		}
	}
}

func TestGrantPlanterDoesNotDependOnCarrierOrder(t *testing.T) {
	nightVision := func(level fxp.Int, merge grantmerge.Mode) *Grant {
		return newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), level, merge)
	}
	for _, tc := range []struct {
		name string
		// grants are the grants of the item, none of which finds it bought. unmet, when not empty, is the name of the
		// carrier whose prerequisites are unmet, so that its grant is not in effect. selfControl, when set, is the
		// self-control roll of the row.
		grants      []*Grant
		unmet       string
		level       fxp.Int
		selfControl selfctrl.Roll
	}{
		{name: "ignore and merge levels", grants: []*Grant{
			nightVision(fxp.Two, grantmerge.Ignore), nightVision(fxp.One, grantmerge.MergeLevels),
		}, level: fxp.Three},
		{name: "two ignores", grants: []*Grant{
			nightVision(fxp.One, grantmerge.Ignore), nightVision(fxp.Two, grantmerge.Ignore),
		}, level: fxp.Two},
		{name: "keep highest and merge levels", grants: []*Grant{
			nightVision(fxp.One, grantmerge.KeepHighest), nightVision(fxp.Two, grantmerge.MergeLevels),
		}, level: fxp.Three},
		{name: "stand-in for an ignore not in effect", grants: []*Grant{
			nightVision(fxp.Three, grantmerge.Ignore), nightVision(fxp.Two, grantmerge.KeepHighest),
			nightVision(fxp.One, grantmerge.MergeLevels),
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
				granted := grantedOf(e.Traits)
				if len(granted) != 1 {
					c.Fatalf("expected one granted row, got %d", len(granted))
				}
				c.True(granted[0].Enabled())
				c.Equal(tc.level, granted[0].CurrentLevel())
				c.Equal("", granted[0].UnsatisfiedReason, "a grant that treats the item's prerequisites as met wins")
				if tc.selfControl != selfctrl.None {
					c.Equal(tc.selfControl, granted[0].SelfControl, "the grant of the item worth more puts it on the sheet")
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
	a := addTraitWithFeatures(e, "A", nightVision())
	b := addTraitWithFeatures(e, "B", nightVision())
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
}

func TestGrantedSkillSatisfiesTechnique(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTraitWithFeatures(e, "Training", newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Three, grantmerge.Ignore))
	disarming := NewTechnique(e, nil, "Brawling")
	disarming.Name = "Disarming"
	e.Skills = append(e.Skills, disarming)
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), disarming.LevelData.Level, "the technique takes its level from the granted skill")
	c.Equal("", disarming.UnsatisfiedReason, "which its tier's points satisfy, as they give it its level")
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
	c.Equal("Grants the skill Guns (Pistol) at level 1, unless already present",
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
}
