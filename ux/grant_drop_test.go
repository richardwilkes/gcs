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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// grantAt returns the panel's feature at the index, which must be a grant.
func grantAt(t *testing.T, p *featuresPanel, index int) *gurps.Grant {
	t.Helper()
	g, ok := (*p.features)[index].(*gurps.Grant)
	if !ok {
		t.Fatalf("feature %d is a %T, not a grant", index, (*p.features)[index])
	}
	return g
}

// newGrantTestTrait returns a trait outside any sheet with the given name and levels, leveled when levels isn't zero.
func newGrantTestTrait(name string, levels fxp.Int) *gurps.Trait {
	t := gurps.NewTrait(nil, nil, false)
	t.Name = name
	if levels != 0 {
		t.CanLevel = true
		t.PointsPerLevel = fxp.One
		t.Levels = levels
	}
	return t
}

// TestGrantItemsFromDraggedRows checks that dragged rows become one grant item each, that a container stands for the
// items within it, and that an item that grants something itself is left out.
func TestGrantItemsFromDraggedRows(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	vision := gurps.NewTrait(nil, nil, true)
	vision.Name = "Vision"
	for _, name := range []string{"Night Vision", "Dark Vision"} {
		child := newGrantTestTrait(name, 0)
		child.SetParent(vision)
		vision.Children = append(vision.Children, child)
	}
	granting := newGrantTestTrait("Granting", 0)
	granting.Features = gurps.Features{gurps.NewGrant(feature.GrantTrait)}
	rows := []*Node[*gurps.Trait]{
		NewNode(table, nil, vision, false),
		NewNode(table, nil, newGrantTestTrait("Fit", 0), false),
		NewNode(table, nil, granting, false),
	}
	kind, items := grantItemsFrom(&unison.TableDragData[*Node[*gurps.Trait]]{Table: table, Rows: rows})
	c.Equal(feature.GrantTrait, kind)
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.name)
	}
	c.Equal([]string{"Night Vision", "Dark Vision", "Fit"}, names)

	kind, _ = grantItemsFrom(nil)
	c.Equal(feature.Unknown, kind, "anything else isn't a drag of items to grant")
}

// TestFeaturesPanelDropsGrants checks that items dropped on an empty grant fill it, with any more added after it, and
// that items dropped elsewhere become new grants, a leveled trait at its level, opened at their level field, each drop
// a single undoable edit.
func TestFeaturesPanelDropsGrants(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	empty := gurps.NewGrant(feature.GrantTrait)
	features := gurps.Features{empty}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	items := func(names ...string) []*grantLibraryItem {
		list := make([]*grantLibraryItem, 0, len(names))
		for _, name := range names {
			list = append(list, &grantLibraryItem{node: newGrantTestTrait(name, fxp.Three)})
		}
		return list
	}
	screen.Do(func() {
		var row *unison.Panel
		for _, child := range p.Children() {
			if path, ok := child.ClientData()[featureDropKey].(string); ok && path == "0" {
				row = child
			}
		}
		c.NotNil(row)
		p.dropGrants(row, dropInto, feature.GrantTrait, items("Night Vision", "Dark Vision"))
	})
	screen.Do(func() {
		c.Equal(2, len(features))
		c.True(features[0] == empty, "the empty grant is filled, not replaced")
		c.Equal("Night Vision", empty.ItemName())
		c.Equal(fxp.Three, empty.Level, "the filled grant takes the trait's level too")
		second, ok := features[1].(*gurps.Grant)
		c.True(ok)
		c.Equal("Dark Vision", second.ItemName(), "a second item is added after it")
		c.Equal("0", p.open)
		p.dropGrants(nil, 0, feature.GrantTrait, items("Fit"))
	})
	screen.Do(func() {
		c.Equal(3, len(features), "dropped anywhere else, it goes at the end")
		last, ok := features[2].(*gurps.Grant)
		c.True(ok)
		c.Equal("Fit", last.ItemName())
		c.Equal(fxp.Three, last.Level, "a leveled trait is granted at its level")
		c.Equal("2", p.open)
		c.Equal("2:level", p.Window().Focus().RefKey, "the new grant's level field takes the focus")
		host.mgr.Undo()
	})
	screen.Do(func() {
		c.Equal(2, len(features), "undo takes the drop back")
		host.mgr.Undo()
	})
	screen.Do(func() {
		c.Equal(1, len(features))
		g, ok := features[0].(*gurps.Grant)
		c.True(ok)
		c.Equal("", string(g.Key()), "and empties the grant again")
	})
}

