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
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// keyBindingRegistrars names the functions in actions.go that take a key binding ID as their first argument. Every
// registration helper must be listed here, or TestKeyBindingIDsAreUnique cannot see the IDs it registers; that test
// fails on a binding registered but not found in the source, which catches a helper missing from this list.
var keyBindingRegistrars = map[string]bool{
	"registerKeyBindableAction":           true,
	"registerFocusAction":                 true,
	"registerFocusActionWithContextTitle": true,
	"registerLibraryAction":               true,
	"registerSheetAction":                 true,
	"registerTitledKeyBindableAction":     true,
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
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
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
		opened := uxtest.OpenedByAction(t, screen, one.action)
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
		uxtest.CloseEditorWithoutPrompt(t, screen, opened)
	}
}

// The per-sheet settings actions must be disabled while no character sheet is active and, once one is, open their
// settings editor for that sheet rather than for the defaults.
func TestPerSheetSettingsActionsFollowTheActiveSheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
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

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
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
		opened := uxtest.OpenedByAction(t, screen, one.action)
		var owner any
		screen.Do(func() { owner = one.owner(opened) })
		c.Equal(any(sheet), owner, "%s must open its settings editor for the active sheet", one.action.Title)
		uxtest.CloseEditorWithoutPrompt(t, screen, opened)
	}
}

// Undo and Redo must take their enabled state and titles from the active window's undo manager and perform the
// matching operation on it. Both are built by one helper handed the undo manager's methods, so this is what catches the
// two being wired to each other's.
func TestUndoAndRedoActionsDriveTheActiveWindowUndoManager(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
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

// isTableDockable reports whether d is a library table holding rows of type T.
func isTableDockable[T gurps.Node[T]](d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*TableDockable[T])
	return ok
}

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

// TestContextMenusUseShortTitles verifies that a list's context menu leaves what the list holds out of its "New" items
// for containers and choices, since it can only add to that list, while the menu bar keeps the full titles.
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

// TestBrowserRequestsAreRecordedHeadless verifies that the places GCS hands a URL to the browser go through
// unison.OpenBrowser, so that a headless session records the request for OpenedURLs() rather than launching the
// host's browser from under a test, and that no error dialog results from the recorded request.
func TestBrowserRequestsAreRecordedHeadless(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")

	var cancel bool
	c.True(screen.Do(func() { cancel = OpenPageReference("https://example.com/ref", "", nil) }))
	c.False(cancel, "a web page reference never asks to cancel further processing")
	c.Equal([]string{"https://example.com/ref"}, screen.OpenedURLs(),
		"a page reference that is a URL is handed to the browser through unison")

	c.True(screen.Do(func() { showWebPage("https://example.com/help") }))
	c.Equal([]string{"https://example.com/help"}, screen.OpenedURLs(),
		"the Help menu's web pages are handed to the browser through unison")
	c.Nil(screen.OpenedURLs(), "and each request is handed back only once")
}

// TestCreatingModifierChoices verifies that the "New ... Modifier Choice" commands are offered by a modifier library
// and by an editor's modifier table, in the menus as well, and that what they create is a mandatory choice.
func TestCreatingModifierChoices(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, nil)
	c.True(library.CanPerformCmd(nil, NewTraitModifierChoiceItemID))
	provider, ok := library.provider.(*modifiersProvider[*gurps.TraitModifier])
	c.True(ok)
	items := provider.ContextMenuItems()
	titles := make([]string, 0, len(items))
	for _, item := range items {
		titles = append(titles, item.Title)
	}
	c.True(slices.Contains(titles, "New Trait Modifier Group"), "a modifier container is now called a group")
	c.True(slices.Contains(titles, "New Trait Modifier Choice"))

	var edited *gurps.TraitModifier
	provider.edit = func(_ Rebuildable, item *gurps.TraitModifier) { edited = item }
	library.PerformCmd(nil, NewTraitModifierChoiceItemID)
	c.Equal([]*gurps.TraitModifier{edited}, provider.RootData(), "the command adds the choice and hands it to its editor")
	c.True(gurps.IsMandatoryModifierChoice(edited))

	e, _, _ := newEquipmentEditorWithModifiers(t)
	c.True(e.CanPerformCmd(nil, NewEquipmentModifierChoiceItemID))
}
