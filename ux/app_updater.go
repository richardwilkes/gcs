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
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/updater"
)

const (
	// appUpdateName is the application's name. It and the two names below are literals rather than the mutable xos
	// globals, which are set at startup: they have to match what the packager baked into the published files, and the
	// helper process has to resolve the same paths and asset names as the application that staged the update.
	appUpdateName = "GCS"
	// appUpdateCmdName is the executable name, matching executable_name in packaging.yml. It is also the base of every
	// release asset name, the name of the Homebrew cask and the name of the repository the releases are published in.
	appUpdateCmdName = "gcs"
	// appUpdateBundleName is the macOS application bundle's name, matching finder_app_name in packaging.yml.
	appUpdateBundleName = "GCS.app"
)

var (
	appUpdaterOnce     sync.Once
	appUpdaterInstance *updater.Updater
	// appUpdateStatusOverride, when set, stands in for what the updater knows. Only tests set it.
	appUpdateStatusOverride func() (title string, releases []updater.Release, updating bool)
)

// appUpdater returns the application's updater, making it on first use.
func appUpdater() *updater.Updater {
	appUpdaterOnce.Do(func() { appUpdaterInstance = updater.New(appUpdaterConfig()) })
	return appUpdaterInstance
}

func appUpdaterConfig() *updater.Config {
	return &updater.Config{
		AppName:         appUpdateName,
		CmdName:         appUpdateCmdName,
		BundleName:      appUpdateBundleName,
		Version:         xos.AppVersion,
		HomebrewCask:    appUpdateCmdName,
		GitHubOwner:     "richardwilkes",
		GitHubRepo:      appUpdateCmdName,
		DownloadPageURL: "https://" + WebSiteDomain,
		HandoffPort:     handoffPort,
		SkipRelease: func(version, _ string) bool {
			// Don't bother showing changes from before 5.0.0, since those were the Java version
			return xstrings.NaturalLess(version, "5.0.0", true)
		},
		FormatVersion:      filterVersion,
		LastSeenVersion:    func() string { return gurps.GlobalSettings().LastSeenGCSVersion },
		SetLastSeenVersion: func(version string) { gurps.GlobalSettings().LastSeenGCSVersion = version },
		ChecksDisabled: func() bool {
			return gurps.GlobalSettings().General.AppUpdateCheck == updatecheck.Never
		},
		StateChanged: SyncAppUpdateButton,
		// Workspace.ErrorHandler is replaced once the workspace exists, so it is looked up when needed.
		ErrorHandler: func(title string, err error) { Workspace.ErrorHandler(title, err) },
		Icon: &unison.DrawableSVG{
			SVG:  svg.Download,
			Size: geom.NewSize(48, 48),
		},
	}
}

// CheckForAppUpdates starts a visible check for application updates: the state shows a check under way, and a release
// not yet shown opens the notification dialog, as does any release found when alwaysNotify is set.
func CheckForAppUpdates(alwaysNotify bool) {
	appUpdater().Check(alwaysNotify)
}

// checkForAppUpdatesQuietly checks for application updates for the repeating schedule. It never opens the dialog, and a
// failed check leaves what is known in place.
func checkForAppUpdatesQuietly() {
	appUpdater().CheckQuietly()
}

// NotifyOfAppUpdate notifies the user of the available update.
func NotifyOfAppUpdate() {
	appUpdater().Notify()
}

// ReportAppUpdateOutcome settles up after an update that was applied while the application was not running, clearing
// away what it left behind and telling the user if anything went wrong.
func ReportAppUpdateOutcome() {
	appUpdater().ReportOutcome()
}

// AppUpdateResult returns the current results of any outstanding app update check.
func AppUpdateResult() (title string, releases []updater.Release, updating bool) {
	if appUpdateStatusOverride != nil {
		return appUpdateStatusOverride()
	}
	return appUpdater().Status()
}

// AppUpdateCheckInProgress returns true while an app update check of either kind, visible or quiet, is in flight.
func AppUpdateCheckInProgress() bool {
	return appUpdater().Checking()
}
