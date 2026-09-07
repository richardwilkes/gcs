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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestMarkModifiedDropsCallsWhileTheUpdateIsUnderWay verifies that a call to MarkModified arriving from inside the
// sync it performs -- a panel reporting a change as it is brought into line with the model -- is dropped rather than
// acted upon, so that the update doesn't repeat itself from inside itself, and that the guard is lifted once the update
// is over, so that the next edit is acted upon.
func TestMarkModifiedDropsCallsWhileTheUpdateIsUnderWay(t *testing.T) {
	c := check.New(t)
	template := newTestTemplateDockable("Reentrant", gurps.NewTemplate())
	counter := &syncCounter{}
	counter.Self = counter
	counter.onSync = func() { template.MarkModified(nil) }
	template.AsPanel().AddChild(counter)

	template.MarkModified(nil)
	c.Equal(1, counter.count, "the call from inside the sync must be dropped")
	c.False(template.awaitingUpdate, "the guard must be lifted once the update is over")

	template.MarkModified(nil)
	c.Equal(2, counter.count, "the next call must be acted upon")
}

// TestLootSheetMarkModifiedBumpsTheTimestamp verifies that marking a loot sheet as modified records when it was
// changed, which is the one thing its MarkModified does beyond what the other page dockables' do.
func TestLootSheetMarkModifiedBumpsTheTimestamp(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	sheet.loot.ModifiedOn = jio.Time{}
	sheet.MarkModified(nil)
	c.NotEqual(jio.Time{}, sheet.loot.ModifiedOn, "marking the loot sheet as modified must bump the timestamp")
}

// TestTemplateRebuildKeepsTheScrollPosition verifies that rebuilding a template puts the page back where the user had
// scrolled it. A rebuild that has to replace a list -- here by hiding the TL column of a long equipment list -- hands
// the focus the old table held to its replacement, and a table scrolls itself into view as it takes the focus: with
// the page scrolled well down into the list, that on its own would jump the page back up to the top of the list.
func TestTemplateRebuildKeepsTheScrollPosition(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	settings := gurps.GlobalSettings().SheetSettings()
	saved := settings.HideTLColumn
	t.Cleanup(func() { screen.Do(func() { settings.HideTLColumn = saved }) })
	screen.Do(func() { settings.HideTLColumn = false })
	template, ok := openedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}

	screen.Do(func() {
		items := make([]*gurps.Equipment, 0, 200)
		for i := range cap(items) {
			item := gurps.NewEquipment(nil, nil, false)
			item.Name = fmt.Sprintf("Item %d", i)
			items = append(items, item)
		}
		template.template.Equipment = items
		template.Rebuild(true)
	})

	// The page has been laid out again by now, so the scroll panel knows how far down it can go.
	const wanted = 300
	var stale *PageList[*gurps.Equipment]
	var focused bool
	var before float32
	screen.Do(func() {
		stale = template.Equipment
		stale.Table.RequestFocus()
		focused = wnd.Focus() == stale.Table.AsPanel()
		template.scroll.SetPosition(0, wanted)
		_, before = template.scroll.Position()
	})
	c.True(focused, "the equipment table takes the focus")
	c.Equal(float32(wanted), before, "the page must be long enough to scroll to where the test wants it")

	var replaced, refocused bool
	var after float32
	screen.Do(func() {
		settings.HideTLColumn = true
		template.Rebuild(true)
		replaced = template.Equipment != stale
		refocused = wnd.Focus() == template.Equipment.Table.AsPanel()
		_, after = template.scroll.Position()
	})
	c.True(replaced, "hiding the TL column must have replaced the equipment list")
	c.True(refocused, "the rebuild must hand the focus to the replacement")
	c.Equal(float32(wanted), after, "the rebuild must put the scroll position back")
}

// pageDockableFixture is one of the three page dockables, built headless, along with what its constructor gave to the
// shared scaffolding and a change to its model that makes it modified.
type pageDockableFixture struct {
	name     string
	dockable pageDockable
	view     *pageView
	content  unison.Paneler
	modify   func()
}

// newPageDockableFixtures builds each of the three page dockables the way their tests do: with the key bindable
// actions registered and a document dock in place, since both the toolbar and the rebuild path reach for them.
func newPageDockableFixtures(t *testing.T) []pageDockableFixture {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	entity := gurps.NewEntity()
	sheet := NewSheet("test"+gurps.SheetExt, entity)
	templateData := gurps.NewTemplate()
	template := NewTemplate("test"+gurps.TemplatesExt, templateData)
	lootData := gurps.NewLoot()
	loot := NewLootSheet("test"+gurps.LootExt, lootData)
	return []pageDockableFixture{
		{"sheet", sheet, &sheet.pageView, sheet.content, func() { entity.Profile.Name = "Bob" }},
		{
			"template", template, &template.pageView, template.content,
			func() { templateData.Notes = append(templateData.Notes, gurps.NewNote(nil, nil, false)) },
		},
		{"loot", loot, &loot.pageView, loot.content, func() { lootData.Name = "Dragon Hoard" }},
	}
}

