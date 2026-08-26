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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// organizeTraitsTestCategories lists the containers the "Organize Traits" command has to end up with, in the order it
// has to put them in, along with the one trait from the test list that belongs in each of them.
var organizeTraitsTestCategories = []struct {
	container string
	trait     string
}{
	{container: "Advantages", trait: "Combat Reflexes"},
	{container: "Perks", trait: "Alcohol Tolerance"},
	{container: "Disadvantages", trait: "Bad Temper"},
	{container: "Quirks", trait: "Nosy"},
	{container: "Features", trait: "Fur"},
	{container: "Languages", trait: "French"},
}

// newOrganizeTraitsTestList returns one non-container trait for each of the categories the command files traits into:
// a 10-point advantage, a 1-point perk, a -10-point disadvantage, a -1-point quirk, a 0-point feature and a trait
// tagged as a language. They are deliberately not in the order the command has to leave them in, so that organizing
// them has to do real work. Pass nil for the owner when the traits aren't backed by an entity.
func newOrganizeTraitsTestList(owner gurps.DataOwner) []*gurps.Trait {
	create := func(name string, points int, tags ...string) *gurps.Trait {
		trait := gurps.NewTrait(owner, nil, false)
		trait.Name = name
		trait.BasePoints = fxp.FromInteger(points)
		trait.Tags = tags
		return trait
	}
	return []*gurps.Trait{
		create("Nosy", -1),
		create("Combat Reflexes", 10),
		create("French", 2, "Language"),
		create("Fur", 0),
		create("Bad Temper", -10),
		create("Alcohol Tolerance", 1),
	}
}

// newTestTraitTableDockable returns a trait library list dockable holding the traits the "Organize Traits" tests work
// with. Building the toolbar reaches for the bindable actions, so they are registered first.
func newTestTraitTableDockable() *TableDockable[*gurps.Trait] {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	return NewTraitTableDockable("test"+gurps.TraitsExt, newOrganizeTraitsTestList(nil))
}

// checkOrganizedTraits verifies that the top-level list holds exactly the expected category containers, in order, each
// of them holding only the one trait that belongs in it.
func checkOrganizedTraits(c check.Checker, list []*gurps.Trait) {
	c.Equal(len(organizeTraitsTestCategories), len(list),
		"the top level must hold one container per category and nothing else, but holds %v", traitNames(list))
	if len(list) != len(organizeTraitsTestCategories) {
		return
	}
	for i, category := range organizeTraitsTestCategories {
		container := list[i]
		c.Equal(category.container, container.Name, "the container at index %d must be %q", i, category.container)
		c.True(container.Container(), "%q must be a container", category.container)
		children := container.NodeChildren()
		c.Equal(1, len(children), "%q must hold exactly one trait, but holds %v", category.container,
			traitNames(children))
		if len(children) == 1 {
			c.Equal(category.trait, children[0].Name, "%q must hold %q", category.container, category.trait)
		}
	}
}

// traitNames returns the names of the traits, for use in failure messages.
func traitNames(list []*gurps.Trait) []string {
	names := make([]string, 0, len(list))
	for _, one := range list {
		names = append(names, one.Name)
	}
	return names
}

// rootRowNames returns the names of the traits the table is showing at its top level.
func rootRowNames(table *unison.Table[*Node[*gurps.Trait]]) []string {
	rows := table.RootRows()
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Data().Name)
	}
	return names
}

// TestOrganizeTraitsOnSheet verifies that the command files a sheet's loose top-level traits into the standard set of
// category containers and reports the change.
func TestOrganizeTraitsOnSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = newOrganizeTraitsTestList(entity)
	sheet.Rebuild(true)
	// The traits were dropped straight into the entity rather than added through the UI, so the baseline the sheet
	// compares against is reset here. Otherwise the sheet would already count as modified and the check below that
	// organizing marks it as modified would pass no matter what the command did.
	sheet.hash = gurps.Hash64(entity)
	c.False(sheet.Modified(), "the sheet must start out unmodified")
	c.Equal(len(organizeTraitsTestCategories), sheet.Traits.Table.RootRowCount(),
		"the sheet must start out with each trait at the top level")

	sheet.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)

	checkOrganizedTraits(c, entity.Traits)
	c.Equal(len(organizeTraitsTestCategories), sheet.Traits.Table.RootRowCount(),
		"the list on screen must show one row per category container, but shows %v", rootRowNames(sheet.Traits.Table))
	c.True(sheet.Modified(), "organizing the traits must mark the sheet as modified")
}

// TestOrganizeTraitsUndoIsASingleStep verifies that organizing a sheet's traits records exactly one undoable edit, so
// that a single undo puts every trait back at the top level and a redo files them all away again.
func TestOrganizeTraitsUndoIsASingleStep(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = newOrganizeTraitsTestList(entity)
	sheet.Rebuild(true)
	original := traitNames(entity.Traits)
	c.False(sheet.undoMgr.CanUndo(), "nothing has been done yet")

	sheet.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)
	c.True(sheet.undoMgr.CanUndo(), "organizing the traits must be undoable")

	sheet.undoMgr.Undo()
	c.Equal(original, traitNames(entity.Traits), "one undo must put every trait back at the top level")
	c.Equal(len(original), sheet.Traits.Table.RootRowCount(),
		"one undo must leave the list on screen showing only the original rows, but it shows %v",
		rootRowNames(sheet.Traits.Table))
	for _, one := range entity.Traits {
		c.False(one.Container(), "%q must be back to being a plain trait", one.Name)
	}
	c.False(sheet.undoMgr.CanUndo(), "organizing the traits must have recorded exactly one edit")

	sheet.undoMgr.Redo()
	checkOrganizedTraits(c, entity.Traits)
	c.Equal(len(organizeTraitsTestCategories), sheet.Traits.Table.RootRowCount(),
		"redo must put the category containers back on screen, but it shows %v", rootRowNames(sheet.Traits.Table))
}

