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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// testNameGeneratorJSON is a hand-written name generator file of the kind the editor must cope with: a compound
// generator with a null child, a child that leaves the depth to the default, and a built-in training set.
const testNameGeneratorJSON = `{
	"type": "compound",
	"separator": " ",
	"compound": [
		{"type": "simple", "training_data": ["Alice", "Bob"]},
		null,
		{"type": "markov_letter", "built_in_training_data": "american_last"}
	]
}`

// newTestNameGeneratorEditorDockable builds a name generator editor around the given generator without placing it in
// the dock or creating a window, neither of which the headless test environment can do; see wireTestFileEditor.
func newTestNameGeneratorEditorDockable(g *gurps.NameGenerator) *nameGeneratorEditorDockable {
	d := newNameGeneratorEditorDockable()
	wireTestFileEditor(&d.fileEditorDockable, g)
	return d
}

// loadedTestNameGeneratorEditorDockable builds a name generator editor holding the generator in the file the reference
// names, the way opening the file does: the file is loaded into a fresh editor before its content is built.
func loadedTestNameGeneratorEditorDockable(t *testing.T, c check.Checker, ref *gurps.NamedFileRef) *nameGeneratorEditorDockable {
	t.Helper()
	d := newNameGeneratorEditorDockable()
	c.NoError(d.load(ref))
	buildTestFileEditor(&d.fileEditorDockable)
	return d
}

// namesFileRef writes the content to a .names file in a fresh temporary directory and returns a reference to it of the
// kind a library scan or the toolbar menu's Open… produces.
func namesFileRef(t *testing.T, c check.Checker, name, content string) *gurps.NamedFileRef {
	t.Helper()
	return testFileRef(t, c, name, gurps.NamesExt, content)
}

// simpleNameGenerator returns a simple generator trained on the given names, each with a weight of 1.
func simpleNameGenerator(names ...string) *gurps.NameGenerator {
	g := gurps.NewNameGenerator()
	for _, name := range names {
		g.Entries = append(g.Entries, &gurps.WeightedStringOption{Weight: 1, Value: name})
	}
	return g
}

// rootNameGeneratorPanel returns the panel editing the generator the file defines, failing the test if there is not
// exactly one.
func rootNameGeneratorPanel(t *testing.T, d *nameGeneratorEditorDockable) *nameGeneratorPanel {
	t.Helper()
	var roots []*nameGeneratorPanel
	for _, p := range panelsOfType[*nameGeneratorPanel](d.AsPanel()) {
		if p.parent == nil {
			roots = append(roots, p)
		}
	}
	if len(roots) != 1 {
		t.Fatalf("expected exactly one root panel, found %d", len(roots))
	}
	return roots[0]
}

// TestNameGeneratorEditorStartsUntitled verifies the state of a freshly created editor: it shows as untitled, is
// unmodified, has no file, and has nothing to undo.
func TestNameGeneratorEditorStartsUntitled(t *testing.T) {
	c := check.New(t)
	d := newTestNameGeneratorEditorDockable(gurps.NewNameGenerator())
	c.Equal("Name Generator: Untitled", d.Title())
	c.False(d.Modified(), "a new generator is unmodified")
	c.Equal("Untitled.names", d.BackingFilePath())
	c.Equal("", d.Tooltip(), "with no file there is no path to show")
	c.False(d.undoMgr.CanUndo(), "nothing to undo")
	c.True(isNameGeneratorEditor(d), "the editor must be recognized as such")
	c.False(isAncestryEditor(d), "it is not the ancestry editor")
	c.False(isNameGeneratorEditor(newTestAncestryEditorDockable(gurps.NewAncestry())), "nor is that this")
}

// TestNameGeneratorEditorTitleFollowsPath verifies that the title and backing path follow the file once there is one,
// since a generator has no name of its own: ancestries select it by the base name of its file.
func TestNameGeneratorEditorTitleFollowsPath(t *testing.T) {
	c := check.New(t)
	d := newTestNameGeneratorEditorDockable(gurps.NewNameGenerator())
	p := filepath.Join(t.TempDir(), "Elven First.names")
	d.SetBackingFilePath(p)
	c.Equal("Name Generator: Elven First", d.Title())
	c.Equal(p, d.BackingFilePath())
	c.Equal(p, d.Tooltip())
}

// TestNameGeneratorEditorSaveButtonTracksModified verifies that the toolbar's Save button is enabled exactly while
// there are unsaved changes.
func TestNameGeneratorEditorSaveButtonTracksModified(t *testing.T) {
	c := check.New(t)
	d := newTestNameGeneratorEditorDockable(gurps.NewNameGenerator())
	d.addToStartToolbar(unison.NewPanel())
	c.False(d.saveButton.Enabled(), "nothing to save yet")
	listPanelFor(t, d, &d.model.Entries).addOption()
	c.True(d.saveButton.Enabled(), "a change enables Save")
	d.markSaved()
	c.False(d.saveButton.Enabled(), "saving disables it again")
}

