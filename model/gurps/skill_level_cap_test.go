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
	"encoding/json/jsontext"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// addTestLevelCap adds a cap on everything of the target's kind that holds the given caps, the lowest winning, and
// returns it.
func addTestLevelCap(e *Entity, target levelcap.Target, caps ...LevelCap) *SkillLevelCap {
	c := NewSkillLevelCap()
	c.Target = target
	c.NameCriteria = criteria.Text{}
	for _, one := range caps {
		one.SetParentGroup(c.Cap)
		c.Cap.Children = append(c.Cap.Children, one)
	}
	e.SheetSettings.SkillLevelCaps = append(e.SheetSettings.SkillLevelCaps, c)
	return c
}

func fixedCap(level int) *FixedLevelCap {
	return NewFixedLevelCap(nil, fxp.FromInteger(level))
}

// refCap returns a cap at the level of the skill, spell or trait with the name, plus the modifier.
func refCap(capType levelcap.Type, name string, modifier int) *ReferenceLevelCap {
	c := NewReferenceLevelCap(nil, nil, capType)
	c.NameCriteria.Qualifier = name
	c.Modifier = fxp.FromInteger(modifier)
	return c
}

// attrCap returns a cap at the value of the attribute with the ID, plus the modifier.
func attrCap(id string, modifier int) *ReferenceLevelCap {
	c := NewReferenceLevelCap(nil, nil, levelcap.Attribute)
	c.Attribute = id
	c.Modifier = fxp.FromInteger(modifier)
	return c
}

// groupCap returns a group of the caps that takes the lowest of them, or the highest.
func groupCap(highest bool, caps ...LevelCap) *LevelCapGroup {
	g := NewLevelCapGroup(nil)
	g.Highest = highest
	for _, one := range caps {
		one.SetParentGroup(g)
		g.Children = append(g.Children, one)
	}
	return g
}

// addLeveledTrait adds an enabled, leveled trait with the name and levels to the entity.
func addLeveledTrait(e *Entity, name string, levels int) *Trait {
	t := addTraitWithFeatures(e, name)
	t.CanLevel = true
	t.Levels = fxp.FromInteger(levels)
	return t
}

// levelCapLine is the start of the line a capped level's tooltip gets.
const levelCapLine = "\nLevel cap of "

func cappedBy(level Level) bool {
	return strings.Contains(level.Tooltip, levelCapLine)
}

// nodeAs returns the node as a T, failing the test when it isn't one.
func nodeAs[T LevelCap](t *testing.T, node LevelCap) T {
	t.Helper()
	typed, ok := node.(T)
	if !ok {
		t.Fatalf("expected a %T, got a %T", typed, node)
	}
	return typed
}

// TestSkillLevelCapHoldsLevel verifies that a cap below a skill's or spell's level lowers the level and the relative
// level by the same amount and says so in the tooltip, while a cap at or above the level changes nothing. The cap
// holds a skill's level without its encumbrance penalty, which still applies to the level held.
func TestSkillLevelCapHoldsLevel(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name       string
		offset     int
		capped     bool
		encumbered bool
	}{
		{name: "two below", offset: -2, capped: true},
		{name: "at the level", offset: 0},
		{name: "one above", offset: 1},
		{name: "two below, encumbered", offset: -2, capped: true, encumbered: true},
	} {
		for _, spell := range []bool{false, true} {
			e := NewEntity()
			var level func() Level
			if spell {
				sp := addTestSpell(e, "Fireball", fxp.Twenty)
				level = func() Level { return sp.LevelData }
			} else {
				sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
				if one.encumbered {
					sk.EncumbrancePenaltyMultiplier = fxp.One
				}
				level = func() Level { return sk.LevelData }
			}
			var penalty fxp.Int
			if one.encumbered {
				addCarriedEquipmentWithFeatures(e, "Pack").BaseWeight = "100 lb"
				if !spell {
					penalty = e.EncumbranceLevel(true).Penalty()
					c.True(penalty < 0, "%s: the skill is encumbered", one.name)
				}
			}
			e.Recalculate()
			before := level()
			c.True(before.Level > fxp.Four, "%s: spell=%t: the uncapped level is high enough to cap", one.name, spell)
			target := levelcap.Skills
			if spell {
				target = levelcap.Spells
			}
			limit := before.Level - penalty + fxp.FromInteger(one.offset)
			addTestLevelCap(e, target, NewFixedLevelCap(nil, limit))
			e.Recalculate()
			after := level()
			if !one.capped {
				c.Equal(before, after, "%s: spell=%t: a cap no lower than the level changes nothing", one.name, spell)
				continue
			}
			c.Equal(limit+penalty, after.Level, "%s: spell=%t: the level is held to the cap", one.name, spell)
			c.Equal(before.RelativeLevel+fxp.FromInteger(one.offset), after.RelativeLevel,
				"%s: spell=%t: the relative level falls by as much", one.name, spell)
			c.Equal(before.Tooltip+levelCapLine+limit.String()+" [-2]", after.Tooltip,
				"%s: spell=%t: the tooltip says what the cap took", one.name, spell)
			checkRecalculationConsistency(c, e, one.name)
		}
	}
}

// TestSkillLevelCapMatcher verifies which skills and spells a cap selects: those of its target's kind that its name,
// specialization, tags, college and power source criteria match, with the specialization consulted only for skills
// and the college and power source only for spells. Techniques count as skills and Ritual Magic spells as spells.
func TestSkillLevelCapMatcher(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name     string
		target   levelcap.Target
		adjust   func(levelCap *SkillLevelCap)
		expected []string
	}{
		{
			name:     "every skill",
			target:   levelcap.Skills,
			expected: []string{"Broadsword", "Hidden Lore", "Feint", "Ritual Magic", "Ignite"},
		},
		{
			name:     "every spell",
			target:   levelcap.Spells,
			expected: []string{"Fireball", "Ignite"},
		},
		{
			name:     "every skill and spell",
			target:   levelcap.SkillsAndSpells,
			expected: []string{"Broadsword", "Hidden Lore", "Feint", "Ritual Magic", "Fireball", "Ignite"},
		},
		{
			name:   "skills by name",
			target: levelcap.Skills,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "lore"}
			},
			expected: []string{"Hidden Lore"},
		},
		{
			name:   "skills by specialization",
			target: levelcap.Skills,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Demons"}
			},
			expected: []string{"Hidden Lore"},
		},
		{
			name:   "skills by tag",
			target: levelcap.Skills,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Melee Combat"}
			},
			expected: []string{"Broadsword"},
		},
		{
			name:   "spells by college",
			target: levelcap.Spells,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.CollegeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
			},
			expected: []string{"Fireball", "Ignite"},
		},
		{
			name:   "spells by power source",
			target: levelcap.Spells,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.PowerSourceCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Divine"}
			},
			expected: []string{"Fireball"},
		},
		{
			name:   "a specialization narrows only the skills",
			target: levelcap.SkillsAndSpells,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Demons"}
			},
			expected: []string{"Hidden Lore", "Fireball", "Ignite"},
		},
		{
			name:   "a college narrows only the spells",
			target: levelcap.SkillsAndSpells,
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.CollegeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
			},
			expected: []string{"Broadsword", "Hidden Lore", "Feint", "Ritual Magic", "Fireball", "Ignite"},
		},
	} {
		e := NewEntity()
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four)
		broadsword.Tags = []string{"Combat", "Melee Combat"}
		addTestSkill(e, "Hidden Lore", "Demons", "", fxp.Four)
		feint := newTechniqueOf("Broadsword", "Feint", fxp.Two)
		feint.SetDataOwner(e)
		e.Skills = append(e.Skills, feint)
		fireball := addTestSpell(e, "Fireball", fxp.Four)
		fireball.College = CollegeList{"Fire"}
		fireball.PowerSource = "Divine"
		addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Four)
		addTestRitualMagicSpell(e, "Ignite", "Fire")
		e.Recalculate()
		Traverse(func(sk *Skill) bool {
			c.True(sk.LevelData.Level > fxp.One, "%s: skill %s starts above the cap", one.name, sk)
			return false
		}, false, true, e.Skills...)
		Traverse(func(sp *Spell) bool {
			c.True(sp.LevelData.Level > fxp.One, "%s: spell %s starts above the cap", one.name, sp)
			return false
		}, false, true, e.Spells...)
		levelCap := addTestLevelCap(e, one.target, fixedCap(1))
		if one.adjust != nil {
			one.adjust(levelCap)
		}
		e.Recalculate()
		// What the cap selects ends at the cap. So does the Ritual Magic spell whenever its ritual skill is capped, since
		// its level rests on that skill's, so the levels are looked at rather than the tooltips.
		var capped []string
		Traverse(func(sk *Skill) bool {
			if sk.LevelData.Level <= fxp.One {
				capped = append(capped, sk.NameWithReplacements())
			}
			return false
		}, false, true, e.Skills...)
		Traverse(func(sp *Spell) bool {
			if sp.LevelData.Level <= fxp.One {
				capped = append(capped, sp.NameWithReplacements())
			}
			return false
		}, false, true, e.Spells...)
		var expected []string
		for _, name := range []string{"Broadsword", "Hidden Lore", "Feint", "Ritual Magic", "Fireball", "Ignite"} {
			for _, want := range one.expected {
				if want == name {
					expected = append(expected, name)
				}
			}
		}
		c.Equal(expected, capped, "%s: the cap selects the right items", one.name)
	}
}

