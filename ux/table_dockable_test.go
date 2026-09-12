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
// enough for the filtering to be told apart row by row. The global settings are pointed at a file in the test's own
// directory and given an empty set of saved filters, holding only the ones passed in, so that the saved filter popup
// is never built out of whatever the developer running the tests happens to have saved. Building the toolbar reaches
// for the bindable actions, so they are registered first.
func newFilterTestTraitDockable(t *testing.T, filters ...*gurps.ListFilter) *TableDockable[*gurps.Trait] {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	swapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	for _, f := range filters {
		gurps.GlobalSettings().AddListFilter(gurps.ListFilterKeyForExtension(gurps.TraitsExt), f)
	}
	return NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{
		newFilterTestTrait("Acute Vision", "Physical"),
		newFilterTestTrait("Combat Reflexes", "Mental"),
		newFilterTestTrait("Fur", "Physical", "Exotic"),
	})
}

// newNameContainsFilter returns a saved filter with the given name that keeps the rows whose own name contains text.
func newNameContainsFilter(name, text string) *gurps.ListFilter {
	return newContainsFilter(name, "name", text)
}

// newContainsFilter returns a saved filter with the given name that keeps the rows whose field with the given key
// contains text.
func newContainsFilter(name, fieldKey, text string) *gurps.ListFilter {
	f := gurps.NewListFilter(name)
	condition := gurps.NewFilterCondition(f.Root, fieldKey)
	condition.Text.Compare = criteria.ContainsText
	condition.Text.Qualifier = text
	f.Root.Children = append(f.Root.Children, condition)
	return f
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
	d := newFilterTestTraitDockable(t)
	c.Equal(3, len(visibleTraitNames(d)), "every trait is shown before any filtering")

	d.filterField.SetText("fur")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "only the matching trait may be shown")

	// The matching trait is renamed so that it no longer matches, which the rows the rebuild produces have to be put
	// through the filter to notice. A rebuild that left the rows that last passed the filter in place would go on
	// showing it under its new name.
	d.provider.RootData()[2].Name = "Pelt"
	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after a rebuild")
	c.Equal(0, len(visibleTraitNames(d)), "the renamed trait no longer matches, so nothing may be shown")

	d.provider.RootData()[2].Name = "Fur"
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

// TestTableDockableNewItemIsDisabledWhileFiltered verifies that a library list turns its new-item commands off while
// a filter is hiding part of the list, since creating an item inserts a row and a filtered table may not have its rows
// modified. Both the quick filter and a saved filter have to turn them off, and clearing either has to turn them back
// on.
func TestTableDockableNewItemIsDisabledWhileFiltered(t *testing.T) {
	c := check.New(t)
	f := newNameContainsFilter("Combat", "combat")
	d := newFilterTestTraitDockable(t, f)
	ids := []int{NewTraitItemID, NewTraitContainerItemID}
	for _, id := range ids {
		c.True(d.AsPanel().CanPerformCmd(nil, id), "an unfiltered library list must be able to create items")
	}

	d.filterField.SetText("fur")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	for _, id := range ids {
		c.False(d.AsPanel().CanPerformCmd(nil, id), "a list filtered by the quick filter must not offer new items")
	}

	d.filterField.SetText("")
	c.False(d.table.IsFiltered(), "emptying the quick filter must show everything")
	for _, id := range ids {
		c.True(d.AsPanel().CanPerformCmd(nil, id), "clearing the quick filter must make the commands available again")
	}

	d.chooseFilter(f)
	c.True(d.table.IsFiltered(), "the saved filter must filter the table")
	for _, id := range ids {
		c.False(d.AsPanel().CanPerformCmd(nil, id), "a list filtered by a saved filter must not offer new items")
	}

	d.chooseFilter(nil)
	c.False(d.table.IsFiltered(), "dropping the saved filter must show everything")
	for _, id := range ids {
		c.True(d.AsPanel().CanPerformCmd(nil, id), "dropping the saved filter must make the commands available again")
	}
	c.Equal(3, len(d.provider.RootData()), "no item may have been created along the way")
}

