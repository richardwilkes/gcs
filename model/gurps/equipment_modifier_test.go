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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEquipmentModifierSyncHashIncludesPerLevelAndPerPoundFlags verifies that the sync hash reflects the
// CostIsPerLevel, CostIsPerPound, and WeightIsPerLevel flags. They change the cost/weight math without necessarily
// changing CostAmount/WeightAmount, so leaving them out let library sync report a modified library modifier as Matched
// and never update the local copy. The TID is not hashed, so a lone flag must change the hash.
func TestEquipmentModifierSyncHashIncludesPerLevelAndPerPoundFlags(t *testing.T) {
	c := check.New(t)

	hashWith := func(mutate func(e *EquipmentModifier)) uint64 {
		e := NewEquipmentModifier(nil, nil, false)
		e.CostAmount = "-60%"
		e.WeightAmount = "2 lb"
		if mutate != nil {
			mutate(e)
		}
		return Hash64(e)
	}

	baseHash := hashWith(nil)

	// A flat "-60%" turning into "-60% per level" flips only CostIsPerLevel; the hash must change.
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.CostIsPerLevel = true }),
		"CostIsPerLevel must affect the sync hash")
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.CostIsPerPound = true }),
		"CostIsPerPound must affect the sync hash")
	c.NotEqual(baseHash, hashWith(func(e *EquipmentModifier) { e.WeightIsPerLevel = true }),
		"WeightIsPerLevel must affect the sync hash")
}

// TestEquipmentModifierCloneDoesNotShareReplacements verifies that a copy of a modifier still carrying the legacy
// replacements map gets its own map, so that later merges into the equipment the copy migrates its map into (see
// SetTarget) cannot write into the library row the copy came from.
func TestEquipmentModifierCloneDoesNotShareReplacements(t *testing.T) {
	c := check.New(t)

	// A library row with legacy replacements, as it arrives from an .eqm file written by an older version.
	library := NewEquipmentModifier(nil, nil, false)
	library.Name = "@Material@ Plating"
	library.Replacements = map[string]string{"Material": "Steel"}

	// Dropping it onto a sheet's equipment row clones it...
	dup := library.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(map[string]string{"Material": "Steel"}, dup.Replacements, "the copy carries the replacements")

	// ...into a map of its own. This is the check the rest of the test rests on: SetTarget's merge clones a map it is
	// handed when the target has none, so the steps below would pass even if the copy still shared the library's map.
	dup.Replacements["Finish"] = "Blued"
	c.Equal(map[string]string{"Material": "Steel"}, library.Replacements,
		"writing into the copy's map must not reach the library row")
	delete(dup.Replacements, "Finish")

	// The next attachment pass migrates the copy's map into the equipment, which has none of its own.
	equipment := NewEquipment(nil, nil, false)
	dup.SetTarget(equipment)
	c.Nil(dup.Replacements, "the copy hands its map off to the equipment")
	c.Equal(map[string]string{"Material": "Steel"}, equipment.Replacements, "the equipment picks up the replacements")

	// A second modifier attaching to the same equipment takes SetTarget's merge branch, writing into the map the
	// equipment now holds. That must not reach back into the library row.
	second := NewEquipmentModifier(nil, nil, false)
	second.Replacements = map[string]string{"Finish": "Blued"}
	second.SetTarget(equipment)
	c.Equal(map[string]string{"Material": "Steel", "Finish": "Blued"}, equipment.Replacements,
		"the equipment merges in the second modifier's replacements")
	c.Equal(map[string]string{"Material": "Steel"}, library.Replacements, "the library row is left untouched")
}

// TestEquipmentModifierApplyNameableKeysKeepsOtherReplacements verifies that applying one modifier's replacements keeps
// the equipment's existing replacements, those held for its disabled modifiers, and those applied for its other
// modifiers.
func TestEquipmentModifierApplyNameableKeysKeepsOtherReplacements(t *testing.T) {
	c := check.New(t)
	equipment := NewEquipment(nil, nil, false)
	equipment.Name = "@Foo@ Thing"
	equipment.Replacements = map[string]string{"Foo": "Bar", "Finish": "Gilded"}
	finish := NewEquipmentModifier(nil, nil, false)
	finish.Name = "@Finish@ Coating"
	finish.SetDisabled(true)
	material := NewEquipmentModifier(nil, nil, false)
	material.Name = "@Material@ Coating"
	color := NewEquipmentModifier(nil, nil, false)
	color.Name = "@Color@ Paint"
	equipment.AddModifiers(finish, material, color)

	// As the prompt does: collect every modifier's keys first, then apply them in turn.
	materialKeys := make(map[string]string)
	material.FillWithNameableKeys(materialKeys, nil)
	colorKeys := make(map[string]string)
	color.FillWithNameableKeys(colorKeys, nil)
	materialKeys["Material"] = "Steel"
	colorKeys["Color"] = "Red"
	material.ApplyNameableKeys(materialKeys)
	color.ApplyNameableKeys(colorKeys)

	c.Equal(map[string]string{"Foo": "Bar", "Finish": "Gilded", "Material": "Steel", "Color": "Red"},
		equipment.Replacements, "the answer held for the switched-off modifier must survive")
	c.Equal("Bar Thing", equipment.NameWithReplacements())
	c.Equal("Steel Coating", material.NameWithReplacements())
	c.Equal("Red Paint", color.NameWithReplacements())
	finish.SetDisabled(false)
	c.Equal("Gilded Coating", finish.NameWithReplacements(), "switching the modifier back on must find its answer")
}

