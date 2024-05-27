// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
)

func fixupMovedSettingsFileIfNeeded() {
	oldPath := filepath.Join(paths.HomeDir(), "AppData", "com.trollworks.gcs", "gcs_prefs.json")
	if fs.FileExists(oldPath) && !fs.FileExists(SettingsPath) {
		if err := os.MkdirAll(filepath.Dir(SettingsPath), 0o755); err != nil {
			errs.Log(err)
		}
		if err := fs.MoveFile(oldPath, SettingsPath); err != nil {
			errs.Log(errs.NewWithCause("unable to move settings from old location to new location", err), "old", oldPath, "new", SettingsPath)
		}
	}
}
