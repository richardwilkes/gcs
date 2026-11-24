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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMixedDefaultChainsTerminate checks that defaults running between skills and spells settle, with no two items
// recording defaults to each other, and that computing their levels afresh, as the prerequisites of an item carrying a
// grant see the sheet and as the points a level set gives are priced at base, comes to an end.
func TestMixedDefaultChainsTerminate(t *testing.T) {
	restsOn := func(h defaultable, name string) bool {
		def := h.recordedDefault()
		return def != nil && def.NameWithReplacements(nil) == name
	}
	// grantCarrierNeeding puts on the sheet a trait carrying a grant, whose prerequisites are judged afresh, needing
	// the named skill at 10.
	grantCarrierNeeding := func(e *Entity, name string) *Trait {
		carrier := addTraitWithFeatures(e, "Initiate",
			newTraitGrant(newGrantLeveledTrait("Night Vision", fxp.One), fxp.One, grantmerge.Ignore))
		carrier.Prereq = NewPrereqList()
		addSkillPrereq(carrier.Prereq, name, fxp.Ten)
		return carrier
	}
	for _, tc := range []struct {
		name string
		// setup builds the sheet and returns a check of what it settles to.
		setup func(c check.Checker, e *Entity) func()
	}{
		{
			name: "a skill and a spell with points defaulting to each other",
			setup: func(c check.Checker, e *Entity) func() {
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.Four)
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Lore", -fxp.Two)}
				lore := addTestSpell(e, "Lore", fxp.Four)
				lore.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				return func() {
					c.True(restsOn(thaumatology, "Lore"), "the skill, updated first, records the spell")
					c.False(restsOn(lore, "Thaumatology"), "so the spell can't record the skill")
				}
			},
		},
		{
			name: "two spells with points defaulting to each other",
			setup: func(c check.Checker, e *Entity) func() {
				lore := addTestSpell(e, "Lore", fxp.Four)
				lore.Defaults = []*SkillDefault{newSpellDefaultTo("Wisdom", -fxp.Two)}
				wisdom := addTestSpell(e, "Wisdom", fxp.Four)
				wisdom.Defaults = []*SkillDefault{newSpellDefaultTo("Lore", -fxp.Two)}
				return func() {
					c.True(restsOn(lore, "Wisdom"), "the spell listed first records the other")
					c.False(restsOn(wisdom, "Lore"), "so the other can't record it")
				}
			},
		},
		{
			// Neither spell records the other, so that each is weighed as the other's default when judged afresh.
			name: "two Ritual Magic spells defaulting to each other, judged afresh for a grant carrier",
			setup: func(c check.Checker, e *Entity) func() {
				var rites []*Spell
				for _, one := range []struct{ name, other string }{{"Bless", "Curse"}, {"Curse", "Bless"}} {
					rite := addTestRitualMagicSpell(e, one.name, "Fire")
					rite.Defaults = []*SkillDefault{
						{DefaultType: IntelligenceID},
						newSpellDefaultTo(one.other, -fxp.One),
					}
					rites = append(rites, rite)
				}
				// Bless at 10 stands for the 2 points that buy IQ+0, which add to Augury's 1.
				augury := addTestSkill(e, "Augury", "", "", fxp.One)
				augury.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", 0)}
				carrier := grantCarrierNeeding(e, "Augury")
				return func() {
					for _, rite := range rites {
						c.Equal(IntelligenceID, rite.DefaultedFrom.Type(), "%s rests on IQ", rite.Name)
					}
					c.Equal(fxp.Ten, augury.LevelData.Level, "Augury rests on Bless")
					c.Equal("", carrier.UnsatisfiedReason, "which meets the carrier's prerequisite")
				}
			},
		},
		{
			name: "a ritual skill defaulting to a Ritual Magic spell resting on it",
			setup: func(c check.Checker, e *Entity) func() {
				ritual := addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Four) // 11
				ritual.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.One)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				bless.PrereqCount = 2 // 9, from the ritual skill
				template := NewRitualMagicSpell(nil, nil, false)
				template.Name = "Bless"
				template.College = CollegeList{"Fire"}
				template.PrereqCount = 2
				addTraitWithFeatures(e, "Blessed",
					newSpellGrantOf(template, grantadjust.SetLevel, fxp.Ten, grantmerge.KeepHighest))
				carrier := grantCarrierNeeding(e, "Ritual Magic")
				return func() {
					c.Equal(fxp.Ten, bless.LevelData.Level, "the level set raises the spell, priced at base")
					c.Equal(fxp.Eleven, ritual.LevelData.Level, "the ritual skill's own points give it more")
					c.Equal("", carrier.UnsatisfiedReason, "which meets the carrier's prerequisite")
				}
			},
		},
		{
			// Bless records Thaumatology while no ritual skill gives it a level, and still does, as the best of its
			// defaults, once Ritual Magic (Fire) gives it 16, though its level no longer rests on it.
			name: "a skill defaulting to a Ritual Magic spell whose ritual skill comes to outrank its default to the skill",
			setup: func(c check.Checker, e *Entity) func() {
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.One) // 9
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.Two)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				bless.Points = fxp.Two
				bless.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				e.Recalculate()
				c.True(restsOn(bless, "Thaumatology"), "precondition: the spell records the skill")
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.FromInteger(24)) // 16
				return func() {
					c.True(restsOn(thaumatology, "Bless"), "the skill records the spell")
					c.Equal(fxp.FromInteger(14), thaumatology.LevelData.Level, "at the spell's 16 less 2")
				}
			},
		},
		{
			name: "a skill defaulting to a spell of its name",
			setup: func(c check.Checker, e *Entity) func() {
				skill := addTestSkill(e, "Shield", "", "", 0)
				skill.Defaults = []*SkillDefault{newSpellDefaultTo("Shield", -fxp.Two)}
				addTestSpell(e, "Shield", fxp.Four) // 10
				return func() {
					c.Equal(fxp.Eight, skill.LevelData.Level, "the spell is not taken for the skill excluded")
				}
			},
		},
	} {
		c := check.New(t)
		e := NewEntity()
		settled := tc.setup(c, e)
		e.Recalculate()
		c.False(e.Unsettled(), "%s: the sheet settles", tc.name)
		settled()
		checkRecalculationConsistency(c, e, tc.name)
	}
}

