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
	_ "embed"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
)

//go:embed images/app-256.png
var appIconBytes []byte

// Start the UI.
func Start(files []string) {
	readyChan := make(chan struct{})
	pathsChan := make(chan []string, 32)
	startHandoffService(readyChan, pathsChan, files)
	if settings := gurps.GlobalSettings(); settings.General.LibraryUpdateCheck.ChecksAtLaunch() {
		settings.Libraries.PerformUpdateChecks()
	}
	unison.Start(
		unison.StartupFinishedCallback(func() {
			unison.DefaultTableColumnHeaderTheme.OnBackgroundInk = colors.OnHeader
			unison.DefaultMarkdownTheme.LinkHandler = HandleLink
			unison.DefaultMarkdownTheme.WorkingDirProvider = WorkingDirProvider
			unison.DefaultMarkdownTheme.AltLinkPrefixes = []string{"md:"}
			if appIcon, err := unison.NewImageFromBytes(appIconBytes, geom.NewPoint(0.5, 0.5)); err != nil {
				errs.Log(err)
			} else {
				unison.DefaultTitleIcons = []*unison.Image{appIcon}
			}
			// Settle up after an update applied while this application was not running, before anything else has a
			// chance to look at the installation directory. This runs only in the primary instance, since the handoff
			// service has already decided that by the time this callback fires.
			ReportAppUpdateOutcome()
			if gurps.GlobalSettings().General.AppUpdateCheck.ChecksAtLaunch() {
				CheckForAppUpdates()
			}
			// The repeating checks are started from here, on the UI thread and after unison is up, since that is where
			// their ticks are delivered and where the state they touch may only be read.
			ApplyUpdateCheckSettings()
			wnd, err := unison.NewWindow(xos.AppName)
			xos.ExitIfErr(err)
			registerWindowDragTypes(wnd)
			SetupMenuBar(wnd)
			InitWorkspace(wnd)
			OpenFiles(files)
			go func() {
				for paths := range pathsChan {
					unison.InvokeTask(func() { OpenFiles(paths) })
				}
			}()
			unison.InvokeTask(performPlatformLateStartup)
			unison.InvokeTask(func() { close(readyChan) })
		}),
		unison.OpenFilesCallback(OpenFiles),
		unison.AllowQuitCallback(func() bool {
			for _, wnd := range unison.Windows() {
				if !wnd.AttemptClose() || wnd.IsValid() {
					return false
				}
			}
			return true
		}),
		// A prepared update is applied from here rather than from wherever the quit was requested, because
		// unison.AttemptQuit does not return when the quit succeeds -- it reaches xos.Exit and the process ends. This
		// callback is reached on every exit that still runs Go code, and does nothing unless an update is waiting.
		unison.QuittingCallback(applyPendingUpdate),
	) // Never returns
}

// AppDescription returns a description of the software.
func AppDescription() string {
	return i18n.Text("GURPS Character Sheet is an interactive character sheet editor for the GURPS Fourth Edition roleplaying game.")
}
