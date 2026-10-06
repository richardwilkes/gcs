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
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// The titles of the saved filter popup's commands.
const (
	newFilterItemTitle    = "New Filter…"
	editFilterItemTitle   = "Edit Filter…"
	deleteFilterItemTitle = "Delete Filter…"
)

// listFilterHeadlessTraitNames are the names of the fixture's traits, in the order the list holds them, which is what
// the table shows whenever nothing is filtering it.
var listFilterHeadlessTraitNames = []string{"Acute Vision", "Combat Reflexes", "Fur"}

// newListFilterHeadlessTraits returns the three traits the tests filter: each bears a name, a tag and a point cost of
// its own, so that every filter below can be told apart row by row. Only one of them is tagged "Physical", which is
// what lets the quick filter be seen searching the tags column rather than the name.
func newListFilterHeadlessTraits() []*gurps.Trait {
	acuteVision := newFilterTestTrait("Acute Vision", "Physical")
	acuteVision.BasePoints = fxp.FromInteger(2)
	combatReflexes := newFilterTestTrait("Combat Reflexes", "Mental")
	combatReflexes.BasePoints = fxp.FromInteger(15)
	fur := newFilterTestTrait("Fur", "Exotic")
	fur.BasePoints = fxp.FromInteger(4)
	return []*gurps.Trait{acuteVision, combatReflexes, fur}
}

// openListFilterTraitDockable opens a trait list holding the fixture's traits in the workspace's document dock and
// returns it. The file lives in the test's own directory, so nothing the dockable might write touches the user's files.
func openListFilterTraitDockable(t *testing.T, screen *unison.HeadlessScreen) *TableDockable[*gurps.Trait] {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test"+gurps.TraitsExt)
	var d *TableDockable[*gurps.Trait]
	screen.Do(func() {
		d = NewTraitTableDockable(path, newListFilterHeadlessTraits())
		DisplayNewDockable(d)
	})
	if d == nil {
		t.Fatal("the trait list could not be opened")
	}
	return d
}

// seedListFilter saves a filter for the trait list type whose lone condition compares the field with the given key
// against the qualifier, and returns it. A text or list field is what the tests use, so the text criteria is the one
// that is filled in.
func seedListFilter(name, fieldKey, qualifier string) *gurps.ListFilter {
	f := gurps.NewListFilter(name)
	condition := gurps.NewFilterCondition(f.Root, fieldKey)
	condition.Text.Compare = criteria.ContainsText
	condition.Text.Qualifier = qualifier
	f.Root.Children = append(f.Root.Children, condition)
	gurps.GlobalSettings().AddListFilter(listFilterPopupTestKey, f)
	return f
}

// listFilterState is what the tests look at after each choice: the rows the list is showing and how the toolbar's two
// filters stand.
type listFilterState struct {
	selected     *gurps.ListFilter
	popupText    string
	fieldText    string
	names        []string
	fieldEnabled bool
	filtered     bool
}

// readListFilterState takes a snapshot of the dockable's filtering, all of it read in one pass on the UI thread.
func readListFilterState(screen *unison.HeadlessScreen, d *TableDockable[*gurps.Trait]) listFilterState {
	var state listFilterState
	screen.Do(func() {
		state.selected = d.selectedFilter
		state.popupText = d.savedFilters.popup.Text()
		state.fieldText = d.filterField.Text()
		state.names = visibleTraitNames(d)
		state.fieldEnabled = d.filterField.Enabled()
		state.filtered = d.table.IsFiltered()
	})
	return state
}

// popupItemIndexOf returns the index of the popup's item with the title, failing the test if there is none.
func popupItemIndexOf(t *testing.T, screen *unison.HeadlessScreen, popup *unison.PopupMenu[string], title string) int {
	t.Helper()
	index := -1
	screen.Do(func() {
		for i := range popup.ItemCount() {
			if one, _ := popup.ItemAt(i); one == title {
				index = i
				return
			}
		}
	})
	if index == -1 {
		t.Fatalf("the popup has no item %q", title)
	}
	return index
}

// dialogButton returns the dialog's button for the given modal response, failing the test if it has none.
func dialogButton(t *testing.T, screen *unison.HeadlessScreen, dialog *unison.Dialog, response int) *unison.Button {
	t.Helper()
	var button *unison.Button
	screen.Do(func() { button = dialog.Button(response) })
	if button == nil {
		t.Fatalf("the dialog has no button for response %d", response)
	}
	return button
}