// TestGrantRowControls checks the controls of an open grant row: the level of a skill or spell grant is a whole number,
// since the grant acts at the tier, and its tooltip prices the tiers as grantTierPoints does; and Cmd-Return in the
// picker's filter is left to the editor, which applies the changes with it.
func TestGrantRowControls(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	features := gurps.Features{g, gurps.NewGrant(feature.GrantSkill)}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		field, ok := p.FindRefKey("0:level").Self.(*DecimalField)
		if !ok {
			t.Fatal("the open grant row has no level field")
		}
		field.SetText("2.5")
		c.Equal(fxp.Two, g.Level, "a skill grant's level is whole")
		p.toggle("1")
	})
	screen.Do(func() {
		filter, ok := p.FindRefKey("1:filter").Self.(*unison.Field)
		if !ok {
			t.Fatal("the empty grant row has no filter")
		}
		c.True(filter.KeyDownCallback(unison.KeyReturn, 0, false), "Return chooses from the picker")
		c.False(filter.KeyDownCallback(unison.KeyReturn, mod.OSMenuCommand(), false), "Cmd-Return is the editor's")
	})
	c.Equal("The level is counted in steps of points: level 1 is what 1 point buys, level 2 what 2 points buy, level 3 "+
		"what 4 points buy, and each level after that 4 points more. A negative level only takes levels or points away "+
		"from one already present.", grantLevelTooltip(feature.GrantSkill))
}

// TestGrantPickerFollowsItsGrant checks that the picker opened to choose a different item for a grant stays with that
// grant as rows move, rather than with the row's position, so another grant moved into that position doesn't get it.
func TestGrantPickerFollowsItsGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	first := gurps.NewGrant(feature.GrantTrait)
	first.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
	second := gurps.NewGrant(feature.GrantTrait)
	second.SetItem(newGrantTestTrait("Dark Vision", 0), gurps.LibraryFile{})
	features := gurps.Features{first, second}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		c.Nil(p.FindRefKey("0:filter"), "a chosen grant shows no picker until asked")
		button, ok := p.FindRefKey("0:item").Self.(*unison.Button)
		if !ok {
			t.Fatal("the open grant row has no item button")
		}
		button.Click()
	})
	screen.Do(func() {
		c.NotNil(p.FindRefKey("0:filter"), "the item button opens the picker")
		for _, entry := range p.moreEntries("0") {
			if entry.Label == "Move Down" {
				entry.Act()
			}
		}
	})
	screen.Do(func() {
		c.True(features[1] == first, "precondition: the grant moved down")
		c.Equal("1", p.open, "and stays open")
		c.NotNil(p.FindRefKey("1:filter"), "with its picker")
		p.toggle("0")
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("0:filter"), "the grant now in its old position has no picker")
	})
}

// TestGrantLibraryItemsFollowTheLibrary checks that the picker's items name the library by its current key once the
// library has been rekeyed, that a file removed from the library drops out, and that an unchanged library gives the
// same list again rather than a fresh one.
func TestGrantLibraryItemsFollowTheLibrary(t *testing.T) {
	c := check.New(t)
	_, _ = uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	path := filepath.Join(dir, "Traits", "Test"+gurps.TraitsExt)
	c.NoError(gurps.SaveTraits([]*gurps.Trait{newGrantTestTrait("Night Vision", 0)}, path))
	find := func(items []*grantLibraryItem) *grantLibraryItem {
		for _, item := range items {
			if item.name == "Night Vision" {
				return item
			}
		}
		return nil
	}
	items := grantLibraryItems(feature.GrantTrait)
	item := find(items)
	if item == nil {
		t.Fatal("the library's trait is not among the items")
	}
	c.Equal("acct/repo", item.from.Library)
	again := grantLibraryItems(feature.GrantTrait)
	c.True(len(again) != 0 && &again[0] == &items[0], "an unchanged library gives the same list again")

	oldKey := lib.Key()
	lib.Configure(library.Config{Title: "Test", GitHubAccountName: "other", RepoName: "repo"})
	libraries.Rekey(oldKey, lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	item = find(grantLibraryItems(feature.GrantTrait))
	if item == nil {
		t.Fatal("the rekeyed library's trait is not among the items")
	}
	c.Equal("other/repo", item.from.Library, "the items name the library by its current key")

	c.NoError(os.Remove(path))
	c.Nil(find(grantLibraryItems(feature.GrantTrait)), "a file removed from the library drops out")
}

// TestGrantItemMatches checks that each word of the filter matches the start of a word of the name, or one a typo
// away once the word is long enough.
func TestGrantItemMatches(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		query string
		name  string
		match bool
	}{
		{query: "night", name: "Night Vision", match: true},
		{query: "vis nig", name: "Night Vision", match: true},
		{query: "nihgt", name: "Night Vision", match: true},
		{query: "nite", name: "Night Vision", match: false},
		{query: "ab", name: "Acute Vision", match: false},
		{query: "", name: "Acute Vision", match: true},
	} {
		c.Equal(tc.match, grantItemMatches(grantFilterWords(tc.query), grantFilterWords(tc.name)), tc.query)
	}
}

