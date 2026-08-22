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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

func addTestSkill(e *Entity, name, specialization, tl string, points fxp.Int) *Skill {
	sk := NewSkill(e, nil, false)
	sk.Name = name
	sk.Specialization = specialization
	sk.Difficulty.Attribute = IntelligenceID
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Points = points
	if tl != "" {
		tlv := tl
		sk.TechLevel = &tlv
	}
	e.Skills = append(e.Skills, sk)
	return sk
}

func newWhenTLSkillDefault(name, specialization string, when criteria.Number) *SkillDefault {
	def := &SkillDefault{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: name}},
		Modifier:    -fxp.Three,
		WhenTL:      when,
	}
	if specialization != "" {
		def.Specialization = criteria.Text{Compare: criteria.IsText, Qualifier: specialization}
	}
	return def
}

func numberCriteria(compare criteria.NumericComparison, qualifier fxp.Int) criteria.Number {
	return criteria.Number{Compare: compare, Qualifier: qualifier}
}

// TestSkillDefaultWhenTL verifies that a skill default's WhenTL constraint is evaluated against the tech level of the
// skill being defaulted *from* (the matched skill), not the character's tech level. See issue #1040.
func TestSkillDefaultWhenTL(t *testing.T) {
	c := check.New(t)

	// Character is at TL5, but holds a TL3 Machinist and a TL5 Smith (Iron).
	e := NewEntity()
	e.Profile.TechLevel = "5"
	addTestSkill(e, "Machinist", "", "3", fxp.Four)
	addTestSkill(e, "Smith", "Iron", "5", fxp.Four)
	addTestSkill(e, "Carpentry", "", "", fxp.Four) // no tech level, so it falls back to the character's TL
	e.Recalculate()

	// The matched skill's TL fails the constraint, so the default must not resolve (the prior bug used the
	// character's TL of 5, which would have let these through).
	c.Equal(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Machinist (TL3) must fail 'when TL at least 5'")
	c.Equal(fxp.Min, newWhenTLSkillDefault("Smith", "Iron",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevel(e, nil, true, nil, false),
		"Smith/Iron (TL5) must fail 'when TL at most 4'")

	// The matched skill's TL satisfies the constraint, so the default resolves to a real level.
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevel(e, nil, true, nil, false),
		"Machinist (TL3) must satisfy 'when TL at most 4'")
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Smith", "Iron",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Smith/Iron (TL5) must satisfy 'when TL at least 5'")

	// A default with no constraint always resolves against the matched skill.
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		criteria.Number{}).SkillLevel(e, nil, true, nil, false),
		"Machinist must resolve when there is no TL constraint")

	// SkillLevelFast (used for weapon defaults) must apply the same filtering.
	c.Equal(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevelFast(e, nil, true, nil, false),
		"SkillLevelFast: Machinist (TL3) must fail 'when TL at least 5'")
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Machinist", "",
		numberCriteria(criteria.AtMostNumber, fxp.Four)).SkillLevelFast(e, nil, true, nil, false),
		"SkillLevelFast: Machinist (TL3) must satisfy 'when TL at most 4'")

	// When the matched skill has no TL of its own, the constraint falls back to the character's TL (5 here).
	c.NotEqual(fxp.Min, newWhenTLSkillDefault("Carpentry", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Five)).SkillLevel(e, nil, true, nil, false),
		"Carpentry (no TL) must use the character's TL of 5")
	c.Equal(fxp.Min, newWhenTLSkillDefault("Carpentry", "",
		numberCriteria(criteria.AtLeastNumber, fxp.Six)).SkillLevel(e, nil, true, nil, false),
		"Carpentry (no TL) must use the character's TL of 5, which fails 'at least 6'")
}

// TestSkillDefaultUnresolvableAttribute verifies that an attribute-based default whose attribute the entity doesn't
// define keeps returning the fxp.Min sentinel, even when "use half stat defaults" is enabled. The conversion used to
// be applied unconditionally, turning the sentinel into a plausible-looking level.
func TestSkillDefaultUnresolvableAttribute(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.SheetSettings.UseHalfStatDefaults = true
	e.Attributes.Set[IntelligenceID].Adjustment = fxp.Four // IQ 14
	e.Recalculate()

	missing := &SkillDefault{DefaultType: "zz", Modifier: -fxp.Three}
	c.Equal(fxp.Min, missing.SkillLevelFast(e, nil, true, nil, false),
		"a default to an undefined attribute must remain unresolvable")
	c.Equal(fxp.Min, missing.SkillLevel(e, nil, true, nil, false),
		"SkillLevel must agree with SkillLevelFast")
	c.Equal(fxp.Min, missing.SkillLevelFast(e, nil, true, nil, true),
		"the rule of 20 must not resurrect an unresolvable default")

	// A defined attribute still gets the half-stat treatment: IQ 14 becomes 14/2+5 = 12, less the -3 modifier.
	iq := &SkillDefault{DefaultType: IntelligenceID, Modifier: -fxp.Three}
	c.Equal(fxp.Nine, iq.SkillLevelFast(e, nil, true, nil, false),
		"IQ 14 with half stat defaults resolves to 12, less the -3 modifier")
}

