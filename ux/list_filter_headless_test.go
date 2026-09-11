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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// The titles of the trait fields the tests drive the condition row's field popup with, and the tooltips that tell the
// widgets a condition row holds apart from one another.
const (
	nameFieldTitle        = "have a name"
	tagsFieldTitle        = "have tags"
	pointsFieldTitle      = "have points"
	containerFieldTitle   = "be a container"
	fieldPopupTooltip     = "The field to test"
	addConditionTooltip   = "Add a condition"
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

// popupItemIndexFromEnd returns the index of the item the given number of places from the end of the popup, along with
// its title, so that a test can both aim at a command and show that it aimed at the right one.
func popupItemIndexFromEnd(screen *unison.HeadlessScreen, popup *unison.PopupMenu[string], fromEnd int) (index int, title string) {
	screen.Do(func() {
		index = popup.ItemCount() - fromEnd
		title, _ = popup.ItemAt(index)
	})
	return index, title
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

// filterEditorRootRow returns the row of the filter editor's root group, which is the only child of the editor's
// panel. Every other row hangs off of it: the group's own columns come first and the rows of its children follow.
func filterEditorRootRow(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) *unison.Panel {
	t.Helper()
	var row *unison.Panel
	screen.Do(func() {
		if panel, ok := firstPanelOfType[*listFilterPanel](dialogWnd.Content()); ok && len(panel.Children()) == 1 {
			row = panel.Children()[0]
		}
	})
	if row == nil {
		t.Fatal("the filter editor has no root group row")
	}
	return row
}

// clickButtonWithTooltip clicks the button within root whose tooltip reads exactly text, failing the test if there is
// none. The filter editor's buttons carry icons rather than titles, so the tooltip is what identifies them.
func clickButtonWithTooltip(t *testing.T, screen *unison.HeadlessScreen, root *unison.Panel, text string) {
	t.Helper()
	var button *unison.Button
	screen.Do(func() { button = buttonWithTooltip(root, text) })
	if button == nil {
		t.Fatalf("no button tooltipped %q", text)
	}
	screen.Click(screen.PanelCenter(button))
}

// traitFilterFieldIndex returns the position of the trait field with the given title among the fields the condition
// row's field popup offers, which are the trait fields in their own order.
func traitFilterFieldIndex(t *testing.T, title string) int {
	t.Helper()
	for i, field := range gurps.TraitFilterFields() {
		if field.Title == title {
			return i
		}
	}
	t.Fatalf("no trait filter field is titled %q", title)
	return -1
}

// conditionFieldPopup returns the popup that chooses which field a condition row tests, found by the tooltip only it
// carries, since the row holds two other popups of the same type.
func conditionFieldPopup(t *testing.T, screen *unison.HeadlessScreen, row *unison.Panel) *unison.PopupMenu[string] {
	t.Helper()
	var popup *unison.PopupMenu[string]
	screen.Do(func() {
		for _, one := range panelsOfType[*unison.PopupMenu[string]](row) {
			if one.Tooltip != nil && tooltipText(one.Tooltip) == fieldPopupTooltip {
				popup = one
				return
			}
		}
	})
	if popup == nil {
		t.Fatal("the condition row has no field popup")
	}
	return popup
}

// conditionComparisonPopup returns the popup that leads a condition row's criteria, which is the third of the row's
// popups, exactly as readConditionRow finds it: the must/must not popup and the field popup come ahead of it.
func conditionComparisonPopup(t *testing.T, screen *unison.HeadlessScreen, row *unison.Panel) *unison.PopupMenu[string] {
	t.Helper()
	var popup *unison.PopupMenu[string]
	screen.Do(func() {
		if popups := panelsOfType[*unison.PopupMenu[string]](row); len(popups) > 2 {
			popup = popups[2]
		}
	})
	if popup == nil {
		t.Fatal("the condition row has no comparison popup")
	}
	return popup
}

// conditionRowShape is what a condition row holds once its field has been chosen: the criteria widgets the field's
// kind calls for, the tooltips those widgets carry, where the criteria's comparison popup stands, and whether the
// row's layout has been brought into line with the children it now has.
type conditionRowShape struct {
	stringTooltips []string
	strings        int
	decimals       int
	comparison     int
	children       int
	columns        int
}

// readConditionRow takes a snapshot of a condition row, all of it read in one pass on the UI thread. The comparison is
// reported as -1 when the row has no comparison popup, which is what a yes/no field leaves behind.
func readConditionRow(screen *unison.HeadlessScreen, row *unison.Panel) conditionRowShape {
	shape := conditionRowShape{comparison: -1}
	screen.Do(func() {
		for _, field := range panelsOfType[*StringField](row) {
			shape.strings++
			shape.stringTooltips = append(shape.stringTooltips, tooltipText(field.Tooltip))
		}
		shape.decimals = len(panelsOfType[*DecimalField](row))
		// The row's popups are, in order, the must/must not popup, the field popup and, when the field's kind calls for
		// criteria, the comparison popup that leads them.
		if popups := panelsOfType[*unison.PopupMenu[string]](row); len(popups) > 2 {
			shape.comparison = popups[2].SelectedIndex()
		}
		shape.children = len(row.Children())
		if layout, ok := row.Layout().(*unison.FlexLayout); ok {
			shape.columns = layout.Columns
		}
	})
	return shape
}

// TestListFilterPopupAppliesSavedFilterHeadless drives the saved filter popup of a trait list inside a headless
// workspace: choosing a saved filter takes the list over and locks the quick filter's field, going back to the quick
// filter hands the list back, and what is typed there searches the tags column as well as the name.
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
		"Quick Filter", separatorTitle, "Mental", separatorTitle, newFilterItemTitle, editFilterItemTitle,
		deleteFilterItemTitle,
	}, titles, "the separator after the quick filter is what puts the saved filter at index 2")

	state := readListFilterState(screen, d)
	c.Equal(listFilterHeadlessTraitNames, state.names, "every trait is shown before any filtering")
	c.Equal("Quick Filter", state.popupText, "the quick filter has the list to start with")
	c.True(state.fieldEnabled, "so its field is usable")
	c.False(state.filtered, "and nothing is filtering the table")

	// Choose the saved filter. The item ahead of it is the separator, which occupies an index of its own.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, 2)
	state = readListFilterState(screen, d)
	c.Equal([]string{"Combat Reflexes"}, state.names, "only the traits the saved filter accepts may be shown")
	c.True(mental == state.selected, "the saved filter itself must be the one in force")
	c.Equal("Mental", state.popupText, "and the popup must show it")
	c.False(state.fieldEnabled, "the quick filter's field must be unusable while a saved filter is in force")
	c.True(state.filtered, "the saved filter must filter the table")
	captureScreen(t, c, screen, "list_filter_applied")

	// Go back to the quick filter, which hands the list back and unlocks the field again.
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, listFilterQuickIndex)
	state = readListFilterState(screen, d)
	c.Nil(state.selected, "the list must go back to the quick filter")
	c.True(state.fieldEnabled, "whose field must be usable again")
	c.False(state.filtered, "the emptied quick filter shows everything")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown again")

	// Type into the quick filter. Only one trait is tagged "Physical" and no trait's name holds the word, so keeping
	// exactly that one shows the quick filter searching the tags column.
	screen.Click(screen.PanelCenter(d.filterField))
	screen.Type("physical")
	state = readListFilterState(screen, d)
	c.Equal("physical", state.fieldText, "the typed text reaches the quick filter's field")
	c.True(state.filtered, "typing in the quick filter must filter the table")
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

	// Put the quick filter to work first, so that the new filter taking the list over can be seen to blank it.
	screen.Click(screen.PanelCenter(d.filterField))
	screen.Type("fur")
	c.Equal([]string{"Fur"}, readListFilterState(screen, d).names, "the quick filter has the list to start with")

	newIndex, newTitle := popupItemIndexFromEnd(screen, d.savedFilters.popup, 3)
	c.Equal(newFilterItemTitle, newTitle, "New Filter… is the third item from the end")
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}

	var nameFields []*StringField
	var contentWidth, contentHeight float32
	screen.Do(func() {
		nameFields = panelsOfType[*StringField](dialogWnd.Content())
		rect := dialogWnd.ContentRect()
		contentWidth = rect.Width
		contentHeight = rect.Height
	})
	if len(nameFields) != 1 {
		t.Fatalf("expected the editor of an empty filter to hold the name field alone, found %d string fields",
			len(nameFields))
	}
	nameField := nameFields[0]
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

	// Add a condition to the root group. It starts out testing the first field, which is a text one, so it brings a
	// qualifier field of its own along.
	clickButtonWithTooltip(t, screen, dialogWnd.Content(), addConditionTooltip)
	rootRow := filterEditorRootRow(t, screen, dialogWnd)
	var rootChildren, stringFields int
	screen.Do(func() {
		rootChildren = len(rootRow.Children())
		stringFields = len(panelsOfType[*StringField](dialogWnd.Content()))
	})
	c.Equal(filterGroupColumns+1, rootChildren, "the group's own columns must be followed by the added condition's row")
	c.Equal(2, stringFields, "the added condition brings a qualifier field of its own")

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
	c.Equal("", state.fieldText, "the quick filter's text must have been cleared, so the two cannot disagree")
	c.False(state.fieldEnabled, "and its field must be unusable")
	c.Equal(listFilterHeadlessTraitNames, state.names,
		"the new filter's lone condition compares against nothing, so every trait passes it")
	c.True(state.filtered, "the saved filter must be driving the rows even though every trait passes it")

	editIndex, editTitle := popupItemIndexFromEnd(screen, d.savedFilters.popup, 2)
	deleteIndex, deleteTitle := popupItemIndexFromEnd(screen, d.savedFilters.popup, 1)
	c.Equal(editFilterItemTitle, editTitle, "Edit Filter… is the second item from the end")
	c.Equal(deleteFilterItemTitle, deleteTitle, "Delete Filter… is the last item")
	var editEnabled, deleteEnabled bool
	screen.Do(func() {
		editEnabled = d.savedFilters.popup.ItemEnabledAt(editIndex)
		deleteEnabled = d.savedFilters.popup.ItemEnabledAt(deleteIndex)
	})
	c.True(editEnabled, "Edit Filter… must be available with a filter in force")
	c.True(deleteEnabled, "Delete Filter… must be available with a filter in force")
}

