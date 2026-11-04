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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
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

// newTestAncestryEditorDockable builds an ancestry editor around the given ancestry without the dock or window the
// headless test environment cannot provide, and with the caller's name generator choices in place of a library scan.
func newTestAncestryEditorDockable(a *gurps.Ancestry, nameGeneratorChoices ...string) *ancestryEditorDockable {
	d := newAncestryEditorDockable()
	d.nameGeneratorLookup = func() []string { return nameGeneratorChoices }
	wireTestFileEditor(&d.fileEditorDockable, a)
	return d
}

func wireTestFileEditor[T fileEditorModel[T]](d *fileEditorDockable[T], model T) {
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.model = model
	d.originalHash = gurps.Hash64(model)
	buildTestFileEditor(d)
}

// buildTestFileEditor builds an editor's content as show would, but with no dock and no window. The content is wrapped
// in a scroll panel below the dockable, as Setup arranges, so that sync() can find a scroll root and the target
// manager, rooted at the dockable, can find the widgets. The undo manager is replaced by one that panics on an error,
// so that a test sees it.
func buildTestFileEditor[T fileEditorModel[T]](d *fileEditorDockable[T]) {
	d.undoMgr = unison.NewUndoManager(100, func(err error) { panic(err) })
	content := unison.NewPanel()
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
	d.AddChild(scroller)
	d.initContent(content)
}

func loadTestFileEditor[T fileEditorModel[T]](t *testing.T, c check.Checker, d *fileEditorDockable[T], ref *library.NamedFileRef) {
	t.Helper()
	c.NoError(d.load(ref))
	buildTestFileEditor(d)
}

func loadedTestAncestryEditorDockable(t *testing.T, c check.Checker, ref *library.NamedFileRef, nameGeneratorChoices ...string) *ancestryEditorDockable {
	t.Helper()
	d := newAncestryEditorDockable()
	d.nameGeneratorLookup = func() []string { return nameGeneratorChoices }
	loadTestFileEditor(t, c, &d.fileEditorDockable, ref)
	return d
}

func ancestryFileRef(t *testing.T, c check.Checker, name, content string) *library.NamedFileRef {
	t.Helper()
	return testFileRef(t, c, name, gurps.AncestryExt, content)
}

// testFileRef writes the content to a file in a fresh temporary directory and returns a reference to it of the kind a
// library scan or the toolbar menu's Open… produces.
func testFileRef(t *testing.T, c check.Checker, name, ext, content string) *library.NamedFileRef {
	t.Helper()
	dir := t.TempDir()
	fileName := name + ext
	c.NoError(os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o640))
	return &library.NamedFileRef{
		Name:       name,
		FileSystem: os.DirFS(dir),
		FilePath:   fileName,
		DiskPath:   filepath.Join(dir, fileName),
	}
}

// widgetFor returns the editor's widget with the given reference key, failing the test if there is no such widget or
// it is of some other type.
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
	return uxtest.PanelsOfType[*genderOptionsPanel](d.AsPanel())
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
// file, after which it tracks the file's base name -- the name traits select an ancestry by.
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

// TestAncestryEditorLoad verifies that an editor opened on a file records its path, starts out unmodified with nothing
// to undo, normalizes the loaded data, assigns key prefixes, and builds its panels from the loaded ancestry.
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

