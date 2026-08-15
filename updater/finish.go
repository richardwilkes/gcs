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

	"github.com/richardwilkes/toolbox/v2/errs"
)

// SpawnFinisher starts the helper that applies a staged update once this process has exited.
//
// It is called from the quitting path, so it must do nothing but start the child: this process is moments from
// terminating, and anything slow here delays the user's quit for no reason.
func SpawnFinisher(state *State, statePath string) error {
	if state == nil || state.Helper == "" {
		return errs.New("there is no prepared update to apply")
	}
	slog.Info("starting the update helper", "helper", state.Helper, "target", state.Target)
	return spawnDetached(state.Helper, []string{"--" + FinishFlag, statePath}, state.LogPath, os.TempDir())
}

// FinishFlag is the command line flag that puts a copy of the application into helper mode. It is not meant for anyone
// to type; it exists because the helper is the same program, started again with something to do other than open a
// window.
const FinishFlag = "finish-update"

// Finish applies a staged update. This runs in the helper process, after the application that staged the update has
// exited, and it is the only part of an update that touches the installation.
//
// There is no user interface here and no one to ask. Everything it learns is written back into the state file, which
// the newly installed application reads on its next start; that is how a failure at this point ever reaches the user.
//
// It deletes nothing. The previous version stays on disk until the replacement has demonstrably started, so a version
// that turns out not to run is still sitting beside one that does, rather than having been removed by the very process
// that installed it.
func Finish(statePath string) error {
	state, err := LoadState(statePath)
	if err != nil {
		return err
	}
	if state == nil {
		return errs.New("there is no prepared update to apply")
	}
	if state.Status != StatusStaged {
		return errs.Newf("the prepared update is %s rather than staged", state.Status)
	}

	if err = waitForPredecessor(state.ParentPID, state.HandoffPort); err != nil {
		// Nothing has been touched. Record why and stop -- the installation is exactly as the user left it.
		return abandon(state, statePath, ReasonPredecessorRunning, err)
	}

	target := state.TargetInfo()
	if err = swap(state.Target, state.Payload, state.Backup); err != nil {
		errs.Log(err)
		if relaunchErr := relaunch(&target, state.LogPath); relaunchErr != nil {
			slog.Error("unable to restart after an update that could not be applied", "error", relaunchErr)
		}
		return abandon(state, statePath, ReasonSwapFailed, err)
	}

	state.Status = StatusApplied
	state.Reason = ReasonNone
	state.Detail = ""
	if err = state.Save(statePath); err != nil {
		// The update is in place; only the record of it is not. Carry on and start the new version, which will simply
		// have nothing to report.
		slog.Error("unable to record that the update was applied", "error", err)
	}

	if err = relaunch(&target, state.LogPath); err != nil {
		// The new version is installed and verified; it just did not start. Rolling back here would replace something
		// known to work with something older for no gain, so record it and leave the user to open it themselves.
		slog.Error("unable to start the updated application", "error", err)
		state.Reason = ReasonRelaunchFailed
		state.Detail = err.Error()
		if saveErr := state.Save(statePath); saveErr != nil {
			slog.Error("unable to record that the restart failed", "error", saveErr)
		}
		return err
	}
	slog.Info("update applied and restarted", "version", state.ToVersion)
	return nil
}

// abandon records why an update was not applied and returns the original failure. The state file is what carries this
// to the next run of the application, since there is nobody to tell right now.
func abandon(state *State, statePath string, reason Reason, cause error) error {
	state.Status = StatusFailed
	state.Reason = reason
	state.Detail = cause.Error()
	if err := state.Save(statePath); err != nil {
		slog.Error("unable to record why the update was not applied", "error", err)
	}
	return cause
}
