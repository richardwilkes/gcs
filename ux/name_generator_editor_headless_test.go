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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// fileListPanel is what the pure-Go file dialog's list of files can be seen as from outside unison: the list is a
// unison.List of an unexported item type, so it is found by the methods it has.
type fileListPanel interface {
	unison.Paneler
	Count() int
	RowRect(row int) geom.Rect
}

// TestNameGeneratorEditorHeadless drives the name generator editor end to end inside a headless GCS workspace, the way
// a user would: it opens the editor from the File menu, adds training names by clicking and typing, toggles the case
// options, changes the type through its popup, imports names through the open dialog and undoes that with the keyboard,
// saves through the file dialog, opens the saved generator from an ancestry editor's edit button, and closes both
// editors. The phases build on one another, so a failure in one that the rest cannot proceed without stops the test
// there.
func TestNameGeneratorEditorHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()

	// Open the editor from the File menu. With no training data, the samples can only say why there are none.
	chooseMenuBarItem(t, screen, wnd, "File", "New Name Generator")
	var d *nameGeneratorEditorDockable
	var editors int
	var title, samplesText, expectedMessage string
	var modified, saveEnabled, inWorkspace, samplesInError bool
	screen.Do(func() {
		matches := AllMatchingDockables(isNameGeneratorEditor)
		editors = len(matches)
		if editors != 1 {
			return
		}
		var ok bool
		if d, ok = matches[0].AsPanel().Self.(*nameGeneratorEditorDockable); !ok {
			return
		}
		title = d.Title()
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		inWorkspace = d.Window() == wnd
		samplesText = d.samples.label.text
		samplesInError = d.samples.label.ink == unison.ThemeError
		if _, err := d.model.SampleNames(sampleNameCount); err != nil {
			expectedMessage = errorMessage(err)
		}
	})
	if editors != 1 || d == nil {
		t.Fatalf("expected exactly one name generator editor after choosing New Name Generator, found %d", editors)
	}
	c.Equal("Name Generator: Untitled", title, "a new generator is untitled")
	c.False(modified, "a new generator is unmodified")
	c.False(saveEnabled, "there is nothing to save yet")
	c.True(inWorkspace, "the editor opens in the workspace window")
	c.NotEqual("", expectedMessage, "an empty generator cannot generate names")
	c.Equal(expectedMessage, samplesText, "the samples show why no names can be generated")
	c.True(samplesInError, "the reason is drawn as an error")

	// Choosing the item again opens a second editor, which stands on its own: the first is left as it was, and closing
	// the second, which is untouched, prompts for nothing and leaves the first in place.
	chooseMenuBarItem(t, screen, wnd, "File", "New Name Generator")
	second := otherNameGeneratorEditor(t, screen, d)
	var secondTitle string
	screen.Do(func() { secondTitle = second.Title() })
	c.Equal("Name Generator: Untitled", secondTitle, "the second editor holds a new generator of its own")
	closeEditorWithoutPrompt(t, screen, second)
	var remaining []unison.Dockable
	screen.Do(func() { remaining = AllMatchingDockables(isNameGeneratorEditor) })
	if len(remaining) != 1 || remaining[0].AsPanel().Self != d {
		t.Fatalf("expected the first editor alone to remain after closing the second, found %d", len(remaining))
	}

	// Add two training names by clicking the list's add button and typing into the value field that takes the focus,
	// then give the second a weight of 3. Every structural edit rebuilds the content, so the button is looked up afresh
	// for each click.
	screen.Click(screen.PanelCenter(buttonInView(t, screen, d.AsPanel(), "Add training name")))
	var entries int
	var valueFocused bool
	screen.Do(func() {
		entries = len(d.model.Entries)
		if entries != 1 {
			return
		}
		valueFocused = wnd.Focus() == d.targetMgr.Find(d.model.Entries[0].KeyPrefix+"value")
	})
	if entries != 1 {
		t.Fatalf("expected one training name after Add training name, found %d", entries)
	}
	c.True(valueFocused, "the new training name's Value field takes the focus")
	screen.Type("Alice")
	screen.KeyPress(unison.KeyTab, mod.None)
	screen.Click(screen.PanelCenter(buttonInView(t, screen, d.AsPanel(), "Add training name")))
	var weightField *unison.Panel
	screen.Do(func() {
		entries = len(d.model.Entries)
		if entries != 2 {
			return
		}
		added := d.model.Entries[1]
		valueFocused = wnd.Focus() == d.targetMgr.Find(added.KeyPrefix+"value")
		weightField = d.targetMgr.Find(added.KeyPrefix + "weight")
	})
	if entries != 2 || weightField == nil {
		t.Fatalf("expected two training names with a weight field after Add training name, found %d", entries)
	}
	c.True(valueFocused, "the second training name's Value field takes the focus")
	screen.Type("Bob")
	screen.KeyPress(unison.KeyTab, mod.None)
	screen.Click(screen.PanelCenter(weightField))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("3")
	screen.KeyPress(unison.KeyTab, mod.None)
	typed := []*gurps.WeightedStringOption{{Weight: 1, Value: "Alice"}, {Weight: 3, Value: "Bob"}}
	var names []*gurps.WeightedStringOption
	screen.Do(func() {
		names = plainEntries(d.model.Entries)
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
	})
	c.Equal(typed, names, "the typed names and weight reach the model")
	c.True(modified, "adding training names is a modification")
	c.True(saveEnabled, "a change enables Save")

	// Regenerate the samples. Every one of the ten names must be one of the two training names.
	var randomize *unison.Button
	screen.Do(func() { randomize = buttonWithSVG(d.samples.AsPanel(), svg.Randomize) })
	if randomize == nil {
		t.Fatal("the samples panel has no Randomize button")
	}
	screen.Click(screen.PanelCenter(randomize))
	var samplesNormal bool
	screen.Do(func() {
		samplesText = d.samples.label.text
		samplesNormal = d.samples.label.ink == unison.DefaultLabelTheme.OnBackgroundInk
	})
	samples := strings.Split(samplesText, ", ")
	c.Equal(sampleNameCount, len(samples), "the samples hold ten names: %q", samplesText)
	for _, name := range samples {
		c.True(name == "Alice" || name == "Bob", "every sample is one of the training names, but found %q", name)
	}
	c.True(samplesNormal, "generated names are drawn normally")

	// Turn each case option off and back on through its checkbox. The checkboxes are phrased positively while the model
	// records negations, so unchecking sets the "No" field.
	for _, one := range []struct {
		key   string
		title string
		get   func() bool
	}{
		{key: "lowered", title: "Lowercase", get: func() bool { return d.model.NoLowered }},
		{key: "first_upper", title: "Capitalize", get: func() bool { return d.model.NoFirstToUpper }},
	} {
		var box *unison.Panel
		screen.Do(func() { box = d.targetMgr.Find(d.model.KeyPrefix + one.key) })
		if box == nil {
			t.Fatalf("the editor has no %s checkbox", one.title)
		}
		var negated bool
		screen.Click(screen.PanelCenter(box))
		screen.Do(func() { negated = one.get() })
		c.True(negated, "unchecking %s turns the option off", one.title)
		screen.Click(screen.PanelCenter(box))
		screen.Do(func() { negated = one.get() })
		c.False(negated, "checking %s turns the option back on", one.title)
	}

	// Change the type through its popup. The fields below it follow the type: Markov Letter adds a depth, Compound adds
	// a separator and a list of generators in place of the training data. The generator's key prefix survives the undo
	// clones that replace it, so it can be read once.
	var rootKey string
	screen.Do(func() { rootKey = d.model.KeyPrefix })
	chooseGeneratorType(t, screen, wnd, d, rootKey, namegen.MarkovLetter)
	var genType namegen.Type
	var hasDepth, hasSeparator, hasBuiltIn bool
	screen.Do(func() {
		genType = d.model.Type
		hasDepth = hasWidget(d, rootKey+"depth")
		hasSeparator = hasWidget(d, rootKey+"separator")
	})
	c.Equal(namegen.MarkovLetter, genType, "the popup writes the model")
	c.True(hasDepth, "a Markov letter generator has a Depth field")
	c.False(hasSeparator, "and no Separator")

	chooseGeneratorType(t, screen, wnd, d, rootKey, namegen.Compound)
	var addGenerator *unison.Button
	screen.Do(func() {
		genType = d.model.Type
		hasDepth = hasWidget(d, rootKey+"depth")
		hasSeparator = hasWidget(d, rootKey+"separator")
		hasBuiltIn = hasWidget(d, rootKey+"built_in")
		if addGenerator = buttonWithTooltip(d.AsPanel(), "Add generator"); addGenerator != nil {
			addGenerator.ScrollIntoView()
		}
	})
	c.Equal(namegen.Compound, genType)
	c.True(hasSeparator, "a compound generator has a Separator field")
	c.False(hasDepth, "and no Depth")
	c.False(hasBuiltIn, "and takes no training data of its own")
	if addGenerator == nil {
		t.Fatal("a compound generator has no Generators header with an Add generator button")
	}
	screen.Click(screen.PanelCenter(addGenerator))
	var children, childEntries int
	var childType namegen.Type
	var childTypeFocused, childHasList, childListEmpty bool
	screen.Do(func() {
		children = len(d.model.Compound)
		if children != 1 {
			return
		}
		child := d.model.Compound[0]
		childType = child.Type
		childEntries = len(child.Entries)
		childTypeFocused = wnd.Focus() == d.targetMgr.Find(child.KeyPrefix+"type")
		list := trainingNamesPanel(d, child)
		childHasList = list != nil
		childListEmpty = list != nil && list.rows == nil
	})
	if children != 1 {
		t.Fatalf("expected one generator after Add generator, found %d", children)
	}
	c.Equal(namegen.Simple, childType, "a new generator is simple")
	c.Equal(0, childEntries, "with no training names")
	c.True(childTypeFocused, "the new generator's Type popup takes the focus")
	c.True(childHasList, "the new generator lists its training names")
	c.True(childListEmpty, "of which there are none yet")

	// Switching back to Simple hides the generators, but the model keeps them, as it keeps the training names when a
	// built-in set is chosen, so that switching back and forth loses nothing.
	chooseGeneratorType(t, screen, wnd, d, rootKey, namegen.Simple)
	var compoundRows, rootRows int
	screen.Do(func() {
		genType = d.model.Type
		children = len(d.model.Compound)
		compoundRows = len(panelsOfType[*compoundGeneratorPanel](d.AsPanel()))
		hasBuiltIn = hasWidget(d, rootKey+"built_in")
		if list := trainingNamesPanel(d, d.model); list != nil && list.rows != nil {
			rootRows = len(list.rows.Children())
		}
	})
	c.Equal(namegen.Simple, genType)
	c.Equal(1, children, "the compound children stay in the model when the type changes")
	c.Equal(0, compoundRows, "but are not shown while the type is not compound")
	c.True(hasBuiltIn, "a simple generator takes training data again")
	c.Equal(2, rootRows, "and its training names are intact")

	// Import training names from a text file through the pure-Go open dialog. The dialog opens in the directory last
	// used for files, which is pointed at the directory holding the file so that the file is the one row in its list;
	// startHeadlessWorkspace puts the directory back afterwards. The file repeats a name already in the list and has a
	// blank line, neither of which adds anything.
	importDir := t.TempDir()
	importPath := filepath.Join(importDir, "names.txt")
	if err := os.WriteFile(importPath, []byte("Carol\n\nDave: 4\nAlice\n"), 0o640); err != nil {
		t.Fatalf("unable to write %s: %v", importPath, err)
	}
	var importButton *unison.Button
	screen.Do(func() {
		gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, importDir)
		if importButton = buttonWithSVG(d.AsPanel(), svg.DownToBracket); importButton != nil {
			importButton.ScrollIntoView()
		}
	})
	if importButton == nil {
		t.Fatal("the training names list has no import button")
	}
	screen.Click(screen.PanelCenter(importButton))
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var fileList fileListPanel
	var dialogTitle string
	var rowCount int
	var openEnabled bool
	screen.Do(func() {
		dialogTitle = dialogWnd.Title()
		if lists := panelsOfType[fileListPanel](dialogWnd.Content()); len(lists) == 1 {
			fileList = lists[0]
			rowCount = fileList.Count()
		}
		openEnabled = dialog.Button(unison.ModalResponseOK).Enabled()
	})
	c.Equal("Open…", dialogTitle)
	if fileList == nil || rowCount != 1 {
		t.Fatalf("expected the open dialog to list the one file in %s, found %d rows", importDir, rowCount)
	}
	c.False(openEnabled, "nothing is selected yet, so Open is disabled")
	var rowPoint geom.Point
	screen.Do(func() {
		list := fileList.AsPanel()
		rowPoint = screenPoint(dialogWnd, list.PointToRoot(fileList.RowRect(0).Center()))
	})
	screen.Click(rowPoint)
	var openButton *unison.Button
	screen.Do(func() {
		openButton = dialog.Button(unison.ModalResponseOK)
		openEnabled = openButton.Enabled()
	})
	if !openEnabled {
		t.Fatal("selecting the file must enable Open")
	}
	screen.Click(screen.PanelCenter(openButton))
	imported := []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Alice"},
		{Weight: 3, Value: "Bob"},
		{Weight: 1, Value: "Carol"},
		{Weight: 4, Value: "Dave"},
	}
	var windows int
	var canUndo bool
	screen.Do(func() {
		windows = len(unison.Windows())
		names = plainEntries(d.model.Entries)
		canUndo = d.undoMgr.CanUndo()
	})
	c.Equal(1, windows, "the open dialog has been dismissed")
	c.Equal(imported, names, "the new names are appended with their weights, skipping the duplicate and the blank line")
	c.True(canUndo, "the import is undoable")

	// The import is a single edit: one undo takes every imported name away and one redo brings them all back. The keys
	// reach the menu bar's Undo and Redo items, whose handlers ask the active window for the undo manager of the
	// focused panel, so a field of the editor is clicked first to make sure that is where the focus is.
	var aliceField *unison.Panel
	screen.Do(func() { aliceField = d.targetMgr.Find(d.model.Entries[0].KeyPrefix + "value") })
	if aliceField == nil {
		t.Fatal("the first training name has no Value field")
	}
	screen.Click(screen.PanelCenter(aliceField))
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	screen.Do(func() { names = plainEntries(d.model.Entries) })
	c.Equal(typed, names, "undo removes every imported name at once")
	screen.KeyPress(unison.KeyY, mod.OSMenuCommand())
	screen.Do(func() { names = plainEntries(d.model.Entries) })
	c.Equal(imported, names, "redo restores them all")

	// Drag the last training name's handle to the upper half of the first, which moves it to the front, then undo that
	// with the keyboard too. The list's rows container holds exactly the rows, which is what lets the drag be aimed by
	// row index; it belongs to the current model, which the undo and redo above replaced, so it is looked up now.
	var trainingRows *unison.Panel
	screen.Do(func() {
		if list := trainingNamesPanel(d, d.model); list != nil {
			trainingRows = list.rows
		}
	})
	if trainingRows == nil {
		t.Fatal("the training names have no rows")
	}
	dragRowAheadOf(t, screen, wnd, trainingRows, 3, 0)
	screen.Do(func() {
		names = plainEntries(d.model.Entries)
		canUndo = d.undoMgr.CanUndo()
	})
	c.Equal([]*gurps.WeightedStringOption{
		{Weight: 4, Value: "Dave"},
		{Weight: 1, Value: "Alice"},
		{Weight: 3, Value: "Bob"},
		{Weight: 1, Value: "Carol"},
	}, names, "dropping on the upper half of the first row moves the name to the front")
	c.True(canUndo, "the drop is undoable")
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	screen.Do(func() { names = plainEntries(d.model.Entries) })
	c.Equal(imported, names, "undo restores the order")

	// Select the last two training names -- a click on Carol's row, in the padding to the right of its value field,
	// then a shift-click on Dave's -- and remove them together with the header's button, then undo that with the
	// keyboard too. The rows were rebuilt by the undo above, so they are looked up again.
	var carolPoint, davePoint geom.Point
	var removeSelected *unison.Button
	screen.Do(func() {
		list := trainingNamesPanel(d, d.model)
		if list == nil {
			return
		}
		rowPoint := func(row *unison.Panel) geom.Point {
			rect := row.ContentRect(false)
			return screen.PanelPoint(row, geom.NewPoint(rect.Width+unison.StdHSpacing/2, rect.Height/2))
		}
		rows := list.rows.Children()
		carolPoint = rowPoint(rows[2])
		davePoint = rowPoint(rows[3])
		removeSelected = list.removeSelectedButton
	})
	if removeSelected == nil {
		t.Fatal("the training names have no remove-selected button")
	}
	var removeEnabled bool
	screen.Do(func() { removeEnabled = removeSelected.Enabled() })
	c.False(removeEnabled, "with nothing selected, there is nothing to remove")
	screen.Click(carolPoint)
	screen.ClickWith(davePoint, unison.ButtonLeft, mod.Shift)
	var selectedNames []string
	screen.Do(func() {
		selectedNames = optionValues(trainingNamesPanel(d, d.model).selectedOptions())
		removeEnabled = removeSelected.Enabled()
	})
	c.Equal([]string{"Carol", "Dave"}, selectedNames, "a click and a shift-click select the range between them")
	c.True(removeEnabled, "a selection enables the button")
	screen.Click(screen.PanelCenter(removeSelected))
	screen.Do(func() { names = plainEntries(d.model.Entries) })
	c.Equal(typed, names, "the button removes the selected names")
	screen.Do(func() { aliceField = d.targetMgr.Find(d.model.Entries[0].KeyPrefix + "value") })
	if aliceField == nil {
		t.Fatal("the first training name has no Value field")
	}
	screen.Click(screen.PanelCenter(aliceField))
	screen.KeyPress(unison.KeyZ, mod.OSMenuCommand())
	screen.Do(func() { names = plainEntries(d.model.Entries) })
	c.Equal(imported, names, "one undo brings them both back")

	// Capture the rendering for a person to look at, when one has asked for it; see captureScreen.
	captureScreen(t, c, screen, "name_generator_editor")

	// Save through the toolbar. A new generator has no file, so the pure-Go save dialog comes up, offering the user
	// library's ancestries folder, where ancestries look for generators, and the placeholder name.
	var saveButton *unison.Button
	screen.Do(func() {
		saveButton = d.saveButton
		saveEnabled = saveButton.Enabled()
	})
	c.True(saveEnabled, "the edited generator can be saved")
	screen.Click(screen.PanelCenter(saveButton))
	dialogWnd, _ = modalDialog(t, screen, wnd)
	var fileNameField *unison.Field
	var fileName, dirName string
	screen.Do(func() {
		dialogTitle = dialogWnd.Title()
		if fields := panelsOfType[*unison.Field](dialogWnd.Content()); len(fields) == 1 {
			fileNameField = fields[0]
			fileName = fileNameField.Text()
		}
		// The directory popup is a PopupMenu of an unexported item type, so it is found by the methods it has.
		if popups := panelsOfType[interface {
			Text() string
			ItemCount() int
		}](dialogWnd.Content()); len(popups) != 0 {
			dirName = popups[0].Text()
		}
	})
	c.Equal("Save…", dialogTitle)
	if fileNameField == nil {
		t.Fatal("the save dialog has no file name field")
	}
	c.Equal("Untitled", fileName, "a generator with no file is offered the placeholder name")
	c.Equal(gurps.AncestriesDirName, dirName, "the dialog opens in the user library's ancestries folder")
	screen.Click(screen.PanelCenter(fileNameField))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("Test Names")
	screen.KeyPress(unison.KeyReturn, mod.None)
	savedPath := filepath.Join(user.AncestriesPath(), "Test Names"+gurps.NamesExt)
	var path string
	var hash uint64
	screen.Do(func() {
		windows = len(unison.Windows())
		path = d.path
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		title = d.Title()
		hash = gurps.Hash64(d.model)
	})
	c.Equal(1, windows, "the save dialog has been dismissed")
	c.Equal(savedPath, path, "the editor records where the file was saved")
	c.False(modified, "the saved generator is unmodified")
	c.False(saveEnabled, "saving disables Save")
	c.Equal("Name Generator: Test Names", title, "the title follows the file's base name")
	loaded := loadSavedNameGenerator(t, c, savedPath)
	c.Equal(hash, gurps.Hash64(loaded), "the file holds exactly what the editor holds")
	// The weighted form is a map, read back sorted by name, which for these names is also the order they were in.
	c.Equal(imported, plainEntries(loaded.Entries), "the file holds the training names and their weights")
	data, err := os.ReadFile(savedPath)
	c.NoError(err, "reading %s", savedPath)
	c.Contains(string(data), `"weighted_training_data"`, "a list with a weight other than 1 is written in the weighted form")

	// Open an ancestry editor and add a name generator to its common options. The choices are gathered afresh each time
	// the ancestry editor's content is built, so it offers the generator just saved along with the built-in ones, and
	// the row's edit button opens the chosen one in the name generator editor.
	chooseMenuBarItem(t, screen, wnd, "File", "New Ancestry")
	var anc *ancestryEditorDockable
	var choices []string
	var addNameGenerator *unison.Button
	screen.Do(func() {
		matches := AllMatchingDockables(isAncestryEditor)
		if len(matches) != 1 {
			return
		}
		var ok bool
		if anc, ok = matches[0].AsPanel().Self.(*ancestryEditorDockable); !ok {
			return
		}
		choices = slices.Clone(anc.nameGeneratorChoices)
		for _, p := range panelsOfType[*nameGeneratorsPanel](anc.AsPanel()) {
			if p.options == anc.model.CommonOptions {
				if addNameGenerator = buttonWithSVG(p.AsPanel(), unison.CircledAddSVG); addNameGenerator != nil {
					addNameGenerator.ScrollIntoView()
				}
				return
			}
		}
	})
	if anc == nil {
		t.Fatal("expected exactly one ancestry editor after choosing New Ancestry")
	}
	if addNameGenerator == nil {
		t.Fatal("the common options have no Add name generator button")
	}
	if !slices.Contains(choices, "Test Names") {
		t.Fatalf("the saved generator must be among the choices, but they are %v", choices)
	}
	screen.Click(screen.PanelCenter(addNameGenerator))
	var row *nameGeneratorRefPanel
	var generators []string
	var selected string
	var editEnabled, popupFocused bool
	var testNamesIndex int
	screen.Do(func() {
		generators = slices.Clone(anc.model.CommonOptions.NameGenerators)
		if rows := panelsOfType[*nameGeneratorRefPanel](anc.AsPanel()); len(rows) == 1 {
			row = rows[0]
			selected = selectedGenerator(row.popup)
			editEnabled = row.editButton.Enabled()
			popupFocused = wnd.Focus() == row.popup.AsPanel()
			testNamesIndex = row.popup.IndexOfItem("Test Names")
		}
	})
	if row == nil {
		t.Fatalf("expected one name generator row after Add name generator, but the model holds %v", generators)
	}
	c.Equal([]string{choices[0]}, generators, "the first available generator is the default")
	c.Equal(choices[0], selected, "and the popup shows it")
	c.True(editEnabled, "an available generator can be edited")
	c.True(popupFocused, "the new row's popup takes the focus")
	choosePopupItem(t, screen, wnd, row.popup, testNamesIndex)
	var editButton *unison.Button
	screen.Do(func() {
		generators = slices.Clone(anc.model.CommonOptions.NameGenerators)
		editButton = row.editButton
		editEnabled = editButton.Enabled()
		editButton.ScrollIntoView()
	})
	c.Equal([]string{"Test Names"}, generators, "choosing a generator writes the model")
	c.True(editEnabled, "the saved generator can be edited")
	screen.Click(screen.PanelCenter(editButton))
	var current bool
	screen.Do(func() {
		editors = len(AllMatchingDockables(isNameGeneratorEditor))
		if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
			if cur := dc.CurrentDockable(); cur != nil {
				current = cur.AsPanel().Self == d
			}
		}
		path = d.path
		modified = d.Modified()
		names = plainEntries(d.model.Entries)
	})
	c.Equal(1, editors, "the edit button opens no second editor for a generator already being edited")
	c.True(current, "but brings its editor to the front")
	c.Equal(savedPath, path, "which shows the saved file")
	c.False(modified, "as saved")
	c.Equal(imported, names, "the editor still holds what was saved")

	// Choosing a built-in generator and pressing the edit button opens it in an editor of its own, alongside the one
	// editing the saved generator. A built-in has no file on disk, so the editor knows it by its name. The two editors
	// share a dock container, so bringing the generator's editor forward put the ancestry editor behind it; it is
	// brought back first, as a user clicking its tab would.
	const builtIn = "Human Last"
	var builtInIndex int
	screen.Do(func() {
		ActivateDockable(anc)
		builtInIndex = row.popup.IndexOfItem(builtIn)
	})
	if builtInIndex < 0 {
		t.Fatalf("the built-in generator %q must be among the choices, but they are %v", builtIn, choices)
	}
	choosePopupItem(t, screen, wnd, row.popup, builtInIndex)
	screen.Do(func() {
		editButton = row.editButton
		editEnabled = editButton.Enabled()
		editButton.ScrollIntoView()
	})
	c.True(editEnabled, "a built-in generator can be edited")
	screen.Click(screen.PanelCenter(editButton))
	builtInEditor := otherNameGeneratorEditor(t, screen, d)
	var builtInTitle, builtInPath, builtInName string
	var builtInModified, builtInCanUndo, builtInCurrent bool
	screen.Do(func() {
		builtInTitle = builtInEditor.Title()
		builtInPath = builtInEditor.path
		builtInName = builtInEditor.loadedName
		builtInModified = builtInEditor.Modified()
		builtInCanUndo = builtInEditor.undoMgr.CanUndo()
		if dc := unison.Ancestor[*unison.DockContainer](builtInEditor); dc != nil {
			if cur := dc.CurrentDockable(); cur != nil {
				builtInCurrent = cur.AsPanel().Self == builtInEditor
			}
		}
		names = plainEntries(d.model.Entries)
	})
	c.Equal("Name Generator: "+builtIn, builtInTitle, "the new editor is known by the built-in's name")
	c.Equal("", builtInPath, "a built-in has no path on disk")
	c.Equal(builtIn, builtInName)
	c.False(builtInModified, "a freshly opened generator is unmodified")
	c.False(builtInCanUndo, "with nothing to undo")
	c.True(builtInCurrent, "the new editor is brought to the front")
	c.Equal(imported, names, "the first editor is left as it was")
	closeEditorWithoutPrompt(t, screen, builtInEditor)

	// A generator that no library holds cannot be opened, so its row's edit button is disabled. The name is put into
	// the model and the content rebuilt, as loading an ancestry file that names a missing generator would.
	screen.Do(func() {
		anc.model.CommonOptions.NameGenerators[0] = "Nope"
		anc.sync()
		selected = ""
		editEnabled = true
		if rows := panelsOfType[*nameGeneratorRefPanel](anc.AsPanel()); len(rows) == 1 {
			selected = selectedGenerator(rows[0].popup)
			editEnabled = rows[0].editButton.Enabled()
		}
	})
	c.Equal("Nope", selected, "a missing generator is still shown")
	c.False(editEnabled, "but cannot be edited")

	// Close both editors. The ancestry has unsaved changes, so closing it prompts and the change is discarded; the
	// generator is as saved, so closing it prompts for nothing. The closes run modal prompts, so they are posted rather
	// than run through Do, which would wait for them to return.
	screen.Post(func() { anc.AttemptClose() })
	screen.Sync()
	_, dialog = modalDialog(t, screen, wnd)
	var discard *unison.Button
	screen.Do(func() { discard = dialog.Button(unison.ModalResponseDiscard) })
	if discard == nil {
		t.Fatal("the save prompt has no discard button")
	}
	screen.Click(screen.PanelCenter(discard))
	screen.Post(func() { d.AttemptClose() })
	screen.Sync()
	var ancestryEditors int
	screen.Do(func() {
		ancestryEditors = len(AllMatchingDockables(isAncestryEditor))
		editors = len(AllMatchingDockables(isNameGeneratorEditor))
		windows = len(unison.Windows())
	})
	c.Equal(0, ancestryEditors, "discarding closes the ancestry editor")
	c.Equal(0, editors, "an unmodified generator closes without a prompt")
	c.Equal(1, windows, "no dialog is left open")
	c.Equal(hash, gurps.Hash64(loadSavedNameGenerator(t, c, savedPath)), "closing leaves the file as it was saved")
}

