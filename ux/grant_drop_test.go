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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xmath"
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

// TestGrantDropZoneReplacesTheChosenItem checks that the open row of a grant keeps its drop zone once an item is chosen,
// that a drop on the zone replaces the item rather than adding a grant, and that a drop elsewhere on the row still goes
// beside it.
func TestGrantDropZoneReplacesTheChosenItem(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
	features := gurps.Features{g}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	var zone *unison.Panel
	screen.Do(func() { zone = p.FindRefKey("0:drop") })
	if zone == nil {
		t.Fatal("the open row of a chosen grant has no drop zone")
	}
	screen.Do(func() {
		where := p.PointFromRoot(zone.PointToRoot(zone.ContentRect(true).Center()))
		target, at := p.grantDropAt(where, feature.GrantTrait)
		c.True(target == zone, "a drop on the zone goes into the grant")
		c.Equal(dropInto, at)
		row, _, _ := p.rowAt(where)
		target, at = p.grantDropAt(p.PointFromRoot(row.PointToRoot(geom.NewPoint(1, 1))), feature.GrantTrait)
		c.True(target == row, "a drop elsewhere on the row goes beside it")
		c.Equal(dropBefore, at)
		p.dropGrants(zone, dropInto, feature.GrantTrait, []*grantLibraryItem{{node: newGrantTestTrait("Dark Vision", 0)}})
	})
	screen.Do(func() {
		c.Equal(1, len(features), "the drop replaces the item rather than adding a grant")
		c.True(features[0] == g, "in the same grant")
		c.Equal("Dark Vision", g.ItemName())
	})
}

