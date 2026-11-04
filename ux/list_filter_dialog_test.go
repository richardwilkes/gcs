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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// dialogFilterPanel returns the filter editor in the dialog.
func dialogFilterPanel(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) *listFilterPanel {
	t.Helper()
	var p *listFilterPanel
	screen.Do(func() { p, _ = uxtest.FirstPanelOfType[*listFilterPanel](dialogWnd.Content()) })
	if p == nil {
		t.Fatal("the dialog holds no filter editor")
	}
	return p
}

// dialogNameField returns the filter editor's name field, failing the test unless it is the one string field in the
// dialog, as it is while no row is open.
func dialogNameField(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) *StringField {
	t.Helper()
	var fields []*StringField
	screen.Do(func() { fields = uxtest.PanelsOfType[*StringField](dialogWnd.Content()) })
	if len(fields) != 1 {
		t.Fatalf("expected the name field to be the dialog's one string field, found %d string fields", len(fields))
	}
	return fields[0]
}

// TestListFilterNewFilterDialogHeadless drives the editor for a new filter inside a headless workspace: OK is held
// back until the name is one no other saved filter bears and the filter has a condition, groups alone not counting,
// the editor opens with room for a few levels of nesting, a condition can be added to the root group, and accepting
// saves the filter and puts it in force.
func TestListFilterNewFilterDialogHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	seedListFilter("Melee", "name", "melee")
	d := openListFilterTraitDockable(t, screen)

	// Put the quick filter to work first, so that the new filter can be seen to combine with it rather than replace
	// it.
	screen.Click(screen.PanelCenter(d.filterField))
	screen.Type("fur")
	c.Equal([]string{"Fur"}, readListFilterState(screen, d).names, "the quick filter has the list to start with")

	newIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle)
	c.Equal(0, newIndex, "New Filter… is the first item")
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}

	nameField := dialogNameField(t, screen, dialogWnd)
	nameInvalid := func() bool {
		var invalid bool
		screen.Do(func() { invalid = nameField.Invalid() })
		return invalid
	}
	var contentWidth, contentHeight float32
	screen.Do(func() {
		rect := dialogWnd.ContentRect()
		contentWidth = rect.Width
		contentHeight = rect.Height
	})
	c.False(okEnabled(), "a filter with no name cannot be accepted")
	c.True(nameInvalid(), "a blank name is marked")
	c.True(contentWidth >= listFilterDialogMinWidth,
		"the editor opens at least %d wide, but is %v", listFilterDialogMinWidth, contentWidth)
	c.True(contentHeight >= listFilterDialogMinHeight,
		"the editor opens at least %d tall, but is %v", listFilterDialogMinHeight, contentHeight)

	// A name another saved filter already bears is refused, whatever its case.
	screen.Click(screen.PanelCenter(nameField))
	screen.Type("melee")
	c.False(okEnabled(), "a name a saved filter already bears cannot be accepted, whatever its case")
	c.True(nameInvalid(), "a name a saved filter already bears is marked")

	// Replace it with a name of its own, which is accepted.
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("Ranged")
	var name string
	screen.Do(func() { name = nameField.Text() })
	c.Equal("Ranged", name, "select-all followed by typing replaces the name")
	c.False(nameInvalid(), "a name of its own is not marked")
	c.False(okEnabled(), "a name no other saved filter bears is not enough while the filter tests nothing")
	uxtest.CaptureScreen(t, c, screen, "list_filter_dialog")

	// Groups holding nothing but other groups still test nothing. The first group chosen for the bare root sets the
	// root's own type, so it takes a second to add one within it.
	p := dialogFilterPanel(t, screen, dialogWnd)
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "Any of Group")() })
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "All of Group")() })
	var sub *gurps.FilterGroup
	screen.Do(func() { sub = asFilterGroup(p.node("r.0")) })
	if sub == nil {
		t.Fatal("a group has been added within the root")
	}
	screen.Do(func() { menuAction(p.treeAddEntries(sub, "r.0"), "Any of Group")() })
	var nested bool
	screen.Do(func() { nested = asFilterGroup(p.node("r.0.0")) != nil })
	c.True(nested, "a group has been nested within it")
	c.False(okEnabled(), "a filter of empty groups cannot be accepted")
	var undoMgr *unison.UndoManager
	screen.Do(func() { undoMgr = unison.UndoManagerFor(p) })
	if undoMgr == nil {
		t.Fatal("the dialog has an undo manager")
	}
	for range 3 {
		screen.Do(undoMgr.Undo)
	}

	// Add a condition to the empty root by clicking its placeholder, which adds one without a menu. With no field
	// chosen before, it starts out testing the first field, open.
	var placeholder *unison.Panel
	screen.Do(func() { placeholder = p.FindRefKey(treeRootPath + ":empty") })
	if placeholder == nil {
		t.Fatal("an empty filter shows its placeholder")
	}
	screen.Click(screen.PanelCenter(placeholder))
	screen.Do(func() {
		c.Equal("r.0", p.open, "the added condition is open")
		c.Equal("r.0:field", focusedRefKey(dialogWnd), "with its field popup focused")
		cond, ok := p.node("r.0").(*gurps.FilterCondition)
		c.True(ok)
		if ok {
			c.Equal(gurps.TraitFilterFields()[0].Key, cond.Field, "testing the first field")
		}
	})
	c.True(okEnabled(), "with a condition to test, the filter can be accepted")

	screen.Click(screen.PanelCenter(okButton))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the editor has been dismissed")

	saved := savedFilters()
	c.Equal(2, len(saved), "the new filter must have been saved alongside the one that was already there")
	if len(saved) != 2 {
		return
	}
	c.Equal("Ranged", saved[1].Name, "the filters are kept sorted by name, so the new one comes second")
	c.Equal(1, len(saved[1].Root.Children), "the condition added in the editor must have been kept")

	state := readListFilterState(screen, d)
	c.True(saved[1] == state.selected, "the new filter itself must be the one in force")
	c.Equal("Ranged", state.popupText, "and the popup must show it")
	c.Equal("fur", state.fieldText, "the quick filter's text must have been kept, since it still applies")
	c.True(state.fieldEnabled, "and its field must still be usable")
	c.Equal([]string{"Fur"}, state.names, "the new filter's lone condition compares against nothing, so every trait "+
		"passes it, and the quick filter narrows those to the one it accepts")
	c.True(state.filtered, "the two together must be driving the rows")

	editIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, editFilterItemTitle)
	deleteIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, deleteFilterItemTitle)
	c.Equal(1, editIndex, "Edit Filter… is the second item")
	c.Equal(2, deleteIndex, "Delete Filter… is the third")
	var editEnabled, deleteEnabled bool
	screen.Do(func() {
		editEnabled = d.savedFilters.popup.ItemEnabledAt(editIndex)
		deleteEnabled = d.savedFilters.popup.ItemEnabledAt(deleteIndex)
	})
	c.True(editEnabled, "Edit Filter… must be available with a filter in force")
	c.True(deleteEnabled, "Delete Filter… must be available with a filter in force")
}

