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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// linkedTrait returns a trait whose own text uses @style@, with a modifier whose @focus@ is linked to it and whose
// @extra@ is not.
func linkedTrait() (*Trait, *TraitModifier) {
	trait := NewTrait(nil, nil, false)
	trait.Name = "Trained in @style@"
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@focus@ focus"
	mod.LocalNotes = "with @extra@"
	trait.Modifiers = []*TraitModifier{mod}
	mod.SetTarget(trait)
	trait.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"focus": "style"}}
	return trait, mod
}

// TestLinkedModifierCodeIsAskedAsOwnerCode verifies that a modifier's linked code is asked for, answered and kept as
// its owner's code, and that its linked and unlinked codes both resolve through the owner's replacements.
func TestLinkedModifierCodeIsAskedAsOwnerCode(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo"}
	m := make(map[string]string)
	trait.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"style": "Judo", "extra": nameable.Unset}, m, "the linked code is asked as the trait's")

	// An answer given for the trait's code is the one the linked code shows.
	m["style"] = "Karate"
	m["extra"] = "zeal"
	trait.ApplyNameableKeys(m)
	c.Equal("Karate focus", mod.NameWithReplacements())
	c.Equal("with zeal", mod.LocalNotesWithReplacements(), "an unlinked code resolves through the owner too")
	c.Equal(map[string]string{"style": "Karate", "extra": "zeal"}, trait.Replacements,
		"nothing is kept under the modifier's own code")
}

func TestCloneRekeysNameableLinks(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo"}
	clone := trait.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 {
		return
	}
	cloned := clone.Modifiers[0]
	c.NotEqual(mod.TID, cloned.TID, "a reference copy gets a fresh ID")
	c.Equal(map[tid.TID]map[string]string{cloned.TID: {"focus": "style"}}, clone.NameableLinks)
	c.Equal("Judo focus", cloned.NameWithReplacements())

	clone.NameableLinks[cloned.TID]["focus"] = "other"
	c.Equal("style", trait.NameableLinks[mod.TID]["focus"], "the original's links are not shared")
}

func TestLinkedEquipmentModifierCode(t *testing.T) {
	c := check.New(t)
	eqp := NewEquipment(nil, nil, false)
	eqp.Name = "Sword of @metal@"
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Name = "@material@ inlay"
	eqp.SetModifiers([]*EquipmentModifier{mod})
	eqp.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"material": "metal"}}
	eqp.Replacements = map[string]string{"metal": "Silver"}
	c.Equal("Silver inlay", mod.NameWithReplacements())
	m := make(map[string]string)
	eqp.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"metal": "Silver"}, m, "the linked code is asked as the equipment's")
}

func TestLinkedModifierCodeReachesItsFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skill := NewSkill(e, nil, false)
	skill.Name = "Broadsword"
	skill.Points = fxp.Four
	e.Skills = append(e.Skills, skill)

	// A trait modifier whose skill bonus names its skill through a code linked to the trait's.
	traitMod := NewTraitModifier(e, nil, false)
	traitMod.Name = "Focused on @skill@"
	traitMod.Features = Features{newSkillBonusTo("@skill@", fxp.Two)}
	trait := addTraitWithFeatures(e, "Weapon Master (@weapon@)")
	trait.SetModifiers([]*TraitModifier{traitMod})
	trait.NameableLinks = map[tid.TID]map[string]string{traitMod.TID: {"skill": "weapon"}}
	trait.Replacements = map[string]string{"weapon": "Broadsword"}

	// Likewise for an equipment modifier.
	eqpMod := NewEquipmentModifier(e, nil, false)
	eqpMod.Name = "Balanced for @use@"
	eqpMod.Features = Features{newSkillBonusTo("@use@", fxp.One)}
	eqp := addCarriedEquipmentWithFeatures(e, "Sword of @skill@")
	eqp.SetModifiers([]*EquipmentModifier{eqpMod})
	eqp.NameableLinks = map[tid.TID]map[string]string{eqpMod.TID: {"use": "skill"}}
	eqp.Replacements = map[string]string{"skill": "Broadsword"}
	eqp.Equipped = true

	e.Recalculate()
	c.Equal("Focused on Broadsword", traitMod.NameWithReplacements())
	c.Equal("Balanced for Broadsword", eqpMod.NameWithReplacements())
	c.Equal(fxp.Three, e.SkillBonusFor("Broadsword", "", "", nil, nil),
		"a modifier's bonus resolves its linked code through its owner, as its name does")
}

func TestApplyDropsSelfLinks(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.NameableLinks[mod.TID]["extra"] = "style"
	var edit TraitEditData
	edit.CopyFrom(trait)
	// Renaming the trait's code to the modifier's in the editor leaves the link pointing at itself.
	edit.NameableLinks[mod.TID]["focus"] = "focus"
	edit.ApplyTo(trait)
	c.Equal(map[tid.TID]map[string]string{mod.TID: {"extra": "style"}}, trait.NameableLinks,
		"a link of a code to itself means the same as none, so it is not kept")

	edit.CopyFrom(trait)
	edit.NameableLinks[mod.TID]["extra"] = "extra"
	edit.ApplyTo(trait)
	c.Nil(trait.NameableLinks, "a modifier left with no links has no entry")
}

func TestNameablesTakePartInSync(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Trained in @style@"
	before := Hash64(trait)
	trait.Nameables = map[string]nameable.Detail{}
	c.Equal(before, Hash64(trait), "an item without nameables hashes as it did before they existed")
	trait.Nameables["style"] = nameable.Detail{Label: "Style", Options: []string{"Judo"}}
	c.NotEqual(before, Hash64(trait), "a change to its nameables puts a copy out of sync with its source")
}