// TestListFilterDeleteAsksAndFallsBackHeadless drives the Delete Filter… command inside a headless workspace: it asks
// before removing anything, a refusal leaves the filter in force, and confirming removes it and hands the list back to
// the quick filter.
func TestListFilterDeleteAsksAndFallsBackHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)

	choosePopupItem(t, screen, wnd, d.savedFilters.popup, 2)
	c.True(mental == readListFilterState(screen, d).selected,
		"the saved filter itself must be in force before it is deleted")
	var itemCount int
	screen.Do(func() { itemCount = d.savedFilters.popup.ItemCount() })
	deleteIndex, deleteTitle := popupItemIndexFromEnd(screen, d.savedFilters.popup, 1)
	c.Equal(deleteFilterItemTitle, deleteTitle, "Delete Filter… is the last item")

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
	c.False(state.fieldEnabled, "and the quick filter's field is still unusable")

	// Confirm it the second time around. The filter goes away and the list falls back to the quick filter.
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
	c.Nil(state.selected, "the list must go back to the quick filter")
	c.Equal("Quick Filter", state.popupText, "which the popup must show")
	c.True(state.fieldEnabled, "the quick filter's field must be usable again")
	c.False(state.filtered, "the emptied quick filter shows everything")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown again")
	c.Equal([]string{
		"Quick Filter", separatorTitle, newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle,
	}, titles, "the deleted filter and the separator that set the saved filters apart must both be gone")
	c.Equal(itemCount-2, len(titles), "so the popup holds two items fewer than it did")
}

