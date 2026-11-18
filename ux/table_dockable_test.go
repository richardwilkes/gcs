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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	traitcontainer "github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
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

// TestTableDockableFilterDisablesNewItemsAndShowsBanner verifies that while the quick filter or a saved filter is in
// force, a library list disables its new-item commands, since a filtered table may not have rows inserted, along with
// its hierarchy and note buttons (see TestTableDockableTogglesAreOffWhileFiltered), and puts the banner that says so
// between the toolbar and the table, and that as soon as the list shows everything again, the commands and buttons are
// re-enabled and the banner is taken away.
func TestTableDockableFilterDisablesNewItemsAndShowsBanner(t *testing.T) {
	c := check.New(t)
	f := newNameContainsFilter("Combat", "combat")
	d := newFilterTestTraitDockable(t, f)
	checkFiltered := func(filtered bool, what string) {
		t.Helper()
		c.Equal(filtered, d.table.IsFiltered(), "%s: whether the table is filtered", what)
		for _, id := range []int{NewTraitItemID, NewTraitContainerItemID} {
			c.Equal(!filtered, d.AsPanel().CanPerformCmd(nil, id), "%s: new items are offered only while unfiltered",
				what)
		}
		c.Equal(!filtered, d.hierarchyButton.Enabled(), "%s: the hierarchy button is on only while unfiltered", what)
		c.Equal(!filtered, d.noteToggleButton.Enabled(), "%s: the note button is on only while unfiltered", what)
		if filtered {
			c.Equal(1, d.IndexOfChild(d.filterBanner), "%s: the banner sits right after the toolbar", what)
			c.Equal(2, d.IndexOfChild(d.scroll), "%s: and right before the table", what)
		} else {
			c.Equal(-1, d.IndexOfChild(d.filterBanner), "%s: an unfiltered list has no banner", what)
		}
	}

	checkFiltered(false, "before any filtering")
	d.filterField.SetText("fur")
	checkFiltered(true, "typing in the quick filter")
	d.filterField.SetText("")
	checkFiltered(false, "emptying the quick filter")
	d.chooseFilter(f)
	checkFiltered(true, "choosing a saved filter")
	d.chooseFilter(nil)
	checkFiltered(false, "dropping the saved filter")
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
// rebuilds the list, while a change to one character's own settings, or to the default block layout or body type,
// which the list has no part in, is none of its business.
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

	d.SheetSettingsUpdated(gurps.NewEntity(), gurps.SheetSettingsAllParts, true)
	c.True(d.table.RowFromIndex(0) == row, "another entity's settings change must not touch the list")
	c.Equal(height, d.table.RowHeights()[0], "another entity's settings change must not touch the list")

	d.SheetSettingsUpdated(nil, gurps.SheetSettingsLayoutPart|gurps.SheetSettingsBodyTypePart, true)
	c.True(d.table.RowFromIndex(0) == row, "a change to the default layout or body type must not touch the list")
	c.Equal(height, d.table.RowHeights()[0], "a change to the default layout or body type must not touch the list")

	global.Sheet.NotesDisplay = display.Inline
	d.SheetSettingsUpdated(nil, gurps.SheetSettingsGeneralPart, false)
	c.True(d.table.RowFromIndex(0) == row, "a change needing no full rebuild must keep the rows")
	c.True(d.table.RowHeights()[0] > height, "the row must be made tall enough for the notes it now shows")

	d.SheetSettingsUpdated(nil, gurps.SheetSettingsAttributesPart, true)
	c.True(d.table.RowFromIndex(0) != row, "a change needing a full rebuild must rebuild the list")
}

// TestTableDockableLockState verifies that a list opened from a file starts locked only while the setting asks for it,
// that a new list starts unlocked whatever the setting, that the lock button toggles the lock and shows its state, and
// that opening the file again starts over from the setting.
func TestTableDockableLockState(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	p := filepath.Join(t.TempDir(), "library"+gurps.TraitsExt)
	c.NoError(gurps.SaveTraits([]*gurps.Trait{newFilterTestTrait("Alertness")}, p))
	open := func() *TableDockable[*gurps.Trait] {
		t.Helper()
		dockable, err := NewTraitTableDockableFromFile(p)
		c.NoError(err)
		d, ok := dockable.(*TableDockable[*gurps.Trait])
		if !ok {
			t.Fatal("a trait list file must open as a trait list")
		}
		return d
	}
	checkLocked := func(d *TableDockable[*gurps.Trait], want bool, what string) {
		t.Helper()
		c.Equal(want, d.locked, "%s: the lock", what)
		c.Equal(want, d.lockButtonGroup.Selected(d.lockButton), "%s: the button is selected while locked", what)
		c.Equal(!want, d.lockButton.HideBase, "%s: the button shows its base while locked", what)
	}

	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialListLocked, false)
	checkLocked(open(), false, "a file opened with the setting off")

	gurps.GlobalSettings().General.InitialListLocked = true
	d := open()
	checkLocked(d, true, "a file opened with the setting on")
	checkLocked(NewTraitTableDockable("new"+gurps.TraitsExt, nil), false, "a new list with the setting on")
	d.lockButton.Click()
	checkLocked(d, false, "a click on the button")
	d.lockButton.Click()
	checkLocked(d, true, "another click on the button")
	d.lockButton.Click()
	checkLocked(open(), true, "the file opened again after its list was unlocked")
}

