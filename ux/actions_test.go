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
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// keyBindingRegistrars names the functions in actions.go that take a key binding ID as their first argument. Every
// registration helper must be listed here, or TestKeyBindingIDsAreUnique cannot see the IDs it registers; that test
// fails on a binding registered but not found in the source, which catches a helper missing from this list.
var keyBindingRegistrars = map[string]bool{
	"registerKeyBindableAction": true,
	"registerFocusAction":       true,
	"registerLibraryAction":     true,
	"registerSheetAction":       true,
}

// gurps.RegisterKeyBinding silently ignores a duplicate ID, so an action that reuses one is never added to the binding
// set: it can't be seen or assigned a binding in the Menu Keys settings, KeyBindings.MakeCurrent() never updates it,
// and any binding the user assigns to that ID applies only to the action that claimed it first.
func TestKeyBindingIDsAreUnique(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	ids := keyBindingIDsInSource(c)
	c.True(len(ids) > 1, "actions.go must contain key binding registrations")
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		c.False(seen[id], "key binding ID %q is registered more than once; only the first registration takes effect", id)
		seen[id] = true
	}
	registered := make(map[string]bool, len(ids))
	for _, one := range gurps.CurrentBindings() {
		registered[one.ID] = true
	}
	for id := range seen {
		c.True(registered[id], "key binding ID %q was never added to the binding set", id)
	}
	for id := range registered {
		c.True(seen[id], "key binding ID %q is registered by a function that keyBindingRegistrars does not name", id)
	}
}

// The two similarly-named equipment library actions must each have their own key binding; they were once registered
// with the same ID.
func TestEquipmentLibraryActionsAreSeparatelyBindable(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	byAction := make(map[*unison.Action]string)
	for _, one := range gurps.CurrentBindings() {
		byAction[one.Action] = one.ID
	}
	equipmentID, ok := byAction[newEquipmentLibraryAction]
	c.True(ok, "New Equipment Library must have a key binding")
	modifiersID, ok := byAction[newEquipmentModifiersLibraryAction]
	c.True(ok, "New Equipment Modifiers Library must have a key binding")
	c.NotEqual(equipmentID, modifiersID, "the two equipment library actions must not share a key binding ID")
}

// keyBindingIDsInSource returns the key binding IDs that actions.go passes as string literals to the functions named
// in keyBindingRegistrars and to gurps.RegisterKeyBinding.
func keyBindingIDsInSource(c check.Checker) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "actions.go", nil, 0)
	c.NoError(err, "actions.go must be parsable")
	var ids []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			if !keyBindingRegistrars[fn.Name] {
				return true
			}
		case *ast.SelectorExpr:
			if fn.Sel.Name != "RegisterKeyBinding" {
				return true
			}
		default:
			return true
		}
		if lit, ok2 := call.Args[0].(*ast.BasicLit); ok2 && lit.Kind == token.STRING {
			var id string
			if id, err = strconv.Unquote(lit.Value); err != nil {
				c.NoError(err, "key binding ID must be a valid string literal")
			} else {
				ids = append(ids, id)
			}
		}
		return true
	})
	return ids
}

// Each "New … Library" action must open a new, empty library dockable of its own kind under the expected name. One
// helper builds all seven from a file name and a constructor, so a slip there would hand a menu item another kind of
// library.
func TestLibraryActionsOpenAnEmptyLibraryOfTheirKind(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	for _, one := range []struct {
		action *unison.Action
		title  string
		isKind func(unison.Dockable) bool
	}{
		{newEquipmentLibraryAction, "Equipment", isTableDockable[*gurps.Equipment]},
		{newEquipmentModifiersLibraryAction, "Equipment Modifiers", isTableDockable[*gurps.EquipmentModifier]},
		{newNotesLibraryAction, "Notes", isTableDockable[*gurps.Note]},
		{newSkillsLibraryAction, "Skills", isTableDockable[*gurps.Skill]},
		{newSpellsLibraryAction, "Spells", isTableDockable[*gurps.Spell]},
		{newTraitModifiersLibraryAction, "Trait Modifiers", isTableDockable[*gurps.TraitModifier]},
		{newTraitsLibraryAction, "Traits", isTableDockable[*gurps.Trait]},
	} {
		opened := openedByAction(t, screen, one.action)
		var title string
		var kind, modified bool
		screen.Do(func() {
			title = opened.Title()
			kind = one.isKind(opened)
			modified = opened.Modified()
		})
		c.Equal(one.title, title, "%s must open a library named for its kind", one.action.Title)
		c.True(kind, "%s must open a %s library", one.action.Title, one.title)
		c.False(modified, "%s must open an unmodified library", one.action.Title)
		closeEditorWithoutPrompt(t, screen, opened)
	}
}