// TestTableDockableSavedFilterAndQuickFilterCombine verifies that a saved filter and the quick filter work together:
// putting a saved filter in force leaves the quick filter's field usable and its text in place, a row has to pass both
// to be shown, and each can be cleared on its own, leaving the other to filter the list alone.
func TestTableDockableSavedFilterAndQuickFilterCombine(t *testing.T) {
	c := check.New(t)
	physical := newContainsFilter("Physical", "tags", "physical")
	d := newFilterTestTraitDockable(t, physical)
	d.filterField.SetText("fur")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "the quick filter alone keeps the one trait named for it")

	d.chooseFilter(physical)
	c.True(physical == d.selectedFilter, "the saved filter must be the one in force")
	c.True(d.filterField.Enabled(), "the quick filter's field must stay usable while a saved filter is in force")
	c.Equal("fur", d.filterField.Text(), "the quick filter's text must be kept, since it still applies")
	c.True(d.table.IsFiltered(), "the two together must filter the table")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "only the trait that passes both may be shown")

	d.filterField.SetText("vision")
	c.Equal([]string{"Acute Vision"}, visibleTraitNames(d),
		"changing the quick filter while a saved filter is in force must narrow that filter's rows afresh")

	d.filterField.SetText("combat")
	c.True(d.table.IsFiltered(), "the two together still filter the table")
	c.Equal(0, len(visibleTraitNames(d)),
		"a trait that passes the quick filter but not the saved filter may not be shown")

	d.filterField.SetText("")
	c.True(d.table.IsFiltered(), "the saved filter still filters the table on its own")
	c.Equal([]string{"Acute Vision", "Fur"}, visibleTraitNames(d), "every trait the saved filter accepts is shown")

	d.filterField.SetText("fur")
	d.chooseFilter(nil)
	c.Nil(d.selectedFilter, "no saved filter may be in force")
	c.Equal("fur", d.filterField.Text(), "dropping the saved filter must leave the quick filter's text alone")
	c.True(d.table.IsFiltered(), "the quick filter still filters the table on its own")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "every trait the quick filter accepts is shown")

	d.filterField.SetText("")
	c.False(d.table.IsFiltered(), "with neither in play, nothing filters the table")
	c.Equal(3, len(visibleTraitNames(d)), "every trait is shown again")
}

// TestTableDockableQuickFilterMatchesTags verifies that the quick filter looks at the tags column, which the tag popup
// it replaced used to be needed for.
func TestTableDockableQuickFilterMatchesTags(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	d.filterField.SetText("mental")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait tagged \"Mental\" must be kept")
}

// TestTableDockableFilterPopupRelayoutsToolbar verifies that changing the saved filters through the popup marks the
// popup and everything above it, up to the dockable, for layout. The popup is as wide as its widest item, so a
// created, renamed or deleted filter changes its size, and the toolbar has to be laid out again to show it properly.
func TestTableDockableFilterPopupRelayoutsToolbar(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			filter.Name = "A filter with a name long enough to widen the popup"
			return true
		})
	d := newFilterTestTraitDockable(t)
	c.NotNil(d.savedFilters.popup, "a trait list must offer saved filters")
	chain := []*unison.Panel{d.savedFilters.popup.AsPanel()}
	for p := d.savedFilters.popup.Parent(); p != nil; p = p.Parent() {
		chain = append(chain, p)
	}
	c.True(len(chain) >= 3, "the popup must sit in a toolbar within the dockable")
	for _, p := range chain {
		p.NeedsLayout = false
	}

	d.savedFilters.popup.ChoiceMadeCallback(d.savedFilters.popup, d.savedFilters.popup.ItemCount()-3, "")
	c.NotNil(d.selectedFilter, "the new filter must be in force")
	for _, p := range chain {
		c.True(p.NeedsLayout, "%T must be marked for layout after the popup's items changed", p.Self)
	}
}

