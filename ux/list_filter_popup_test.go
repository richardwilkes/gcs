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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// listFilterPopupTestKey is the list type key the saved filter popup tests work with, the one trait lists use.
const listFilterPopupTestKey = "adq"

// separatorTitle stands in for a separator when the popup's items are listed out, since a separator has no title of
// its own but still occupies an index.
const separatorTitle = "<separator>"

// The indexes of the saved filter popup's No Filter entry and its first saved filter, after the three commands, a
// separator, No Filter and another separator. They are spelled out rather than taken from the popup, so that the tests
// check where the popup puts them.
const (
	popupNoneIndex       = 4
	popupFirstSavedIndex = 6
)

// listFilterPopupHarness drives a saved filter popup the way a list dockable would, recording every filter the popup
// puts in force so that a test can see what the user's choice produced.
type listFilterPopupHarness struct {
	popup   *listFilterPopup
	current *gurps.ListFilter
	chosen  []*gurps.ListFilter
}

// newListFilterPopupHarness builds a popup over a temporary set of saved filters holding one filter per name given, and
// points the global settings at a file in the test's own directory so that saving them touches nothing the user owns.
func newListFilterPopupHarness(t *testing.T, names ...string) *listFilterPopupHarness {
	t.Helper()
	uxtest.SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	for _, name := range names {
		gurps.GlobalSettings().AddListFilter(listFilterPopupTestKey, gurps.NewListFilter(name))
	}
	h := &listFilterPopupHarness{}
	h.popup = newListFilterPopup(listFilterPopupSpec{
		key:     listFilterPopupTestKey,
		fields:  filterFieldInfos(gurps.TraitFilterFields(), nil),
		current: func() *gurps.ListFilter { return h.current },
		choose: func(f *gurps.ListFilter) {
			h.current = f
			h.chosen = append(h.chosen, f)
		},
	})
	return h
}

// choose picks the item at the given index the way the popup's menu does when the user clicks one.
func (h *listFilterPopupHarness) choose(index int) {
	item, _ := h.popup.popup.ItemAt(index)
	h.popup.popup.ChoiceMadeCallback(h.popup.popup, index, item)
}

// lastChosen returns the filter most recently put in force, and false if none ever was.
func (h *listFilterPopupHarness) lastChosen() (*gurps.ListFilter, bool) {
	if len(h.chosen) == 0 {
		return nil, false
	}
	return h.chosen[len(h.chosen)-1], true
}

// savedFilters returns the saved filters for the list type the tests use.
func savedFilters() []*gurps.ListFilter {
	return gurps.GlobalSettings().ListFiltersFor(listFilterPopupTestKey)
}

// popupItemTitles returns the title of every item in the popup, in order, with separators standing out so that their
// positions can be checked too.
func popupItemTitles(p *unison.PopupMenu[string]) []string {
	titles := make([]string, p.ItemCount())
	for i := range titles {
		if item, ok := p.ItemAt(i); ok {
			titles[i] = item
		} else {
			titles[i] = separatorTitle
		}
	}
	return titles
}

