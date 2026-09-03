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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// newTraitEditorWithModifiers returns a trait editor's table of trait modifiers, along with the non-container modifier
// belonging to the trait itself. The trait carries one modifier that can be toggled followed by one container, which
// cannot.
func newTraitEditorWithModifiers(t *testing.T) (*editor[*gurps.Trait, *gurps.TraitEditData],
	*unison.Table[*Node[*gurps.TraitModifier]], *gurps.TraitModifier,
) {
	t.Helper()
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Sharp"
	container := gurps.NewTraitModifier(entity, nil, true)
	container.Name = "Variations"
	trait.Modifiers = []*gurps.TraitModifier{modifier, container}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := findPanel[*traitModifiersPanel](content)
	c.True(ok, "expected a trait modifiers panel in the trait editor")
	return e, panel.table, modifier
}

// newEquipmentEditorWithModifiers does the same for a piece of equipment and its modifiers.
func newEquipmentEditorWithModifiers(t *testing.T) (*editor[*gurps.Equipment, *gurps.EquipmentEditData],
	*unison.Table[*Node[*gurps.EquipmentModifier]], *gurps.EquipmentModifier,
) {
	t.Helper()
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	equipment := gurps.NewEquipment(entity, nil, false)
	equipment.Name = "Sword"
	modifier := gurps.NewEquipmentModifier(entity, nil, false)
	modifier.Name = "Fine"
	container := gurps.NewEquipmentModifier(entity, nil, true)
	container.Name = "Variations"
	equipment.Modifiers = []*gurps.EquipmentModifier{modifier, container}
	e, content := buildEditorContent(sheet, equipment, initEquipmentEditor(true))
	panel, ok := findPanel[*equipmentModifiersPanel](content)
	c.True(ok, "expected an equipment modifiers panel in the equipment editor")
	return e, panel.table, modifier
}

// generalModifier reports the given editor copy through the interface both kinds of modifier implement, so that a
// check can be written once for either.
func generalModifier[T gurps.Node[T]](c check.Checker, node T) gurps.GeneralModifier {
	m, ok := any(node).(gurps.GeneralModifier)
	c.True(ok, "a modifier must provide the general modifier interface")
	return m
}

// checkModifierToggleInEditor drives Toggle State over a modifiers table holding one non-container modifier followed by
// one container, both selected, and verifies that only the editor's copy of the non-container one is flipped and that
// the change can be taken back and put back again. The trait and equipment editors differ only in the type of modifier
// they hold, so the two subtests share this.
func checkModifierToggleInEditor[T gurps.Node[T]](t *testing.T, table *unison.Table[*Node[T]], modifiers []T,
	isModified func() bool, targetModifier gurps.GeneralModifier, undoName string,
) {
	t.Helper()
	c := check.New(t)
	c.Equal(2, len(table.RootRows()), "the modifiers table must show both of the modifiers")
	table.SelectByIndex(0, 1)
	c.True(table.CanPerformCmd(nil, ToggleStateItemID),
		"a selection holding a modifier that can be toggled must offer Toggle State")

	table.PerformCmd(nil, ToggleStateItemID)
	c.False(generalModifier(c, modifiers[0]).Enabled(),
		"the command must turn the selected modifier off in the editor's copy of the data")
	c.True(generalModifier(c, modifiers[1]).Enabled(),
		"a container is always enabled, so the command must leave it alone")
	c.True(targetModifier.Enabled(), "the target's own modifier must not be touched until the edit is applied")
	c.True(isModified(), "toggling a modifier must give the editor unsaved changes")

	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the editor's undo manager")
	c.True(mgr.CanUndo(), "toggling a modifier must be undoable")
	c.Equal(fmt.Sprintf(i18n.Text("Undo %s"), undoName), mgr.UndoTitle(),
		"the Edit menu must name the kind of modifier that was toggled")
	mgr.Undo()
	c.True(generalModifier(c, modifiers[0]).Enabled(), "undo must turn the modifier back on")
	c.False(isModified(), "undo must leave the editor with no unsaved changes")

	c.True(mgr.CanRedo(), "an undone toggle must be redoable")
	mgr.Redo()
	c.False(generalModifier(c, modifiers[0]).Enabled(), "redo must turn the modifier off again")
	c.True(isModified(), "redo must give the editor unsaved changes again")
}

