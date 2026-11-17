// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// namedTrait returns a trait prerequisite for the name.
func namedTrait(name string) *gurps.TraitPrereq {
	p := gurps.NewTraitPrereq()
	p.NameCriteria.Qualifier = name
	return p
}

// TestPrereqDescribe verifies the description of each prerequisite type, including which names and qualifiers are
// passed through the emphasis func and which criteria are left out.
func TestPrereqDescribe(t *testing.T) {
	c := check.New(t)
	trait := func(name string, compare criteria.NumericComparison, level fxp.Int) *gurps.TraitPrereq {
		p := namedTrait(name)
		p.LevelCriteria.Compare = compare
		p.LevelCriteria.Qualifier = level
		return p
	}
	skill := func(name, specialization string, specCompare criteria.StringComparison) *gurps.SkillPrereq {
		p := gurps.NewSkillPrereq()
		p.NameCriteria.Qualifier = name
		p.SpecializationCriteria.Compare = specCompare
		p.SpecializationCriteria.Qualifier = specialization
		return p
	}
	spell := func(subType spellcmp.Type, qualifier string, quantity fxp.Int) *gurps.SpellPrereq {
		p := gurps.NewSpellPrereq()
		p.SubType = subType
		p.QualifierCriteria.Qualifier = qualifier
		p.QuantityCriteria.Qualifier = quantity
		p.SamePowerSource = false
		return p
	}
	attribute := func(which, combined string) *gurps.AttributePrereq {
		p := gurps.NewAttributePrereq(nil)
		p.Which = which
		p.CombinedWith = combined
		p.QualifierCriteria.Qualifier = fxp.FromInteger(20)
		return p
	}
	equipped := func(name string, nameCompare criteria.StringComparison, tag string,
		tagCompare criteria.StringComparison,
	) *gurps.EquippedEquipmentPrereq {
		p := gurps.NewEquippedEquipmentPrereq()
		p.NameCriteria.Compare = nameCompare
		p.NameCriteria.Qualifier = name
		p.TagsCriteria.Compare = tagCompare
		p.TagsCriteria.Qualifier = tag
		return p
	}

	notes := trait("Mag", criteria.AtLeastNumber, fxp.Two)
	notes.Has = false
	notes.NameCriteria.Compare = criteria.ContainsText
	notes.NotesCriteria.Compare = criteria.ContainsText
	notes.NotesCriteria.Qualifier = "Fire"
	optional := skill("Broadsword", "", criteria.AnyText)
	optional.OptionalSpecializationCriteria.Compare = criteria.IsText
	optional.OptionalSpecializationCriteria.Qualifier = "Rapier"
	optional.LevelCriteria.Qualifier = fxp.Twelve
	sameSource := spell(spellcmp.Name, "Fireball", fxp.One)
	sameSource.SamePowerSource = true
	explicitSource := spell(spellcmp.Any, "", fxp.Three)
	explicitSource.PowerSourceCriteria.Compare = criteria.IsText
	explicitSource.PowerSourceCriteria.Qualifier = "Clerical"
	unknownSpell := spell(spellcmp.Tag, "Fire", fxp.One)
	unknownSpell.Has = false
	noTitleNote := spell(spellcmp.TitleNote, "", fxp.One)
	noTitleNote.QualifierCriteria.Compare = criteria.IsText
	noTags := spell(spellcmp.Tag, "", fxp.One)
	noTags.QualifierCriteria.Compare = criteria.IsText
	anyCollege := spell(spellcmp.College, "", fxp.One)
	anyCollege.QualifierCriteria.Compare = criteria.IsNotText
	eitherTag := spell(spellcmp.Tag, "Necromantic, Holy", fxp.One)
	eitherTag.QualifierCriteria.Compare = criteria.IsText
	quantity := gurps.NewContainedQuantityPrereq()
	quantity.Has = false
	pointsAndTags := trait("Magery", criteria.AtLeastNumber, 0)
	pointsAndTags.PointsCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Ten}
	pointsAndTags.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Mental"}
	withModifier := trait("Magery", criteria.AtLeastNumber, 0)
	withModifier.ModifierNameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "One College"}
	withModifier.ModifierNotesCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "Fire"}
	withModifier.ModifierTagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Limitation"}
	skillPoints := skill("Broadsword", "", criteria.AnyText)
	skillPoints.PointsCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Four}
	skillPoints.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Combat"}
	notEquipped := equipped("Shield", criteria.IsText, "", criteria.AnyText)
	notEquipped.Not = true
	anyName := trait("", criteria.AtLeastNumber, 0)
	anyName.NameCriteria.Compare = criteria.AnyText

	for _, one := range []struct {
		name     string
		prereq   gurps.Prereq
		expected string
	}{
		{"trait at least 0", trait("Magery", criteria.AtLeastNumber, 0), "Has trait [Magery]"},
		{"trait with level", trait("Magery", criteria.AtLeastNumber, fxp.One), "Has trait [Magery] at level at least 1"},
		{"trait exact level", trait("Magery", criteria.EqualsNumber, fxp.Two), "Has trait [Magery] at level 2"},
		{"trait any level", trait("Magery", criteria.AnyNumber, 0), "Has trait [Magery]"},
		{"trait of any name", anyName, "Has trait of any name"},
		{
			"trait with notes", notes,
			`Does not have trait whose name contains "[Mag]" at level at least 2 whose notes contains "[Fire]"`,
		},
		{
			"trait with points and tags", pointsAndTags,
			"Has trait [Magery] whose point total is at least 10 tagged [Mental]",
		},
		{
			"trait with a modifier", withModifier,
			`Has trait [Magery] with a modifier whose name is [One College] and whose notes contains "[Fire]" and ` +
				"tagged [Limitation]",
		},
		{
			"skill with points and tags", skillPoints,
			"Has skill [Broadsword] at level at least 0 whose point total is at least 4 tagged [Combat]",
		},
		{"skill", skill("Thaumatology", "", criteria.AnyText), "Has skill [Thaumatology] at level at least 0"},
		{
			"skill with specialization", skill("Guns", "Pistol", criteria.IsText),
			"Has skill [Guns] ([Pistol]) at level at least 0",
		},
		{
			"skill with specialization that contains", skill("Guns", "Pis", criteria.ContainsText),
			`Has skill [Guns] with a specialization that contains "[Pis]" at level at least 0`,
		},
		{
			"skill with optional specialization", optional,
			"Has skill [Broadsword] with an optional specialization that is [Rapier] at level at least 12",
		},
		{
			"spell college", spell(spellcmp.College, "Fire", fxp.Two),
			"Knows at least 2 spells whose college is [Fire]",
		},
		{"spell college count", spell(spellcmp.CollegeCount, "", fxp.One), "Knows spells from at least 1 college"},
		{
			"spell same power source", sameSource,
			"Knows at least 1 spell whose name is [Fireball] and whose power source is the same as this spell's",
		},
		{"spell explicit power source", explicitSource, "Knows at least 3 spells whose power source is [Clerical]"},
		{"spell not known", unknownSpell, "Does not know at least 1 spell whose tag is [Fire]"},
		// An empty "is" on a title note, tag or college means the spell has none, and an empty "is not" at least
		// one, as the matches count them; a tag or college qualifier lists the alternatives it matches.
		{"spell without title notes", noTitleNote, "Knows at least 1 spell without title notes"},
		{"spell without tags", noTags, "Knows at least 1 spell without tags"},
		{"spell with a college", anyCollege, "Knows at least 1 spell with at least one college"},
		{"spell with either tag", eitherTag, `Knows at least 1 spell whose tag is "[Necromantic]" or "[Holy]"`},
		{"attribute", attribute(gurps.StrengthID, ""), "Has [ST] at least 20"},
		{"attribute combined", attribute(gurps.StrengthID, gurps.DexterityID), "Has [ST+DX] at least 20"},
		{
			"attribute beyond the sheet's", attribute(gurps.SizeModifierID, gurps.ParryID),
			"Has [Size Modifier+Parry] at least 20",
		},
		{"attribute unrecognized", attribute("xyz", ""), "Has [xyz] at least 20"},
		{
			"equipped by name and tag", equipped("Sword", criteria.IsText, "Weapon", criteria.IsText),
			"Has [Sword] equipped tagged [Weapon]",
		},
		{"equipped anything", equipped("", criteria.AnyText, "", criteria.AnyText), "Has any equipment equipped"},
		{"not equipped", notEquipped, "Does not have [Shield] equipped"},
		{
			"equipped without a tag", equipped("", criteria.AnyText, "Cursed", criteria.DoesNotContainText),
			`Has equipped equipment with all tags that do not contain "[Cursed]"`,
		},
		{
			"equipped name contains", equipped("Sw", criteria.ContainsText, "", criteria.AnyText),
			`Has equipped equipment whose name contains "[Sw]"`,
		},
		{"contained quantity", quantity, "Does not have a contained quantity of at most 1"},
		{"contained weight", gurps.NewContainedWeightPrereq(nil), "Has a contained weight of at most 5 lb"},
		{"script", gurps.NewScriptPrereq(), "A custom check"},
		{
			"unknown", gurps.NewUnknownPrereq("future", []byte(`{"type":"future"}`)),
			`Meets an unknown type of prerequisite ("[future]") that needs a newer version of GCS`,
		},
	} {
		c.Equal(one.expected, one.prereq.Describe(nil, nil, func(s string) string { return "[" + s + "]" }), one.name)
	}
}

