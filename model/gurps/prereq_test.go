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

// TestPrereqDescribe verifies the description of each prerequisite type, including which names and qualifiers are
// passed through the emphasis func and which criteria are left out.
func TestPrereqDescribe(t *testing.T) {
	c := check.New(t)
	trait := func(name string, compare criteria.NumericComparison, level fxp.Int) *gurps.TraitPrereq {
		p := gurps.NewTraitPrereq()
		p.NameCriteria.Qualifier = name
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
	quantity := gurps.NewContainedQuantityPrereq()
	quantity.Has = false

	for _, one := range []struct {
		name     string
		prereq   gurps.Prereq
		expected string
	}{
		{"trait at least 0", trait("Magery", criteria.AtLeastNumber, 0), "Has trait [Magery]"},
		{"trait with level", trait("Magery", criteria.AtLeastNumber, fxp.One), "Has trait [Magery] at level at least 1"},
		{"trait exact level", trait("Magery", criteria.EqualsNumber, fxp.Two), "Has trait [Magery] at level 2"},
		{"trait any level", trait("Magery", criteria.AnyNumber, 0), "Has trait [Magery]"},
		{
			"trait with notes", notes,
			`Does not have trait whose name contains "[Mag]" at level at least 2 whose notes contains "[Fire]"`,
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
		{"script", gurps.NewScriptPrereq(), "Passes a custom check"},
		{
			"unknown", gurps.NewUnknownPrereq("future", []byte(`{"type":"future"}`)),
			`Meets an unknown type of prerequisite ("[future]") that needs a newer version of GCS`,
		},
	} {
		c.Equal(one.expected, one.prereq.Describe(nil, func(s string) string { return "[" + s + "]" }), one.name)
	}
}

// TestPrereqDescribeReplacements verifies that a description applies the nameable replacements it is given.
func TestPrereqDescribeReplacements(t *testing.T) {
	c := check.New(t)
	p := gurps.NewTraitPrereq()
	p.NameCriteria.Qualifier = "@Trait@"
	c.Equal("Has trait Magery", p.Describe(map[string]string{"Trait": "Magery"}, func(s string) string { return s }))
}

// TestPrereqListDescribe verifies that a list joins its children to match its mode, parenthesizes nested lists of more
// than one child, and notes a tech level condition.
func TestPrereqListDescribe(t *testing.T) {
	c := check.New(t)
	newTrait := func(name string) *gurps.TraitPrereq {
		p := gurps.NewTraitPrereq()
		p.NameCriteria.Qualifier = name
		return p
	}
	root := gurps.NewPrereqList()
	root.Prereqs = append(root.Prereqs, newTrait("Magery"))
	anyOf := gurps.NewPrereqList()
	anyOf.All = false
	anyOf.WhenTL.Compare = criteria.AtLeastNumber
	anyOf.WhenTL.Qualifier = fxp.Three
	anyOf.Prereqs = append(anyOf.Prereqs, newTrait("Luck"), newTrait("DX"))
	single := gurps.NewPrereqList()
	single.Prereqs = append(single.Prereqs, newTrait("Fearlessness"))
	root.Prereqs = append(root.Prereqs, anyOf, single)
	c.Equal("Has trait Magery and (has trait Luck or has trait DX) (only when TL at least 3) and has trait Fearlessness",
		root.Describe(nil, func(s string) string { return s }))
	c.Equal("", gurps.NewPrereqList().Describe(nil, func(s string) string { return s }))
}
