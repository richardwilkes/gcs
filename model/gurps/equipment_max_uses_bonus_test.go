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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

func newMaxUsesBonus(sel equipmentsel.Type, amount string) *EquipmentMaxUsesBonus {
	b := NewEquipmentMaxUsesBonus()
	b.SelectionType = sel
	b.Amount = amount
	return b
}

// TestEquipmentMaxUsesBonusThisEquipment covers the locally-applied "to this equipment" bonuses, the operation implied
// by the entered text, the Add -> % -> x stacking order, the [0, MaxEquipmentMaxUses] clamp, and the non-positive
// multiplier guard.
func TestEquipmentMaxUsesBonusThisEquipment(t *testing.T) {
	c := check.New(t)

	eqp := NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	c.Equal(10, eqp.ResolvedMaxUses(), "no bonuses")

	// Add -> % -> x order: ((10 + 2) * 1.5) * 2 = 36.
	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Features = Features{
		newMaxUsesBonus(equipmentsel.ThisEquipment, "+2"),
		newMaxUsesBonus(equipmentsel.ThisEquipment, "50%"),
		newMaxUsesBonus(equipmentsel.ThisEquipment, "x2"),
	}
	c.Equal(36, eqp.ResolvedMaxUses(), "add -> percent -> multiply")

	// Minimum clamp: 1 + (-100) = -99, clamped to 0.
	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 1
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "-100")}
	c.Equal(0, eqp.ResolvedMaxUses(), "clamped to minimum of 0")

	// Maximum clamp: 1,000,000 * 100 = 100,000,000, clamped to MaxEquipmentMaxUses.
	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 1000000
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "x100")}
	c.Equal(MaxEquipmentMaxUses, eqp.ResolvedMaxUses(), "clamped to maximum")

	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "x-5")}
	c.Equal(10, eqp.ResolvedMaxUses(), "non-positive multiplier treated as 1")

	eqp = NewEquipment(nil, nil, false)
	eqp.Equipped = false
	eqp.MaxUses = 10
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "+5")}
	c.Equal(15, eqp.ResolvedMaxUses(), "applies while unequipped")
}

// TestEquipmentMaxUsesBonusFromModifier verifies that a "to this equipment" bonus carried by an equipment modifier is
// applied only while that modifier is enabled.
func TestEquipmentMaxUsesBonusFromModifier(t *testing.T) {
	c := check.New(t)
	eqp := NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "+3")}
	eqp.Modifiers = []*EquipmentModifier{mod}

	mod.Disabled = false
	c.Equal(13, eqp.ResolvedMaxUses(), "enabled modifier bonus applies")

	mod.Disabled = true
	c.Equal(10, eqp.ResolvedMaxUses(), "disabled modifier bonus ignored")
}

// TestEquipmentMaxUsesBonusPerLevel verifies per-level scaling off the item the "to this equipment" bonus is attached
// to (the equipment's own level).
func TestEquipmentMaxUsesBonusPerLevel(t *testing.T) {
	c := check.New(t)

	bonus := newMaxUsesBonus(equipmentsel.ThisEquipment, "+2")
	bonus.PerLevel = true

	eqp := NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Level = fxp.Three
	eqp.Features = Features{bonus}
	c.Equal(16, eqp.ResolvedMaxUses(), "per-level bonus scales by the item's level: 10 + 2*3")
}

