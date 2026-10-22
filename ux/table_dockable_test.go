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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	traitcontainer "github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func newFilterTestTrait(name string, tags ...string) *gurps.Trait {
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = name
	trait.Tags = tags
	return trait
}

// newFilterTestTraitDockable returns a trait library list dockable holding three traits with distinct names and tags.
// The global settings are pointed at a file in the test's directory and hold only the saved filters passed in, so the
// saved filter popup never reflects the developer's own saved filters. Building the toolbar needs the bindable actions,
// so they are registered first.
func newFilterTestTraitDockable(t *testing.T, filters ...*gurps.ListFilter) *TableDockable[*gurps.Trait] {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
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

// visibleTraitNames returns the names of the table's root rows, which under a filter are those it keeps at the top
// level.
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

	// Rename the matching trait so it no longer matches. A rebuild that kept the rows that last passed the filter would
	// go on showing it under its new name.
	d.provider.RootData()[2].Name = "Pelt"
	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after a rebuild")
	c.Equal(0, len(visibleTraitNames(d)), "the renamed trait no longer matches, so nothing may be shown")

	d.provider.RootData()[2].Name = "Fur"
	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after a rebuild")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "the rebuilt rows must be filtered the same way")

	// A newly added trait that matches has to show up after a rebuild.
	d.provider.SetRootData(append(d.provider.RootData(), newFilterTestTrait("Furry Coat", "Physical")))
	d.Rebuild(false)
	c.True(d.table.IsFiltered(), "the filter must still be in force after the data changed")
	c.Equal([]string{"Fur", "Furry Coat"}, visibleTraitNames(d), "the added trait must be filtered along with the rest")
}

// TestTableDockableNewItemIsDisabledWhileFiltered verifies that a library list disables its new-item commands while the
// quick filter or a saved filter is in force, since a filtered table may not have rows inserted, and re-enables them
// when the filter is cleared.
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

// TestModifierLibraryNewItemIsDisabledWhileFiltered verifies that a modifier library turns its new-item commands, the
// one for a new choice included, off while the quick filter is hiding part of the list, as other library lists do.
func TestModifierLibraryNewItemIsDisabledWhileFiltered(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	traitMod := gurps.NewTraitModifier(nil, nil, false)
	traitMod.Name = "Accessibility"
	checkNewItemsOffWhileFiltered(c,
		NewTraitModifierTableDockable("test"+gurps.TraitModifiersExt, []*gurps.TraitModifier{traitMod}),
		NewTraitModifierItemID, NewTraitContainerModifierItemID, NewTraitModifierChoiceItemID)
	eqpMod := gurps.NewEquipmentModifier(nil, nil, false)
	eqpMod.Name = "Balanced"
	checkNewItemsOffWhileFiltered(c,
		NewEquipmentModifierTableDockable("test"+gurps.EquipmentModifiersExt, []*gurps.EquipmentModifier{eqpMod}),
		NewEquipmentModifierItemID, NewEquipmentContainerModifierItemID, NewEquipmentModifierChoiceItemID)
}

// checkNewItemsOffWhileFiltered checks that the library list offers each of the new-item commands only while its
// quick filter is empty.
func checkNewItemsOffWhileFiltered[T gurps.Node[T]](c check.Checker, d *TableDockable[T], ids ...int) {
	c.Helper()
	for _, id := range ids {
		c.True(d.AsPanel().CanPerformCmd(nil, id), "an unfiltered library list must be able to create items")
	}
	d.filterField.SetText("zzz")
	c.True(d.table.IsFiltered())
	for _, id := range ids {
		c.False(d.AsPanel().CanPerformCmd(nil, id), "a filtered library list must not offer new items")
	}
	d.filterField.SetText("")
	for _, id := range ids {
		c.True(d.AsPanel().CanPerformCmd(nil, id), "clearing the filter must make the commands available again")
	}
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

// TestTableDockableQuickFilterMatchesTags verifies that the quick filter looks at the tags column.
func TestTableDockableQuickFilterMatchesTags(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	d.filterField.SetText("mental")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait tagged \"Mental\" must be kept")
}

