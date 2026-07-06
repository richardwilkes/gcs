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
	"encoding/json/v2"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/toolbox/v2/check"
)

const (
	potionName    = "Potion"
	consumableTag = "Consumable"
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

	// No bonuses: the resolved value equals the raw MaxUses.
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

	// A non-positive multiplier is treated as 1 (no change).
	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "x-5")}
	c.Equal(10, eqp.ResolvedMaxUses(), "non-positive multiplier treated as 1")

	// A "to this equipment" bonus applies even when the item is not equipped.
	eqp = NewEquipment(nil, nil, false)
	eqp.Equipped = false
	eqp.MaxUses = 10
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "+5")}
	c.Equal(15, eqp.ResolvedMaxUses(), "applies while unequipped")
}

// TestEquipmentMaxUsesBonusOperationFromText verifies that the entered text alone determines the operation.
func TestEquipmentMaxUsesBonusOperationFromText(t *testing.T) {
	c := check.New(t)
	cases := []struct {
		amount   string
		maxUses  int
		expected int
		note     string
	}{
		{amount: "1", maxUses: 10, expected: 11, note: "bare number is an addition"},
		{amount: "+1", maxUses: 10, expected: 11, note: "signed positive addition"},
		{amount: "-1", maxUses: 10, expected: 9, note: "signed negative addition"},
		{amount: "10%", maxUses: 10, expected: 11, note: "trailing percent is a percentage"},
		{amount: "+10%", maxUses: 10, expected: 11, note: "signed positive percentage"},
		{amount: "-10%", maxUses: 10, expected: 9, note: "signed negative percentage"},
		{amount: "x2", maxUses: 10, expected: 20, note: "leading x is a multiplier"},
		{amount: "2x", maxUses: 10, expected: 20, note: "trailing x is a multiplier"},
	}
	for _, one := range cases {
		eqp := NewEquipment(nil, nil, false)
		eqp.MaxUses = one.maxUses
		eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, one.amount)}
		c.Equal(one.expected, eqp.ResolvedMaxUses(), one.note)
	}
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

	// The same bonus on a non-leveled item contributes nothing.
	eqp = NewEquipment(nil, nil, false)
	eqp.MaxUses = 10
	eqp.Features = Features{bonus}
	c.Equal(10, eqp.ResolvedMaxUses(), "per-level bonus contributes 0 on a non-leveled item")
}

// TestEquipmentMaxUsesBonusEquipmentWithName covers the entity-collected "to equipment whose name" selector, including
// name + tag matching and per-level scaling from the attaching trait.
func TestEquipmentMaxUsesBonusEquipmentWithName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	potion := NewEquipment(e, nil, false)
	potion.Name = potionName
	potion.Tags = []string{consumableTag}
	potion.MaxUses = 5
	e.CarriedEquipment = append(e.CarriedEquipment, potion)

	sword := NewEquipment(e, nil, false)
	sword.Name = "Sword"
	sword.Tags = []string{"Weapon"}
	sword.MaxUses = 5
	e.CarriedEquipment = append(e.CarriedEquipment, sword)

	// A trait grants +2 max uses to consumables named "Potion".
	bonus := newMaxUsesBonus(equipmentsel.EquipmentWithName, "+2")
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = potionName
	bonus.TagsCriteria.Compare = criteria.IsText
	bonus.TagsCriteria.Qualifier = consumableTag
	trait := NewTrait(e, nil, false)
	trait.Features = append(trait.Features, bonus)
	e.Traits = append(e.Traits, trait)
	e.Recalculate()

	c.Equal(7, potion.ResolvedMaxUses(), "matching name + tag receives the bonus")
	c.Equal(5, sword.ResolvedMaxUses(), "non-matching item is unaffected")

	// Per-level scaling comes from the attaching trait's level.
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

	// With no bonus, the stored and displayed values match.
	c.Equal(8, eqp.ResolvedUses(), "displayed uses without a bonus")

	// A feature that lowers the maximum below the stored Uses caps the displayed value but leaves the stored value.
	eqp.Features = Features{newMaxUsesBonus(equipmentsel.ThisEquipment, "x0.3")} // 10 * 0.3 = 3
	c.Equal(3, eqp.ResolvedMaxUses(), "reduced maximum")
	c.Equal(3, eqp.ResolvedUses(), "displayed uses capped at the reduced maximum")
	c.Equal(8, eqp.Uses, "stored uses left unchanged until an edit or save")

	// The save-time adjustment brings the stored value down to the cap.
	AdjustEquipmentUsesForSave([]*Equipment{eqp})
	c.Equal(3, eqp.Uses, "stored uses adjusted to the cap on save")

	// A stored value already within range is not modified by the save-time adjustment.
	eqp.Uses = 2
	AdjustEquipmentUsesForSave([]*Equipment{eqp})
	c.Equal(2, eqp.Uses, "in-range stored uses left alone on save")
}

// TestEquipmentMaxUsesBonusRoundTrip verifies that each selector/operation combination survives a JSON round-trip.
func TestEquipmentMaxUsesBonusRoundTrip(t *testing.T) {
	c := check.New(t)
	this := newMaxUsesBonus(equipmentsel.ThisEquipment, "x2")
	byName := newMaxUsesBonus(equipmentsel.EquipmentWithName, "+50%")
	byName.PerLevel = true
	byName.NameCriteria.Compare = criteria.IsText
	byName.NameCriteria.Qualifier = potionName
	byName.TagsCriteria.Compare = criteria.IsText
	byName.TagsCriteria.Qualifier = consumableTag
	original := Features{this, byName}

	data, err := json.Marshal(original)
	c.NoError(err)
	var restored Features
	c.NoError(json.Unmarshal(data, &restored))
	c.Equal(len(original), len(restored), "feature count preserved")
	again, err := json.Marshal(restored)
	c.NoError(err)
	c.Equal(string(data), string(again), "re-marshaled JSON is stable across a round-trip")
}