// TestTableDockableJumpToSearchFilter verifies that the command that jumps to the quick filter's field is available
// whether or not a saved filter is in force, since the field is usable either way.
func TestTableDockableJumpToSearchFilter(t *testing.T) {
	c := check.New(t)
	f := newNameContainsFilter("Combat", "combat")
	d := newFilterTestTraitDockable(t, f)
	c.True(d.AsPanel().CanPerformCmd(nil, JumpToSearchFilterItemID),
		"the quick filter's field must be reachable with no saved filter in force")

	d.chooseFilter(f)
	c.True(d.AsPanel().CanPerformCmd(nil, JumpToSearchFilterItemID),
		"the field is still usable while a saved filter is in force, so it must still be reachable")

	d.chooseFilter(nil)
	c.True(d.AsPanel().CanPerformCmd(nil, JumpToSearchFilterItemID),
		"dropping the saved filter must leave the field reachable")

	// The command asks for the focus, which a dockable that is in no window has nobody to ask.
	c.NotPanics(func() { d.AsPanel().PerformCmd(nil, JumpToSearchFilterItemID) },
		"performing the command outside of a window must not panic")
	c.False(d.filterField.Focused(), "there is no window, so nothing can take the focus")
}

// newFilterTestWeapon returns a melee weapon with the given usage, which is one of the columns the quick filter looks
// at.
func newFilterTestWeapon(usage string) *gurps.Weapon {
	w := gurps.NewWeapon(nil, true)
	w.Usage = usage
	return w
}

// TestTableDockableWithoutFilterKeyOmitsSavedFilters verifies that a list type with no filter key -- the weapon lists,
// which are never shown in a library list dockable -- gets no saved filter popup, yet still filters through the quick
// filter and takes the calls a popup would otherwise make without one.
func TestTableDockableWithoutFilterKeyOmitsSavedFilters(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	lists := &listsForTest{melee: []*gurps.Weapon{newFilterTestWeapon("Punch"), newFilterTestWeapon("Kick")}}
	d := NewTableDockable("test.wpn", ".wpn", NewWeaponsProvider(lists, true, false),
		func(_ string) error { return nil })
	c.Equal("", d.provider.FilterKey(), "the weapon lists have no saved filters of their own")
	c.Nil(d.savedFilters, "a list type with no filter key must get no saved filter popup, nor anything driving one")
	c.Nil(d.provider.FilterFields(), "nor any fields for a saved filter to test")
	for _, popup := range panelsOfType[*unison.PopupMenu[string]](d.AsPanel()) {
		c.NoPrefix(tooltipText(popup.Tooltip), "Saved Filters", "no popup in the toolbar may be the saved filter popup")
	}

	d.filterField.SetText("punch")
	c.True(d.table.IsFiltered(), "the quick filter must work without a saved filter popup")
	c.Equal(1, len(d.table.RootRows()), "only the matching weapon may be shown")

	d.filterField.SetText("")
	c.False(d.table.IsFiltered(), "emptying the quick filter must show everything")
	c.Equal(2, len(d.table.RootRows()), "every weapon is shown again")

	c.NotPanics(func() {
		d.chooseFilter(nil)
		d.listFiltersChanged(listFilterPopupTestKey, nil)
		d.Rebuild(false)
	}, "a list with no saved filter popup must take the filter calls all the same")
}

// filterFieldsCounter wraps a table provider to count how often the fields a saved filter may test are asked for,
// which the dockable does once each time it puts the rows through a saved filter.
type filterFieldsCounter struct {
	TableProvider[*gurps.Trait]
	calls int
}

func (p *filterFieldsCounter) FilterFields() []*gurps.FilterField[*gurps.Trait] {
	p.calls++
	return p.TableProvider.FilterFields()
}