// TestListFilterEditorKeysHeadless drives the filter editor's keys inside a headless workspace: the key bindings of
// Undo and Redo work within the dialog, which has no menu bar of its own, whether the focus is in the name field, the
// conditions or a button, and undoing the name brings the OK button into line. Escape closes the open row before it
// cancels the dialog, wherever the focus is, and Return accepts it.
func TestListFilterEditorKeysHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	uxtest.SwapForTest(t, &dialogMenuTakesUndoKeys, false)
	seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)
	newIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	cancelButton := dialogButton(t, screen, dialog, unison.ModalResponseCancel)
	nameField := dialogNameField(t, screen, dialogWnd)
	type state struct {
		name     string
		open     string
		children int
		ok       bool
	}
	read := func() state {
		var s state
		screen.Do(func() {
			s = state{
				name:     nameField.Text(),
				open:     p.open,
				children: len(p.filter.Root.Children),
				ok:       okButton.Enabled(),
			}
		})
		return s
	}
	focus := func(panel unison.Paneler) { screen.Do(func() { panel.AsPanel().RequestFocus() }) }
	windows := func() int {
		var n int
		screen.Do(func() { n = len(unison.Windows()) })
		return n
	}
	undo := func() { screen.KeyPress(unison.KeyZ, mod.OSMenuCommand()) }
	redo := func() { screen.KeyPress(unison.KeyY, mod.OSMenuCommand()) }

	screen.Type("Ranged")
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")() })
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "a named filter with a condition")

	focus(nameField)
	undo()
	c.Equal(state{name: "Ranged"}, read(), "Undo in the name field takes the condition back, and OK off with it")
	undo()
	c.Equal(state{}, read(), "and then the name, which leaves OK off")
	redo()
	c.Equal(state{name: "Ranged"}, read(), "Redo puts the name back, with OK still off for want of a condition")
	redo()
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "and then the condition, open")

	focus(okButton)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape on OK closes the open row")
	c.Equal(2, windows(), "and leaves the dialog up")
	focus(cancelButton)
	undo()
	c.Equal(0, read().children, "Undo on Cancel works too")
	var inPanel bool
	screen.Do(func() {
		focused := dialogWnd.Focus()
		inPanel = focused != nil && p.HasInSelfOrDescendants(func(one *unison.Panel) bool { return one == focused })
	})
	c.True(inPanel, "and moves the focus into the rows it rebuilds")
	focus(cancelButton)
	redo()
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "as does Redo, opening the row again")
	focus(okButton)
	undo()
	c.Equal(0, read().children, "Undo works with OK focused")
	focus(okButton)
	redo()
	c.Equal(1, read().children, "as does Redo")
	focus(cancelButton)
	c.Equal("r.0", read().open, "the row is open again")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape on Cancel closes it")
	screen.Do(func() { p.toggle("r.0") })
	focus(nameField)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape in the name field closes it too")
	c.Equal(2, windows(), "leaving the dialog up")

	// The dialog follows the key bindings as they stand, and a cleared one matches no key.
	uxtest.SwapForTest(t, &undoAction.KeyBinding, unison.KeyBinding{KeyCode: unison.KeyU, Modifiers: mod.OSMenuCommand()})
	uxtest.SwapForTest(t, &redoAction.KeyBinding, unison.KeyBinding{})
	undo()
	c.Equal(1, read().children, "the old Undo key does nothing once Undo is bound to another")
	screen.KeyPress(unison.KeyU, mod.OSMenuCommand())
	c.Equal(0, read().children, "and the new one undoes")
	redo()
	c.Equal(0, read().children, "the old Redo key does nothing once Redo is cleared")
	var taken bool
	screen.Do(func() { taken = dialogWnd.KeyDownCallback(0, mod.None, false) })
	c.False(taken, "and the cleared binding matches no key")

	focus(cancelButton)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windows(), "with no row open, Escape cancels the dialog")
	c.Equal(1, len(savedFilters()), "which saves nothing")

	// Return accepts the dialog even with a row open.
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, _ = modalDialog(t, screen, wnd)
	p = dialogFilterPanel(t, screen, dialogWnd)
	nameField = dialogNameField(t, screen, dialogWnd)
	screen.Type("Ranged")
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")() })
	focus(nameField)
	c.Equal("r.0", read().open, "a row is open")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal(1, windows(), "Return accepts the dialog")
	saved := savedFilters()
	c.Equal(2, len(saved))
	if len(saved) == 2 {
		c.Equal("Ranged", saved[1].Name)
		c.Equal(1, len(saved[1].Root.Children), "with its condition")
	}
}