// TestTableDockableFilterSuggestionsFollowTheList verifies that the filter editor a list opens suggests the values the
// list holds when they are asked for, so that a change made to the list after the editor was given its fields shows up,
// that a field with nothing to suggest has no suggestions, and that the container type offers every type there is.
func TestTableDockableFilterSuggestionsFollowTheList(t *testing.T) {
	c := check.New(t)
	var fields []filterFieldInfo
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, _ *gurps.ListFilter, offered []filterFieldInfo, _ *gurps.ListFilter) bool {
			fields = offered
			return false
		})
	d := newFilterTestTraitDockable(t)
	d.savedFilters.newFilter(gurps.NewListFilter(""))
	info := func(key string) filterFieldInfo {
		for _, one := range fields {
			if one.key == key {
				return one
			}
		}
		t.Fatalf("the editor was given no %s field", key)
		return filterFieldInfo{}
	}
	suggested := func(key string) []string {
		if one := info(key); one.suggestions != nil {
			return one.suggestions(false)
		}
		return nil
	}
	c.Equal([]string{"Exotic", "Mental", "Physical"}, suggested("tags"))
	c.Nil(info("name").suggestions, "a name has nothing to suggest")
	types := make([]string, 0, len(traitcontainer.Types))
	for _, one := range traitcontainer.Types {
		types = append(types, one.String())
	}
	c.Equal(types, suggested("container_type"))

	d.provider.SetRootData(append(d.provider.RootData(), newFilterTestTrait("Night Vision", "physical", "Zoology")))
	c.Equal([]string{"Exotic", "Mental", "Physical", "Zoology"}, suggested("tags"),
		"the tags are those in the list as it is now")
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
	c.Nil(d.tagFilter, "nor a tag filter")
	c.Nil(d.provider.FilterFields(), "nor any fields for a saved filter to test")
	for _, popup := range uxtest.PanelsOfType[*unison.PopupMenu[string]](d.AsPanel()) {
		c.NoPrefix(tooltipText(popup.Tooltip), "Saved Filters", "no popup in the toolbar may be the saved filter popup")
		c.NoPrefix(tooltipText(popup.Tooltip), "Tag Filter", "nor the tag filter")
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
	uxtest.SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
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

// filterTestTraits returns the three traits newFilterTestTraitDockable fills its list with, failing the test unless
// each is the one expected.
func filterTestTraits(t *testing.T, d *TableDockable[*gurps.Trait]) (vision, reflexes, fur *gurps.Trait) {
	t.Helper()
	traits := d.provider.RootData()
	vision, reflexes, fur = traits[0], traits[1], traits[2]
	for _, one := range []struct {
		trait *gurps.Trait
		name  string
	}{{vision, "Acute Vision"}, {reflexes, "Combat Reflexes"}, {fur, "Fur"}} {
		if one.trait.Name != one.name {
			t.Fatalf("expected the trait %q, found %q", one.name, one.trait.Name)
		}
	}
	return vision, reflexes, fur
}

// chooseTag picks the tag filter's item with the given value the way its menu does when the user clicks it, with the
// given modifier keys held.
func chooseTag(t *testing.T, d *TableDockable[*gurps.Trait], value string, mods mod.Modifiers) {
	t.Helper()
	uxtest.SwapForTest(t, &tagFilterModifiers, func(*unison.PopupMenu[string]) mod.Modifiers { return mods })
	popup := d.tagFilter.popup
	popup.WillShowMenuCallback(popup)
	i := popup.IndexOfItem(value)
	if i == -1 || !popup.ItemEnabledAt(i) {
		t.Fatalf("the tag filter has no item %q that its menu would choose", value)
	}
	popup.ChoiceMadeCallback(popup, i, value)
}

// tagFilterItems prepares the tag filter's menu as opening it does, and returns what its items show.
func tagFilterItems(d *TableDockable[*gurps.Trait]) []string {
	d.tagFilter.popup.WillShowMenuCallback(d.tagFilter.popup)
	titles := popupItemTitles(d.tagFilter.popup)
	for i, value := range titles {
		if value != separatorTitle {
			titles[i] = d.tagFilter.title(value)
		}
	}
	return titles
}

// checkedTags returns what the tag filter's checked items show.
func checkedTags(d *TableDockable[*gurps.Trait]) []string {
	indexes := d.tagFilter.popup.SelectedIndexes()
	titles := make([]string, 0, len(indexes))
	for _, i := range indexes {
		value, _ := d.tagFilter.popup.ItemAt(i)
		titles = append(titles, d.tagFilter.title(value))
	}
	return titles
}

// swapFilterEditor has the filter editor apply change to the filter it is given and close with accept, returning the
// filter it was given through edited.
func swapFilterEditor(t *testing.T, accept bool, change func(f *gurps.ListFilter)) (edited **gurps.ListFilter) {
	t.Helper()
	edited = new(*gurps.ListFilter)
	uxtest.SwapForTest(t, &showFilterEditor,
		func(_, _ string, filter *gurps.ListFilter, _ []filterFieldInfo, _ *gurps.ListFilter) bool {
			*edited = filter
			change(filter)
			return accept
		})
	return edited
}

// conditionAt returns the condition at the given index of the filter's root, failing the test if there is none.
func conditionAt(t *testing.T, f *gurps.ListFilter, index int) *gurps.FilterCondition {
	t.Helper()
	if f == nil || index >= len(f.Root.Children) {
		t.Fatalf("the filter has no node at %d", index)
	}
	condition, ok := f.Root.Children[index].(*gurps.FilterCondition)
	if !ok {
		t.Fatalf("the node at %d is not a condition", index)
	}
	return condition
}

// TestTableDockableTagFilter verifies that the tag filter sits between the quick filter and the saved filters and
// offers the list's tags, that a plain click chooses just one tag while a Shift-click or a menu command click adds or
// takes off one, that a row has to carry every chosen tag and pass the quick filter too, that Any Tag drops them all,
// and that a chosen tag no trait carries any more is dropped.
func TestTableDockableTagFilter(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	_, reflexes, fur := filterTestTraits(t, d)
	toolbar := d.filterField.Parent().Children()
	c.True(slices.Equal([]*unison.Panel{
		d.filterField.AsPanel(), d.tagFilter.popup.AsPanel(),
		d.savedFilters.popup.AsPanel(),
	}, toolbar[len(toolbar)-3:]),
		"the tag filter sits between the quick filter and the saved filters")
	c.Equal([]string{"Any Tag", separatorTitle, "Exotic", "Mental", "Physical", separatorTitle, "Save Filter…"},
		tagFilterItems(d), "every tag in the list is offered, between Any Tag and Save Filter")
	c.Equal([]string{"Any Tag"}, checkedTags(d), "Any Tag is checked while no tag is chosen")
	c.False(d.tagFilter.popup.ItemEnabledAt(d.tagFilter.popup.IndexOfItem(saveFilterValue)),
		"with no tag chosen there is nothing to save")
	c.Contains(tooltipText(d.tagFilter.popup.Tooltip),
		"Shift-Click or "+strings.TrimSuffix(mod.OSMenuCommand().String(), "+")+"-Click",
		"the tooltip says how to choose more than one tag")

	chooseTag(t, d, "physical", 0)
	c.Equal([]string{"Acute Vision", "Fur"}, visibleTraitNames(d), "the traits tagged Physical are shown")
	chooseTag(t, d, "exotic", mod.Shift)
	c.Equal([]string{"Exotic", "Physical"}, checkedTags(d), "a Shift-click adds a tag")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "a trait has to carry every chosen tag")
	c.Contains(tooltipText(d.tagFilter.popup.Tooltip), "Physical\n● Exotic",
		"the tooltip lists the chosen tags in the order chosen")
	chooseTag(t, d, "mental", mod.OSMenuCommand())
	c.Equal([]string{"Exotic", "Mental", "Physical"}, checkedTags(d), "a menu command click adds a tag")
	chooseTag(t, d, "mental", mod.OSMenuCommand())
	c.Equal([]string{"Exotic", "Physical"}, checkedTags(d), "a menu command click on a chosen tag takes it off")
	chooseTag(t, d, "physical", mod.Shift)
	c.Equal([]string{"Exotic"}, checkedTags(d), "a Shift-click on a chosen tag takes it off")
	chooseTag(t, d, "physical", mod.Shift)

	chooseTag(t, d, "mental", 0)
	c.Equal([]string{"Mental"}, checkedTags(d), "a plain click chooses just the one tag")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait tagged Mental is shown")
	d.filterField.SetText("vision")
	c.Equal(0, len(visibleTraitNames(d)), "a trait has to pass the quick filter as well as carry the tags")
	d.filterField.SetText("")

	chooseTag(t, d, "mental", mod.Shift)
	c.Equal([]string{"Any Tag"}, checkedTags(d), "taking off the last tag checks Any Tag")
	c.False(d.table.IsFiltered(), "with no tag chosen, nothing filters the table")
	chooseTag(t, d, "mental", 0)
	chooseTag(t, d, anyTagValue, 0)
	c.Equal([]string{"Any Tag"}, checkedTags(d), "Any Tag drops every chosen tag")
	c.False(d.table.IsFiltered(), "with no tag chosen, nothing filters the table")

	chooseTag(t, d, "exotic", 0)
	chooseTag(t, d, "physical", mod.Shift)
	fur.Tags = []string{"Physical"}
	d.Rebuild(false)
	c.Equal([]string{"Physical"}, checkedTags(d), "a chosen tag no trait carries any more is dropped")
	c.Equal([]string{"Acute Vision", "Fur"}, visibleTraitNames(d), "the tag still carried alone filters the list")

	reflexes.Tags = []string{"Mental", "Rare"}
	d.Rebuild(false)
	chooseTag(t, d, "rare", 0)
	reflexes.Tags = []string{"Mental"}
	c.Equal([]string{"Any Tag", separatorTitle, "Mental", "Physical", separatorTitle, "Save Filter…"},
		tagFilterItems(d), "opening the menu drops a chosen tag no trait carries any more")
	c.False(d.table.IsFiltered(), "the list is filtered again without it")
}