// TestEquipmentMaxUsesBonusEquipmentWithName covers the entity-collected "to equipment whose name" selector, including
// name + tag matching and per-level scaling from the attaching trait.
func TestEquipmentMaxUsesBonusEquipmentWithName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	potion := addCarriedEquipmentWithFeatures(e, "Potion")
	potion.Tags = []string{"Consumable"}
	potion.MaxUses = 5

	sword := addCarriedEquipmentWithFeatures(e, "Sword")
	sword.Tags = []string{"Weapon"}
	sword.MaxUses = 5

	// A trait grants +2 max uses to consumables named "Potion".
	bonus := newMaxUsesBonus(equipmentsel.EquipmentWithName, "+2")
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = "Potion"
	bonus.TagsCriteria.Compare = criteria.IsText
	bonus.TagsCriteria.Qualifier = "Consumable"
	trait := addTraitWithFeatures(e, "", bonus)
	e.Recalculate()

	c.Equal(7, potion.ResolvedMaxUses(), "matching name + tag receives the bonus")
	c.Equal(5, sword.ResolvedMaxUses(), "non-matching item is unaffected")

	bonus.PerLevel = true
	trait.CanLevel = true
	trait.Levels = fxp.Three
	e.Recalculate()
	c.Equal(11, potion.ResolvedMaxUses(), "per-level whose-name bonus scales by the trait's level: 5 + 2*3")
}

// TestEquipmentResolvedUses verifies that a feature-reduced maximum caps the displayed Uses without mutating the
// stored value, and that the save-time adjustment brings the stored value back into range.
func TestEquipmentResolvedUses(t *testing.T) {
	c := check.New(t)

	eqp := NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Uses = 8

	c.Equal(8, eqp.ResolvedUses(), "displayed uses without a bonus")

	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "x0.3")} // 10 * 0.3 = 3
	c.Equal(3, eqp.ResolvedMaxUses(), "reduced maximum")
	c.Equal(3, eqp.ResolvedUses(), "displayed uses capped at the reduced maximum")
	c.Equal(8, eqp.Uses, "stored uses left unchanged until an edit or save")

	AdjustEquipmentUsesForSave([]*Equipment{eqp})
	c.Equal(3, eqp.Uses, "stored uses adjusted to the cap on save")

	eqp.Uses = 2
	AdjustEquipmentUsesForSave([]*Equipment{eqp})
	c.Equal(2, eqp.Uses, "in-range stored uses left alone on save")
}

// TestEquipmentMaxUsesBonusCannotCreateMaximum verifies that an item with no maximum uses stays unlimited even when a
// bonus matches it, as a trait with no maximum level does. Bonuses used to be computed from a base of zero, so a +2
// turned "no maximum" into a maximum of 2.
func TestEquipmentMaxUsesBonusCannotCreateMaximum(t *testing.T) {
	c := check.New(t)

	for _, amount := range []string{"+2", "50%", "x2", "-2"} {
		eqp := NewEquipment(nil, nil, false)
		eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, amount)}
		c.Equal(0, eqp.ResolvedMaxUses(), "a %q bonus must not create a maximum", amount)
	}

	eqp := NewEquipment(nil, nil, false)
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "+3")}
	eqp.Modifiers = []*EquipmentModifier{mod}
	c.Equal(0, eqp.ResolvedMaxUses(), "a modifier bonus must not create a maximum")

	e := NewEntity()
	bonus := newMaxUsesBonus(equipmentsel.EquipmentWithName, "+5")
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = "Potion"
	addTraitWithFeatures(e, "", bonus)
	potion := addCarriedEquipmentWithFeatures(e, "Potion")
	e.Recalculate()
	c.Equal(0, potion.ResolvedMaxUses(), "an entity-wide bonus must not create a maximum")
	potion.MaxUses = 1
	c.Equal(6, potion.ResolvedMaxUses(), "precondition: the same bonus raises a declared maximum")
}

// TestEquipmentMaxUsesBonusRoundTrip verifies that each selector/operation combination survives a JSON round-trip.
func TestEquipmentMaxUsesBonusRoundTrip(t *testing.T) {
	c := check.New(t)
	this := newMaxUsesBonus(equipmentsel.ThisEquipment, "x2")
	byName := newMaxUsesBonus(equipmentsel.EquipmentWithName, "+50%")
	byName.PerLevel = true
	byName.NameCriteria.Compare = criteria.IsText
	byName.NameCriteria.Qualifier = "Potion"
	byName.TagsCriteria.Compare = criteria.IsText
	byName.TagsCriteria.Qualifier = "Consumable"
	checkFeaturesJSONRoundTrip(c, Features{this, byName})
}
