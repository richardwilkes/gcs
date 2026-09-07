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
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// The file editor base needs a real model and content to work on. What every kind of editor must do is checked by
// checkFileEditorContract, which each kind's tests call with a description of it; the tests here that use the ancestry
// editor depend on nothing about its content.

// fileEditorContract describes one kind of file editor to checkFileEditorContract: how to make one, and what
// distinguishes it from the other kinds.
type fileEditorContract[T fileEditorModel[T]] struct {
	// newEditor makes an editor as its constructor does: holding a blank model, with no content built yet. The checks
	// finish it with wireTestFileEditor or loadTestFileEditor.
	newEditor func() *fileEditorDockable[T]
	// titlePrefix is what the title starts with, such as "Ancestry".
	titlePrefix string
	// ext is the extension of the files the editor edits.
	ext string
	// fileName is the name, without extension, that the checks give the files they write and load.
	fileName string
	// validJSON is the content of a file that loads into something other than a blank model.
	validJSON string
	// newModel returns a blank model, as the editor's constructor gives it.
	newModel func() T
	// read reads a file of the editor's kind, as opening one does.
	read func(fileSystem fs.FS, filePath string) (T, error)
	// edit makes one change to the model through the editor's widgets, as a user would, failing the test if they are
	// not there. What the change is does not matter, but it must survive being saved and loaded.
	edit func(t *testing.T, d *fileEditorDockable[T])
}

// checkFileEditorContract runs, as subtests, the checks every kind of file editor must pass -- those of the behavior
// the base provides.
func checkFileEditorContract[T fileEditorModel[T]](t *testing.T, contract fileEditorContract[T]) {
	t.Helper()
	untitled := contract.titlePrefix + ": Untitled"
	fileName := contract.fileName + contract.ext
	newEditor := func() *fileEditorDockable[T] {
		d := contract.newEditor()
		wireTestFileEditor(d, contract.newModel())
		return d
	}

	// A freshly created editor shows as untitled, is unmodified, has no file, and has nothing to undo.
	t.Run("StartsUntitled", func(t *testing.T) {
		c := check.New(t)
		d := newEditor()
		c.Equal(untitled, d.Title())
		c.False(d.Modified(), "a new model is unmodified")
		c.Equal("Untitled"+contract.ext, d.BackingFilePath())
		c.Equal("", d.Tooltip(), "with no file there is no path to show")
		c.False(d.undoMgr.CanUndo(), "nothing to undo")
	})

	// The toolbar's Save button is enabled exactly while there are unsaved changes.
	t.Run("SaveButtonTracksModified", func(t *testing.T) {
		c := check.New(t)
		d := newEditor()
		d.addToStartToolbar(unison.NewPanel())
		c.False(d.saveButton.Enabled(), "nothing to save yet")
		contract.edit(t, d)
		c.True(d.Modified(), "the edit is a modification")
		c.True(d.saveButton.Enabled(), "a change enables Save")
		d.markSaved()
		c.False(d.saveButton.Enabled(), "saving disables it again")
	})

	// A file that fails to load leaves the editor exactly as it was made: holding a new, untitled model with no file.
	t.Run("LoadInvalidFileChangesNothing", func(t *testing.T) {
		c := check.New(t)
		d := contract.newEditor()
		before := gurps.Hash64(d.model)
		c.HasError(d.load(testFileRef(t, c, "Broken", contract.ext, "this is not a file the editor can load")))
		c.Equal(before, gurps.Hash64(d.model), "the model is untouched")
		c.Equal("", d.path, "no path is recorded")
		c.Equal(untitled, d.Title())
		c.False(d.Modified())
	})

	// Reset replaces the model with a new one while the editor keeps its file, so that it shows as modified until
	// saved, and rebuilds the content around the new model; undoing brings back the previous model.
	t.Run("Reset", func(t *testing.T) {
		c := check.New(t)
		ref := testFileRef(t, c, contract.fileName, contract.ext, contract.validJSON)
		d := contract.newEditor()
		loadTestFileEditor(t, c, d, ref)
		loadedHash := gurps.Hash64(d.model)
		blankHash := gurps.Hash64(contract.newModel())
		c.NotEqual(blankHash, loadedHash, "precondition: the file holds something other than a new model")

		d.reset()
		c.Equal(ref.DiskPath, d.path, "reset keeps the file")
		c.Equal(contract.titlePrefix+": "+contract.fileName, d.Title(), "so the editor is still known by it")
		c.Equal(blankHash, gurps.Hash64(d.model), "reset returns to a new model")
		c.True(d.Modified(), "which is not what the file holds")
		c.True(d.undoMgr.CanUndo(), "the reset is undoable")
		contract.edit(t, d)
		c.NotEqual(blankHash, gurps.Hash64(d.model), "the content is rebuilt around the new model, so an edit reaches it")

		for d.undoMgr.CanUndo() {
			d.undoMgr.Undo()
		}
		c.Equal(loadedHash, gurps.Hash64(d.model), "undoing everything restores the loaded model")
		c.Equal(ref.DiskPath, d.path)
		c.False(d.Modified(), "and the editor is back in step with its file")
	})

	// Save on a model that has a file writes the model to it and leaves the editor unmodified.
	t.Run("SaveWritesFile", func(t *testing.T) {
		c := check.New(t)
		failOnWorkspaceError(t)
		d := newEditor()
		dir := t.TempDir()
		d.path = filepath.Join(dir, fileName)
		contract.edit(t, d)
		c.True(d.Modified(), "precondition: there is something to save")

		c.True(d.save(false), "save must succeed")
		c.False(d.Modified(), "the saved model is unmodified")
		loaded, err := contract.read(os.DirFS(dir), fileName)
		c.NoError(err, "the written file must load")
		c.Equal(gurps.Hash64(d.model), gurps.Hash64(loaded), "the file holds exactly what the editor holds")
	})
}

