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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// weaponUsages returns the usages of the weapons, in order.
func weaponUsages(list []*gurps.Weapon) []string {
	usages := make([]string, len(list))
	for i, w := range list {
		usages[i] = w.Usage
	}
	return usages
}

// TestWeaponsPanelKeepsTheKindsInterleaved verifies that putting one kind of weapon back into an item's list, which
// holds both kinds in whatever order the user gave them, leaves the other kind where it was. Moving a melee weapon and
// undoing the move is a net no-op, so the list must come back as it was and the editor must not count as modified.
func TestWeaponsPanelKeepsTheKindsInterleaved(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Claws"
	for i, usage := range []string{"Thrust", "Thrown", "Swung"} {
		w := gurps.NewWeapon(trait, i != 1)
		w.Usage = usage
		trait.Weapons = append(trait.Weapons, w)
	}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	var panel *weaponsPanel
	for _, one := range uxtest.PanelsOfType[*weaponsPanel](content) {
		if one.melee {
			panel = one
		}
	}
	if panel == nil {
		t.Fatal("the trait editor must hold a melee weapons panel")
	}
	mgr := unison.UndoManagerFor(panel.table)
	c.True(mgr == e.undoMgr, "the weapons table must use the editor's undo manager")
	c.Equal([]string{"Thrust", "Thrown", "Swung"}, weaponUsages(e.editorData.Weapons), "precondition")
	c.False(e.isModified(), "precondition: the editor starts out unmodified")

	panel.table.SetSelectionMap(map[tid.TID]bool{e.editorData.Weapons[0].ID(): true})
	MoveSelection(panel.table, MoveDown)
	c.Equal([]string{"Swung", "Thrown", "Thrust"}, weaponUsages(e.editorData.Weapons),
		"the move must swap the melee weapons and leave the ranged one in place")
	c.True(e.isModified(), "the move must count as a modification")

	mgr.Undo()
	c.Equal([]string{"Thrust", "Thrown", "Swung"}, weaponUsages(e.editorData.Weapons),
		"undoing the move must put the list back as it was")
	c.False(e.isModified(), "a move and its undo must leave the editor unmodified")
}

// TestWeaponsPanelKeepsAModifiersWeaponsHeldByIt verifies that the weapons a structural undo gives back to the weapons
// panel of a modifier's editor still belong to the trait the modifier modifies and are still held by the modifier,
// through which they resolve their nameable codes.
func TestWeaponsPanelKeepsAModifiersWeaponsHeldByIt(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	mod := gurps.NewTraitModifier(entity, nil, false)
	mod.Weapons = []*gurps.Weapon{gurps.NewWeapon(nil, true)}
	trait := gurps.NewTrait(entity, nil, false)
	trait.SetModifiers([]*gurps.TraitModifier{mod})
	e, _ := buildEditorContent(sheet, mod, initTraitModifierEditor)
	if e.meleeWeapons == nil {
		t.Fatal("the modifier's editor must hold a melee weapons panel")
	}
	table := e.meleeWeapons.table
	table.SetSelectionMap(map[tid.TID]bool{e.editorData.Weapons[0].ID(): true})
	DuplicateSelection(table)
	c.Equal(2, len(e.editorData.Weapons), "precondition: the weapon is duplicated")
	unison.UndoManagerFor(table).Undo()
	c.Equal(1, len(e.editorData.Weapons))
	restored := e.editorData.Weapons[0]
	c.True(restored.Owner == gurps.WeaponOwner(trait), "the restored weapon belongs to the modifier's trait")
	c.True(restored.HoldingModifier() == gurps.WeaponHoldingModifier(mod), "and is held by the modifier")
}