// TestTableDockableTagFilterFixesWidth verifies that the tag filter's button keeps the same width whatever its tags.
func TestTableDockableTagFilterFixesWidth(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	vision, _, _ := filterTestTraits(t, d)
	popup := d.tagFilter.popup
	_, short, _ := popup.Sizes(geom.Size{})
	vision.Tags = []string{"Sensory Enhancements of Every Kind and Description"}
	d.Rebuild(false)
	popup.WillShowMenuCallback(popup)
	_, uncapped, _ := popup.DefaultSizes(geom.Size{})
	_, pref, _ := popup.Sizes(geom.Size{})
	c.True(pref.Width < uncapped.Width, "the long tag may not widen the button to fit it")
	c.Equal(short.Width, pref.Width, "the button keeps its width whatever the tags")
}

// TestTableDockableQuickFilterKeepsItsMinimumWidth verifies that the quick filter in a toolbar too narrow for its row
// shrinks no further than the room for quickFilterWidthSample, rather than out of sight.
func TestTableDockableQuickFilterKeepsItsMinimumWidth(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	toolbar := d.filterField.Parent()
	toolbar.SetFrameRect(geom.NewRect(0, 0, 100, 30))
	toolbar.ValidateLayout()
	sample := unison.NewText(quickFilterWidthSample, &unison.TextDecoration{Font: d.filterField.Font}).Width()
	width := d.filterField.FrameRect().Width
	c.True(width >= sample, "the quick filter is %v wide, less than the sample's %v", width, sample)
}

