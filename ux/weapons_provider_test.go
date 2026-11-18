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
// the editor of the item that owns it and, inside that, the editor of the item editor's own copy of the weapon or, for
// a weapon an item's modifier holds, the editor of the item editor's copy of the modifier and, inside that, the editor
// of the modifier editor's copy of the weapon. The editors stage their data with Copy-mode clones, which keep the IDs
// of what they copy, so each copy is found by its ID.
func TestOpeningASheetWeaponOpensItsEditorInsideItsOwners(t *testing.T) {
	for _, tc := range []struct {
		name string
		// place puts an item on the entity that holds the weapon it returns, itself or through its one modifier.
		place func(entity *gurps.Entity) *gurps.Weapon
		check func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, weapon *gurps.Weapon)
	}{
		{
			name: "trait",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				trait := gurps.NewTrait(entity, nil, false)
				trait.Name = "Claws"
				weapon := gurps.NewWeapon(trait, true)
				weapon.Usage = "Thrust"
				trait.Weapons = []*gurps.Weapon{weapon}
				entity.Traits = []*gurps.Trait{trait}
				return weapon
			},
			check: ownWeaponEditorsCheck[*gurps.Trait](func(d *gurps.TraitEditData) []*gurps.Weapon { return d.Weapons }),
		},
		{
			name: "skill",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				skill := gurps.NewSkill(entity, nil, false)
				weapon := gurps.NewWeapon(skill, true)
				weapon.Usage = "Strike"
				skill.Weapons = []*gurps.Weapon{weapon}
				entity.Skills = []*gurps.Skill{skill}
				return weapon
			},
			check: ownWeaponEditorsCheck[*gurps.Skill](func(d *gurps.SkillEditData) []*gurps.Weapon { return d.Weapons }),
		},
		{
			name: "spell",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				spell := gurps.NewSpell(entity, nil, false)
				weapon := gurps.NewWeapon(spell, true)
				weapon.Usage = "Strike"
				spell.Weapons = []*gurps.Weapon{weapon}
				entity.Spells = []*gurps.Spell{spell}
				return weapon
			},
			check: ownWeaponEditorsCheck[*gurps.Spell](func(d *gurps.SpellEditData) []*gurps.Weapon { return d.Weapons }),
		},
		{
			// The sheet lists the weapons of equipment that isn't carried only when it shows them all, and they are
			// looked for on the other equipment's list once the carried equipment's list turns out not to hold them.
			name: "other equipment",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				entity.SheetSettings.ShowAllWeapons = true
				eqp := gurps.NewEquipment(entity, nil, false)
				weapon := gurps.NewWeapon(eqp, true)
				weapon.Usage = "Strike"
				eqp.Weapons = []*gurps.Weapon{weapon}
				entity.OtherEquipment = []*gurps.Equipment{eqp}
				return weapon
			},
			check: ownWeaponEditorsCheck[*gurps.Equipment](func(d *gurps.EquipmentEditData) []*gurps.Weapon { return d.Weapons }),
		},
		{
			name: "trait modifier",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				weapon := gurps.NewWeapon(nil, true)
				weapon.Usage = "Strike"
				mod := gurps.NewTraitModifier(entity, nil, false)
				mod.Weapons = []*gurps.Weapon{weapon}
				trait := gurps.NewTrait(entity, nil, false)
				trait.SetModifiers([]*gurps.TraitModifier{mod})
				entity.Traits = []*gurps.Trait{trait}
				return weapon
			},
			check: modifierWeaponEditorsCheck[*gurps.Trait, *gurps.TraitEditData, *gurps.TraitModifier](
				func(d *gurps.TraitEditData) []*gurps.TraitModifier { return d.Modifiers },
				func(d *gurps.TraitModifierEditData) []*gurps.Weapon { return d.Weapons },
			),
		},
		{
			name: "equipment modifier",
			place: func(entity *gurps.Entity) *gurps.Weapon {
				weapon := gurps.NewWeapon(nil, true)
				weapon.Usage = "Strike"
				mod := gurps.NewEquipmentModifier(entity, nil, false)
				mod.Weapons = []*gurps.Weapon{weapon}
				eqp := gurps.NewEquipment(entity, nil, false)
				eqp.SetModifiers([]*gurps.EquipmentModifier{mod})
				entity.CarriedEquipment = []*gurps.Equipment{eqp}
				return weapon
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
			var weapon *gurps.Weapon
			screen.Do(func() {
				weapon = tc.place(sheet.Entity())
				sheet.Rebuild(true)
				table := sheet.MeleeWeapons.Table
				table.SetSelectionMap(map[tid.TID]bool{weapon.ID(): true})
				c.Equal(1, len(table.SelectedRows(false)), "precondition: the sheet lists the weapon")
				table.PerformCmd(nil, OpenEditorItemID)
			})
			tc.check(t, c, screen, weapon)
		})
	}
}

// ownWeaponEditorsCheck returns the check of TestOpeningASheetWeaponOpensItsEditorInsideItsOwners for a weapon an item
// of type N holds itself, which weapons takes from the data of the item's editor.
func ownWeaponEditorsCheck[N gurps.Node[N], DN gurps.EditorData[N]](
	weapons func(DN) []*gurps.Weapon,
) func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, weapon *gurps.Weapon) {
	return func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, weapon *gurps.Weapon) {
		itemEditor := soleEditorOfType[*editor[N, DN]](t, screen)
		weaponEditor := soleEditorOfType[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen)
		screen.Do(func() {
			c.True(weaponEditor.target == weapons(itemEditor.editorData)[0],
				"the weapon's editor must edit the item editor's copy of the weapon")
			c.True(weaponEditor.target != weapon, "rather than the weapon on the sheet")
			c.True(weaponEditor.owner == Rebuildable(itemEditor), "and belong to the item's editor")
		})
	}
}

// modifierWeaponEditorsCheck returns the check of TestOpeningASheetWeaponOpensItsEditorInsideItsOwners for a weapon
// the modifier of an item of type N holds, with modifiers of type M, which modifiers and weapons take from the data of
// their editors.
func modifierWeaponEditorsCheck[N gurps.Node[N], DN gurps.EditorData[N], M gurps.Node[M], DM gurps.EditorData[M]](
	modifiers func(DN) []M, weapons func(DM) []*gurps.Weapon,
) func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, weapon *gurps.Weapon) {
	return func(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, _ *gurps.Weapon) {
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