// dialogFilterPanel returns the filter editor in the dialog.
func dialogFilterPanel(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) *listFilterPanel {
	t.Helper()
	var p *listFilterPanel
	screen.Do(func() { p, _ = firstPanelOfType[*listFilterPanel](dialogWnd.Content()) })
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
	screen.Do(func() { fields = panelsOfType[*StringField](dialogWnd.Content()) })
	if len(fields) != 1 {
		t.Fatalf("expected the name field to be the dialog's one string field, found %d string fields", len(fields))
	}
	return fields[0]
}

// TestListFilterPopupAppliesSavedFilterHeadless drives the saved filter popup of a trait list inside a headless
// workspace: choosing a saved filter narrows the list and leaves the quick filter's field usable, what is typed there
// narrows the list further while the saved filter is in force, choosing None drops the saved filter alone, and the
// quick filter searches the tags column as well as the name.
func TestListFilterPopupAppliesSavedFilterHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)

	var inWorkspace bool
	var titles []string
	screen.Do(func() {
		inWorkspace = d.Window() == wnd
		titles = popupItemTitles(d.savedFilters.popup)
	})
	c.True(inWorkspace, "the trait list opens in the workspace window")
	c.Equal([]string{
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "None", separatorTitle,
		"Mental",
	}, titles, "the commands lead, and the separator after None puts the saved filter at index 6")

	state := readListFilterState(screen, d)
	c.Equal(listFilterHeadlessTraitNames, state.names, "every trait is shown before any filtering")
	c.Equal("None", state.popupText, "no saved filter is in force to start with")
	c.True(state.fieldEnabled, "and the quick filter's field is usable")
	c.False(state.filtered, "and nothing is filtering the table")

	// Choose the saved filter. The item ahead of it is the separator, which occupies an index of its own.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, popupFirstSavedIndex)
	state = readListFilterState(screen, d)
	c.Equal([]string{"Combat Reflexes"}, state.names, "only the traits the saved filter accepts may be shown")
	c.True(mental == state.selected, "the saved filter itself must be the one in force")
	c.Equal("Mental", state.popupText, "and the popup must show it")
	c.True(state.fieldEnabled, "the quick filter's field must stay usable while a saved filter is in force")
	c.True(state.filtered, "the saved filter must filter the table")
	captureScreen(t, c, screen, "list_filter_applied")

	// Type into the quick filter while the saved filter is in force. The one trait the saved filter keeps is not
	// tagged "Physical", so nothing passes both, which shows the two being applied together rather than the typed
	// text taking over.
	screen.Click(screen.PanelCenter(d.filterField))
	screen.Type("physical")
	state = readListFilterState(screen, d)
	c.Equal("physical", state.fieldText, "the typed text reaches the quick filter's field")
	c.True(mental == state.selected, "typing in the quick filter must leave the saved filter in force")
	c.True(state.filtered, "the two together must filter the table")
	c.Equal(0, len(state.names), "no trait passes both the saved filter and the quick filter")
	captureScreen(t, c, screen, "list_filter_combined")

	// Choose None, which drops the saved filter and leaves the quick filter's text to filter the list on its own.
	// Only one trait is tagged "Physical" and no trait's name holds the word, so keeping exactly that one shows the
	// quick filter searching the tags column.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, popupNoneIndex)
	state = readListFilterState(screen, d)
	c.Nil(state.selected, "the saved filter must have been dropped")
	c.Equal("physical", state.fieldText, "dropping the saved filter must leave the quick filter's text alone")
	c.True(state.fieldEnabled, "and its field usable")
	c.True(state.filtered, "the quick filter must filter the table on its own")
	c.Equal([]string{"Acute Vision"}, state.names, "the quick filter searches the tags column, not just the name")
}

