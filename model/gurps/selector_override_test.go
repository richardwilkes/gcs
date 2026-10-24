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
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// newDamageTypeOverride returns an override setting the weapon damage type to the value, selecting weapons as given.
func newDamageTypeOverride(selection wsel.Type, value string) *SelectorOverride {
	o := NewSelectorOverride(selector.WeaponDamageType)
	o.SelectionType = selection
	o.Value = value
	return o
}

// newCrushingWeapon returns a melee weapon of the owner with the given usage that does crushing damage. It is not
// added to the owner.
func newCrushingWeapon(owner WeaponOwner, usage string) *Weapon {
	w := NewWeapon(owner, true)
	w.Usage = usage
	return w
}

// resolvedDamageType returns the weapon's damage type after any overrides.
func resolvedDamageType(w *Weapon, tooltip *xbytes.InsertBuffer) string {
	return w.ResolveSelector(selector.WeaponDamageType, w.Damage.Type, tooltip)
}

// addAverageDXSkill adds a DX/A skill with the given points to the entity's skills and returns it. The entity is not
// recalculated.
func addAverageDXSkill(e *Entity, name string, points fxp.Int) *Skill {
	sk := NewSkill(e, nil, false)
	sk.Name = name
	sk.Difficulty.Attribute = DexterityID
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Points = points
	e.Skills = append(e.Skills, sk)
	return sk
}

// TestSelectorOverrideThisWeapon verifies that an override selecting this weapon applies to every weapon of the item
// carrying it, and to nothing else, narrowed by its usage criterion.
func TestSelectorOverrideThisWeapon(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	override := newDamageTypeOverride(wsel.ThisWeapon, "cut")
	gadget := addTraitWithFeatures(e, "Gadget", override)
	swung := newCrushingWeapon(gadget, "Swung")
	thrust := newCrushingWeapon(gadget, "Thrust")
	gadget.Weapons = []*Weapon{swung, thrust}
	other := addTraitWithFeatures(e, "Other")
	otherWeapon := newCrushingWeapon(other, "Swung")
	other.Weapons = []*Weapon{otherWeapon}
	e.Recalculate()
	c.Equal("cut", resolvedDamageType(swung, nil), "an override selecting this weapon applies to its item's weapon")
	c.Equal("cut", resolvedDamageType(thrust, nil), "and to every weapon of the item")
	c.Equal("cr", resolvedDamageType(otherWeapon, nil), "but not to another item's weapon")
	c.Contains(swung.Damage.ResolvedDamage(nil), "cut", "the resolved damage shows it")
	override.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Swung"}
	c.Equal("cut", resolvedDamageType(swung, nil), "a usage criterion keeps the weapons it matches")
	c.Equal("cr", resolvedDamageType(thrust, nil), "and excludes the rest")
}

// TestSelectorOverrideThisWeaponFromModifier verifies that an override selecting this weapon carried by a trait
// modifier applies to the trait's weapon, follows the trait's switch, and is attributed to the modifier.
func TestSelectorOverrideThisWeaponFromModifier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	override := newDamageTypeOverride(wsel.ThisWeapon, "cut")
	override.Switchable = true
	gadget := addTraitWithFeatures(e, "Gadget")
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Scoped"
	mod.Features = Features{override}
	gadget.Modifiers = []*TraitModifier{mod}
	w := newCrushingWeapon(gadget, "Swung")
	gadget.Weapons = []*Weapon{w}
	e.Recalculate()
	c.Equal("cr", resolvedDamageType(w, nil),
		"a modifier's switchable override is ignored while the trait's switch is off")
	gadget.SetSwitchedOn(true)
	e.Recalculate()
	var tooltip xbytes.InsertBuffer
	c.Equal("cut", resolvedDamageType(w, &tooltip), "and applies while it is on")
	c.Contains(tooltip.String(), "Gadget (Scoped)", "the tooltip names the trait and its modifier")
}