// TestListFilterPopupItems verifies the items the popup holds, and their state, with no saved filters, one, and
// several.
func TestListFilterPopupItems(t *testing.T) {
	for _, one := range []struct {
		name  string
		saved []string
		want  []string
	}{
		{
			name: "no saved filters",
			want: []string{"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "No Filter"},
		},
		{
			name:  "one saved filter",
			saved: []string{"Cheap"},
			want: []string{
				"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "No Filter", separatorTitle, "Cheap",
			},
		},
		{
			// The saved filters arrive sorted by name, ignoring case, whatever order they were added in.
			name:  "several saved filters",
			saved: []string{"Gamma", "alpha", "Beta"},
			want: []string{
				"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "No Filter", separatorTitle, "alpha",
				"Beta", "Gamma",
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			h := newListFilterPopupHarness(t, one.saved...)
			p := h.popup
			c.Equal(one.want, popupItemTitles(p.popup), "the popup must hold these items, in this order")

			c.Equal(len(one.want), len(p.filters), "the filters must run parallel to the items")
			saved := savedFilters()
			for i := range p.filters {
				if j := i - popupFirstSavedIndex; j >= 0 && j < len(saved) {
					c.True(saved[j] == p.filters[i], "the item at %d must stand for the saved filter at %d", i, j)
				} else {
					c.Nil(p.filters[i], "the item at %d stands for no filter", i)
				}
			}

			c.Equal(0, p.newIndex, "New Filter… is the first item")
			c.Equal(1, p.editIndex, "Edit Filter… is the second")
			c.Equal(2, p.deleteIndex, "Delete Filter… is the third")

			c.Equal(popupNoneIndex, p.popup.SelectedIndex(), "No Filter starts out selected")
			c.True(p.popup.ItemEnabledAt(p.newIndex), "New Filter… is always available")
			c.False(p.popup.ItemEnabledAt(p.editIndex), "Edit Filter… needs a saved filter in force")
			c.False(p.popup.ItemEnabledAt(p.deleteIndex), "Delete Filter… needs a saved filter in force")
			_, chosen := h.lastChosen()
			c.False(chosen, "filling the popup in must not put a filter in force")
		})
	}
}

// TestListFilterPopupChooseFilter verifies that picking a saved filter puts it in force and turns on the commands that
// act on it.
func TestListFilterPopupChooseFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	p := h.popup
	h.choose(popupFirstSavedIndex + 1) // "Beta"

	c.Equal(popupFirstSavedIndex+1, p.popup.SelectedIndex(), "the chosen filter must become the selection")
	chosen, ok := h.lastChosen()
	c.True(ok, "choosing a filter must put one in force")
	c.True(savedFilters()[1] == chosen, "the filter put in force must be the one that was picked")
	c.True(p.popup.ItemEnabledAt(p.editIndex), "Edit Filter… must be available with a filter in force")
	c.True(p.popup.ItemEnabledAt(p.deleteIndex), "Delete Filter… must be available with a filter in force")

	h.choose(popupNoneIndex)
	chosen, ok = h.lastChosen()
	c.True(ok, "going back to No Filter must be reported")
	c.Nil(chosen, "No Filter is reported as no filter at all")
	c.False(p.popup.ItemEnabledAt(p.editIndex), "Edit Filter… must be off again")
	c.False(p.popup.ItemEnabledAt(p.deleteIndex), "Delete Filter… must be off again")
}

// TestListFilterPopupNewFilter verifies that accepting the editor for a new filter saves it and puts it in force.
func TestListFilterPopupNewFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha")
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, except *gurps.ListFilter) bool {
			c.Nil(except, "a new filter has no name of its own to keep")
			filter.Name = "Zeta"
			return true
		})
	h.choose(h.popup.newIndex)

	saved := savedFilters()
	c.Equal(2, len(saved), "the new filter must have been saved")
	c.Equal("Zeta", saved[1].Name, "the name the editor set must have been kept")
	chosen, ok := h.lastChosen()
	c.True(ok, "a new filter must be put in force")
	c.True(saved[1] == chosen, "the filter put in force must be the new one")
	c.Equal(popupFirstSavedIndex+1, h.popup.popup.SelectedIndex(),
		"the new filter's item, not the command, must be the selection")

	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, _ *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool { return false })
	h.choose(h.popup.newIndex)
	c.Equal(2, len(savedFilters()), "a canceled editor must save nothing")
	chosen, _ = h.lastChosen()
	c.True(saved[1] == chosen, "a canceled editor must leave the filter in force alone")
}