// TestSkillLevelCapReferences verifies the caps that rest on something on the sheet: the level of a skill, spell or
// trait, or the value of an attribute, plus the adjustment, with the highest matching item taken, a disabled trait
// ignored, the capped level of a referenced skill read, a level resting on the capped skill read as it is without that
// skill's cap, which the item with that level keeps, along with the points a grant's level set on it costs, and has
// as the prerequisites of an item carrying a grant see it too, a grant's level set on a technique priced against the
// skill it rests on as capped, even by a technique granted after it, or by one resting on it in turn, each reaching
// the level set whichever is listed first, a technique a grant puts on the sheet read as in
// effect once it is, a technique resting on a skill held to another technique resting on it reading that skill as
// held, a technique the default chain leaves with no level passed over, and a reference to nothing on the sheet, or to
// nothing with a level, counted as 0.
func TestSkillLevelCapReferences(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		// setup adds what the cap rests on and returns the caps and the level the capped skill should have, given its
		// uncapped level, along with the text the tooltip names the cap's source with.
		setup  func(e *Entity) (caps []LevelCap, expected func(uncapped fxp.Int) fxp.Int, source string)
		target levelcap.Target
		// restsOn, when set, names the skill or spell the cap rests on, whose level rests on a skill the cap holds and
		// so keeps the level and points it has without the cap, the level also as the prerequisites of an item carrying
		// a grant see it, which price a level set afresh.
		restsOn  string
		capSpell bool
		// costs gives, by name, the points of the skills grants put on the sheet whose level sets are priced against a
		// skill the cap holds.
		costs map[string]fxp.Int
		// levels gives, by name, the levels of the skills grants put on the sheet at a level that rests on a skill a cap
		// holds.
		levels map[string]fxp.Int
	}{
		{
			name: "a skill",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				e.Recalculate()
				level := thaumatology.LevelData.Level
				return []LevelCap{refCap(levelcap.Skill, "Thaumatology", -2)},
					func(fxp.Int) fxp.Int { return level - fxp.Two }, "Thaumatology -2"
			},
			target: levelcap.Skills,
		},
		{
			name: "a spell",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				fireball := addTestSpell(e, "Fireball", fxp.Four)
				e.Recalculate()
				level := fireball.LevelData.Level
				return []LevelCap{refCap(levelcap.Spell, "Fireball", -1)},
					func(fxp.Int) fxp.Int { return level - fxp.One }, "Fireball -1"
			},
			target: levelcap.Skills,
		},
		{
			name: "a technique resting on the skill the cap holds",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				feint := newTechniqueOf("Broadsword", "Feint", 0)
				feint.SetDataOwner(e)
				e.Skills = append(e.Skills, feint)
				return []LevelCap{refCap(levelcap.Skill, "Feint", -1)},
					func(uncapped fxp.Int) fxp.Int { return uncapped - fxp.One }, "Feint -1"
			},
			target:  levelcap.Skills,
			restsOn: "Feint",
		},
		{
			name: "a technique resting on the parry of the skill the cap holds",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				riposte := newTechniqueOf("Broadsword", "Riposte", 0)
				riposte.TechniqueDefault.DefaultType = ParryID
				riposte.SetDataOwner(e)
				e.Skills = append(e.Skills, riposte)
				return []LevelCap{refCap(levelcap.Skill, "Riposte", 0)},
					defenseLevelFromSkill, "Riposte"
			},
			target:  levelcap.Skills,
			restsOn: "Riposte",
		},
		{
			name: "a Ritual Magic spell resting on a skill the cap holds",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Four)
				ignite := addTestRitualMagicSpell(e, "Ignite", "Fire")
				e.Recalculate()
				level := ignite.LevelData.Level
				return []LevelCap{refCap(levelcap.Spell, "Ignite", -1)},
					func(fxp.Int) fxp.Int { return level - fxp.One }, "Ignite -1"
			},
			target:  levelcap.Skills,
			restsOn: "Ignite",
		},
		{
			// Feint is set to 17, 2 levels above Broadsword at 15, which costs 2 points.
			name: "a technique a grant sets the level of, resting on the skill the cap holds",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTraitWithFeatures(e, "Training", newSkillGrantOf(newTechniqueOf("Broadsword", "Feint", 0),
					grantadjust.SetLevel, fxp.FromInteger(17), grantmerge.Ignore))
				return []LevelCap{refCap(levelcap.Skill, "Feint", -3)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(14) }, "Feint -3"
			},
			target:  levelcap.Skills,
			restsOn: "Feint",
		},
		{
			// Kick is set to 17, 3 levels above Karate as the cap holds it, at Feint, set to 17, -3, which costs 3
			// points, though the grant of Feint comes after Kick's.
			name: "a technique a grant sets the level of, resting on a skill held to a technique granted after it",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Karate", "", "", fxp.Twenty)
				addTraitWithFeatures(e, "Training",
					newSkillGrantOf(newTechniqueOf("Karate", "Kick", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore),
					newSkillGrantOf(newTechniqueOf("Broadsword", "Feint", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore))
				return []LevelCap{refCap(levelcap.Skill, "Feint", -3)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(14) }, "Feint -3"
			},
			target:  levelcap.Skills,
			restsOn: "Feint",
			costs:   map[string]fxp.Int{"Kick": fxp.Three},
		},
		{
			// Ignite is set to 10, a level above its default, Ritual Magic (Fire) at 11 less its 2 prerequisites, which
			// costs 2 points.
			name: "a Ritual Magic spell a grant sets the level of, resting on a skill the cap holds",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Four)
				ignite := NewRitualMagicSpell(nil, nil, false)
				ignite.Name = "Ignite"
				ignite.College = CollegeList{"Fire"}
				ignite.PrereqCount = 2
				addTraitWithFeatures(e, "Magery", newSpellGrantOf(ignite, grantadjust.SetLevel, fxp.Ten,
					grantmerge.Ignore))
				return []LevelCap{refCap(levelcap.Spell, "Ignite", -2)},
					func(fxp.Int) fxp.Int { return fxp.Eight }, "Ignite -2"
			},
			target:  levelcap.Skills,
			restsOn: "Ignite",
		},
		{
			name: "a skill whose default rests on the capped skill",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				shortsword := addTestSkill(e, "Shortsword", "", "", 0)
				shortsword.Defaults = []*SkillDefault{{
					DefaultType: SkillID,
					Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Broadsword"}},
					Modifier:    -fxp.Two,
				}}
				return []LevelCap{refCap(levelcap.Skill, "Shortsword", 1)},
					func(uncapped fxp.Int) fxp.Int { return uncapped - fxp.One }, "Shortsword +1"
			},
			target: levelcap.Skills,
		},
		{
			name: "the highest of skills whose caps rest on each other",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.Twenty)
				addTestSkill(e, "Occultism", "", "", fxp.FromInteger(40))
				e.Recalculate()
				level := thaumatology.LevelData.Level
				return []LevelCap{groupCap(true, refCap(levelcap.Skill, "Thaumatology", -2),
						refCap(levelcap.Skill, "Occultism", 0))},
					func(fxp.Int) fxp.Int { return level - fxp.Two }, "Thaumatology -2"
			},
			target: levelcap.Skills,
		},
		{
			name: "the highest enabled trait",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addLeveledTrait(e, "Magery", 1)
				addLeveledTrait(e, "Magery", 3)
				addLeveledTrait(e, "Magery", 5).Disabled = true
				return []LevelCap{refCap(levelcap.Trait, "Magery", 10)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(13) }, "Magery 3 +10"
			},
			target: levelcap.Skills,
		},
		{
			name: "an attribute",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				return []LevelCap{attrCap(IntelligenceID, 1)},
					func(fxp.Int) fxp.Int { return e.ResolveAttributeCurrent(IntelligenceID) + fxp.One }, "IQ +1"
			},
			target: levelcap.Skills,
		},
		{
			name: "the dodge",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				return []LevelCap{attrCap(DodgeID, 0)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(e.Dodge(e.EncumbranceLevel(false))) }, "Dodge"
			},
			target: levelcap.Skills,
		},
		{
			name: "nothing on the sheet",
			setup: func(_ *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				return []LevelCap{refCap(levelcap.Skill, "Nope", -2)}, func(fxp.Int) fxp.Int { return -fxp.Two },
					"no skill Nope -2"
			},
			target: levelcap.Skills,
		},
		{
			name: "an attribute the sheet doesn't have",
			setup: func(_ *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				return []LevelCap{attrCap("nope", 3)}, func(fxp.Int) fxp.Int { return fxp.Three }, "no nope +3"
			},
			target: levelcap.Skills,
		},
		{
			name: "a spell with no level",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSpell(e, "Fireball", 0)
				return []LevelCap{refCap(levelcap.Spell, "Fireball", 1)}, func(fxp.Int) fxp.Int { return fxp.One },
					"no spell Fireball +1"
			},
			target: levelcap.Skills,
		},
		{
			name: "the capped level of a skill",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
				only := addTestLevelCap(e, levelcap.Skills, fixedCap(11))
				only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Thaumatology"}
				return []LevelCap{refCap(levelcap.Skill, "Thaumatology", -2)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(9) }, "Thaumatology -2"
			},
			target:   levelcap.Spells,
			capSpell: true,
		},
		{
			// Kick and Feint, each set to 17, rest on each other through the caps holding Karate and Broadsword, so the
			// grant of Feint is priced while Kick is not yet in effect, but the levels resting on Kick take it as in
			// effect once it is, and each is priced again with the other in effect, so that both reach the level set:
			// Broadsword is held to Kick -3 at 14, and Lightning to Broadsword -2.
			name: "the capped level of a skill held to a technique granted after one resting on the skill is priced",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
				addTestSkill(e, "Karate", "", "", fxp.Twenty)
				for _, held := range []struct{ skill, technique string }{{"Broadsword", "Kick"}, {"Karate", "Feint"}} {
					only := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, held.technique, -3))
					only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: held.skill}
				}
				addTraitWithFeatures(e, "Training",
					newSkillGrantOf(newTechniqueOf("Karate", "Kick", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore),
					newSkillGrantOf(newTechniqueOf("Broadsword", "Feint", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore))
				return []LevelCap{refCap(levelcap.Skill, "Broadsword", -2)},
					func(fxp.Int) fxp.Int { return fxp.Twelve }, "Broadsword -2"
			},
			target:   levelcap.Spells,
			capSpell: true,
			levels:   map[string]fxp.Int{"Kick": fxp.FromInteger(17), "Feint": fxp.FromInteger(17)},
		},
		{
			// As above, with the grants listed the other way round.
			name: "the capped level of a skill held to a technique granted before one resting on the skill is priced",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
				addTestSkill(e, "Karate", "", "", fxp.Twenty)
				for _, held := range []struct{ skill, technique string }{{"Broadsword", "Kick"}, {"Karate", "Feint"}} {
					only := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, held.technique, -3))
					only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: held.skill}
				}
				addTraitWithFeatures(e, "Training",
					newSkillGrantOf(newTechniqueOf("Broadsword", "Feint", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore),
					newSkillGrantOf(newTechniqueOf("Karate", "Kick", 0), grantadjust.SetLevel, fxp.FromInteger(17),
						grantmerge.Ignore))
				return []LevelCap{refCap(levelcap.Skill, "Broadsword", -2)},
					func(fxp.Int) fxp.Int { return fxp.Twelve }, "Broadsword -2"
			},
			target:   levelcap.Spells,
			capSpell: true,
			levels:   map[string]fxp.Int{"Kick": fxp.FromInteger(17), "Feint": fxp.FromInteger(17)},
		},
		{
			// Broadsword is held to Feint -1 at 14, and Riposte rests on it as held, though Feint, resting on it too, takes
			// it as it is without the cap, so Lightning is held to Riposte -2 at 12.
			name: "the level of a technique resting on a skill held to another technique resting on it",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
				for _, name := range []string{"Feint", "Riposte"} {
					technique := newTechniqueOf("Broadsword", name, 0)
					technique.SetDataOwner(e)
					e.Skills = append(e.Skills, technique)
				}
				only := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Feint", -1))
				only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"}
				return []LevelCap{refCap(levelcap.Skill, "Riposte", -2)},
					func(fxp.Int) fxp.Int { return fxp.Twelve }, "Riposte -2"
			},
			target:   levelcap.Spells,
			capSpell: true,
		},
		{
			// Shortsword defaults to Broadsword -2, which is held to Feint -1, Feint resting on Shortsword, so Shortsword's
			// default is taken with Feint passed over, at 13, and Lightning is held to Shortsword at 13.
			name: "the level of a skill defaulting to a skill held to a technique resting on it",
			setup: func(e *Entity) ([]LevelCap, func(fxp.Int) fxp.Int, string) {
				addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
				shortsword := addTestSkill(e, "Shortsword", "", "", fxp.One)
				shortsword.Defaults = []*SkillDefault{{
					DefaultType: SkillID,
					Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Broadsword"}},
					Modifier:    -fxp.Two,
				}}
				feint := newTechniqueOf("Shortsword", "Feint", 0)
				feint.SetDataOwner(e)
				e.Skills = append(e.Skills, feint)
				only := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Feint", -1))
				only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"}
				return []LevelCap{refCap(levelcap.Skill, "Shortsword", 0)},
					func(fxp.Int) fxp.Int { return fxp.FromInteger(13) }, "Shortsword"
			},
			target:   levelcap.Spells,
			capSpell: true,
		},
	} {
		e := NewEntity()
		var level func() Level
		if one.capSpell {
			sp := addTestSpell(e, "Lightning", fxp.Twenty)
			level = func() Level { return sp.LevelData }
		} else {
			sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
			level = func() Level { return sk.LevelData }
		}
		caps, expected, source := one.setup(e)
		var judge *Trait
		if one.restsOn != "" {
			judge = addTraitWithFeatures(e, "Insight",
				newTraitGrant(newGrantLeveledTrait("Focus", fxp.One), fxp.One, grantmerge.Ignore))
		}
		// restsOnLevel returns the level and points of the skill or spell named restsOn, and its level as the
		// prerequisites of judge see it.
		restsOnLevel := func() (level, points, judged fxp.Int) {
			level = fxp.Min
			Traverse(func(sk *Skill) bool {
				if sk.NameWithReplacements() == one.restsOn {
					level, points = sk.LevelData.Level, sk.AdjustedPoints(nil)
					e.JudgingPrereqsOf(judge, func() { judged = sk.levelForPrereqsOf(judge) })
				}
				return false
			}, false, true, e.Skills...)
			Traverse(func(sp *Spell) bool {
				if sp.NameWithReplacements() == one.restsOn {
					level, points = sp.LevelData.Level, sp.AdjustedPoints(nil)
					e.JudgingPrereqsOf(judge, func() { judged = sp.levelForPrereqsOf(judge) })
				}
				return false
			}, false, true, e.Spells...)
			return level, points, judged
		}
		e.Recalculate()
		uncapped := level().Level
		restsOn, restsOnPoints, _ := restsOnLevel()
		addTestLevelCap(e, one.target, caps...)
		e.Recalculate()
		want := expected(uncapped)
		c.Equal(want, level().Level, "%s: the level is held to the cap", one.name)
		c.True(strings.HasSuffix(level().Tooltip, levelCapLine+want.String()+" ("+source+") ["+
			(want-uncapped).StringWithSign()+"]"), "%s: the tooltip names the cap's source: %q", one.name,
			level().Tooltip)
		if one.restsOn != "" {
			level, points, judged := restsOnLevel()
			c.Equal(restsOn, level, "%s: what the cap rests on keeps its level", one.name)
			c.Equal(restsOnPoints, points, "%s: what the cap rests on keeps its points", one.name)
			c.Equal(restsOn, judged, "%s: the prerequisites of an item carrying a grant see that level", one.name)
		}
		for name, points := range one.costs {
			Traverse(func(sk *Skill) bool {
				if sk.NameWithReplacements() == name {
					c.Equal(points, sk.AdjustedPoints(nil), "%s: %s is priced against the capped skill", one.name, name)
				}
				return false
			}, false, true, e.Skills...)
		}
		for name, want := range one.levels {
			Traverse(func(sk *Skill) bool {
				if sk.NameWithReplacements() == name {
					c.Equal(want, sk.LevelData.Level, "%s: %s has the level set (%s points)", one.name, name,
						sk.AdjustedPoints(nil))
				}
				return false
			}, false, true, e.Skills...)
		}
		checkRecalculationConsistency(c, e, one.name)
	}
}

