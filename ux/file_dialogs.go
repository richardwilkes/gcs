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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// chooseFilesToOpen runs the open-file dialog and returns the files the user chose, or false if the dialog was
// canceled. The dialog opens in the directory last used under lastDirKey (see gurps.Settings.LastDir) and, once a
// choice is made, the directory holding the first chosen file is recorded there for next time. Only files may be
// chosen, and only those with one of the given extensions; pass none to allow any file.
func chooseFilesToOpen(lastDirKey string, multiple bool, extensions ...string) ([]string, bool) {
	dialog := unison.NewOpenDialog()
	dialog.SetAllowsMultipleSelection(multiple)
	dialog.SetResolvesAliases(true)
	dialog.SetAllowedExtensions(extensions...)
	dialog.SetCanChooseDirectories(false)
	dialog.SetCanChooseFiles(true)
	global := gurps.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(lastDirKey))
	if !dialog.RunModal() {
		return nil, false
	}
	paths := dialog.Paths()
	if len(paths) == 0 {
		return nil, false
	}
	global.SetLastDir(lastDirKey, filepath.Dir(paths[0]))
	return paths, true
}

// chooseFileToOpen is chooseFilesToOpen for a single file: it returns the file the user chose, or false if the dialog
// was canceled.
func chooseFileToOpen(lastDirKey string, extensions ...string) (string, bool) {
	paths, ok := chooseFilesToOpen(lastDirKey, false, extensions...)
	if !ok {
		return "", false
	}
	return paths[0], true
}