// TestListFilterConditionRowRebuildsForFieldKindHeadless drives a condition row's field popup inside a headless
// workspace, showing that the criteria trailing the popup are thrown away and rebuilt to suit each field's kind, that
// the row's layout follows, and that canceling the editor saves nothing.
func TestListFilterConditionRowRebuildsForFieldKindHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &lastFilterFieldKeyUsed, make(map[string]string))
	d := openListFilterTraitDockable(t, screen)

	newIndex, newTitle := popupItemIndexFromEnd(screen, d.savedFilters.popup, 3)
	c.Equal(newFilterItemTitle, newTitle, "New Filter… is the third item from the end, even with no saved filters")
	choosePopupItem(t, screen, wnd, d.savedFilters.popup, newIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)

	clickButtonWithTooltip(t, screen, dialogWnd.Content(), addConditionTooltip)
	rootRow := filterEditorRootRow(t, screen, dialogWnd)
	var row *unison.Panel
	screen.Do(func() {
		if children := rootRow.Children(); len(children) == filterGroupColumns+1 {
			row = children[filterGroupColumns]
		}
	})
	if row == nil {
		t.Fatal("the root group must hold the row of the condition that was added")
	}
	// The popups within the row are read from the row rather than from the dialog, so that the popups of the group's
	// own row cannot be mistaken for the condition's.
	fieldPopup := conditionFieldPopup(t, screen, row)

	// A list field is compared against text that may hold several values.
	choosePopupItem(t, screen, dialogWnd, fieldPopup, traitFilterFieldIndex(t, tagsFieldTitle))
	shape := readConditionRow(screen, row)
	c.Equal(1, shape.strings, "a list field is compared against text")
	c.Equal(0, shape.decimals, "and against no number")
	if shape.strings == 1 {
		c.Contains(shape.stringTooltips[0], "commas", "a list field's qualifier may hold several values")
	}
	c.Equal(shape.children, shape.columns, "the row's layout must hold one column per child")

	// Move the comparison off its default. A list field is compared through the same text criteria a text field is, so
	// what the field changes below do to it shows when the name field is reached.
	choosePopupItem(t, screen, dialogWnd, conditionComparisonPopup(t, screen, row), 1)
	c.Equal(1, readConditionRow(screen, row).comparison, "the comparison must be the one just chosen")

	// A yes/no field needs no criteria at all: the must/must not popup ahead of the field says all there is to say.
	choosePopupItem(t, screen, dialogWnd, fieldPopup, traitFilterFieldIndex(t, containerFieldTitle))
	shape = readConditionRow(screen, row)
	c.Equal(0, shape.strings, "a yes/no field has nothing to compare against")
	c.Equal(0, shape.decimals, "of either kind")
	c.Equal(-1, shape.comparison, "so the row keeps no comparison popup")
	c.Equal(shape.children, shape.columns,
		"and the row's layout must be brought into line with the children it now has")

	// A number field is compared against a number.
	choosePopupItem(t, screen, dialogWnd, fieldPopup, traitFilterFieldIndex(t, pointsFieldTitle))
	shape = readConditionRow(screen, row)
	c.Equal(1, shape.decimals, "a number field is compared against a number")
	c.Equal(0, shape.strings, "and against no text")
	c.Equal(shape.children, shape.columns, "the row's layout must hold one column per child")

	// Move this field's comparison off its default as well, so that the criteria the field change resets is not the
	// text one alone.
	choosePopupItem(t, screen, dialogWnd, conditionComparisonPopup(t, screen, row), 1)
	c.Equal(1, readConditionRow(screen, row).comparison, "the comparison must be the one just chosen")

	// A text field is compared against text, with the criteria reset rather than carried over from the field before.
	choosePopupItem(t, screen, dialogWnd, fieldPopup, traitFilterFieldIndex(t, nameFieldTitle))
	shape = readConditionRow(screen, row)
	c.Equal(1, shape.strings, "a text field is compared against text")
	c.Equal(0, shape.decimals, "and against no number")
	c.Equal(0, shape.comparison, "changing the field resets the criteria, so the comparison goes back to the first")
	c.Equal(shape.children, shape.columns, "the row's layout must hold one column per child")

	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the editor has been dismissed")
	c.Equal(0, len(savedFilters()), "a canceled editor must save nothing")
	c.Nil(readListFilterState(screen, d).selected, "and must leave the list with the quick filter")
}