// TestListFilterPopupEditFilter verifies that accepting the editor copies the clone that was edited into the saved
// filter, which keeps its identity so that every dockable with it in force keeps it, and puts the list through it
// again even though the name, and so the selection, did not move.
func TestListFilterPopupEditFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha")
	original := savedFilters()[0]
	h.choose(popupFirstSavedIndex) // "alpha"
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, except *gurps.ListFilter) bool {
			c.True(original == except, "the filter being edited may keep its own name")
			c.True(original != filter, "a clone must be edited, so that a cancel changes nothing")
			filter.Root.Children = append(filter.Root.Children, gurps.NewFilterCondition(filter.Root, "name"))
			return true
		})
	h.choose(h.popup.editIndex)

	saved := savedFilters()
	c.Equal(1, len(saved), "editing a filter must not add one")
	c.True(original == saved[0], "the saved filter must keep its identity, since other dockables may hold it")
	c.Equal("alpha", saved[0].Name, "the name must be unchanged")
	c.Equal(1, len(saved[0].Root.Children), "the edit must have been copied into it")
	condition, isCondition := saved[0].Root.Children[0].(*gurps.FilterCondition)
	c.True(isCondition, "the copied node must be the condition the editor added")
	if isCondition {
		c.Equal("name", condition.Field, "the condition the editor added must be the one the saved filter now holds")
	}
	chosen, ok := h.lastChosen()
	c.True(ok, "an edited filter must be put in force again, since its contents changed")
	c.True(original == chosen, "the saved filter must be the one in force")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(), "its item must be the selection")

	// Renaming moves the filter among its siblings, so the list has to be sorted again.
	h = newListFilterPopupHarness(t, "alpha", "gamma")
	h.choose(popupFirstSavedIndex) // "alpha"
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "omega"
			return true
		})
	h.choose(h.popup.editIndex)
	c.Equal([]string{"gamma", "omega"}, listFilterNames(savedFilters()), "a renamed filter must be re-sorted")
	c.Equal(popupFirstSavedIndex+1, h.popup.popup.SelectedIndex(),
		"the renamed filter's new item must be the selection")
	c.Equal("omega", h.popup.popup.Text(), "and the popup must show the new name")
}

// listFilterNames returns the names of the filters, in order.
func listFilterNames(filters []*gurps.ListFilter) []string {
	names := make([]string, len(filters))
	for i, f := range filters {
		names[i] = f.Name
	}
	return names
}

// TestListFilterPopupDeleteFilter verifies that the filter in force is only removed once the user confirms it, and
// that the list is left with no saved filter when it is.
func TestListFilterPopupDeleteFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	original := savedFilters()[0]
	h.choose(popupFirstSavedIndex) // "alpha"

	confirm := false
	uxtest.SwapForTest(t, &confirmFilterDeletion, func(name string) bool {
		c.Equal("alpha", name, "the filter in force is the one being deleted")
		return confirm
	})
	h.choose(h.popup.deleteIndex)
	c.Equal(2, len(savedFilters()), "a refused deletion must remove nothing")
	chosen, _ := h.lastChosen()
	c.True(original == chosen, "a refused deletion must leave the filter in force alone")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(), "a refused deletion must leave the selection alone")

	confirm = true
	h.choose(h.popup.deleteIndex)
	saved := savedFilters()
	c.Equal(1, len(saved), "the filter must have been removed")
	c.Equal("Beta", saved[0].Name, "the other filter must have been left alone")
	chosen, _ = h.lastChosen()
	c.Nil(chosen, "the list must be left with no saved filter")
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "No Filter must become the selection")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must be off again")
}

// TestListFilterPopupCommandsNeedAFilterInForce verifies that Edit and Delete do nothing at all while no saved filter
// is in force. They are turned off in that state, but the popup is driven by callbacks that a command could still
// reach, so neither may put up its dialog or touch the saved filters.
func TestListFilterPopupCommandsNeedAFilterInForce(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha")
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, _ *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			c.Fatal("the editor must not be put up with no filter in force")
			return false
		})
	uxtest.SwapForTest(t, &confirmFilterDeletion, func(_ string) bool {
		c.Fatal("a deletion must not be confirmed with no filter in force")
		return false
	})

	h.choose(h.popup.editIndex)
	h.choose(h.popup.deleteIndex)

	c.Equal([]string{"alpha"}, listFilterNames(savedFilters()), "the saved filters must have been left alone")
	_, chosen := h.lastChosen()
	c.False(chosen, "neither command may put a filter in force")
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "No Filter must stay selected")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must still be off")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.deleteIndex), "Delete Filter… must still be off")
}

