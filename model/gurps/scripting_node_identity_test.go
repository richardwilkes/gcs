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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/toolbox/v2/check"
)

// Covers the properties every node script wrapper shares -- parent, container and tags -- both as a top-level node and
// as a child. parent must be the wrapper for the actual parent, reachable recursively up the chain, and undefined at
// the top; container must follow the node's kind; and tags must be a copy, so a script cannot alter the node.
func TestScriptNodeIdentityProperties(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	trait := NewTrait(e, nil, true)
	childTrait := NewTrait(e, trait, false)
	trait.Children = []*Trait{childTrait}
	trait.Tags = []string{"Mental", "Physical"}

	skill := NewSkill(e, nil, true)
	childSkill := NewSkill(e, skill, false)
	skill.Children = []*Skill{childSkill}
	skill.Tags = []string{"Combat"}

	spell := NewSpell(e, nil, true)
	childSpell := NewSpell(e, spell, false)
	spell.Children = []*Spell{childSpell}
	spell.Tags = []string{"Fire", "Air"}

	equipment := NewEquipment(e, nil, true)
	childEquipment := NewEquipment(e, equipment, false)
	equipment.Children = []*Equipment{childEquipment}
	equipment.Tags = []string{"Gear"}

	note := NewNote(e, nil, true)
	childNote := NewNote(e, note, false)
	note.Children = []*Note{childNote}
	note.Tags = []string{"Reminder"}

	for _, tc := range []struct {
		name     string
		top      ScriptSelfProvider
		child    ScriptSelfProvider
		wantTags string
	}{
		{
			name:     "trait",
			top:      deferredNewScriptTrait(trait),
			child:    deferredNewScriptTrait(childTrait),
			wantTags: "Mental,Physical",
		},
		{
			name:     "skill",
			top:      deferredNewScriptSkill(skill),
			child:    deferredNewScriptSkill(childSkill),
			wantTags: "Combat",
		},
		{
			name:     "spell",
			top:      deferredNewScriptSpell(spell),
			child:    deferredNewScriptSpell(childSpell),
			wantTags: "Fire,Air",
		},
		{
			name:     "equipment",
			top:      deferredNewScriptEquipment(equipment),
			child:    deferredNewScriptEquipment(childEquipment),
			wantTags: "Gear",
		},
		{
			name:     "note",
			top:      deferredNewScriptNote(note),
			child:    deferredNewScriptNote(childNote),
			wantTags: "Reminder",
		},
	} {
		c.Equal("undefined", ResolveScript(e, tc.top, "typeof self.parent"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.top, "self.container"), "case %q", tc.name)
		c.Equal(tc.wantTags, ResolveScript(e, tc.top, "self.tags.join(',')"), "case %q", tc.name)

		// A child reaches its parent's wrapper through parent, and the chain ends there.
		c.Equal("false", ResolveScript(e, tc.child, "self.container"), "case %q", tc.name)
		c.Equal("", ResolveScript(e, tc.child, "self.tags.join(',')"), "case %q", tc.name)
		c.Equal(tc.top.ID, ResolveScript(e, tc.child, "self.parent.id"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.child, "self.parent.id === self.parentID"), "case %q", tc.name)
		c.Equal("true", ResolveScript(e, tc.child, "self.parent.container"), "case %q", tc.name)
		c.Equal(tc.wantTags, ResolveScript(e, tc.child, "self.parent.tags.join(',')"), "case %q", tc.name)
		c.Equal("undefined", ResolveScript(e, tc.child, "typeof self.parent.parent"), "case %q", tc.name)

		// tags is a copy: a script that pushes onto it does not alter the node.
		c.Equal("1", ResolveScript(e, tc.child, "self.tags.push('x')"), "case %q", tc.name)
		c.Equal("", ResolveScript(e, tc.child, "self.tags.join(',')"), "case %q", tc.name)
	}
}

// The notes property every wrapper exposes yields the node's secondary text even when the sheet is configured not to
// display notes at all, since a script wants the content, not the layout decision.
func TestScriptNotesReportAllSecondaryText(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.SheetSettings.NotesDisplay = display.NotShown

	trait := NewTrait(e, nil, false)
	trait.LocalNotes = "trait notes"
	traitMod := NewTraitModifier(e, nil, false)
	traitMod.LocalNotes = "trait modifier notes"
	skill := NewSkill(e, nil, false)
	skill.LocalNotes = "skill notes"
	spell := NewSpell(e, nil, false)
	spell.LocalNotes = "spell notes"
	equipment := NewEquipment(e, nil, false)
	equipment.LocalNotes = "equipment notes"
	equipmentMod := NewEquipmentModifier(e, nil, false)
	equipmentMod.LocalNotes = "equipment modifier notes"

	for _, tc := range []struct {
		name     string
		provider ScriptSelfProvider
		want     string
	}{
		{name: "trait", provider: deferredNewScriptTrait(trait), want: "trait notes"},
		{name: "trait modifier", provider: deferredNewScriptTraitModifier(traitMod), want: "trait modifier notes"},
		{name: "skill", provider: deferredNewScriptSkill(skill), want: "skill notes"},
		{name: "spell", provider: deferredNewScriptSpell(spell), want: "spell notes"},
		{name: "equipment", provider: deferredNewScriptEquipment(equipment), want: "equipment notes"},
		{
			name:     "equipment modifier",
			provider: deferredNewScriptEquipmentModifier(equipmentMod),
			want:     "equipment modifier notes",
		},
	} {
		c.Equal("true", ResolveScript(e, tc.provider, fmt.Sprintf("self.notes.includes(%q)", tc.want)), "case %q",
			tc.name)
	}
}