// TestGrantRowControls checks the controls of an open grant row: the level of a skill or spell grant is a whole number,
// since the grant acts at the tier, so its field refuses fractions, where a trait grant's takes them, and its tooltip
// prices the tiers as grantTierPoints does; and Cmd-Return in the picker's filter is left to the editor, which applies
// the changes with it.
func TestGrantRowControls(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	// A loaded grant may hold a merge that means nothing for its item, which the grant treats as Ignore.
	traitGrant := gurps.NewGrant(feature.GrantTrait)
	traitGrant.SetItem(newGrantTestTrait("Night Vision", fxp.One), gurps.LibraryFile{})
	traitGrant.Merge = grantmerge.MergePoints
	features := gurps.Features{g, gurps.NewGrant(feature.GrantSkill), traitGrant}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("2") })
	screen.Do(func() {
		popup, ok := p.FindRefKey("2:merge").Self.(*unison.PopupMenu[grantmerge.Mode])
		if !ok {
			t.Fatal("the open trait grant row has no merge popup")
		}
		selected, has := popup.Selected()
		c.True(has, "the popup shows the merge the grant uses")
		c.Equal(grantmerge.Ignore, selected)
		field, ok := p.FindRefKey("2:level").Self.(*DecimalField)
		if !ok {
			t.Fatal("the open trait grant row has no level field")
		}
		field.SetText("1.5")
		c.Equal(fxp.OneAndAHalf, traitGrant.Level, "a trait grant's level may be fractional")
		p.toggle("0")
	})
	screen.Do(func() {
		field, ok := p.FindRefKey("0:level").Self.(*DecimalField)
		if !ok {
			t.Fatal("the open grant row has no level field")
		}
		field.SetText("2.5")
		c.NotEqual("", field.tooltipTextForValidation(), "a skill grant's level is whole, so a pasted fraction is invalid")
		field.SetText("2")
		c.Equal(fxp.Two, g.Level)
		field.SetSelectionToEnd()
		c.False(field.RuneTypedCallback('.'), "and a typed decimal point is refused")
		c.Equal("2", field.Text())
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

// TestGrantLibraryItemsFollowSymlinkedFolders checks that the picker's items include those in a folder a symlink in
// the library points at, as the Library Explorer lists them, and that a folder a symlink leads to again, one of the
// library's own or one back up the tree, is not walked again, which would offer its items twice or walk forever.
func TestGrantLibraryItemsFollowSymlinkedFolders(t *testing.T) {
	c := check.New(t)
	_, _ = uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	traits := filepath.Join(dir, "Traits")
	c.NoError(os.MkdirAll(traits, 0o750))
	if err := os.Symlink(traits, filepath.Join(dir, "Alias")); err != nil {
		t.Skipf("unable to create a symlink here: %v", err) // Windows requires a privilege for this by default
	}
	linked := filepath.Join(t.TempDir(), "Linked Traits")
	c.NoError(os.MkdirAll(linked, 0o750))
	c.NoError(os.Symlink(linked, filepath.Join(dir, "Linked")))
	c.NoError(os.Symlink(dir, filepath.Join(linked, "Loop")))
	c.NoError(gurps.SaveTraits([]*gurps.Trait{newGrantTestTrait("Night Vision", 0)},
		filepath.Join(traits, "Test"+gurps.TraitsExt)))
	c.NoError(gurps.SaveTraits([]*gurps.Trait{newGrantTestTrait("Dark Vision", 0)},
		filepath.Join(linked, "Test"+gurps.TraitsExt)))
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	paths := make(map[string][]string)
	for _, item := range grantLibraryItems(feature.GrantTrait) {
		if item.from.Library == lib.Key() {
			paths[item.name] = append(paths[item.name], item.from.Path)
		}
	}
	c.Equal([]string{"Linked/Test" + gurps.TraitsExt}, paths["Dark Vision"],
		"the trait in the folder a symlink points at is offered once, by its path within the library")
	c.Equal([]string{"Traits/Test" + gurps.TraitsExt}, paths["Night Vision"],
		"the trait in a folder of the library is offered once, by its path there, though a symlink leads to it too")
}

// TestGrantPickerLeavesOutTheItemEdited checks that the picker doesn't offer the item being edited, or a copy of it,
// as the drop and Add as Grant leave it out, since the item would otherwise grant a copy of itself.
func TestGrantPickerLeavesOutTheItemEdited(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	libFile := gurps.LibraryFile{Library: lib.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	nightVision := newGrantTestTrait("Night Vision", 0)
	darkVision := newGrantTestTrait("Dark Vision", 0)
	c.NoError(gurps.SaveTraits([]*gurps.Trait{nightVision, darkVision}, filepath.Join(dir, filepath.FromSlash(libFile.Path))))
	entity := gurps.NewEntity()
	owner := nightVision.Clone(libFile, entity, nil, gurps.Reference)
	features := gurps.Features{gurps.NewGrant(feature.GrantTrait)}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	var names []string
	screen.Do(func() {
		lists := uxtest.PanelsOfType[*unison.List[*grantLibraryItem]](p.AsPanel())
		if len(lists) != 1 {
			t.Fatalf("expected one picker list, found %d", len(lists))
		}
		for i := range lists[0].Count() {
			if item := lists[0].DataAtIndex(i); item.from == libFile {
				names = append(names, item.name)
			}
		}
	})
	c.Equal([]string{"Dark Vision"}, names, "the item being edited is left out")
}

// TestGrantPickerListFollowsTheFilter checks that the picker's list is resized to the rows the filter leaves, no taller
// than its view when they fit, since the scroll panel otherwise keeps the frame of the full list, leaving blank space to
// scroll through below a filtered list and clipping the rows of one that grew.
func TestGrantPickerListFollowsTheFilter(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	traits := make([]*gurps.Trait, 0, grantPickerRows+3)
	traits = append(traits, newGrantTestTrait("Night Vision", 0))
	for i := range grantPickerRows + 2 {
		traits = append(traits, newGrantTestTrait(fmt.Sprintf("Acute Sense %d", i), 0))
	}
	c.NoError(gurps.SaveTraits(traits, filepath.Join(dir, "Traits", "Test"+gurps.TraitsExt)))
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := gurps.Features{gurps.NewGrant(feature.GrantTrait)}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	var list *unison.List[*grantLibraryItem]
	var filter *unison.Field
	screen.Do(func() {
		lists := uxtest.PanelsOfType[*unison.List[*grantLibraryItem]](p.AsPanel())
		if len(lists) != 1 {
			t.Fatalf("expected one picker list, found %d", len(lists))
		}
		list = lists[0]
		var ok bool
		if filter, ok = p.FindRefKey("0:filter").Self.(*unison.Field); !ok {
			t.Fatal("the picker has no filter field")
		}
	})
	rowHeight := xmath.Ceil(list.Factory.CellHeight())
	screen.Do(func() { filter.SetText("Night Vision") })
	screen.Do(func() {
		c.Equal(1, list.Count(), "precondition: the filter leaves one row")
		c.Equal(list.Parent().ContentRect(true).Height, list.FrameRect().Height,
			"the list shrinks to its view once the filter leaves fewer rows than fit")
	})
	screen.Do(func() { filter.SetText("") })
	screen.Do(func() {
		c.True(list.Count() > grantPickerRows, "precondition: the cleared filter shows every row again")
		c.Equal(rowHeight*float32(list.Count()), list.FrameRect().Height, "the list grows back with the filter cleared")
	})
}

// TestGrantPickerKeepsItsFilterAcrossRebuilds checks that the filter typed into a picker, and the row chosen in its list,
// survive a rebuild of the panel, which an edit of the open row's other controls causes while the picker is open.
func TestGrantPickerKeepsItsFilterAcrossRebuilds(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	traits := make([]*gurps.Trait, 0, 4)
	traits = append(traits, newGrantTestTrait("Night Vision", 0))
	for i := range 3 {
		traits = append(traits, newGrantTestTrait(fmt.Sprintf("Acute Sense %d", i), 0))
	}
	c.NoError(gurps.SaveTraits(traits, filepath.Join(dir, "Traits", "Test"+gurps.TraitsExt)))
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	features := gurps.Features{gurps.NewGrant(feature.GrantTrait)}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	picker := func() (*unison.Field, *unison.List[*grantLibraryItem]) {
		filter, ok := p.FindRefKey("0:filter").Self.(*unison.Field)
		if !ok {
			t.Fatal("the picker has no filter field")
		}
		lists := uxtest.PanelsOfType[*unison.List[*grantLibraryItem]](p.AsPanel())
		if len(lists) != 1 {
			t.Fatalf("expected one picker list, found %d", len(lists))
		}
		return filter, lists[0]
	}
	screen.Do(func() {
		filter, list := picker()
		filter.SetText("Acute")
		c.Equal(3, list.Count(), "precondition: the filter leaves the acute senses")
		list.Select(false, 2)
		p.rebuild("")
	})
	screen.Do(func() {
		filter, list := picker()
		c.Equal("Acute", filter.Text(), "the filter typed survives the rebuild")
		c.Equal(3, list.Count(), "and so does the list it leaves")
		c.Equal(2, list.Selection.FirstSet(), "with the same row chosen")
		p.toggle("0")
	})
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		filter, _ := picker()
		c.Equal("", filter.Text(), "a picker opened afresh starts with no filter")
	})
}

// TestGrantPickerSelectsOneRow checks that the picker's list holds one selected row, the one Return chooses, so that a
// Shift-click or Cmd-click on another row can't put a second one above it, and that Up and Down with a modifier down
// are the filter's rather than the list's, as Shift-Down selects to the end of the filter's text.
func TestGrantPickerSelectsOneRow(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	traits := make([]*gurps.Trait, 0, 3)
	for i := range 3 {
		traits = append(traits, newGrantTestTrait(fmt.Sprintf("Acute Sense %d", i), 0))
	}
	c.NoError(gurps.SaveTraits(traits, filepath.Join(dir, "Traits", "Test"+gurps.TraitsExt)))
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantTrait)
	features := gurps.Features{g}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		filter, ok := p.FindRefKey("0:filter").Self.(*unison.Field)
		if !ok {
			t.Fatal("the picker has no filter field")
		}
		lists := uxtest.PanelsOfType[*unison.List[*grantLibraryItem]](p.AsPanel())
		if len(lists) != 1 {
			t.Fatalf("expected one picker list, found %d", len(lists))
		}
		list := lists[0]
		filter.SetText("Acute")
		c.Equal(3, list.Count(), "precondition: the filter leaves the acute senses")
		c.Equal(0, list.Selection.FirstSet(), "precondition: the first is selected")
		filter.SetSelectionToStart()
		c.True(filter.KeyDownCallback(unison.KeyDown, mod.Shift, false), "Shift-Down is taken")
		start, end := filter.Selection()
		c.Equal(0, start, "by the filter, which selects from the caret")
		c.Equal(len("Acute"), end, "to the end of its text")
		c.Equal(1, list.Selection.Count(), "the list's selection is left alone")
		c.Equal(0, list.Selection.FirstSet())
		c.True(filter.KeyDownCallback(unison.KeyDown, 0, false), "Down alone moves the list's row")
		c.Equal(1, list.Selection.FirstSet())
		list.Select(true, 2) // A Shift-click or Cmd-click on another row.
		c.Equal(1, list.Selection.Count(), "the list holds one selected row")
		c.Equal(2, list.Selection.FirstSet(), "the one last chosen")
		c.True(filter.KeyDownCallback(unison.KeyReturn, 0, false), "Return chooses it")
	})
	screen.Do(func() { c.Equal("Acute Sense 2", g.ItemName(), "the grant gets the row the selection moved to") })
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

