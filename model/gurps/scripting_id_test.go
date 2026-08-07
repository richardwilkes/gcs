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
	"fmt"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestScriptIDsArePrimitiveStrings verifies that the id and parentID properties every script wrapper exposes arrive in
// JavaScript as primitive strings. Handing goja a tid.TID (a named string type) rather than a plain string makes it
// build a String *object* instead: typeof reports "object", === against a string literal is always false, and
// Array.prototype.includes — which compares with SameValueZero — never matches, so the ids are unusable for the
// comparisons scripts actually need to make.
func TestScriptIDsArePrimitiveStrings(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, true)
	childTrait := NewTrait(e, trait, false)
	trait.Children = []*Trait{childTrait}
	traitMod := NewTraitModifier(e, nil, false)

	skill := NewSkill(e, nil, true)
	childSkill := NewSkill(e, skill, false)
	skill.Children = []*Skill{childSkill}

	spell := NewSpell(e, nil, true)
	childSpell := NewSpell(e, spell, false)
	spell.Children = []*Spell{childSpell}

	equipment := NewEquipment(e, nil, true)
	childEquipment := NewEquipment(e, equipment, false)
	equipment.Children = []*Equipment{childEquipment}
	equipmentMod := NewEquipmentModifier(e, nil, false)

	note := NewNote(e, nil, true)
	childNote := NewNote(e, note, false)
	note.Children = []*Note{childNote}

	weapon := NewWeapon(equipment, true)

	for _, tc := range []struct {
		name     string
		provider ScriptSelfProvider
		expr     string // yields the id under test
		want     tid.TID
	}{
		{name: "trait id", provider: deferredNewScriptTrait(trait), expr: "self.id", want: trait.TID},
		{name: "trait parentID", provider: deferredNewScriptTrait(childTrait), expr: "self.parentID", want: trait.TID},
		{name: "trait modifier id", provider: deferredNewScriptTraitModifier(traitMod), expr: "self.id", want: traitMod.TID},
		{name: "skill id", provider: deferredNewScriptSkill(skill), expr: "self.id", want: skill.TID},
		{name: "skill parentID", provider: deferredNewScriptSkill(childSkill), expr: "self.parentID", want: skill.TID},
		{name: "spell id", provider: deferredNewScriptSpell(spell), expr: "self.id", want: spell.TID},
		{name: "spell parentID", provider: deferredNewScriptSpell(childSpell), expr: "self.parentID", want: spell.TID},
		{name: "equipment id", provider: deferredNewScriptEquipment(equipment), expr: "self.id", want: equipment.TID},
		{
			name:     "equipment parentID",
			provider: deferredNewScriptEquipment(childEquipment),
			expr:     "self.parentID",
			want:     equipment.TID,
		},
		{
			name:     "equipment modifier id",
			provider: deferredNewScriptEquipmentModifier(equipmentMod),
			expr:     "self.id",
			want:     equipmentMod.TID,
		},
		{name: "note id", provider: deferredNewScriptNote(note), expr: "self.id", want: note.TID},
		{name: "note parentID", provider: deferredNewScriptNote(childNote), expr: "self.parentID", want: note.TID},
		{name: "weapon id", provider: deferredNewScriptWeapon(weapon), expr: "self.id", want: weapon.TID},
	} {
		// The value must be a primitive string, not a String object.
		c.Equal("string", ResolveScript(e, tc.provider, "typeof "+tc.expr), "case %q", tc.name)

		// It must hold the expected id...
		c.Equal(string(tc.want), ResolveScript(e, tc.provider, tc.expr), "case %q", tc.name)

		// ...and compare equal to that id with the operators scripts use. A String object fails every one of these.
		for _, cmp := range []struct {
			label  string
			script string
		}{
			{label: "===", script: fmt.Sprintf("%s === %q", tc.expr, tc.want)},
			{label: "includes", script: fmt.Sprintf("[%q].includes(%s)", tc.want, tc.expr)},
			{label: "indexOf", script: fmt.Sprintf("[%q].indexOf(%s) === 0", tc.want, tc.expr)},
			{label: "object key", script: fmt.Sprintf("({%q: 1})[%s] === 1", tc.want, tc.expr)},
		} {
			c.Equal("true", ResolveScript(e, tc.provider, cmp.script), "case %q via %s", tc.name, cmp.label)
		}
	}

	// A missing parent still reports undefined rather than an id.
	c.Equal("undefined", ResolveScript(e, deferredNewScriptTrait(trait), "typeof self.parentID"))
	c.Equal("undefined", ResolveScript(e, deferredNewScriptEquipment(equipment), "typeof self.parentID"))
	c.Equal("undefined", ResolveScript(e, deferredNewScriptNote(note), "typeof self.parentID"))
}

// TestScriptCollectionIDsArePrimitiveStrings verifies that the ids reachable through the collection-valued accessors
// (children, weapons, activeModifiers, and the entity-level find functions) are primitive strings too, since those are
// the paths a script most often uses to gather ids for comparison.
func TestScriptCollectionIDsArePrimitiveStrings(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, true)
	trait.Name = "Container"
	child := NewTrait(e, trait, false)
	child.Name = "Child"
	trait.Children = []*Trait{child}
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Mod"
	child.Modifiers = []*TraitModifier{mod}
	e.Traits = []*Trait{trait}

	item := NewEquipment(e, nil, false)
	item.Name = "Gear"
	weapon := NewWeapon(item, true)
	item.Weapons = []*Weapon{weapon}
	itemMod := NewEquipmentModifier(e, nil, false)
	itemMod.Name = "Gear Mod"
	item.Modifiers = []*EquipmentModifier{itemMod}
	e.CarriedEquipment = []*Equipment{item}

	e.Recalculate()

	for _, tc := range []struct {
		name string
		expr string
		want tid.TID
	}{
		{name: "trait children", expr: `entity.findTraits("Container", "")[0].children[0].id`, want: child.TID},
		{
			name: "trait modifiers",
			expr: `entity.findTraits("Child", "")[0].activeModifiers[0].id`,
			want: mod.TID,
		},
		{name: "equipment weapons", expr: `entity.findEquipment("Gear", "")[0].weapons[0].id`, want: weapon.TID},
		{
			name: "equipment modifiers",
			expr: `entity.findEquipment("Gear", "")[0].activeModifiers[0].id`,
			want: itemMod.TID,
		},
	} {
		c.Equal("string", ResolveScript(e, ScriptSelfProvider{}, "typeof "+tc.expr), "case %q", tc.name)
		c.Equal(string(tc.want), ResolveScript(e, ScriptSelfProvider{}, tc.expr), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, ScriptSelfProvider{}, fmt.Sprintf("[%q].includes(%s)", tc.want, tc.expr)),
			"case %q", tc.name)
	}
}
