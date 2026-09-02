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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
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

// addTaggedTestSkill adds a skill carrying the given tags. It is otherwise identical to addTestSkill: an IQ/Average
// skill with no tech level, so that a given point total always produces the same level.
func addTaggedTestSkill(e *Entity, name, specialization string, points fxp.Int, tags ...string) *Skill {
	sk := addTestSkill(e, name, specialization, "", points)
	sk.Tags = tags
	return sk
}

// newTaggedSkillDefault creates a skill default that names no skill outright, selecting its skills by tag instead.
func newTaggedSkillDefault(compare criteria.StringComparison, qualifier string, modifier fxp.Int) *SkillDefault {
	return &SkillDefault{
		DefaultType: SkillID,
		Tags:        textCriteria(compare, qualifier),
		Modifier:    modifier,
	}
}

// skillNames returns the sorted names of the skills in the list, so that a selection can be compared as a whole rather
// than one probe at a time.
func skillNames(list []*Skill) []string {
	names := make([]string, 0, len(list))
	for _, sk := range list {
		names = append(names, sk.NameWithReplacements())
	}
	slices.Sort(names)
	return names
}

// TestSkillDefaultTagCriteria verifies that a skill default carrying a tag criteria resolves to the best skill bearing
// that tag -- not to the best skill overall -- and resolves to nothing at all when no skill bears it. This is the
// mechanism issue #1112 asks for: Stage Combat defaults to "any actual combat skill" at -3.
func TestSkillDefaultTagCriteria(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	broadsword := addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat")       // level 12
	brawling := addTaggedTestSkill(e, "Brawling", "", fxp.Two, "Combat")             // level 10
	cooking := addTaggedTestSkill(e, "Cooking", "", fxp.FromInteger(16), "Domestic") // level 14
	e.Recalculate()
	c.Equal(fxp.Twelve, broadsword.LevelData.Level, "precondition: Broadsword should be at level 12")
	c.Equal(fxp.Ten, brawling.LevelData.Level, "precondition: Brawling should be at level 10")
	c.Equal(fxp.FromInteger(14), cooking.LevelData.Level, "precondition: Cooking should be at level 14")

	combat := newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)
	c.Equal([]string{"Brawling", "Broadsword"}, skillNames(combat.matchingSkills(e, nil, true, nil)),
		"only the Combat-tagged skills may be selected")
	c.Equal(fxp.Nine, combat.SkillLevel(e, nil, true, nil, false),
		"the default must resolve to Broadsword-3, not to the higher-level but untagged Cooking")
	c.Equal(fxp.Nine, combat.SkillLevelFast(e, nil, true, nil, false),
		"SkillLevelFast must agree with SkillLevel")

	missing := newTaggedSkillDefault(criteria.IsText, "Sorcery", -fxp.Three)
	c.Equal(0, len(missing.matchingSkills(e, nil, true, nil)), "a tag no skill carries must select nothing")
	c.Equal(fxp.Min, missing.SkillLevel(e, nil, true, nil, false), "a tag no skill carries must not resolve")
	c.Equal(fxp.Min, missing.SkillLevelFast(e, nil, true, nil, false), "SkillLevelFast must agree with SkillLevel")
}

// TestSkillDefaultTagCriteriaList verifies the comma-list semantics a tag criteria inherits from the tag matching used
// by skill bonuses: a positive comparison matches a skill having any one of the listed tags, while a negative one
// requires every tag of the skill to fail against every entry in the list. A skill with no tags at all counts as
// having a single, empty tag, so it passes a negative comparison.
func TestSkillDefaultTagCriteriaList(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat", "Weapon") // level 12
	addTaggedTestSkill(e, "Stealth", "", fxp.Two, "Thief")                 // level 10
	addTaggedTestSkill(e, "Cooking", "", fxp.Four, "Domestic")             // level 11
	addTaggedTestSkill(e, "Carousing", "", fxp.One)                        // level 9, no tags at all
	e.Recalculate()

	either := newTaggedSkillDefault(criteria.IsText, "Combat, Thief", -fxp.Three)
	c.Equal([]string{"Broadsword", "Stealth"}, skillNames(either.matchingSkills(e, nil, true, nil)),
		"a comma-separated list must match a skill carrying any one of the listed tags")
	c.Equal(fxp.Nine, either.SkillLevel(e, nil, true, nil, false), "the best of the two is Broadsword-3")

	not := newTaggedSkillDefault(criteria.IsNotText, "Combat", -fxp.Three)
	c.Equal([]string{"Carousing", "Cooking", "Stealth"}, skillNames(not.matchingSkills(e, nil, true, nil)),
		"'is not Combat' must reject a skill tagged Combat even though it carries other tags, and must accept both a "+
			"differently-tagged and an untagged skill")
	c.Equal(fxp.Eight, not.SkillLevel(e, nil, true, nil, false), "the best of those is Cooking-3")
}

