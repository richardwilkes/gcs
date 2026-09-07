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
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// The shared open-file dialog opens in the directory recorded under the caller's key, lists only the files with the
// allowed extensions, hands back what was chosen and records the directory for next time, but records nothing when it
// is canceled.
func TestChooseFilesToOpen(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)

	// A directory holding a text file and an image, so an extension filter has something to leave out. The last-used
	// directories are process-wide; startHeadlessWorkspace puts them back afterwards.
	dir := t.TempDir()
	textPath := filepath.Join(dir, "notes.txt")
	imagePath := filepath.Join(dir, "picture.png")
	for _, one := range []string{textPath, imagePath} {
		if err := os.WriteFile(one, []byte("x"), 0o640); err != nil {
			t.Fatalf("unable to write %s: %v", one, err)
		}
	}
	const key = "choose-files-test"
	screen.Do(func() { gurps.GlobalSettings().SetLastDir(key, dir) })

	// Choose the one text file. The dialog runs a nested modal loop, so the call is posted rather than run through Do
	// and its results are read once the dialog has been dismissed.
	var chosen string
	var chosenOK, returned bool
	c.True(screen.Post(func() {
		chosen, chosenOK = chooseFileToOpen(key, "txt")
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	fileList, rowCount := openDialogFileList(t, screen, dialogWnd)
	if rowCount != 1 {
		t.Fatalf("expected the dialog to list only the text file in %s, found %d rows", dir, rowCount)
	}
	screen.Click(fileListRowPoint(screen, dialogWnd, fileList, 0))
	var openButton *unison.Button
	var openEnabled bool
	screen.Do(func() {
		openButton = dialog.Button(unison.ModalResponseOK)
		openEnabled = openButton.Enabled()
	})
	if !openEnabled {
		t.Fatal("selecting the file must enable Open")
	}
	screen.Click(screen.PanelCenter(openButton))
	var windows int
	var lastDir string
	screen.Do(func() {
		windows = len(unison.Windows())
		lastDir = gurps.GlobalSettings().LastDir(key)
	})
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the choice has been made")
	c.True(chosenOK, "a chosen file is reported as chosen")
	c.Equal(textPath, chosen)
	c.Equal(dir, lastDir, "the chosen file's directory is recorded under the key")

	// Choose both files at once when any file is allowed: a click and a shift-click select the range.
	var paths []string
	returned = false
	c.True(screen.Post(func() {
		paths, chosenOK = chooseFilesToOpen(key, true)
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog = modalDialog(t, screen, wnd)
	fileList, rowCount = openDialogFileList(t, screen, dialogWnd)
	if rowCount != 2 {
		t.Fatalf("expected the dialog to list both files in %s, found %d rows", dir, rowCount)
	}
	screen.Click(fileListRowPoint(screen, dialogWnd, fileList, 0))
	screen.ClickWith(fileListRowPoint(screen, dialogWnd, fileList, 1), unison.ButtonLeft, mod.Shift)
	screen.Do(func() {
		openButton = dialog.Button(unison.ModalResponseOK)
		openEnabled = openButton.Enabled()
	})
	if !openEnabled {
		t.Fatal("selecting the files must enable Open")
	}
	screen.Click(screen.PanelCenter(openButton))
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the choice has been made")
	c.True(chosenOK, "chosen files are reported as chosen")
	c.Equal([]string{textPath, imagePath}, paths, "every selected file is returned, in list order")

	// Cancel a dialog opened under a key nothing has been recorded for: the key stays unrecorded.
	const freshKey = "choose-files-test-fresh"
	returned = false
	c.True(screen.Post(func() {
		paths, chosenOK = chooseFilesToOpen(freshKey, false, "txt")
		returned = true
	}))
	screen.Sync()
	modalDialog(t, screen, wnd)
	screen.KeyPress(unison.KeyEscape, mod.None)
	var recorded bool
	screen.Do(func() {
		windows = len(unison.Windows())
		_, recorded = gurps.GlobalSettings().LastDirs[freshKey]
	})
	c.Equal(1, windows, "escape has dismissed the dialog")
	c.True(returned, "the canceled call has returned")
	c.False(chosenOK, "a canceled dialog reports that nothing was chosen")
	c.Nil(paths, "and returns no paths")
	c.False(recorded, "a canceled dialog records no directory")
}

// The shared save-file dialog opens in the directory it is given offering the sanitized initial name, hands back the
// chosen path with the required extension on it and records the directory under the caller's key, but records nothing
// when canceled or when given no key.
func TestChooseFileToSave(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	dir := t.TempDir()
	const key = "choose-file-to-save-test"

	// Accept the offered name as is. The name has a character no file name may hold, which the dialog is offered
	// sanitized, and lacks the extension, which the returned path carries. The dialog runs a nested modal loop, so the
	// call is posted rather than run through Do and its results are read once the dialog has been dismissed.
	var chosen string
	var chosenOK, returned bool
	c.True(screen.Post(func() {
		chosen, chosenOK = chooseFileToSave(dir, "Alpha: Beta", "txt", key)
		returned = true
	}))
	screen.Sync()
	dialogWnd, _ := modalDialog(t, screen, wnd)
	fileNameField, fileName, dirName := saveDialogFields(t, screen, dialogWnd)
	c.Equal("Alpha@6 Beta", fileName, "the initial name is offered sanitized for the file system")
	c.Equal(filepath.Base(dir), dirName, "the dialog opens in the directory it was given")
	screen.Click(screen.PanelCenter(fileNameField))
	screen.KeyPress(unison.KeyReturn, mod.None)
	var windows int
	var lastDir string
	screen.Do(func() {
		windows = len(unison.Windows())
		lastDir = gurps.GlobalSettings().LastDir(key)
	})
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the choice has been made")
	c.True(chosenOK, "a chosen file is reported as chosen")
	c.Equal(filepath.Join(dir, "Alpha@6 Beta.txt"), chosen, "the chosen path carries the required extension")
	c.Equal(dir, lastDir, "the chosen file's directory is recorded under the key")

	// Choose again with no key: the choice is made, but nothing is recorded.
	const unusedKey = "choose-file-to-save-test-unused"
	returned = false
	c.True(screen.Post(func() {
		chosen, chosenOK = chooseFileToSave(dir, "Gamma", "txt", "")
		returned = true
	}))
	screen.Sync()
	dialogWnd, _ = modalDialog(t, screen, wnd)
	fileNameField, _, _ = saveDialogFields(t, screen, dialogWnd)
	screen.Click(screen.PanelCenter(fileNameField))
	screen.KeyPress(unison.KeyReturn, mod.None)
	var lastDirs map[string]string
	screen.Do(func() {
		windows = len(unison.Windows())
		lastDirs = maps.Clone(gurps.GlobalSettings().LastDirs)
	})
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the choice has been made")
	c.True(chosenOK, "a chosen file is reported as chosen")
	c.Equal(filepath.Join(dir, "Gamma.txt"), chosen)
	c.Equal(dir, lastDirs[key], "the earlier key is untouched")
	_, recorded := lastDirs[""]
	c.False(recorded, "an empty key records nothing")

	// Cancel a dialog opened under a key nothing has been recorded for: the key stays unrecorded.
	returned = false
	c.True(screen.Post(func() {
		chosen, chosenOK = chooseFileToSave(dir, "Delta", "txt", unusedKey)
		returned = true
	}))
	screen.Sync()
	modalDialog(t, screen, wnd)
	screen.KeyPress(unison.KeyEscape, mod.None)
	screen.Do(func() {
		windows = len(unison.Windows())
		_, recorded = gurps.GlobalSettings().LastDirs[unusedKey]
	})
	c.Equal(1, windows, "escape has dismissed the dialog")
	c.True(returned, "the canceled call has returned")
	c.False(chosenOK, "a canceled dialog reports that nothing was chosen")
	c.Equal("", chosen, "and returns no path")
	c.False(recorded, "a canceled dialog records no directory")
}

// openDialogFileList returns the file list of the pure-Go open dialog in dialogWnd and the number of rows it holds,
// failing the test if the window is not the open dialog or has no file list.
func openDialogFileList(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) (fileList fileListPanel, rowCount int) {
	t.Helper()
	var title string
	screen.Do(func() {
		title = dialogWnd.Title()
		if lists := panelsOfType[fileListPanel](dialogWnd.Content()); len(lists) == 1 {
			fileList = lists[0]
			rowCount = fileList.Count()
		}
	})
	if title != "Open…" {
		t.Fatalf("expected the open dialog, found one titled %q", title)
	}
	if fileList == nil {
		t.Fatal("the open dialog has no file list")
	}
	return fileList, rowCount
}

// fileListRowPoint returns the screen point at the center of row in fileList, which lives in dialogWnd.
func fileListRowPoint(screen *unison.HeadlessScreen, dialogWnd *unison.Window, fileList fileListPanel, row int) geom.Point {
	var pt geom.Point
	screen.Do(func() {
		list := fileList.AsPanel()
		pt = screenPoint(dialogWnd, list.PointToRoot(fileList.RowRect(row).Center()))
	})
	return pt
}
