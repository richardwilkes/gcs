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
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestLoadingDropsNullNodes verifies that a null in any list of nodes a file holds -- the lists of a sheet, template,
// loot sheet or list file, and the children, modifiers and weapons of a node -- is dropped on load rather than becoming
// a nil node that panics the first thing to walk the list.
func TestLoadingDropsNullNodes(t *testing.T) {
	c := check.New(t)
	id := func(node interface{ ID() tid.TID }) string { return string(node.ID()) }
	traits := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"modifiers":[null,{"id":%q,"children":[null]}],"weapons":[null]}]}]`,
		id(NewTrait(nil, nil, true)), id(NewTrait(nil, nil, false)), id(NewTraitModifier(nil, nil, true)))
	skills := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"weapons":[null]}]}]`,
		id(NewSkill(nil, nil, true)), id(NewSkill(nil, nil, false)))
	spells := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"weapons":[null]}]}]`,
		id(NewSpell(nil, nil, true)), id(NewSpell(nil, nil, false)))
	equipment := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"modifiers":[null,{"id":%q,"children":[null]}],"weapons":[null]}]}]`,
		id(NewEquipment(nil, nil, true)), id(NewEquipment(nil, nil, false)), id(NewEquipmentModifier(nil, nil, true)))
	notes := fmt.Sprintf(`[null,{"id":%q,"children":[null]}]`, id(NewNote(nil, nil, true)))

	checkTraits := func(list []*Trait) {
		c.Equal(1, len(list), "traits")
		c.Equal(1, len(list[0].Children), "trait children")
		c.Equal(1, len(list[0].Children[0].Modifiers), "trait modifiers")
		c.Equal(0, len(list[0].Children[0].Modifiers[0].Children), "trait modifier children")
		c.Equal(0, len(list[0].Children[0].Weapons), "trait weapons")
	}
	checkSkills := func(list []*Skill) {
		c.Equal(1, len(list), "skills")
		c.Equal(1, len(list[0].Children), "skill children")
		c.Equal(0, len(list[0].Children[0].Weapons), "skill weapons")
	}
	checkSpells := func(list []*Spell) {
		c.Equal(1, len(list), "spells")
		c.Equal(1, len(list[0].Children), "spell children")
		c.Equal(0, len(list[0].Children[0].Weapons), "spell weapons")
	}
	checkEquipment := func(list []*Equipment) {
		c.Equal(1, len(list), "equipment")
		c.Equal(1, len(list[0].Children), "equipment children")
		c.Equal(1, len(list[0].Children[0].Modifiers), "equipment modifiers")
		c.Equal(0, len(list[0].Children[0].Modifiers[0].Children), "equipment modifier children")
		c.Equal(0, len(list[0].Children[0].Weapons), "equipment weapons")
	}
	checkNotes := func(list []*Note) {
		c.Equal(1, len(list), "notes")
		c.Equal(0, len(list[0].Children), "note children")
	}

	var entity Entity
	c.NoError(json.Unmarshal([]byte(`{"traits":`+traits+`,"skills":`+skills+`,"spells":`+spells+`,"equipment":`+
		equipment+`,"other_equipment":`+equipment+`,"notes":`+notes+`}`), &entity))
	checkTraits(entity.Traits)
	checkSkills(entity.Skills)
	checkSpells(entity.Spells)
	checkEquipment(entity.CarriedEquipment)
	checkEquipment(entity.OtherEquipment)
	checkNotes(entity.Notes)
	var legacyEntity Entity
	c.NoError(json.Unmarshal([]byte(`{"advantages":`+traits+`}`), &legacyEntity))
	checkTraits(legacyEntity.Traits)

	var tmpl Template
	c.NoError(json.Unmarshal([]byte(`{"traits":`+traits+`,"skills":`+skills+`,"spells":`+spells+`,"equipment":`+
		equipment+`,"notes":`+notes+`}`), &tmpl))
	checkTraits(tmpl.Traits)
	checkSkills(tmpl.Skills)
	checkSpells(tmpl.Spells)
	checkEquipment(tmpl.Equipment)
	checkNotes(tmpl.Notes)
	var legacyTemplate Template
	c.NoError(json.Unmarshal([]byte(`{"advantages":`+traits+`}`), &legacyTemplate))
	checkTraits(legacyTemplate.Traits)

	var loot Loot
	c.NoError(json.Unmarshal([]byte(`{"equipment":`+equipment+`,"notes":`+notes+`}`), &loot))
	checkEquipment(loot.Equipment)
	checkNotes(loot.Notes)

	listFile := func(rows string) fstest.MapFS {
		data := fmt.Sprintf(`{"version":%d,"rows":%s}`, jio.CurrentDataVersion, rows)
		return fstest.MapFS{"list": &fstest.MapFile{Data: []byte(data)}}
	}
	loadedTraits, err := NewTraitsFromFile(listFile(traits), "list")
	c.NoError(err)
	checkTraits(loadedTraits)
	loadedSkills, err := NewSkillsFromFile(listFile(skills), "list")
	c.NoError(err)
	checkSkills(loadedSkills)
	loadedSpells, err := NewSpellsFromFile(listFile(spells), "list")
	c.NoError(err)
	checkSpells(loadedSpells)
	loadedEquipment, err := NewEquipmentFromFile(listFile(equipment), "list")
	c.NoError(err)
	checkEquipment(loadedEquipment)
	loadedNotes, err := NewNotesFromFile(listFile(notes), "list")
	c.NoError(err)
	checkNotes(loadedNotes)
	loadedTraitModifiers, err := NewTraitModifiersFromFile(listFile(fmt.Sprintf(`[null,{"id":%q,"children":[null]}]`,
		id(NewTraitModifier(nil, nil, true)))), "list")
	c.NoError(err)
	c.Equal(1, len(loadedTraitModifiers), "trait modifier rows")
	c.Equal(0, len(loadedTraitModifiers[0].Children), "trait modifier row children")
	loadedEquipmentModifiers, err := NewEquipmentModifiersFromFile(listFile(fmt.Sprintf(`[null,{"id":%q,"children":[null]}]`,
		id(NewEquipmentModifier(nil, nil, true)))), "list")
	c.NoError(err)
	c.Equal(1, len(loadedEquipmentModifiers), "equipment modifier rows")
	c.Equal(0, len(loadedEquipmentModifiers[0].Children), "equipment modifier row children")
}
