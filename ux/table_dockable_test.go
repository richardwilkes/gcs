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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newFilterTestTrait returns a trait with the given name and tags for the filtering tests to work with.
func newFilterTestTrait(name string, tags ...string) *gurps.Trait {
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = name
	trait.Tags = tags
	return trait
}

// newFilterTestTraitDockable returns a trait library list dockable holding three traits with distinct names and tags,
// enough for the filtering to be told apart row by row. Building the toolbar reaches for the bindable actions, so they
// are registered first.
func newFilterTestTraitDockable() *TableDockable[*gurps.Trait] {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	return NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{
		newFilterTestTrait("Acute Vision", "Physical"),
		newFilterTestTrait("Combat Reflexes", "Mental"),
		newFilterTestTrait("Fur", "Physical", "Exotic"),
	})
}

// visibleTraitNames returns the names of the rows the table is showing, which are the rows that passed the filter when
// one is applied.
func visibleTraitNames(d *TableDockable[*gurps.Trait]) []string {
	rows := d.table.RootRows()
	names := make([]string, len(rows))
	for i, row := range rows {
		names[i] = row.Data().NameWithReplacements()
	}
	return names
}

// TestTableDockableRebuildKeepsFilter verifies that a rebuild puts the rows it produced back through the filter, rather
// than showing everything again.
func TestTableDockableRebuildKeepsFilter(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable()
	c.Equal(3, len(visibleTraitNames(d)), "every trait is shown before any filtering")

	d.filterField.SetText("fur")
	c.True(d.table.IsFiltered(), "typing in the content filter must filter the table")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "only the matching trait may be shown")

	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after a rebuild")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "the rebuilt rows must be filtered the same way")

	// A rebuild follows a change to the data, and the rows it produces have to go through the filter as well, so a
	// newly added trait that matches has to show up.
	d.provider.SetRootData(append(d.provider.RootData(), newFilterTestTrait("Furry Coat", "Physical")))
	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after the data changed")
	c.Equal([]string{"Fur", "Furry Coat"}, visibleTraitNames(d), "the added trait must be filtered along with the rest")
}

// TestTableDockableSavedFilterDisablesQuickFilter verifies that putting a saved filter in force takes over the list and
// blanks the quick filter's field, and that dropping it hands the list back.
func TestTableDockableSavedFilterDisablesQuickFilter(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	f := gurps.NewListFilter("Combat")
	condition := gurps.NewFilterCondition(f.Root, "name")
	condition.Text.Compare = criteria.ContainsText
	condition.Text.Qualifier = "combat"
	f.Root.Children = append(f.Root.Children, condition)
	gurps.GlobalSettings().AddListFilter(gurps.ListFilterKeyForExtension(gurps.TraitsExt), f)

	d := newFilterTestTraitDockable()
	d.filterField.SetText("fur")
	d.chooseFilter(f)
	c.Equal(f, d.selectedFilter, "the saved filter must be the one in force")
	c.False(d.filterField.Enabled(), "the quick filter's field must be unusable while a saved filter is in force")
	c.Equal("", d.filterField.Text(), "the quick filter's text must be cleared, so the two cannot disagree")
	c.True(d.table.IsFiltered(), "the saved filter must filter the table")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "only the traits the saved filter accepts may be shown")

	d.chooseFilter(nil)
	c.Nil(d.selectedFilter, "the list must go back to the quick filter")
	c.True(d.filterField.Enabled(), "the quick filter's field must be usable again")
	c.False(d.table.IsFiltered(), "the emptied quick filter shows everything")
	c.Equal(3, len(visibleTraitNames(d)), "every trait is shown again")
}

// TestTableDockableQuickFilterMatchesTags verifies that the quick filter looks at the tags column, which the tag popup
// it replaced used to be needed for.
func TestTableDockableQuickFilterMatchesTags(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable()
	d.filterField.SetText("mental")
	c.True(d.table.IsFiltered(), "typing in the content filter must filter the table")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait tagged \"Mental\" must be kept")
}

// TestTableDockableFilterPopupRelayoutsToolbar verifies that changing the saved filters through the popup marks the
// popup and everything above it, up to the dockable, for layout. The popup is as wide as its widest item, so a
// created, renamed or deleted filter changes its size, and the toolbar has to be laid out again to show it properly.
func TestTableDockableFilterPopupRelayoutsToolbar(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "A filter with a name long enough to widen the popup"
			return true
		})
	d := newFilterTestTraitDockable()
	c.NotNil(d.filterPopup, "a trait list must offer saved filters")
	chain := []*unison.Panel{d.filterPopup.AsPanel()}
	for p := d.filterPopup.Parent(); p != nil; p = p.Parent() {
		chain = append(chain, p)
	}
	c.True(len(chain) >= 3, "the popup must sit in a toolbar within the dockable")
	for _, p := range chain {
		p.NeedsLayout = false
	}

	d.filterPopup.ChoiceMadeCallback(d.filterPopup, d.filterPopup.ItemCount()-3, "")
	c.NotNil(d.selectedFilter, "the new filter must be in force")
	for _, p := range chain {
		c.True(p.NeedsLayout, "%T must be marked for layout after the popup's items changed", p.Self)
	}
}