// TestSkillDefaultTagCombinesWithNameAndTL verifies that the tag criteria narrows the name and WhenTL criteria rather
// than replacing them. All three must pass for a skill to be used, and dropping any one of them lets a different,
// higher-level skill win, which shows each is doing real work.
func TestSkillDefaultTagCombinesWithNameAndTL(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	e.Profile.TechLevel = "3"
	broadsword := addTestSkill(e, "Broadsword", "", "3", fxp.Eight) // level 12; passes all three criteria
	broadsword.Tags = []string{"Combat"}
	smallsword := addTestSkill(e, "Smallsword", "", "3", fxp.Twelve) // level 13; fails the tag
	smallsword.Tags = []string{"Weapon"}
	greatsword := addTestSkill(e, "Greatsword", "", "5", fxp.FromInteger(16)) // level 14; fails the tech level
	greatsword.Tags = []string{"Combat"}
	brawling := addTestSkill(e, "Brawling", "", "3", fxp.Twelve) // level 13; fails the name
	brawling.Tags = []string{"Combat"}
	e.Recalculate()

	all := &SkillDefault{
		DefaultType: SkillID,
		Name:        textCriteria(criteria.ContainsText, "sword"),
		Tags:        textCriteria(criteria.IsText, "Combat"),
		Modifier:    -fxp.Three,
		WhenTL:      numberCriteria(criteria.AtMostNumber, fxp.Four),
	}
	c.Equal(fxp.Nine, all.SkillLevel(e, nil, true, nil, false),
		"only Broadsword satisfies the name, tag and tech level criteria together")

	withoutTag := *all
	withoutTag.Tags = criteria.Text{}
	c.Equal(fxp.Ten, withoutTag.SkillLevel(e, nil, true, nil, false),
		"without the tag criteria the higher-level, differently-tagged Smallsword wins")

	withoutTL := *all
	withoutTL.WhenTL = criteria.Number{}
	c.Equal(fxp.Eleven, withoutTL.SkillLevel(e, nil, true, nil, false),
		"without the tech level criteria the higher-level TL5 Greatsword wins")

	withoutName := *all
	withoutName.Name = criteria.Text{}
	c.Equal(fxp.Ten, withoutName.SkillLevel(e, nil, true, nil, false),
		"without the name criteria the higher-level Brawling wins")
}

// TestSkillDefaultTagsJSON verifies that the tag criteria round-trips through JSON, that it is omitted entirely when
// unset so existing files are written back byte-identical, and that a default written before the criteria objects
// existed still migrates without picking up a tag criteria.
func TestSkillDefaultTagsJSON(t *testing.T) {
	c := check.New(t)

	untagged := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	data, err := jio.Marshal(untagged)
	c.NoError(err)
	c.NotContains(string(data), "tags", "a default with no tag criteria must not write a tags key")
	var loaded SkillDefault
	c.NoError(jio.Unmarshal(data, &loaded))
	c.True(loaded.Tags.IsZero(), "a default with no tags key must load with a zero tag criteria")

	tagged := newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)
	data, err = jio.Marshal(tagged)
	c.NoError(err)
	c.Contains(string(data), `"tags"`, "a default with a tag criteria must write a tags key")
	loaded = SkillDefault{}
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(criteria.IsText, loaded.Tags.Compare, "the tag comparison must survive the round trip")
	c.Equal("Combat", loaded.Tags.Qualifier, "the tag qualifier must survive the round trip")

	// A default written before the criteria objects existed still migrates its name and carries no tag criteria.
	loaded = SkillDefault{}
	c.NoError(jio.Unmarshal([]byte(`{"type":"skill","name":"Broadsword","modifier":-3}`), &loaded))
	c.Equal(criteria.IsText, loaded.Name.Compare, "a legacy string name migrates to an 'is' criteria")
	c.Equal("Broadsword", loaded.Name.Qualifier, "a legacy string name migrates its qualifier")
	c.Equal(-fxp.Three, loaded.Modifier, "a legacy modifier still loads")
	c.True(loaded.Tags.IsZero(), "a legacy default must have no tag criteria")
}

