// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build race

package ux

import "testing"

// TestRace re-runs the tests that put two or more goroutines over shared state, which is all the race detector can
// report on. build.sh runs the full suite uninstrumented and then only this wrapper under the race detector; see
// model/gurps/race_coverage_test.go for the full rationale. List any new concurrency test here as well.
func TestRace(t *testing.T) {
	t.Run("BuildContentCacheReuseAndFailureCaching", TestBuildContentCacheReuseAndFailureCaching)
	t.Run("ContentLoadYieldsToBuildsAndInvalidations", TestContentLoadYieldsToBuildsAndInvalidations)
	t.Run("DeepSearchCachesFileWhoseParsePanics", TestDeepSearchCachesFileWhoseParsePanics)
	t.Run("DeepSearchDisabledTypeIgnoresCache", TestDeepSearchDisabledTypeIgnoresCache)
	t.Run("DeepSearchOfSheetProfileIsCaseInsensitive", TestDeepSearchOfSheetProfileIsCaseInsensitive)
	t.Run("DeepSearchServesCacheHitsUntilInvalidated", TestDeepSearchServesCacheHitsUntilInvalidated)
	t.Run("DockableAwaitingItsWindowIsOpen", TestDockableAwaitingItsWindowIsOpen)
	t.Run("HandoffRefusesWithoutSending", TestHandoffRefusesWithoutSending)
	t.Run("HandoffRoundTrip", TestHandoffRoundTrip)
	t.Run("LibraryFilterAwaitingWindowHearsOfListing", TestLibraryFilterAwaitingWindowHearsOfListing)
	t.Run("LibraryFilterCopyToSheetCarriesLibrarySource", TestLibraryFilterCopyToSheetCarriesLibrarySource)
	t.Run("LibraryFilterDiscardsSupersededLoads", TestLibraryFilterDiscardsSupersededLoads)
	t.Run("LibraryFilterIsReadOnly", TestLibraryFilterIsReadOnly)
	t.Run("LibraryFilterKeepsRowsUntilLoad", TestLibraryFilterKeepsRowsUntilLoad)
	t.Run("LibraryFilterMeasuresRowsReadAgain", TestLibraryFilterMeasuresRowsReadAgain)
	t.Run("LibraryFilterMergesFilesWithFileColumn", TestLibraryFilterMergesFilesWithFileColumn)
	t.Run("LibraryFilterQuickFilterSkipsFileColumn", TestLibraryFilterQuickFilterSkipsFileColumn)
	t.Run("LibraryFilterReloadClosesEditorsOfReplacedRows", TestLibraryFilterReloadClosesEditorsOfReplacedRows)
	t.Run("LibraryFilterReloadsChangedFilesOnly", TestLibraryFilterReloadsChangedFilesOnly)
	t.Run("LibraryFilterReportsUnreadableFiles", TestLibraryFilterReportsUnreadableFiles)
	t.Run("LibraryFilterReusesWhatSupersededLoadsRead", TestLibraryFilterReusesWhatSupersededLoadsRead)
	t.Run("LibraryFilterSelectionStaysWithReIDedRow", TestLibraryFilterSelectionStaysWithReIDedRow)
	t.Run("LibraryFilterShowInSourceFile", TestLibraryFilterShowInSourceFile)
	t.Run("LibraryFilterSizesColumnsToEveryRow", TestLibraryFilterSizesColumnsToEveryRow)
	t.Run("NavigatorSearchOffersMatches", TestNavigatorSearchOffersMatches)
	t.Run("OpenStateChangesReachOtherViews", TestOpenStateChangesReachOtherViews)
	t.Run("PDFQueueRemoval", TestPDFQueueRemoval)
	t.Run("PDFQueueStickyPriority", TestPDFQueueStickyPriority)
	t.Run("PDFQueueUserSignalBeatsContinuations", TestPDFQueueUserSignalBeatsContinuations)
	t.Run("PrewarmWithNoDeepSearchPathsRefreshesSearch", TestPrewarmWithNoDeepSearchPathsRefreshesSearch)
	t.Run("ReloadWalksTheLibrariesOffTheUIThread", TestReloadWalksTheLibrariesOffTheUIThread)
	t.Run("SearchReadsMissingContentInTheBackground", TestSearchReadsMissingContentInTheBackground)
	t.Run("SupersededCacheBuildHandsEntriesToItsSuccessor", TestSupersededCacheBuildHandsEntriesToItsSuccessor)
}