// The per-sheet settings actions must be disabled while no character sheet is active and, once one is, open their
// settings editor for that sheet rather than for the defaults.
func TestPerSheetSettingsActionsFollowTheActiveSheet(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	actions := []*unison.Action{perSheetAttributeSettingsAction, perSheetBodyTypeSettingsAction, perSheetSettingsAction}
	enabled := make([]bool, len(actions))
	screen.Do(func() {
		for i, a := range actions {
			enabled[i] = a.Enabled(nil)
		}
	})
	for i, a := range actions {
		c.False(enabled[i], "%s must be disabled while no sheet is active", a.Title)
	}

	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var active *Sheet
	screen.Do(func() {
		active = ActiveSheet()
		for i, a := range actions {
			enabled[i] = a.Enabled(nil)
		}
	})
	c.Equal(sheet, active, "the new sheet must be the active one")
	for i, a := range actions {
		c.True(enabled[i], "%s must be enabled while a sheet is active", a.Title)
	}

	for _, one := range []struct {
		action *unison.Action
		owner  func(unison.Dockable) any
	}{
		{perSheetAttributeSettingsAction, func(d unison.Dockable) any {
			if s, ok2 := d.AsPanel().Self.(*attributeSettingsDockable); ok2 {
				return s.owner
			}
			return nil
		}},
		{perSheetBodyTypeSettingsAction, func(d unison.Dockable) any {
			if s, ok2 := d.AsPanel().Self.(*bodySettingsDockable); ok2 {
				return s.owner
			}
			return nil
		}},
		{perSheetSettingsAction, func(d unison.Dockable) any {
			if s, ok2 := d.AsPanel().Self.(*sheetSettingsDockable); ok2 {
				return s.owner
			}
			return nil
		}},
	} {
		opened := openedByAction(t, screen, one.action)
		var owner any
		screen.Do(func() { owner = one.owner(opened) })
		c.Equal(any(sheet), owner, "%s must open its settings editor for the active sheet", one.action.Title)
		closeEditorWithoutPrompt(t, screen, opened)
	}
}

// Undo and Redo must take their enabled state and titles from the active window's undo manager and perform the
// matching operation on it. Both are built by one helper handed the undo manager's methods, so this is what catches the
// two being wired to each other's.
func TestUndoAndRedoActionsDriveTheActiveWindowUndoManager(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	type state struct {
		undoTitle, redoTitle     string
		undoEnabled, redoEnabled bool
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.undoEnabled = undoAction.Enabled(nil)
			s.undoTitle = undoAction.Title
			s.redoEnabled = redoAction.Enabled(nil)
			s.redoTitle = redoAction.Title
		})
		return s
	}
	c.Equal(state{undoTitle: unison.CannotUndoTitle(), redoTitle: unison.CannotRedoTitle()}, current(),
		"nothing can be undone or redone in a fresh sheet")

	var undone, redone int
	screen.Do(func() {
		sheet.UndoManager().Add(&unison.UndoEdit[int]{
			ID:       unison.NextUndoID(),
			EditName: "Probe",
			UndoFunc: func(*unison.UndoEdit[int]) { undone++ },
			RedoFunc: func(*unison.UndoEdit[int]) { redone++ },
		})
	})
	c.Equal(state{undoEnabled: true, undoTitle: "Undo Probe", redoTitle: unison.CannotRedoTitle()}, current(),
		"an edit must make Undo available under the edit's name")

	screen.Do(func() { undoAction.Execute(nil) })
	c.Equal(1, undone, "Undo must undo the edit")
	c.Equal(0, redone, "Undo must not redo the edit")
	c.Equal(state{undoTitle: unison.CannotUndoTitle(), redoEnabled: true, redoTitle: "Redo Probe"}, current(),
		"undoing must make Redo available under the edit's name")

	screen.Do(func() { redoAction.Execute(nil) })
	c.Equal(1, undone, "Redo must not undo the edit again")
	c.Equal(1, redone, "Redo must redo the edit")
	c.Equal(state{undoEnabled: true, undoTitle: "Undo Probe", redoTitle: unison.CannotRedoTitle()}, current(),
		"redoing must make Undo available again")
}

// openedByAction executes action on the UI thread and returns the one dockable it opened, failing the test if it
// opened any other number of them.
func openedByAction(t *testing.T, screen *unison.HeadlessScreen, action *unison.Action) interface {
	unison.Dockable
	unison.TabCloser
} {
	t.Helper()
	var opened []unison.Dockable
	screen.Do(func() {
		before := make(map[unison.Dockable]bool)
		for _, d := range AllDockables() {
			before[d] = true
		}
		action.Execute(nil)
		for _, d := range AllDockables() {
			if !before[d] {
				opened = append(opened, d)
			}
		}
	})
	if len(opened) != 1 {
		t.Fatalf("%s opened %d dockables; expected exactly one", action.Title, len(opened))
	}
	d, ok := opened[0].AsPanel().Self.(interface {
		unison.Dockable
		unison.TabCloser
	})
	if !ok {
		t.Fatalf("%s opened a %T, which cannot be closed as a tab", action.Title, opened[0].AsPanel().Self)
	}
	return d
}

// isTableDockable reports whether d is a library table holding rows of type T.
func isTableDockable[T gurps.Node[T]](d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*TableDockable[T])
	return ok
}
