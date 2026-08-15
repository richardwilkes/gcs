// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updater

const (
	// workDirPrefix names the staging directory. It is created inside the installation's directory, not in the system
	// temporary directory, so that moving the staged copy into place is a rename within one filesystem rather than a
	// copy across two -- a rename is atomic and instant, a cross-device copy is neither. The leading dot keeps it out
	// of the Finder and out of ordinary directory listings while it exists.
	workDirPrefix = ".gcs-update-"
	// backupSuffix separates the displaced installation's original name from the unique suffix that follows it. See
	// Target.BackupPath for why the resulting name is shaped the way it is.
	backupSuffix = ".old-"
	// helperName is the copy of the running application that finishes the update after that application exits. It is a
	// copy rather than the installed file so that, once the application has quit, nothing under the installation
	// directory is in use by any GCS process -- which is what makes the swap and the cleanup unconditionally safe,
	// notably on Windows, where a running image cannot be deleted. It names a file everywhere except macOS, where a
	// copy of the whole bundle is required and this names that; see stageHelper.
	helperName = "helper"
	// stateName is the record shared between the application that stages an update, the helper that applies it, and
	// the newly installed application that reports and cleans up after it.
	stateName = "update-state.json"
	// logName is the helper's log. The helper has no UI, so this and the application log are the only places its
	// failures can surface.
	logName = "update.log"
)

// sweepGlobs match the leftovers an update can produce in the installation's directory. Anything matching these that is
// old enough is removed at startup; see ReportAndSweep.
var sweepGlobs = []string{workDirPrefix + "*", ".*" + backupSuffix + "*"}