// TestToggleStateFlipsModifiersInsideEditors verifies that Toggle State reaches the modifier rows of a detail editor,
// which is what issue #1074 asked for: until now the only way to turn a modifier on or off was to click its checkmark
// cell one row at a time, since the editor tables installed no handler for the command and so both greyed out the Edit
// menu item and dropped it from their context menus. The whole selection is flipped as a single undoable edit, and
// only the editor's copy of the data is touched, so nothing reaches the item being edited until Apply.
func TestToggleStateFlipsModifiersInsideEditors(t *testing.T) {
	t.Run("trait", func(t *testing.T) {
		e, table, modifier := newTraitEditorWithModifiers(t)
		checkModifierToggleInEditor(t, table, e.editorData.Modifiers, e.isModified, modifier,
			i18n.Text("Toggle Trait Modifier"))
	})

	t.Run("equipment", func(t *testing.T) {
		e, table, modifier := newEquipmentEditorWithModifiers(t)
		checkModifierToggleInEditor(t, table, e.editorData.Modifiers, e.isModified, modifier,
			i18n.Text("Toggle Equipment Modifier"))
	})
}

// TestToggleStateSkipsModifierContainers verifies that the command is offered only when the selection holds something
// it can actually change. A container is always enabled and shows no checkmark cell, so a selection of nothing but
// containers must leave the menu item greyed out rather than registering an edit that changes nothing.
func TestToggleStateSkipsModifierContainers(t *testing.T) {
	c := check.New(t)
	_, table, _ := newTraitEditorWithModifiers(t)
	c.False(table.CanPerformCmd(nil, ToggleStateItemID), "with nothing selected there is nothing to toggle")

	table.SelectByIndex(1)
	c.False(table.CanPerformCmd(nil, ToggleStateItemID),
		"a selection holding only a container must not offer Toggle State")

	table.ClearSelection()
	table.SelectByIndex(0)
	c.True(table.CanPerformCmd(nil, ToggleStateItemID),
		"a selection holding a modifier that can be toggled must offer Toggle State")
}

// checkModifierCheckmarkClickInEditor clicks the checkmark cell of the first modifier row and verifies that the click
// turns the modifier off in the editor's copy and that the change can be taken back and put back again. Both kinds of
// modifier cell go through the same handleCheck case, so the two subtests share this.
func checkModifierCheckmarkClickInEditor[T gurps.Node[T]](t *testing.T, table *unison.Table[*Node[T]], modifiers []T,
	isModified func() bool, enabledColumnID int, undoName string,
) {
	t.Helper()
	c := check.New(t)
	c.Equal(enabledColumnID, table.Columns[0].ID, "the enabled column must come first in an editor")

	label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the enabled cell must be a label")
	// The cell has to be part of the editor's panel tree for the undo manager and the owning editor to be found, so the
	// test attaches it itself, standing in for what the table does for the duration of a real click. A cell is never
	// one of the table's children beyond the event it is handling, so it is taken back off again as soon as the click
	// has been dispatched.
	table.AddChild(label)
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	label.RemoveFromParent()
	c.False(generalModifier(c, modifiers[0]).Enabled(),
		"clicking the cell must turn the modifier off in the editor's copy")
	c.True(isModified(), "clicking the cell must give the editor unsaved changes")

	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the editor's undo manager")
	c.True(mgr.CanUndo(), "clicking the cell must be undoable")
	c.Equal(fmt.Sprintf(i18n.Text("Undo %s"), undoName), mgr.UndoTitle(),
		"the Edit menu must name the kind of modifier that was toggled")
	mgr.Undo()
	c.True(generalModifier(c, modifiers[0]).Enabled(), "undo must turn the modifier back on")
	c.False(isModified(), "undo must leave the editor with no unsaved changes")

	c.True(mgr.CanRedo(), "an undone click must be redoable")
	mgr.Redo()
	c.False(generalModifier(c, modifiers[0]).Enabled(), "redo must turn the modifier off again")
	c.True(isModified(), "redo must give the editor unsaved changes again")
}

// TestModifierCheckmarkClickIsUndoable verifies that clicking a modifier's checkmark cell inside an editor still flips
// the editor's copy and can be taken back and put back again, now that the click is routed through the same machinery
// as the command rather than through an undo payload of its own. Both kinds of modifier cell take that route, so both
// are exercised here.
func TestModifierCheckmarkClickIsUndoable(t *testing.T) {
	t.Run("trait", func(t *testing.T) {
		e, table, _ := newTraitEditorWithModifiers(t)
		checkModifierCheckmarkClickInEditor(t, table, e.editorData.Modifiers, e.isModified,
			gurps.TraitModifierEnabledColumn, i18n.Text("Toggle Trait Modifier"))
	})

	t.Run("equipment", func(t *testing.T) {
		e, table, _ := newEquipmentEditorWithModifiers(t)
		checkModifierCheckmarkClickInEditor(t, table, e.editorData.Modifiers, e.isModified,
			gurps.EquipmentModifierEnabledColumn, i18n.Text("Toggle Equipment Modifier"))
	})
}