// TestListFilterPopupShowUpdatesItems verifies that opening the popup brings its items up to date, and that it drops
// the saved filter when the one that was in force has vanished behind its back, which is what happens when another
// dockable showing the same list type deleted it.
func TestListFilterPopupShowUpdatesItems(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	original := savedFilters()[0]
	h.choose(popupFirstSavedIndex) // "alpha"
	count := len(h.chosen)

	h.popup.popup.WillShowMenuCallback(h.popup.popup)
	c.Equal(count, len(h.chosen), "the filter in force is still there, so showing the popup must put nothing in force")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(), "the filter in force must stay selected")
	c.True(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must still be available")

	gurps.GlobalSettings().RemoveListFilter(listFilterPopupTestKey, original)
	h.popup.popup.WillShowMenuCallback(h.popup.popup)
	c.Equal(count+1, len(h.chosen), "the loss of the filter in force must be reported")
	chosen, _ := h.lastChosen()
	c.Nil(chosen, "the list must be left with no saved filter")
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "No Filter must become the selection")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must be off with no filter in force")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.deleteIndex), "Delete Filter… must be off with no filter in force")
	c.False(slices.Contains(popupItemTitles(h.popup.popup), "alpha"),
		"the filter that vanished must be gone from the items")
}

// TestListFilterPopupFilterNamedNone verifies that a saved filter that happens to bear the same name as the popup's
// own No Filter entry is still tracked, since the popup goes by index and identity rather than by name.
func TestListFilterPopupFilterNamedNone(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "No Filter", "Zeta")
	c.Equal([]string{
		"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "No Filter", separatorTitle, "No Filter", "Zeta",
	}, popupItemTitles(h.popup.popup),
		"a saved filter may bear the same name as the No Filter entry")

	h.choose(popupFirstSavedIndex)
	chosen, ok := h.lastChosen()
	c.True(ok, "the saved filter must be put in force")
	c.True(savedFilters()[0] == chosen, "the saved filter, not No Filter, must be in force")
	c.False(h.popup.rebuildItems(), "the saved filter is still there")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(), "the saved filter's own item must stay selected")
}

// TestListFilterPopupFixesWidth verifies that the popup's button keeps the same width whatever its filter names, that
// the button shows ten Ms whole and cuts anything wider to as much of its start as fits, with the ellipsis counted in
// the width and no space before it, and that the tooltip leads with a filter name only while the button cuts it short.
func TestListFilterPopupFixesWidth(t *testing.T) {
	c := check.New(t)
	long := "Mental or Social Advantages Worth Keeping"
	h := newListFilterPopupHarness(t, long, "Short")
	popup := h.popup.popup
	_, uncapped, _ := popup.DefaultSizes(geom.Size{})
	_, pref, _ := popup.Sizes(geom.Size{})
	c.True(pref.Width < uncapped.Width, "the long name may not widen the button to fit it")
	_, short, _ := newListFilterPopupHarness(t, "Short").popup.popup.Sizes(geom.Size{})
	c.Equal(short.Width, pref.Width, "the button keeps its width whatever the filter names")

	ten := strings.Repeat("M", 10)
	c.Equal(ten, popupFittedText(popup, ten), "ten Ms are shown whole")
	c.Equal(strings.Repeat("M", 9)+"…", popupFittedText(popup, ten+"M"), "eleven are cut to fit the ellipsis")
	c.Equal("Mental…", popupFittedText(popup, "Mental"+strings.Repeat(" ", 200)+"x"),
		"no space is left before the ellipsis")

	h.choose(popupFirstSavedIndex)
	if _, cut := h.popup.button.textCut(); !cut {
		t.Fatal("the button must cut the long name short")
	}
	c.Contains(tooltipText(popup.Tooltip), long, "the tooltip leads with the long name in full")
	h.choose(popupFirstSavedIndex + 1)
	c.NotContains(tooltipText(popup.Tooltip), long, "the tooltip drops the long name once another is chosen")
	c.NotContains(tooltipText(popup.Tooltip), "Short", "a name the button shows whole is kept out of the tooltip")
}

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
// against the qualifier, and returns it. A text or list field is what the tests use, so the text criterion is the one
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

