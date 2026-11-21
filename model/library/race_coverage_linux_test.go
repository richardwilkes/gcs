// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build race

package library

import "testing"

// platformRaceTests re-runs, for TestRace, the tests that exist only on this platform.
func platformRaceTests(t *testing.T) {
	t.Run("LibraryWatchLogsFoldersLeftUnwatchedAtMostOnce", TestLibraryWatchLogsFoldersLeftUnwatchedAtMostOnce)
	t.Run("LibraryWatchPassesOverFoldersThatCannotBeRead", TestLibraryWatchPassesOverFoldersThatCannotBeRead)
	t.Run("WatchFolderEntryLogsALastingFailureAtMostOnce", TestWatchFolderEntryLogsALastingFailureAtMostOnce)
}