// TestAncestryEditorHeadless drives the ancestry editor end to end inside a headless GCS workspace, the way a user
// would: it opens the editor from the File menu, edits it by clicking and typing, reorders a list by dragging, undoes
// with the keyboard, saves through the file dialog, reloads the file from the toolbar menu, and closes the tab with
// unsaved changes. The phases build on one another, so a failure the rest cannot proceed without stops the test there.
func TestAncestryEditorHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()

	// Open the editor from the File menu.
	uxtest.ChooseMenuBarItem(t, screen, wnd, "File", "New", "New Ancestry")
	d := uxtest.SoleEditor[*ancestryEditorDockable](t, screen, isAncestryEditor)
	var title string
	var modified, saveEnabled, inWorkspace bool
	screen.Do(func() {
		title = d.Title()
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		inWorkspace = d.Window() == wnd
	})
	c.Equal("Ancestry: Untitled", title, "a new ancestry is untitled")
	c.False(modified, "a new ancestry is unmodified")
	c.False(saveEnabled, "there is nothing to save yet")
	c.True(inWorkspace, "the editor opens in the workspace window")

	// Choosing the item again opens a second editor, which stands on its own: the first is left as it was, and closing
	// the second, which is untouched, prompts for nothing and leaves the first in place.
	uxtest.ChooseMenuBarItem(t, screen, wnd, "File", "New", "New Ancestry")
	second := uxtest.OtherEditor(t, screen, d, isAncestryEditor)
	var secondTitle string
	screen.Do(func() { secondTitle = second.Title() })
	c.Equal("Ancestry: Untitled", secondTitle, "the second editor holds a new ancestry of its own")
	uxtest.CloseEditorWithoutPrompt(t, screen, second)
	var remaining []unison.Dockable
	screen.Do(func() { remaining = AllMatchingDockables(isAncestryEditor) })
	if len(remaining) != 1 || remaining[0].AsPanel().Self != d {
		t.Fatalf("expected the first editor alone to remain after closing the second, found %d", len(remaining))
	}

	// Type a name into the Name field and commit it by tabbing out.
	var nameField *unison.Panel
	screen.Do(func() { nameField = d.targetMgr.Find(d.model.KeyPrefix + "name") })
	if nameField == nil {
		t.Fatal("the editor has no Name field")
	}
	screen.Click(screen.PanelCenter(nameField))
	screen.Type("Elf")
	screen.KeyPress(unison.KeyTab, mod.None)
	var name, tabTitle string
	screen.Do(func() {
		name = d.model.Name
		title = d.Title()
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		tabTitle = dockTabTitle(d)
	})
	c.Equal("Elf", name, "typing in the Name field writes the model")
	c.Equal("Ancestry: Elf", title, "the title follows the name until the ancestry has a file")
	c.True(modified, "changing the name is a modification")
	c.True(saveEnabled, "a change enables Save")
	c.Equal("*Ancestry: Elf", tabTitle, "the dock tab marks a modified dockable with an asterisk")

	// Add a gender. The button is below the common options, which may be off the bottom of the view, so scroll it into
	// view first, as a user would.
	var addGender *unison.Button
	screen.Do(func() {
		if addGender = buttonWithTooltip(d.AsPanel(), "Add gender"); addGender != nil {
			addGender.ScrollIntoView()
		}
	})
	if addGender == nil {
		t.Fatal("the editor has no Add gender button")
	}
	screen.Click(screen.PanelCenter(addGender))
	var genders int
	var weight int
	var newGenderNameFocused bool
	screen.Do(func() {
		genders = len(d.model.GenderOptions)
		if genders != 3 {
			return
		}
		added := d.model.GenderOptions[2]
		weight = added.Weight
		newGenderNameFocused = wnd.Focus() == d.targetMgr.Find(added.Value.KeyPrefix+"name")
	})
	if genders != 3 {
		t.Fatalf("expected 3 genders after Add gender, found %d", genders)
	}
	c.Equal(1, weight, "a new gender has a weight of 1")
	c.True(newGenderNameFocused, "the new gender's Name field takes the focus")
	screen.Type("Other")
	screen.KeyPress(unison.KeyTab, mod.None)
	var names []string
	screen.Do(func() { names = genderNames(d.model) })
	c.Equal([]string{"Male", "Female", "Other"}, names, "the typed name is the new gender's")

	// Add a hair option to the common options, give it a value and a weight.
	var addHair *unison.Button
	screen.Do(func() {
		for _, list := range uxtest.PanelsOfType[*weightedStringOptionsPanel](d.AsPanel()) {
			if list.spec.list == &d.model.CommonOptions.HairOptions {
				if addHair = buttonWithSVG(list.AsPanel(), unison.CircledAddSVG); addHair != nil {
					addHair.ScrollIntoView()
				}
				return
			}
		}
	})
	if addHair == nil {
		t.Fatal("the common options have no Add hair option button")
	}
	screen.Click(screen.PanelCenter(addHair))
	var hairOptions int
	var weightField *unison.Panel
	var valueFocused bool
	screen.Do(func() {
		hairOptions = len(d.model.CommonOptions.HairOptions)
		if hairOptions != 1 {
			return
		}
		option := d.model.CommonOptions.HairOptions[0]
		valueFocused = wnd.Focus() == d.targetMgr.Find(option.KeyPrefix+"value")
		weightField = d.targetMgr.Find(option.KeyPrefix + "weight")
	})
	if hairOptions != 1 || weightField == nil {
		t.Fatalf("expected one hair option with a weight field after Add hair option, found %d", hairOptions)
	}
	c.True(valueFocused, "the new option's Value field takes the focus")
	screen.Type("Silver")
	screen.Click(screen.PanelCenter(weightField))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("3")
	screen.KeyPress(unison.KeyTab, mod.None)
	var hair []*gurps.WeightedStringOption
	screen.Do(func() {
		for _, one := range d.model.CommonOptions.HairOptions {
			hair = append(hair, &gurps.WeightedStringOption{Weight: one.Weight, Value: one.Value})
		}
	})
	c.Equal([]*gurps.WeightedStringOption{{Weight: 3, Value: "Silver"}}, hair, "the typed value and weight reach the model")

	// Drag the third gender row's handle to the upper half of the first gender row. The gender list's children are
	// exactly the gender rows, which is what lets the drag be aimed by row index.
	var genderList *unison.Panel
	screen.Do(func() {
		if roots := uxtest.PanelsOfType[*ancestryEditorPanel](d.AsPanel()); len(roots) == 1 {
			genderList = roots[0].genders
		}
	})
	if genderList == nil {
		t.Fatal("the editor has no gender list")
	}
	dragRowAheadOf(t, screen, wnd, genderList, 2, 0)
	var canUndo bool
	screen.Do(func() {
		names = genderNames(d.model)
		canUndo = d.undoMgr.CanUndo()
	})
	c.Equal([]string{"Other", "Male", "Female"}, names, "dropping on the upper half of the first row moves the gender to the front")
	c.True(canUndo, "the drop is undoable")

	// Undo with the keyboard. The key reaches the menu bar's Undo item, whose handler asks the active window for the
	// undo manager of the focused panel, which is one of the editor's fields.
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	screen.Do(func() { names = genderNames(d.model) })
	c.Equal([]string{"Male", "Female", "Other"}, names, "undo restores the original order")

	// Capture the rendering for a person to look at, when one has asked for it.
	uxtest.CaptureScreen(t, c, screen, "ancestry_editor")

	// Save through the toolbar. A new ancestry has no file, so the pure-Go save dialog comes up, offering the user
	// library's ancestries folder and the ancestry's name as the file name, which is accepted as it stands.
	savedPath, loaded, _ := saveNewFileEditor(t, c, screen, wnd, &d.fileEditorDockable, "Ancestry", "Elf", "",
		gurps.AncestryExt, gurps.NewAncestryFromFile)
	c.Equal("Elf", loaded.Name)

	// Save As is always available and always prompts, opening in the directory of the current file with its name
	// offered; canceling leaves the editor exactly as it was.
	var saveAsButton *unison.Button
	screen.Do(func() { saveAsButton = buttonWithSVG(d.AsPanel(), svg.SaveAs) })
	if saveAsButton == nil {
		t.Fatal("the toolbar has no Save As button")
	}
	var saveAsEnabled bool
	screen.Do(func() { saveAsEnabled = saveAsButton.Enabled() })
	c.True(saveAsEnabled, "Save As is available even when the ancestry is unmodified")
	screen.Click(screen.PanelCenter(saveAsButton))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	_, fileName, _ := saveDialogFields(t, screen, dialogWnd)
	c.Equal("Elf", fileName, "Save As offers the current file name")
	screen.KeyPress(unison.KeyEscape, mod.None)
	var path string
	var windows int
	screen.Do(func() {
		windows = len(unison.Windows())
		path = d.path
		modified = d.Modified()
	})
	c.Equal(1, windows, "the Save As dialog has been dismissed")
	c.Equal(savedPath, path, "canceling Save As leaves the path alone")
	c.False(modified, "canceling Save As leaves the ancestry unmodified")
	c.Equal([]string{"Male", "Female", "Other"}, genderNames(loaded), "the file holds the edited genders")
	c.Equal(1, len(loaded.CommonOptions.HairOptions), "the file holds the added hair option")
	if len(loaded.CommonOptions.HairOptions) == 1 {
		c.Equal(3, loaded.CommonOptions.HairOptions[0].Weight)
		c.Equal("Silver", loaded.CommonOptions.HairOptions[0].Value)
	}

	// Change the name, then choose the saved file from the toolbar menu's library list. The editor already shows that
	// file, so the choice brings it forward and changes nothing, not even the edit in progress. The name is at the top
	// of the content, which the steps above scrolled away from, so it is brought back into view first. A second
	// ancestry is put into the library too, so that the menu has one to open in an editor of its own.
	dwarfPath := filepath.Join(user.AncestriesPath(true), "Dwarf"+gurps.AncestryExt)
	c.NoError(os.WriteFile(dwarfPath, []byte(`{"version": 5, "name": "Dwarf"}`), 0o640))
	screen.Do(func() {
		if nameField = d.targetMgr.Find(d.model.KeyPrefix + "name"); nameField != nil {
			nameField.ScrollIntoView()
		}
	})
	if nameField == nil {
		t.Fatal("the editor lost its Name field")
	}
	screen.Click(screen.PanelCenter(nameField))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("Elf2")
	screen.Do(func() {
		name = d.model.Name
		modified = d.Modified()
	})
	c.Equal("Elf2", name)
	c.True(modified, "the renamed ancestry is modified again")
	var menuButton *unison.Button
	screen.Do(func() { menuButton = buttonWithSVG(d.toolbar, svg.Menu) })
	if menuButton == nil {
		t.Fatal("the toolbar has no Menu button")
	}
	screen.Click(screen.PanelCenter(menuButton))
	var items []*unison.Panel
	screen.Do(func() { items = uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd)) })
	// The menu holds Open…, a separator, the user library's title, and its two ancestries in name order: Dwarf, then
	// Elf. The master library has no ancestries, so it is not listed. The items are panels without titles of their own,
	// so the one to click is found by its position.
	if len(items) != 5 {
		t.Fatalf("expected the toolbar menu to hold 5 items (Open…, a separator, the library and its two ancestries), found %d",
			len(items))
	}
	screen.Click(screen.PanelCenter(items[4]))
	var editors int
	screen.Do(func() {
		editors = len(AllMatchingDockables(isAncestryEditor))
		name = d.model.Name
		modified = d.Modified()
		path = d.path
	})
	c.Equal(1, editors, "choosing the file the editor shows opens no second editor")
	c.Equal("Elf2", name, "and keeps the edit in progress")
	c.True(modified)
	c.Equal(savedPath, path)

	// Choosing the other ancestry opens it in an editor of its own, leaving this one as it is.
	screen.Click(screen.PanelCenter(menuButton))
	screen.Do(func() { items = uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd)) })
	if len(items) != 5 {
		t.Fatalf("expected the toolbar menu to hold 5 items again, found %d", len(items))
	}
	screen.Click(screen.PanelCenter(items[3]))
	dwarf := uxtest.OtherEditor(t, screen, d, isAncestryEditor)
	var dwarfName, dwarfPathShown, dwarfTitle string
	var dwarfModified, dwarfCanUndo, dwarfCurrent bool
	screen.Do(func() {
		dwarfName = dwarf.model.Name
		dwarfPathShown = dwarf.path
		dwarfTitle = dwarf.Title()
		dwarfModified = dwarf.Modified()
		dwarfCanUndo = dwarf.undoMgr.CanUndo()
		if dc := unison.Ancestor[*unison.DockContainer](dwarf); dc != nil {
			if cur := dc.CurrentDockable(); cur != nil {
				dwarfCurrent = cur.AsPanel().Self == dwarf
			}
		}
		name = d.model.Name
		modified = d.Modified()
	})
	c.Equal("Dwarf", dwarfName, "the new editor holds the chosen file's ancestry")
	c.Equal(dwarfPath, dwarfPathShown, "and records where it lives")
	c.Equal("Ancestry: Dwarf", dwarfTitle)
	c.False(dwarfModified, "a freshly opened ancestry is unmodified")
	c.False(dwarfCanUndo, "with nothing to undo")
	c.True(dwarfCurrent, "the new editor is brought to the front")
	c.Equal("Elf2", name, "the first editor still holds its edit")
	c.True(modified)
	uxtest.CloseEditorWithoutPrompt(t, screen, dwarf)

	// Close the tab with the unsaved change and discard it. The close runs a modal prompt, so it is posted rather than
	// run through Do, which would wait for it to return.
	screen.Post(func() { d.AttemptClose() })
	screen.Sync()
	_, dialog := modalDialog(t, screen, wnd)
	var discard *unison.Button
	screen.Do(func() { discard = dialog.Button(unison.ModalResponseDiscard) })
	if discard == nil {
		t.Fatal("the save prompt has no discard button")
	}
	screen.Click(screen.PanelCenter(discard))
	screen.Do(func() {
		editors = len(AllMatchingDockables(isAncestryEditor))
		windows = len(unison.Windows())
	})
	c.Equal(0, editors, "discarding closes the editor")
	c.Equal(1, windows, "the prompt has been dismissed")
	c.Equal("Elf", loadSavedFile(t, c, savedPath, gurps.NewAncestryFromFile).Name, "discarding leaves the file as it was saved")
}

// dockTabTitle returns the text of the tab the dock shows for the dockable, or "" if it cannot be found. The tab's
// label is the one within the dockable's dock container that ends with the dockable's title, since a modified
// dockable's tab is prefixed with an asterisk. unison does not export the label's type, so it is told apart by what it
// does: alone among a dock container's panels, it both has a title to set and withholds its context menu
// (unison.ContextMenuWithholder).
func dockTabTitle(d unison.Dockable) string {
	dc := unison.Ancestor[*unison.DockContainer](d)
	if dc == nil {
		return ""
	}
	title := d.Title()
	var text string
	dc.AsPanel().HasInSelfOrDescendants(func(p *unison.Panel) bool {
		label, ok := p.Self.(interface {
			String() string
			SetTitle(string)
			WithholdsContextMenu() bool
		})
		if !ok || !strings.HasSuffix(label.String(), title) {
			return false
		}
		text = label.String()
		return true
	})
	return text
}
