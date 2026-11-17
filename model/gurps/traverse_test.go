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

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTraverseVisitsDepthFirst verifies that Traverse visits each node before its children, and a node's descendants
// before its next sibling, which decides the node a walk that stops early finds.
func TestTraverseVisitsDepthFirst(t *testing.T) {
	c := check.New(t)
	newTrait := func(name string, children ...*Trait) *Trait {
		trait := NewTrait(nil, nil, len(children) != 0)
		trait.Name = name
		trait.Children = children
		return trait
	}
	list := []*Trait{newTrait("A", newTrait("A1", newTrait("A1a")), newTrait("A2")), newTrait("B")}
	var names []string
	Traverse(func(trait *Trait) bool {
		names = append(names, trait.Name)
		return false
	}, false, false, list...)
	c.Equal([]string{"A", "A1", "A1a", "A2", "B"}, names)
}

// TestCountNodes verifies that CountNodes descends into containers, honors the enabled-only flag the way Traverse does,
// and applies the optional filter.
func TestCountNodes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewTrait(e, nil, true)
	child := NewTrait(e, group, false)
	group.Children = append(group.Children, child)
	disabled := NewTrait(e, nil, true)
	disabled.SetDisabled(true)
	hidden := NewTrait(e, disabled, false)
	disabled.Children = append(disabled.Children, hidden)
	list := []*Trait{group, disabled}
	c.Equal(4, CountNodes(list, false, nil), "a nil filter counts every node, containers included")
	c.Equal(2, CountNodes(list, true, nil), "a disabled node and its descendants are skipped")
	c.Equal(2, CountNodes(list, false, func(t *Trait) bool { return !t.Container() }), "the filter picks the nodes")
	c.Equal(0, CountNodes[*Trait](nil, false, nil))
}

// TestForEachSourcedNodeVisitsEveryNodeOnce verifies that the walk shared by source syncing and hashing reaches every
// node a provider holds exactly once: nested children, the modifiers of traits and equipment, disabled nodes, and both
// equipment lists, while the lists a provider doesn't have (a loot sheet's traits, skills and spells) contribute
// nothing.
func TestForEachSourcedNodeVisitsEveryNodeOnce(t *testing.T) {
	c := check.New(t)
	visits := func(provider ListProvider) map[sourcedNode]int {
		m := make(map[sourcedNode]int)
		forEachSourcedNode(provider, func(node sourcedNode) { m[node]++ })
		return m
	}

	tmpl := NewTemplate()
	traitContainer := NewTrait(tmpl, nil, true)
	trait := NewTrait(tmpl, traitContainer, false)
	trait.SetDisabled(true)
	traitMod := NewTraitModifier(tmpl, nil, false)
	trait.Modifiers = append(trait.Modifiers, traitMod)
	traitContainer.Children = append(traitContainer.Children, trait)
	tmpl.Traits = append(tmpl.Traits, traitContainer)
	skill := NewSkill(tmpl, nil, false)
	tmpl.Skills = append(tmpl.Skills, skill)
	spell := NewSpell(tmpl, nil, false)
	tmpl.Spells = append(tmpl.Spells, spell)
	eqp := NewEquipment(tmpl, nil, false)
	eqpMod := NewEquipmentModifier(tmpl, nil, false)
	eqp.Modifiers = append(eqp.Modifiers, eqpMod)
	tmpl.Equipment = append(tmpl.Equipment, eqp)
	note := NewNote(tmpl, nil, false)
	tmpl.Notes = append(tmpl.Notes, note)
	got := visits(tmpl)
	c.Equal(8, len(got), "every node of the template is visited")
	for _, node := range []sourcedNode{traitContainer, trait, traitMod, skill, spell, eqp, eqpMod, note} {
		c.Equal(1, got[node], "each node is visited exactly once")
	}

	e := NewEntity()
	before := len(visits(e))
	carried := NewEquipment(e, nil, false)
	e.CarriedEquipment = append(e.CarriedEquipment, carried)
	other := NewEquipment(e, nil, false)
	e.OtherEquipment = append(e.OtherEquipment, other)
	got = visits(e)
	c.Equal(before+2, len(got), "both equipment lists of an entity are visited")
	c.Equal(1, got[carried], "carried equipment is visited once")
	c.Equal(1, got[other], "other equipment is visited once")

	loot := NewLoot()
	lootEqp := NewEquipment(loot, nil, false)
	lootMod := NewEquipmentModifier(loot, nil, false)
	lootEqp.Modifiers = append(lootEqp.Modifiers, lootMod)
	loot.Equipment = append(loot.Equipment, lootEqp)
	lootNote := NewNote(loot, nil, false)
	loot.Notes = append(loot.Notes, lootNote)
	got = visits(loot)
	c.Equal(3, len(got), "a loot sheet's equipment, its modifier and its note are visited and nothing else")
	c.Equal(1, got[lootMod], "the loot equipment's modifier is visited once")
}