// TestEquipmentModifierFillWithNameableKeys verifies that the nameable keys of an equipment modifier are gathered from
// its text, its features and its weapons.
func TestEquipmentModifierFillWithNameableKeys(t *testing.T) {
	c := check.New(t)
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Name = "@Name@"
	mod.LocalNotes = "@Notes@"
	mod.Features = Features{newSkillBonusTo("@Skill@", fxp.One)}
	w := NewWeapon(nil, true)
	w.Usage = "@Usage@"
	mod.Weapons = []*Weapon{w}
	m := make(map[string]string)
	mod.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{
		"Name":  nameable.Unset,
		"Notes": nameable.Unset,
		"Skill": nameable.Unset,
		"Usage": nameable.Unset,
	}, m)
}

// TestEquipmentModifierSyncWithSource verifies that a modifier which has drifted from its library source pulls its
// cost, features and weapons back across as copies of its own, the weapons belonging to its equipment and held by it.
func TestEquipmentModifierSyncWithSource(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + EquipmentModifiersExt}
	source := NewEquipmentModifier(nil, nil, false)
	source.Name = "Fine"
	source.CostType = emcost.Base
	source.CostAmount = "+3 CF"
	source.Features = Features{newSkillBonusTo("Broadsword", fxp.One)}
	source.Weapons = []*Weapon{NewWeapon(nil, true)}
	local := NewEquipmentModifier(e, nil, false)
	local.Name = "Fine"
	local.CostAmount = "+10"
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	eqp := addCarriedEquipmentWithFeatures(e, "Sword")
	eqp.AddModifiers(local)
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)

	local.SyncWithSource()
	c.Equal(emcost.Base, local.CostType)
	c.Equal("+3 CF", local.CostAmount)
	c.Equal(1, len(local.Features), "the source's feature is synced")
	if len(local.Features) == 1 {
		c.True(local.Features[0] != source.Features[0], "the synced feature is a copy, not the source's")
	}
	c.Equal(1, len(local.Weapons), "the source's weapon is synced")
	if len(local.Weapons) == 1 {
		w := local.Weapons[0]
		c.True(w != source.Weapons[0], "the synced weapon is a copy, not the source's")
		c.True(w.Owner == eqp, "the synced weapon belongs to the modifier's equipment")
		c.True(w.HoldingModifier() == local, "the synced weapon is held by the modifier")
	}
}

// TestEquipmentModifierEnabledCell verifies that a modifier's enabled cell is a checkbox showing whether it is enabled,
// and that a group, which can't be disabled, has none.
func TestEquipmentModifierEnabledCell(t *testing.T) {
	c := check.New(t)
	leaf := NewEquipmentModifier(nil, nil, false)
	var data CellData
	leaf.CellData(EquipmentModifierEnabledColumn, &data)
	c.Equal(cell.Toggle, data.Type)
	c.True(data.Checked, "an enabled modifier is checked")
	leaf.SetDisabled(true)
	data = CellData{}
	leaf.CellData(EquipmentModifierEnabledColumn, &data)
	c.False(data.Checked, "a disabled modifier is unchecked")
	data = CellData{}
	NewEquipmentModifier(nil, nil, true).CellData(EquipmentModifierEnabledColumn, &data)
	c.NotEqual(cell.Toggle, data.Type)
}

// TestEquipmentModifierAdjustmentDescriptions verifies how a modifier describes its cost and weight adjustments, alone
// and together, including in its full description when the sheet shows them.
func TestEquipmentModifierAdjustmentDescriptions(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.ShowEquipmentModifierAdj = true
	for _, one := range []struct {
		name   string
		adjust func(mod *EquipmentModifier)
		cost   string
		weight string
		both   string
	}{
		{
			"cost", func(mod *EquipmentModifier) { mod.CostAmount = "+10" },
			"+10 to original cost", "", "+10 to original cost",
		},
		{
			"cost per level and per pound", func(mod *EquipmentModifier) {
				mod.CostAmount = "+5"
				mod.CostIsPerLevel = true
				mod.CostIsPerPound = true
			},
			"+5 per level per pound to original cost", "", "+5 per level per pound to original cost",
		},
		{
			"weight", func(mod *EquipmentModifier) { mod.WeightAmount = "+2 lb" },
			"", "+2 lb to original weight", "+2 lb to original weight",
		},
		{
			"weight per level", func(mod *EquipmentModifier) {
				mod.WeightType = emweight.Final
				mod.WeightAmount = "x2"
				mod.WeightIsPerLevel = true
			},
			"", "x2 per level to final weight", "x2 per level to final weight",
		},
		{
			"both", func(mod *EquipmentModifier) {
				mod.CostType = emcost.Base
				mod.CostAmount = "+1 CF"
				mod.WeightAmount = "+2 lb"
			},
			"+1 CF to base cost", "+2 lb to original weight", "+1 CF to base cost; +2 lb to original weight",
		},
	} {
		mod := NewEquipmentModifier(e, nil, false)
		mod.Name = "Mod"
		one.adjust(mod)
		c.Equal(one.cost, mod.CostDescription(), "%s: cost", one.name)
		c.Equal(one.weight, mod.WeightDescription(), "%s: weight", one.name)
		c.Equal(one.both, mod.FullCostDescription(), "%s: both", one.name)
		c.Equal("Mod ["+one.both+"]", mod.FullDescription(), "%s: the full description shows both", one.name)
	}
}