// TestListFilterEditorSuggestionsEscapeHeadless checks that, in the filter editor's dialog, whose window takes Escape
// to close the open row, Escape with a value's suggestions showing hides them and nothing more.
func TestListFilterEditorSuggestionsEscapeHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	newIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, _ := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")() })
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[string]](t, p.AsPanel(), "r.0:field"); ok {
			popup.Select("tags")
		}
	})
	screen.Do(func() {
		if popup, ok := refAs[*unison.PopupMenu[criteria.StringComparison]](t, p.AsPanel(), "r.0:textcmp"); ok {
			popup.Select(criteria.IsText)
		}
	})
	var field *StringField
	screen.Do(func() { field, _ = refAs[*StringField](t, p.AsPanel(), "r.0:text") })
	if field == nil {
		t.Fatal("the tags condition's value must be a string field")
	}
	screen.Click(screen.PanelCenter(field))
	screen.Type("ment")
	c.NotNil(screen.SuggestionsWindow(dialogWnd), "typing offers the tags of the list")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Nil(screen.SuggestionsWindow(dialogWnd), "Escape hides them")
	var open string
	var windows int
	screen.Do(func() {
		open = p.open
		windows = len(unison.Windows())
	})
	c.Equal("r.0", open, "leaving the row open")
	c.Equal(2, windows, "and the dialog up")
	screen.KeyPress(unison.KeyEscape, mod.None)
	screen.Do(func() { open = p.open })
	c.Equal("", open, "the next Escape closes the row")
	screen.KeyPress(unison.KeyEscape, mod.None)
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "and the one after cancels the dialog")
}

