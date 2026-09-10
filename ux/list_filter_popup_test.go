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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// listFilterPopupTestKey is the list type key the saved filter popup tests work with, the one trait lists use.
const listFilterPopupTestKey = "adq"

// separatorTitle stands in for a separator when the popup's items are listed out, since a separator has no title of
// its own but still occupies an index.
const separatorTitle = "<separator>"

// listFilterPopupHarness drives a saved filter popup the way a list dockable would, recording every filter the popup
// puts in force so that a test can see what the user's choice produced.
type listFilterPopupHarness struct {
	popup   *listFilterPopup
	current *gurps.ListFilter
	chosen  []*gurps.ListFilter
}

// newListFilterPopupHarness builds a popup over an empty, temporary set of saved filters holding one filter per name
// given, and points the global settings at a file in the test's own directory so that saving them touches nothing the
// user owns.
func newListFilterPopupHarness(t *testing.T, names ...string) *listFilterPopupHarness {
	t.Helper()
	swapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	for _, name := range names {
		gurps.GlobalSettings().AddListFilter(listFilterPopupTestKey, gurps.NewListFilter(name))
	}
	h := &listFilterPopupHarness{}
	h.popup = newListFilterPopup(listFilterPopupSpec{
		key:     listFilterPopupTestKey,
		fields:  filterFieldInfos(gurps.TraitFilterFields()),
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
			want: []string{"Quick Filter", separatorTitle, "New Filter…", "Edit Filter…", "Delete Filter…"},
		},
		{
			name:  "one saved filter",
			saved: []string{"Cheap"},
			want: []string{
				"Quick Filter", separatorTitle, "Cheap", separatorTitle, "New Filter…", "Edit Filter…",
				"Delete Filter…",
			},
		},
		{
			// The saved filters arrive sorted by name, ignoring case, whatever order they were added in.
			name:  "several saved filters",
			saved: []string{"Gamma", "alpha", "Beta"},
			want: []string{
				"Quick Filter", separatorTitle, "alpha", "Beta", "Gamma", separatorTitle, "New Filter…",
				"Edit Filter…", "Delete Filter…",
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
				if i > listFilterQuickIndex && i-2 >= 0 && i-2 < len(saved) {
					c.Equal(saved[i-2], p.filters[i], "the item at %d must stand for the saved filter at %d", i, i-2)
				} else {
					c.Nil(p.filters[i], "the item at %d stands for no filter", i)
				}
			}

			c.Equal(len(one.want)-3, p.newIndex, "New Filter… is the third item from the end")
			c.Equal(len(one.want)-2, p.editIndex, "Edit Filter… is the second item from the end")
			c.Equal(len(one.want)-1, p.deleteIndex, "Delete Filter… is the last item")

			c.Equal(listFilterQuickIndex, p.popup.SelectedIndex(), "the quick filter starts out selected")
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
	h.choose(3) // "Beta"

	c.Equal(3, p.popup.SelectedIndex(), "the chosen filter must become the selection")
	chosen, ok := h.lastChosen()
	c.True(ok, "choosing a filter must put one in force")
	c.Equal(savedFilters()[1], chosen, "the filter put in force must be the one that was picked")
	c.True(p.popup.ItemEnabledAt(p.editIndex), "Edit Filter… must be available with a filter in force")
	c.True(p.popup.ItemEnabledAt(p.deleteIndex), "Delete Filter… must be available with a filter in force")

	h.choose(listFilterQuickIndex)
	chosen, ok = h.lastChosen()
	c.True(ok, "going back to the quick filter must be reported")
	c.Nil(chosen, "the quick filter is reported as no filter at all")
	c.False(p.popup.ItemEnabledAt(p.editIndex), "Edit Filter… must be off again")
	c.False(p.popup.ItemEnabledAt(p.deleteIndex), "Delete Filter… must be off again")
}

// TestListFilterPopupNewFilter verifies that accepting the editor for a new filter saves it and puts it in force.
func TestListFilterPopupNewFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha")
	swapForTest(t, &showFilterEditor,
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
	c.Equal(saved[1], chosen, "the filter put in force must be the new one")
	c.Equal(3, h.popup.popup.SelectedIndex(), "the new filter's item, not the command, must be the selection")

	swapForTest(t, &showFilterEditor,
		func(_, _ string, _ *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool { return false })
	h.choose(h.popup.newIndex)
	c.Equal(2, len(savedFilters()), "a canceled editor must save nothing")
	chosen, _ = h.lastChosen()
	c.Equal(saved[1], chosen, "a canceled editor must leave the filter in force alone")
}

// TestListFilterPopupEditFilter verifies that accepting the editor copies the clone that was edited into the saved
// filter, which keeps its identity so that every dockable with it in force keeps it, and puts the list through it
// again even though the name, and so the selection, did not move.
func TestListFilterPopupEditFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha")
	original := savedFilters()[0]
	h.choose(2) // "alpha"
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, except *gurps.ListFilter) bool {
			c.Equal(original, except, "the filter being edited may keep its own name")
			c.NotEqual(original, filter, "a clone must be edited, so that a cancel changes nothing")
			filter.Root.Children = append(filter.Root.Children, gurps.NewFilterCondition(filter.Root, "name"))
			return true
		})
	h.choose(h.popup.editIndex)

	saved := savedFilters()
	c.Equal(1, len(saved), "editing a filter must not add one")
	c.Equal(original, saved[0], "the saved filter must keep its identity, since other dockables may hold it")
	c.Equal("alpha", saved[0].Name, "the name must be unchanged")
	c.Equal(1, len(saved[0].Root.Children), "the edit must have been copied into it")
	c.Equal(original.Root, original.Root.Children[0].ParentGroup(), "the copied tree must hang off the saved filter")
	chosen, ok := h.lastChosen()
	c.True(ok, "an edited filter must be put in force again, since its contents changed")
	c.Equal(original, chosen, "the saved filter must be the one in force")
	c.Equal(2, h.popup.popup.SelectedIndex(), "its item must be the selection")

	// Renaming moves the filter among its siblings, so the list has to be sorted again.
	h = newListFilterPopupHarness(t, "alpha", "gamma")
	h.choose(2) // "alpha"
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "omega"
			return true
		})
	h.choose(h.popup.editIndex)
	c.Equal([]string{"gamma", "omega"}, listFilterNames(savedFilters()), "a renamed filter must be re-sorted")
	c.Equal(3, h.popup.popup.SelectedIndex(), "the renamed filter's new item must be the selection")
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
// that the list goes back to the quick filter when it is.
func TestListFilterPopupDeleteFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	original := savedFilters()[0]
	h.choose(2) // "alpha"

	confirm := false
	swapForTest(t, &confirmFilterDeletion, func(name string) bool {
		c.Equal("alpha", name, "the filter in force is the one being deleted")
		return confirm
	})
	h.choose(h.popup.deleteIndex)
	c.Equal(2, len(savedFilters()), "a refused deletion must remove nothing")
	chosen, _ := h.lastChosen()
	c.Equal(original, chosen, "a refused deletion must leave the filter in force alone")
	c.Equal(2, h.popup.popup.SelectedIndex(), "a refused deletion must leave the selection alone")

	confirm = true
	h.choose(h.popup.deleteIndex)
	saved := savedFilters()
	c.Equal(1, len(saved), "the filter must have been removed")
	c.Equal("Beta", saved[0].Name, "the other filter must have been left alone")
	chosen, _ = h.lastChosen()
	c.Nil(chosen, "the list must go back to the quick filter")
	c.Equal(listFilterQuickIndex, h.popup.popup.SelectedIndex(), "the quick filter must become the selection")
	c.False(h.popup.popup.ItemEnabledAt(h.popup.editIndex), "Edit Filter… must be off again")
}

