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

// TestOpeningASheetWeaponOpensItsEditorInsideItsOwners verifies that opening a weapon from a sheet's weapon list opens
// the editor of the trait that owns it and, inside that, the editor of the trait editor's own copy of the weapon. The
// editor stages its data with Copy-mode clones, which keep the weapon's ID, so the copy is found by that ID.
func TestOpeningASheetWeaponOpensItsEditorInsideItsOwners(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var weapon *gurps.Weapon
	screen.Do(func() {
		entity := sheet.Entity()
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = "Claws"
		weapon = gurps.NewWeapon(trait, true)
		weapon.Usage = "Thrust"
		trait.Weapons = []*gurps.Weapon{weapon}
		entity.Traits = []*gurps.Trait{trait}
		sheet.Rebuild(true)
		table := sheet.MeleeWeapons.Table
		table.SetSelectionMap(map[tid.TID]bool{weapon.ID(): true})
		table.PerformCmd(nil, OpenEditorItemID)
	})
	traitEditor := uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return isEditor
	})
	weaponEditor := uxtest.SoleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Weapon, *gurps.Weapon])
		return isEditor
	})
	var editedWeapon, stagedWeapon *gurps.Weapon
	var owner Rebuildable
	screen.Do(func() {
		editedWeapon = weaponEditor.target
		stagedWeapon = traitEditor.editorData.Weapons[0]
		owner = weaponEditor.owner
	})
	c.True(editedWeapon == stagedWeapon, "the weapon's editor must edit the trait editor's copy of the weapon")
	c.True(editedWeapon != weapon, "rather than the weapon on the sheet")
	c.True(owner == Rebuildable(traitEditor), "and belong to the trait's editor")
}