// TestTableDockableChoosingSavedFilterAppliesOnce verifies that putting a saved filter in force while the quick filter
// holds text puts the rows through the saved filter once, with the quick filter's text narrowing them in the same
// pass, since each pass walks every row and syncs and sorts the table.
func TestTableDockableChoosingSavedFilterAppliesOnce(t *testing.T) {
	c := check.New(t)
	f := newNameContainsFilter("Combat", "combat")
	d := newFilterTestTraitDockable(t, f)
	counter := &filterFieldsCounter{TableProvider: d.provider}
	d.provider = counter
	d.filterField.SetText("fur")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "only the trait the quick filter accepts may be shown")
	c.Equal(0, counter.calls, "the quick filter has no saved filter to look the fields up for")

	d.chooseFilter(f)
	c.Equal(1, counter.calls, "the saved filter must be applied exactly once")
	c.Equal("fur", d.filterField.Text(), "the quick filter's text must have been kept")
	c.Equal(0, len(visibleTraitNames(d)), "no trait passes both the saved filter and the quick filter")

	d.chooseFilter(f)
	c.Equal(2, counter.calls, "choosing the saved filter again applies it once more")

	d.filterField.SetText("reflexes")
	c.Equal(3, counter.calls, "changing the quick filter puts the rows through the saved filter once more as well")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait that passes both is shown")
}

// TestTableDockableFilterKeepsHierarchy verifies that a filtered list keeps its hierarchy: a matching row is shown
// beneath the containers that hold it, which are shown open regardless of their own open state and are dimmed when they
// did not match themselves, and clearing the filter puts the list back the way it was.
func TestTableDockableFilterKeepsHierarchy(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	swapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	swapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Exotic Features"
	container.Tags = []string{"Exotic"}
	vision := gurps.NewTrait(nil, container, false)
	vision.Name = "Acute Vision"
	fur := gurps.NewTrait(nil, container, false)
	fur.Name = "Fur"
	container.Children = []*gurps.Trait{vision, fur}
	container.SetOpen(false)
	d := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{
		container,
		newFilterTestTrait("Combat Reflexes", "Mental"),
	})
	c.Equal(2, d.table.LastRowIndex()+1, "the closed container hides its children before any filtering")

	d.filterField.SetText("fur")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.Equal([]string{"Exotic Features"}, visibleTraitNames(d), "the matching trait's container must be the only root")
	c.Equal(2, d.table.LastRowIndex()+1, "the container and the matching trait beneath it must be shown")
	containerRow := d.table.RowFromIndex(0)
	furRow := d.table.RowFromIndex(1)
	c.Equal("Exotic Features", containerRow.Data().NameWithReplacements())
	c.Equal("Fur", furRow.Data().NameWithReplacements())
	c.True(d.table.IsFilterContextRow(containerRow), "the container did not match, so it is only context")
	c.False(d.table.IsFilterContextRow(furRow), "the trait matched, so it is not context")
	c.True(containerRow.cellData(0, true).Dim, "a container shown only as context must be dimmed")
	c.False(containerRow.cellData(0, false).Dim, "sorting and matching must not see the dimming")
	c.False(furRow.cellData(0, true).Dim, "a matching trait must not be dimmed")
	c.False(container.IsOpen(), "showing the rows beneath a closed container must not open it")

	d.filterField.SetText("exotic")
	c.Equal([]string{"Exotic Features"}, visibleTraitNames(d))
	c.Equal(1, d.table.LastRowIndex()+1, "a matching container is shown without the children that did not match")
	c.False(d.table.IsFilterContextRow(d.table.RowFromIndex(0)), "the container matched, so it is not context")
	c.False(d.table.RowFromIndex(0).cellData(0, true).Dim, "a matching container must not be dimmed")

	d.filterField.SetText("")
	c.False(d.table.IsFiltered(), "emptying the quick filter must show everything")
	c.Equal(2, d.table.LastRowIndex()+1, "the closed container must hide its children again")
	c.False(container.IsOpen(), "the container must still be closed")
}
