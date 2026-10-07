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
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/updater"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// stageTimeout bounds the whole prepare step. It is generous because the download is tens of megabytes and some
// connections are slow; the Cancel button, not this, is how an impatient user stops it.
const stageTimeout = 30 * time.Minute

// progressResolution is how finely the progress bar is subdivided. Download progress is quantized to it so that only
// visible changes are posted to the UI thread (see throttledProgress).
const progressResolution = 1000

// pendingUpdate holds an update that has been prepared and is waiting for the application to quit. It is applied from
// the quitting callback, since unison.AttemptQuit does not return when the quit succeeds -- it reaches xos.Exit -- so
// code after the call runs only when the quit was refused.
var pendingUpdate struct {
	state     *updater.State
	statePath string
	lock      sync.Mutex
	applying  bool
}

// applyPendingUpdate starts the helper that finishes a prepared update. It is registered as unison's quitting callback,
// so it runs on every exit that still executes Go code -- a quit request, the last window closing, or a termination
// signal -- and does nothing at all unless an update is waiting.
func applyPendingUpdate() {
	pendingUpdate.lock.Lock()
	state, statePath := pendingUpdate.state, pendingUpdate.statePath
	pendingUpdate.state = nil
	pendingUpdate.lock.Unlock()
	if state == nil {
		return
	}
	if err := updater.SpawnFinisher(state, statePath); err != nil {
		// There is no interface left to report this through; the application is already on its way out. The prepared
		// update is left behind, and the next launch reports it and clears it away.
		errs.Log(errs.NewWithCause("unable to start the update", err))
	}
}

// InitiateAppUpdate prepares the update described by plan and, once it is verified and ready, asks the application to
// quit so that it can be applied. The work happens in a staging directory, so canceling, failing or a refused quit
// leaves the installation untouched.
func InitiateAppUpdate(plan *updater.Plan) {
	pendingUpdate.lock.Lock()
	if pendingUpdate.applying {
		pendingUpdate.lock.Unlock()
		return
	}
	pendingUpdate.applying = true
	pendingUpdate.lock.Unlock()
	defer func() {
		pendingUpdate.lock.Lock()
		pendingUpdate.applying = false
		pendingUpdate.lock.Unlock()
	}()

	staged, ok := stageAppUpdate(plan)
	if !ok {
		return
	}

	state := plan.State(staged, handoffPort)
	statePath := updater.StatePath()
	if err := state.Save(statePath); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to prepare the update"), err)
		updater.Discard(staged)
		return
	}

	pendingUpdate.lock.Lock()
	pendingUpdate.state = state
	pendingUpdate.statePath = statePath
	pendingUpdate.lock.Unlock()

	// AttemptQuit only returns if the quit was refused, which is what reaching the next line means. An open document
	// with unsaved changes is the usual reason.
	unison.AttemptQuit()

	pendingUpdate.lock.Lock()
	pendingUpdate.state = nil
	pendingUpdate.lock.Unlock()
	updater.Discard(staged)
	if err := updater.ClearState(statePath); err != nil {
		errs.Log(err)
	}
	unison.WarningDialogWithMessage(i18n.Text("The update was not installed"),
		xstrings.Wrap("", i18n.Text("%s could not quit, so the update was discarded. It will be offered again the next time you check for updates.",
			xos.AppName), 100))
}

// stageAppUpdate runs the download and verification behind a progress window, returning the prepared update. It reports
// its own failures, so the caller only needs to know whether to carry on.
func stageAppUpdate(plan *updater.Plan) (*updater.Staged, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), stageTimeout)
	defer cancel()
	progress := unison.NewProgressBar(progressResolution)
	wnd, label, err := newProgressWindow(i18n.Text("Updating…"), phaseTitle(updater.PhaseDownloading, plan.ToVersion),
		progress, cancel)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to prepare the update"), err)
		return nil, false
	}

	// The channel must be buffered and read only after RunModal returns (see runInBackground).
	resultChan := make(chan stageResult, 1)
	runInBackground(resultChan,
		func() stageResult {
			staged, stageErr := plan.Stage(ctx, &http.Client{}, throttledProgress(progress), func(phase updater.Phase) {
				unison.InvokeTask(func() { label.SetTitle(phaseTitle(phase, plan.ToVersion)) })
			})
			return stageResult{staged: staged, err: stageErr}
		},
		func() { unison.InvokeTask(func() { wnd.StopModal(unison.ModalResponseOK) }) })
	wnd.RunModal()

	result := <-resultChan
	if result.err != nil {
		if ctx.Err() != nil {
			// The user asked for this, so there is nothing to report.
			return nil, false
		}
		Workspace.ErrorHandler(i18n.Text("Unable to prepare the update"), result.err)
		return nil, false
	}
	return result.staged, true
}

// stageResult carries the outcome of the preparation back to the UI thread.
type stageResult struct {
	staged *updater.Staged
	err    error
}

// throttledProgress returns a progress reporter that posts to the UI thread only when the bar would actually move,
// rather than once per read, which would flood the task queue during a fast download.
func throttledProgress(bar *unison.ProgressBar) func(float64) {
	last := -1
	return func(fraction float64) {
		current := int(fraction * progressResolution)
		if current == last {
			return
		}
		last = current
		unison.InvokeTask(func() { bar.SetCurrent(float32(current)) })
	}
}

// phaseTitle describes what the update is doing at the moment.
func phaseTitle(phase updater.Phase, version string) string {
	switch phase {
	case updater.PhaseExtracting:
		return i18n.Text("Unpacking the update…")
	case updater.PhaseVerifying:
		return i18n.Text("Verifying the update…")
	case updater.PhasePreparing:
		return i18n.Text("Preparing to install…")
	default:
		return i18n.Text("Downloading %s %s…", xos.AppName, filterVersion(version))
	}
}