// TestSkillDefaultTagsHashStability verifies that the tag criteria joins the source-data hash only once it is really
// set. Hashing it unconditionally would change the hash of every default in every library, marking data that has not
// been touched as modified.
func TestSkillDefaultTagsHashStability(t *testing.T) {
	c := check.New(t)

	plain := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	leftover := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	// Qualifier text left behind by a tag criteria that was set and then switched back to "is anything".
	leftover.Tags = textCriteria(criteria.AnyText, "Combat")
	tagged := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	tagged.Tags = textCriteria(criteria.IsText, "Combat")

	c.Equal(Hash64(plain), Hash64(leftover), "an 'is anything' tag criteria must not alter the hash")
	c.NotEqual(Hash64(plain), Hash64(tagged), "a real tag criteria must participate in the hash")

	// The same must hold for the skill holding the default, since that is what a library's modified marker looks at.
	e := NewEntity()
	sk := addTestSkill(e, "Stage Combat", "", "", fxp.One)
	sk.Defaults = []*SkillDefault{plain}
	before := Hash64(sk)
	sk.Defaults = []*SkillDefault{leftover}
	c.Equal(before, Hash64(sk), "an 'is anything' tag criteria must not alter the owning skill's hash")
	sk.Defaults = []*SkillDefault{tagged}
	c.NotEqual(before, Hash64(sk), "a real tag criteria must alter the owning skill's hash")
}

// TestSkillDefaultTagFillWithNameableKeys verifies that a substitution placeholder used in the tag criteria is
// collected just like the ones in the name and specialization, so the user is prompted to fill it in.
func TestSkillDefaultTagFillWithNameableKeys(t *testing.T) {
	c := check.New(t)

	def := newTaggedSkillDefault(criteria.IsText, "@tag@", -fxp.Three)
	def.Name = textCriteria(criteria.IsText, "@name@")
	def.Specialization = textCriteria(criteria.IsText, "@spec@")
	m := make(map[string]string)
	def.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"name": nameable.Unset, "spec": nameable.Unset, "tag": nameable.Unset}, m,
		"the tag criteria's placeholder must be collected alongside the name and specialization ones")
}

// TestSkillDefaultTagResolvesToBestTaggedSkill covers the scenario from issue #1112 end to end: a Stage Combat skill
// whose only default selects "any skill tagged Combat" at -3 must adopt the best Combat skill rather than the best
// skill on the sheet, and the default it records must be pinned to that skill by name, with the tag criteria spent.
func TestSkillDefaultTagResolvesToBestTaggedSkill(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	broadsword := addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat")       // level 12
	addTaggedTestSkill(e, "Brawling", "", fxp.Two, "Combat")                         // level 10
	cooking := addTaggedTestSkill(e, "Cooking", "", fxp.FromInteger(16), "Domestic") // level 14
	stage := addTestSkill(e, "Stage Combat", "", "", 0)
	stage.Defaults = []*SkillDefault{newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)}
	e.Recalculate()

	c.True(cooking.LevelData.Level > broadsword.LevelData.Level,
		"precondition: the untagged Cooking must be the highest-level skill on the sheet")
	c.NotNil(stage.DefaultedFrom, "Stage Combat must adopt its tag default")
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil),
		"the tag default must resolve to the best Combat skill, not to the higher-level Cooking")
	c.Equal(criteria.IsText, stage.DefaultedFrom.Name.Compare,
		"the recorded default must be pinned to the matched skill by name")
	c.True(stage.DefaultedFrom.Tags.IsZero(),
		"the recorded default is already pinned to one skill, so it must not keep the tag criteria")
	c.Equal(broadsword.LevelData.Level-fxp.Three, stage.LevelData.Level, "Stage Combat is Broadsword-3")
	c.True(stage.DefaultSkill() == broadsword, "the resolved default skill must be Broadsword")
}