// TestSkillDefaultUnresolvableAttributeNotChosen verifies end-to-end that a skill won't adopt a default pointing at an
// attribute the entity doesn't define, which would persist a nonsense adjusted level.
func TestSkillDefaultUnresolvableAttributeNotChosen(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.SheetSettings.UseHalfStatDefaults = true
	sk := addTestSkill(e, "Hidden Lore", "Bogus", "", fxp.Four)
	sk.Defaults = []*SkillDefault{{DefaultType: "zz", Modifier: -fxp.Three}}
	e.Recalculate()

	c.Nil(sk.DefaultedFrom, "an unresolvable attribute default must not be adopted")
	c.Equal(0, len(sk.resolvableDefaults()), "an unresolvable attribute default must not be offered for swapping")
}

// TestSkillDefaultWhenTLChosenDefault reproduces issue #1040 end-to-end: an Armory (Body Armor) skill must not default
// to a skill whose tech level fails the default's WhenTL constraint, even after the character's TL is raised.
func TestSkillDefaultWhenTLChosenDefault(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.Profile.TechLevel = "5"
	addTestSkill(e, "Machinist", "", "3", fxp.Four)
	addTestSkill(e, "Smith", "Iron", "5", fxp.Four)

	armory := addTestSkill(e, "Armory", "Body Armor", "3", 0)
	armory.Defaults = []*SkillDefault{
		{
			DefaultType: IntelligenceID,
			Modifier:    -fxp.Five,
		},
		newWhenTLSkillDefault("Machinist", "", numberCriteria(criteria.AtLeastNumber, fxp.Five)),
		newWhenTLSkillDefault("Smith", "Iron", numberCriteria(criteria.AtMostNumber, fxp.Four)),
	}
	e.Recalculate()

	// Neither the TL3 Machinist (fails "at least 5") nor the TL5 Smith/Iron (fails "at most 4") is eligible, so the
	// Armory skill must fall back to its IQ-5 default rather than a skill default.
	c.NotNil(armory.DefaultedFrom, "Armory should have a resolved default")
	c.False(armory.DefaultedFrom.SkillBased(), "Armory must not default to a skill whose TL fails the constraint")
	c.Equal(IntelligenceID, armory.DefaultedFrom.DefaultType, "Armory must default to IQ")
}

// addTagMatchingSkillBonus installs a trait carrying a skill bonus that applies to any skill bearing the given tag.
func addTagMatchingSkillBonus(e *Entity, tag string, amount fxp.Int) {
	bonus := NewSkillBonus()
	bonus.NameCriteria.Compare = criteria.AnyText
	bonus.TagsCriteria = textCriteria(criteria.IsText, tag)
	bonus.Amount = amount
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
}

