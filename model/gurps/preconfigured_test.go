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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLoadingASheetClearsPreconfigured verifies that a character sheet or a loot sheet clears the preconfigured mark on
// every row it loads, however deeply nested, since the mark means nothing on a sheet.
func TestLoadingASheetClearsPreconfigured(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	group := NewTrait(entity, nil, true)
	group.Preconfigured = true
	nested := NewTrait(entity, group, false)
	nested.Preconfigured = true
	group.Children = []*Trait{nested}
	entity.Traits = []*Trait{group}
	skill := NewSkill(entity, nil, false)
	skill.Preconfigured = true
	entity.Skills = []*Skill{skill}
	spell := NewSpell(entity, nil, false)
	spell.Preconfigured = true
	entity.Spells = []*Spell{spell}
	carried := NewEquipment(entity, nil, false)
	carried.Preconfigured = true
	entity.CarriedEquipment = []*Equipment{carried}
	other := NewEquipment(entity, nil, false)
	other.Preconfigured = true
	entity.OtherEquipment = []*Equipment{other}
	note := NewNote(entity, nil, false)
	note.Preconfigured = true
	entity.Notes = []*Note{note}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loaded Entity
	c.NoError(json.Unmarshal(data, &loaded))
	c.False(loaded.Traits[0].Preconfigured, "a trait container must lose the mark")
	c.False(loaded.Traits[0].Children[0].Preconfigured, "a nested trait must lose the mark")
	c.False(loaded.Skills[0].Preconfigured, "a skill must lose the mark")
	c.False(loaded.Spells[0].Preconfigured, "a spell must lose the mark")
	c.False(loaded.CarriedEquipment[0].Preconfigured, "carried equipment must lose the mark")
	c.False(loaded.OtherEquipment[0].Preconfigured, "other equipment must lose the mark")
	c.False(loaded.Notes[0].Preconfigured, "a note must lose the mark")

	loot := NewLoot()
	lootItem := NewEquipment(loot, nil, false)
	lootItem.Preconfigured = true
	loot.Equipment = []*Equipment{lootItem}
	lootNote := NewNote(loot, nil, false)
	lootNote.Preconfigured = true
	loot.Notes = []*Note{lootNote}
	data, err = json.Marshal(loot)
	c.NoError(err)
	var loadedLoot Loot
	c.NoError(json.Unmarshal(data, &loadedLoot))
	c.False(loadedLoot.Equipment[0].Preconfigured, "loot equipment must lose the mark")
	c.False(loadedLoot.Notes[0].Preconfigured, "a loot note must lose the mark")
}