// TestListFilterNewFilterDialogHeadless drives the editor for a new filter inside a headless workspace: OK is held
// back until the name is one no other saved filter bears, the editor opens with room for a few levels of nesting, a
// condition can be added to the root group, and accepting saves the filter and puts it in force.
func TestListFilterNewFilterDialogHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	seedListFilter("Melee", "name", "melee")
	d := openListFilterTraitDockable(t, screen)

	// Put the quick filter to work first, so that the new filter can be seen to combine with it rather than replace
	// it.
	screen.Click(screen.PanelCenter(d.filterField))
	screen.Type("fur")
	c.Equal([]string{"Fur"}, readListFilterState(screen, d).names, "the quick filter has the list to start with")

	newIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle)
	c.Equal(0, newIndex, "New Filter… is the first item")
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}

	nameField := dialogNameField(t, screen, dialogWnd)
	var contentWidth, contentHeight float32
	screen.Do(func() {
		rect := dialogWnd.ContentRect()
		contentWidth = rect.Width
		contentHeight = rect.Height
	})
	c.False(okEnabled(), "a filter with no name cannot be accepted")
	c.True(contentWidth >= listFilterDialogMinWidth,
		"the editor opens at least %d wide, but is %v", listFilterDialogMinWidth, contentWidth)
	c.True(contentHeight >= listFilterDialogMinHeight,
		"the editor opens at least %d tall, but is %v", listFilterDialogMinHeight, contentHeight)

	// A name another saved filter already bears is refused, whatever its case.
	screen.Click(screen.PanelCenter(nameField))
	screen.Type("melee")
	c.False(okEnabled(), "a name a saved filter already bears cannot be accepted, whatever its case")

	// Replace it with a name of its own, which is accepted.
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("Ranged")
	var name string
	screen.Do(func() { name = nameField.Text() })
	c.Equal("Ranged", name, "select-all followed by typing replaces the name")
	c.True(okEnabled(), "a name no other saved filter bears can be accepted")
	captureScreen(t, c, screen, "list_filter_dialog")

	// Add a condition to the empty root through its placeholder. It starts out testing the first field, open.
	p := dialogFilterPanel(t, screen, dialogWnd)
	var placeholder *unison.Panel
	screen.Do(func() { placeholder = p.FindRefKey(treeRootPath + ":empty") })
	c.NotNil(placeholder, "an empty filter shows its placeholder")
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")() })
	screen.Do(func() {
		c.Equal("r.0", p.open, "the added condition is open")
		c.Equal("r.0:field", dialogWnd.Focus().RefKey, "with its field popup focused")
	})

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

// TestListFilterDeleteAsksAndFallsBackHeadless drives the Delete Filter… command inside a headless workspace: it asks
// before removing anything, a refusal leaves the filter in force, and confirming removes it and leaves the list with no
// saved filter.
func TestListFilterDeleteAsksAndFallsBackHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)

	choosePopupItem(t, screen, wnd, d.savedFilters.popup, popupFirstSavedIndex)
	c.True(mental == readListFilterState(screen, d).selected,
		"the saved filter itself must be in force before it is deleted")
	var itemCount int
	screen.Do(func() { itemCount = d.savedFilters.popup.ItemCount() })
	deleteIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, deleteFilterItemTitle)
	c.Equal(2, deleteIndex, "Delete Filter… is the third item")

	// Refuse the confirmation. The prompt names the filter in force, and turning it down removes nothing.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, deleteIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var prompt string
	screen.Do(func() { prompt = strings.Join(labelTexts(dialogWnd.Content()), "\n") })
	c.Contains(prompt, `Delete the filter 'Mental'?`, "the prompt must name the filter in force")
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	c.Equal(1, len(savedFilters()), "a refused deletion must remove nothing")
	state := readListFilterState(screen, d)
	c.True(mental == state.selected, "and must leave the filter itself in force")
	c.Equal([]string{"Combat Reflexes"}, state.names, "so the list is still filtered by it")

	// Confirm it the second time around. The filter goes away and the list is left with no saved filter.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, deleteIndex)
	_, dialog = modalDialog(t, screen, wnd)
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	var windows int
	var titles []string
	screen.Do(func() {
		windows = len(unison.Windows())
		titles = popupItemTitles(d.savedFilters.popup)
	})
	c.Equal(1, windows, "the prompt has been dismissed")
	c.Equal(0, len(savedFilters()), "confirming must remove the filter")
	state = readListFilterState(screen, d)
	c.Nil(state.selected, "the list must be left with no saved filter")
	c.Equal("None", state.popupText, "which the popup must show")
	c.False(state.filtered, "with the quick filter empty, nothing filters the table")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown again")
	c.Equal([]string{
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "None",
	}, titles, "the deleted filter and the separator that set the saved filters apart must both be gone")
	c.Equal(itemCount-2, len(titles), "so the popup holds two items fewer than it did")
}

