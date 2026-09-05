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

// The file editor base is exercised through the ancestry editor, since it needs a real model and content to work on;
// nothing here depends on what the content is.

// failOnWorkspaceError makes any error the workspace would show in a dialog fail the test instead, for the duration of
// the test.
func failOnWorkspaceError(t *testing.T) {
	t.Helper()
	saved := Workspace.ErrorHandler
	t.Cleanup(func() { Workspace.ErrorHandler = saved })
	Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }
}

// builtInAncestryRef returns a reference to an ancestry file that, like one built into the application, has no path on
// disk.
func builtInAncestryRef(t *testing.T, c check.Checker, name, content string) *gurps.NamedFileRef {
	t.Helper()
	ref := ancestryFileRef(t, c, name, content)
	ref.DiskPath = ""
	return ref
}

// TestFileEditorUndoPastSaveLeavesModified verifies that undoing an edit that has been saved leaves the editor showing
// as modified, since what it then holds is not what its file holds, so that Save is offered and closing prompts; and
// that redoing the edit brings it back into step with the file.
func TestFileEditorUndoPastSaveLeavesModified(t *testing.T) {
	c := check.New(t)
	failOnWorkspaceError(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	d.addToStartToolbar(unison.NewPanel())
	d.path = filepath.Join(t.TempDir(), "Dwarf.ancestry")
	d.editStructure("Rename", func() { d.model.Name = "Dwarf" }, "")
	c.True(d.save(false), "save must succeed")
	c.False(d.Modified(), "the saved ancestry is unmodified")
	c.False(d.saveButton.Enabled())

	d.undoMgr.Undo()
	c.Equal("", d.model.Name, "undo reverses the edit")
	c.True(d.Modified(), "the file holds the edit, so the editor is modified again")
	c.True(d.saveButton.Enabled(), "and Save is offered")

	d.undoMgr.Redo()
	c.Equal("Dwarf", d.model.Name)
	c.False(d.Modified(), "redo puts the editor back in step with the file")
	c.False(d.saveButton.Enabled())
}

// TestFileEditorBuiltInFileKeepsItsName verifies that a file with no path on disk, as a built-in one has, is known by
// its name for as long as it is loaded: in the title and as the file name a Save As is offered, so that a copy saved
// into a library takes precedence over the built-in as the user expects. A reset changes the content, not the file, so
// the name stays.
func TestFileEditorBuiltInFileKeepsItsName(t *testing.T) {
	c := check.New(t)
	ref := builtInAncestryRef(t, c, "Elf", testAncestryJSON)
	d := loadedTestAncestryEditorDockable(t, c, ref)
	c.Equal("Ancestry: Elf", d.Title())
	c.Equal("Elf.ancestry", d.BackingFilePath(), "the name is what a Save As offers")
	c.Equal("", d.Tooltip(), "there is no path to show")
	c.True(d.showsFile(ref), "the editor knows which file it holds")
	c.False(d.showsFile(ancestryFileRef(t, c, "Elf", testAncestryJSON)), "a file on disk of the same name is another file")

	d.reset()
	c.Equal("Ancestry: Elf", d.Title(), "a reset keeps the name")
	c.True(d.showsFile(ref))
	c.True(d.Modified(), "and leaves the editor modified, since the built-in holds something else")

	// Saving the file somewhere gives it a path, which takes over from the name.
	dir := t.TempDir()
	d.path = filepath.Join(dir, "Wood Elf.ancestry")
	d.loadedName = ""
	c.Equal("Ancestry: Wood Elf", d.Title())
	c.False(d.showsFile(ref), "the editor no longer holds the built-in")
}

// TestOpenFileEditorBrokenFileOpensNothing verifies that opening a file that fails to load makes no editor at all: the
// error is returned to the caller to report, and with no document dock to place an editor in, showing one would fail,
// so returning quietly is the proof that none was shown.
func TestOpenFileEditorBrokenFileOpensNothing(t *testing.T) {
	c := check.New(t)
	saved := Workspace
	t.Cleanup(func() { Workspace = saved })
	Workspace.DocumentDock = nil
	ref := ancestryFileRef(t, c, "Broken", "this is not an ancestry")
	var d *ancestryEditorDockable
	var err error
	c.NotPanics(func() { d, err = openAncestryRef(ref) })
	c.HasError(err)
	c.Nil(d, "no editor is handed back")
	dockable, err := openAncestryFile(ref.DiskPath)
	c.HasError(err)
	c.True(dockable == nil, "the file type registry's loader hands back a nil dockable, not a typed nil")
}

// TestDiskFileRef verifies that a reference made from a path on disk names the file the way a library scan would, and
// reads the same file.
func TestDiskFileRef(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "Elf.ancestry")
	c.NoError(os.WriteFile(p, []byte(testAncestryJSON), 0o640))
	ref := diskFileRef(p)
	c.Equal("Elf", ref.Name)
	c.Equal("Elf.ancestry", ref.FilePath)
	c.Equal(p, ref.DiskPath)
	a, err := gurps.NewAncestryFromFile(ref.FileSystem, ref.FilePath)
	c.NoError(err)
	c.Equal("Elf", a.Name)
}