// TestSelectorOverrideOnUnequippedEquipment verifies that overrides aimed at the weapons of equipment that isn't
// equipped, which never reach the entity's feature lists, still apply to those weapons and name the equipment, and
// that a switchable one keeps the equipment's switch meaningful.
func TestSelectorOverrideOnUnequippedEquipment(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	this := newDamageTypeOverride(wsel.ThisWeapon, "cut")
	this.Switchable = true
	named := NewSelectorOverride(selector.WeaponArmorDivisor)
	named.SelectionType = wsel.WithName
	named.Value = "2"
	named.NameCriteria.Qualifier = "Amulet"
	eqp := addCarriedEquipmentWithFeatures(e, "Amulet", this, named)
	eqp.Equipped = false
	w := newCrushingWeapon(eqp, "Swung")
	eqp.Weapons = []*Weapon{w}
	e.Recalculate()
	c.False(eqp.ReallyEquipped(), "precondition: the equipment is not equipped")
	c.True(eqp.switchMattersWhileUnequipped(),
		"a switchable override aimed at the equipment's own weapon matters while it is unequipped")
	c.Equal("cr", resolvedDamageType(w, nil), "the switchable override is ignored while the switch is off")
	eqp.SetSwitchedOn(true)
	var tooltip xbytes.InsertBuffer
	c.Equal("cut", resolvedDamageType(w, &tooltip), "an override selecting this weapon applies while it is on")
	c.Contains(tooltip.String(), "**Amulet**", "the tooltip names the equipment")
	c.NotContains(tooltip.String(), "Unknown")
	c.Equal("2", w.ResolveSelector(selector.WeaponArmorDivisor, "1", nil), "an override naming the weapon applies too")
}

// TestSelectorOverrideWithRequiredSkill verifies that an override selecting weapons by their required skill applies
// to the weapons used with that skill, narrowed by the specialization and relative skill level criteria, and that a
// weapon using no skill is never selected, even by a new override whose name criterion is empty.
func TestSelectorOverrideWithRequiredSkill(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addAverageDXSkill(e, "Broadsword", fxp.Four) // DX+1
	addAverageDXSkill(e, "Axe/Mace", fxp.One)
	override := newDamageTypeOverride(wsel.WithRequiredSkill, "cut")
	override.NameCriteria.Qualifier = "Broadsword"
	addTraitWithFeatures(e, "Imbuement", override)
	gadget := addTraitWithFeatures(e, "Gadget")
	sword := newCrushingWeapon(gadget, "Swung")
	sword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, 0)}
	axe := newCrushingWeapon(gadget, "Swung")
	axe.Defaults = []*SkillDefault{newSkillDefaultTo("Axe/Mace", "", true, 0)}
	unskilled := newCrushingWeapon(gadget, "Swung")
	unskilled.Defaults = []*SkillDefault{{DefaultType: DexterityID}}
	gadget.Weapons = []*Weapon{sword, axe, unskilled}
	e.Recalculate()
	c.Equal("cut", resolvedDamageType(sword, nil), "an override selecting by required skill applies to a weapon using it")
	c.Equal("cr", resolvedDamageType(axe, nil), "not to one using another skill")
	c.Equal("cr", resolvedDamageType(unskilled, nil), "nor to one using no skill")

	fresh := newDamageTypeOverride(wsel.WithRequiredSkill, "cut")
	addTraitWithFeatures(e, "Fresh", fresh)
	e.Recalculate()
	c.Equal("cr", resolvedDamageType(unskilled, nil),
		"a new override, whose name criterion is empty, leaves a weapon using no skill alone")
	c.Equal("cr", resolvedDamageType(axe, nil), "and every other")

	override.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Two}
	c.Equal("cr", resolvedDamageType(sword, nil), "a relative skill level the skill falls short of excludes the weapon")
	override.RelativeLevelCriteria.Qualifier = fxp.One
	c.Equal("cut", resolvedDamageType(sword, nil), "one it reaches keeps it")
	override.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Rapier"}
	c.Equal("cr", resolvedDamageType(sword, nil), "a specialization the skill lacks excludes the weapon")
	override.SpecializationCriteria = criteria.Text{Compare: criteria.IsText}
	c.Equal("cut", resolvedDamageType(sword, nil), `an "is" specialization of nothing keeps an unspecialized skill`)
}