// TestSkillLevelCapReferenceSkipsItself verifies that a skill matched by a cap resting on its own level keeps that
// level, even when the reference also matches a skill with no level, while the other skills the cap matches are held
// to it.
func TestSkillLevelCapReferenceSkipsItself(t *testing.T) {
	c := check.New(t)
	for _, levelless := range []bool{false, true} {
		e := NewEntity()
		thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
		if levelless {
			addTestSkill(e, "Thaumatology", "Fire", "", 0)
		}
		broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
		e.Recalculate()
		level := thaumatology.LevelData.Level
		c.True(broadsword.LevelData.Level > level, "levelless=%t: the other skill starts above the cap", levelless)
		addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Thaumatology", -2))
		e.Recalculate()
		c.Equal(level, thaumatology.LevelData.Level, "levelless=%t: the skill the cap rests on keeps its level", levelless)
		c.False(cappedBy(thaumatology.LevelData), "levelless=%t: the skill the cap rests on is not capped", levelless)
		c.Equal(level-fxp.Two, broadsword.LevelData.Level, "levelless=%t: the other skill is held to the cap", levelless)
		checkRecalculationConsistency(c, e, "self reference")
	}
}

// TestSkillLevelCapGroups verifies how a group resolves: the lowest of its children, or the highest, a reference to
// nothing on the sheet counting as 0, with an empty group capping nothing, and nested groups.
func TestSkillLevelCapGroups(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name     string
		group    *LevelCapGroup
		expected int // 0 for no cap
	}{
		{name: "lowest", group: groupCap(false, fixedCap(12), fixedCap(10)), expected: 10},
		{name: "highest", group: groupCap(true, fixedCap(10), fixedCap(12)), expected: 12},
		{name: "highest with a missing reference", group: groupCap(true, refCap(levelcap.Skill, "Nope", 5), fixedCap(12)), expected: 12},
		{name: "lowest with a missing reference", group: groupCap(false, refCap(levelcap.Skill, "Nope", 5), fixedCap(12)), expected: 5},
		{name: "empty", group: groupCap(false)},
		{name: "nested", group: groupCap(false, groupCap(true, fixedCap(10), fixedCap(11)), fixedCap(13)), expected: 11},
	} {
		e := NewEntity()
		sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
		e.Recalculate()
		uncapped := sk.LevelData.Level
		c.True(uncapped > fxp.FromInteger(13), "%s: the uncapped level is above every cap", one.name)
		addTestLevelCap(e, levelcap.Skills, one.group)
		e.Recalculate()
		if one.expected == 0 {
			c.Equal(uncapped, sk.LevelData.Level, "%s: nothing caps the level", one.name)
			c.False(cappedBy(sk.LevelData), "%s: the tooltip is untouched", one.name)
		} else {
			c.Equal(fxp.FromInteger(one.expected), sk.LevelData.Level, "%s: the level is held to the group's result", one.name)
		}
	}
}

