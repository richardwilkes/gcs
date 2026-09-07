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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestPromptForFileSystemNameGatesOKOnTarget drives the name prompt the rename and new-folder commands share inside a
// headless workspace: OK is disabled while the entry holds an invalid name or one whose target already exists and
// enabled once the target is free, the path handed back is the target of the trimmed name, so that the path validated
// is the one the caller acts on, a current name is shown in a disabled row of its own, and canceling returns nothing.
func TestPromptForFileSystemNameGatesOKOnTarget(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "taken"), 0o750))
	targetPath := func(name string) string { return newFolderPath(dir, name) }

	// Rename "taken" to something else. The dialog runs a nested modal loop, so the call is posted rather than run
	// through Do and its results are read once the dialog has been dismissed.
	var path string
	var ok, returned bool
	c.True(screen.Post(func() {
		path, ok = promptForFileSystemName("taken", "New Name", "taken", targetPath)
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var fields []*StringField
	var okButton *unison.Button
	screen.Do(func() {
		fields = panelsOfType[*StringField](dialogWnd.Content())
		okButton = dialog.Button(unison.ModalResponseOK)
	})
	if len(fields) != 2 {
		t.Fatalf("expected a current name field and an entry, found %d fields", len(fields))
	}
	current, entry := fields[0], fields[1]
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}
	screen.Do(func() {
		c.False(current.Enabled(), "the current name is shown, not edited")
		c.Equal("taken", current.Text())
		c.Equal("taken", entry.Text(), "the entry starts out holding the current name")
	})
	c.False(okEnabled(), "the current name's target exists, so OK starts out disabled")

	screen.Click(screen.PanelCenter(entry))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("  ")
	c.False(okEnabled(), "a blank name is invalid")
	screen.Type("fresh  ")
	c.True(okEnabled(), "a valid name whose target is free enables OK")
	screen.Click(screen.PanelCenter(okButton))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the prompt has returned")
	c.True(ok, "an accepted name is reported as accepted")
	c.Equal(filepath.Join(dir, "fresh"), path, "the path handed back is the trimmed name's target")

	// A prompt with no current name has only the entry, and canceling it hands back nothing.
	returned = false
	c.True(screen.Post(func() {
		path, ok = promptForFileSystemName("", "Folder Name", "", targetPath)
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog = modalDialog(t, screen, wnd)
	screen.Do(func() {
		fields = panelsOfType[*StringField](dialogWnd.Content())
		okButton = dialog.Button(unison.ModalResponseOK)
	})
	c.Equal(1, len(fields), "with no current name, only the entry is shown")
	c.False(okEnabled(), "an empty entry leaves OK disabled")
	screen.Type("new folder")
	c.True(okEnabled())
	screen.KeyPress(unison.KeyEscape, mod.None)
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the prompt has returned")
	c.False(ok, "a canceled prompt is reported as such")
	c.Equal("", path)
	_, err := os.Stat(filepath.Join(dir, "new folder"))
	c.True(os.IsNotExist(err), "the prompt itself creates nothing")
}
