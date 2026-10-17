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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// addOutputTemplates creates the library's "Output Templates" directory and populates it with the named files. Passing
// no names leaves the directory empty.
func addOutputTemplates(c check.Checker, lib *library.Library, names ...string) {
	dir := filepath.Join(lib.Path(false), library.OutputTemplatesDirName)
	c.NoError(os.MkdirAll(dir, 0o750))
	for _, name := range names {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("template"), 0o640))
	}
}

// menuItemTitles returns the titles of the items in the menu, with separators reported as "---".
func menuItemTitles(menu unison.Menu) []string {
	titles := make([]string, 0, menu.Count())
	for i := range menu.Count() {
		item := menu.ItemAtIndex(i)
		if item.IsSeparator() {
			titles = append(titles, "---")
			continue
		}
		titles = append(titles, item.Title())
	}
	return titles
}

// newExportToMenu builds the "Export To…" menu the same way the menu bar does.
func newExportToMenu() unison.Menu {
	var s menuBarScope
	registerKeyBindingsOnce.Do(registerActions)
	return unison.NewInWindowMenuFactory().NewMenu(ExportToMenuID, "Export To…", s.exportToUpdater)
}

// exportToMenuTitles builds and populates the "Export To…" menu, returning the titles of the items following the fixed
// prologue of image export formats, the portrait export and the separators after them.
func exportToMenuTitles(c check.Checker, menu unison.Menu) []string {
	var s menuBarScope
	s.exportToUpdater(menu)
	titles := menuItemTitles(menu)
	c.Equal([]string{"PDF", "WEBP", "PNG", "JPEG", "---", "Export Portrait", "---"}, titles[:7], "fixed prologue")
	return titles[7:]
}

