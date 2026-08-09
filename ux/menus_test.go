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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// useTestLibraries points the global settings at a fresh library set rooted in a temporary directory, restoring the
// original set when the test finishes. It returns the master and user libraries, both of which start out with no
// "Output Templates" directory at all.
func useTestLibraries(t *testing.T, c check.Checker) (master, user *gurps.Library) {
	t.Helper()
	global := gurps.GlobalSettings()
	saved := global.LibrarySet
	t.Cleanup(func() { global.LibrarySet = saved })
	global.LibrarySet = gurps.NewLibraries()
	dir := t.TempDir()
	master = global.LibrarySet.Master()
	c.NoError(master.SetPath(filepath.Join(dir, "master")))
	user = global.LibrarySet.User()
	c.NoError(user.SetPath(filepath.Join(dir, "user")))
	return master, user
}

// addOutputTemplates creates the library's "Output Templates" directory and populates it with the named files. Passing
// no names leaves the directory empty.
func addOutputTemplates(c check.Checker, lib *gurps.Library, names ...string) {
	dir := filepath.Join(lib.Path(), "Output Templates")
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
// prologue of image export formats and the separator after them.
func exportToMenuTitles(c check.Checker, menu unison.Menu) []string {
	var s menuBarScope
	s.exportToUpdater(menu)
	titles := menuItemTitles(menu)
	c.Equal([]string{"PDF", "WEBP", "PNG", "JPEG", "---"}, titles[:5], "fixed prologue")
	return titles[5:]
}

// TestExportToMenuWithoutTemplateDirs verifies that the placeholder is shown when no library has an "Output Templates"
// directory at all. This is the regression the stale count check caused: the fixed prologue of four export formats plus
// its separator already puts the menu item count at 5, so testing for a count of 2 could never be true and the
// placeholder was never displayed.
func TestExportToMenuWithoutTemplateDirs(t *testing.T) {
	c := check.New(t)
	useTestLibraries(t, c)
	c.Equal([]string{"No export templates available"}, exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithEmptyMasterTemplateDir verifies that the placeholder is still shown when the master library has
// an "Output Templates" directory that holds no templates. The master library always contributes its title in that
// case, so a check against a fixed item count would miss this and leave a dangling library title with nothing under it.
func TestExportToMenuWithEmptyMasterTemplateDir(t *testing.T) {
	c := check.New(t)
	master, _ := useTestLibraries(t, c)
	addOutputTemplates(c, master)
	c.Equal([]string{master.Data().Title, "No export templates available"},
		exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithHiddenFilesOnly verifies that a directory holding nothing but dot files, which are skipped, is
// treated the same as an empty one.
func TestExportToMenuWithHiddenFilesOnly(t *testing.T) {
	c := check.New(t)
	master, _ := useTestLibraries(t, c)
	addOutputTemplates(c, master, ".DS_Store")
	c.Equal([]string{master.Data().Title, "No export templates available"},
		exportToMenuTitles(c, newExportToMenu()), "menu body")
}

// TestExportToMenuWithTemplates verifies that the placeholder is not shown once any library supplies a template, that
// the templates are listed under their library's title, and that they are sorted naturally.
func TestExportToMenuWithTemplates(t *testing.T) {
	c := check.New(t)
	master, user := useTestLibraries(t, c)
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
	master, _ := useTestLibraries(t, c)
	menu := newExportToMenu()
	c.Equal([]string{"No export templates available"}, exportToMenuTitles(c, menu), "before any template exists")
	addOutputTemplates(c, master, "Sheet.gcs")
	c.Equal([]string{master.Data().Title, "    Sheet"}, exportToMenuTitles(c, menu), "after a template is added")
}
