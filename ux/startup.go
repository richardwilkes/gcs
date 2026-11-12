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

// GCSStartupConfig contains configuration for starting the GURPS Character Sheet UI.
type GCSStartupConfig struct {
	// EnableHandoff indicates whether the handoff service should be started. It should be true for the main UI and
	// false for headless sessions.
	EnableHandoff bool
	// UpdateDesktopInfo indicates whether the desktop info (app icons, registry entries, etc) should be updated. It
	// should be true for the main UI and false for headless sessions.
	UpdateDesktopInfo bool
	// ManageAppUpdates indicates whether to settle up after an application update and to check for a new one at launch.
	// It should be true for the main UI and false for headless sessions, which are not the instance that owns the
	// installation and so must not clear away an update another instance may be in the middle of applying.
	ManageAppUpdates bool
}

// DefaultGCSStartupConfig returns a GCSStartupConfig suitable for starting the main GURPS Character Sheet UI.
func DefaultGCSStartupConfig() *GCSStartupConfig {
	return &GCSStartupConfig{
		EnableHandoff:     true,
		UpdateDesktopInfo: true,
		ManageAppUpdates:  true,
	}
}

// DefaultHeadlessGCSStartupConfig returns a GCSStartupConfig suitable for starting a headless GURPS Character Sheet
// session.
func DefaultHeadlessGCSStartupConfig() *GCSStartupConfig {
	return &GCSStartupConfig{
		EnableHandoff:     false,
		UpdateDesktopInfo: false,
		ManageAppUpdates:  false,
	}
}

// StartOptions returns the options to pass to unison.Start for starting the GURPS Character Sheet UI.
func StartOptions(files []string, config *GCSStartupConfig) []unison.StartupOption {
	readyChan := make(chan struct{})
	pathsChan := make(chan []string, 32)
	if config.EnableHandoff {
		startHandoffService(readyChan, pathsChan, files)
	}
	if settings := gurps.GlobalSettings(); settings.General.LibraryUpdateCheck.ChecksAtLaunch() {
		settings.Libraries.PerformUpdateChecks()
	}
	return []unison.StartupOption{
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
			if config.ManageAppUpdates {
				// Settle up after an update applied while this application was not running, before anything else looks
				// at the installation directory. With handoff enabled, only the primary instance gets here, since the
				// handoff service has already decided that by the time this callback fires.
				ReportAppUpdateOutcome()
				if gurps.GlobalSettings().General.AppUpdateCheck.ChecksAtLaunch() {
					CheckForAppUpdates(false)
				}
			}
			// The repeating checks are started here, on the UI thread and after unison is up, since that is where their
			// ticks are delivered and where the state they touch may only be read.
			ApplyUpdateCheckSettings()
			wnd, err := unison.NewWindow(xos.AppName)
			xos.ExitIfErr(err)
			RegisterWindowDragTypes(wnd)
			SetupMenuBar(wnd)
			InitWorkspace(wnd)
			OpenFiles(files)
			if config.EnableHandoff {
				go func() {
					for paths := range pathsChan {
						unison.InvokeTask(func() { OpenFiles(paths) })
					}
				}()
			}
			unison.InvokeTask(func() { performPlatformLateStartup(config.UpdateDesktopInfo) })
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
	}
}

// AppDescription returns a description of the software.
func AppDescription() string {
	return i18n.Text("GURPS Character Sheet is an interactive character sheet editor for the GURPS Fourth Edition roleplaying game.")
}