// TestSwapDefaultsOf checks which skills and spells Swap Defaults can act on and what it does to them: a skill and a
// spell defaulting to each other trade places from either one, as two spells do, a skill trades places with a Ritual
// Magic spell either way, but not where the spell would then have no level or its ritual skill would outrank its
// default to the skill, a spell with several defaults moves to the next, and a granted spell, a technique and a Ritual
// Magic spell, which always takes its best default, are left alone.
func TestSwapDefaultsOf(t *testing.T) {
	// restsOn returns the name the default the item recorded selects, the type of one that selects no skill or spell,
	// or "" for none.
	restsOn := func(h defaultable) string {
		def := h.recordedDefault()
		switch {
		case def == nil:
			return ""
		case def.SkillOrSpellBased():
			return def.NameWithReplacements(nil)
		default:
			return def.Type()
		}
	}
	// lore puts Thaumatology and the spell Lore on the sheet, each at 11 and defaulting to the other at -2, with
	// Thaumatology, updated first, recording Lore.
	lore := func(e *Entity) (thaumatology *Skill, spell *Spell) {
		thaumatology = addTestSkill(e, "Thaumatology", "", "", fxp.Four)
		thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Lore", -fxp.Two)}
		spell = addTestSpell(e, "Lore", fxp.Four)
		spell.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
		e.Recalculate()
		return thaumatology, spell
	}
	for _, tc := range []struct {
		name string
		// setup builds and recalculates the sheet, returning the item to swap and its partner, or nil where it has
		// none, whose defaults select, once the swap is done, what itemRestsOn and partnerRestsOn say (see restsOn).
		setup          func(c check.Checker, e *Entity) (item, partner defaultable)
		can            bool
		itemRestsOn    string
		partnerRestsOn string
	}{
		{
			name: "a skill trades places with the spell it defaults to", can: true, partnerRestsOn: "Thaumatology",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				thaumatology, spell := lore(e)
				c.Equal("Lore", restsOn(thaumatology), "precondition: the skill rests on the spell")
				return thaumatology, spell
			},
		},
		{
			name: "a spell trades places with the skill it defaults to", can: true, partnerRestsOn: "Lore",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				thaumatology, spell := lore(e)
				thaumatology.DefaultedFrom = nil
				spell.DefaultedFrom = newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)
				e.Recalculate()
				c.Equal("Thaumatology", restsOn(spell), "precondition: the spell rests on the skill")
				return spell, thaumatology
			},
		},
		{
			// Broadsword keeps Axe-4 at 11 over Mace-4 at 12, and both beat its Shortsword-2.
			name: "a skill trades places with the skill it defaults to, which rests on a third", can: true,
			partnerRestsOn: "Shortsword",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four)
				broadsword.Defaults = []*SkillDefault{
					newSkillDefaultTo("Axe", "", true, -fxp.Four),
					newSkillDefaultTo("Shortsword", "", true, -fxp.Two),
					newSkillDefaultTo("Mace", "", true, -fxp.Four),
				}
				broadsword.DefaultedFrom = newSkillDefaultTo("Axe", "", true, -fxp.Four)
				shortsword := addTestSkill(e, "Shortsword", "", "", fxp.One)
				shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
				addTestSkill(e, "Axe", "", "", fxp.FromInteger(20))
				addTestSkill(e, "Mace", "", "", fxp.FromInteger(24))
				e.Recalculate()
				c.Equal("Axe", restsOn(broadsword), "precondition: the skill defaulted to rests on a third")
				c.Equal("Broadsword", restsOn(shortsword), "precondition: the skill rests on the other")
				return shortsword, broadsword
			},
		},
		{
			name: "a spell trades places with the skill defaulting to it", can: true, itemRestsOn: "Thaumatology",
			setup: func(_ check.Checker, e *Entity) (item, partner defaultable) {
				thaumatology, spell := lore(e)
				return spell, thaumatology
			},
		},
		{
			// Bless is at 12 from its default to Thaumatology, above the 9 Ritual Magic (Fire) gives it, as a swap from
			// Thaumatology leaves it.
			name: "a skill trades places with the Ritual Magic spell defaulting to it", can: true, itemRestsOn: "Bless",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.One)
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.Two)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				bless.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				bless.DefaultedFrom = newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)
				e.Recalculate()
				c.Equal("Thaumatology", restsOn(bless), "precondition: the spell rests on the skill")
				return thaumatology, bless
			},
		},
		{
			// With no ritual skill, Bless has no level but the 12 its default to Thaumatology gives it, which leaves
			// nothing for Thaumatology to rest on once Bless drops that default.
			name:           "a skill can't trade places with a Ritual Magic spell that has no other level",
			partnerRestsOn: "Thaumatology",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.Two)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				bless.Points = fxp.Two
				bless.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				e.Recalculate()
				c.Equal("Thaumatology", restsOn(bless), "precondition: the spell rests on the skill")
				return thaumatology, bless
			},
		},
		{
			// Bless at 12 from its default to Thaumatology would beat the 9 Ritual Magic (Fire) gives it.
			name:           "a skill trades places with the Ritual Magic spell it defaults to",
			can:            true,
			partnerRestsOn: "Thaumatology",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.One)
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.Two)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				e.Recalculate()
				// Given its default only now, Bless can't take it, since Thaumatology already rests on Bless and keeps
				// it.
				bless.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				e.Recalculate()
				c.Equal("Bless", restsOn(thaumatology), "precondition: the skill rests on the spell")
				return thaumatology, bless
			},
		},
		{
			// Bless at 12 from its default to Thaumatology would only tie the 12 Ritual Magic (Fire) gives it, so its
			// level would not rest on that default.
			name:        "a skill can't trade places with a Ritual Magic spell whose ritual skill would outrank it",
			itemRestsOn: "Bless",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Eight)
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				thaumatology.Defaults = []*SkillDefault{newSpellDefaultTo("Bless", -fxp.Two)}
				bless := addTestRitualMagicSpell(e, "Bless", "Fire")
				bless.Defaults = []*SkillDefault{newSkillDefaultTo("Thaumatology", "", true, -fxp.Two)}
				e.Recalculate()
				c.Equal("Bless", restsOn(thaumatology), "precondition: the skill rests on the spell")
				return thaumatology, bless
			},
		},
		{
			name: "a spell trades places with the spell defaulting to it", can: true, itemRestsOn: "Lore",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				first := addTestSpell(e, "Lore", fxp.Four)
				first.Defaults = []*SkillDefault{newSpellDefaultTo("Wisdom", -fxp.Two)}
				second := addTestSpell(e, "Wisdom", fxp.Four)
				second.Defaults = []*SkillDefault{newSpellDefaultTo("Lore", -fxp.Two)}
				e.Recalculate()
				c.Equal("Wisdom", restsOn(first), "precondition: the spell listed first rests on the other")
				return second, first
			},
		},
		{
			// Broadsword's Axe-4 at 12 comes first, then its Mace-4 at 8, then its Shortsword-2 at 7, since Shortsword
			// is at 9 without the 11 its default to Broadsword gives it.
			name: "a skill moves to its next default ahead of one leading back to it", can: true, itemRestsOn: "Mace",
			partnerRestsOn: "Broadsword",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four)
				broadsword.Defaults = []*SkillDefault{
					newSkillDefaultTo("Axe", "", true, -fxp.Four),
					newSkillDefaultTo("Shortsword", "", true, -fxp.Two),
					newSkillDefaultTo("Mace", "", true, -fxp.Four),
				}
				shortsword := addTestSkill(e, "Shortsword", "", "", fxp.One)
				shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
				addTestSkill(e, "Axe", "", "", fxp.FromInteger(24))
				addTestSkill(e, "Mace", "", "", fxp.Eight)
				e.Recalculate()
				c.Equal("Axe", restsOn(broadsword), "precondition: the skill rests on its best default")
				c.Equal("Broadsword", restsOn(shortsword), "precondition: the other skill rests on it")
				return broadsword, shortsword
			},
		},
		{
			// Thaumatology-6 at 8 is the better default, over Occultism-6 at 4.
			name: "a spell moves to its next default", can: true, itemRestsOn: "Occultism",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				addTestSkill(e, "Occultism", "", "", fxp.Two)
				spell := addTestSpell(e, "Lore", fxp.Four)
				spell.Defaults = []*SkillDefault{
					newSkillDefaultTo("Thaumatology", "", true, -fxp.Six),
					newSkillDefaultTo("Occultism", "", true, -fxp.Six),
				}
				e.Recalculate()
				c.Equal("Thaumatology", restsOn(spell), "precondition: the spell rests on its better default")
				return spell, nil
			},
		},
		{
			name: "a granted spell is left alone", itemRestsOn: "Thaumatology",
			setup: func(_ check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				addTestSkill(e, "Occultism", "", "", fxp.Two)
				template := newHardSpell("Lore", fxp.Four)
				template.Defaults = []*SkillDefault{
					newSkillDefaultTo("Thaumatology", "", true, -fxp.Six),
					newSkillDefaultTo("Occultism", "", true, -fxp.Six),
				}
				addTraitWithFeatures(e, "Training", newSpellGrant(template, fxp.Two, grantmerge.Ignore))
				e.Recalculate()
				return grantedOf(e.Spells)[0], nil
			},
		},
		{
			name:           "a skill can't trade places with a granted skill defaulting to it",
			partnerRestsOn: "Shortsword",
			setup: func(c check.Checker, e *Entity) (item, partner defaultable) {
				shortsword := addTestSkill(e, "Shortsword", "", "", fxp.One)
				template := newEasySkill("Broadsword", fxp.One)
				template.Defaults = []*SkillDefault{newSkillDefaultTo("Shortsword", "", true, -fxp.Two)}
				addTraitWithFeatures(e, "Training", newSkillGrant(template, fxp.Four, grantmerge.Ignore))
				e.Recalculate()
				// Given its default only now, Shortsword can't take it, as the granted Broadsword already rests on
				// Shortsword.
				shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
				e.Recalculate()
				granted := grantedOf(e.Skills)[0]
				c.Equal("Shortsword", restsOn(granted), "precondition: the granted skill rests on the other")
				return shortsword, granted
			},
		},
		{
			// A technique's level rests on its technique default; these are what a file edited by hand may hold.
			name: "a technique is left alone",
			setup: func(_ check.Checker, e *Entity) (item, partner defaultable) {
				addTestSkill(e, "Karate", "", "", fxp.Four)
				kicking := NewTechnique(e, nil, "Karate")
				kicking.Points = fxp.Two
				kicking.Defaults = []*SkillDefault{
					{DefaultType: IntelligenceID, Modifier: -fxp.Two},
					{DefaultType: DexterityID, Modifier: -fxp.Two},
				}
				e.Skills = append(e.Skills, kicking)
				e.Recalculate()
				return kicking, nil
			},
		},
		{
			// IQ-2 and DX-2 are both 8, so the first is the best.
			name: "a Ritual Magic spell is left alone", itemRestsOn: IntelligenceID,
			setup: func(_ check.Checker, e *Entity) (item, partner defaultable) {
				rite := addTestRitualMagicSpell(e, "Bless", "Fire")
				rite.Points = fxp.Two
				rite.Defaults = []*SkillDefault{
					{DefaultType: IntelligenceID, Modifier: -fxp.Two},
					{DefaultType: DexterityID, Modifier: -fxp.Two},
				}
				e.Recalculate()
				return rite, nil
			},
		},
	} {
		c := check.New(t)
		e := NewEntity()
		item, partner := tc.setup(c, e)
		var can bool
		switch one := item.(type) {
		case *Skill:
			can = CanSwapDefaultsOf(one)
			SwapDefaultsOf(one)
		case *Spell:
			can = CanSwapDefaultsOf(one)
			SwapDefaultsOf(one)
		}
		c.Equal(tc.can, can, "%s: whether it can swap", tc.name)
		e.Recalculate()
		c.Equal(tc.itemRestsOn, restsOn(item), "%s: what it rests on afterwards", tc.name)
		if partner != nil {
			c.Equal(tc.partnerRestsOn, restsOn(partner), "%s: what its partner rests on afterwards", tc.name)
		}
	}
}
