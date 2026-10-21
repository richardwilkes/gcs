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
	"github.com/richardwilkes/toolbox/v2/tid"
)

// linkedTrait returns a trait whose own text uses @style@, with a modifier whose @focus@ is linked to it and whose
// @extra@ is not.
func linkedTrait() (*Trait, *TraitModifier) {
	trait := NewTrait(nil, nil, false)
	trait.Name = "Trained in @style@"
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = "@focus@ focus"
	mod.LocalNotes = "with @extra@"
	trait.Modifiers = []*TraitModifier{mod}
	mod.SetTarget(trait)
	trait.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"focus": "style"}}
	return trait, mod
}

func TestLinkedModifierCodeResolvesThroughOwner(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo", "extra": "zeal"}
	c.Equal("Judo focus", mod.NameWithReplacements())
	c.Equal("with zeal", mod.LocalNotesWithReplacements())
}

func TestLinkedModifierCodeIsAskedAsOwnerCode(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo"}
	m := make(map[string]string)
	trait.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"style": "Judo", "extra": nameable.Unset}, m, "the linked code is asked as the trait's")

	// An answer given for the trait's code is the one the linked code shows.
	m["style"] = "Karate"
	m["extra"] = "zeal"
	trait.ApplyNameableKeys(m)
	c.Equal("Karate focus", mod.NameWithReplacements())
	c.Equal(map[string]string{"style": "Karate", "extra": "zeal"}, trait.Replacements,
		"nothing is kept under the modifier's own code")
}

func TestCloneRekeysNameableLinks(t *testing.T) {
	c := check.New(t)
	trait, mod := linkedTrait()
	trait.Replacements = map[string]string{"style": "Judo"}
	clone := trait.Clone(LibraryFile{}, nil, nil, Reference)
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 {
		return
	}
	cloned := clone.Modifiers[0]
	c.NotEqual(mod.TID, cloned.TID, "a reference copy gets a fresh ID")
	c.Equal(map[tid.TID]map[string]string{cloned.TID: {"focus": "style"}}, clone.NameableLinks)
	c.Equal("Judo focus", cloned.NameWithReplacements())

	clone.NameableLinks[cloned.TID]["focus"] = "other"
	c.Equal("style", trait.NameableLinks[mod.TID]["focus"], "the original's links are not shared")
}

func TestLinkedEquipmentModifierCode(t *testing.T) {
	c := check.New(t)
	eqp := NewEquipment(nil, nil, false)
	eqp.Name = "Sword of @metal@"
	mod := NewEquipmentModifier(nil, nil, false)
	mod.Name = "@material@ inlay"
	eqp.SetModifiers([]*EquipmentModifier{mod})
	eqp.NameableLinks = map[tid.TID]map[string]string{mod.TID: {"material": "metal"}}
	eqp.Replacements = map[string]string{"metal": "Silver"}
	c.Equal("Silver inlay", mod.NameWithReplacements())
	m := make(map[string]string)
	eqp.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"metal": "Silver"}, m, "the linked code is asked as the equipment's")
}

func TestNameablesTakePartInSync(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.Name = "Trained in @style@"
	before := Hash64(trait)
	trait.Nameables = map[string]nameable.Detail{}
	c.Equal(before, Hash64(trait), "an item without nameables hashes as it did before they existed")
	trait.Nameables["style"] = nameable.Detail{Label: "Style", Options: []string{"Judo"}}
	c.NotEqual(before, Hash64(trait), "a change to its nameables puts a copy out of sync with its source")
}