// TestPrereqDescribeQuotesValues checks how a "whose ..." clause shows its value: bare after "is" when it reads
// plainly, without the space at either end that "is" ignores, and otherwise quoted as it is, in single quotes when it
// holds a double quote. The quotes sit outside what em emphasizes.
func TestPrereqDescribeQuotesValues(t *testing.T) {
	c := check.New(t)
	plain := func(s string) string { return s }
	for i, one := range []struct {
		compare criteria.StringComparison
		notes   string
		em      func(string) string
		want    string
	}{
		{criteria.IsText, "Sharp", plain, "Has trait Acute Vision whose notes is Sharp"},
		{criteria.IsText, "Sharp, Keen", plain, `Has trait Acute Vision whose notes is "Sharp, Keen"`},
		{criteria.IsText, " Sharp", plain, "Has trait Acute Vision whose notes is Sharp"},
		{criteria.StartsWithText, " Sharp", plain, `Has trait Acute Vision whose notes starts with " Sharp"`},
		{criteria.ContainsText, " ", plain, `Has trait Acute Vision whose notes contains " "`},
		{criteria.ContainsText, "", plain, `Has trait Acute Vision whose notes contains ""`},
		{criteria.IsText, `Say "hi"`, plain, `Has trait Acute Vision whose notes is 'Say "hi"'`},
		{criteria.IsText, "Sharp", bracket, "Has trait [Acute Vision] whose notes is [Sharp]"},
		{criteria.IsText, "Sharp, Keen", bracket, `Has trait [Acute Vision] whose notes is "[Sharp, Keen]"`},
		{criteria.IsText, `Say "hi"`, bracket, `Has trait [Acute Vision] whose notes is '[Say "hi"]'`},
	} {
		p := namedTrait("Acute Vision")
		p.NotesCriteria = criteria.Text{Compare: one.compare, Qualifier: one.notes}
		c.Equal(one.want, p.Describe(nil, nil, one.em), "case %d", i)
	}
}

