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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
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
// names; see loadTestFileEditor.
func loadedTestNameGeneratorEditorDockable(t *testing.T, c check.Checker, ref *library.NamedFileRef) *nameGeneratorEditorDockable {
	t.Helper()
	d := newNameGeneratorEditorDockable()
	loadTestFileEditor(t, c, &d.fileEditorDockable, ref)
	return d
}

// namesFileRef writes the content to a .names file in a fresh temporary directory and returns a reference to it of the
// kind a library scan or the toolbar menu's Open… produces.
func namesFileRef(t *testing.T, c check.Checker, name, content string) *library.NamedFileRef {
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

// TestNameGeneratorEditorContract runs the checks every file editor must pass against the name generator editor.
func TestNameGeneratorEditorContract(t *testing.T) {
	checkFileEditorContract(t, fileEditorContract[*gurps.NameGenerator]{
		newEditor: func() *fileEditorDockable[*gurps.NameGenerator] {
			return &newNameGeneratorEditorDockable().fileEditorDockable
		},
		titlePrefix: "Name Generator",
		ext:         gurps.NamesExt,
		fileName:    "Elven",
		validJSON:   testNameGeneratorJSON,
		newModel:    gurps.NewNameGenerator,
		read:        gurps.ReadNameGeneratorFromFS,
		edit: func(t *testing.T, d *fileEditorDockable[*gurps.NameGenerator]) {
			listPanelFor(t, d, &d.model.Entries).addOption()
			stringFieldFor(t, d, d.model.Entries[len(d.model.Entries)-1].KeyPrefix+"value").SetText("Arwen")
		},
	})
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

// TestOpenNameGeneratorInEditorIgnoresUnknownName verifies that asking to edit a generator no library holds does
// nothing at all: with no document dock to place an editor in, showing one would fail, so returning quietly is the
// proof.
func TestOpenNameGeneratorInEditorIgnoresUnknownName(t *testing.T) {
	c := check.New(t)
	useTestLibraries(t, c)
	swapForTest(t, &Workspace.DocumentDock, nil)
	c.NotPanics(func() { OpenNameGeneratorInEditor("No Such Generator") })
}
