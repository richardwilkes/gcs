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

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCloneWeaponsBelongToTheirNewHolder verifies that the weapon copies made when a trait, skill, spell or piece of
// equipment is cloned -- or synced from its library source -- belong to the node now holding them. Weapon.Clone() copies
// the owner over from the original, since it has no way to know the new one, so every caller has to supply it. A copy
// left pointing at the node it came from resolves its nameable replacements (and its skill defaults and minimum ST)
// through that node, so a duplicated row's weapons reported the original's values and followed the original's later
// edits.
func TestCloneWeaponsBelongToTheirNewHolder(t *testing.T) {
	c := check.New(t)

	// Each case builds a node named "@kind@ Attack" holding one weapon, clones it, then diverges the clone's
	// replacements. The clone's weapon must read the clone's map.
	for _, one := range []struct {
		name  string
		build func() (source, clone any, sourceWeapon, cloneWeapon *Weapon)
	}{
		{
			name: "trait",
			build: func() (any, any, *Weapon, *Weapon) {
				n := NewTrait(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(t *Trait) *[]*Weapon { return &t.Weapons })
			},
		},
		{
			name: "skill",
			build: func() (any, any, *Weapon, *Weapon) {
				n := NewSkill(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(s *Skill) *[]*Weapon { return &s.Weapons })
			},
		},
		{
			name: "spell",
			build: func() (any, any, *Weapon, *Weapon) {
				n := NewSpell(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, false, func(s *Spell) *[]*Weapon { return &s.Weapons })
			},
		},
		{
			name: "equipment",
			build: func() (any, any, *Weapon, *Weapon) {
				n := NewEquipment(nil, nil, false)
				n.Name = "@kind@ Attack"
				return cloneWithWeapon(n, true, func(e *Equipment) *[]*Weapon { return &e.Weapons })
			},
		},
	} {
		source, clone, sourceWeapon, cloneWeapon := one.build()
		c.True(cloneWeapon.Owner == clone, "%s: the cloned weapon belongs to the clone", one.name)
		c.True(sourceWeapon.Owner == source, "%s: the original's weapon still belongs to the original", one.name)
		c.Equal(map[string]string{"kind": "Ice"}, cloneWeapon.NameableReplacements(),
			"%s: the cloned weapon resolves against the clone's replacements", one.name)
		c.Equal(map[string]string{"kind": "Fire"}, sourceWeapon.NameableReplacements(),
			"%s: the original's weapon is unaffected", one.name)
		c.True(cloneWeapon.Damage.Owner == cloneWeapon, "%s: the cloned weapon's damage belongs to it", one.name)
	}
}

// cloneWithWeapon gives n, whose name is expected to mention "@kind@", a "Fire" replacement for it and one weapon, then
// clones n and diverges the clone's replacement to "Ice", so the two weapons must resolve differently. Node supplies
// the clone and owner plumbing, WeaponOwner is what NewWeapon needs and nameable.Setter (promoted from each type's edit
// data) sets the replacements; only the Weapons slice is a plain field on every node type, so it needs the accessor.
func cloneWithWeapon[T interface {
	Node[T]
	WeaponOwner
	nameable.Setter
}](n T, melee bool, weapons func(T) *[]*Weapon) (source, clone any, sourceWeapon, cloneWeapon *Weapon) {
	n.SetNameableReplacements(map[string]string{"kind": "Fire"})
	*weapons(n) = []*Weapon{NewWeapon(n, melee)}
	n.SetDataOwner(nil)
	dup := n.Clone(LibraryFile{}, nil, nil, Reference)
	dup.NameableReplacements()["kind"] = "Ice"
	return n, dup, (*weapons(n))[0], (*weapons(dup))[0]
}
