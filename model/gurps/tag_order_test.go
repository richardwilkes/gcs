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

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// unsortedTags is the order a user might enter tags in; sortedTags is the order they are kept in.
var (
	unsortedTags = []string{"Zeta", "alpha", "Mid"}
	sortedTags   = []string{"Mid", "Zeta", "alpha"}
)

// savedLists is the part of a saved node these tests look at.
type savedLists struct {
	Tags    []string `json:"tags"`
	College []string `json:"college"`
}

// savedListsOf saves the node and returns the tags and colleges in the order they were written.
func savedListsOf(t *testing.T, node any) savedLists {
	t.Helper()
	data, err := jio.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	var lists savedLists
	if err = json.Unmarshal(data, &lists); err != nil {
		t.Fatal(err)
	}
	return lists
}

// TestApplyToSortsTags verifies that committing an editor's data sorts the tags, whatever order they were entered in,
// so that the order shows neither in the item nor in the file it is saved to (#1217).
func TestApplyToSortsTags(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	var traitData TraitEditData
	traitData.CopyFrom(trait)
	traitData.Tags = unsortedTags
	traitData.ApplyTo(trait)
	c.Equal(sortedTags, trait.Tags, "trait")

	traitModifier := NewTraitModifier(nil, nil, false)
	var traitModifierData TraitModifierEditData
	traitModifierData.CopyFrom(traitModifier)
	traitModifierData.Tags = unsortedTags
	traitModifierData.ApplyTo(traitModifier)
	c.Equal(sortedTags, traitModifier.Tags, "trait modifier")

	skill := NewSkill(nil, nil, false)
	var skillData SkillEditData
	skillData.CopyFrom(skill)
	skillData.Tags = unsortedTags
	skillData.ApplyTo(skill)
	c.Equal(sortedTags, skill.Tags, "skill")

	spell := NewSpell(nil, nil, false)
	var spellData SpellEditData
	spellData.CopyFrom(spell)
	spellData.Tags = unsortedTags
	spellData.College = CollegeList{"Water", "Air", "Fire"}
	spellData.ApplyTo(spell)
	c.Equal(sortedTags, spell.Tags, "spell")
	c.Equal(CollegeList{"Air", "Fire", "Water"}, spell.College, "spell colleges")

	equipment := NewEquipment(nil, nil, false)
	var equipmentData EquipmentEditData
	equipmentData.CopyFrom(equipment)
	equipmentData.Tags = unsortedTags
	equipmentData.ApplyTo(equipment)
	c.Equal(sortedTags, equipment.Tags, "equipment")

	equipmentModifier := NewEquipmentModifier(nil, nil, false)
	var equipmentModifierData EquipmentModifierEditData
	equipmentModifierData.CopyFrom(equipmentModifier)
	equipmentModifierData.Tags = unsortedTags
	equipmentModifierData.ApplyTo(equipmentModifier)
	c.Equal(sortedTags, equipmentModifier.Tags, "equipment modifier")

	note := NewNote(nil, nil, false)
	var noteData NoteEditData
	noteData.CopyFrom(note)
	noteData.Tags = unsortedTags
	noteData.ApplyTo(note)
	c.Equal(sortedTags, note.Tags, "note")
}

// TestMarshalSortsTags verifies that saving a node writes its tags sorted even when it holds them in another order, as
// one built by code rather than through an editor might, and that a spell's colleges are written the same way.
func TestMarshalSortsTags(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	trait.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, trait).Tags, "trait")

	traitModifier := NewTraitModifier(nil, nil, false)
	traitModifier.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, traitModifier).Tags, "trait modifier")

	skill := NewSkill(nil, nil, false)
	skill.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, skill).Tags, "skill")

	spell := NewSpell(nil, nil, false)
	spell.Tags = append([]string(nil), unsortedTags...)
	spell.College = CollegeList{"Water", "Air", "Fire"}
	lists := savedListsOf(t, spell)
	c.Equal(sortedTags, lists.Tags, "spell")
	c.Equal([]string{"Air", "Fire", "Water"}, lists.College, "spell colleges")

	equipment := NewEquipment(nil, nil, false)
	equipment.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, equipment).Tags, "equipment")

	equipmentModifier := NewEquipmentModifier(nil, nil, false)
	equipmentModifier.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, equipmentModifier).Tags, "equipment modifier")

	note := NewNote(nil, nil, false)
	note.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, note).Tags, "note")
}

// TestCollegeListLoadsSorted verifies that colleges come back sorted from both the list and the legacy string formats,
// as tags do.
func TestCollegeListLoadsSorted(t *testing.T) {
	c := check.New(t)
	var list CollegeList
	c.NoError(json.Unmarshal([]byte(`["Water", " Air ", "", "Fire"]`), &list))
	c.Equal(CollegeList{"Air", "Fire", "Water"}, list, "list format")
	c.NoError(json.Unmarshal([]byte(`"Water or Air/Fire"`), &list))
	c.Equal(CollegeList{"Air", "Fire", "Water"}, list, "legacy string format")
}