// TestSkillDefaultTagDoesNotDefaultToItself verifies that a skill bearing the very tag its own default selects on does
// not default to itself, even when it is the best skill bearing that tag.
func TestSkillDefaultTagDoesNotDefaultToItself(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	broadsword := addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat")        // level 12
	stage := addTaggedTestSkill(e, "Stage Combat", "", fxp.FromInteger(16), "Combat") // level 14
	stage.Defaults = []*SkillDefault{newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)}
	e.Recalculate()

	c.Equal(fxp.FromInteger(14), stage.LevelData.Level,
		"precondition: Stage Combat must be the highest-level Combat skill in its own right")
	c.NotNil(stage.DefaultedFrom, "Stage Combat must still record a default")
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil),
		"the default must be the other Combat skill, not itself")
	c.True(stage.DefaultSkill() == broadsword, "the resolved default skill must be Broadsword")
	c.Equal([]string{"Broadsword"},
		skillNames(stage.Defaults[0].matchingSkills(e, nil, true, map[string]bool{stage.String(): true})),
		"the excluded skill must be kept out of the tag default's selection")
}

// TestSkillDefaultTagOffersEachTaggedSkillOnce verifies that a tag default expands into one swappable choice per
// matching skill: Swap Defaults cycles through the tagged skills in level order and wraps around, an untagged skill is
// never offered, and an explicit default naming one of those same skills at the same modifier is recognized as the
// same choice rather than being offered twice.
func TestSkillDefaultTagOffersEachTaggedSkillOnce(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat")          // level 12
	addTaggedTestSkill(e, "Judo", "", fxp.Four, "Combat")                 // level 11
	addTaggedTestSkill(e, "Brawling", "", fxp.Two, "Combat")              // level 10
	addTaggedTestSkill(e, "Cooking", "", fxp.FromInteger(16), "Domestic") // level 14
	stage := addTestSkill(e, "Stage Combat", "", "", fxp.One)
	stage.Defaults = []*SkillDefault{newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)}
	e.Recalculate()

	offered := func() []string {
		list := stage.resolvableDefaults()
		names := make([]string, 0, len(list))
		for _, def := range list {
			names = append(names, def.NameWithReplacements(nil))
		}
		return names
	}
	c.Equal([]string{"Broadsword", "Judo", "Brawling"}, offered(),
		"each Combat skill must be offered exactly once, best first, and the untagged Cooking not at all")

	c.NotNil(stage.DefaultedFrom, "Stage Combat must start out on its best default")
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil), "the best Combat skill is chosen first")
	for _, expected := range []string{"Judo", "Brawling", "Broadsword"} {
		stage.SwapToNextDefault()
		c.Equal(expected, stage.DefaultedFrom.NameWithReplacements(nil),
			"swapping defaults must move to %s", expected)
	}

	// An explicit default naming a skill the tag default already reaches, at the same modifier, is the same choice.
	stage.Defaults = append(stage.Defaults, newSkillDefaultTo("Broadsword", "", true, -fxp.Three))
	e.Recalculate()
	c.Equal([]string{"Broadsword", "Judo", "Brawling"}, offered(),
		"an explicit default to a skill the tag default already reaches must not be offered a second time")
}

// TestSkillDefaultTagEquivalent verifies that the tag criteria takes part in the equivalence check used to pair a
// recorded default with the one it came from and to dedupe the swap list, that it survives a clone, and that the
// comparison of the default type itself is made on the normalized form.
func TestSkillDefaultTagEquivalent(t *testing.T) {
	c := check.New(t)

	plain := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	tagged := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	tagged.Tags = textCriteria(criteria.IsText, "Combat")
	otherTag := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	otherTag.Tags = textCriteria(criteria.IsText, "Thief")

	c.False(plain.Equivalent(nil, tagged), "a tag criteria must make the default a different one")
	c.False(tagged.Equivalent(nil, otherTag), "defaults differing only in the tag qualifier are not equivalent")
	c.True(tagged.Equivalent(nil, tagged.CloneWithoutLevelOrPoints()), "a clone must remain equivalent")
	c.Equal(tagged.Tags, tagged.CloneWithoutLevelOrPoints().Tags, "a clone must keep the tag criteria verbatim")

	// The tag qualifier is compared after substitution, as the name and specialization qualifiers are.
	placeholder := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	placeholder.Tags = textCriteria(criteria.IsText, "@tag@")
	c.True(tagged.Equivalent(map[string]string{"tag": "Combat"}, placeholder),
		"a filled-in placeholder must be equivalent to the text it resolves to")
	c.False(tagged.Equivalent(map[string]string{"tag": "Thief"}, placeholder),
		"a placeholder filled in with different text must not be equivalent")

	// The default type is compared in its normalized form, as everything else that classifies it does.
	messy := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	messy.DefaultType = " Skill "
	c.True(plain.Equivalent(nil, messy), "types differing only in case and whitespace name the same default type")

	// A tag criteria of "is anything" selects nothing in particular however it came to be that way: a qualifier left
	// behind when the popup was switched back, or a comparison no one recognizes, changes nothing about what it
	// selects. Hash already treats all of them as one and the same, and Equivalent has to agree with it, or a
	// recorded default cannot be paired with the declared default it came from.
	leftover := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	leftover.Tags = textCriteria(criteria.AnyText, "Combat")
	unknown := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	unknown.Tags = textCriteria("any", "Combat")
	for _, other := range []*SkillDefault{leftover, unknown} {
		c.True(plain.Equivalent(nil, other), "an \"is anything\" tag criteria of %q is the same as none at all",
			other.Tags.Compare)
		c.Equal(Hash64(plain), Hash64(other), "an \"is anything\" tag criteria of %q must hash the same as none at all",
			other.Tags.Compare)
	}
	c.True(tagged.Equivalent(nil, tagged.CloneWithoutLevelOrPoints()),
		"a tag criteria that does select something must still be compared in full")
}