// TestGrantLibraryItemRefusesWhatGrants checks that the picker leaves out an item that grants something itself, whether
// the grant is among its own features or those of one of its modifiers, as SetItem strips grants from both.
func TestGrantLibraryItemRefusesWhatGrants(t *testing.T) {
	c := check.New(t)
	grant := func() *gurps.Grant {
		g := gurps.NewGrant(feature.GrantTrait)
		g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
		return g
	}
	plain := newGrantTestTrait("Plain", 0)
	plain.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
	c.NotNil(newGrantLibraryItem(plain, gurps.LibraryFile{}), "a trait whose modifier grants nothing is offered")
	direct := newGrantTestTrait("Direct", 0)
	direct.Features = gurps.Features{grant()}
	c.Nil(newGrantLibraryItem(direct, gurps.LibraryFile{}), "a trait that grants something is not")
	viaModifier := newGrantTestTrait("Via Modifier", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Features = gurps.Features{grant()}
	viaModifier.Modifiers = []*gurps.TraitModifier{modifier}
	c.Nil(newGrantLibraryItem(viaModifier, gurps.LibraryFile{}), "nor is one whose modifier grants something")
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

// TestGrantItemEditorFollowsTheGrantsType checks that the editor opened on a grant's item, and the hash by which the
// panel tells that the item changed, go by the grant's type, as the grant itself does, when a loaded grant also holds
// an item of another kind.
func TestGrantItemEditorFollowsTheGrantsType(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	owner := gurps.NewTrait(entity, nil, false)
	owner.Name = "Training"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	skillOnly := grantItemHash(g)
	g.Trait = newGrantTestTrait("Night Vision", 0)
	c.Equal(skillOnly, grantItemHash(g), "a stray trait doesn't change which item the grant gives")
	owner.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{owner}
	sheet.Rebuild(true)
	var skillEditors, traitEditors int
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer := EditTrait(sheet, owner)
		p := uxtest.PanelsOfType[*featuresPanel](outer.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
		skillEditors = len(AllMatchingDockables(func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.Skill, *gurps.SkillEditData])
			return ok
		}))
		traitEditors = len(AllMatchingDockables(func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
			return ok
		}))
	})
	c.Equal(1, skillEditors, "the skill the grant gives is edited")
	c.Equal(1, traitEditors, "beside the editor of what carries the grant")
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

