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
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

const (
	// staleAge is how old a leftover has to be before the startup sweep will remove it on sight. Anything younger
	// might belong to an update being applied right now by a helper this instance knows nothing about.
	staleAge = 7 * 24 * time.Hour
	// abandonedStagingAge is the shorter bound used just before a new update is prepared. By that point any helper
	// still working would have had to survive since before the application started, and the wait it performs gives up
	// long before this; anything older is from an attempt that did not clean up after itself. Without this, each
	// failed attempt would leave most of a hundred megabytes sitting in the user's applications folder for a week.
	abandonedStagingAge = time.Hour
)

// removeAttempts and removeDelay bound the retries when clearing leftovers. On Windows the helper may still be exiting
// as the newly installed application starts, and its image cannot be deleted until it has.
const (
	removeAttempts = 5
	removeDelay    = 500 * time.Millisecond
)

// Outcome reports what became of an update that was in progress when the application last exited. A nil Outcome means
// there was nothing to report.
type Outcome struct {
	Reason    Reason
	Detail    string
	ToVersion string
	Applied   bool
}

// ReportAndSweep settles up after an update: it repairs an installation left half-swapped, reports what happened, and
// clears away what the helper deliberately left behind.
//
// The helper never deletes anything. It moves the old installation aside, puts the new one in place, and starts it.
// Removing the old copy is left to here, which means it only happens once the replacement has demonstrably started --
// so a version that cannot run is still sitting on disk next to one that can, rather than having been deleted by the
// process that installed it.
//
// This must run in the primary instance only, after the handoff service has established that this process is it.
// Failures are logged rather than surfaced: the application has already started successfully by this point, and a
// leftover directory is not worth interrupting the user for.
func ReportAndSweep() *Outcome {
	statePath := StatePath()
	state, err := LoadState(statePath)
	if err != nil {
		// An unreadable or unrecognized state file cannot be acted on, and keeping it would mean reporting the same
		// confusion on every launch from here on.
		errs.Log(errs.NewWithCause("discarding an unusable update state", err), "path", statePath)
		removeIgnoringErrors(statePath)
		sweepStrays(currentSweepDirs(), staleAge)
		return nil
	}
	if state == nil {
		sweepStrays(currentSweepDirs(), staleAge)
		return nil
	}
	outcome := settle(state)
	removeIgnoringErrors(statePath)
	dirs := currentSweepDirs()
	if state.Target != "" {
		dirs = append(dirs, filepath.Dir(state.Target))
	}
	sweepStrays(dirs, staleAge)
	return outcome
}

// settle repairs and interprets a state left behind by a previous run, removing whatever is safe to remove.
func settle(state *State) *Outcome {
	// Repair first. Reaching here means the helper died between moving the installation aside and putting the
	// replacement in its place -- a window of microseconds that its own rollback normally covers, so this is a net
	// rather than an expected path.
	if state.Backup != "" && state.Target != "" && !exists(state.Target) && exists(state.Backup) {
		if err := os.Rename(state.Backup, state.Target); err != nil {
			errs.Log(errs.NewWithCause("unable to restore the previous version", err), "backup", state.Backup,
				"target", state.Target)
		} else {
			slog.Warn("restored the previous version after an interrupted update", "target", state.Target)
		}
		removeWithRetry(state.WorkDir)
		return &Outcome{Reason: ReasonSwapFailed, Detail: state.Detail, ToVersion: state.ToVersion}
	}

	switch state.Status {
	case StatusApplied:
		if state.ToVersion != "" && state.ToVersion != xos.AppVersion {
			// The swap claimed success, yet this is not the version it installed. Something else is going on, so keep
			// the previous version rather than deleting the only other copy on the machine.
			errs.Log(errs.New("the installed version is not the one the update reported installing"),
				"expected", state.ToVersion, "actual", xos.AppVersion, "backup", state.Backup)
			removeWithRetry(state.WorkDir)
			return &Outcome{Reason: ReasonVersionMismatch, ToVersion: state.ToVersion}
		}
		slog.Info("update applied", "version", state.ToVersion, "from", state.FromVersion)
		removeWithRetry(state.Backup)
		removeWithRetry(state.WorkDir)
		if state.Reason != ReasonNone {
			// Installed, but something afterwards went wrong -- the relaunch, most likely, which is why we are here
			// rather than having been started by it.
			return &Outcome{Applied: true, Reason: state.Reason, Detail: state.Detail, ToVersion: state.ToVersion}
		}
		return &Outcome{Applied: true, ToVersion: state.ToVersion}
	case StatusFailed:
		errs.Log(errs.Newf("the update was not applied: %s", state.Reason), "detail", state.Detail)
		removeWithRetry(state.WorkDir)
		// The installation was left untouched by a failure at this point, so any backup is a stray.
		if exists(state.Target) {
			removeWithRetry(state.Backup)
		}
		return &Outcome{Reason: state.Reason, Detail: state.Detail, ToVersion: state.ToVersion}
	case StatusStaged:
		// Staged but never applied: the helper never ran, or never got as far as recording a result. Nothing was
		// touched, so there is nothing to repair -- just clear the staging away.
		slog.Info("discarding an update that was staged but never applied", "version", state.ToVersion)
		removeWithRetry(state.WorkDir)
		return &Outcome{Reason: ReasonAbandoned, ToVersion: state.ToVersion}
	default:
		errs.Log(errs.Newf("unrecognized update status %q", state.Status))
		removeWithRetry(state.WorkDir)
		return nil
	}
}

// currentSweepDirs returns the directories an update to this installation would have left things in.
func currentSweepDirs() []string {
	target, err := CurrentTarget()
	if err != nil {
		errs.Log(errs.NewWithCause("unable to determine the installation directory", err))
		return nil
	}
	return []string{target.Parent}
}

// sweepStrays removes staging directories and displaced installations old enough that no update in progress could still
// need them. This is the backstop for a helper that was killed before it could record anything, and it is what
// eventually clears a Windows image that could not be deleted while it was still running.
func sweepStrays(dirs []string, olderThan time.Duration) {
	seen := make(map[string]bool, len(dirs))
	cutoff := time.Now().Add(-olderThan)
	for _, dir := range dirs {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		for _, glob := range sweepGlobs {
			matches, err := filepath.Glob(filepath.Join(dir, glob))
			if err != nil {
				continue
			}
			for _, match := range matches {
				fi, statErr := os.Lstat(match)
				if statErr != nil || fi.ModTime().After(cutoff) {
					continue
				}
				slog.Info("removing a leftover from an earlier update", "path", match)
				removeWithRetry(match)
			}
		}
	}
}

// removeWithRetry removes a path, retrying briefly. The retries are for Windows, where the helper's own image cannot be
// deleted until it has finished exiting, which may be a moment after it started the replacement.
func removeWithRetry(path string) {
	if path == "" {
		return
	}
	for i := range removeAttempts {
		err := os.RemoveAll(path)
		if err == nil || os.IsNotExist(err) {
			return
		}
		if i < removeAttempts-1 {
			time.Sleep(removeDelay)
		} else {
			errs.Log(errs.NewWithCause("unable to remove a leftover from an update", err), "path", path)
		}
	}
}

func exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Lstat(path)
	return err == nil
}