// TestListFilterPopupAppliesSavedFilterHeadless drives the saved filter popup of a trait list inside a headless
// workspace: choosing a saved filter narrows the list and leaves the quick filter's field usable, what is typed there
// narrows the list further while the saved filter is in force, choosing No Filter drops the saved filter alone, and the
// quick filter searches the tags column as well as the name.
func TestListFilterPopupAppliesSavedFilterHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
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
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "No Filter", separatorTitle,
		"Mental",
	}, titles, "the commands lead, and the separator after No Filter puts the saved filter at index 6")

	state := readListFilterState(screen, d)
	c.Equal(listFilterHeadlessTraitNames, state.names, "every trait is shown before any filtering")
	c.Equal("No Filter", state.popupText, "no saved filter is in force to start with")
	c.True(state.fieldEnabled, "and the quick filter's field is usable")
	c.False(state.filtered, "and nothing is filtering the table")

	// Choose the saved filter. The item ahead of it is the separator, which occupies an index of its own.
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, popupFirstSavedIndex)
	state = readListFilterState(screen, d)
	c.Equal([]string{"Combat Reflexes"}, state.names, "only the traits the saved filter accepts may be shown")
	c.True(mental == state.selected, "the saved filter itself must be the one in force")
	c.Equal("Mental", state.popupText, "and the popup must show it")
	c.True(state.fieldEnabled, "the quick filter's field must stay usable while a saved filter is in force")
	c.True(state.filtered, "the saved filter must filter the table")
	uxtest.CaptureScreen(t, c, screen, "list_filter_applied")

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
	uxtest.CaptureScreen(t, c, screen, "list_filter_combined")

	// Choose No Filter, which drops the saved filter and leaves the quick filter's text to filter the list on its own.
	// Only one trait is tagged "Physical" and no trait's name holds the word, so keeping exactly that one shows the
	// quick filter searching the tags column.
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, popupNoneIndex)
	state = readListFilterState(screen, d)
	c.Nil(state.selected, "the saved filter must have been dropped")
	c.Equal("physical", state.fieldText, "dropping the saved filter must leave the quick filter's text alone")
	c.True(state.fieldEnabled, "and its field usable")
	c.True(state.filtered, "the quick filter must filter the table on its own")
	c.Equal([]string{"Acute Vision"}, state.names, "the quick filter searches the tags column, not just the name")
}

// TestListFilterDeleteAsksAndFallsBackHeadless drives the Delete Filter… command inside a headless workspace: it asks
// before removing anything, a refusal leaves the filter in force, and confirming removes it and leaves the list with no
// saved filter.
func TestListFilterDeleteAsksAndFallsBackHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)

	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, popupFirstSavedIndex)
	c.True(mental == readListFilterState(screen, d).selected,
		"the saved filter itself must be in force before it is deleted")
	var itemCount int
	screen.Do(func() { itemCount = d.savedFilters.popup.ItemCount() })
	deleteIndex := popupItemIndexOf(t, screen, d.savedFilters.popup, deleteFilterItemTitle)
	c.Equal(2, deleteIndex, "Delete Filter… is the third item")

	// Refuse the confirmation. The prompt names the filter in force, and turning it down removes nothing.
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, deleteIndex)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var prompt string
	screen.Do(func() { prompt = strings.Join(uxtest.LabelTexts(dialogWnd.Content()), "\n") })
	c.Contains(prompt, `Delete the filter 'Mental'?`, "the prompt must name the filter in force")
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	c.Equal(1, len(savedFilters()), "a refused deletion must remove nothing")
	state := readListFilterState(screen, d)
	c.True(mental == state.selected, "and must leave the filter itself in force")
	c.Equal([]string{"Combat Reflexes"}, state.names, "so the list is still filtered by it")

	// Confirm it the second time around. The filter goes away and the list is left with no saved filter.
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, deleteIndex)
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
	c.Equal("No Filter", state.popupText, "which the popup must show")
	c.False(state.filtered, "with the quick filter empty, nothing filters the table")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown again")
	c.Equal([]string{
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "No Filter",
	}, titles, "the deleted filter and the separator that set the saved filters apart must both be gone")
	c.Equal(itemCount-2, len(titles), "so the popup holds two items fewer than it did")
}