// TestGrantedItemEditLandsOnTheGrant checks that an edit of a grant's item, made in its editor on a copy, goes to the
// grant the panel holds once its own undo has replaced the grant with a copy.
func TestGrantedItemEditLandsOnTheGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Night Vision", fxp.One), gurps.LibraryFile{})
	features := gurps.Features{g}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() {
		item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o := &grantItemOwner{panel: p, key: g.Key(), item: item, commit: func(g *gurps.Grant) {
			g.Trait = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		}}
		p.edit("Level", "0:level", "", func() { g.Level = fxp.Two })
		host.mgr.Undo()
		item.Name = "Dark Vision"
		o.Rebuild(true)
		current, ok := features[0].(*gurps.Grant)
		c.True(ok)
		c.False(current == g, "undo put a copy of the grant in its place")
		c.Equal("Dark Vision", current.ItemName())
		c.Equal("Night Vision", g.ItemName(), "the grant undo replaced is left alone")
	})
}

// newTestSheetWithGrant returns a sheet holding the skill Stealth and, inside a closed container, the trait Training,
// which grants the skill Brawling.
func newTestSheetWithGrant(t *testing.T) (sheet *Sheet, training *gurps.Trait, granted, stealth *gurps.Skill) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	group := gurps.NewTrait(entity, nil, true)
	group.SetOpen(false)
	training = gurps.NewTrait(entity, group, false)
	training.Name = "Training"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	group.Children = []*gurps.Trait{training}
	entity.Traits = []*gurps.Trait{group}
	stealth = newTestSkill("Stealth", fxp.One, nil)
	stealth.SetDataOwner(entity)
	entity.Skills = []*gurps.Skill{stealth}
	sheet.Rebuild(true)
	for _, one := range entity.Skills {
		if gurps.IsGranted(one) {
			granted = one
		}
	}
	return sheet, training, granted, stealth
}

// TestGrantedRowsOnSheet checks that the commands that change rows are off for a selection of granted rows alone, and
// that the mark of a granted row selects the item granting it, opening its container.
func TestGrantedRowsOnSheet(t *testing.T) {
	c := check.New(t)
	sheet, training, granted, stealth := newTestSheetWithGrant(t)
	c.NotNil(granted)
	table := sheet.Skills.Table
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true})
	c.False(hasUngrantedSelection(table), "a granted row changes only with its grant")
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true, stealth.ID(): true})
	c.True(hasUngrantedSelection(table))

	selectGrantingItem(table, gurps.GrantingItem(granted))
	c.True(training.Parent().IsOpen(), "the container holding the grantor is opened")
	selected := sheet.Traits.Table.SelectedRows(false)
	c.Equal(1, len(selected))
	c.True(selected[0].Data() == training, "the grantor is selected")
}

// TestGrantedItemEditorIsReadOnly checks that a granted item's editor only shows it.
func TestGrantedItemEditorIsReadOnly(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		sheet, _, granted, stealth := newTestSheetWithGrant(t)
		DisplayNewDockable(sheet)
		c.True(EditSkill(sheet, granted).readOnly)
		c.False(EditSkill(sheet, stealth).readOnly)
	})
}

