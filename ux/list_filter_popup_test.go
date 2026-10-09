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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// listFilterPopupTestKey is the list type key the saved filter popup tests work with, the one trait lists use.
const listFilterPopupTestKey = "adq"

// separatorTitle stands in for a separator when the popup's items are listed out, since a separator has no title of
// its own but still occupies an index.
const separatorTitle = "<separator>"

// The indexes of the saved filter popup's None entry and its first saved filter, after the three commands, a
// separator, None and another separator. They are spelled out rather than taken from the popup, so that the tests
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
			want: []string{"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "None"},
		},
		{
			name:  "one saved filter",
			saved: []string{"Cheap"},
			want: []string{
				"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "None", separatorTitle, "Cheap",
			},
		},
		{
			// The saved filters arrive sorted by name, ignoring case, whatever order they were added in.
			name:  "several saved filters",
			saved: []string{"Gamma", "alpha", "Beta"},
			want: []string{
				"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "None", separatorTitle, "alpha",
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

			c.Equal(popupNoneIndex, p.popup.SelectedIndex(), "None starts out selected")
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
	c.True(ok, "going back to None must be reported")
	c.Nil(chosen, "None is reported as no filter at all")
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
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "None must become the selection")
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
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "None must stay selected")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must still be off")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.deleteIndex), "Delete Filter… must still be off")
}

// TestListFilterPopupRebuildAfterExternalRemoval verifies that a rebuild reports the loss of a filter that was deleted
// behind the popup's back, which is what happens when another dockable showing the same list type deletes it.
func TestListFilterPopupRebuildAfterExternalRemoval(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	original := savedFilters()[0]
	h.choose(popupFirstSavedIndex) // "alpha"
	c.False(h.popup.rebuildItems(), "nothing has changed, so the filter in force is still there")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(),
		"the filter in force must stay selected across a rebuild")

	gurps.GlobalSettings().RemoveListFilter(listFilterPopupTestKey, original)
	c.True(h.popup.rebuildItems(), "the filter in force is gone, so the rebuild must report it")
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "the selection must fall back to None")
	chosen, _ := h.lastChosen()
	c.True(original == chosen, "the rebuild itself must not put a filter in force; the caller does that")
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
	c.Equal(popupNoneIndex, h.popup.popup.SelectedIndex(), "None must become the selection")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must be off with no filter in force")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.deleteIndex), "Delete Filter… must be off with no filter in force")
	c.False(slices.Contains(popupItemTitles(h.popup.popup), "alpha"),
		"the filter that vanished must be gone from the items")
}

// TestListFilterPopupFilterNamedNone verifies that a saved filter that happens to bear the same name as the popup's
// own None entry is still tracked, since the popup goes by index and identity rather than by name.
func TestListFilterPopupFilterNamedNone(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "None", "Zeta")
	c.Equal([]string{
		"New Filter…", "Edit Filter…", "Delete Filter…", separatorTitle, "None", separatorTitle, "None", "Zeta",
	}, popupItemTitles(h.popup.popup),
		"a saved filter may bear the same name as the None entry")

	h.choose(popupFirstSavedIndex)
	chosen, ok := h.lastChosen()
	c.True(ok, "the saved filter must be put in force")
	c.True(savedFilters()[0] == chosen, "the saved filter, not None, must be in force")
	c.False(h.popup.rebuildItems(), "the saved filter is still there")
	c.Equal(popupFirstSavedIndex, h.popup.popup.SelectedIndex(), "the saved filter's own item must stay selected")
}