// TestSkillLevelCapLowestEntryWins verifies that of the caps that select an item the lowest applies, and one that
// doesn't select it, however low, is ignored.
func TestSkillLevelCapLowestEntryWins(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	addTestLevelCap(e, levelcap.Skills, fixedCap(12))
	addTestLevelCap(e, levelcap.Skills, fixedCap(10))
	addTestLevelCap(e, levelcap.Spells, fixedCap(5))
	e.Recalculate()
	c.Equal(fxp.Ten, sk.LevelData.Level, "the lowest cap that applies wins")
	c.True(strings.HasSuffix(sk.LevelData.Tooltip, levelCapLine+"10 [-5]"), "the tooltip names the cap that won: %q",
		sk.LevelData.Tooltip)
}

// TestSkillLevelCapSettles verifies that caps resting on capped skills settle in a pass rather than lowering the skill
// they rest on with every pass, that skills whose caps rest on each other settle too, and that nothing is left marked
// as having its caps resolved.
func TestSkillLevelCapSettles(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.FromInteger(16))
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	alpha := addTestSkill(e, "Alpha", "", "", fxp.FromInteger(8))
	beta := addTestSkill(e, "Beta", "", "", fxp.Twenty)
	e.Recalculate()
	thaumatologyLevel := thaumatology.LevelData.Level
	alphaLevel := alpha.LevelData.Level
	betaLevel := beta.LevelData.Level
	c.True(alphaLevel < betaLevel, "beta starts above alpha")
	addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Thaumatology", -2))
	onAlpha := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Beta", 0))
	onAlpha.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Alpha"}
	onBeta := addTestLevelCap(e, levelcap.Skills, refCap(levelcap.Skill, "Alpha", -1))
	onBeta.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Beta"}
	c.True(e.recalculate(maxRecalculationPassesInAll) <= 2, "the sheet settles at once")
	c.False(e.Unsettled(), "the sheet settles")
	c.Equal(thaumatologyLevel, thaumatology.LevelData.Level, "the skill the cap rests on keeps its level")
	c.Equal(thaumatologyLevel-fxp.Two, broadsword.LevelData.Level, "the other skill is held to the cap")
	c.Equal(alphaLevel.Min(betaLevel), alpha.LevelData.Level, "alpha rests on beta as it is without alpha's cap")
	c.Equal(betaLevel.Min(alphaLevel-fxp.One), beta.LevelData.Level, "beta rests on alpha as it is without beta's cap")
	for range 2 {
		e.Recalculate()
		c.Equal(thaumatologyLevel-fxp.Two, broadsword.LevelData.Level, "recalculating again changes nothing")
		c.Equal(alphaLevel.Min(betaLevel), alpha.LevelData.Level, "recalculating again changes nothing for alpha")
	}
	c.Equal(0, len(e.levelCapsResolving), "nothing is left marked as having its caps resolved")
	checkRecalculationConsistency(c, e, "settled")
}

