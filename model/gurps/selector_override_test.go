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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
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

// TestSelectorFieldDescriptors verifies the field wiring: every selector.Field has a descriptor, and a constrained
// field's suggested states are all valid stored values with a non-empty picker label, which catches enum-key typos in
// the descriptor table.
func TestSelectorFieldDescriptors(t *testing.T) {
	c := check.New(t)
	for _, field := range selector.Fields {
		d := SelectorFieldDescriptorFor(field)
		c.Equal(field, d.Field, "descriptor is registered for %v", field)
		if d.FreeForm || d.StateTitle == nil {
			continue
		}
		for _, state := range d.SuggestedStates {
			c.True(d.StateTitle(state) != "", "state %q of %v has a label", state, field)
		}
	}

	// A duplicated or misspelled strength-basis key would collapse two picker entries onto the same stdmg option.
	d := SelectorFieldDescriptorFor(selector.WeaponDamageStrengthBasis)
	seen := make(map[stdmg.Option]bool)
	for _, state := range d.SuggestedStates {
		opt := stdmg.ExtractOption(state)
		c.Equal(state, opt.Key(), "state %q round-trips through stdmg", state)
		c.False(seen[opt], "state %q is not a duplicate", state)
		seen[opt] = true
	}

	// The numeric damage fields validate fixed-point input; the dice-spec fields are free-form with no validator.
	for _, field := range []selector.Field{
		selector.WeaponArmorDivisor, selector.WeaponFragmentationArmorDivisor,
		selector.WeaponDamageStrengthMultiplier, selector.WeaponDamagePerDieModifier,
	} {
		nd := SelectorFieldDescriptorFor(field)
		c.True(nd.Validate != nil, "%v has a validator", field)
		if nd.Validate != nil {
			c.True(nd.Validate("2"), "%v accepts a number", field)
			c.False(nd.Validate("abc"), "%v rejects a non-number", field)
		}
	}
	for _, field := range []selector.Field{
		selector.WeaponBaseDamageDice, selector.WeaponBaseDamageDicePerLevel, selector.WeaponFragmentationDice,
	} {
		c.True(SelectorFieldDescriptorFor(field).Validate == nil, "%v is free-form (no validator)", field)
	}

	c.Equal(SelectorScopeWeapon, SelectorFieldDescriptorFor(selector.WeaponDamageType).Scope, "damage type is weapon-scoped")
	freq := SelectorFieldDescriptorFor(selector.TraitFrequency)
	c.Equal(SelectorScopeTrait, freq.Scope, "trait frequency is trait-scoped")

	// A bad frequency key would mislabel or drop a choice.
	c.Equal(len(frequency.Rolls), len(freq.SuggestedStates), "one state per frequency roll")
	for _, state := range freq.SuggestedStates {
		n, err := strconv.Atoi(state)
		c.True(err == nil, "state %q is numeric", state)
		roll := frequency.Roll(n)
		c.Equal(roll, roll.EnsureValid(), "state %q is a valid frequency roll", state)
		c.Equal(roll.String(), freq.StateTitle(state), "state %q label matches the roll", state)
	}

	// The self-control fields are trait-scoped constrained enums whose states round-trip through their enums.
	for _, field := range []selector.Field{selector.TraitSelfControlRoll, selector.TraitSelfControlAdjustment} {
		c.Equal(SelectorScopeTrait, SelectorFieldDescriptorFor(field).Scope, "%v is trait-scoped", field)
	}
	cr := SelectorFieldDescriptorFor(selector.TraitSelfControlRoll)
	c.Equal(len(selfctrl.Rolls), len(cr.SuggestedStates), "one state per self-control roll")
	for _, state := range cr.SuggestedStates {
		n, err := strconv.Atoi(state)
		c.True(err == nil, "state %q is numeric", state)
		roll := selfctrl.Roll(n)
		c.Equal(roll, roll.EnsureValid(), "state %q is a valid self-control roll", state)
		c.Equal(roll.String(), cr.StateTitle(state), "state %q label matches the roll", state)
	}
	adj := SelectorFieldDescriptorFor(selector.TraitSelfControlAdjustment)
	c.Equal(len(selfctrl.Adjustments), len(adj.SuggestedStates), "one state per adjustment")
	for _, state := range adj.SuggestedStates {
		option := selfctrl.ExtractAdjustment(state)
		c.Equal(state, option.Key(), "state %q round-trips through the adjustment key", state)
		c.Equal(option.String(), adj.StateTitle(state), "state %q label matches the adjustment", state)
	}
}

