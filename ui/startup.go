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

	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/menus"
	"github.com/richardwilkes/gcs/v5/ui/updates"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

// AppDescription of the software
var AppDescription = i18n.Text("GURPS Character Sheet is an interactive character sheet editor for the GURPS Fourth Edition roleplaying game.")

// AppIconBytes holds the GCS application icon in a 256x256 format
//
//go:embed app-256.png
var AppIconBytes []byte

//go:embed doc-256.png
var docIconBytes []byte

// Start the UI.
func Start(files []string) {
	libs := settings.Global().LibrarySet
	go libs.PerformUpdateChecks()
	unison.Start(
		unison.StartupFinishedCallback(func() {
			performPlatformStartup()
			if appIcon, err := unison.NewImageFromBytes(AppIconBytes, 0.5); err != nil {
				jot.Error(err)
			} else {
				unison.DefaultTitleIcons = []*unison.Image{appIcon}
			}
			updates.CheckForAppUpdates()
			wnd, err := unison.NewWindow(cmdline.AppName)
			jot.FatalIfErr(err)
			menus.Setup(wnd)
			workspace.NewWorkspace(wnd)
			workspace.OpenFiles(files)
		}),
		unison.OpenFilesCallback(workspace.OpenFiles),
		unison.AllowQuitCallback(func() bool {
			for _, wnd := range unison.Windows() {
				wnd.AttemptClose()
				if wnd.IsValid() {
					return false
				}
			}
			return true
		}),
	) // Never returns
}
