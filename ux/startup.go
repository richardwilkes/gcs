// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

//go:embed images/app-256.png
var appIconBytes []byte

// Start the UI.
func Start(files []string) {
	readyChan := make(chan struct{})
	pathsChan := make(chan []string, 32)
	startHandoffService(readyChan, pathsChan, files)
	libs := gurps.GlobalSettings().LibrarySet
	go libs.PerformUpdateChecks()
	unison.Start(
		unison.StartupFinishedCallback(func() {
			unison.DefaultTableColumnHeaderTheme.OnBackgroundInk = colors.OnHeader
			unison.DefaultMarkdownTheme.LinkHandler = HandleLink
			unison.DefaultMarkdownTheme.WorkingDirProvider = WorkingDirProvider
			unison.DefaultMarkdownTheme.AltLinkPrefixes = []string{"md:"}
			if appIcon, err := unison.NewImageFromBytes(appIconBytes, 0.5); err != nil {
				errs.Log(err)
			} else {
				unison.DefaultTitleIcons = []*unison.Image{appIcon}
			}
			CheckForAppUpdates()
			wnd, err := unison.NewWindow(cmdline.AppName)
			fatal.IfErr(err)
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
	) // Never returns
}

// AppDescription returns a description of the software.
func AppDescription() string {
	return i18n.Text("GURPS Character Sheet is an interactive character sheet editor for the GURPS Fourth Edition roleplaying game.")
}
