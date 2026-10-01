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
	"errors"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
)

// pendingAppRelease is a release the user has already been told about, versioned far enough ahead that it can never
// match the version the tests are built as.
var pendingAppRelease = []library.Release{{Version: "99.0.0"}}

func TestQuietCheckFailureKeepsAKnownUpdate(t *testing.T) {
	c := check.New(t)
	var u appUpdater
	u.SetReleases(pendingAppRelease)
	wantTitle, _, _ := u.Result()
	c.Contains(wantTitle, "99", "the seeded title must name the release it is about")
	seq, ok := u.beginQuiet()
	c.True(ok)
	u.finishQuiet(seq, nil, errors.New("update site unreachable"))
	title, releases, updating := u.Result()
	c.Equal(wantTitle, title, "a failed quiet check must not change the title")
	c.Equal(1, len(releases), "a failed quiet check must not discard the known release")
	c.Equal("99.0.0", releases[0].Version)
	c.False(updating)
	c.False(u.quiet, "the quiet check must no longer be marked as in flight")
}

// TestQuietCheckWithNoUpdateClearsAStaleRelease covers a release no longer on offer, such as one the user has since
// installed.
func TestQuietCheckWithNoUpdateClearsAStaleRelease(t *testing.T) {
	c := check.New(t)
	var u appUpdater
	u.SetReleases(pendingAppRelease)
	seq, ok := u.beginQuiet()
	c.True(ok)
	u.finishQuiet(seq, nil, nil)
	title, releases, updating := u.Result()
	c.Equal(noAppUpdatesText(), title)
	c.Nil(releases, "a quiet check that found nothing must clear the release")
	c.False(updating)
}

// TestQuietCheckFindingAnUpdateRecordsIt verifies that the quiet path records a release as the visible one does, since
// the toolbar button and the Help menu can't tell which check produced it.
func TestQuietCheckFindingAnUpdateRecordsIt(t *testing.T) {
	c := check.New(t)
	var u appUpdater
	seq, ok := u.beginQuiet()
	c.True(ok)
	u.finishQuiet(seq, pendingAppRelease, nil)
	title, releases, updating := u.Result()
	c.Contains(title, "99", "the title must name the release that was found")
	c.Equal(1, len(releases))
	c.Equal("99.0.0", releases[0].Version)
	c.False(updating)
}

// TestQuietResultIsDiscardedAfterAVisibleCheckStarts verifies that a late quiet result can't overwrite the "Checking…"
// state of a visible check the user started.
func TestQuietResultIsDiscardedAfterAVisibleCheckStarts(t *testing.T) {
	c := check.New(t)
	var u appUpdater
	seq, ok := u.beginQuiet()
	c.True(ok)
	c.True(u.Reset(), "a visible check must be able to start while a quiet one is in flight")
	checkingTitle, _, _ := u.Result()
	u.finishQuiet(seq, pendingAppRelease, nil)
	title, releases, updating := u.Result()
	c.Equal(checkingTitle, title, "the stale quiet result must not replace the visible check's state")
	c.Nil(releases)
	c.True(updating, "the visible check must still be marked as running")
	c.False(u.quiet, "the quiet check must no longer be marked as in flight")
}

func TestBeginQuietRefusesWhileAnotherCheckRuns(t *testing.T) {
	c := check.New(t)
	var visible appUpdater
	c.True(visible.Reset())
	_, ok := visible.beginQuiet()
	c.False(ok, "a quiet check must not start while a visible check is running")

	var quiet appUpdater
	_, ok = quiet.beginQuiet()
	c.True(ok)
	_, ok = quiet.beginQuiet()
	c.False(ok, "a second quiet check must not start while the first is in flight")
}

// TestUncheckedUpdaterReportsAStatus verifies that the title before any result is never blank and follows the setting.
// It once claimed the checks were off whenever none had run, which was wrong once the setting left Never mid-session.
func TestUncheckedUpdaterReportsAStatus(t *testing.T) {
	c := check.New(t)
	option := updatecheck.Never
	u := appUpdater{frequency: func() updatecheck.Option { return option }}
	title, releases, updating := u.Result()
	c.NotEqual("", title, "an updater that has never checked must still report a status")
	c.Contains(title, "off", "with the checks off, the title must say so")
	c.Nil(releases)
	c.False(updating)

	option = updatecheck.Hourly
	title, _, _ = u.Result()
	c.NotEqual("", title)
	c.True(!strings.Contains(title, "off"), "with the checks on, the title must not claim they are off: %q", title)

	seq, ok := u.beginQuiet()
	c.True(ok)
	title, _, _ = u.Result()
	c.Equal(checkingForAppUpdatesText(), title, "while a quiet check runs with nothing known, the title must say so")
	u.finishQuiet(seq, nil, nil)
	title, _, _ = u.Result()
	c.Equal(noAppUpdatesText(), title)
}

func TestQuietCheckFailureWithNothingKnownIsReported(t *testing.T) {
	c := check.New(t)
	u := appUpdater{frequency: func() updatecheck.Option { return updatecheck.Hourly }}
	seq, ok := u.beginQuiet()
	c.True(ok)
	u.finishQuiet(seq, nil, errors.New("update site unreachable"))
	title, releases, updating := u.Result()
	c.Equal(unableToAccessAppUpdateSiteText(), title)
	c.Nil(releases)
	c.False(updating)

	// A later success replaces the failure like any other result.
	seq, ok = u.beginQuiet()
	c.True(ok)
	u.finishQuiet(seq, nil, nil)
	title, _, _ = u.Result()
	c.Equal(noAppUpdatesText(), title)
}

// TestCheckingCoversBothKindsOfCheck verifies that Checking reports both kinds of check while Result reports only a
// visible one as updating: the Help menu's check item reads Checking, while the status item and toolbar button read
// Result and must keep showing a known update during a quiet check.
func TestCheckingCoversBothKindsOfCheck(t *testing.T) {
	c := check.New(t)
	var u appUpdater
	u.SetReleases(pendingAppRelease)
	c.False(u.Checking(), "nothing is in flight to begin with")

	seq, ok := u.beginQuiet()
	c.True(ok)
	c.True(u.Checking(), "a quiet check must count as a check in flight")
	_, releases, updating := u.Result()
	c.False(updating, "a quiet check must not report as a visible one")
	c.Equal(1, len(releases), "a quiet check must leave the known update on display")
	u.finishQuiet(seq, pendingAppRelease, nil)
	c.False(u.Checking(), "a finished quiet check must no longer count")

	c.True(u.Reset())
	c.True(u.Checking(), "a visible check must count as a check in flight")
	_, _, updating = u.Result()
	c.True(updating)
	u.SetResult(noAppUpdatesText())
	c.False(u.Checking(), "a finished visible check must no longer count")
}

func TestShouldShowAppUpdateDialog(t *testing.T) {
	c := check.New(t)
	c.True(shouldShowAppUpdateDialog("99.0.0", "98.0.0"), "a release that hasn't been seen must be announced")
	c.True(shouldShowAppUpdateDialog("99.0.0", ""), "a manual check clears the last seen version to force the dialog")
	c.False(shouldShowAppUpdateDialog("99.0.0", "99.0.0"), "a release already seen must not interrupt the user again")
}