// TestListFilterEditorHeldKeysHeadless checks that a held Escape or Return acts once in the filter editor, so that it
// can't close or open a row and then cancel or accept the dialog too; that Return on Done closes its row rather than
// accepting the dialog; and that Escape in the name field closes the open row and leaves the focus there.
func TestListFilterEditorHeldKeysHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	nameField := dialogNameField(t, screen, dialogWnd)
	screen.Type("Ranged")
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")() })
	open := func() string {
		var path string
		screen.Do(func() { path = p.open })
		return path
	}
	windows := func() int {
		var n int
		screen.Do(func() { n = len(unison.Windows()) })
		return n
	}
	hold := func(key unison.KeyCode) {
		screen.KeyDown(key, mod.None)
		screen.KeyDown(key, mod.None)
		screen.KeyUp(key, mod.None)
	}

	c.Equal("r.0", open(), "the new condition is open")
	hold(unison.KeyEscape)
	c.Equal("", open(), "a held Escape closes the row")
	c.Equal(2, windows(), "and doesn't cancel the dialog")

	screen.Do(func() { c.Equal("r.0"+keySentence, focusedRefKey(dialogWnd), "closing the row focuses its sentence") })
	hold(unison.KeyReturn)
	c.Equal("r.0", open(), "a held Return opens the row")
	c.Equal(2, windows(), "and doesn't accept the dialog")

	screen.Do(func() { c.True(p.focusOn("r.0:done"), "Done can take the focus") })
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal("", open(), "Return on Done closes the row")
	c.Equal(2, windows(), "rather than accepting the dialog")

	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() { nameField.RequestFocus() })
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", open(), "Escape in the name field closes the row")
	var focused bool
	screen.Do(func() { focused = dialogWnd.Focus() == nameField.AsPanel() })
	c.True(focused, "and leaves the focus in the name field")
	screen.KeyPress(unison.KeyEnd, mod.None)
	screen.Type("X")
	var name string
	screen.Do(func() { name = nameField.Text() })
	c.Equal("RangedX", name, "so typing goes on in it")
}

// TestListFilterDialogUndoMenuHeadless checks that the Undo and Redo commands find the dialog's undo manager with a
// button focused, outside the dialog's content, as macOS's menu bar looks them up, and that where the menu bar takes
// their keys, the dialog leaves the keys to it.
func TestListFilterDialogUndoMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle))
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	children := func() int {
		var n int
		screen.Do(func() { n = len(p.filter.Root.Children) })
		return n
	}
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")() })
	ok := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	screen.Do(func() { ok.RequestFocus() })
	var enabled bool
	screen.Do(func() { enabled = undoAction.Enabled(nil) })
	c.True(enabled, "Undo is available with OK focused")
	screen.Do(func() { undoAction.Execute(nil) })
	c.Equal(0, children(), "and undoes")
	screen.Do(func() { redoAction.Execute(nil) })
	c.Equal(1, children(), "as Redo redoes")

	uxtest.SwapForTest(t, &dialogMenuTakesUndoKeys, true)
	screen.Do(func() { ok.RequestFocus() })
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	c.Equal(1, children(), "where the menu bar takes the Undo key, the dialog leaves it alone")
}

// TestListFilterDialogDragHeadless drags a row by its sentence within the filter editor's dialog, which is a window of
// its own, so the drop lands only because the dialog takes the rows' drag type.
func TestListFilterDialogDragHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	uxtest.SwapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	var first, second gurps.FilterNode
	screen.Do(func() {
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")()
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")()
		p.toggle("r.1")
	})
	var sentence, target *unison.Panel
	var above geom.Point
	screen.Do(func() {
		first = p.node("r.0")
		second = p.node("r.1")
		sentence, _ = refAs[*unison.Panel](t, p.AsPanel(), "r.1"+keySentence)
		if more, ok := refAs[*unison.Panel](t, p.AsPanel(), "r.0"+keyMore); ok {
			target = more.Parent()
			above = geom.NewPoint(40, target.FrameRect().Height*0.2)
		}
	})
	if sentence == nil || target == nil {
		return
	}
	c.NotNil(first, "the first condition is at r.0")
	c.NotNil(second, "the second condition is at r.1")
	screen.Drag(screen.PanelCenter(sentence), screen.PanelPoint(target, above), 10)
	screen.Do(func() {
		c.True(second == p.node("r.0"), "the dragged condition goes first")
		c.True(first == p.node("r.1"), "ahead of the other")
	})
}
