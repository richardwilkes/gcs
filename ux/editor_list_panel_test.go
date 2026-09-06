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
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEditorNewItemCommandsAddToTheirPanels verifies that each "new item" command an editor offers for its modifier and
// weapon lists lands in the right list of the editor's data and on the right table. The list panels install these
// handlers themselves, so one that failed to would leave a menu item that does nothing.
func TestEditorNewItemCommandsAddToTheirPanels(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	te, content := buildEditorContent(sheet, trait, initTraitEditor)
	traitModifiers, ok := firstPanelOfType[*traitModifiersPanel](content)
	c.True(ok, "expected a trait modifiers panel in the trait editor")
	te.AsPanel().PerformCmd(nil, NewTraitModifierItemID)
	c.Equal(1, len(te.editorData.Modifiers), "the command must add a modifier to the editor's copy")
	c.False(te.editorData.Modifiers[0].Container(), "the plain command must add a non-container")
	te.AsPanel().PerformCmd(nil, NewTraitContainerModifierItemID)
	c.Equal(2, len(te.editorData.Modifiers), "the container command must add a second modifier")
	c.True(te.editorData.Modifiers[1].Container(), "the container command must add a container")
	c.Equal(2, traitModifiers.table.RootRowCount(), "both modifiers must be rows in the panel's table")
	c.Equal(0, len(trait.Modifiers), "the trait itself must not change until the editor is applied")

	te.AsPanel().PerformCmd(nil, NewMeleeWeaponItemID)
	c.Equal(1, len(te.editorData.Weapons), "the melee command must add a weapon to the editor's copy")
	c.True(te.editorData.Weapons[0].IsMelee(), "the melee command must add a melee weapon")
	te.AsPanel().PerformCmd(nil, NewRangedWeaponItemID)
	c.Equal(2, len(te.editorData.Weapons), "the ranged command must add a second weapon")
	c.False(te.editorData.Weapons[1].IsMelee(), "the ranged command must add a ranged weapon")
	c.Equal(1, te.meleeWeapons.table.RootRowCount(), "only the melee weapon may be in the melee table")
	c.Equal(1, te.rangedWeapons.table.RootRowCount(), "only the ranged weapon may be in the ranged table")
	c.Equal(0, len(trait.Weapons), "the trait itself must not change until the editor is applied")

	equipment := gurps.NewEquipment(entity, nil, false)
	equipment.Name = "Sword"
	ee, content := buildEditorContent(sheet, equipment, initEquipmentEditor(true))
	equipmentModifiers, ok := firstPanelOfType[*equipmentModifiersPanel](content)
	c.True(ok, "expected an equipment modifiers panel in the equipment editor")
	ee.AsPanel().PerformCmd(nil, NewEquipmentModifierItemID)
	ee.AsPanel().PerformCmd(nil, NewEquipmentContainerModifierItemID)
	c.Equal(2, len(ee.editorData.Modifiers), "both commands must add a modifier to the editor's copy")
	c.False(ee.editorData.Modifiers[0].Container(), "the plain command must add a non-container")
	c.True(ee.editorData.Modifiers[1].Container(), "the container command must add a container")
	c.Equal(2, equipmentModifiers.table.RootRowCount(), "both modifiers must be rows in the panel's table")
	c.Equal(0, len(equipment.Modifiers), "the equipment itself must not change until the editor is applied")
}