// TestGrantItemEditorLeavesLevelAndEnabledToTheGrant checks that the editor of a grant's trait offers neither a Level
// field nor an Enabled check box, since the grant's row sets the level of the trait it puts on the sheet and that trait
// is enabled whenever the grant is in effect, so nothing would read either.
func TestGrantItemEditorLeavesLevelAndEnabledToTheGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Acute Vision", fxp.Two), gurps.LibraryFile{})
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
	hasLevelField := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*DecimalField](content), func(f *DecimalField) bool {
			return f.undoTitle == "Level"
		})
	}
	hasEnabledCheckBox := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*CheckBox](content), func(cb *CheckBox) bool {
			return cb.Text.String() == "Enabled"
		})
	}
	screen.Do(func() {
		_, isNested := nested.owner.(*grantItemOwner)
		c.True(isNested, "precondition: the other trait editor is the grant item's")
		c.True(hasLevelField(outer.content), "precondition: a trait's own editor has a Level field")
		c.True(hasEnabledCheckBox(outer.content), "and an Enabled check box")
		c.False(hasLevelField(nested.content), "the grant's row sets the level")
		c.False(hasEnabledCheckBox(nested.content), "and the granted trait is enabled while the grant is in effect")
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
		// With deep set, a modified editor is open on one of the item's modifiers too, which closes along with the
		// item's editor, since what it would save has nowhere to go either.
		for _, deep := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s deep=%t", tc.name, deep), func(t *testing.T) {
				c := check.New(t)
				screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
				sheet := newTestSheetForTemplate(t)
				entity := sheet.Entity()
				training := gurps.NewTrait(entity, nil, false)
				training.Name = "Training"
				g := gurps.NewGrant(feature.GrantTrait)
				nightVision := newGrantTestTrait("Night Vision", 0)
				modifier := gurps.NewTraitModifier(nil, nil, false)
				modifier.Name = "Long"
				nightVision.Modifiers = []*gurps.TraitModifier{modifier}
				g.SetItem(nightVision, gurps.LibraryFile{})
				training.Features = gurps.Features{g}
				entity.Traits = []*gurps.Trait{training}
				sheet.Rebuild(true)
				isTraitEditor := func(d unison.Dockable) bool {
					_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
					return ok
				}
				isModifierEditor := func(d unison.Dockable) bool {
					_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
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
				})
				if deep {
					screen.Do(func() { EditTraitModifier(nested, nested.editorData.Modifiers[0]) })
					modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t,
						screen, isModifierEditor)
					screen.Do(func() {
						modifierEditor.editorData.Name = "Longer"
						modifierEditor.MarkModified(nil)
					})
				}
				screen.Do(func() { tc.change(p, grantAt(t, p, 0)) })
				var open, modifierEditors, windows int
				screen.Do(func() {
					open = len(AllMatchingDockables(isTraitEditor))
					modifierEditors = len(AllMatchingDockables(isModifierEditor))
					windows = len(unison.Windows())
				})
				c.Equal(1, open, "the grant item's editor closes without asking, leaving the outer editor")
				c.Equal(0, modifierEditors, "as does the editor open on one of the item's modifiers")
				c.Equal(1, windows, "nothing asks about the changes: %v", wnd.Title())
			})
		}
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

	// So is the trait whose modifier is being edited, since its modifiers' grants are its own.
	screen.Do(func() {
		modifier := gurps.NewTraitModifier(entity, nil, false)
		modifier.Name = "Long"
		e.editorData.Modifiers = []*gurps.TraitModifier{modifier}
		EditTraitModifier(e, modifier)
	})
	modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	names = nil
	screen.Do(func() {
		p := uxtest.PanelsOfType[*featuresPanel](modifierEditor.content)[0]
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{fit, newGrantTestTrait("Night Vision", 0)})
		for _, item := range p.grantable(items) {
			names = append(names, item.name)
		}
	})
	c.Equal([]string{"Night Vision"}, names, "the trait whose modifier is being edited is left out")
}

