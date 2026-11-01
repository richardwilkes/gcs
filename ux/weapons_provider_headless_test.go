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

// TestOpeningASheetWeaponOfAModifierOpensItsEditorInsideItsModifiers verifies that opening a weapon an item's modifier
// holds from a sheet's weapon list opens the editor of the item that owns it, inside that the editor of the item
// editor's copy of the modifier, and inside that the editor of the modifier editor's copy of the weapon. The copies keep
// the IDs of what they copy, so each is found by its ID.
func TestOpeningASheetWeaponOfAModifierOpensItsEditorInsideItsModifiers(t *testing.T) {
	for _, tc := range []struct {
		name string
		// place puts an item on the entity whose one modifier holds the weapon.
		place func(entity *gurps.Entity, weapon *gurps.Weapon)
		check func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen)
	}{
		{
			name: "trait",
			place: func(entity *gurps.Entity, weapon *gurps.Weapon) {
				mod := gurps.NewTraitModifier(entity, nil, false)
				mod.Weapons = []*gurps.Weapon{weapon}
				trait := gurps.NewTrait(entity, nil, false)
				trait.SetModifiers([]*gurps.TraitModifier{mod})
				entity.Traits = []*gurps.Trait{trait}
			},
			check: modifierWeaponEditorsCheck[*gurps.Trait, *gurps.TraitEditData, *gurps.TraitModifier](
				func(d *gurps.TraitEditData) []*gurps.TraitModifier { return d.Modifiers },
				func(d *gurps.TraitModifierEditData) []*gurps.Weapon { return d.Weapons },
			),
		},
		{
			name: "equipment",
			place: func(entity *gurps.Entity, weapon *gurps.Weapon) {
				mod := gurps.NewEquipmentModifier(entity, nil, false)
				mod.Weapons = []*gurps.Weapon{weapon}
				eqp := gurps.NewEquipment(entity, nil, false)
				eqp.SetModifiers([]*gurps.EquipmentModifier{mod})
				entity.CarriedEquipment = []*gurps.Equipment{eqp}
			},
			check: modifierWeaponEditorsCheck[*gurps.Equipment, *gurps.EquipmentEditData, *gurps.EquipmentModifier](
				func(d *gurps.EquipmentEditData) []*gurps.EquipmentModifier { return d.Modifiers },
				func(d *gurps.EquipmentModifierEditData) []*gurps.Weapon { return d.Weapons },
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
			if !ok {
				t.Fatal("New Character Sheet must open a character sheet")
			}
			screen.Do(func() {
				weapon := gurps.NewWeapon(nil, true)
				weapon.Usage = "Strike"
				tc.place(sheet.Entity(), weapon)
				sheet.Rebuild(true)
				table := sheet.MeleeWeapons.Table
				table.SetSelectionMap(map[tid.TID]bool{weapon.ID(): true})
				c.Equal(1, len(table.SelectedRows(false)), "precondition: the sheet lists the modifier's weapon")
				table.PerformCmd(nil, OpenEditorItemID)
			})
			tc.check(t, c, screen)
		})
	}
}

// modifierWeaponEditorsCheck returns the check of TestOpeningASheetWeaponOfAModifierOpensItsEditorInsideItsModifiers
// for an item of type N with modifiers of type M, which modifiers and weapons take from the data of their editors.
func modifierWeaponEditorsCheck[N gurps.Node[N], DN gurps.EditorData[N], M gurps.Node[M], DM gurps.EditorData[M]](
	modifiers func(DN) []M, weapons func(DM) []*gurps.Weapon,
) func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen) {
	return func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen) {
		itemEditor := soleEditorOfType[*editor[N, DN]](t, screen)
		modEditor := soleEditorOfType[*editor[M, DM]](t, screen)
		weaponEditor := soleEditorOfType[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen)
		screen.Do(func() {
			c.True(modEditor.target == modifiers(itemEditor.editorData)[0],
				"the modifier's editor must edit the item editor's copy of the modifier")
			c.True(modEditor.owner == Rebuildable(itemEditor), "and belong to the item's editor")
			c.True(weaponEditor.target == weapons(modEditor.editorData)[0],
				"the weapon's editor must edit the modifier editor's copy of the weapon")
			c.True(weaponEditor.owner == Rebuildable(modEditor), "and belong to the modifier's editor")
		})
	}
}

// soleEditorOfType returns the one open dockable that is a T, failing the test if there is not exactly one.
func soleEditorOfType[T unison.Dockable](t *testing.T, screen *unison.HeadlessScreen) T {
	t.Helper()
	return uxtest.SoleEditor[T](t, screen, func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(T)
		return ok
	})
}