// otherNameGeneratorEditor returns the one name generator editor open besides d, failing the test if there is not
// exactly one.
func otherNameGeneratorEditor(t *testing.T, screen *unison.HeadlessScreen, d *nameGeneratorEditorDockable) *nameGeneratorEditorDockable {
	t.Helper()
	var others []*nameGeneratorEditorDockable
	screen.Do(func() {
		for _, match := range AllMatchingDockables(isNameGeneratorEditor) {
			if other, ok := match.AsPanel().Self.(*nameGeneratorEditorDockable); ok && other != d {
				others = append(others, other)
			}
		}
	})
	if len(others) != 1 {
		t.Fatalf("expected exactly one other name generator editor, found %d", len(others))
	}
	return others[0]
}

// buttonInView returns the button within root whose tooltip reads exactly text, scrolled into view so that it can be
// clicked, failing the test if there is none.
func buttonInView(t *testing.T, screen *unison.HeadlessScreen, root *unison.Panel, text string) *unison.Button {
	t.Helper()
	var button *unison.Button
	screen.Do(func() {
		if button = buttonWithTooltip(root, text); button != nil {
			button.ScrollIntoView()
		}
	})
	if button == nil {
		t.Fatalf("no button with the tooltip %q", text)
	}
	return button
}

// chooseGeneratorType chooses typ in the Type popup of the generator with the given key prefix, the way a user would.
// The popup is rebuilt with the rest of the content whenever the type changes, so it is looked up afresh each time.
func chooseGeneratorType(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, d *nameGeneratorEditorDockable, keyPrefix string, typ namegen.Type) {
	t.Helper()
	var popup *Popup[namegen.Type]
	var index int
	screen.Do(func() {
		if p := d.targetMgr.Find(keyPrefix + "type"); p != nil {
			var ok bool
			if popup, ok = p.Self.(*Popup[namegen.Type]); ok {
				index = popup.IndexOfItem(typ)
			}
		}
	})
	if popup == nil {
		t.Fatalf("no Type popup for the generator with key prefix %q", keyPrefix)
	}
	choosePopupItem(t, screen, wnd, popup, index)
}

// plainEntries returns copies of the options holding only what the file holds, the weight and the value, so that they
// can be compared against literals.
func plainEntries(list []*gurps.WeightedStringOption) []*gurps.WeightedStringOption {
	result := make([]*gurps.WeightedStringOption, 0, len(list))
	for _, one := range list {
		result = append(result, &gurps.WeightedStringOption{Weight: one.Weight, Value: one.Value})
	}
	return result
}

// loadSavedNameGenerator reads the name generator file at path, failing the test if it does not exist or does not
// parse.
func loadSavedNameGenerator(t *testing.T, c check.Checker, path string) *gurps.NameGenerator {
	t.Helper()
	_, err := os.Stat(path)
	c.NoError(err, "the name generator file must exist at %s", path)
	loaded, err := gurps.ReadNameGeneratorFromFS(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("the name generator file at %s must parse: %v", path, err)
	}
	return loaded
}