// TestSkillDefaultRemovesBonusUsingOtherSkillsTags verifies that the skill bonus baked into a default's level is undone
// using the *defaulted-to* skill's tags. calcSkillDefaultLevel passed the defaulting skill's tags instead, so a bonus
// that only the other skill qualified for was never removed (and one that only the defaulting skill qualified for was
// removed even though it had never been added). The wrong level is then persisted into DefaultedFrom.
func TestSkillDefaultRemovesBonusUsingOtherSkillsTags(t *testing.T) {
	c := check.New(t)

	// The +2 applies to Broadsword only. Its level carries the bonus, but the level Shortsword defaults from must not:
	// Broadsword IQ10-1+1pt = 9, +2 = 11; less the -2 default and the +2 bonus that doesn't apply to Shortsword = 7.
	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.One)
	broadsword.Tags = []string{"Combat"}
	shortsword := addTestSkill(e, "Shortsword", "", "", 0)
	shortsword.Tags = []string{"Weapon"}
	shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
	addTagMatchingSkillBonus(e, "Combat", fxp.Two)
	e.Recalculate()

	c.Equal(fxp.Eleven, broadsword.LevelData.Level, "Broadsword gets the Combat bonus")
	c.NotNil(shortsword.DefaultedFrom, "Shortsword should have a resolved default")
	c.Equal(fxp.Seven, shortsword.DefaultedFrom.Level, "the Combat bonus must be removed from the default's level")
	c.Equal(fxp.Seven, shortsword.DefaultedFrom.AdjLevel, "the persisted adjusted level must match")
	c.Equal(fxp.Seven, shortsword.LevelData.Level, "Shortsword must not inherit a bonus it doesn't qualify for")

	// Swapping which skill carries the tag: nothing was baked into Broadsword's level, so nothing may be removed, and
	// Shortsword still earns the bonus on its own. Broadsword = 9; less the -2 default = 7; plus Shortsword's own +2 = 9.
	e = NewEntity()
	broadsword = addTestSkill(e, "Broadsword", "", "", fxp.One)
	broadsword.Tags = []string{"Weapon"}
	shortsword = addTestSkill(e, "Shortsword", "", "", 0)
	shortsword.Tags = []string{"Combat"}
	shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
	addTagMatchingSkillBonus(e, "Combat", fxp.Two)
	e.Recalculate()

	c.Equal(fxp.Nine, broadsword.LevelData.Level, "Broadsword doesn't qualify for the Combat bonus")
	c.NotNil(shortsword.DefaultedFrom, "Shortsword should have a resolved default")
	c.Equal(fxp.Seven, shortsword.DefaultedFrom.Level, "no bonus was baked in, so none may be removed")
	c.Equal(fxp.Nine, shortsword.LevelData.Level, "Shortsword still earns the Combat bonus for itself")
}

// TestSkillDefaultTypeCaseInsensitive verifies that level resolution classifies the default type the same way
// everything else does. Only SetType() sanitizes the type, so a data file not written by GCS can hold "Parry"; such a
// default was displayed and treated as a parry default everywhere except in SkillLevel/SkillLevelFast, which switched
// on the exact string and dropped it into the attribute branch, where it silently never resolved.
func TestSkillDefaultTypeCaseInsensitive(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.Attributes.Set[DexterityID].Adjustment = fxp.Four // DX 14
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four)
	broadsword.Difficulty.Attribute = DexterityID
	e.Recalculate()

	withType := func(defaultType string) *SkillDefault {
		return &SkillDefault{
			DefaultType: defaultType,
			Name:        textCriteria(criteria.IsText, "Broadsword"),
			Modifier:    -fxp.Two,
		}
	}
	for _, canonical := range []string{SkillID, ParryID, BlockID, DodgeID, DexterityID} {
		expected := withType(canonical).SkillLevel(e, nil, false, nil, false)
		expectedFast := withType(canonical).SkillLevelFast(e, nil, false, nil, false)
		c.NotEqual(fxp.Min, expected, "%q must resolve to a real level", canonical)
		c.NotEqual(fxp.Min, expectedFast, "%q must resolve to a real level via SkillLevelFast", canonical)
		for _, variant := range []string{
			strings.ToUpper(canonical), " " + canonical + " ",
			strings.ToUpper(canonical[:1]) + canonical[1:],
		} {
			def := withType(variant)
			c.Equal(expected, def.SkillLevel(e, nil, false, nil, false),
				"%q must resolve the same as %q", variant, canonical)
			c.Equal(expectedFast, def.SkillLevelFast(e, nil, false, nil, false),
				"%q must resolve the same as %q via SkillLevelFast", variant, canonical)
			c.Equal(withType(canonical).SkillBased(), def.SkillBased(),
				"%q must classify the same as %q", variant, canonical)
			c.Equal(withType(canonical).FullName(e, nil), def.FullName(e, nil),
				"%q must name the same as %q", variant, canonical)
		}
	}

	// The defenses are re-pointed at each other by name, which is the check the weapon parry/block resolution makes.
	c.Equal(BlockID, withType("Parry").asDefense(BlockID).Type(), "a mixed-case parry default re-points to block")
	c.Equal(ParryID, withType("Parry").asDefense(ParryID).Type(), "an already-parry default is left alone")

	// A skill-based default must actually reach the named skill, not merely resolve to something.
	parry := withType("Parry").SkillLevel(e, nil, false, nil, false)
	c.Equal(broadsword.LevelData.Level.Div(fxp.Two).Floor()+fxp.Three+e.ParryBonus-fxp.Two, parry,
		"a parry default halves the named skill's level")
}
