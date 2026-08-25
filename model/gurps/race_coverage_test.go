// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build race

package gurps

import "testing"

// TestRace re-runs the tests that put two or more goroutines over shared state, which is the only situation the race
// detector can ever report on: a test that never leaves the test goroutine cannot produce a report, so instrumenting it
// is pure overhead. build.sh exploits that by running the full suite uninstrumented and then running only this wrapper
// under the race detector (go test -race -run '^TestRace$' ./...). The build tag keeps the wrapper out of plain builds,
// where the tests named here already run directly.
//
// When adding a test that involves concurrency — spawning goroutines itself, starting a library watch, or exercising a
// script timeout — list it here as well so the race pass covers it.
func TestRace(t *testing.T) {
	t.Run("LibraryConcurrentAccess", TestLibraryConcurrentAccess)
	t.Run("LibrarySetPathWhenWatchFails", TestLibrarySetPathWhenWatchFails)
	t.Run("LibrarySetPathAfterWatchFailsRestartsWatch", TestLibrarySetPathAfterWatchFailsRestartsWatch)
	t.Run("LibraryWatchDeliversNothingUntilSomethingHappens", TestLibraryWatchDeliversNothingUntilSomethingHappens)
	t.Run("LibrarySetPathSyncsEachWatcherOnce", TestLibrarySetPathSyncsEachWatcherOnce)
	t.Run("NotifyOfLibraryChangeConcurrent", TestNotifyOfLibraryChangeConcurrent)
	t.Run("ScriptResolutionConcurrency", TestScriptResolutionConcurrency)
	t.Run("ScriptObjectResultConcurrency", TestScriptObjectResultConcurrency)
	t.Run("ScriptResultConversionHonorsTimeout", TestScriptResultConversionHonorsTimeout)
	t.Run("ScriptTimeoutDoesNotAffectNextRun", TestScriptTimeoutDoesNotAffectNextRun)
	t.Run("ScriptExecTimeLimitOverride", TestScriptExecTimeLimitOverride)
}