// TestListFilterChangesReachOtherDockablesHeadless verifies that a change made to the saved filters through one
// list's popup reaches every other open list of the same type: a renamed filter that is in force there keeps its
// place under its new name, a new filter appears in its popup, and a deleted filter that was in force there hands the
// list back to the quick filter. Each time, the other list's toolbar is laid out again, which shows in the width of
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
	choosePopupItem(t, screen, wnd, second.savedFilters.popup, 2)
	choosePopupItemDirectly(screen, first, 2)
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
	editIndex, _ := popupItemIndexFromEnd(screen, first.savedFilters.popup, 2)
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
	newIndex, _ := popupItemIndexFromEnd(screen, first.savedFilters.popup, 3)
	choosePopupItemDirectly(screen, first, newIndex)
	var titles []string
	screen.Do(func() { titles = popupItemTitles(second.savedFilters.popup) })
	c.Equal([]string{
		"Quick Filter", separatorTitle, "A much longer filter name", "Mind", separatorTitle, newFilterItemTitle,
		editFilterItemTitle, deleteFilterItemTitle,
	}, titles, "the popup in front must list the new filter")
	state = readListFilterState(screen, second)
	c.True(mental == state.selected, "the list in front must keep the very filter it had in force")
	c.Equal("Mind", state.popupText, "and show it")
	widthAfterAdd := popupWidth(screen, second)
	c.True(widthAfterAdd > widthBefore, "the popup in front must have been laid out again to fit the new item")

	// Delete the filter in force through the list behind, which put the new one in force when it created it, so it
	// has to be put back first. The list in front, still on the deleted filter, falls back to the quick filter.
	choosePopupItemDirectly(screen, first, 3) // "Mind"
	swapForTest(t, &confirmFilterDeletion, func(_ string) bool { return true })
	deleteIndex, _ := popupItemIndexFromEnd(screen, first.savedFilters.popup, 1)
	choosePopupItemDirectly(screen, first, deleteIndex)
	state = readListFilterState(screen, second)
	c.Nil(state.selected, "the list in front must fall back to the quick filter once its filter is gone")
	c.Equal("Quick Filter", state.popupText, "and its popup must say so")
	c.True(state.fieldEnabled, "and its field must be usable again")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown there again")
	c.Nil(readListFilterState(screen, first).selected, "the list behind must be on the quick filter as well")

	// Delete the remaining filter too. The popup in front shrinks back, having only the quick filter and the commands
	// left to fit.
	choosePopupItemDirectly(screen, first, 2) // "A much longer filter name"
	deleteIndex, _ = popupItemIndexFromEnd(screen, first.savedFilters.popup, 1)
	choosePopupItemDirectly(screen, first, deleteIndex)
	screen.Do(func() { titles = popupItemTitles(second.savedFilters.popup) })
	c.Equal([]string{"Quick Filter", separatorTitle, newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle},
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