// TestPrereqDescribeEmptyNames verifies that an "is" name that is empty, or becomes empty once its markers are
// replaced, reads as "", and that such a specialization, or optional specialization, says it picks only a skill without
// one.
func TestPrereqDescribeEmptyNames(t *testing.T) {
	c := check.New(t)
	plain := func(s string) string { return s }
	replacements := map[string]string{"Blank": ""}
	trait := gurps.NewTraitPrereq()
	trait.NameCriteria.Qualifier = ""
	c.Equal(`Has trait ""`, trait.Describe(nil, nil, plain))
	trait.NameCriteria.Qualifier = "@Blank@"
	c.Equal(`Has trait ""`, trait.Describe(nil, replacements, plain))
	skill := gurps.NewSkillPrereq()
	skill.NameCriteria.Qualifier = "Guns"
	skill.SpecializationCriteria = criteria.Text{Compare: criteria.IsText}
	c.Equal("Has skill Guns without a specialization at level at least 0", skill.Describe(nil, nil, plain))
	skill.SpecializationCriteria.Qualifier = "@Blank@"
	c.Equal("Has skill Guns without a specialization at level at least 0", skill.Describe(nil, replacements, plain))
	skill.SpecializationCriteria.Qualifier = " "
	c.Equal("Has skill Guns without a specialization at level at least 0", skill.Describe(nil, nil, plain))
	skill.SpecializationCriteria = criteria.Text{}
	skill.OptionalSpecializationCriteria = criteria.Text{Compare: criteria.IsText}
	c.Equal("Has skill Guns without an optional specialization at level at least 0", skill.Describe(nil, nil, plain))
	bonus := gurps.NewSkillBonus()
	bonus.NameCriteria.Qualifier = "Guns"
	bonus.SpecializationCriteria = criteria.Text{Compare: criteria.IsText}
	bonus.Amount = fxp.One
	c.Equal("+1 to skill Guns without a specialization", bonus.Describe(nil, nil, plain))
	blankBonus := gurps.NewSkillBonus()
	blankBonus.NameCriteria.Qualifier = "@Blank@"
	blankBonus.Amount = fxp.One
	c.Equal(criteria.IsText, blankBonus.NameCriteria.Compare)
	c.Equal(`+1 to skill ""`, blankBonus.Describe(nil, replacements, plain), "the bonus names it as the prerequisite does")
	blankSkill := gurps.NewSkillPrereq()
	blankSkill.NameCriteria.Qualifier = "@Blank@"
	c.Equal(criteria.IsText, blankSkill.NameCriteria.Compare)
	c.Equal(`Has skill "" at level at least 0`, blankSkill.Describe(nil, replacements, plain))
}
