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
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestMarkModifiedDropsCallsWhileTheUpdateIsUnderWay verifies that a call to MarkModified arriving from inside the
// sync it performs -- a panel reporting a change as it is brought into line with the model -- is dropped, and that the
// guard is lifted once the update is over, so that the next edit is acted upon.
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

// TestTemplateRebuildKeepsTheScrollPosition verifies that rebuilding a template puts the page back where the user had
// scrolled it. A rebuild that replaces a list -- here by hiding the TL column of a long equipment list -- hands the
// focus the old table held to its replacement, and a table scrolls itself into view as it takes the focus, which on
// its own would jump the page back up to the top of the list.
func TestTemplateRebuildKeepsTheScrollPosition(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	settings := gurps.GlobalSettings().SheetSettings()
	saved := settings.HideTLColumn
	t.Cleanup(func() { screen.Do(func() { settings.HideTLColumn = saved }) })
	screen.Do(func() { settings.HideTLColumn = false })
	template, ok := uxtest.OpenedByAction(t, screen, newCharacterTemplateAction).(*Template)
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

// pageDockableFixture is one of the three page dockables, along with what its constructor gave to the shared
// scaffolding and a change to its model that makes it modified.
type pageDockableFixture struct {
	name     string
	dockable pageDockable
	view     *pageView
	content  unison.Paneler
	modify   func()
}

// newPageDockableFixtures builds each of the three page dockables with the key bindable actions registered and a
// document dock in place, since both the toolbar and the rebuild path reach for them.
func newPageDockableFixtures(t *testing.T) []pageDockableFixture {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	entity := gurps.NewEntity()
	sheet := NewSheet("test"+gurps.SheetExt, entity)
	templateData := gurps.NewTemplate()
	template := NewTemplate("test"+gurps.TemplatesExt, templateData)
	lootData := gurps.NewLoot()
	loot := NewLootSheet("test"+gurps.LootExt, lootData)
	return []pageDockableFixture{
		{"sheet", sheet, &sheet.pageView, sheet.grid, func() { entity.Profile.Name = "Bob" }},
		{
			"template", template, &template.pageView, template.content,
			func() { templateData.Notes = append(templateData.Notes, gurps.NewNote(nil, nil, false)) },
		},
		{"loot", loot, &loot.pageView, loot.content, func() { lootData.Name = "Dragon Hoard" }},
	}
}

// TestPageDockablesShareTheScaffold verifies that each of the three page dockables comes out of its constructor with
// the scaffolding initPageDockable and finishPageDockable put up around its own content: the dockable is what the
// panel, the file-backed panel and the target manager refer to; it has an undo manager and the initial UI scale; the
// toolbar sits above the scroll panel holding the content; drops are rerouted; and Save As is always on offer, while
// Save waits for a change.
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

// sourceCellOfFirstRow returns the library source cell the list shows for its first row, ending the test if the list
// has no such column or row.
func sourceCellOfFirstRow[T gurps.Node[T]](t *testing.T, list *PageList[T], columnID int) gurps.CellData {
	t.Helper()
	col := list.Table.ColumnIndexForID(columnID)
	if col < 0 {
		t.Fatal("the list must show the library source column")
	}
	row := list.Table.RowFromIndex(0)
	if row == nil {
		t.Fatal("the list must show the sourced row")
	}
	return row.cellData(col, true)
}

// TestPageDockablesShowSourceStatesOnOpen verifies that a template or loot sheet shows how its sourced rows compare
// with their library sources as soon as it opens. A row keeps the cell data it derives until a change is reported, so
// lists built before the source matcher has loaded the library hashes it compares against show every sourced row as
// missing its source until something else reports a change.
func TestPageDockablesShowSourceStatesOnOpen(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().HideSourceMismatch, false)
	traitFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	libTrait := gurps.NewTrait(nil, nil, false)
	libTrait.Name = "Resistant"
	c.NoError(gurps.SaveTraits([]*gurps.Trait{libTrait},
		filepath.Join(user.Path(false), filepath.FromSlash(traitFile.Path))))
	eqpFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	libEqp := gurps.NewEquipment(nil, nil, false)
	libEqp.Name = "Rope"
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{libEqp},
		filepath.Join(user.Path(false), filepath.FromSlash(eqpFile.Path))))
	for _, one := range []struct {
		name string
		open func() gurps.CellData
	}{
		{"template", func() gurps.CellData {
			data := gurps.NewTemplate()
			data.Traits = []*gurps.Trait{libTrait.Clone(traitFile, data, nil, gurps.Reference)}
			return sourceCellOfFirstRow(t, NewTemplate("test"+gurps.TemplatesExt, data).Traits, gurps.TraitLibSrcColumn)
		}},
		{"loot sheet", func() gurps.CellData {
			data := gurps.NewLoot()
			data.Equipment = []*gurps.Equipment{libEqp.Clone(eqpFile, data, nil, gurps.Reference)}
			return sourceCellOfFirstRow(t, NewLootSheet("test"+gurps.LootExt, data).Equipment,
				gurps.EquipmentLibSrcColumn)
		}},
	} {
		state, _, _ := strings.Cut(one.open().Tooltip, "\n")
		c.Equal(srcstate.Matched.String(), state, "%s: the sourced row must show that it matches its source", one.name)
	}
}