// TestListFilterEditorKeysHeadless drives the filter editor's keys inside a headless workspace: the key bindings of
// Undo and Redo work within the dialog, which has no menu bar of its own, whether the focus is in the name field, the
// conditions or a button, and undoing the name brings the OK button into line. Escape closes the open row before it
// cancels the dialog, wherever the focus is, and Return accepts it.
func TestListFilterEditorKeysHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	swapForTest(t, &dialogMenuTakesUndoKeys, false)
	seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)
	newIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle)
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
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
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")() })
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "a named filter with a condition")

	focus(nameField)
	undo()
	c.Equal(0, read().children, "Undo in the name field takes the condition back")
	undo()
	c.Equal(state{}, read(), "and then the name, which leaves OK off")
	redo()
	c.Equal(state{name: "Ranged", ok: true}, read(), "Redo puts the name back, and OK on")
	redo()
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "and then the condition, open")

	focus(okButton)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape on OK closes the open row")
	c.Equal(2, windows(), "and leaves the dialog up")
	focus(cancelButton)
	undo()
	c.Equal(0, read().children, "Undo on Cancel works too")
	redo()
	c.Equal(state{name: "Ranged", open: "r.0", children: 1, ok: true}, read(), "as does Redo, opening the row again")
	focus(cancelButton)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape on Cancel closes it")
	screen.Do(func() { p.toggle("r.0") })
	focus(nameField)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", read().open, "Escape in the name field closes it too")
	c.Equal(2, windows(), "leaving the dialog up")

	// The dialog follows the key bindings as they stand, and a cleared one matches no key.
	swapForTest(t, &undoAction.KeyBinding, unison.KeyBinding{KeyCode: unison.KeyU, Modifiers: mod.OSMenuCommand()})
	swapForTest(t, &redoAction.KeyBinding, unison.KeyBinding{})
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
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, _ = modalDialog(t, screen, wnd)
	p = dialogFilterPanel(t, screen, dialogWnd)
	nameField = dialogNameField(t, screen, dialogWnd)
	screen.Type("Ranged")
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")() })
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

// TestListFilterDialogUndoMenuHeadless checks that the Undo and Redo commands find the dialog's undo manager with a
// button focused, outside the dialog's content, as macOS's menu bar looks them up, and that where the menu bar takes
// their keys, the dialog leaves the keys to it.
func TestListFilterDialogUndoMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	choosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle))
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	children := func() int {
		var n int
		screen.Do(func() { n = len(p.filter.Root.Children) })
		return n
	}
	screen.Do(func() { menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")() })
	ok := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	screen.Do(func() { ok.RequestFocus() })
	var enabled bool
	screen.Do(func() { enabled = undoAction.Enabled(nil) })
	c.True(enabled, "Undo is available with OK focused")
	screen.Do(func() { undoAction.Execute(nil) })
	c.Equal(0, children(), "and undoes")
	screen.Do(func() { redoAction.Execute(nil) })
	c.Equal(1, children(), "as Redo redoes")

	swapForTest(t, &dialogMenuTakesUndoKeys, true)
	screen.Do(func() { ok.RequestFocus() })
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	c.Equal(1, children(), "where the menu bar takes the Undo key, the dialog leaves it alone")
}

// TestListFilterEditCancelHeadless edits a saved filter through the real dialog and cancels it, checking that the
// saved filter is left as it was.
func TestListFilterEditCancelHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	before := gurps.Hash64(mental)
	d := openListFilterTraitDockable(t, screen)
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, popupItemIndexOf(t, screen, d.savedFilters.popup, "Mental"))
	choosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, editFilterItemTitle))
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	screen.Type("Changed")
	screen.Do(func() {
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")()
		menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Delete")()
	})
	screen.Do(func() {
		c.True(p.filter != mental, "the dialog edits a copy")
		c.Equal("Changed", p.filter.Name, "which takes the typed name")
		c.Equal([]string{gurps.TraitFilterFields()[0].Key}, filterShape(p.filter.Root),
			"and the change to its conditions")
	})
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "Cancel dismisses the dialog")
	c.Equal("Mental", mental.Name, "the saved filter keeps its name")
	c.Equal(before, gurps.Hash64(mental), "and its conditions")
	c.True(mental == readListFilterState(screen, d).selected, "and stays in force")
}