// TestGrantItemEditorOffersNoGrants checks that the editor of a grant's item, and an editor opened from it for one of
// the item's modifiers, offer no grant feature types and take no dropped grants, since a granted item grants nothing.
func TestGrantItemEditorOffersNoGrants(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	nightVision := newGrantTestTrait("Night Vision", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Mod"
	nightVision.Modifiers = []*gurps.TraitModifier{modifier}
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(nightVision, gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{training}
	sheet.Rebuild(true)
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var outer *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer = EditTrait(sheet, training)
		panels := uxtest.PanelsOfType[*featuresPanel](outer.content)
		c.Equal(1, len(panels))
		panels[0].editGrantedItem(grantAt(t, panels[0], 0))
	})
	nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
	var modEditor *editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]
	screen.Do(func() {
		_, isNested := nested.owner.(*grantItemOwner)
		c.True(isNested, "precondition: the other trait editor is the grant item's")
		EditTraitModifier(nested, nested.editorData.Modifiers[0])
	})
	modEditor = uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	screen.Do(func() {
		for _, e := range []*unison.Panel{nested.content, modEditor.content} {
			panels := uxtest.PanelsOfType[*featuresPanel](e)
			c.Equal(1, len(panels))
			c.False(slices.ContainsFunc(panels[0].featureTypesList(), gurps.IsGrant), "no grant types are offered")
			c.False(panels[0].takesGrants(), "and no grants are taken")
		}
	})
}

// TestGrantItemEditorClosesWhenItsGrantChanges checks that the editor open on a grant's item closes, without asking,
// once the outer editor's panel drops the grant or gives it another item, since what the editor would save has nowhere
// to go, or would overwrite the new item.
func TestGrantItemEditorClosesWhenItsGrantChanges(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(p *featuresPanel, g *gurps.Grant)
	}{
		{name: "deleted", change: func(p *featuresPanel, _ *gurps.Grant) {
			for _, entry := range p.moreEntries("0") {
				if entry.Label == "Delete" {
					entry.Act()
				}
			}
		}},
		{name: "another item", change: func(p *featuresPanel, g *gurps.Grant) {
			// As choosing from the picker does.
			p.sentenceRows.restructure("Choose Granted Item", "0:filter", "", func() string {
				g.SetItem(newGrantTestTrait("Dark Vision", 0), gurps.LibraryFile{})
				return grantFocusKey("0", g)
			})
		}},
		{name: "item changed", change: func(p *featuresPanel, g *gurps.Grant) {
			// As Sync Granted Item with Source does.
			p.edit("Sync Granted Item with Source", "0:more", "0:more", func() { g.Trait.Name = "Dark Vision" })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			sheet := newTestSheetForTemplate(t)
			entity := sheet.Entity()
			training := gurps.NewTrait(entity, nil, false)
			training.Name = "Training"
			g := gurps.NewGrant(feature.GrantTrait)
			g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
			training.Features = gurps.Features{g}
			entity.Traits = []*gurps.Trait{training}
			sheet.Rebuild(true)
			isTraitEditor := func(d unison.Dockable) bool {
				_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
				return ok
			}
			var outer *editor[*gurps.Trait, *gurps.TraitEditData]
			var p *featuresPanel
			screen.Do(func() {
				DisplayNewDockable(sheet)
				outer = EditTrait(sheet, training)
				p = uxtest.PanelsOfType[*featuresPanel](outer.content)[0]
				p.editGrantedItem(grantAt(t, p, 0))
			})
			nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
			screen.Do(func() {
				nested.editorData.Name = "Changed"
				nested.MarkModified(nil)
				tc.change(p, grantAt(t, p, 0))
			})
			var open int
			screen.Do(func() { open = len(AllMatchingDockables(isTraitEditor)) })
			c.Equal(1, open, "the grant item's editor closes without asking, leaving the outer editor")
		})
	}
}