// TestTableDockableFilterPopupsTurnCompact verifies that the filter popups turn into icons while the toolbar is too
// narrow to show them in full, that the saved filter's name then leads its tooltip, and that they turn back once the
// toolbar is wide enough again.
func TestTableDockableFilterPopupsTurnCompact(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	toolbar := d.filterField.Parent()
	minSize, _, _ := toolbar.Sizes(geom.Size{})
	layOut := func(width float32) {
		toolbar.SetFrameRect(geom.NewRect(0, 0, width, 30))
		toolbar.ValidateLayout()
	}

	layOut(minSize.Width - 1)
	for _, b := range []*filterPopupButton{d.tagFilter.button, d.savedFilters.button} {
		c.True(b.compact, "%s must turn compact in a toolbar too narrow for it", b.popup.Accessibility.Name)
		size := b.popup.FrameRect().Size
		c.Equal(size.Height, size.Width, "%s must be a square icon", b.popup.Accessibility.Name)
	}
	c.Contains(tooltipText(d.savedFilters.popup.Tooltip), "No Filter", "the icon's tooltip names the saved filter")

	layOut(minSize.Width)
	c.False(d.tagFilter.button.compact, "the tag filter must turn back once the toolbar is wide enough")
	c.False(d.savedFilters.button.compact, "and so must the saved filters")
	c.NotContains(tooltipText(d.savedFilters.popup.Tooltip), "No Filter", "a name shown whole is kept out of the tooltip")
}