// TestValueAdjustedForModifiers verifies how modifiers adjust a value: percentages of the original value and
// multipliers to the original cost, then cost factors, which take off no more than 80%, then the same adjustments to
// the final base cost and then to the final cost, never coming to less than nothing.
func TestValueAdjustedForModifiers(t *testing.T) {
	c := check.New(t)
	mod := func(costType emcost.Type, amount string) *EquipmentModifier {
		one := NewEquipmentModifier(nil, nil, false)
		one.CostType = costType
		one.CostAmount = amount
		return one
	}
	for _, one := range []struct {
		name     string
		mods     []*EquipmentModifier
		expected int
	}{
		{"percentage", []*EquipmentModifier{mod(emcost.Original, "+10%")}, 110},
		{"multiplier", []*EquipmentModifier{mod(emcost.Original, "x2")}, 200},
		{
			"percentages are of the value the step starts with",
			[]*EquipmentModifier{mod(emcost.Original, "x2"), mod(emcost.Original, "+10%")},
			210,
		},
		{"cost factor", []*EquipmentModifier{mod(emcost.Base, "+2 CF")}, 300},
		{"cost factors add up", []*EquipmentModifier{mod(emcost.Base, "+1 CF"), mod(emcost.Base, "+0.5 CF")}, 250},
		{"multiplier as a cost factor", []*EquipmentModifier{mod(emcost.Base, "x3")}, 300},
		{"cost factors take off no more than 80%", []*EquipmentModifier{mod(emcost.Base, "-0.9 CF")}, 20},
		{
			"final base cost follows the cost factors",
			[]*EquipmentModifier{mod(emcost.Base, "+1 CF"), mod(emcost.FinalBase, "+10")},
			210,
		},
		{
			"final cost comes last",
			[]*EquipmentModifier{mod(emcost.Base, "+1 CF"), mod(emcost.FinalBase, "+50%"), mod(emcost.Final, "+10")},
			310,
		},
		{"never less than nothing", []*EquipmentModifier{mod(emcost.Original, "-150")}, 0},
	} {
		c.Equal(fxp.FromInteger(one.expected), ValueAdjustedForModifiers(NewEquipment(nil, nil, false),
			fxp.Hundred, 0, one.mods), one.name)
	}
}

// TestWeightAdjustedForModifiers verifies how modifiers adjust a weight: percentages of the original weight, then at
// each of the base, final base and final weight steps in turn, multipliers followed by additions, never coming to less
// than nothing.
func TestWeightAdjustedForModifiers(t *testing.T) {
	c := check.New(t)
	mod := func(weightType emweight.Type, amount string) *EquipmentModifier {
		one := NewEquipmentModifier(nil, nil, false)
		one.WeightType = weightType
		one.WeightAmount = amount
		return one
	}
	for _, one := range []struct {
		name     string
		mods     []*EquipmentModifier
		expected int
	}{
		{"percentage", []*EquipmentModifier{mod(emweight.Original, "+50%")}, 15},
		{
			"percentages are of the original weight",
			[]*EquipmentModifier{mod(emweight.Original, "+2 lb"), mod(emweight.Original, "+50%")},
			17,
		},
		{"multiplier", []*EquipmentModifier{mod(emweight.Base, "x2")}, 20},
		{"percentage multiplier", []*EquipmentModifier{mod(emweight.Base, "x50%")}, 5},
		{"fractional multiplier", []*EquipmentModifier{mod(emweight.FinalBase, "x1/2")}, 5},
		{
			"multipliers before additions",
			[]*EquipmentModifier{mod(emweight.Base, "+1 lb"), mod(emweight.Base, "x2")},
			21,
		},
		{
			"steps in turn",
			[]*EquipmentModifier{
				mod(emweight.Original, "+2 lb"), mod(emweight.Base, "x2"), mod(emweight.FinalBase, "+1 lb"),
				mod(emweight.Final, "x2"),
			},
			50,
		},
		{"never less than nothing", []*EquipmentModifier{mod(emweight.Original, "-20 lb")}, 0},
	} {
		c.Equal(fxp.WeightFromInteger(one.expected, fxp.Pound),
			WeightAdjustedForModifiers(NewEquipment(nil, nil, false), fxp.WeightFromInteger(10, fxp.Pound), one.mods,
				fxp.Pound), one.name)
	}
}