// TestTechniqueDefaultTypeCaseInsensitive verifies that a technique's default is classified by its normalized type, the
// way everything else classifies it. A data file not written by GCS can hold "Skill" rather than "skill"; such a
// default used to fall out of the skill branch of the level calculation and be resolved by the attribute branch's fast
// path instead, which requires the base skill to have points of its own -- so a technique based on a skill that is
// itself defaulted came out unresolvable.
func TestTechniqueDefaultTypeCaseInsensitive(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTestSkill(e, "Cooking", "", "", fxp.FromInteger(16)) // level 14
	// Broadsword has no points of its own; it holds a level only by way of its own default.
	broadsword := addTestSkill(e, "Broadsword", "", "", 0)
	broadsword.Defaults = []*SkillDefault{newSkillDefaultTo("Cooking", "", true, -fxp.Three)}
	e.Recalculate()
	c.Equal(fxp.Eleven, broadsword.LevelData.Level, "precondition: Broadsword defaults to Cooking-3")

	withType := func(defaultType string) *SkillDefault {
		def := newSkillDefaultTo("Broadsword", "", true, 0)
		def.DefaultType = defaultType
		return def
	}
	expected := CalculateTechniqueLevel(e, nil, "Stage Fighting", "", nil, withType(SkillID), difficulty.Average,
		fxp.One, false, nil, nil)
	c.Equal(broadsword.LevelData.Level+fxp.One, expected.Level,
		"precondition: the technique must resolve to its base skill's level plus its one point")
	for _, variant := range []string{"Skill", " skill ", "SKILL"} {
		result := CalculateTechniqueLevel(e, nil, "Stage Fighting", "", nil, withType(variant), difficulty.Average,
			fxp.One, false, nil, nil)
		c.Equal(expected.Level, result.Level, "%q must resolve the same as %q", variant, SkillID)
	}
}

