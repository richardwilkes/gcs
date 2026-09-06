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
	"github.com/richardwilkes/unison/enums/behavior"
)

// testAncestryJSON is a hand-written ancestry file of the kind the editor must cope with: no common options, a null
// entry in a list, a null gender, and a gender with no value.
const testAncestryJSON = `{
	"version": 5,
	"name": "Elf",
	"gender_options": [
		{"weight": 3, "value": {"name": "Male", "hair_options": [null, {"weight": 2, "value": "Silver"}]}},
		null,
		{"weight": 1}
	]
}`

// newTestAncestryEditorDockable builds an ancestry editor around the given ancestry without placing it in the dock or
// creating a window, neither of which the headless test environment can do, and without scanning the libraries for name
// generators, which the caller supplies instead.
func newTestAncestryEditorDockable(a *gurps.Ancestry, nameGeneratorChoices ...string) *ancestryEditorDockable {
	d := newAncestryEditorDockable()
	d.nameGeneratorLookup = func() []string { return nameGeneratorChoices }
	wireTestFileEditor(&d.fileEditorDockable, a)
	return d
}

// wireTestFileEditor finishes an editor its constructor has made, giving it the model to edit and building its content;
// see buildTestFileEditor.
func wireTestFileEditor[T fileEditorModel[T]](d *fileEditorDockable[T], model T) {
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.model = model
	d.originalHash = gurps.Hash64(model)
	buildTestFileEditor(d)
}

// buildTestFileEditor builds an editor's content from the model it holds as show would, but with no dock and no window:
// the content is wrapped in a scroll panel that is itself a child of the dockable, as Setup arranges, so that sync() can
// find a scroll root and the target manager, which is rooted at the dockable, can find the widgets. The undo manager is
// replaced by one that panics on an error, so that a test sees it.
func buildTestFileEditor[T fileEditorModel[T]](d *fileEditorDockable[T]) {
	d.undoMgr = unison.NewUndoManager(100, func(err error) { panic(err) })
	content := unison.NewPanel()
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
	d.AddChild(scroller)
	d.initContent(content)
}

// loadTestFileEditor finishes an editor its constructor has made the way opening a file does: the file the reference
// names is loaded into it before its content is built; see buildTestFileEditor.
func loadTestFileEditor[T fileEditorModel[T]](t *testing.T, c check.Checker, d *fileEditorDockable[T], ref *gurps.NamedFileRef) {
	t.Helper()
	c.NoError(d.load(ref))
	buildTestFileEditor(d)
}

// loadedTestAncestryEditorDockable builds an ancestry editor holding the ancestry in the file the reference names; see
// loadTestFileEditor.
func loadedTestAncestryEditorDockable(t *testing.T, c check.Checker, ref *gurps.NamedFileRef, nameGeneratorChoices ...string) *ancestryEditorDockable {
	t.Helper()
	d := newAncestryEditorDockable()
	d.nameGeneratorLookup = func() []string { return nameGeneratorChoices }
	loadTestFileEditor(t, c, &d.fileEditorDockable, ref)
	return d
}

// ancestryFileRef writes the content to an ancestry file in a fresh temporary directory and returns a reference to it
// of the kind a library scan or the toolbar menu's Open… produces.
func ancestryFileRef(t *testing.T, c check.Checker, name, content string) *gurps.NamedFileRef {
	t.Helper()
	return testFileRef(t, c, name, gurps.AncestryExt, content)
}

// testFileRef writes the content to a file with the given name and extension in a fresh temporary directory and returns
// a reference to it of the kind a library scan or the toolbar menu's Open… produces.
func testFileRef(t *testing.T, c check.Checker, name, ext, content string) *gurps.NamedFileRef {
	t.Helper()
	dir := t.TempDir()
	fileName := name + ext
	c.NoError(os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o640))
	return &gurps.NamedFileRef{
		Name:       name,
		FileSystem: os.DirFS(dir),
		FilePath:   fileName,
		DiskPath:   filepath.Join(dir, fileName),
	}
}

// widgetFor returns the widget of the given type with the given reference key within the editor, failing the test if
// there is no such widget or it is of some other type.
func widgetFor[T any](t *testing.T, d structuralEditor, key string) T {
	t.Helper()
	panel := d.targetManager().Find(key)
	if panel == nil {
		t.Fatalf("no widget with key %q", key)
	}
	widget, ok := panel.Self.(T)
	if !ok {
		t.Fatalf("the widget with key %q is a %T", key, panel.Self)
	}
	return widget
}