// TestTableDockableTagFilterIgnoresCase verifies that tags differing only in case are offered once, and that a chosen
// tag stays chosen, and is saved as the list now spells it, when the spelling it is offered under changes.
func TestTableDockableTagFilterIgnoresCase(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	vision, _, _ := filterTestTraits(t, d)
	vision.Tags = []string{"exotic"}
	d.Rebuild(false)
	c.Equal([]string{"Any Tag", separatorTitle, "exotic", "Mental", "Physical", separatorTitle, "Save Filter…"},
		tagFilterItems(d), "tags that differ only in case are offered once, as first found")
	chooseTag(t, d, "exotic", 0)
	c.Equal([]string{"Acute Vision", "Fur"}, visibleTraitNames(d), "the tag matches whatever its case")

	vision.Tags = []string{"Physical"}
	d.Rebuild(false)
	c.Equal([]string{"Exotic"}, checkedTags(d), "the chosen tag stays checked under its new spelling")
	edited := swapFilterEditor(t, false, func(*gurps.ListFilter) {})
	chooseTag(t, d, saveFilterValue, 0)
	c.Equal("Exotic", conditionAt(t, *edited, 0).Text.Qualifier, "the tag is saved as the list now spells it")
}

// TestTableDockableTagNamedAnyTag verifies that a tag bearing the title of the Any Tag entry is offered and chosen as a
// tag, and that a tag holding the value of one of the popup's own entries is never offered.
func TestTableDockableTagNamedAnyTag(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	_, reflexes, _ := filterTestTraits(t, d)
	reflexes.Tags = []string{"Any Tag", anyTagValue, saveFilterValue}
	d.Rebuild(false)
	c.Equal([]string{"Any Tag", separatorTitle, "Any Tag", "Exotic", "Physical", separatorTitle, "Save Filter…"},
		tagFilterItems(d), "the tag is offered beside the entry, and the tags holding the entries' values are not")

	chooseTag(t, d, "any tag", 0)
	c.Equal([]int{2}, d.tagFilter.popup.SelectedIndexes(), "the tag's own item is checked, not the Any Tag entry")
	c.Equal([]string{"Combat Reflexes"}, visibleTraitNames(d), "the trait carrying the tag is shown")
	chooseTag(t, d, anyTagValue, 0)
	c.False(d.table.IsFiltered(), "the Any Tag entry still drops the chosen tags")
}

// TestTableDockableTagFilterFollowsSavedFilter verifies that the tag filter offers only the tags of the traits the
// saved filter in force lets through, that editing or renaming that filter keeps the chosen tags it still offers, and
// that choosing another saved filter, or deleting the one in force, drops them.
func TestTableDockableTagFilterFollowsSavedFilter(t *testing.T) {
	c := check.New(t)
	fur := newNameContainsFilter("Fur", "fur")
	physical := newContainsFilter("Physical", "tags", "physical")
	d := newFilterTestTraitDockable(t, fur, physical)
	uxtest.SwapForTest(t, &confirmFilterDeletion, func(string) bool { return true })
	qualifier := func(text string) func(f *gurps.ListFilter) {
		return func(f *gurps.ListFilter) { conditionAt(t, f, 0).Text.Qualifier = text }
	}

	d.chooseFilter(fur)
	c.Equal([]string{"Any Tag", separatorTitle, "Exotic", "Physical", separatorTitle, "Save Filter…"},
		tagFilterItems(d), "only the tags of the traits the saved filter lets through are offered")
	chooseTag(t, d, "exotic", 0)
	swapFilterEditor(t, true, func(f *gurps.ListFilter) { f.Name = "Furry" })
	d.savedFilters.editFilter()
	c.Equal([]string{"Exotic"}, checkedTags(d), "renaming the saved filter keeps the chosen tags")
	swapFilterEditor(t, true, qualifier("u"))
	d.savedFilters.editFilter()
	c.Equal([]string{"Exotic"}, checkedTags(d), "editing the saved filter keeps a chosen tag it still offers")
	swapFilterEditor(t, true, qualifier("vision"))
	d.savedFilters.editFilter()
	c.Equal([]string{"Any Tag"}, checkedTags(d), "editing the saved filter drops a chosen tag it no longer offers")
	c.Equal([]string{"Acute Vision"}, visibleTraitNames(d), "the edited saved filter alone filters the list")

	chooseTag(t, d, "physical", 0)
	d.chooseFilter(physical)
	c.Equal([]string{"Any Tag"}, checkedTags(d), "another saved filter drops the chosen tags, though it offers them")
	chooseTag(t, d, "physical", 0)
	d.savedFilters.deleteFilter()
	c.Nil(d.selectedFilter, "deleting the saved filter in force leaves none")
	c.Equal([]string{"Any Tag"}, checkedTags(d), "deleting the saved filter in force drops the chosen tags")
}