// TestExportToMenuWithoutTemplateDirs verifies that the placeholder is shown when no library has an "Output Templates"
// directory at all. A stale check for an item count of 2 once kept it from ever showing, since the fixed prologue alone
// already holds more items than that.
func TestExportToMenuWithoutTemplateDirs(t *testing.T) {
	c := check.New(t)
	uxtest.UseTestLibraries(t, c)
	c.Equal([]string{"No export templates available"}, exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithEmptyMasterTemplateDir verifies that the placeholder is still shown when the master library has
// an "Output Templates" directory that holds no templates. The master library always contributes its title in that
// case, so a check against a fixed item count would miss this and leave a dangling library title with nothing under it.
func TestExportToMenuWithEmptyMasterTemplateDir(t *testing.T) {
	c := check.New(t)
	master, _ := uxtest.UseTestLibraries(t, c)
	addOutputTemplates(c, master)
	c.Equal([]string{master.Data().Title, "No export templates available"},
		exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithHiddenFilesOnly verifies that a directory holding nothing but dot files, which are skipped, is
// treated the same as an empty one.
func TestExportToMenuWithHiddenFilesOnly(t *testing.T) {
	c := check.New(t)
	master, _ := uxtest.UseTestLibraries(t, c)
	addOutputTemplates(c, master, ".DS_Store")
	c.Equal([]string{master.Data().Title, "No export templates available"},
		exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithTemplates verifies that the placeholder is not shown once any library supplies a template, that
// the templates are listed under their library's title, and that they are sorted naturally.
func TestExportToMenuWithTemplates(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	addOutputTemplates(c, master, "Sheet 10.gcs", "Sheet 2.gcs")
	addOutputTemplates(c, user, "Mine.gcs")
	c.Equal([]string{
		user.Data().Title,
		"    Mine",
		master.Data().Title,
		"    Sheet 2",
		"    Sheet 10",
	}, exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuRepopulates verifies that rebuilding the menu, which happens each time it is opened, reflects the
// templates present at that moment rather than accumulating the previous contents.
func TestExportToMenuRepopulates(t *testing.T) {
	c := check.New(t)
	master, _ := uxtest.UseTestLibraries(t, c)
	menu := newExportToMenu()
	c.Equal([]string{"No export templates available"}, exportToMenuTitles(c, menu), "before any template exists")
	addOutputTemplates(c, master, "Sheet.gcs")
	c.Equal([]string{master.Data().Title, "    Sheet"}, exportToMenuTitles(c, menu), "after a template is added")
}

// menuBarTitles builds the menu bar the way SetupMenuBar does, with the in-window menu factory, and returns the titles
// of the menus in it, along with the bar itself for looking into those menus.
func menuBarTitles() (bar unison.Menu, titles []string) {
	registerKeyBindingsOnce.Do(registerActions)
	bar = unison.NewInWindowMenuFactory().NewMenu(0, "", nil)
	menuBarScope{}.populate(bar)
	return bar, menuItemTitles(bar)
}

// TestMenuBarStructure pins down the arrangement of the menu bar: which menus it holds, in what order, and what each
// of the application's own menus and sub-menus holds. The items the platform contributes, such as Quit and the Window
// menu's own entries, are left to unison, so the File menu is checked only up to its last application item.
func TestMenuBarStructure(t *testing.T) {
	c := check.New(t)
	bar, titles := menuBarTitles()
	c.Equal([]string{"File", "Edit", "Insert", "Item", "Sheet", "Settings", "View", "Window", "Help"}, titles, "menu bar")

	file := bar.Menu(unison.FileMenuID)
	titles = menuItemTitles(file)
	c.Equal([]string{"New", "Open…", "Open Recent", "Close", "---", "Save", "Save As…", "Export", "Print…", "---"},
		titles[:10], "File menu, up to the separator before the Quit item unison adds to a per-window bar")
	c.Equal([]string{
		"New Character Sheet",
		"New Character Template",
		"New Loot Sheet",
		"---",
		"New Traits Library",
		"New Trait Modifiers Library",
		"New Skills Library",
		"New Spells Library",
		"New Equipment Library",
		"New Equipment Modifiers Library",
		"New Notes Library",
		"---",
		"New Ancestry",
		"New Name Generator",
		"New Markdown File",
	}, menuItemTitles(file.Menu(FileNewMenuID)), "File > New")

	c.Equal([]string{
		undoAction.Title,
		redoAction.Title,
		"---",
		"Cut",
		"Copy",
		"Paste",
		"Duplicate",
		"Delete",
		"Select All",
		"---",
		"Jump to Search/Filter Field",
	}, menuItemTitles(bar.Menu(unison.EditMenuID)), "Edit menu")

	insert := bar.Menu(InsertMenuID)
	c.Equal([]string{"Trait", "Skill", "Spell", "Equipment", "Note", "---", "New Melee Attack", "New Ranged Attack"},
		menuItemTitles(insert), "Insert menu")
	c.Equal([]string{
		"New Trait",
		"New Trait Container",
		"New Trait Choice",
		"---",
		"New Trait Modifier",
		"New Trait Modifier Group",
		"New Trait Modifier Choice",
	}, menuItemTitles(insert.Menu(InsertTraitMenuID)), "Insert > Trait")
	c.Equal([]string{"New Skill", "New Skill Container", "New Skill Choice", "New Technique"},
		menuItemTitles(insert.Menu(InsertSkillMenuID)), "Insert > Skill")
	c.Equal([]string{"New Spell", "New Spell Container", "New Spell Choice", "New Ritual Magic Spell"},
		menuItemTitles(insert.Menu(InsertSpellMenuID)), "Insert > Spell")
	c.Equal([]string{
		"New Carried Equipment",
		"New Carried Equipment Container",
		"New Carried Equipment Group",
		"---",
		"New Other Equipment",
		"New Other Equipment Container",
		"New Other Equipment Group",
		"---",
		"New Equipment Choice",
		"---",
		"New Equipment Modifier",
		"New Equipment Modifier Group",
		"New Equipment Modifier Choice",
	}, menuItemTitles(insert.Menu(InsertEquipmentMenuID)), "Insert > Equipment")
	c.Equal([]string{"New Note", "New Note Container"}, menuItemTitles(insert.Menu(InsertNoteMenuID)), "Insert > Note")

	item := bar.Menu(ItemMenuID)
	c.Equal([]string{
		"Open Detail Editor",
		"Open Page Reference",
		"Open Each Page Reference",
		"---",
		"Move",
		"Adjust",
		"Convert",
		"Toggle State",
		"Swap Defaults",
		"Apply Modifier",
		"---",
		"Copy to Character Sheet",
		"Copy to Template",
		"---",
		"Sync with Source",
		"Clear Source",
	}, menuItemTitles(item), "Item menu")
	c.Equal([]string{
		"Move Up",
		"Move Down",
		"Move Out of Container",
		"Move Into Container",
		"---",
		"Move to Carried Equipment",
		"Move to Other Equipment",
	}, menuItemTitles(item.Menu(MoveMenuID)), "Item > Move")
	c.Equal([]string{
		"Increment",
		"Decrement",
		"---",
		"Increase Uses",
		"Decrease Uses",
		"Reset Uses to Maximum",
		"---",
		"Increase Skill Level",
		"Decrease Skill Level",
		"---",
		"Increase Tech Level",
		"Decrease Tech Level",
		"---",
		"Increase Equipment Level",
		"Decrease Equipment Level",
	}, menuItemTitles(item.Menu(AdjustMenuID)), "Item > Adjust")
	c.Equal([]string{"Convert to Container", "Convert to Non-Container", "Convert to Choice", "Convert to Group"},
		menuItemTitles(item.Menu(ConvertMenuID)), "Item > Convert")

	c.Equal([]string{
		"Apply Template to Character Sheet",
		"New Character Sheet from Template",
		"Clone Character Sheet & Re-Randomize Fields",
		"---",
		"Add Natural Attacks",
		"Organize Traits",
		"---",
		"Buy Equipment…",
		"Buy Selected Equipment…",
		"Sell Equipment…",
		"Shopping Record",
		"---",
		"Clear Portrait",
		"---",
		"Sheet Settings…",
		"Attributes…",
		"Body Type…",
		"Edit Sheet Layout",
	}, menuItemTitles(bar.Menu(SheetMenuID)), "Sheet menu")

	c.Equal([]string{
		"General Settings…",
		"Colors…",
		"Fonts…",
		"Menu Keys…",
		"Saved Filters…",
		"---",
		"Default Sheet Settings…",
		"Default Attributes…",
		"Default Body Type…",
		"---",
		"Page Reference Mappings…",
		"Download GURPS Rules Lookup File",
	}, menuItemTitles(bar.Menu(SettingsMenuID)), "Settings menu")

	view := bar.Menu(ViewMenuID)
	titles = menuItemTitles(view)
	c.Equal([]string{"Default Scale", "Scale Up", "Scale Down", "Scale To", "---", "Calculators"}, titles[:6],
		"View menu, up to the separator macOS adds ahead of the items the system contributes")
	c.Equal([]string{
		"25% Scale",
		"50% Scale",
		"75% Scale",
		"100% Scale",
		"200% Scale",
		"300% Scale",
		"400% Scale",
		"500% Scale",
		"600% Scale",
	}, menuItemTitles(view.Menu(ScaleToMenuID)), "View > Scale To")

	window := menuItemTitles(bar.Menu(unison.WindowMenuID))
	c.Equal("Undock From Workspace", window[len(window)-1], "the Window menu ends with the undock item")
}
