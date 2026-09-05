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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestAncestryEditorHeadless drives the ancestry editor end to end inside a headless GCS workspace, the way a user
// would: it opens the editor from the File menu, edits it by clicking and typing, reorders a list by dragging, undoes
// with the keyboard, saves through the file dialog, reloads the file from the toolbar menu, and closes the tab with
// unsaved changes. The phases build on one another, so a failure in one that the rest cannot proceed without stops the
// test there.
func TestAncestryEditorHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()

	// Open the editor from the File menu.
	chooseMenuBarItem(t, screen, wnd, "File", "New Ancestry")
	var d *ancestryEditorDockable
	var editors int
	var title string
	var modified, saveEnabled, inWorkspace bool
	screen.Do(func() {
		matches := AllMatchingDockables(isAncestryEditor)
		editors = len(matches)
		if editors != 1 {
			return
		}
		var ok bool
		if d, ok = matches[0].AsPanel().Self.(*ancestryEditorDockable); !ok {
			return
		}
		title = d.Title()
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		inWorkspace = d.Window() == wnd
	})
	if editors != 1 || d == nil {
		t.Fatalf("expected exactly one ancestry editor after choosing New Ancestry, found %d", editors)
	}
	c.Equal("Ancestry: Untitled", title, "a new ancestry is untitled")
	c.False(modified, "a new ancestry is unmodified")
	c.False(saveEnabled, "there is nothing to save yet")
	c.True(inWorkspace, "the editor opens in the workspace window")

	// Choosing the item again opens a second editor, which stands on its own: the first is left as it was, and closing
	// the second, which is untouched, prompts for nothing and leaves the first in place.
	chooseMenuBarItem(t, screen, wnd, "File", "New Ancestry")
	second := otherAncestryEditor(t, screen, d)
	var secondTitle string
	screen.Do(func() { secondTitle = second.Title() })
	c.Equal("Ancestry: Untitled", secondTitle, "the second editor holds a new ancestry of its own")
	closeEditorWithoutPrompt(t, screen, second)
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
		for _, list := range panelsOfType[*weightedStringOptionsPanel](d.AsPanel()) {
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
		if roots := panelsOfType[*ancestryEditorPanel](d.AsPanel()); len(roots) == 1 {
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

	// Capture the rendering for a person to look at, when one has asked for it; see captureScreen.
	captureScreen(t, c, screen, "ancestry_editor")

	// Save through the toolbar. A new ancestry has no file, so the pure-Go save dialog comes up, offering the user
	// library's ancestries folder and the ancestry's name as the file name.
	var saveButton *unison.Button
	screen.Do(func() {
		saveButton = d.saveButton
		saveEnabled = saveButton.Enabled()
	})
	c.True(saveEnabled, "the edited ancestry can be saved")
	screen.Click(screen.PanelCenter(saveButton))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	var fileNameField *unison.Field
	var fileName, dirName, dialogTitle string
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
	c.Equal("Elf", fileName, "the ancestry's name is offered as the file name")
	c.Equal(gurps.AncestriesDirName, dirName, "the dialog opens in the user library's ancestries folder")
	screen.Click(screen.PanelCenter(fileNameField))
	screen.KeyPress(unison.KeyReturn, mod.None)
	savedPath := filepath.Join(user.AncestriesPath(), "Elf"+gurps.AncestryExt)
	var path, tooltip string
	var hash uint64
	var windows int
	screen.Do(func() {
		windows = len(unison.Windows())
		path = d.path
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		title = d.Title()
		tooltip = d.Tooltip()
		hash = gurps.Hash64(d.model)
	})
	c.Equal(1, windows, "the save dialog has been dismissed")
	c.Equal(savedPath, path, "the editor records where the file was saved")
	c.False(modified, "the saved ancestry is unmodified")
	c.False(saveEnabled, "saving disables Save")
	c.Equal("Ancestry: Elf", title)
	c.Equal(savedPath, tooltip, "the tooltip shows the path")
	loaded := loadSavedAncestry(t, c, savedPath)
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
	dialogWnd, _ = modalDialog(t, screen, wnd)
	screen.Do(func() {
		dialogTitle = dialogWnd.Title()
		fileName = ""
		if fields := panelsOfType[*unison.Field](dialogWnd.Content()); len(fields) == 1 {
			fileName = fields[0].Text()
		}
	})
	c.Equal("Save…", dialogTitle)
	c.Equal("Elf", fileName, "Save As offers the current file name")
	screen.KeyPress(unison.KeyEscape, mod.None)
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
	c.Equal(hash, gurps.Hash64(loaded), "the file holds exactly what the editor holds")

	// Change the name, then choose the saved file from the toolbar menu's library list. The editor already shows that
	// file, so the choice brings it forward and changes nothing, not even the edit in progress. The name is at the top
	// of the content, which the steps above scrolled away from, so it is brought back into view first, as a user would.
	// A second ancestry is put into the library first, so that the menu has one to open in an editor of its own.
	dwarfPath := filepath.Join(user.AncestriesPath(), "Dwarf"+gurps.AncestryExt)
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
	screen.Do(func() { items = menuItemPanels(openMenuPopup(wnd)) })
	// The menu holds Open…, a separator, the user library's title, and its two ancestries in name order: Dwarf, then
	// Elf. The master library has no ancestries, so it is not listed. The items are panels without titles of their own,
	// so the one to click is found by its position.
	if len(items) != 5 {
		t.Fatalf("expected the toolbar menu to hold 5 items (Open…, a separator, the library and its two ancestries), found %d",
			len(items))
	}
	screen.Click(screen.PanelCenter(items[4]))
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
	screen.Do(func() { items = menuItemPanels(openMenuPopup(wnd)) })
	if len(items) != 5 {
		t.Fatalf("expected the toolbar menu to hold 5 items again, found %d", len(items))
	}
	screen.Click(screen.PanelCenter(items[3]))
	dwarf := otherAncestryEditor(t, screen, d)
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
	closeEditorWithoutPrompt(t, screen, dwarf)

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
	c.Equal("Elf", loadSavedAncestry(t, c, savedPath).Name, "discarding leaves the file as it was saved")
}

// otherAncestryEditor returns the one ancestry editor open besides d, failing the test if there is not exactly one.
func otherAncestryEditor(t *testing.T, screen *unison.HeadlessScreen, d *ancestryEditorDockable) *ancestryEditorDockable {
	t.Helper()
	var others []*ancestryEditorDockable
	screen.Do(func() {
		for _, match := range AllMatchingDockables(isAncestryEditor) {
			if other, ok := match.AsPanel().Self.(*ancestryEditorDockable); ok && other != d {
				others = append(others, other)
			}
		}
	})
	if len(others) != 1 {
		t.Fatalf("expected exactly one other ancestry editor, found %d", len(others))
	}
	return others[0]
}

// closeEditorWithoutPrompt closes a dockable that is expected to close without a prompt, and fails the test if it is
// still open or a dialog came up afterwards. The close is posted rather than run through Do, since a prompt, were there
// one, would be modal and Do would wait for it.
func closeEditorWithoutPrompt(t *testing.T, screen *unison.HeadlessScreen, d interface {
	unison.Dockable
	unison.TabCloser
},
) {
	t.Helper()
	screen.Post(func() { d.AttemptClose() })
	screen.Sync()
	var stillOpen bool
	var windows int
	screen.Do(func() {
		stillOpen = slices.ContainsFunc(AllDockables(), func(open unison.Dockable) bool { return open.AsPanel().Self == d })
		windows = len(unison.Windows())
	})
	if stillOpen {
		t.Fatalf("%s is still open", d.Title())
	}
	if windows != 1 {
		t.Fatalf("closing %s left %d windows open; expected the workspace alone", d.Title(), windows)
	}
}

// dockTabTitle returns the text of the tab the dock shows for the dockable, or "" if it cannot be found. The tab's
// label is the one within the dockable's dock container that ends with the dockable's title, since a modified
// dockable's tab is prefixed with an asterisk.
func dockTabTitle(d unison.Dockable) string {
	dc := unison.Ancestor[*unison.DockContainer](d)
	if dc == nil {
		return ""
	}
	title := d.Title()
	for _, label := range panelsOfType[*unison.Label](dc.AsPanel()) {
		if text := label.String(); strings.HasSuffix(text, title) {
			return text
		}
	}
	return ""
}

// loadSavedAncestry reads the ancestry file at path, failing the test if it does not exist or does not parse.
func loadSavedAncestry(t *testing.T, c check.Checker, path string) *gurps.Ancestry {
	t.Helper()
	_, err := os.Stat(path)
	c.NoError(err, "the ancestry file must exist at %s", path)
	loaded, err := gurps.NewAncestryFromFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("the ancestry file at %s must parse: %v", path, err)
	}
	return loaded
}
