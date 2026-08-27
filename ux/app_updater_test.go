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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
)

// pendingAppRelease is the release the tests seed as one the user has already been told about. Its version is far
// enough ahead of anything real that it can never coincide with the version the tests are built as.
var pendingAppRelease = []gurps.Release{{Version: "99.0.0"}}

// TestQuietCheckFailureKeepsAKnownUpdate verifies that a background check which can't reach the update site leaves the
// update the user already knows about on display. The quiet checks run unattended, so a network hiccup that quietly
// erased the notice would take the update away with nothing to tell the user it had happened.
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

// TestQuietCheckWithNoUpdateClearsAStaleRelease verifies that a background check which finds nothing on offer takes
// down a release that is no longer being offered -- the user having installed it, for instance -- and says so.
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

// TestQuietCheckFindingAnUpdateRecordsIt verifies that the quiet path records a release the same way the visible one
// does, since the toolbar button and the Help menu read that state without knowing which check produced it.
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

// TestQuietResultIsDiscardedAfterAVisibleCheckStarts verifies that a background result which arrives after the user
// started a check of their own is thrown away. The visible check is the newer question, and letting the older answer
// land would overwrite the "Checking…" state with a result the user never asked for.
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

// TestBeginQuietRefusesWhileAnotherCheckRuns verifies that a quiet check never starts on top of a visible one, whose
// answer is the one the user is waiting for, nor on top of another quiet one, which would mean two requests to the
// update site for the same information.
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

// TestUncheckedUpdaterReportsAStatus verifies that an updater which has never recorded a result still has something to
// say, and that what it says follows the setting. The Help menu shows this title verbatim, so an empty string would
// leave a blank menu item, and it used to claim that the checks were off whenever no check had run -- which, after the
// setting was switched away from Never mid-session, was no longer true.
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

// TestQuietCheckFailureWithNothingKnownIsReported verifies that a background check which fails when nothing is known
// records the failure. There is nothing to protect in that state, and leaving the title claiming that no check has
// run would hide the fact that one did and couldn't reach the update site.
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

// TestShouldShowAppUpdateDialog verifies the rule that keeps the update dialog from reappearing at every launch: it
// opens for a release that hasn't been seen and stays shut for the one recorded as already seen.
func TestShouldShowAppUpdateDialog(t *testing.T) {
	c := check.New(t)
	c.True(shouldShowAppUpdateDialog("99.0.0", "98.0.0"), "a release that hasn't been seen must be announced")
	c.True(shouldShowAppUpdateDialog("99.0.0", ""), "a manual check clears the last seen version to force the dialog")
	c.False(shouldShowAppUpdateDialog("99.0.0", "99.0.0"), "a release already seen must not interrupt the user again")
}
