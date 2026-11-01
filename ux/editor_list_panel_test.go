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
	"github.com/richardwilkes/unison"
)

// Each "new item" command must land in the right list of the editor's data and on the right table. The list panels
// install these handlers themselves, so one that failed to would leave a menu item that does nothing.
func TestEditorNewItemCommandsAddToTheirPanels(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	te, content := buildEditorContent(sheet, trait, initTraitEditor)
	traitModifiers, ok := uxtest.FirstPanelOfType[*traitModifiersPanel](content)
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
	equipmentModifiers, ok := uxtest.FirstPanelOfType[*equipmentModifiersPanel](content)
	c.True(ok, "expected an equipment modifiers panel in the equipment editor")
	ee.AsPanel().PerformCmd(nil, NewEquipmentModifierItemID)
	ee.AsPanel().PerformCmd(nil, NewEquipmentContainerModifierItemID)
	c.Equal(2, len(ee.editorData.Modifiers), "both commands must add a modifier to the editor's copy")
	c.False(ee.editorData.Modifiers[0].Container(), "the plain command must add a non-container")
	c.True(ee.editorData.Modifiers[1].Container(), "the container command must add a container")
	c.Equal(2, equipmentModifiers.table.RootRowCount(), "both modifiers must be rows in the panel's table")
	c.Equal(0, len(equipment.Modifiers), "the equipment itself must not change until the editor is applied")
}

// TestDroppedModifierLegacyReplacementsJoinTheEditorsData verifies that the replacements an older version saved on a
// modifier dropped into an item editor's list of modifiers go into the editor's data, which keeps its own value for a
// code both have, leaving the item alone until the editor is applied, that undoing the drop takes them back out and
// redoing it puts them back, and that applying the editor gives them to the item.
func TestDroppedModifierLegacyReplacementsJoinTheEditorsData(t *testing.T) {
	for _, tc := range []struct {
		name string
		// drop opens an editor for an item on the sheet whose substitutions give Foo the value Bar, drops a modifier
		// carrying the legacy replacements into the group of modifiers it holds, and returns the item's replacements,
		// the editor data's, the editor's undo manager and a function that applies the editor.
		drop func(t *testing.T, sheet *Sheet, legacy map[string]string) (item, data *map[string]string,
			undo *unison.UndoManager, apply func())
	}{
		{
			name: "trait",
			drop: func(t *testing.T, sheet *Sheet, legacy map[string]string) (item, data *map[string]string,
				undo *unison.UndoManager, apply func(),
			) {
				trait := gurps.NewTrait(sheet.Entity(), nil, false)
				trait.Replacements = map[string]string{"Foo": "Bar"}
				trait.SetModifiers([]*gurps.TraitModifier{gurps.NewTraitModifier(sheet.Entity(), nil, true)})
				e, content := buildEditorContent(sheet, trait, initTraitEditor)
				panel, ok := uxtest.FirstPanelOfType[*traitModifiersPanel](content)
				if !ok {
					t.Fatal("expected a trait modifiers panel in the trait editor")
				}
				dropped := gurps.NewTraitModifier(nil, nil, false)
				dropped.Name = "@Material@ Coating"
				dropped.Replacements = legacy
				from := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
				simulateCopyDrop(from, panel.table, []*Node[*gurps.TraitModifier]{NewNode(from, nil, dropped, false)},
					e.editorData.Modifiers[0].ID())
				return &trait.Replacements, &e.editorData.Replacements, e.undoMgr, e.applyEdits
			},
		},
		{
			name: "equipment",
			drop: func(t *testing.T, sheet *Sheet, legacy map[string]string) (item, data *map[string]string,
				undo *unison.UndoManager, apply func(),
			) {
				eqp := gurps.NewEquipment(sheet.Entity(), nil, false)
				eqp.Replacements = map[string]string{"Foo": "Bar"}
				eqp.SetModifiers([]*gurps.EquipmentModifier{gurps.NewEquipmentModifier(sheet.Entity(), nil, true)})
				e, content := buildEditorContent(sheet, eqp, initEquipmentEditor(true))
				panel, ok := uxtest.FirstPanelOfType[*equipmentModifiersPanel](content)
				if !ok {
					t.Fatal("expected an equipment modifiers panel in the equipment editor")
				}
				dropped := gurps.NewEquipmentModifier(nil, nil, false)
				dropped.Name = "@Material@ Coating"
				dropped.Replacements = legacy
				from := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{})
				simulateCopyDrop(from, panel.table,
					[]*Node[*gurps.EquipmentModifier]{NewNode(from, nil, dropped, false)}, e.editorData.Modifiers[0].ID())
				return &eqp.Replacements, &e.editorData.Replacements, e.undoMgr, e.applyEdits
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			forbidModifierPrompts(t)
			sheet := newTestSheetForTemplate(t)
			// As an older version saved it.
			item, data, undo, apply := tc.drop(t, sheet, map[string]string{"Material": "Steel", "Foo": "Baz"})
			c.Equal(map[string]string{"Foo": "Bar"}, *item, "the item must be left alone until the editor is applied")
			c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel"}, *data,
				"the editor's data must take the dropped modifier's replacements, keeping its own value for Foo")
			undo.Undo()
			c.Equal(map[string]string{"Foo": "Bar"}, *data, "undoing the drop must take them back out")
			undo.Redo()
			c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel"}, *data, "redoing it must put them back")
			apply()
			c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel"}, *item, "applying must give them to the item")
		})
	}
}