// TestDeletingContainerKeepsGrantedRows checks that deleting a container the user moved a granted row into leaves that
// row on the sheet, since a granted row goes only with its grant.
func TestDeletingContainerKeepsGrantedRows(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var sheet *Sheet
	var granted *gurps.Skill
	var group *gurps.Skill
	isSkillEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Skill, *gurps.SkillEditData])
		return ok
	}
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
		EditSkill(sheet, granted)
		sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		DeleteSelection(sheet.Skills.Table, true)
	})
	var kept bool
	var groups, editors int
	screen.Do(func() {
		for _, s := range sheet.Entity().Skills {
			if s == granted {
				kept = true
			}
			if s.Container() {
				groups++
			}
		}
		editors = len(AllMatchingDockables(isSkillEditor))
	})
	c.Equal(0, groups, "the container goes")
	c.True(kept, "the granted row within it stays, as the row it was")
	c.Equal(1, editors, "and so does its editor")
}

// newTestSheetWithEquipmentGrant returns a sheet holding a piece of carried equipment that grants the skill Brawling,
// along with the carrier and the granted row.
func newTestSheetWithEquipmentGrant(t *testing.T) (sheet *Sheet, carrier *gurps.Equipment, granted *gurps.Skill) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	carrier = gurps.NewEquipment(entity, nil, false)
	carrier.Name = "Goggles"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	carrier.Features = gurps.Features{g}
	entity.CarriedEquipment = []*gurps.Equipment{carrier}
	sheet.Rebuild(true)
	for _, one := range entity.Skills {
		if gurps.IsGranted(one) {
			granted = one
		}
	}
	return sheet, carrier, granted
}