// TestTableDockableLockedListRefusesChanges verifies that every command, drop and undo that would change a list's file
// is refused while the list is locked and offered again once it is unlocked, while what only reads the list, or writes
// it to another file, stays on, as does the table, so that its rows can still be selected and dragged out.
func TestTableDockableLockedListRefusesChanges(t *testing.T) {
	c := check.New(t)
	traits := newFilterTestTraitDockable(t)
	traits.table.SelectByIndex(0)
	traits.undoMgr.Add(&unison.UndoEdit[int]{ID: unison.NextUndoID(), EditName: "Change"})
	equipment := NewEquipmentTableDockable("test"+gurps.EquipmentExt,
		[]*gurps.Equipment{gurps.NewEquipment(nil, nil, false)})
	equipment.table.SelectByIndex(0)
	modifiers := NewTraitModifierTableDockable("test"+gurps.TraitModifiersExt,
		[]*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, true)})
	modifiers.table.SelectByIndex(0)
	modifierDrag := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}
	for _, one := range []struct {
		name        string
		list        interface{ setLocked(locked bool) }
		check       func() bool
		whileLocked bool
	}{
		{"New Trait", traits, func() bool { return traits.CanPerformCmd(nil, NewTraitItemID) }, false},
		{"Delete", traits, func() bool { return traits.CanPerformCmd(nil, unison.DeleteItemID) }, false},
		{"Organize Traits", traits, func() bool { return traits.CanPerformCmd(nil, OrganizeTraitsItemID) }, false},
		{"a modifier drop", traits, func() bool { return traits.table.CanAcceptDropCallback(modifierDrag) }, false},
		{"Undo", traits, func() bool {
			mgr := unison.UndoManagerFor(traits.table)
			return mgr != nil && mgr.CanUndo()
		}, false},
		{"Convert to Container", equipment, func() bool {
			return equipment.CanPerformCmd(nil, ConvertToContainerItemID)
		}, false},
		{"Increment Quantity", equipment, func() bool {
			return equipment.table.CanPerformCmd(nil, IncrementItemID)
		}, false},
		{"Convert to Choice", modifiers, func() bool {
			return modifiers.CanPerformCmd(nil, ConvertToChoiceContainerItemID)
		}, false},
		{"Open Editor", traits, func() bool { return traits.CanPerformCmd(nil, OpenEditorItemID) }, true},
		{"Save As", traits, func() bool { return traits.CanPerformCmd(nil, SaveAsItemID) }, true},
		{"the table", traits, traits.table.Enabled, true},
	} {
		c.True(one.check(), "%s: on while unlocked", one.name)
		one.list.setLocked(true)
		c.Equal(one.whileLocked, one.check(), "%s: while locked", one.name)
		one.list.setLocked(false)
		c.True(one.check(), "%s: on again once unlocked", one.name)
	}
}

// TestTableDockableLockLeavesScaleUndoable verifies that a list's scale, which isn't part of its file, stays undoable
// while the list is locked, and that once it is unlocked, Undo takes back the scale change made while locked before
// the one made before locking.
func TestTableDockableLockLeavesScaleUndoable(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	initial := d.scale
	undo := func(want int, what string) {
		t.Helper()
		mgr := unison.UndoManagerFor(d.table)
		c.True(mgr != nil && mgr.CanUndo(), "%s: Undo must be on", what)
		if mgr != nil {
			mgr.Undo()
		}
		c.Equal(want, d.scale, "%s: Undo must restore the scale", what)
	}
	d.PerformCmd(nil, Scale200ItemID)
	d.setLocked(true)
	d.PerformCmd(nil, Scale300ItemID)
	undo(200, "while locked")
	d.PerformCmd(nil, Scale400ItemID)
	d.setLocked(false)
	undo(200, "once unlocked")
	undo(initial, "after that")
}

// TestTableDockableLockedListKeepsItsRowOrder verifies that no sort can reorder a locked list's file: a press on a
// column header is refused, and clearing a filter while a sort is in force, which re-sorts an unlocked list, leaves the
// rows in their order. The sort is kept, so it applies again once the list is unlocked.
func TestTableDockableLockedListKeepsItsRowOrder(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t)
	names := func() []string {
		rows := d.provider.RootData()
		list := make([]string, 0, len(rows))
		for _, one := range rows {
			list = append(list, one.Name)
		}
		return list
	}
	col := d.table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	press := func() bool {
		req := accessibility.ActionRequest{Key: col, Action: accessibility.Press}
		return d.tableHeader.PerformAccessibilityAction(req)
	}
	refilter := func() {
		d.filterField.SetText("fur")
		d.filterField.SetText("")
	}
	unsorted := names()
	sorted := slices.Clone(unsorted)
	slices.Reverse(sorted)
	d.tableHeader.ColumnHeaders[col].SetSortState(unison.SortState{Order: 0, Sortable: true})

	d.setLocked(true)
	refilter()
	c.Equal(unsorted, names(), "clearing a filter must not re-sort a locked list")
	c.False(press(), "a header press must be refused while locked")
	c.Equal(unsorted, names(), "the refused press must leave the rows in their order")

	d.setLocked(false)
	refilter()
	c.Equal(sorted, names(), "clearing a filter must re-sort an unlocked list by the sort kept from before")
	c.True(press(), "a header press must be taken once unlocked")
	c.Equal(unsorted, names(), "pressing the sorted column must flip its direction")
}