// TestPageDockablesShareTheScaffold verifies that each of the three page dockables comes out of its constructor with
// the scaffolding initPageDockable and finishPageDockable put up around its own content: the dockable itself is what
// the panel, the file-backed panel and the target manager refer to; it has an undo manager that the undo edits made
// within it find, and starts at the initial UI scale; the toolbar sits above the scroll panel holding the content;
// drops onto it are rerouted; and Save As is always on offer, while Save waits for a change to the content.
func TestPageDockablesShareTheScaffold(t *testing.T) {
	c := check.New(t)
	for _, f := range newPageDockableFixtures(t) {
		p := f.dockable.AsPanel()
		c.True(f.view.Self == f.dockable, "%s: the panel's Self must be the dockable", f.name)
		c.True(f.view.dockable == f.dockable, "%s: the file-backed panel must refer to the dockable", f.name)
		c.True(f.view.targetMgr.root == p, "%s: the target manager must be rooted at the dockable", f.name)
		c.NotNil(f.view.undoMgr, "%s: the dockable must have an undo manager", f.name)
		c.True(unison.UndoManagerFor(f.view.scroll) == f.view.undoMgr,
			"%s: the undo manager must be the one found from within the page", f.name)
		c.Equal(gurps.GlobalSettings().General.InitialSheetUIScale, f.view.scale, "%s: the initial UI scale", f.name)
		c.True(f.view.scroll.Content().AsPanel() == f.content.AsPanel(),
			"%s: the scroll panel must hold the content", f.name)
		children := p.Children()
		c.Equal(2, len(children), "%s: the dockable must hold the toolbar and the scroll panel", f.name)
		c.True(children[0] == f.view.toolbar, "%s: the toolbar must come first", f.name)
		c.True(children[1] == f.view.scroll.AsPanel(), "%s: the scroll panel must come after the toolbar", f.name)
		c.True(p.CanAcceptDropCallback != nil && p.DropCallback != nil, "%s: drops must be rerouted", f.name)
		c.True(p.CanPerformCmd(nil, SaveAsItemID), "%s: Save As must always be on offer", f.name)
		c.False(p.CanPerformCmd(nil, SaveItemID), "%s: Save must wait for a change", f.name)
		f.modify()
		c.True(p.CanPerformCmd(nil, SaveItemID), "%s: Save must be on offer once there is a change", f.name)
	}
}

// TestPageDockablesOfferTheCommandsForTheirLists verifies that each page dockable installs the "New ..." commands for
// the lists it has and none for the lists it lacks -- a loot sheet never offers to add a trait, a template never to
// add other equipment -- and that the commands acting on the trait list as a whole go only to the dockables with one.
func TestPageDockablesOfferTheCommandsForTheirLists(t *testing.T) {
	c := check.New(t)
	fixtures := newPageDockableFixtures(t)
	for _, cmd := range []struct {
		name                  string
		id                    int
		sheet, template, loot bool
	}{
		{"New Trait", NewTraitItemID, true, true, false},
		{"New Trait Container", NewTraitContainerItemID, true, true, false},
		{"New Skill", NewSkillItemID, true, true, false},
		{"New Skill Container", NewSkillContainerItemID, true, true, false},
		{"New Technique", NewTechniqueItemID, true, true, false},
		{"New Spell", NewSpellItemID, true, true, false},
		{"New Spell Container", NewSpellContainerItemID, true, true, false},
		{"New Ritual Magic Spell", NewRitualMagicSpellItemID, true, true, false},
		{"New Carried Equipment", NewCarriedEquipmentItemID, true, true, false},
		{"New Carried Equipment Container", NewCarriedEquipmentContainerItemID, true, true, false},
		{"New Other Equipment", NewOtherEquipmentItemID, true, false, true},
		{"New Other Equipment Container", NewOtherEquipmentContainerItemID, true, false, true},
		{"New Note", NewNoteItemID, true, true, true},
		{"New Note Container", NewNoteContainerItemID, true, true, true},
		{"Add Natural Attacks", AddNaturalAttacksItemID, true, true, false},
		{"Organize Traits", OrganizeTraitsItemID, true, true, false},
	} {
		for i, wanted := range []bool{cmd.sheet, cmd.template, cmd.loot} {
			f := fixtures[i]
			c.Equal(wanted, f.dockable.AsPanel().CanPerformCmd(nil, cmd.id), "%s: %s", f.name, cmd.name)
		}
	}
}

// TestAddNaturalAttacksGoesToTheTraitListOfTheDockable verifies that the Add Natural Attacks command adds the trait to
// the trait list of the dockable it was invoked on, and that on a sheet the trait is made for the sheet's entity,
// while on a template, whose traits have no entity until it is applied, it is made for none. A new entity may already
// hold a natural attacks trait of its own (see gurps.GeneralSettings.AutoAddNaturalAttacks), so only the growth of the
// lists is looked at.
func TestAddNaturalAttacksGoesToTheTraitListOfTheDockable(t *testing.T) {
	c := check.New(t)
	fixtures := newPageDockableFixtures(t)
	sheet, ok := fixtures[0].dockable.(*Sheet)
	c.True(ok, "the first fixture must be the sheet")
	template, ok := fixtures[1].dockable.(*Template)
	c.True(ok, "the second fixture must be the template")
	sheetTraits := len(sheet.entity.Traits)
	templateTraits := len(template.template.Traits)

	sheet.AsPanel().PerformCmd(nil, AddNaturalAttacksItemID)
	c.Equal(sheetTraits+1, len(sheet.entity.Traits), "the sheet must gain the trait")
	added := sheet.entity.Traits[len(sheet.entity.Traits)-1]
	c.Equal("Natural Attacks", added.Name)
	c.True(added.DataOwner() == sheet.entity, "the sheet's trait must be made for its entity")
	c.Equal(templateTraits, len(template.template.Traits), "the template must be left alone")

	template.AsPanel().PerformCmd(nil, AddNaturalAttacksItemID)
	c.Equal(templateTraits+1, len(template.template.Traits), "the template must gain the trait")
	added = template.template.Traits[len(template.template.Traits)-1]
	c.Equal("Natural Attacks", added.Name)
	c.Equal(sheetTraits+1, len(sheet.entity.Traits), "the sheet must be left alone")
}