// TestNameGeneratorEditorLoad verifies that an editor opened on a file records its path, starts out unmodified with
// nothing to undo, normalizes the loaded data, assigns key prefixes, and builds its panels from the loaded generator.
func TestNameGeneratorEditorLoad(t *testing.T) {
	c := check.New(t)
	ref := namesFileRef(t, c, "Elven", testNameGeneratorJSON)
	d := loadedTestNameGeneratorEditorDockable(t, c, ref)
	g := d.model
	c.Equal(namegen.Compound, g.Type)
	c.Equal(ref.DiskPath, d.path, "the loaded file's path is recorded")
	c.Equal("Name Generator: Elven", d.Title())
	c.False(d.Modified(), "a freshly loaded generator is unmodified")
	c.False(d.undoMgr.CanUndo(), "opening a file is not an edit")
	c.True(d.showsFile(ref), "the editor knows which file it holds")
	c.Equal(2, len(g.Compound), "the null child is dropped")
	c.Equal(0, g.Compound[1].Depth, "an omitted depth stays at zero, so the file is not altered by being opened")
	c.NotEqual("", g.KeyPrefix, "the generator gets a key prefix")
	c.NotEqual("", g.Compound[0].KeyPrefix, "each child gets a key prefix")
	c.NotEqual("", g.Compound[0].Entries[1].KeyPrefix, "each training name gets a key prefix")
	c.Equal(3, len(panelsOfType[*nameGeneratorPanel](d.AsPanel())), "the panels are built: the root and two children")
	c.Equal("Bob", stringFieldFor(t, d, g.Compound[0].Entries[1].KeyPrefix+"value").Text())
	c.Equal(" ", stringFieldFor(t, d, g.KeyPrefix+"separator").Text())
}

// TestNameGeneratorEditorLoadInvalidFileChangesNothing verifies that a file that fails to load leaves the editor exactly
// as it was made: holding a new, untitled generator with no file.
func TestNameGeneratorEditorLoadInvalidFileChangesNothing(t *testing.T) {
	c := check.New(t)
	d := newNameGeneratorEditorDockable()
	before := gurps.Hash64(d.model)
	c.HasError(d.load(namesFileRef(t, c, "Broken", "this is not a name generator")))
	c.Equal(before, gurps.Hash64(d.model), "the generator is untouched")
	c.Equal("", d.path, "no path is recorded")
	c.Equal("Name Generator: Untitled", d.Title())
	c.False(d.Modified())
}

// TestNameGeneratorEditorReset verifies that Reset replaces the generator with a new one while the editor keeps its
// file, so that it shows as modified until saved, and that undoing it brings back the previous generator.
func TestNameGeneratorEditorReset(t *testing.T) {
	c := check.New(t)
	ref := namesFileRef(t, c, "Elven", testNameGeneratorJSON)
	d := loadedTestNameGeneratorEditorDockable(t, c, ref)

	d.reset()
	c.Equal(ref.DiskPath, d.path, "reset keeps the file")
	c.Equal("Name Generator: Elven", d.Title(), "so the editor is still known by it")
	c.Equal(gurps.Hash64(gurps.NewNameGenerator()), gurps.Hash64(d.model), "reset returns to a new generator")
	c.True(d.Modified(), "which is not what the file holds")
	c.Equal(1, len(panelsOfType[*nameGeneratorPanel](d.AsPanel())), "the panels show the new generator alone")
	c.True(d.undoMgr.CanUndo(), "the reset is undoable")

	d.undoMgr.Undo()
	c.Equal(namegen.Compound, d.model.Type, "undo restores the loaded generator")
	c.Equal(ref.DiskPath, d.path)
	c.False(d.Modified(), "and the editor is back in step with its file")
}

// TestNameGeneratorEditorSaveWritesFile verifies that Save on a generator that has a file writes the generator to it
// and leaves the editor unmodified.
func TestNameGeneratorEditorSaveWritesFile(t *testing.T) {
	c := check.New(t)
	saved := Workspace.ErrorHandler
	t.Cleanup(func() { Workspace.ErrorHandler = saved })
	Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }

	d := newTestNameGeneratorEditorDockable(simpleNameGenerator("Alice"))
	dir := t.TempDir()
	d.path = filepath.Join(dir, "Elven.names")
	stringFieldFor(t, d, d.model.Entries[0].KeyPrefix+"value").SetText("Arwen")
	c.True(d.Modified(), "precondition: there is something to save")

	c.True(d.save(false), "save must succeed")
	c.False(d.Modified(), "the saved generator is unmodified")
	loaded, err := gurps.ReadNameGeneratorFromFS(os.DirFS(dir), "Elven.names")
	c.NoError(err, "the written file must load")
	c.Equal([]string{"Arwen"}, optionValues(loaded.Entries))
	c.Equal(gurps.Hash64(d.model), gurps.Hash64(loaded), "the file holds exactly what the editor holds")
}

// TestOpenNameGeneratorInEditorIgnoresUnknownName verifies that asking to edit a generator no library holds does
// nothing at all: with no document dock to place an editor in, showing one would fail, so returning quietly is the
// proof.
func TestOpenNameGeneratorInEditorIgnoresUnknownName(t *testing.T) {
	c := check.New(t)
	useTestLibraries(t, c)
	saved := Workspace
	t.Cleanup(func() { Workspace = saved })
	Workspace.DocumentDock = nil
	c.NotPanics(func() { OpenNameGeneratorInEditor("No Such Generator") })
}