// TestSkillLevelCapSeenByPrerequisites verifies that a prerequisite on a skill's level is judged against the capped
// level.
func TestSkillLevelCapSeenByPrerequisites(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	e.Recalculate()
	needs := addTraitWithFeatures(e, "Needs Broadsword")
	needs.Prereq = NewPrereqList()
	addSkillPrereq(needs.Prereq, "Broadsword", sk.LevelData.Level)
	e.Recalculate()
	c.Equal("", needs.UnsatisfiedReason, "the prerequisite is met at the uncapped level")
	addTestLevelCap(e, levelcap.Skills, NewFixedLevelCap(nil, sk.LevelData.Level-fxp.One))
	e.Recalculate()
	c.NotEqual("", needs.UnsatisfiedReason, "the prerequisite is unmet once the level is capped below it")
}

// TestSkillLevelCapSeenByDefaults verifies that a skill defaulting to a capped skill rests on the capped level.
func TestSkillLevelCapSeenByDefaults(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	shortsword := addTestSkill(e, "Shortsword", "", "", 0)
	shortsword.Defaults = []*SkillDefault{{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Broadsword"}},
		Modifier:    -fxp.Two,
	}}
	e.Recalculate()
	c.Equal(broadsword.LevelData.Level-fxp.Two, shortsword.LevelData.Level, "the default rests on the uncapped level")
	only := addTestLevelCap(e, levelcap.Skills, NewFixedLevelCap(nil, broadsword.LevelData.Level-fxp.Three))
	only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"}
	e.Recalculate()
	c.Equal(broadsword.LevelData.Level-fxp.Two, shortsword.LevelData.Level, "the default rests on the capped level")
	c.False(cappedBy(shortsword.LevelData), "the defaulting skill is not itself capped")
}

// TestSkillLevelCapStepping verifies that stepping a capped skill's level up adds no points, since no points raise
// it, and that stepping it down lands a level below the cap.
func TestSkillLevelCapStepping(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	addTestLevelCap(e, levelcap.Skills, fixedCap(13))
	e.Recalculate()
	c.Equal(fxp.FromInteger(13), sk.LevelData.Level, "the skill starts at the cap")
	sk.IncrementSkillLevel()
	c.Equal(fxp.Twenty, sk.RawPoints(), "no points are added when none raise the level")
	sk.DecrementSkillLevel()
	e.Recalculate()
	c.Equal(fxp.Twelve, sk.LevelData.Level, "stepping down lands a level below the cap")
	c.True(sk.RawPoints() < fxp.Twenty, "stepping down takes points away")
}