// TestListFilterDialogDragHeadless drags a row by its sentence within the filter editor's dialog, which is a window of
// its own, so the drop lands only because the dialog takes the rows' drag type.
func TestListFilterDialogDragHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)
	choosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, newFilterItemTitle))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	var first, second gurps.FilterNode
	screen.Do(func() {
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")()
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Condition")()
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

// TestListFilterChangesReachOtherDockablesHeadless verifies that a change made to the saved filters through one
// list's popup reaches every other open list of the same type: a renamed filter that is in force there keeps its
// place under its new name, a new filter appears in its popup, and a deleted filter that was in force there leaves the
// list with no saved filter. Each time, the other list's toolbar is laid out again, which shows in the width of
// its popup, since that follows the widest item. The two lists share a tab group, so only the second, in front, can
// be clicked; the first is driven through its popup's callback directly. The editor is stubbed, since it is the
// popups' bookkeeping that is under test here.
func TestListFilterChangesReachOtherDockablesHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	first := openListFilterTraitDockable(t, screen)
	second := openListFilterTraitDockable(t, screen)
	choosePopupItem(t, screen, wnd, second.savedFilters.popup, popupFirstSavedIndex)
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex)
	state := readListFilterState(screen, second)
	c.True(mental == state.selected, "the saved filter itself must be in force in the list in front")
	c.Equal([]string{"Combat Reflexes"}, state.names, "and filtering it")
	c.True(mental == readListFilterState(screen, first).selected, "and in force in the list behind")

	// Rename the filter through the list behind. The one in front keeps it in force, under the new name.
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "Mind"
			return true
		})
	editIndex := popupItemIndexOf(t, screen, first.savedFilters.popup, editFilterItemTitle)
	choosePopupItemDirectly(screen, first, editIndex)
	state = readListFilterState(screen, second)
	c.True(mental == state.selected, "the renamed filter itself must still be in force in the list in front")
	c.Equal("Mind", state.popupText, "under its new name")
	c.Equal([]string{"Combat Reflexes"}, state.names, "and still filtering it")

	// Create a filter through the list behind. The popup in front lists it, and grows to fit its long name, while the
	// list keeps what it had in force.
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "A much longer filter name"
			return true
		})
	widthBefore := popupWidth(screen, second)
	newIndex := popupItemIndexOf(t, screen, first.savedFilters.popup, newFilterItemTitle)
	choosePopupItemDirectly(screen, first, newIndex)
	var titles []string
	screen.Do(func() { titles = popupItemTitles(second.savedFilters.popup) })
	c.Equal([]string{
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "None", separatorTitle,
		"A much longer filter name", "Mind",
	}, titles, "the popup in front must list the new filter")
	state = readListFilterState(screen, second)
	c.True(mental == state.selected, "the list in front must keep the very filter it had in force")
	c.Equal("Mind", state.popupText, "and show it")
	widthAfterAdd := popupWidth(screen, second)
	c.True(widthAfterAdd > widthBefore, "the popup in front must have been laid out again to fit the new item")

	// Delete the filter in force through the list behind, which put the new one in force when it created it, so it
	// has to be put back first. The list in front, still on the deleted filter, falls back to None.
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex+1) // "Mind"
	swapForTest(t, &confirmFilterDeletion, func(_ string) bool { return true })
	deleteIndex := popupItemIndexOf(t, screen, first.savedFilters.popup, deleteFilterItemTitle)
	choosePopupItemDirectly(screen, first, deleteIndex)
	state = readListFilterState(screen, second)
	c.Nil(state.selected, "the list in front must fall back to no saved filter once its filter is gone")
	c.Equal("None", state.popupText, "and its popup must say so")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown there again")
	c.Nil(readListFilterState(screen, first).selected, "the list behind must be on no saved filter as well")

	// Delete the remaining filter too. The popup in front shrinks back, having only None and the commands left to
	// fit.
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex) // "A much longer filter name"
	deleteIndex = popupItemIndexOf(t, screen, first.savedFilters.popup, deleteFilterItemTitle)
	choosePopupItemDirectly(screen, first, deleteIndex)
	screen.Do(func() { titles = popupItemTitles(second.savedFilters.popup) })
	c.Equal([]string{newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "None"},
		titles, "the popup in front must have lost the deleted filter")
	c.True(popupWidth(screen, second) < widthAfterAdd, "the popup in front must have been laid out again to shrink")
}

// choosePopupItemDirectly makes the choice at the given index in the dockable's saved filter popup through the popup's
// own callback, the way the menu would, without opening the menu. It is for a dockable that can't be clicked because
// another one in its tab group is in front of it.
func choosePopupItemDirectly(screen *unison.HeadlessScreen, d *TableDockable[*gurps.Trait], index int) {
	screen.Do(func() {
		title, _ := d.savedFilters.popup.ItemAt(index)
		d.savedFilters.popup.ChoiceMadeCallback(d.savedFilters.popup, index, title)
	})
}

// popupWidth returns the width the dockable's saved filter popup was last laid out at.
func popupWidth(screen *unison.HeadlessScreen, d *TableDockable[*gurps.Trait]) float32 {
	var width float32
	screen.Do(func() { width = d.savedFilters.popup.FrameRect().Width })
	return width
}
