/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ui

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
)

// See https://developer.gnome.org/documentation/guidelines/maintainer/integrating.html

//go:embed app-256.png
var appIcon []byte

func performPlatformStartup() {
	exePath, err := os.Executable()
	if err != nil {
		jot.Error(errs.Wrap(err))
		return
	}
	if filepath.Base(exePath) != cmdline.AppCmdName {
		jot.Warnf("skipping desktop integration since executable name '%s' is not '%s'", filepath.Base(exePath),
			cmdline.AppCmdName)
		return
	}
	if err = installIcons(); err != nil {
		jot.Error(err)
	}
	if err = installDesktopFiles(exePath); err != nil {
		jot.Error(err)
	}
}

func installDesktopFiles(exePath string) error {
	dir := filepath.Join(paths.HomeDir(), ".local", "share", "applications")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	data := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=%s
Exec=%s %%F
Icon=%s
MimeType=%s;
Categories=Game;Utility;RolePlaying;
Keywords=gurps;character;sheet;rpg;roleplaying;utility;
Terminal=false
`, cmdline.AppName, AppDescription, exePath, cmdline.AppIdentifier, strings.Join(library.RegisteredMimeTypes(), ";"))
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".desktop"), []byte(data), 0o640); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func installIcons() error {
	dir := filepath.Join(paths.HomeDir(), ".local", "share", "icons", "hicolor", "256x256", "apps")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".png"), appIcon, 0o640); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
