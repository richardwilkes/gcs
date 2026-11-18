// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// A disabled skill, spell or piece of equipment offers none of the increment and decrement adjustments, in either
// direction, and neither does one inside a disabled container. Disabled equipment can't be equipped or unequipped
// either. A trait's levels stay adjustable; see TestTraitLevelsStayAdjustableWhileDisabled.
func TestAdjustmentsSkipDisabledItems(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	skill := gurps.NewSkill(entity, nil, false)
	skill.Points = fxp.Four
	skill.TechLevel = new("3")
	entity.Skills = []*gurps.Skill{skill}

	spell := gurps.NewSpell(entity, nil, false)
	spell.Points = fxp.Four
	entity.Spells = []*gurps.Spell{spell}

	parent := gurps.NewEquipment(entity, nil, true)
	eqp := gurps.NewEquipment(entity, parent, false)
	eqp.Quantity = fxp.Two
	eqp.Level = fxp.Two
	eqp.TechLevel = "3"
	eqp.MaxUses = 3
	eqp.Uses = 1
	parent.Children = []*gurps.Equipment{eqp}
	entity.CarriedEquipment = []*gurps.Equipment{parent}

	type adjustment struct {
		name    string
		allowed func() bool
	}
	skillChecks := []adjustment{
		{"skill points up", func() bool { _, ok := rawPointsExtractor[*gurps.Skill](true)(skill); return ok }},
		{"skill points down", func() bool { _, ok := rawPointsExtractor[*gurps.Skill](false)(skill); return ok }},
		{"skill level up", func() bool { _, ok := skillLevelExtractor[*gurps.Skill](true)(skill); return ok }},
		{"skill level down", func() bool { _, ok := skillLevelExtractor[*gurps.Skill](false)(skill); return ok }},
		{"skill TL up", func() bool { _, ok := techLevelExtractor[*gurps.Skill](fxp.One)(skill); return ok }},
		{"skill TL down", func() bool { _, ok := techLevelExtractor[*gurps.Skill](-fxp.One)(skill); return ok }},
	}
	spellChecks := []adjustment{
		{"spell points up", func() bool { _, ok := rawPointsExtractor[*gurps.Spell](true)(spell); return ok }},
		{"spell points down", func() bool { _, ok := rawPointsExtractor[*gurps.Spell](false)(spell); return ok }},
		{"spell level up", func() bool { _, ok := skillLevelExtractor[*gurps.Spell](true)(spell); return ok }},
		{"spell level down", func() bool { _, ok := skillLevelExtractor[*gurps.Spell](false)(spell); return ok }},
	}
	eqpChecks := []adjustment{
		{"quantity up", func() bool { _, ok := quantityExtractor(true)(eqp); return ok }},
		{"quantity down", func() bool { _, ok := quantityExtractor(false)(eqp); return ok }},
		{"equipment level up", func() bool { _, ok := equipmentLevelExtractor(fxp.One)(eqp); return ok }},
		{"equipment level down", func() bool { _, ok := equipmentLevelExtractor(-fxp.One)(eqp); return ok }},
		{"equipment TL up", func() bool { _, ok := techLevelExtractor[*gurps.Equipment](fxp.One)(eqp); return ok }},
		{"equipment TL down", func() bool { _, ok := techLevelExtractor[*gurps.Equipment](-fxp.One)(eqp); return ok }},
		{"uses up", func() bool { _, ok := usesExtractor(1)(eqp); return ok }},
		{"uses down", func() bool { _, ok := usesExtractor(-1)(eqp); return ok }},
		{"uses reset", func() bool { _, ok := usesResetExtractor(eqp); return ok }},
		{"toggle equipped", func() bool { _, ok := equippedExtractor(eqp); return ok }},
	}
	expect := func(checks []adjustment, allowed bool, state string) {
		for _, one := range checks {
			c.Equal(allowed, one.allowed(), one.name+" when "+state)
		}
	}

	expect(skillChecks, true, "enabled")
	expect(spellChecks, true, "enabled")
	expect(eqpChecks, true, "enabled")

	skill.SetDisabled(true)
	spell.SetDisabled(true)
	eqp.SetDisabled(true)
	expect(skillChecks, false, "disabled")
	expect(spellChecks, false, "disabled")
	expect(eqpChecks, false, "disabled")

	eqp.SetDisabled(false)
	parent.SetDisabled(true)
	expect(eqpChecks, false, "inside a disabled container")
}

// TestQuantityCommandsSkipGroups verifies that Increment and Decrement, which adjust equipment's quantity, leave a
// group alone, since its quantity is always one.
func TestQuantityCommandsSkipGroups(t *testing.T) {
	c := check.New(t)
	item := gurps.NewEquipment(nil, nil, false)
	group := gurps.NewEquipmentGroup(nil, nil)
	sheet := newTestSheetWithEquipment(t, item, group)
	table := sheet.CarriedEquipment.Table

	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	c.False(canAdjustQuantity(table, true), "a group alone must not offer to be incremented")
	c.False(canAdjustQuantity(table, false), "a group alone must not offer to be decremented")

	table.SetSelectionMap(map[tid.TID]bool{item.ID(): true, group.ID(): true})
	adjustQuantity(sheet, table, true)
	c.Equal(fxp.FromInteger(2), item.Quantity, "the item must be incremented")
	c.Equal(fxp.One, group.Quantity, "the group must be left at one")
}

// TestQuantityCommandsStepToWholeNumbers verifies that Increment and Decrement move a quantity to the next whole number
// up or down, that Decrement isn't offered once nothing is left, and that undo puts the quantity back.
func TestQuantityCommandsStepToWholeNumbers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		quantity  fxp.Int
		increment bool
		offered   bool
		want      fxp.Int
	}{
		{name: "increment a whole number", quantity: fxp.Two, increment: true, offered: true, want: fxp.Three},
		{name: "increment a fraction", quantity: fxp.TwoAndAHalf, increment: true, offered: true, want: fxp.Three},
		{name: "decrement a whole number", quantity: fxp.Two, offered: true, want: fxp.One},
		{name: "decrement a fraction", quantity: fxp.TwoAndAHalf, offered: true, want: fxp.Two},
		{name: "decrement a fraction of one", quantity: fxp.Half, offered: true, want: 0},
		{name: "decrement nothing", quantity: 0, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			item := gurps.NewEquipment(nil, nil, false)
			item.Quantity = tc.quantity
			sheet := newTestSheetWithEquipment(t, item)
			table := sheet.CarriedEquipment.Table
			table.SetSelectionMap(map[tid.TID]bool{item.ID(): true})
			c.Equal(tc.offered, canAdjustQuantity(table, tc.increment), "whether the command is offered")
			adjustQuantity(sheet, table, tc.increment)
			c.Equal(tc.want, item.Quantity, "the adjusted quantity")
			mgr := unison.UndoManagerFor(table)
			c.Equal(tc.offered, mgr.CanUndo(), "whether the adjustment is undoable")
			if tc.offered {
				mgr.Undo()
				c.Equal(tc.quantity, item.Quantity, "undo must put the quantity back")
			}
		})
	}
}