// TestGrantDestinationsAreToldApart checks that the editors offered as destinations for Add as Grant are named so that
// two modifier editors with the same title, opened from the editors of different traits, can be told apart.
func TestGrantDestinationsAreToldApart(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traits := make([]*gurps.Trait, 0, 2)
	for _, name := range []string{"Training", "Schooling"} {
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = name
		modifier := gurps.NewTraitModifier(entity, nil, false)
		modifier.Name = "Mod"
		trait.Modifiers = []*gurps.TraitModifier{modifier}
		gurps.AttachModifiers(trait, trait.Modifiers)
		traits = append(traits, trait)
	}
	entity.Traits = traits
	sheet.Rebuild(true)
	var names []string
	screen.Do(func() {
		DisplayNewDockable(sheet)
		for _, trait := range traits {
			e := EditTrait(sheet, trait)
			EditTraitModifier(e, e.editorData.Modifiers[0])
		}
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{newGrantTestTrait("Night Vision", 0)})
		for _, d := range grantDestinations(items) {
			names = append(names, d.String())
		}
	})
	c.Equal(4, len(names))
	c.Equal(len(names), len(slices.Compact(slices.Clone(names))), "every destination has a name of its own: %v", names)
	c.True(slices.ContainsFunc(names, func(s string) bool {
		return strings.HasPrefix(s, "Mod (") && strings.Contains(s, "Schooling")
	}), "a modifier's editor names the trait it was opened from: %v", names)
}

// TestGrantNotAddedToItsOwnEditor checks that an item is never added as a grant to the editor of that same item, since
// the item would then grant a copy of itself.
func TestGrantNotAddedToItsOwnEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	fit := gurps.NewTrait(entity, nil, false)
	fit.Name = "Fit"
	entity.Traits = []*gurps.Trait{fit}
	sheet.Rebuild(true)
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		DisplayNewDockable(sheet)
		e = EditTrait(sheet, fit)
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{fit, newGrantTestTrait("Night Vision", 0)})
		c.Equal(2, len(items))
		addAsGrant(feature.GrantTrait, items)
	})
	var names []string
	screen.Do(func() {
		for _, f := range e.editorData.Features {
			if g, ok := f.(*gurps.Grant); ok {
				names = append(names, g.ItemName())
			}
		}
	})
	c.Equal([]string{"Night Vision"}, names, "the item being edited is left out")
}

// TestDeletingContainerKeepsGrantedRows checks that deleting a container the user moved a granted row into leaves that
// row on the sheet, since a granted row goes only with its grant.
func TestDeletingContainerKeepsGrantedRows(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var sheet *Sheet
	var granted *gurps.Skill
	var group *gurps.Skill
	screen.Do(func() {
		sheet, _, granted, _ = newTestSheetWithGrant(t)
		entity := sheet.Entity()
		group = gurps.NewSkill(entity, nil, true)
		group.Name = "Group"
		entity.Skills = slices.DeleteFunc(entity.Skills, func(s *gurps.Skill) bool { return s == granted })
		granted.SetParent(group)
		group.Children = []*gurps.Skill{granted}
		entity.Skills = append(entity.Skills, group)
		sheet.Rebuild(true)
		DisplayNewDockable(sheet)
		sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		DeleteSelection(sheet.Skills.Table, true)
	})
	var kept bool
	var groups int
	screen.Do(func() {
		for _, s := range sheet.Entity().Skills {
			if s == granted {
				kept = true
			}
			if s.Container() {
				groups++
			}
		}
	})
	c.Equal(0, groups, "the container goes")
	c.True(kept, "the granted row within it stays, as the row it was")
}

// TestNestedEditorTitle checks that an editor opened from a document is named for the document, while one opened from
// another editor, for a grant's item or for one of the item's modifiers, is named for its item.
func TestNestedEditorTitle(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	outer := newGrantTestTrait("Training", 0)
	inner := newGrantTestTrait("Night Vision", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Mod"
	inner.Modifiers = []*gurps.TraitModifier{modifier}
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(inner, gurps.LibraryFile{})
	outer.Features = gurps.Features{g}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{outer}
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var outerTitle string
	screen.Do(func() {
		template := newTestTemplateDockable("Template", data)
		DisplayNewDockable(template)
		e = EditTrait(template, outer)
		outerTitle = e.Title()
		p := uxtest.PanelsOfType[*featuresPanel](e.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
	})
	nested := uxtest.OtherEditor(t, screen, e, isTraitEditor)
	var innerTitle, modTitle string
	screen.Do(func() {
		innerTitle = nested.Title()
		EditTraitModifier(nested, nested.editorData.Modifiers[0])
	})
	modEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	screen.Do(func() { modTitle = modEditor.Title() })
	c.Equal("Trait Editor for Template", outerTitle)
	c.Equal("Trait Editor for Night Vision", innerTitle)
	c.Equal("Trait Modifier Editor for Mod", modTitle)
}