// TestListFilterEditCancelHeadless edits a saved filter through the real dialog and cancels it, checking that the
// saved filter is left as it was.
func TestListFilterEditCancelHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	before := gurps.Hash64(mental)
	d := openListFilterTraitDockable(t, screen)
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, popupItemIndexOf(t, screen, d.savedFilters.popup, "Mental"))
	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup,
		popupItemIndexOf(t, screen, d.savedFilters.popup, editFilterItemTitle))
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	p := dialogFilterPanel(t, screen, dialogWnd)
	screen.Type("Changed")
	screen.Do(func() {
		menuAction(p.treeAddEntries(p.filter.Root, treeRootPath), "New Filter")()
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

// TestListFilterChangesReachOtherDockablesHeadless verifies that a change made to the saved filters through one
// list's popup reaches every other open list of the same type: a renamed filter that is in force there keeps its
// place under its new name, a new filter appears in its popup, and a deleted filter that was in force there leaves the
// list with no saved filter. Each time, the other list's popup keeps its width. The two lists share a tab group, so
// only the second, in front, can be clicked; the first is driven through its popup's callback directly. The editor is
// stubbed, since it is the popups' bookkeeping that is under test here.
func TestListFilterChangesReachOtherDockablesHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	mental := seedListFilter("Mental", "tags", "Mental")
	first := openListFilterTraitDockable(t, screen)
	second := openListFilterTraitDockable(t, screen)
	uxtest.ChoosePopupItem(t, screen, wnd, second.savedFilters.popup, popupFirstSavedIndex)
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex)
	state := readListFilterState(screen, second)
	c.True(mental == state.selected, "the saved filter itself must be in force in the list in front")
	c.Equal([]string{"Combat Reflexes"}, state.names, "and filtering it")
	c.True(mental == readListFilterState(screen, first).selected, "and in force in the list behind")

	// Rename the filter through the list behind. The one in front keeps it in force, under the new name.
	uxtest.SwapForTest(t, &showFilterEditor,
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

	// Create a filter through the list behind. The popup in front lists it without growing to fit its long name, while
	// the list keeps what it had in force.
	uxtest.SwapForTest(t, &showFilterEditor,
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
		newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "No Filter", separatorTitle,
		"A much longer filter name", "Mind",
	}, titles, "the popup in front must list the new filter")
	state = readListFilterState(screen, second)
	c.True(mental == state.selected, "the list in front must keep the very filter it had in force")
	c.Equal("Mind", state.popupText, "and show it")
	c.Equal(widthBefore, popupWidth(screen, second), "the popup in front must keep its width")

	// Delete the filter in force through the list behind, which put the new one in force when it created it, so it
	// has to be put back first. The list in front, still on the deleted filter, falls back to No Filter.
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex+1) // "Mind"
	uxtest.SwapForTest(t, &confirmFilterDeletion, func(_ string) bool { return true })
	deleteIndex := popupItemIndexOf(t, screen, first.savedFilters.popup, deleteFilterItemTitle)
	choosePopupItemDirectly(screen, first, deleteIndex)
	state = readListFilterState(screen, second)
	c.Nil(state.selected, "the list in front must fall back to no saved filter once its filter is gone")
	c.Equal("No Filter", state.popupText, "and its popup must say so")
	c.Equal(listFilterHeadlessTraitNames, state.names, "so every trait is shown there again")
	c.Nil(readListFilterState(screen, first).selected, "the list behind must be on no saved filter as well")

	// Delete the remaining filter too. The popup in front still keeps its width.
	choosePopupItemDirectly(screen, first, popupFirstSavedIndex) // "A much longer filter name"
	deleteIndex = popupItemIndexOf(t, screen, first.savedFilters.popup, deleteFilterItemTitle)
	choosePopupItemDirectly(screen, first, deleteIndex)
	screen.Do(func() { titles = popupItemTitles(second.savedFilters.popup) })
	c.Equal([]string{newFilterItemTitle, editFilterItemTitle, deleteFilterItemTitle, separatorTitle, "No Filter"},
		titles, "the popup in front must have lost the deleted filter")
	c.Equal(widthBefore, popupWidth(screen, second), "the popup in front must still keep its width")
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

// TestListFilterPopupCutsNameHeadless verifies, with the popup drawn in a headless workspace, that a filter name too
// long for the button is cut short there while the menu it opens still lists it in full.
func TestListFilterPopupCutsNameHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	long := "Mental or Social Advantages Worth Keeping"
	seedListFilter(long, "tags", "Mental")
	d := openListFilterTraitDockable(t, screen)

	uxtest.ChoosePopupItem(t, screen, wnd, d.savedFilters.popup, popupFirstSavedIndex)
	var cut bool
	screen.Do(func() { _, cut = d.savedFilters.button.textCut() })
	c.True(cut, "the button cuts the long name short")

	screen.Click(screen.PanelCenter(d.savedFilters.popup))
	chooseOpenMenuItem(t, screen, wnd, long) // Fails unless the menu lists the long name in full.
}
