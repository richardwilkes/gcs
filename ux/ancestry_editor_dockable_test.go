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

// loadedTestAncestryEditorDockable builds an ancestry editor holding the ancestry in the file the reference names, the
// way opening the file does: the file is loaded into a fresh editor before its content is built.
func loadedTestAncestryEditorDockable(t *testing.T, c check.Checker, ref *gurps.NamedFileRef, nameGeneratorChoices ...string) *ancestryEditorDockable {
	t.Helper()
	d := newAncestryEditorDockable()
	d.nameGeneratorLookup = func() []string { return nameGeneratorChoices }
	c.NoError(d.load(ref))
	buildTestFileEditor(&d.fileEditorDockable)
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

// panelsOfType returns every panel of the given type within the subtree rooted at root, in pre-order.
func panelsOfType[T any](root *unison.Panel) []T {
	var found []T
	var walk func(p *unison.Panel)
	walk = func(p *unison.Panel) {
		if match, ok := p.Self.(T); ok {
			found = append(found, match)
		}
		for _, child := range p.Children() {
			walk(child)
		}
	}
	walk(root)
	return found
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

// TestAncestryEditorStartsUntitled verifies the state of a freshly created editor: it shows as untitled, is unmodified,
// has no file, and has nothing to undo.
func TestAncestryEditorStartsUntitled(t *testing.T) {
	c := check.New(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	c.Equal("Ancestry: Untitled", d.Title())
	c.False(d.Modified(), "a new ancestry is unmodified")
	c.Equal("Untitled.ancestry", d.BackingFilePath())
	c.Equal("", d.Tooltip(), "with no file there is no path to show")
	c.False(d.undoMgr.CanUndo(), "nothing to undo")
	c.True(isAncestryEditor(d), "the editor must be recognized as such")
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	c.False(isAncestryEditor(recorder), "an unrelated dockable must not be")
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

// TestAncestryEditorSaveButtonTracksModified verifies that the toolbar's Save button is enabled exactly while there are
// unsaved changes.
func TestAncestryEditorSaveButtonTracksModified(t *testing.T) {
	c := check.New(t)
	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	d.addToStartToolbar(unison.NewPanel())
	c.False(d.saveButton.Enabled(), "nothing to save yet")
	stringFieldFor(t, d, d.model.KeyPrefix+"name").SetText("Elf")
	c.True(d.saveButton.Enabled(), "a change enables Save")
	d.markSaved()
	c.False(d.saveButton.Enabled(), "saving disables it again")
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

// TestAncestryEditorLoadInvalidFileChangesNothing verifies that a file that fails to load leaves the editor exactly as
// it was made: holding a new, untitled ancestry with no file.
func TestAncestryEditorLoadInvalidFileChangesNothing(t *testing.T) {
	c := check.New(t)
	d := newAncestryEditorDockable()
	before := gurps.Hash64(d.model)
	c.HasError(d.load(ancestryFileRef(t, c, "Broken", "this is not an ancestry")))
	c.Equal(before, gurps.Hash64(d.model), "the ancestry is untouched")
	c.Equal("", d.path, "no path is recorded")
	c.Equal("Ancestry: Untitled", d.Title())
	c.False(d.Modified())
}

// TestAncestryEditorReset verifies that Reset replaces the ancestry with a new one while the editor keeps its file, so
// that it shows as modified until saved, and that undoing it brings back the previous ancestry.
func TestAncestryEditorReset(t *testing.T) {
	c := check.New(t)
	ref := ancestryFileRef(t, c, "Elf", testAncestryJSON)
	d := loadedTestAncestryEditorDockable(t, c, ref)

	d.reset()
	c.Equal(ref.DiskPath, d.path, "reset keeps the file")
	c.Equal("Ancestry: Elf", d.Title(), "so the editor is still known by it")
	c.Equal(gurps.Hash64(gurps.NewAncestry()), gurps.Hash64(d.model), "reset returns to a new ancestry")
	c.True(d.Modified(), "which is not what the file holds")
	c.Equal(2, len(genderRows(d)), "the panels show the new ancestry's genders")
	c.True(d.undoMgr.CanUndo(), "the reset is undoable")

	d.undoMgr.Undo()
	c.Equal("Elf", d.model.Name, "undo restores the loaded ancestry")
	c.Equal(ref.DiskPath, d.path)
	c.False(d.Modified(), "and the editor is back in step with its file")
}

// TestAncestryEditorSaveWritesFile verifies that Save on an ancestry that has a file writes the ancestry to it and
// leaves the editor unmodified.
func TestAncestryEditorSaveWritesFile(t *testing.T) {
	c := check.New(t)
	saved := Workspace.ErrorHandler
	t.Cleanup(func() { Workspace.ErrorHandler = saved })
	Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }

	d := newTestAncestryEditorDockable(gurps.NewAncestry())
	dir := t.TempDir()
	d.path = filepath.Join(dir, "Dwarf.ancestry")
	stringFieldFor(t, d, d.model.KeyPrefix+"name").SetText("Dwarf")
	c.True(d.Modified(), "precondition: there is something to save")

	c.True(d.save(false), "save must succeed")
	c.False(d.Modified(), "the saved ancestry is unmodified")
	loaded, err := gurps.NewAncestryFromFile(os.DirFS(dir), "Dwarf.ancestry")
	c.NoError(err, "the written file must load")
	c.Equal("Dwarf", loaded.Name)
	c.Equal(gurps.Hash64(d.model), gurps.Hash64(loaded), "the file holds exactly what the editor holds")
}