// TestSkillDefaultTagFullName verifies that a skill-based default which names no skill outright still describes itself.
// The technique paths are the ones that show this text -- the "Requires a skill named ..." tooltip, for one -- where an
// empty description would leave the reader with a sentence that trails off.
func TestSkillDefaultTagFullName(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	named := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	c.Equal("Broadsword", named.FullName(e, nil), "a default naming a skill is described by that name")

	anySkill := &SkillDefault{DefaultType: SkillID, Modifier: -fxp.Three}
	c.Equal("any skill", anySkill.FullName(e, nil), "a default naming no skill at all must still describe itself")

	tagged := newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)
	c.Equal(`any skill where at least one tag is "Combat"`, tagged.FullName(e, nil),
		"a tag default is described by its tag criteria")

	not := newTaggedSkillDefault(criteria.IsNotText, "Combat", -fxp.Three)
	c.Equal(`any skill where all tags are not "Combat"`, not.FullName(e, nil),
		"a negative tag criteria is described in its plural form")

	// The defense suffix is still appended to whichever description was produced.
	parry := newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)
	parry.DefaultType = ParryID
	c.Equal(`any skill where at least one tag is "Combat" Parry`, parry.FullName(e, nil),
		"a parry default must still be marked as one")

	// A default that names its skill and narrows it by tag, or that selects it by any other criteria, is spelled out in
	// full: nothing it matches on may go unmentioned, and the name alone would pass it off as a plain default.
	namedAndTagged := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	namedAndTagged.Tags = textCriteria(criteria.IsText, "Combat")
	c.Equal(`any skill whose name is "Broadsword" and at least one tag is "Combat"`, namedAndTagged.FullName(e, nil),
		"a tag criteria alongside a name must be described as well")
	contains := &SkillDefault{DefaultType: SkillID, Name: textCriteria(criteria.ContainsText, "sword")}
	c.Equal(`any skill whose name contains "sword"`, contains.FullName(e, nil),
		"a name comparison other than \"is\" must be described, not shown as the bare qualifier")
	contains.Specialization = textCriteria(criteria.IsText, "Fencing")
	contains.Tags = textCriteria(criteria.IsNotText, "Cinematic")
	c.Equal(`any skill whose name contains "sword" and whose specialization is "Fencing" and all tags are not "Cinematic"`,
		contains.FullName(e, nil), "every criteria must be described, in the order the editor lists them")

	// A specialization on its own gets a subject rather than dangling in parentheses after nothing.
	specOnly := &SkillDefault{DefaultType: SkillID, Specialization: textCriteria(criteria.IsText, "Fencing")}
	c.Equal(`any skill whose specialization is "Fencing"`, specOnly.FullName(e, nil),
		"a specialization without a name must still have a subject")

	// A name that "is" nothing matches no skill, and must not be passed off as matching any.
	emptyName := &SkillDefault{DefaultType: SkillID, Name: textCriteria(criteria.IsText, "")}
	c.Equal(`any skill whose name is ""`, emptyName.FullName(e, nil), "an empty name must be described as such")

	// The ordinary forms are unchanged: the name, with the specialization in parentheses when it names one, and without
	// a qualifier left behind by a specialization criteria since switched back to "is anything".
	c.Equal("Broadsword (Fencing)", newSkillDefaultTo("Broadsword", "Fencing", false, -fxp.Three).FullName(e, nil),
		"a named specialization is shown in parentheses")
	c.Equal("Broadsword", newSkillDefaultTo("Broadsword", "", false, -fxp.Three).FullName(e, nil),
		"an empty \"is\" specialization names a skill without one, and is not shown")
	stale := newSkillDefaultTo("Broadsword", "", true, -fxp.Three)
	stale.Specialization = textCriteria(criteria.AnyText, "Fencing")
	c.Equal("Broadsword", stale.FullName(e, nil),
		"a qualifier left behind on an \"is anything\" specialization plays no part and must not be shown")
}

// TestSkillDefaultRechosenWhenDefaultsEdited verifies that editing a skill's declared defaults makes it choose its
// default afresh. A skill with points keeps the default it recorded rather than jumping to whichever currently yields
// the highest level, but that recorded choice was made from the defaults as they were before the edit: left in place,
// a Stage Combat that had recorded a low combat skill would go on using it after its default was widened to "any
// skill tagged Combat", even though a far better combat skill is now reachable. Edits that leave the defaults alone
// must still honor the recorded choice.
func TestSkillDefaultRechosenWhenDefaultsEdited(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat") // level 12
	addTaggedTestSkill(e, "Knife", "", fxp.One, "Combat")        // level 9
	stage := addTestSkill(e, "Stage Combat", "", "", fxp.Two)
	stage.Defaults = []*SkillDefault{
		newSkillDefaultTo("Broadsword", "", true, -fxp.Three),
		newSkillDefaultTo("Knife", "", true, -fxp.Three),
	}
	e.Recalculate()
	c.NotNil(stage.DefaultedFrom, "precondition: Stage Combat must resolve a default")
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil), "precondition: the best default is chosen first")
	stage.SwapToNextDefault()
	c.Equal("Knife", stage.DefaultedFrom.NameWithReplacements(nil), "precondition: the user has swapped to Knife")

	// An edit that leaves the defaults alone keeps the user's choice, even though Broadsword would yield more.
	var edit SkillEditData
	edit.CopyFrom(stage)
	edit.Points = fxp.Four
	edit.ApplyTo(stage)
	e.Recalculate()
	c.Equal(fxp.Four, stage.Points, "precondition: the points edit was applied")
	c.NotNil(stage.DefaultedFrom, "Stage Combat must still resolve a default")
	c.Equal("Knife", stage.DefaultedFrom.NameWithReplacements(nil),
		"an edit that does not touch the defaults must keep the recorded choice")

	// Replacing the defaults with a tag default that still reaches Knife must not keep Knife on the strength of that:
	// the choice is made afresh, and the best Combat skill wins. The editor takes a snapshot before the edit to undo it
	// with, so one is taken here too.
	var before SkillEditData
	before.CopyFrom(stage)
	edit = SkillEditData{}
	edit.CopyFrom(stage)
	edit.Defaults = []*SkillDefault{newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)}
	edit.ApplyTo(stage)
	e.Recalculate()
	c.NotNil(stage.DefaultedFrom, "Stage Combat must resolve a default from the edited defaults")
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil),
		"editing the defaults must re-choose the default rather than keep the one recorded before the edit")

	// Undoing the edit applies the snapshot taken before it. The default that snapshot recorded was chosen from the
	// snapshot's own defaults, so it is exactly the one to bring back: the user's earlier swap to Knife returns along
	// with the defaults it was made from, rather than being thrown away for the best default a second time.
	before.ApplyTo(stage)
	e.Recalculate()
	c.Equal(2, len(stage.Defaults), "precondition: the undo restored the declared defaults")
	c.NotNil(stage.DefaultedFrom, "Stage Combat must resolve a default from the restored defaults")
	c.Equal("Knife", stage.DefaultedFrom.NameWithReplacements(nil),
		"undoing a defaults edit must bring back the default recorded along with those defaults")

	// Redoing it applies the edit again, which is judged the same way it was the first time.
	edit.ApplyTo(stage)
	e.Recalculate()
	c.Equal("Broadsword", stage.DefaultedFrom.NameWithReplacements(nil),
		"redoing a defaults edit must re-choose the default just as the edit did")
}

