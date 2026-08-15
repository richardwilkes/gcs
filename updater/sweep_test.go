// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// installation builds a fake install directory holding a target, a backup and a staging directory, and returns the
// state describing them. Which of the three actually exist is up to the caller, so that each way an update can be
// interrupted can be set up precisely.
func installation(t *testing.T, status Status, present ...string) (state *State, dir string) {
	t.Helper()
	dir = t.TempDir()
	state = &State{
		Schema:    StateSchema,
		Status:    status,
		Target:    filepath.Join(dir, "gcs"),
		Backup:    filepath.Join(dir, "."+CmdName+backupSuffix+"123"),
		WorkDir:   filepath.Join(dir, workDirPrefix+"abc"),
		ToVersion: xos.AppVersion,
	}
	for _, which := range present {
		switch which {
		case "target":
			write(t, state.Target, "new")
		case "backup":
			write(t, state.Backup, "old")
		case "workdir":
			if err := os.MkdirAll(state.WorkDir, 0o755); err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(state.WorkDir, "helper"), "helper")
		}
	}
	return state, dir
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // Test-controlled path
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestSettleRepairsAnInterruptedSwap is the case that matters most: the helper moved the installation aside and died
// before putting the replacement in its place, so the user has no application at all. The previous version must come
// back rather than the user being left with nothing.
func TestSettleRepairsAnInterruptedSwap(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusStaged, "backup", "workdir")

	outcome := settle(state)

	c.True(exists(state.Target), "the previous version must be restored")
	c.Equal("old", read(t, state.Target))
	c.False(exists(state.WorkDir), "the staging directory must be cleared")
	c.NotNil(outcome)
	c.Equal(ReasonSwapFailed, outcome.Reason)
	c.False(outcome.Applied)
}

// TestSettleRemovesTheBackupOnlyAfterASuccessfulStart verifies the rollback window. The helper deliberately leaves the
// previous version on disk; it is removed here, which is the first moment the replacement has been shown to run.
func TestSettleRemovesTheBackupOnlyAfterASuccessfulStart(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusApplied, "target", "backup", "workdir")

	outcome := settle(state)

	c.True(exists(state.Target))
	c.False(exists(state.Backup), "the previous version is removed once the new one has started")
	c.False(exists(state.WorkDir))
	c.NotNil(outcome)
	c.True(outcome.Applied)
	c.Equal(ReasonNone, outcome.Reason)
}

// TestSettleKeepsTheBackupOnAVersionMismatch verifies that when the swap claims to have installed a version other than
// the one now running, the previous version is kept. Deleting it would leave only a copy that is already known not to
// be what it claimed.
func TestSettleKeepsTheBackupOnAVersionMismatch(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusApplied, "target", "backup", "workdir")
	state.ToVersion = "99.99.99"

	outcome := settle(state)

	c.True(exists(state.Backup), "the previous version must be kept when the result cannot be trusted")
	c.False(exists(state.WorkDir))
	c.NotNil(outcome)
	c.Equal(ReasonVersionMismatch, outcome.Reason)
	c.False(outcome.Applied)
}

// TestSettleReportsAnInstalledButUnlaunchedUpdate verifies that a relaunch failure is still reported as applied. The
// new version is in place and running -- that is how we got here -- so it is a notice, not a failure.
func TestSettleReportsAnInstalledButUnlaunchedUpdate(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusApplied, "target", "backup", "workdir")
	state.Reason = ReasonRelaunchFailed

	outcome := settle(state)

	c.NotNil(outcome)
	c.True(outcome.Applied)
	c.Equal(ReasonRelaunchFailed, outcome.Reason)
	c.False(exists(state.Backup))
}

// TestSettleOnAFailedUpdateLeavesTheInstallationAlone verifies that a failure recorded before anything was touched
// clears the staging and leaves the installed application exactly as it was.
func TestSettleOnAFailedUpdateLeavesTheInstallationAlone(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusFailed, "target", "workdir")
	state.Reason = ReasonPredecessorRunning

	outcome := settle(state)

	c.True(exists(state.Target))
	c.Equal("new", read(t, state.Target), "the installed application must be untouched")
	c.False(exists(state.WorkDir))
	c.NotNil(outcome)
	c.Equal(ReasonPredecessorRunning, outcome.Reason)
	c.False(outcome.Applied)
}