// TestSkillLevelCapEditPreview verifies that the level an editor previews is held to the caps that select the skill
// or spell as edited rather than as it is on the sheet, with the references matching it as edited and the levels they
// rest on, such as a technique's or a default's resting on the skill, taken with it as edited, and the row of a grant
// the edits merge into it left out, so that the preview is the level the sheet shows once the edits are applied.
func TestSkillLevelCapEditPreview(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := addTestSkill(e, "Broadsword", "", "", fxp.Twenty)
	sp := addTestSpell(e, "Fireball", fxp.Twenty)
	only := addTestLevelCap(e, levelcap.SkillsAndSpells, fixedCap(5))
	only.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Capped"}
	e.Recalculate()
	level := Level{Level: fxp.FromInteger(15), RelativeLevel: fxp.Five}
	levelAt := func(fxp.Int) Level { return level }
	skillData := sk.SkillEditData
	c.Equal(level, sk.LevelWithGrants(&skillData, level, levelAt), "a skill the caps don't select keeps its level")
	skillData.Name = "Capped"
	c.Equal(fxp.Five, sk.LevelWithGrants(&skillData, level, levelAt).Level, "a skill renamed into a cap is held to it")
	spellData := sp.SpellEditData
	c.Equal(level, sp.LevelWithGrants(&spellData, level, levelAt), "a spell the caps don't select keeps its level")
	spellData.Name = "Capped"
	c.Equal(fxp.Five, sp.LevelWithGrants(&spellData, level, levelAt).Level, "a spell renamed into a cap is held to it")
	// feint adds Feint, a technique resting on Broadsword, and returns a cap resting on it.
	feint := func(e *Entity) LevelCap {
		technique := newTechniqueOf("Broadsword", "Feint", 0)
		technique.SetDataOwner(e)
		e.Skills = append(e.Skills, technique)
		return refCap(levelcap.Skill, "Feint", -1)
	}
	for _, one := range []struct {
		name     string
		from, to string
		// points are those the skill has on the sheet, which the edits make 20.
		points fxp.Int
		// setup adds what the cap rests on, when that isn't Thaumatology, and returns the cap.
		setup  func(e *Entity) LevelCap
		capped bool
	}{
		{name: "a skill renamed out of a reference", from: "Thaumatology", to: "Occultism", points: fxp.Twenty, capped: true},
		{name: "a skill renamed into a reference", from: "Broadsword", to: "Thaumatology", points: fxp.Twenty},
		{
			name: "a skill a technique the cap rests on rests on", from: "Broadsword", to: "Broadsword", points: fxp.Four,
			setup: feint, capped: true,
		},
		{
			name: "a skill whose block a technique the cap rests on rests on", from: "Broadsword", to: "Broadsword",
			points: fxp.One,
			setup: func(e *Entity) LevelCap {
				technique := newTechniqueOf("Broadsword", "Cross Block", 0)
				technique.TechniqueDefault.DefaultType = BlockID
				technique.SetDataOwner(e)
				e.Skills = append(e.Skills, technique)
				return refCap(levelcap.Skill, "Cross Block", 3)
			},
			capped: true,
		},
		{
			name: "a skill renamed into a grant a technique the cap rests on rests on", from: "Saber", to: "Broadsword",
			points: fxp.Twenty,
			setup: func(e *Entity) LevelCap {
				granted := NewSkill(nil, nil, false)
				granted.Name = "Broadsword"
				granted.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Average}
				granted.Points = fxp.One
				addTraitWithFeatures(e, "Training", newSkillGrant(granted, fxp.Three, grantmerge.Ignore))
				return feint(e)
			},
			capped: true,
		},
		{
			// The grant merges into the skill once it is renamed, which takes the grant's own row off the sheet.
			name: "a skill renamed into a grant that keeps it at a level the cap rests on", from: "Saber",
			to: "Thaumatology", points: fxp.Twenty,
			setup: func(e *Entity) LevelCap {
				granted := NewSkill(nil, nil, false)
				granted.Name = "Thaumatology"
				granted.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Average}
				addTraitWithFeatures(e, "Training", newSkillGrantOf(granted, grantadjust.SetLevel, fxp.FromInteger(18),
					grantmerge.KeepHighest))
				return refCap(levelcap.Skill, "Thaumatology", -3)
			},
		},
		{
			// The grant merges into the skill once it is renamed, since it comes first, rather than into the other one.
			name: "a skill renamed into a grant that leaves the skill the cap rests on", from: "Saber",
			to: "Thaumatology", points: fxp.Twenty,
			setup: func(e *Entity) LevelCap {
				addTestSkill(e, "Thaumatology", "", "", fxp.Twenty)
				granted := NewSkill(nil, nil, false)
				granted.Name = "Thaumatology"
				granted.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Average}
				addTraitWithFeatures(e, "Training", newSkillGrant(granted, fxp.Three, grantmerge.Ignore))
				return refCap(levelcap.Skill, "Thaumatology", -3)
			},
			capped: true,
		},
		{
			// The grant merges into the skill, which comes first, until it is renamed, and then into the other one.
			name: "a skill renamed out of a grant that moves to the skill the cap rests on", from: "Thaumatology",
			to: "Occultism", points: fxp.Twenty,
			setup: func(e *Entity) LevelCap {
				addTestSkill(e, "Thaumatology", "", "", fxp.Twenty)
				granted := NewSkill(nil, nil, false)
				granted.Name = "Thaumatology"
				granted.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Average}
				addTraitWithFeatures(e, "Training", newSkillGrant(granted, fxp.Three, grantmerge.Ignore))
				return refCap(levelcap.Skill, "Thaumatology", -3)
			},
		},
		{
			// The grant merges into the skill until it is renamed, and then puts its own row on the sheet.
			name: "a skill renamed out of a grant that then grants the skill the cap rests on", from: "Thaumatology",
			to: "Occultism", points: fxp.Twenty,
			setup: func(e *Entity) LevelCap {
				granted := NewSkill(nil, nil, false)
				granted.Name = "Thaumatology"
				granted.Difficulty = AttributeDifficulty{Attribute: IntelligenceID, Difficulty: difficulty.Average}
				addTraitWithFeatures(e, "Training", newSkillGrant(granted, fxp.Three, grantmerge.Ignore))
				return refCap(levelcap.Skill, "Thaumatology", -3)
			},
			capped: true,
		},
		{
			name: "a skill bought that a default the cap rests on rests on", from: "Broadsword", to: "Broadsword",
			setup: func(e *Entity) LevelCap {
				shortsword := addTestSkill(e, "Shortsword", "", "", 0)
				shortsword.Defaults = []*SkillDefault{{
					DefaultType: SkillID,
					Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Broadsword"}},
					Modifier:    -fxp.Two,
				}}
				return refCap(levelcap.Skill, "Shortsword", 1)
			},
			capped: true,
		},
	} {
		e = NewEntity()
		sk = addTestSkill(e, one.from, "", "", one.points)
		ref := LevelCap(refCap(levelcap.Skill, "Thaumatology", -2))
		if one.setup != nil {
			ref = one.setup(e)
		}
		addTestLevelCap(e, levelcap.Skills, ref)
		e.Recalculate()
		skillData = sk.SkillEditData
		skillData.Name = one.to
		skillData.Points = fxp.Twenty
		edited := sk.withEditData(&skillData)
		uncapped := edited.levelBeforeGrants(nil, skillData.DefaultedFrom, nil)
		preview := sk.LevelWithGrants(&skillData, uncapped, func(points fxp.Int) Level {
			return edited.levelAt(e, nil, points, nil)
		})
		sk.SkillEditData = skillData
		e.Recalculate()
		c.Equal(one.capped, cappedBy(sk.LevelData), "%s: the edited skill is capped on the sheet", one.name)
		c.Equal(sk.LevelData.Level, preview.Level, "%s: the preview is the level the sheet shows", one.name)
	}
	for _, one := range []struct {
		name     string
		from, to string
		modifier int
		// grant, when set, adds a grant that keeps Fireball at 18, which merges into the spell once it is renamed,
		// taking the grant's own row off the sheet.
		grant bool
		// rival, when set, adds Fireball after the spell.
		rival bool
		// levels, when set, adds a grant that adds 3 levels to Fireball, which merges into the first Fireball on the
		// sheet, or puts its own row there when there is none.
		levels bool
		capped bool
	}{
		{name: "a spell renamed out of a reference", from: "Fireball", to: "Ignite", modifier: -1, capped: true},
		{
			name: "a spell renamed into a grant that keeps it at a level the cap rests on", from: "Flash", to: "Fireball",
			modifier: -3, grant: true,
		},
		{
			name: "a spell renamed into a grant that leaves the spell the cap rests on", from: "Flash", to: "Fireball",
			modifier: -3, rival: true, levels: true, capped: true,
		},
		{
			name: "a spell renamed out of a grant that moves to the spell the cap rests on", from: "Fireball",
			to: "Ignite", modifier: -3, rival: true, levels: true,
		},
		{
			name: "a spell renamed out of a grant that then grants the spell the cap rests on", from: "Fireball",
			to: "Ignite", modifier: -3, levels: true, capped: true,
		},
	} {
		e = NewEntity()
		sp = addTestSpell(e, one.from, fxp.Twenty)
		if one.grant {
			granted := NewSpell(nil, nil, false)
			granted.Name = "Fireball"
			addTraitWithFeatures(e, "Training", newSpellGrantOf(granted, grantadjust.SetLevel, fxp.FromInteger(18),
				grantmerge.KeepHighest))
		}
		if one.rival {
			addTestSpell(e, "Fireball", fxp.Twenty)
		}
		if one.levels {
			granted := NewSpell(nil, nil, false)
			granted.Name = "Fireball"
			addTraitWithFeatures(e, "Training", newSpellGrant(granted, fxp.Three, grantmerge.Ignore))
		}
		addTestLevelCap(e, levelcap.Spells, refCap(levelcap.Spell, "Fireball", one.modifier))
		e.Recalculate()
		spellData = sp.SpellEditData
		spellData.Name = one.to
		edited := sp.withEditData(&spellData)
		preview := sp.LevelWithGrants(&spellData, edited.levelAt(e, spellData.Points), func(points fxp.Int) Level {
			return edited.levelAt(e, points)
		})
		sp.SpellEditData = spellData
		e.Recalculate()
		c.Equal(one.capped, cappedBy(sp.LevelData), "%s: the renamed spell is capped on the sheet", one.name)
		c.Equal(sp.LevelData.Level, preview.Level, "%s: the preview is the level the sheet shows", one.name)
	}
}