// TestTechniqueDefaultCriteriaClearedWhenNotSkillBased verifies that a technique whose default is an attribute rather
// than a skill carries no skill criteria on that default, however it came by them. The editor hides the criteria
// fields for such a default, so a name, specialization or tag left behind from an earlier skill selection could not be
// seen or removed, yet would still be written to disk and hashed, making the technique look modified relative to its
// library source over criteria that nothing consults.
func TestTechniqueDefaultCriteriaClearedWhenNotSkillBased(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	newFeint := func(owner DataOwner) *Skill {
		tech := NewTechnique(owner, nil, "Broadsword")
		tech.Name = "Feint"
		tech.PageRef = "B231"
		tech.TechniqueDefault.DefaultType = DexterityID
		return tech
	}
	withLeftovers := func(tech *Skill) *Skill {
		tech.TechniqueDefault.Specialization = textCriteria(criteria.IsText, "Fencing")
		tech.TechniqueDefault.Tags = textCriteria(criteria.IsText, "Combat")
		return tech
	}
	clean := newFeint(e)
	clean.TechniqueDefault.Name = criteria.Text{}
	stale := withLeftovers(newFeint(e))
	c.NotEqual(Hash64(clean), Hash64(stale), "precondition: the leftover criteria count toward the hash")

	// Committing an edit of the technique clears them.
	var edit SkillEditData
	edit.CopyFrom(stale)
	edit.ApplyTo(stale)
	c.True(stale.TechniqueDefault.Name.IsZero(), "an edit must clear a leftover name criteria")
	c.True(stale.TechniqueDefault.Specialization.IsZero(), "an edit must clear a leftover specialization criteria")
	c.True(stale.TechniqueDefault.Tags.IsZero(), "an edit must clear a leftover tag criteria")
	c.Equal(Hash64(clean), Hash64(stale), "once cleared, the technique must hash as one that never had them")

	// So does syncing the technique with a library source that still carries them.
	source := withLeftovers(newFeint(nil))
	local := newFeint(e)
	local.TechniqueDefault.Name = criteria.Text{}
	local.PageRef = "B232"
	e.Skills = append(e.Skills, local)
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + SkillsExt}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: the technique differs from its source")
	local.SyncWithSource()
	c.Equal("B231", local.PageRef, "the sync brings the source's page reference across")
	c.True(local.TechniqueDefault.Name.IsZero(), "a sync must not bring a leftover name criteria across")
	c.True(local.TechniqueDefault.Specialization.IsZero(),
		"a sync must not bring a leftover specialization criteria across")
	c.True(local.TechniqueDefault.Tags.IsZero(), "a sync must not bring a leftover tag criteria across")
}