// TestUndoOfTakingGrantsAwayRestoresGrantedRowInOtherList checks that undoing an edit that takes a carrier's grants
// away -- deleting the carrier, applying an editor that removes the grant, disabling the carrier or moving it to the
// other equipment list -- brings back the row the grant put in the skills list as the row it was, where the user had
// moved it, rather than having the restored grant make a fresh one at the end of the list.
func TestUndoOfTakingGrantsAwayRestoresGrantedRowInOtherList(t *testing.T) {
	for _, tc := range []struct {
		name string
		// equipment is set when the carrier is a piece of carried equipment rather than a trait.
		equipment bool
		edit      func(sheet *Sheet, trait *gurps.Trait, eqp *gurps.Equipment)
	}{
		{name: "delete selection", edit: func(sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, trait)
			DeleteSelection(sheet.Traits.Table, true)
		}},
		{name: "editor removes the grant", edit: func(sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			e, _ := buildEditorContent(sheet, trait, initTraitEditor)
			e.editorData.Features = nil
			e.applyEdits()
		}},
		{name: "toggle enablement", edit: func(sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, trait)
			toggleDisabled(sheet, sheet.Traits.Table)
		}},
		{name: "move to other equipment", equipment: true, edit: func(sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, eqp)
			moveSelectedEquipment(sheet, sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table)
		}},
		{name: "drag to other equipment", equipment: true, edit: func(sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			// Driven the way unison drives a drop from one table into another: the undo data is collected, the rows
			// are moved and the drop completion is called.
			from, to := sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table
			selectGrantingItem(sheet.Skills.Table, eqp)
			draggedTableData = &unison.TableDragData[*Node[*gurps.Equipment]]{Table: from, Rows: from.SelectedRows(true)}
			defer func() { draggedTableData = nil }()
			undo := willDropCallback(from, to, true)
			entity := sheet.Entity()
			entity.CarriedEquipment = nil
			entity.OtherEquipment = []*gurps.Equipment{eqp}
			from.SyncToModel()
			to.SyncToModel()
			to.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
			didDropCallback(undo, from, to, true)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			var sheet *Sheet
			var trait *gurps.Trait
			var eqp *gurps.Equipment
			var granted *gurps.Skill
			if tc.equipment {
				sheet, eqp, granted = newTestSheetWithEquipmentGrant(t)
			} else {
				sheet, trait, granted, _ = newTestSheetWithGrant(t)
			}
			if granted == nil {
				t.Fatal("the grant must put its row on the sheet")
			}
			entity := sheet.Entity()
			group := gurps.NewSkill(entity, nil, true)
			group.Name = "Group"
			entity.Skills = slices.DeleteFunc(entity.Skills, func(s *gurps.Skill) bool { return s == granted })
			granted.SetParent(group)
			group.Children = []*gurps.Skill{granted}
			entity.Skills = append(entity.Skills, group)
			sheet.Rebuild(true)
			tc.edit(sheet, trait, eqp)
			grantedSkill := func() *gurps.Skill {
				var found *gurps.Skill
				gurps.Traverse(func(s *gurps.Skill) bool {
					if gurps.IsGranted(s) {
						found = s
					}
					return false
				}, false, false, entity.Skills...)
				return found
			}
			c.Nil(grantedSkill(), "the granted row goes with its grant")
			if !sheet.undoMgr.CanUndo() {
				t.Fatal("the edit must be undoable")
			}
			sheet.undoMgr.Undo()
			restored := grantedSkill()
			if restored == nil {
				t.Fatal("the granted row must be back")
			}
			c.Equal(granted.ID(), restored.ID(), "the granted row comes back as the row it was")
			c.True(restored.Parent() == group, "and where the user had put it")
		})
	}
}

// TestMovingGrantedRowLeavesItAlone checks that what a move does to the rows moved, filling in an empty tech level and
// recalculating the level, leaves a granted row alone, since it changes only with its grant.
func TestMovingGrantedRowLeavesItAlone(t *testing.T) {
	c := check.New(t)
	sheet, _, granted, _ := newTestSheetWithGrant(t)
	sheet.Entity().Profile.TechLevel = "3"
	techLevel := ""
	granted.TechLevel = &techLevel
	table := sheet.Skills.Table
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true})
	provider, ok := any(table.Model).(TableProvider[*gurps.Skill])
	if !ok {
		t.Fatal("the skills table has no provider")
	}
	provider.ProcessDropData(nil, table)
	c.Equal("", *granted.TechLevel, "the granted row's tech level is left as its grant made it")
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