// TestSettleDiscardsAnAbandonedStaging verifies that an update that was prepared but never applied -- the user vetoed
// the quit, or the helper never ran -- is cleaned up without touching the installation.
func TestSettleDiscardsAnAbandonedStaging(t *testing.T) {
	c := check.New(t)
	state, _ := installation(t, StatusStaged, "target", "workdir")

	outcome := settle(state)

	c.True(exists(state.Target))
	c.Equal("new", read(t, state.Target))
	c.False(exists(state.WorkDir))
	c.NotNil(outcome)
	c.Equal(ReasonAbandoned, outcome.Reason)
}

// TestSweepStraysRemovesOnlyOldLeftovers verifies both halves of the sweep: it clears what an interrupted update leaves
// behind, and it leaves alone anything recent enough to belong to an update being applied right now by a helper this
// process knows nothing about. It must also not touch the user's own files.
func TestSweepStraysRemovesOnlyOldLeftovers(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	old := time.Now().Add(-staleAge - time.Hour)

	stale := []string{
		filepath.Join(dir, workDirPrefix+"aaa"),
		filepath.Join(dir, "."+CmdName+backupSuffix+"111"),
		filepath.Join(dir, "."+BundleName+backupSuffix+"222"),
	}
	for _, path := range stale {
		write(t, path, "leftover")
		c.NoError(os.Chtimes(path, old, old))
	}

	recent := filepath.Join(dir, workDirPrefix+"bbb")
	write(t, recent, "in progress")

	untouched := []string{
		filepath.Join(dir, "gcs"),
		filepath.Join(dir, "GCS.app"),
		filepath.Join(dir, "gcs_prefs.json"),
		filepath.Join(dir, "my-notes.old-backup"),
	}
	for _, path := range untouched {
		write(t, path, "keep")
		c.NoError(os.Chtimes(path, old, old))
	}

	sweepStrays([]string{dir}, staleAge)

	for _, path := range stale {
		c.False(exists(path), "a stale leftover survived: %s", path)
	}
	c.True(exists(recent), "a recent staging directory must be left for the update still using it")
	for _, path := range untouched {
		c.True(exists(path), "the sweep removed something it did not create: %s", path)
	}
}

// TestSweepStraysToleratesMissingDirectories verifies the sweep is harmless when the installation directory cannot be
// determined or no longer exists. It runs on every launch, so it must never be able to fail the startup path.
func TestSweepStraysToleratesMissingDirectories(t *testing.T) {
	sweepStrays(nil, staleAge)
	sweepStrays([]string{""}, staleAge)
	sweepStrays([]string{filepath.Join(t.TempDir(), "gone")}, staleAge)
}

// TestSweepStraysClearsAbandonedStagingSooner verifies the shorter bound used just before a new update is prepared.
// Each staging directory holds a whole copy of the application, so leaving a few failed attempts to sit for a week
// would quietly cost the user hundreds of megabytes.
func TestSweepStraysClearsAbandonedStagingSooner(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()

	abandoned := filepath.Join(dir, workDirPrefix+"old")
	write(t, abandoned, "an attempt that never finished")
	stamp := time.Now().Add(-abandonedStagingAge - time.Minute)
	c.NoError(os.Chtimes(abandoned, stamp, stamp))

	inFlight := filepath.Join(dir, workDirPrefix+"new")
	write(t, inFlight, "an attempt still running")

	sweepStrays([]string{dir}, abandonedStagingAge)

	c.False(exists(abandoned), "an abandoned staging directory must be cleared before a new attempt")
	c.True(exists(inFlight), "a staging directory a helper could still be using must be left alone")

	// The same directory is left alone by the startup sweep, which uses the longer bound.
	write(t, abandoned, "again")
	c.NoError(os.Chtimes(abandoned, stamp, stamp))
	sweepStrays([]string{dir}, staleAge)
	c.True(exists(abandoned), "the startup sweep must not remove something this recent")
}