// Each kind of editor must be recognized as itself and as nothing else, and an unrelated dockable as neither.
func TestFileEditorRecognizersTellTheEditorsApart(t *testing.T) {
	c := check.New(t)
	ancestry := newTestAncestryEditorDockable(gurps.NewAncestry())
	names := newTestNameGeneratorEditorDockable(gurps.NewNameGenerator())
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	c.True(isAncestryEditor(ancestry))
	c.False(isAncestryEditor(names))
	c.False(isAncestryEditor(recorder))
	c.True(isNameGeneratorEditor(names))
	c.False(isNameGeneratorEditor(ancestry))
	c.False(isNameGeneratorEditor(recorder))
}

// failOnWorkspaceError makes any error the workspace would show in a dialog fail the test instead.
func failOnWorkspaceError(t *testing.T) {
	t.Helper()
	swapForTest(t, &Workspace.ErrorHandler, func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) })
}

// builtInAncestryRef returns a reference to an ancestry file that, like one built into the application, has no path on
// disk.
func builtInAncestryRef(t *testing.T, c check.Checker, name, content string) *gurps.NamedFileRef {
	t.Helper()
	ref := ancestryFileRef(t, c, name, content)
	ref.DiskPath = ""
	return ref
}

// Undoing an edit that has been saved must leave the editor showing as modified, since what it then holds is not what
// its file holds, so Save is offered and closing prompts; redoing the edit brings it back into step with the file.
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

// A file with no path on disk, as a built-in one has, is known by its name for as long as it is loaded: in the title
// and as the file name a Save As is offered, so a copy saved into a library takes precedence over the built-in as the
// user expects. A reset changes the content, not the file, so the name stays.
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

// Opening a file that fails to load must make no editor at all: the error is returned to the caller to report, and with
// no document dock to place an editor in, showing one would fail, so returning quietly proves none was shown.
func TestOpenFileEditorBrokenFileOpensNothing(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &Workspace.DocumentDock, nil)
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

// A reference made from a path on disk must name the file the way a library scan would, and read the same file.
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