// TestTechniqueSatisfiedTooltipForTagDefault verifies that the tooltip for an unsatisfied technique is phrased around
// the description of a default that names no skill outright, rather than presenting that description as though it
// were a skill's name.
func TestTechniqueSatisfiedTooltipForTagDefault(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	tech := NewTechnique(e, nil, "")
	tech.Name = "Feint"
	tech.TechniqueDefault = newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Two)
	e.Skills = append(e.Skills, tech)
	var tooltip xbytes.InsertBuffer
	c.False(tech.TechniqueSatisfied(&tooltip, "- "), "precondition: there is no Combat skill to satisfy the technique")
	c.Equal(`- Requires any skill where at least one tag is "Combat"`, tooltip.String(),
		"a missing skill selected by tag must be asked for by its description")

	// With a Combat skill that has no points in it, it is the points that are missing.
	addTaggedTestSkill(e, "Knife", "", 0, "Combat")
	tooltip = xbytes.InsertBuffer{}
	c.False(tech.TechniqueSatisfied(&tooltip, "- "), "precondition: a Combat skill without points does not satisfy it")
	c.Equal(`- Requires at least 1 point in any skill where at least one tag is "Combat"`, tooltip.String(),
		"points missing from a skill selected by tag must be asked for by its description")

	// A default that names its skill keeps the plainer wording.
	named := NewTechnique(e, nil, "Broadsword")
	e.Skills = append(e.Skills, named)
	tooltip = xbytes.InsertBuffer{}
	c.False(named.TechniqueSatisfied(&tooltip, "- "), "precondition: there is no Broadsword skill")
	c.Equal("- Requires a skill named Broadsword", tooltip.String(), "a missing named skill is asked for by name")
}

// TestSkillDefaultRechosenWhenSyncedWithSource verifies that a sync which brings changed defaults across from the
// library re-chooses the default the same way an edit does, while a sync that leaves the defaults alone keeps the
// recorded choice.
func TestSkillDefaultRechosenWhenSyncedWithSource(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	addTaggedTestSkill(e, "Broadsword", "", fxp.Eight, "Combat") // level 12
	addTaggedTestSkill(e, "Knife", "", fxp.One, "Combat")        // level 9
	newStage := func(owner DataOwner, pageRef string, defaults ...*SkillDefault) *Skill {
		sk := NewSkill(owner, nil, false)
		sk.Name = "Stage Combat"
		sk.Difficulty.Attribute = IntelligenceID
		sk.Difficulty.Difficulty = difficulty.Average
		sk.PageRef = pageRef
		sk.Defaults = defaults
		return sk
	}
	source := newStage(nil, "B222", newSkillDefaultTo("Knife", "", true, -fxp.Three))
	local := newStage(e, "B222", newSkillDefaultTo("Knife", "", true, -fxp.Three))
	local.Points = fxp.Two
	e.Skills = append(e.Skills, local)
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + SkillsExt}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	setSource := func() {
		e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
			libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
		}
	}
	setSource()
	e.Recalculate()
	c.NotNil(local.DefaultedFrom, "precondition: Stage Combat must resolve a default")
	c.Equal("Knife", local.DefaultedFrom.NameWithReplacements(nil), "precondition: Knife is the only default")

	// A sync that changes something other than the defaults keeps the recorded choice.
	source.PageRef = "B223"
	setSource()
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: a differing page reference is a mismatch")
	local.SyncWithSource()
	e.Recalculate()
	c.Equal("B223", local.PageRef, "the sync brings the source's page reference across")
	c.NotNil(local.DefaultedFrom, "Stage Combat must still resolve a default")
	c.Equal("Knife", local.DefaultedFrom.NameWithReplacements(nil),
		"a sync that leaves the defaults alone must keep the recorded choice")

	// A sync that widens the defaults to a tag default re-chooses, and the best Combat skill wins.
	source.Defaults = []*SkillDefault{newTaggedSkillDefault(criteria.IsText, "Combat", -fxp.Three)}
	setSource()
	state, _ = e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: differing defaults are a mismatch")
	local.SyncWithSource()
	e.Recalculate()
	c.NotNil(local.DefaultedFrom, "Stage Combat must resolve a default from the synced defaults")
	c.Equal("Broadsword", local.DefaultedFrom.NameWithReplacements(nil),
		"syncing changed defaults must re-choose the default rather than keep the one recorded before the sync")
}
