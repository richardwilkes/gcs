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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/toolbox/v2/check"
)

// expectedExtendedValue works out the extended value of equipment presenting no choice the way it has always been
// worked out, independently of the ranges it now comes from.
func expectedExtendedValue(e *Equipment) fxp.Int {
	if e.Quantity <= 0 {
		return 0
	}
	value := e.AdjustedValue()
	for _, one := range e.Children {
		value += expectedExtendedValue(one)
	}
	return value.Mul(e.Quantity)
}

// expectedExtendedWeight works out the extended weight of equipment presenting no choice the way it has always been
// worked out, independently of the ranges it now comes from.
func expectedExtendedWeight(e *Equipment) fxp.Weight {
	if e.Quantity <= 0 {
		return 0
	}
	var contents fxp.Int
	for _, one := range e.Children {
		contents += fxp.Int(expectedExtendedWeight(one))
	}
	if len(e.Children) != 0 {
		reduction := containedWeightReductionFor(e, fxp.Pound, e.Modifiers, e.Features)
		contents = fxp.Int(reduction.apply(fxp.Weight(contents)))
	}
	return fxp.Weight((fxp.Int(e.AdjustedWeight(false, fxp.Pound)) + contents).Mul(e.Quantity))
}

// TestEquipmentTotalsAgreeWithTheirRanges verifies, for equipment presenting no choice, that the extended value and
// weight worked out from the ranges are exactly what they always were, that the ranges are settled on them, and that
// the editor's own working of the extended weight agrees, across modifiers, levels, quantities, nesting, groups and
// contained weight reductions, switched on and off.
func TestEquipmentTotalsAgreeWithTheirRanges(t *testing.T) {
	c := check.New(t)
	perLevel := func(parent *Equipment) *Equipment {
		one := newEquipmentItem("Sword", "100", "3 lb")
		one.Level = fxp.FromInteger(2)
		one.Quantity = fxp.FromInteger(3)
		mod := NewEquipmentModifier(nil, nil, false)
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
		mod.CostIsPerLevel = true
		mod.WeightType = emweight.Original
		mod.WeightAmount = "+1 lb"
		mod.WeightIsPerLevel = true
		one.SetModifiers([]*EquipmentModifier{mod})
		if parent != nil {
			one.SetParent(parent)
			parent.Children = append(parent.Children, one)
		}
		return one
	}
	reducing := func(reduction string, switchable, switchedOn bool) *Equipment {
		bag := NewEquipment(nil, nil, true)
		bag.BaseValue = "25"
		bag.BaseWeight = "2 lb"
		bag.Quantity = fxp.FromInteger(2)
		bag.SwitchedOn = switchedOn
		cwr := NewContainedWeightReduction()
		cwr.Reduction = reduction
		cwr.Switchable = switchable
		bag.Features = Features{cwr}
		perLevel(bag)
		item(bag, "Rope", "5", "2 lb")
		none := item(bag, "Torch", "3", "1 lb")
		none.Quantity = 0
		inner := NewEquipment(nil, bag, true)
		inner.BaseWeight = "0.5 lb"
		item(inner, "Coin", "1", "0.02 lb").Quantity = fxp.FromInteger(40)
		bag.Children = append(bag.Children, inner)
		return bag
	}
	group := NewEquipmentGroup(nil, nil)
	item(group, "Knife", "40", "1 lb")
	bagInGroup := reducing("25%", false, false)
	bagInGroup.SetParent(group)
	group.Children = append(group.Children, bagInGroup)
	fractional := newEquipmentItem("Flour", "2", "1 lb")
	fractional.Quantity = fxp.FromStringForced("2.5")

	for _, one := range []struct {
		name string
		eqp  *Equipment
	}{
		{"per-level modifiers", perLevel(nil)},
		{"percentage reduction", reducing("50%", false, false)},
		{"fixed reduction", reducing("2 lb", false, false)},
		{"full reduction", reducing("100%", false, false)},
		{"switchable reduction, switched off", reducing("50%", true, false)},
		{"switchable reduction, switched on", reducing("50%", true, true)},
		{"group", group},
		{"fractional quantity", fractional},
	} {
		value, settled := one.eqp.ExtendedValueRange().Settled()
		c.True(settled, "%s: the value range must be settled", one.name)
		c.Equal(expectedExtendedValue(one.eqp), value, "%s: the value range", one.name)
		c.Equal(expectedExtendedValue(one.eqp), one.eqp.ExtendedValue(), "%s: the extended value", one.name)
		weight, settled := one.eqp.ExtendedWeightRange(fxp.Pound).Settled()
		c.True(settled, "%s: the weight range must be settled", one.name)
		c.Equal(fxp.Int(expectedExtendedWeight(one.eqp)), weight, "%s: the weight range", one.name)
		c.Equal(expectedExtendedWeight(one.eqp), one.eqp.ExtendedWeight(false, fxp.Pound), "%s: the extended weight",
			one.name)
		c.Equal(one.eqp.ExtendedWeight(false, fxp.Pound), ExtendedWeightAdjustedForModifiers(one.eqp, fxp.Pound,
			one.eqp.Quantity, one.eqp.ResolvedBaseWeight(), one.eqp.Modifiers, one.eqp.Features, one.eqp.Children,
			false, false), "%s: the editor's working of the extended weight", one.name)
	}

	// The weight that counts toward the encumbrance for skills leaves out equipped equipment marked "Ignore weight for
	// skills", which the ranges don't, so it is worked out on its own.
	pack := NewEquipment(nil, nil, true)
	pack.BaseWeight = "3 lb"
	armor := item(pack, "Armor", "100", "20 lb")
	armor.WeightIgnoredForSkills = true
	item(pack, "Rope", "5", "2 lb")
	c.Equal(fxp.Weight(fxp.FromInteger(25)), pack.ExtendedWeight(false, fxp.Pound))
	c.Equal(fxp.Weight(fxp.FromInteger(5)), pack.ExtendedWeight(true, fxp.Pound),
		"the armor's weight must not count for skills")
}