// TestLevelCapDescribe verifies the sentences caps and their nodes read as.
func TestLevelCapDescribe(t *testing.T) {
	c := check.New(t)
	thaumatology := refCap(levelcap.Skill, "Thaumatology", -2)
	thaumatology.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
	for _, one := range []struct {
		name     string
		node     LevelCap
		expected string
	}{
		{name: "fixed", node: fixedCap(12), expected: "12"},
		{name: "skill", node: thaumatology, expected: "skill Thaumatology (Fire) -2"},
		{name: "spell", node: refCap(levelcap.Spell, "Enchantment", -2), expected: "spell Enchantment -2"},
		{name: "trait", node: refCap(levelcap.Trait, "Magery", 10), expected: "trait Magery +10"},
		{name: "attribute", node: attrCap(IntelligenceID, 2), expected: "IQ +2"},
		{name: "no modifier", node: refCap(levelcap.Skill, "Thaumatology", 0), expected: "skill Thaumatology"},
		{name: "lowest", node: groupCap(false, fixedCap(12), refCap(levelcap.Skill, "Thaumatology", -2)), expected: "the lowest of 12 and skill Thaumatology -2"},
		{name: "highest of three", node: groupCap(true, fixedCap(1), fixedCap(2), fixedCap(3)), expected: "the highest of 1, 2 and 3"},
		{name: "one child", node: groupCap(true, fixedCap(12)), expected: "12"},
		{name: "empty", node: groupCap(false), expected: "no cap"},
		{name: "empty subgroup left out", node: groupCap(false, groupCap(true), fixedCap(12)), expected: "12"},
		{name: "unknown", node: NewUnknownLevelCap("future", jsontext.Value(`{"type":"future"}`)), expected: `an unknown kind of level cap ("future") that needs a newer version of GCS`},
	} {
		c.Equal(one.expected, one.node.Describe(nil, plainText), one.name)
	}
	for _, one := range []struct {
		name     string
		adjust   func(levelCap *SkillLevelCap)
		expected string
	}{
		{name: "every skill", expected: "Skills are capped at 12"},
		{
			name: "named skill with a specialization",
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Hidden Lore"}
				levelCap.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Demons"}
			},
			expected: "Skills named Hidden Lore (Demons) are capped at 12",
		},
		{
			name: "skills and spells by name and tag",
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.Target = levelcap.SkillsAndSpells
				levelCap.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Magic"}
				levelCap.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Occult"}
			},
			expected: `Skills and spells whose name contains "Magic" tagged Occult are capped at 12`,
		},
		{
			name: "spells by college and power source",
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.Target = levelcap.Spells
				levelCap.CollegeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
				levelCap.PowerSourceCriteria = criteria.Text{Compare: criteria.StartsWithText, Qualifier: "Arc"}
			},
			expected: `Spells whose college is Fire whose power source starts with "Arc" are capped at 12`,
		},
		{
			name: "spell criteria left out of a cap on skills",
			adjust: func(levelCap *SkillLevelCap) {
				levelCap.CollegeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
				levelCap.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Demons"}
				levelCap.Target = levelcap.Spells
			},
			expected: "Spells whose college is Fire are capped at 12",
		},
	} {
		levelCap := NewSkillLevelCap()
		levelCap.Cap.Children = LevelCaps{fixedCap(12)}
		if one.adjust != nil {
			one.adjust(levelCap)
		}
		c.Equal(one.expected, levelCap.Describe(nil, plainText), one.name)
	}
}

// newTestSkillLevelCap returns a cap holding every kind of node this version knows, nested.
func newTestSkillLevelCap() *SkillLevelCap {
	levelCap := NewSkillLevelCap()
	levelCap.Target = levelcap.SkillsAndSpells
	levelCap.NameCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Magic"}
	levelCap.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Occult"}
	levelCap.CollegeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
	levelCap.PowerSourceCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Arcane"}
	thaumatology := refCap(levelcap.Skill, "Thaumatology", -2)
	thaumatology.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}
	levelCap.Cap = groupCap(false, fixedCap(12), thaumatology,
		groupCap(true, refCap(levelcap.Spell, "Enchantment", 1), refCap(levelcap.Trait, "Magery", 10),
			attrCap(IntelligenceID, 2)))
	levelCap.EnsureValidity()
	return levelCap
}

// TestSkillLevelCapsJSON verifies that sheet settings holding every kind of cap node round trip through JSON, with
// the nodes decoded to their types and their parents set, and a node of a kind this version doesn't know preserved as
// it was, as is a target this version doesn't know, which selects nothing.
func TestSkillLevelCapsJSON(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	levelCap := newTestSkillLevelCap()
	unknown := NewUnknownLevelCap("future", jsontext.Value(`{"type":"future","foo":1}`))
	unknown.SetParentGroup(levelCap.Cap)
	levelCap.Cap.Children = append(levelCap.Cap.Children, unknown)
	s.SkillLevelCaps = SkillLevelCaps{levelCap}
	data, err := jio.Marshal(s)
	c.NoError(err)
	c.Contains(string(data), `"skill_level_caps"`)
	var loaded SheetSettings
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(1, len(loaded.SkillLevelCaps), "the cap is loaded")
	levelCap = loaded.SkillLevelCaps[0]
	c.Equal(Hash64(s.SkillLevelCaps[0]), Hash64(levelCap), "the cap loads as it was saved")
	root := levelCap.Cap
	c.Equal(4, len(root.Children), "every child is loaded")
	c.Equal(levelcap.Fixed, root.Children[0].CapType())
	c.Equal(levelcap.Skill, root.Children[1].CapType())
	c.Equal(levelcap.Group, root.Children[2].CapType())
	c.Equal(levelcap.Unknown, root.Children[3].CapType())
	sub := nodeAs[*LevelCapGroup](t, root.Children[2])
	c.True(sub.Highest, "the nested group keeps its kind")
	c.Equal(levelcap.Spell, sub.Children[0].CapType())
	c.Equal(levelcap.Trait, sub.Children[1].CapType())
	c.Equal(levelcap.Attribute, sub.Children[2].CapType())
	c.Equal(IntelligenceID, nodeAs[*ReferenceLevelCap](t, sub.Children[2]).Attribute)
	for _, child := range root.Children {
		c.Equal(root, child.ParentGroup(), "each child knows its group")
	}
	for _, child := range sub.Children {
		c.Equal(sub, child.ParentGroup(), "each nested child knows its group")
	}
	c.Equal("future", nodeAs[*UnknownLevelCap](t, root.Children[3]).Kind)
	again, err := jio.Marshal(&loaded)
	c.NoError(err)
	c.Equal(string(data), string(again), "saving what was loaded writes the same JSON")
	c.Contains(string(again), `{"type":"future","foo":1}`, "the unknown node is written back as it was")
	data, err = jio.Marshal(loaded.SkillLevelCaps)
	c.NoError(err)
	future := strings.Replace(string(data), `"target":"skills_and_spells"`, `"target":"techniques"`, 1)
	var futureCaps SkillLevelCaps
	c.NoError(jio.Unmarshal([]byte(future), &futureCaps))
	c.Equal(1, len(futureCaps), "the cap with an unknown target is loaded")
	subject := &levelCapSubject{name: "Magic", tags: []string{"Occult"}, colleges: []string{"Fire"}, powerSource: "Arcane"}
	for _, spell := range []bool{false, true} {
		subject.spell = spell
		c.True(levelCap.applies(subject), "spell=%t: the known target selects the subject", spell)
		c.False(futureCaps[0].applies(subject), "spell=%t: an unknown target selects nothing", spell)
	}
	again, err = jio.Marshal(futureCaps)
	c.NoError(err)
	c.Equal(future, string(again), "an unknown target is written back as it was read")
}

