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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

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

// TestListTagFilterChoosesTags verifies that the tag filter sits between the quick filter and the saved filters and
// offers the list's tags, that a plain click chooses just one tag while a Shift-click or a menu command click adds or
// takes off one, that a row has to carry every chosen tag and pass the quick filter too, that Any Tag drops them all,
// and that a chosen tag no trait carries any more is dropped.
func TestListTagFilterChoosesTags(t *testing.T) {
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

// TestListTagFilterIgnoresCase verifies that tags differing only in case are offered once, and that a chosen tag stays
// chosen, and is saved as the list now spells it, when the spelling it is offered under changes.
func TestListTagFilterIgnoresCase(t *testing.T) {
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

// TestListTagFilterOffersTagNamedAnyTag verifies that a tag bearing the title of the Any Tag entry is offered and
// chosen as a tag, and that a tag holding the value of one of the popup's own entries is never offered.
func TestListTagFilterOffersTagNamedAnyTag(t *testing.T) {
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

// TestListTagFilterHeadless drives the tag filter of a trait list inside a headless workspace, where the modifier keys
// are read from the window as they are for a person's clicks: a click chooses a tag, and a Shift-click on another adds
// it, so a trait has to carry both.
func TestListTagFilterHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	d := openListFilterTraitDockable(t, screen)
	var physical, exotic int
	screen.Do(func() {
		physical = d.tagFilter.popup.IndexOfItem("physical")
		exotic = d.tagFilter.popup.IndexOfItem("exotic")
	})
	read := func() (checked, names []string, text string) {
		screen.Do(func() { checked, names, text = checkedTags(d), visibleTraitNames(d), d.tagFilter.popup.Text() })
		return checked, names, text
	}

	uxtest.ChoosePopupItem(t, screen, wnd, d.tagFilter.popup, physical)
	checked, names, _ := read()
	c.Equal([]string{"Physical"}, checked, "the click chooses the tag")
	c.Equal([]string{"Acute Vision"}, names, "only the trait tagged Physical is shown")

	screen.Click(screen.PanelCenter(d.tagFilter.popup))
	var item *unison.Panel
	screen.Do(func() {
		if items := uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd)); exotic >= 0 && exotic < len(items) {
			item = items[exotic]
		}
	})
	if item == nil {
		t.Fatal("the tag filter's menu has no Exotic item")
	}
	screen.ClickWith(screen.PanelCenter(item), unison.ButtonLeft, mod.Shift)
	checked, names, text := read()
	c.Equal([]string{"Exotic", "Physical"}, checked, "the Shift-click adds the tag")
	c.Equal("Multiple", text, "the popup says that more than one tag is chosen")
	c.Equal(0, len(names), "no trait carries both tags")
}
