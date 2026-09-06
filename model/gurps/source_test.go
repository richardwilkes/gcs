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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSourcePathSeparatorNormalization verifies that source paths are stored and round-tripped using forward slashes,
// regardless of the separator they were authored with. See issue #1005, where files created on Windows stored source
// paths with backslash separators (e.g. `Basic Set\Basic Set Traits.adq`). Those paths could not be located on other
// platforms, causing the library source match status to show as a question mark.
func TestSourcePathSeparatorNormalization(t *testing.T) {
	c := check.New(t)

	// A path loaded with backslash separators is normalized to forward slashes.
	const windowsAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set\\Basic Set Traits.adq","id":"x"}}`
	var loaded SourcedID
	c.NoError(jio.Unmarshal([]byte(windowsAuthored), &loaded), "backslash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded.Source.Path, "path should be normalized on load")

	// An already-normalized path is left unchanged on load.
	const unixAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set/Basic Set Traits.adq","id":"x"}}`
	var loaded2 SourcedID
	c.NoError(jio.Unmarshal([]byte(unixAuthored), &loaded2), "forward-slash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded2.Source.Path, "forward-slash path should be unchanged")

	// A source carrying a backslash path in memory is always written out with forward slashes.
	inMemory := SourcedID{
		TID: "a",
		Source: Source{
			LibraryFile: LibraryFile{Library: "Master Library", Path: `Basic Set\Basic Set Traits.adq`},
			TID:         "x",
		},
	}
	data, err := jio.Marshal(&inMemory)
	c.NoError(err, "source should marshal")
	c.False(strings.Contains(string(data), `\`), "marshaled output must not contain backslashes")
	c.True(strings.Contains(string(data), "Basic Set/Basic Set Traits.adq"), "marshaled output should use forward slashes")

	// Round-tripping the marshaled output preserves the normalized path.
	var roundTripped SourcedID
	c.NoError(jio.Unmarshal(data, &roundTripped), "marshaled source should load")
	c.Equal("Basic Set/Basic Set Traits.adq", roundTripped.Source.Path, "round-tripped path should remain normalized")
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
	trait.Disabled = true
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