// TestSelectorOverrideWithRequiredSkillAndNoSkill verifies how an override selecting weapons by required skill treats
// a weapon for which no skill-based default resolves, matching what a weapon bonus does: a name criterion that is
// anything, or is not some skill, selects it, while one naming a skill outright does not, even the unfilled name a new
// override starts with.
func TestSelectorOverrideWithRequiredSkillAndNoSkill(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	override := newDamageTypeOverride(wsel.WithRequiredSkill, "cut")
	override.NameCriteria = criteria.Text{Compare: criteria.AnyText}
	addTraitWithFeatures(e, "Imbuement", override)
	gadget := addTraitWithFeatures(e, "Gadget")
	byAttribute := newCrushingWeapon(gadget, "Punch")
	byAttribute.Defaults = []*SkillDefault{{DefaultType: DexterityID}}
	byAbsentSkill := newCrushingWeapon(gadget, "Kick")
	byAbsentSkill.Defaults = []*SkillDefault{newSkillDefaultTo("Brawling", "", true, 0)}
	gadget.Weapons = []*Weapon{byAttribute, byAbsentSkill}
	e.Recalculate()
	c.Equal("cut", resolvedDamageType(byAttribute, nil),
		"a name criterion of anything selects a weapon whose defaults are all attribute-based")
	c.Equal("cut", resolvedDamageType(byAbsentSkill, nil),
		"and one whose skill default names a skill the character lacks")

	override.NameCriteria = criteria.Text{Compare: criteria.IsNotText, Qualifier: "Broadsword"}
	c.Equal("cut", resolvedDamageType(byAttribute, nil), "as does one that is not some skill")

	override.NameCriteria = criteria.Text{Compare: criteria.IsText}
	c.Equal("cr", resolvedDamageType(byAttribute, nil), "the unfilled name a new override starts with selects nothing")

	override.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Brawling"}
	c.Equal("cr", resolvedDamageType(byAbsentSkill, nil), "nor does naming the absent skill the weapon defaults to")
}

// TestSelectorOverrideWithRequiredSkillThroughTechnique verifies that an override qualified by the skill a technique
// is based on applies to a weapon whose best default is the technique, as the weapon bonuses of issue #767 do.
func TestSelectorOverrideWithRequiredSkillThroughTechnique(t *testing.T) {
	c := check.New(t)
	e, err := NewEntityFromFile(os.DirFS("testdata"), "issue767.gcs")
	c.NoError(err)
	var kick *Weapon
	Traverse(func(tr *Trait) bool {
		for _, w := range tr.Weapons {
			if w.UsageWithReplacements() == "Kick" {
				kick = w
			}
		}
		return false
	}, false, false, e.Traits...)
	c.NotNil(kick, "precondition: the file has a Kick")
	if kick == nil {
		return
	}
	override := newDamageTypeOverride(wsel.WithRequiredSkill, "cut")
	override.NameCriteria.Qualifier = "Brawling"
	addTraitWithFeatures(e, "Imbuement", override)
	e.Recalculate()
	c.Equal("cut", resolvedDamageType(kick, nil),
		"an override qualified by the skill the Kicking technique is based on applies to the Kick")
}

// TestSelectorOverrideWithRequiredSkillThroughTaggedTechnique verifies that the technique whose base skill qualifies
// the overrides is the one the weapon's default selects by tag, not a better one of the same name based on another
// skill.
func TestSelectorOverrideWithRequiredSkillThroughTaggedTechnique(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addAverageDXSkill(e, "Brawling", fxp.Twelve) // DX+4, so 14
	addAverageDXSkill(e, "Karate", fxp.Four)     // DX+2, so 12
	byBrawling := NewTechnique(e, nil, "Brawling")
	byBrawling.Name = "Kicking"
	byBrawling.TechniqueDefault.Modifier = -fxp.Two
	byKarate := NewTechnique(e, nil, "Karate")
	byKarate.Name = "Kicking"
	byKarate.TechniqueDefault.Modifier = -fxp.Two
	byKarate.Tags = []string{"Karate"}
	e.Skills = append(e.Skills, byBrawling, byKarate)
	style := addTraitWithFeatures(e, "Style")
	kick := newCrushingWeapon(style, "Kick")
	kick.Defaults = []*SkillDefault{{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Kicking"}},
		Tags:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Karate"}},
	}}
	style.Weapons = []*Weapon{kick}
	override := newDamageTypeOverride(wsel.WithRequiredSkill, "cut")
	override.NameCriteria.Qualifier = "Karate"
	addTraitWithFeatures(e, "Imbuement", override)
	e.Recalculate()
	c.True(byBrawling.LevelData.Level > byKarate.LevelData.Level,
		"precondition: the technique based on Brawling is the better of the two")
	c.Equal(byKarate.LevelData.Level, kick.SkillLevel(nil), "precondition: the Kick uses the technique tagged Karate")
	c.Equal("cut", resolvedDamageType(kick, nil),
		"an override qualified by the base skill of the technique the Kick uses applies to it")
	override.NameCriteria.Qualifier = "Brawling"
	c.Equal("cr", resolvedDamageType(kick, nil),
		"but not one qualified by the base skill of the technique the Kick doesn't use")
}