// TestOrganizeTraitsOnTemplate verifies that the command works on a template, whose traits have no entity behind them.
func TestOrganizeTraitsOnTemplate(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	data.Traits = newOrganizeTraitsTestList(nil)
	template := newTestTemplateDockable("Organize", data)
	original := traitNames(data.Traits)
	c.False(template.Modified(), "the template must start out unmodified")

	template.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)

	checkOrganizedTraits(c, template.template.Traits)
	c.Equal(len(organizeTraitsTestCategories), template.Traits.Table.RootRowCount(),
		"the list on screen must show one row per category container, but shows %v",
		rootRowNames(template.Traits.Table))
	c.True(template.Modified(), "organizing the traits must mark the template as modified")

	c.True(template.undoMgr.CanUndo(), "organizing the traits must be undoable")
	template.undoMgr.Undo()
	c.Equal(original, traitNames(template.template.Traits), "one undo must put every trait back at the top level")
	c.False(template.undoMgr.CanUndo(), "organizing the traits must have recorded exactly one edit")
}

// TestOrganizeTraitsOnAlreadyOrganizedTraitsRecordsNothing verifies that running the command a second time, when there
// is nothing left to file away, records no edit. A second edit would leave the user having to undo twice to get back
// to the traits they started with, with the first undo appearing to do nothing at all.
func TestOrganizeTraitsOnAlreadyOrganizedTraitsRecordsNothing(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = newOrganizeTraitsTestList(entity)
	sheet.Rebuild(true)
	original := traitNames(entity.Traits)

	sheet.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)
	organized := traitNames(entity.Traits)
	sheet.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)
	c.Equal(organized, traitNames(entity.Traits), "the second run must leave the already organized traits alone")

	c.True(sheet.undoMgr.CanUndo(), "the first run must still be undoable")
	sheet.undoMgr.Undo()
	c.Equal(original, traitNames(entity.Traits), "one undo must get back to the traits the sheet started with")
	c.False(sheet.undoMgr.CanUndo(), "the second run must not have recorded an edit of its own")
}

// TestOrganizeTraitsOnLibraryList verifies that the command works on a trait library list, whose traits have no data
// owner behind them at all, and that it is a single undoable step there too.
func TestOrganizeTraitsOnLibraryList(t *testing.T) {
	c := check.New(t)
	d := newTestTraitTableDockable()
	c.Nil(d.provider.DataOwner(), "a library list's traits have no data owner")
	original := traitNames(d.provider.RootData())
	c.False(d.Modified(), "the list must start out unmodified")
	c.Equal(len(organizeTraitsTestCategories), d.table.RootRowCount(),
		"the list must start out with each trait at the top level")
	c.False(d.undoMgr.CanUndo(), "nothing has been done yet")

	d.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)

	checkOrganizedTraits(c, d.provider.RootData())
	c.Equal(len(organizeTraitsTestCategories), d.table.RootRowCount(),
		"the list on screen must show one row per category container, but shows %v", rootRowNames(d.table))
	c.True(d.Modified(), "organizing the traits must mark the list as modified")

	c.True(d.undoMgr.CanUndo(), "organizing the traits must be undoable")
	d.undoMgr.Undo()
	c.Equal(original, traitNames(d.provider.RootData()), "one undo must put every trait back at the top level")
	c.Equal(len(original), d.table.RootRowCount(),
		"one undo must leave the list on screen showing only the original rows, but it shows %v", rootRowNames(d.table))
	c.False(d.undoMgr.CanUndo(), "organizing the traits must have recorded exactly one edit")

	d.undoMgr.Redo()
	checkOrganizedTraits(c, d.provider.RootData())
	c.Equal(len(organizeTraitsTestCategories), d.table.RootRowCount(),
		"redo must put the category containers back on screen, but it shows %v", rootRowNames(d.table))
}

// TestOrganizeTraitsIsDisabledWhileLibraryListIsFiltered verifies that a library list turns the command off while its
// content filter is hiding part of the list, since organizing moves rows around and a filtered table may not have its
// rows modified.
func TestOrganizeTraitsIsDisabledWhileLibraryListIsFiltered(t *testing.T) {
	c := check.New(t)
	d := newTestTraitTableDockable()
	c.True(d.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"an unfiltered library list must be able to organize its traits")

	d.table.ApplyFilter(func(_ *Node[*gurps.Trait]) bool { return false })
	c.False(d.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"a filtered library list must not offer the command")

	d.table.ApplyFilter(nil)
	c.True(d.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"clearing the filter must make the command available again")
}

// TestOrganizeTraitsIsAvailableOnEveryTraitsList verifies that the command is enabled for every dockable that has a
// traits list -- sheets, templates and trait library lists -- and nowhere else, since it is routed to whatever has the
// focus.
func TestOrganizeTraitsIsAvailableOnEveryTraitsList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	c.True(sheet.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID), "a sheet must be able to organize its traits")

	template := newTestTemplateDockable("Organize", gurps.NewTemplate())
	c.True(template.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"a template must be able to organize its traits")

	library := newTestTraitTableDockable()
	c.True(library.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"a trait library list must be able to organize its traits")

	loot := NewLootSheet("test"+gurps.LootExt, gurps.NewLoot())
	c.False(loot.AsPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"a loot sheet holds no traits, so it must not offer the command")

	c.False(unison.NewPanel().CanPerformCmd(nil, OrganizeTraitsItemID),
		"something with no traits list must not offer the command")
}
