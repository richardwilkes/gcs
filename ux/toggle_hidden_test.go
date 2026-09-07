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

// checkToggleUndoRedo verifies what must hold once something has been toggled inside an editor, no matter what the
// toggle was or how it was made: the editor must have unsaved changes, the table must be able to reach the editor's
// undo manager, the Edit menu must name the change, undo must put the old state back and leave the editor with no
// unsaved changes, and redo must make the change again. verb names the change for the failure messages ("hide",
// "click", "toggle"), while checkToggled and checkRestored make whatever assertions the caller needs about the state
// itself, the first after the redo and the second after the undo.
func checkToggleUndoRedo(c check.Checker, table unison.Paneler, isModified func() bool, undoName, verb string,
	checkToggled, checkRestored func(),
) {
	c.Helper()
	c.True(isModified(), "the %s must give the editor unsaved changes", verb)

	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the editor's undo manager")
	c.True(mgr.CanUndo(), "the %s must be undoable", verb)
	c.Equal(fmt.Sprintf(i18n.Text("Undo %s"), undoName), mgr.UndoTitle(),
		"the Edit menu must name the change that was made")
	mgr.Undo()
	checkRestored()
	c.False(isModified(), "undo must leave the editor with no unsaved changes")

	c.True(mgr.CanRedo(), "an undone %s must be redoable", verb)
	mgr.Redo()
	checkToggled()
	c.True(isModified(), "redo must give the editor unsaved changes again")
}

// newTraitEditorWithWeapons returns a trait editor for a trait carrying one melee weapon and one ranged weapon, so that
// a test can check that hiding one of them leaves the other alone.
func newTraitEditorWithWeapons(t *testing.T) *editor[*gurps.Trait, *gurps.TraitEditData] {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Claws"
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true), gurps.NewWeapon(trait, false)}
	e, _ := buildEditorContent(sheet, trait, initTraitEditor)
	return e
}

// forEachWeaponKind runs f as a subtest for the melee weapons of an editor and again for the ranged ones, since the two
// weapon panels differ only in which kind of weapon they hold.
func forEachWeaponKind(t *testing.T, f func(t *testing.T, melee bool)) {
	t.Helper()
	for _, melee := range []bool{true, false} {
		name := "ranged"
		if melee {
			name = "melee"
		}
		t.Run(name, func(t *testing.T) { f(t, melee) })
	}
}

// weaponsTableInEditor returns the editor's table of melee or ranged weapons along with the single weapon of that kind
// that newTraitEditorWithWeapons gave the trait, verifying along the way that the panel exists, that the table shows
// just that one weapon and that the weapon starts out visible.
func weaponsTableInEditor(t *testing.T, e *editor[*gurps.Trait, *gurps.TraitEditData], melee bool) (
	*unison.Table[*Node[*gurps.Weapon]], *gurps.Weapon,
) {
	t.Helper()
	c := check.New(t)
	panel := e.rangedWeapons
	if melee {
		panel = e.meleeWeapons
	}
	c.NotNil(panel, "expected a weapons panel in the trait editor")
	rows := panel.table.RootRows()
	c.Equal(1, len(rows), "the weapons table must show the one weapon of its kind")
	weapon := rows[0].Data()
	c.False(weapon.Hide, "the weapon starts out visible")
	return panel.table, weapon
}

// checkWeaponHideUndoRedo verifies what must hold once a weapon has been hidden inside an editor, whether by the Toggle
// State command or by a click on its checkmark cell: nothing but the editor's copy of that one weapon changed, and the
// hide can be taken back and put back again. verb names the change for the failure messages.
func checkWeaponHideUndoRedo(c check.Checker, e *editor[*gurps.Trait, *gurps.TraitEditData],
	table *unison.Table[*Node[*gurps.Weapon]], weapon *gurps.Weapon, verb string,
) {
	c.Helper()
	for _, other := range e.editorData.Weapons {
		if other != weapon {
			c.False(other.Hide, "the editor's other weapon must be left alone")
		}
	}
	for _, own := range e.target.Weapons {
		c.False(own.Hide, "the target's own weapons must not be touched until the edit is applied")
	}
	checkToggleUndoRedo(c, table, e.isModified, i18n.Text("Toggle Hidden"), verb,
		func() { c.True(weapon.Hide, "redo must hide the weapon again") },
		func() { c.False(weapon.Hide, "undo must show the weapon again") })
}

// TestToggleStateHidesWeaponsInsideEditors verifies that Toggle State reaches the weapon rows of a detail editor, where
// the checkmark column means "hidden". As with the modifier tables, the only way to hide a weapon used to be clicking
// its cell one row at a time. Only the editor's copy of the data is touched, so nothing reaches the item being edited
// until Apply, and the whole selection is flipped as a single undoable edit.
func TestToggleStateHidesWeaponsInsideEditors(t *testing.T) {
	forEachWeaponKind(t, func(t *testing.T, melee bool) {
		c := check.New(t)
		e := newTraitEditorWithWeapons(t)
		table, weapon := weaponsTableInEditor(t, e, melee)

		table.SelectByIndex(0)
		c.True(table.CanPerformCmd(nil, ToggleStateItemID), "a selected weapon must offer Toggle State")
		table.PerformCmd(nil, ToggleStateItemID)
		c.True(weapon.Hide, "the command must hide the selected weapon in the editor's copy of the data")

		checkWeaponHideUndoRedo(c, e, table, weapon, "hide")
	})
}

// TestWeaponHideCheckmarkClickIsUndoable verifies that clicking a weapon's Hide checkmark cell inside an editor hides
// the weapon in the editor's copy and that the change can be taken back and put back again. The click and the command
// now share adjustTargets and snapshotList.apply, so the click path -- and the redo that goes with it -- needs a test
// of its own rather than being taken on faith from the command's.
func TestWeaponHideCheckmarkClickIsUndoable(t *testing.T) {
	forEachWeaponKind(t, func(t *testing.T, melee bool) {
		c := check.New(t)
		e := newTraitEditorWithWeapons(t)
		table, weapon := weaponsTableInEditor(t, e, melee)
		c.Equal(gurps.WeaponHideColumn, table.Columns[0].ID, "the Hide column must come first in an editor")

		label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false, false).(*unison.Label)
		c.True(ok, "the Hide cell must be a label")
		// The cell has to be part of the editor's panel tree for the undo manager and the owning editor to be found, so
		// the test attaches it itself, standing in for what the table does for the duration of a real click. A cell is
		// never one of the table's children beyond the event it is handling, so it is taken back off again as soon as
		// the click has been dispatched.
		table.AddChild(label)
		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
		label.RemoveFromParent()
		c.True(weapon.Hide, "clicking the cell must hide the weapon in the editor's copy of the data")

		checkWeaponHideUndoRedo(c, e, table, weapon, "click")
	})
}

// TestToggleStateNotOfferedOnSheetWeaponLists verifies that the command stays off the sheet's own weapon lists. Those
// lists carry no Hide column -- a weapon is listed there only when it isn't hidden -- so the command is installed on
// the editor tables alone rather than on every table that happens to hold weapons.
func TestToggleStateNotOfferedOnSheetWeaponLists(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)

	table := sheet.MeleeWeapons.Table
	c.Equal(1, len(table.RootRows()), "the sheet's melee weapons list must show the trait's weapon")
	table.SelectByIndex(0)
	c.False(table.CanPerformCmd(nil, ToggleStateItemID),
		"the sheet's weapon lists have nothing to toggle, so the command must not be offered there")
}