// TestSelectorOverrideSpecificityScope verifies that only the criteria the override actually matches on count toward
// specificity. A trait-scoped override never consults the usage criterion (traits have no usage), and a weapon-scoped
// one consults the specialization and relative skill level only when selecting by required skill, so a leftover
// criterion -- which the editor hides once the field or the selection no longer uses it -- must not silently win a
// tie-break it does nothing to constrain. An override selecting the weapons of its own item counts as having pinned
// their name and tags, since it is at least that narrow.
func TestSelectorOverrideSpecificityScope(t *testing.T) {
	c := check.New(t)

	trait := NewSelectorOverride(selector.TraitSelfControlRoll)
	c.Equal(1, trait.OverrideSpecificity(), "only the name criterion constrains a trait-scoped override")
	trait.UsageCriteria.Compare = criteria.IsText
	trait.UsageCriteria.Qualifier = "Swung"
	c.Equal(1, trait.OverrideSpecificity(), "a usage criterion a trait can't match adds no specificity")
	trait.TagsCriteria.Compare = criteria.ContainsText
	trait.TagsCriteria.Qualifier = "Mental"
	c.Equal(2, trait.OverrideSpecificity(), "the tag criterion does constrain a trait-scoped override")

	weapon := NewSelectorOverride(selector.WeaponDamageType)
	c.Equal(wsel.WithRequiredSkill, weapon.SelectionType, "precondition: a new weapon-scoped override selects by skill")
	c.Equal(1, weapon.OverrideSpecificity(), "only the name criterion is set initially")
	weapon.UsageCriteria.Compare = criteria.IsText
	weapon.UsageCriteria.Qualifier = "Swung"
	c.Equal(2, weapon.OverrideSpecificity(), "usage still counts for a weapon-scoped override")
	weapon.SpecializationCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Rapier"}
	c.Equal(3, weapon.OverrideSpecificity(), "the specialization constrains a selection by required skill")
	weapon.RelativeLevelCriteria = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One}
	c.Equal(4, weapon.OverrideSpecificity(), "as does the relative skill level")
	weapon.SelectionType = wsel.WithName
	c.Equal(2, weapon.OverrideSpecificity(), "neither constrains a selection by name")
	weapon.SelectionType = wsel.ThisWeapon
	c.Equal(3, weapon.OverrideSpecificity(), "this weapon counts as name and tags, plus the usage")
	weapon.DamageTypeCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "cut"}
	c.Equal(4, weapon.OverrideSpecificity(), "the damage type constrains a selection of this weapon")
	weapon.SelectionType = wsel.WithName
	c.Equal(3, weapon.OverrideSpecificity(), "and a selection by name")
	weapon.SelectionType = wsel.WithRequiredSkill
	c.Equal(5, weapon.OverrideSpecificity(), "and a selection by required skill")
	trait.DamageTypeCriteria = weapon.DamageTypeCriteria
	c.Equal(2, trait.OverrideSpecificity(), "but not a trait-scoped override, which traits can't match")
	weapon.DamageTypeCriteria = criteria.Text{Compare: criteria.AnyText}
	weapon.SelectionType = wsel.ThisWeapon
	weapon.UsageCriteria = criteria.Text{Compare: criteria.AnyText}
	c.Equal(2, weapon.OverrideSpecificity(), "and as name and tags alone without one")
	weapon.SelectionType = wsel.WithName
	c.Equal(1, weapon.OverrideSpecificity(), "where a selection by name with the same criteria counts the name alone")
}

// TestSelectorFieldDescriptorForUnknownField verifies that an unrecognized field falls back to the zero-value free-form
// descriptor. Normalizing the field first mapped every out-of-range value onto the first enum value, so an unknown
// field quietly received the trait self-control roll descriptor -- wrong scope, wrong suggested states, and a
// self-control key seeded into any new override built from it.
func TestSelectorFieldDescriptorForUnknownField(t *testing.T) {
	c := check.New(t)

	unknown := selector.LastField + 1
	d := SelectorFieldDescriptorFor(unknown)
	c.Equal(unknown, d.Field, "the unknown field is echoed back")
	c.True(d.FreeForm, "an unknown field is free-form")
	c.Equal(0, len(d.SuggestedStates), "an unknown field suggests no states")
	c.True(d.StateTitle == nil, "an unknown field has no state labeler")
	c.True(d.Validate == nil, "an unknown field has no validator")
	c.Equal(SelectorScopeWeapon, d.Scope, "an unknown field uses the zero-value scope")
	c.Equal("", NewSelectorOverride(unknown).Value, "no state is seeded for an unknown field")
}

// TestSelectorOverrideUsesOwnerSubstitutions verifies that an override's criteria are filled from its owner's
// substitutions, as with any bonus, not from those of the weapon or trait being matched, since those are the
// substitutions its markers were filed under.
func TestSelectorOverrideUsesOwnerSubstitutions(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	weaponOverride := NewSelectorOverride(selector.WeaponDamageType)
	weaponOverride.SelectionType = wsel.WithName
	weaponOverride.Value = "cut"
	weaponOverride.NameCriteria.Qualifier = "@Weapon@"
	traitOverride := NewSelectorOverride(selector.TraitSelfControlRoll)
	traitOverride.Value = strconv.Itoa(int(selfctrl.CR6))
	traitOverride.NameCriteria.Qualifier = "@Target@"
	master := NewTrait(e, nil, false)
	master.Name = "Master"
	master.Features = Features{weaponOverride, traitOverride}
	master.Replacements = map[string]string{"Weapon": "Gadget", "Target": "Greed"}

	gadget := NewTrait(e, nil, false)
	gadget.Name = "Gadget"
	w := NewWeapon(gadget, true)
	w.Damage.Type = "cr"
	gadget.Weapons = []*Weapon{w}

	greed := NewTrait(e, nil, false)
	greed.Name = "Greed"
	greed.SelfControl = selfctrl.CR12

	e.Traits = append(e.Traits, master, gadget, greed)
	e.Recalculate()
	c.Equal("cut", w.ResolveSelector(selector.WeaponDamageType, w.Damage.Type, nil),
		"the weapon override matches through its owner's substitution")
	c.Equal(selfctrl.CR6, greed.ResolvedSelfControl(nil), "the trait override matches through its owner's substitution")
}