// TestListFilterPopupRebuildAfterExternalRemoval verifies that a rebuild reports the loss of a filter that was deleted
// behind the popup's back, which is what happens when another dockable showing the same list type deletes it.
func TestListFilterPopupRebuildAfterExternalRemoval(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "alpha", "Beta")
	original := savedFilters()[0]
	h.choose(2) // "alpha"
	c.False(h.popup.rebuildItems(), "nothing has changed, so the filter in force is still there")
	c.Equal(2, h.popup.popup.SelectedIndex(), "the filter in force must stay selected across a rebuild")

	gurps.GlobalSettings().RemoveListFilter(listFilterPopupTestKey, original)
	c.True(h.popup.rebuildItems(), "the filter in force is gone, so the rebuild must report it")
	c.Equal(listFilterQuickIndex, h.popup.popup.SelectedIndex(), "the selection must fall back to the quick filter")
	chosen, _ := h.lastChosen()
	c.Equal(original, chosen, "the rebuild itself must not put a filter in force; the caller does that")
}

// TestListFilterPopupFilterNamedQuickFilter verifies that a saved filter that happens to bear the same name as the
// popup's own first entry is still tracked, since the popup goes by index and identity rather than by name.
func TestListFilterPopupFilterNamedQuickFilter(t *testing.T) {
	c := check.New(t)
	h := newListFilterPopupHarness(t, "Quick Filter", "Zeta")
	c.Equal([]string{
		"Quick Filter", separatorTitle, "Quick Filter", "Zeta", separatorTitle, "New Filter…",
		"Edit Filter…", "Delete Filter…",
	}, popupItemTitles(h.popup.popup),
		"a saved filter may bear the same name as the quick filter")

	h.choose(2)
	chosen, ok := h.lastChosen()
	c.True(ok, "the saved filter must be put in force")
	c.Equal(savedFilters()[0], chosen, "the saved filter, not the quick filter, must be in force")
	c.False(h.popup.rebuildItems(), "the saved filter is still there")
	c.Equal(2, h.popup.popup.SelectedIndex(), "the saved filter's own item must stay selected")
}