func stringFieldFor(t *testing.T, d structuralEditor, key string) *StringField {
	t.Helper()
	return widgetFor[*StringField](t, d, key)
}

func integerFieldFor(t *testing.T, d structuralEditor, key string) *IntegerField {
	t.Helper()
	return widgetFor[*IntegerField](t, d, key)
}

func genderRows(d *ancestryEditorDockable) []*genderOptionsPanel {
	return panelsOfType[*genderOptionsPanel](d.AsPanel())
}

// TestAncestryEditorContract runs the checks every file editor must pass against the ancestry editor.
func TestAncestryEditorContract(t *testing.T) {
	checkFileEditorContract(t, fileEditorContract[*gurps.Ancestry]{
		newEditor: func() *fileEditorDockable[*gurps.Ancestry] {
			d := newAncestryEditorDockable()
			d.nameGeneratorLookup = func() []string { return nil }
			return &d.fileEditorDockable
		},
		titlePrefix: "Ancestry",
		ext:         gurps.AncestryExt,
		fileName:    "Elf",
		validJSON:   testAncestryJSON,
		newModel:    gurps.NewAncestry,
		read:        gurps.NewAncestryFromFile,
		edit: func(t *testing.T, d *fileEditorDockable[*gurps.Ancestry]) {
			stringFieldFor(t, d, d.model.KeyPrefix+"name").SetText("Dwarf")
		},
	})
}

// TestAncestryEditorTitleFollowsNameThenPath verifies that the title tracks the Name field until the ancestry has a
// file, after which it tracks the file's base name -- the name traits select an ancestry by -- and the tooltip shows
// the path.
func TestAncestryEditorTitleFollowsNameThenPath(t *testing.T) {
	c := check.New(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	stringFieldFor(t, d, d.model.KeyPrefix+"name").SetText("Elf")
	c.Equal("Elf", d.model.Name, "typing in the Name field writes the model")
	c.Equal("Ancestry: Elf", d.Title())
	c.Equal("Elf.ancestry", d.BackingFilePath(), "the name becomes the initial file name")
	c.True(d.Modified(), "changing the name is a modification")

	p := filepath.Join(t.TempDir(), "Dwarf.ancestry")
	d.SetBackingFilePath(p)
	c.Equal("Ancestry: Dwarf", d.Title(), "once on disk, the file name is the ancestry's name")
	c.Equal(p, d.BackingFilePath())
	c.Equal(p, d.Tooltip())
}

// TestAncestryEditorLoad verifies that an editor opened on a file records its path, starts out unmodified with nothing to
// undo, normalizes the loaded data, assigns key prefixes, and builds its panels from the loaded ancestry.
func TestAncestryEditorLoad(t *testing.T) {
	c := check.New(t)
	ref := ancestryFileRef(t, c, "Elf", testAncestryJSON)
	d := loadedTestAncestryEditorDockable(t, c, ref)
	c.Equal("Elf", d.model.Name)
	c.Equal(ref.DiskPath, d.path, "the loaded file's path is recorded")
	c.Equal("Ancestry: Elf", d.Title())
	c.False(d.Modified(), "a freshly loaded ancestry is unmodified")
	c.False(d.undoMgr.CanUndo(), "opening a file is not an edit")
	c.True(d.showsFile(ref), "the editor knows which file it holds")
	c.False(d.showsFile(ancestryFileRef(t, c, "Elf", testAncestryJSON)), "another file of the same name is another file")
	c.NotNil(d.model.CommonOptions, "the common options are materialized")
	c.Equal(2, len(d.model.GenderOptions), "the null gender is dropped")
	c.NotNil(d.model.GenderOptions[1].Value, "a valueless gender gets an empty options block")
	c.Equal(1, len(d.model.GenderOptions[0].Value.HairOptions), "the null option is dropped")
	c.NotEqual("", d.model.KeyPrefix, "the ancestry gets a key prefix")
	c.NotEqual("", d.model.CommonOptions.KeyPrefix, "the common options get a key prefix")
	c.NotEqual("", d.model.GenderOptions[0].Value.HairOptions[0].KeyPrefix, "each option gets a key prefix")
	c.Equal(2, len(genderRows(d)), "the panels are built for the loaded ancestry")
	c.Equal("Silver", stringFieldFor(t, d, d.model.GenderOptions[0].Value.HairOptions[0].KeyPrefix+"value").Text())
}
