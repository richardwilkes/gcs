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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestContextMenuItemsMatchTheirActions checks that every item a table provider puts in its context menu names a
// registered action and carries that action's own title, or the one contextMenuShortTitles registers for it, so that a
// menu can never show a stale copy of a title that was changed on the action. Separators are the only items that name
// no action.
func TestContextMenuItemsMatchTheirActions(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	actions := map[int]*unison.Action{unison.DeleteItemID: unison.DeleteAction()}
	for _, binding := range gurps.CurrentBindings() {
		actions[binding.Action.ID] = binding.Action
	}
	lists := &listsForTest{}
	entity := gurps.NewEntity()
	providers := map[string]interface{ ContextMenuItems() []ContextMenuItem }{
		"traits":              NewTraitsProvider(lists, false),
		"trait modifiers":     NewTraitModifiersProvider(&traitModifierListProvider{}, false),
		"skills":              NewSkillsProvider(entity, false),
		"spells":              NewSpellsProvider(entity, false),
		"carried equipment":   NewEquipmentProvider(lists, true, false),
		"other equipment":     NewEquipmentProvider(lists, false, false),
		"equipment modifiers": NewEquipmentModifiersProvider(&equipmentModifierListProvider{}, false),
		"notes":               NewNotesProvider(lists, false),
		"melee weapons":       NewWeaponsProvider(lists, true, false),
		"ranged weapons":      NewWeaponsProvider(lists, false, false),
	}
	shared := AppendDefaultContextMenuItems(nil)
	for name, provider := range providers {
		items := provider.ContextMenuItems()
		if len(items) <= len(shared) {
			t.Errorf("%s: the provider adds nothing of its own to the context menu", name)
			continue
		}
		c.Equal(shared, items[len(items)-len(shared):], "%s: the shared items follow the provider's own", name)
		for i, item := range items {
			where := fmt.Sprintf("%s item %d", name, i)
			if item.ID == -1 {
				c.Equal("", item.Title, "%s: a separator has no title", where)
				continue
			}
			action, ok := actions[item.ID]
			if !ok {
				t.Errorf("%s: no registered action has ID %d", where, item.ID)
				continue
			}
			want := action.Title
			if title, has := contextMenuShortTitles[item.ID]; has {
				want = title
			}
			c.Equal(want, item.Title, "%s: the item carries its action's title", where)
		}
	}
}

// TestContextMenusUseShortTitles verifies that a list's context menu leaves out what the list holds from all but its
// first "New" item, since it can only add to that list, while the menu bar keeps the full titles.
func TestContextMenusUseShortTitles(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	lists := &listsForTest{}
	entity := gurps.NewEntity()
	titles := func(provider interface{ ContextMenuItems() []ContextMenuItem }) map[int]string {
		m := make(map[int]string)
		for _, one := range provider.ContextMenuItems() {
			m[one.ID] = one.Title
		}
		return m
	}
	traits := titles(NewTraitsProvider(lists, false))
	c.Equal("New Trait", traits[NewTraitItemID])
	c.Equal("New Container", traits[NewTraitContainerItemID])
	c.Equal("New Choice", traits[NewTraitChoiceContainerItemID])
	skills := titles(NewSkillsProvider(entity, false))
	c.Equal("New Skill", skills[NewSkillItemID])
	c.Equal("New Container", skills[NewSkillContainerItemID])
	c.Equal("New Technique", skills[NewTechniqueItemID], "a distinct kind of item keeps its name")
	spells := titles(NewSpellsProvider(entity, false))
	c.Equal("New Container", spells[NewSpellContainerItemID])
	c.Equal("New Choice", spells[NewSpellChoiceContainerItemID])
	c.Equal("New Ritual Magic Spell", spells[NewRitualMagicSpellItemID])
	notes := titles(NewNotesProvider(lists, false))
	c.Equal("New Note", notes[NewNoteItemID])
	c.Equal("New Container", notes[NewNoteContainerItemID])
	c.Equal("New Trait Container", newTraitContainerAction.Title, "the menu bar keeps the full title")
}