// TestSelectorOverrideThisWeaponOutranksNamed verifies that, at equal priority, an override selecting the weapons of
// its own item beats one that names the weapon, without the tie being reported as a conflict.
func TestSelectorOverrideThisWeaponOutranksNamed(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	this := newDamageTypeOverride(wsel.ThisWeapon, "cut")
	gadget := addTraitWithFeatures(e, "Gadget", this)
	w := newCrushingWeapon(gadget, "Swung")
	gadget.Weapons = []*Weapon{w}
	named := newDamageTypeOverride(wsel.WithName, "imp")
	named.NameCriteria.Qualifier = "Gadget"
	addTraitWithFeatures(e, "Rule", named)
	e.Recalculate()
	var tooltip xbytes.InsertBuffer
	c.Equal("cut", resolvedDamageType(w, &tooltip), "the override of the weapon's own item wins over one naming it")
	c.NotContains(tooltip.String(), "conflict", "as the more specific of the two")
	c.Contains(tooltip.String(), `overriding **"imp"**`, "and the loser is reported")
	named.Priority = 1
	c.Equal("imp", resolvedDamageType(w, nil), "a higher priority still wins")
}

// TestSelectorOverrideJSONSelectionType verifies that a file from before overrides had a selection type reads as
// selecting weapons by name, which is what it did, and that the selection a new override starts with, the zero value,
// is nonetheless written and read back.
func TestSelectorOverrideJSONSelectionType(t *testing.T) {
	c := check.New(t)
	var legacy Features
	c.NoError(jio.Unmarshal([]byte(`[{"type":"selector_override","field":"weapon_damage_type","value":"cut",`+
		`"name":{"compare":"is","qualifier":"Axe"}}]`), &legacy))
	c.Equal(1, len(legacy))
	if len(legacy) == 1 {
		o, ok := legacy[0].(*SelectorOverride)
		c.True(ok)
		if ok {
			c.Equal(wsel.WithName, o.SelectionType, "a file without a selection type selects weapons by name")
		}
	}
	bySkill := NewSelectorOverride(selector.WeaponDamageType)
	bySkill.NameCriteria.Qualifier = "Broadsword"
	c.Equal(wsel.WithRequiredSkill, bySkill.SelectionType, "precondition: a new override selects by skill")
	data, err := jio.Marshal(Features{bySkill})
	c.NoError(err)
	c.Contains(string(data), `"selection_type":"weapons_with_required_skill"`, "the zero-valued selection is written")
	var restored Features
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal(1, len(restored))
	if len(restored) == 1 {
		o, ok := restored[0].(*SelectorOverride)
		c.True(ok)
		if ok {
			c.Equal(wsel.WithRequiredSkill, o.SelectionType, "and read back")
		}
	}
	checkFeaturesJSONRoundTrip(c, Features{
		bySkill, newDamageTypeOverride(wsel.ThisWeapon, "cut"),
		newDamageTypeOverride(wsel.WithName, "cut"),
	})
}

// TestSelectorOverrideNameableKeysBySelection verifies that only the criteria a selection consults contribute nameable
// keys.
func TestSelectorOverrideNameableKeysBySelection(t *testing.T) {
	c := check.New(t)
	o := NewSelectorOverride(selector.WeaponDamageType)
	o.NameCriteria.Qualifier = "@Name@"
	o.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Spec@"}
	o.UsageCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Usage@"}
	o.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "@Type@"}
	o.TagsCriteria = criteria.Text{Compare: criteria.ContainsText, Qualifier: "@Tag@"}
	keys := func() []string {
		m := make(map[string]string)
		o.FillWithNameableKeys(m, nil)
		return slices.Sorted(maps.Keys(m))
	}
	c.Equal([]string{"Name", "Spec", "Tag", "Type", "Usage"}, keys(), "selecting by required skill consults every criterion")
	o.SelectionType = wsel.WithName
	c.Equal([]string{"Name", "Tag", "Type", "Usage"}, keys(), "selecting by name has no specialization")
	o.SelectionType = wsel.ThisWeapon
	c.Equal([]string{"Type", "Usage"}, keys(), "selecting this weapon has only a usage and damage type")
	o.Field = selector.TraitFrequency
	c.Equal([]string{"Name", "Tag"}, keys(), "a trait-scoped override has a name and tags, whatever its selection")
}