// TestTableDockableEmptySavedFilterFilters verifies that a saved filter with no conditions, whether it requires all or
// any of them, still counts as filtering the list while it is in force, as it did before there were tags to choose.
func TestTableDockableEmptySavedFilterFilters(t *testing.T) {
	c := check.New(t)
	all := gurps.NewListFilter("All")
	anyOf := gurps.NewListFilter("Any")
	anyOf.Root.All = false
	d := newFilterTestTraitDockable(t, all, anyOf)
	for _, f := range []*gurps.ListFilter{all, anyOf} {
		d.chooseFilter(f)
		c.True(d.table.IsFiltered(), "the empty saved filter %q filters the list", f.Name)
		c.Equal(3, len(visibleTraitNames(d)), "the empty saved filter %q passes every trait", f.Name)
	}
}

// TestTableDockableSaveTagFilter verifies that Save Filter puts up the editor for a new filter that requires what the
// saved filter in force does and the chosen tags as well, and once it is accepted, saves it and puts it in force in
// their place.
func TestTableDockableSaveTagFilter(t *testing.T) {
	c := check.New(t)
	physical := newContainsFilter("Physical", "tags", "physical")
	d := newFilterTestTraitDockable(t, physical)
	d.chooseFilter(physical)
	chooseTag(t, d, "exotic", 0)
	key := gurps.ListFilterKeyForExtension(gurps.TraitsExt)

	edited := swapFilterEditor(t, false, func(f *gurps.ListFilter) { f.Name = "Exotic Physical" })
	chooseTag(t, d, saveFilterValue, 0)
	c.Equal(2, len((*edited).Root.Children), "the new filter holds the saved filter's condition and the tag's")
	c.Equal("physical", conditionAt(t, *edited, 0).Text.Qualifier, "the saved filter's condition comes first")
	c.Equal("Exotic", conditionAt(t, *edited, 1).Text.Qualifier, "the tag follows, spelled as the list spells it")
	c.Equal(1, len(gurps.GlobalSettings().ListFiltersFor(key)), "a canceled editor saves nothing")
	c.Equal([]string{"Exotic"}, checkedTags(d), "a canceled editor leaves the chosen tags alone")

	swapFilterEditor(t, true, func(f *gurps.ListFilter) { f.Name = "Exotic Physical" })
	chooseTag(t, d, saveFilterValue, 0)
	saved := gurps.GlobalSettings().ListFiltersFor(key)
	c.Equal(2, len(saved), "the new filter is saved")
	c.True(slices.Contains(saved, d.selectedFilter), "the new filter is put in force")
	c.Equal("Exotic Physical", d.selectedFilter.Name, "the filter in force is the new one")
	c.Equal([]string{"Any Tag"}, checkedTags(d), "the chosen tags are dropped, since the new filter holds them")
	c.Equal([]string{"Fur"}, visibleTraitNames(d), "the new filter shows what the two did together")
}

// TestTableDockableKeepsItsRowsAcrossAGlobalSettingsSync verifies that a change to the global sheet settings that can't
// alter a library list's columns or rows, reported without a full rebuild, leaves its rows in place and settles their
// heights for what the cells now show, rather than building every row and cell anew. Typing in a field of the global
// sheet settings reports such a change on every keystroke, and used to rebuild every open library list each time. A
// change to the defaults that calls for a full rebuild, such as applying the default attribute definitions, still
// rebuilds the list, and a change to one character's own settings is none of its business.
func TestTableDockableKeepsItsRowsAcrossAGlobalSettingsSync(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))
	global.Sheet.NotesDisplay = display.NotShown
	d := newFilterTestTraitDockable(t)
	row := d.table.RowFromIndex(0)
	c.NotNil(row, "the list must show its traits")
	row.Data().LocalNotes = "Sees well in dim light"
	height := d.table.RowHeights()[0]

	d.SheetSettingsUpdated(gurps.NewEntity(), true)
	c.True(d.table.RowFromIndex(0) == row, "another entity's settings change must not touch the list")
	c.Equal(height, d.table.RowHeights()[0], "another entity's settings change must not touch the list")

	global.Sheet.NotesDisplay = display.Inline
	d.SheetSettingsUpdated(nil, false)
	c.True(d.table.RowFromIndex(0) == row, "a change needing no full rebuild must keep the rows")
	c.True(d.table.RowHeights()[0] > height, "the row must be made tall enough for the notes it now shows")

	d.SheetSettingsUpdated(nil, true)
	c.True(d.table.RowFromIndex(0) != row, "a change needing a full rebuild must rebuild the list")
}
