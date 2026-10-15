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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
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

// A trait's levels are plain data that no disabling changes the meaning of, so the level commands stay available for a
// disabled trait: on a sheet enforcing trait prerequisites, decrementing is the one-step fix for a level above the
// maximum, which is what disabled the trait, and on a template, disabled alternatives are kept on purpose and still
// need their levels set.
func TestTraitLevelsStayAdjustableWhileDisabled(t *testing.T) {
	c := check.New(t)
	allowed := func(trait *gurps.Trait, increment bool) bool {
		_, ok := traitLevelExtractor(increment)(trait)
		return ok
	}

	entity := gurps.NewEntity()
	entity.SheetSettings.EnforceTraitPrereqs = true
	trait := gurps.NewTrait(entity, nil, false)
	trait.CanLevel = true
	trait.Levels = fxp.Four
	trait.MaxLevels = "3"
	entity.Traits = []*gurps.Trait{trait}
	entity.Recalculate()
	c.True(trait.DisabledByPrereqs(), "a level above the maximum disables the trait")
	c.True(allowed(trait, false), "decrementing the level of a trait its level disabled")
	c.True(allowed(trait, true), "incrementing the level of a trait its level disabled")

	trait.SetDisabled(true)
	c.True(allowed(trait, false), "decrementing the level of a trait the user disabled")
	c.True(allowed(trait, true), "incrementing the level of a trait the user disabled")

	tmpl := gurps.NewTemplate()
	alternative := gurps.NewTrait(tmpl, nil, false)
	alternative.CanLevel = true
	alternative.Levels = fxp.Two
	alternative.SetDisabled(true)
	tmpl.Traits = []*gurps.Trait{alternative}
	c.True(allowed(alternative, false), "decrementing the level of a disabled alternative on a template")
	c.True(allowed(alternative, true), "incrementing the level of a disabled alternative on a template")
}

// Clicking the equipped checkmark of disabled equipment is refused, leaving the item and the checkmark as they were.
func TestEquippedCheckmarkRefusesDisabledEquipment(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Equipped = true
	eqp.SetDisabled(true)
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	_, table := NewNodeTable(NewEquipmentProvider(entity, true, false), nil)
	node := NewNode(table, nil, eqp, false)
	cellData := &gurps.CellData{Checked: true}
	label, ok := node.createToggleCell(cellData, unison.Black).(*unison.Label)
	c.True(ok, "the toggle cell must be a label")

	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(eqp.Equipped, "the click must not unequip disabled equipment")
	c.True(cellData.Checked, "the checkmark must be put back")
}