// TestTableDockableLockingClosesEditors verifies that locking or unlocking a list closes the editors opened from it
// first: canceling the save prompt of one with changes leaves the list unlocked and the editor open, saving applies the
// changes before the list locks, and unlocking closes the read-only editor, so that the row's editor opens editable.
func TestTableDockableLockingClosesEditors(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	var d *TableDockable[*gurps.Trait]
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	trait := newFilterTestTrait("Alertness")
	isOpen := func(dockable unison.Dockable) bool {
		return slices.ContainsFunc(AllDockables(), func(one unison.Dockable) bool {
			return one.AsPanel().Self == dockable
		})
	}
	screen.Do(func() {
		d = NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{trait})
		DisplayNewDockable(d)
		e = EditTrait(d, trait)
		e.editorData.LocalNotes = "Keen"
	})
	// The click prompts, so it is posted rather than run through Do, which would wait for the prompt to return.
	clickLockAndAnswer := func(response int) {
		t.Helper()
		screen.Post(d.lockButton.Click)
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, response)))
	}

	clickLockAndAnswer(unison.ModalResponseCancel)
	screen.Do(func() {
		c.False(d.locked, "canceling the prompt must leave the list unlocked")
		c.False(d.lockButtonGroup.Selected(d.lockButton), "and the button unlatched")
		c.True(isOpen(e), "and the editor open")
		c.Equal("", trait.LocalNotes, "with its changes pending")
	})

	clickLockAndAnswer(unison.ModalResponseOK)
	screen.Do(func() {
		c.True(d.locked, "saving at the prompt must let the list lock")
		c.False(isOpen(e), "once the editor has closed")
		c.Equal("Keen", trait.LocalNotes, "having applied its changes")
		e = EditTrait(d, trait)
		c.True(e.readOnly, "an editor opened from a locked list must be read-only")
	})

	screen.Post(d.lockButton.Click)
	screen.Sync()
	screen.Do(func() {
		c.False(d.locked, "a click must unlock the list without a prompt")
		c.False(isOpen(e), "closing the read-only editor")
		c.False(EditTrait(d, trait).readOnly, "so that the row's editor opens editable")
	})
}

// listFilterHeadlessTraitNames are the names of the fixture's traits, in the order the list holds them, which is what
// the table shows whenever nothing is filtering it.
var listFilterHeadlessTraitNames = []string{"Acute Vision", "Combat Reflexes", "Fur"}

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

// TestTableDockableTogglesAreOffWhileFiltered verifies that a list dockable's hierarchy and note toggles, whose buttons
// are off while either kind of filter is applied (see TestTableDockableFilterDisablesNewItemsAndShowsBanner), leave the
// rows alone should they be reached all the same. A filtered table shows every container it keeps as open, so a toggle
// would silently change disclosure states that nothing shows until the filter is cleared.
func TestTableDockableTogglesAreOffWhileFiltered(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	first := newTestContainer("First", true)
	first.LocalNotes = "A note"
	second := newTestContainer("Second", false)
	dockable := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{first, second})

	dockable.filterField.SetText("first")
	c.True(dockable.table.IsFiltered(), "typing in the quick filter must filter the table")
	dockable.toggleHierarchy()
	c.True(first.IsOpen(), "the matched container must be left open")
	c.False(second.IsOpen(), "the unmatched container must be left closed")
	dockable.toggleNotes()
	c.Equal(1, dockable.FirstNoteState(), "the note must be left shown")

	dockable.filterField.SetText("")
	c.False(dockable.table.IsFiltered(), "emptying the quick filter must show everything")

	dockable.chooseFilter(newNameContainsFilter("First", "first"))
	c.True(dockable.table.IsFiltered(), "the saved filter must filter the table")
	dockable.toggleHierarchy()
	c.True(first.IsOpen(), "the matched container must be left open under the saved filter as well")
	c.False(second.IsOpen(), "the unmatched container must be left closed under the saved filter as well")

	dockable.chooseFilter(nil)
	c.False(dockable.table.IsFiltered(), "dropping the saved filter must show everything")
	dockable.toggleHierarchy()
	c.False(first.IsOpen(), "with no filter, the open first container sets the direction, so it must be closed")
	c.False(second.IsOpen(), "and the second one must be closed along with it")
}