// TestSkillLevelCapsSkipWhatCanNotBeRead verifies that a cap that can't be read is left out rather than failing the
// load of the settings holding it.
func TestSkillLevelCapsSkipWhatCanNotBeRead(t *testing.T) {
	c := check.New(t)
	var s SheetSettings
	c.NoError(jio.Unmarshal([]byte(`{"skill_level_caps":[{"target":"skills","cap":5},{"target":"spells","cap":{"type":"group"}}]}`), &s))
	c.Equal(1, len(s.SkillLevelCaps), "the cap that can't be read is left out")
	c.Equal(levelcap.Spells, s.SkillLevelCaps[0].Target, "the cap that can be read is kept")
	c.NoError(jio.Unmarshal([]byte(`{"skill_level_caps":5}`), &s))
	c.Equal(0, len(s.SkillLevelCaps), "caps that aren't a list are left out")
}

// TestSheetSettingsSkillLevelCapsDefaultToNone verifies that the factory settings have no caps and write none.
func TestSheetSettingsSkillLevelCapsDefaultToNone(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	c.Equal(0, len(s.SkillLevelCaps))
	data, err := jio.Marshal(s)
	c.NoError(err)
	c.False(strings.Contains(string(data), "skill_level_caps"), "no caps are written when there are none")
}

// TestSheetSettingsSkillLevelCapsClone verifies that a clone of the settings holds copies of the caps, so that editing
// either leaves the other alone, with the copies' nodes owned by the copies.
func TestSheetSettingsSkillLevelCapsClone(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	s.SkillLevelCaps = SkillLevelCaps{newTestSkillLevelCap()}
	before := Hash64(s.SkillLevelCaps[0])
	clone := s.Clone(nil)
	c.Equal(before, Hash64(clone.SkillLevelCaps[0]), "the clone holds the same caps")
	cloned := clone.SkillLevelCaps[0]
	nodeAs[*FixedLevelCap](t, cloned.Cap.Children[0]).Level = fxp.One
	cloned.Cap.Children = append(cloned.Cap.Children, fixedCap(3))
	cloned.Target = levelcap.Spells
	c.Equal(before, Hash64(s.SkillLevelCaps[0]), "editing the clone leaves the original alone")
	c.Equal(cloned.Cap, cloned.Cap.Children[0].ParentGroup(), "the clone's nodes are owned by the clone")
	c.Equal(cloned.Cap, nodeAs[*LevelCapGroup](t, cloned.Cap.Children[2]).Children[0].ParentGroup().ParentGroup(),
		"the clone's nested nodes are owned by the clone")
}

// TestSheetSettingsSkillLevelCapsEnsureValidity verifies that making the settings valid drops a missing cap, gives one
// without a root a root, drops missing nodes, sets the parents and types of the rest and makes a reference of no
// reference type a skill reference.
func TestSheetSettingsSkillLevelCapsEnsureValidity(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	s.SkillLevelCaps = SkillLevelCaps{
		nil,
		{Target: levelcap.Target(99)},
		{Cap: &LevelCapGroup{Children: LevelCaps{nil, &ReferenceLevelCap{
			Type:                   levelcap.Group,
			SpecializationCriteria: criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"}, Attribute: IntelligenceID,
		}}}},
	}
	s.EnsureValidity()
	c.Equal(2, len(s.SkillLevelCaps), "the missing cap is dropped")
	c.Equal(levelcap.UnknownTarget, s.SkillLevelCaps[0].Target, "an invalid target becomes unknown")
	c.NotNil(s.SkillLevelCaps[0].Cap, "a cap without a root gets one")
	c.Equal(levelcap.Group, s.SkillLevelCaps[0].Cap.Type)
	root := s.SkillLevelCaps[1].Cap
	c.Equal(levelcap.Group, root.Type, "the root's type is set")
	c.Equal(1, len(root.Children), "the missing node is dropped")
	ref := nodeAs[*ReferenceLevelCap](t, root.Children[0])
	c.Equal(levelcap.Skill, ref.Type, "a reference of no reference type becomes a skill reference")
	c.Equal(root, ref.Parent, "the node's parent is set")
	c.Equal("", ref.Attribute, "a skill reference holds no attribute")
	c.Equal("Fire", ref.SpecializationCriteria.Qualifier, "a skill reference keeps its specialization")
}

// TestSkillLevelCapsFileRoundTrip verifies that caps saved to a file of their own load back as they were saved, valid
// and at the current data version, and that a file holding none loads as none.
func TestSkillLevelCapsFileRoundTrip(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	name := "caps" + SkillLevelCapsExt
	caps := SkillLevelCaps{newTestSkillLevelCap(), NewSkillLevelCap()}
	c.NoError(SaveSkillLevelCaps(caps, filepath.Join(dir, name)))
	data, err := os.ReadFile(filepath.Join(dir, name))
	c.NoError(err)
	c.Contains(string(data), `"version": `+strconv.Itoa(jio.CurrentDataVersion))
	loaded, err := NewSkillLevelCapsFromFile(os.DirFS(dir), name)
	c.NoError(err)
	c.Equal(len(caps), len(loaded))
	for i := range caps {
		c.Equal(Hash64(caps[i]), Hash64(loaded[i]), "cap %d loads as it was saved", i)
		for _, child := range loaded[i].Cap.Children {
			c.Equal(loaded[i].Cap, child.ParentGroup(), "cap %d: the loaded nodes are owned by the loaded root", i)
		}
	}
	c.NoError(SaveSkillLevelCaps(nil, filepath.Join(dir, name)))
	loaded, err = NewSkillLevelCapsFromFile(os.DirFS(dir), name)
	c.NoError(err)
	c.Equal(0, len(loaded), "a file holding no caps loads as none")
}
