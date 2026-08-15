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
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/v2/errs"
)

// Phase is the step a staging operation has reached, so the caller can say what is happening.
type Phase byte

const (
	// PhaseDownloading means the release is being fetched.
	PhaseDownloading Phase = iota
	// PhaseExtracting means the downloaded archive is being unpacked.
	PhaseExtracting
	// PhaseVerifying means the unpacked application is being checked.
	PhaseVerifying
	// PhasePreparing means the update is being made ready to apply.
	PhasePreparing
)

// Staged is a verified replacement sitting beside the installation, ready to be swapped in.
type Staged struct {
	WorkDir string
	Payload string
	Helper  string
	LogPath string
}

// Stage downloads the release, unpacks it, verifies it, and proves it runs -- all without touching the installation.
// Everything it creates lives in a single staging directory, and an abort at any point removes that directory and
// leaves the installation exactly as it was.
//
// The staging directory is created beside the installation rather than in the system temporary directory. That is what
// makes the swap a rename within one filesystem, which is atomic and instant, instead of a copy across two, which is
// neither and which cannot be undone halfway.
//
// progress reports the fraction of the download completed, from 0 to 1. phase reports which step is running. Both are
// called from this goroutine, may be nil, and must not block.
func (p *Plan) Stage(ctx context.Context, client *http.Client, progress func(fraction float64), phase func(Phase)) (staged *Staged, err error) {
	report := func(ph Phase) {
		if phase != nil {
			phase(ph)
		}
	}

	// Clear away anything left by an attempt that did not clean up after itself. Each staging directory holds a full
	// copy of the application, so without this a few failed attempts would quietly consume hundreds of megabytes of
	// the user's disk and sit there for a week until the startup sweep got round to them.
	//
	// This does not affect the directory created below: os.MkdirTemp generates a fresh random name and retries until
	// it creates one that did not already exist, so a new staging directory can never collide with, or reuse, one that
	// is already there -- including one a helper is actively working in.
	sweepStrays([]string{p.Target.Parent}, abandonedStagingAge)

	workDir, err := os.MkdirTemp(p.Target.Parent, workDirPrefix)
	if err != nil {
		return nil, errs.NewWithCause("unable to prepare the update", err)
	}
	// The archive goes to the system temporary directory rather than the staging directory: it is thrown away as soon
	// as it has been unpacked, and keeping it out of the installation's directory means a failure part-way through
	// leaves nothing large sitting in the user's Applications folder.
	dlDir, err := os.MkdirTemp("", "gcs-download-")
	if err != nil {
		removeIgnoringErrors(workDir)
		return nil, errs.NewWithCause("unable to prepare the update", err)
	}
	defer removeIgnoringErrors(dlDir)
	defer func() {
		if err != nil {
			removeIgnoringErrors(workDir)
		}
	}()

	report(PhaseDownloading)
	archivePath := filepath.Join(dlDir, p.Asset.Name)
	var reportProgress func(int64)
	if progress != nil && p.Asset.Size > 0 {
		reportProgress = func(read int64) { progress(float64(read) / float64(p.Asset.Size)) }
	}
	if err = Download(ctx, client, p.Asset, archivePath, reportProgress); err != nil {
		return nil, err
	}

	report(PhaseExtracting)
	payload := p.Target.PayloadPath(workDir)
	if err = stagePayload(ctx, archivePath, payload); err != nil {
		return nil, err
	}

	report(PhaseVerifying)
	if err = verifyPayload(ctx, payload, p.Target.Path); err != nil {
		return nil, err
	}
	if err = verifyRuns(ctx, p.Target.ExecWithin(payload), p.ToVersion); err != nil {
		return nil, err
	}

	// The helper is a copy of the *running* build rather than the downloaded one because the command line between the
	// two is a contract: having the same compilation write it and read it means the compiler checks it. Handing the job
	// to the new build would make it a permanent cross-version interface, where a rename years from now silently breaks
	// every older version's ability to update.
	//
	// It is a copy rather than the installed file so that, once the application has exited, nothing under the
	// installation directory is open by any GCS process. That is what lets the swap and the cleanup be unconditional --
	// Windows will not delete a running image, and would otherwise leave the displaced copy behind for a week.
	report(PhasePreparing)
	helper, err := stageHelper(ctx, &p.Target, workDir)
	if err != nil {
		return nil, err
	}

	return &Staged{
		WorkDir: workDir,
		Payload: payload,
		Helper:  helper,
		LogPath: filepath.Join(workDir, logName),
	}, nil
}

// Discard throws away a prepared update. Nothing in the installation has been touched at this point, so this only has
// to remove the staging directory.
func Discard(staged *Staged) {
	if staged != nil {
		removeIgnoringErrors(staged.WorkDir)
	}
}

// ClearState removes the record of an update, for the case where one was prepared and then abandoned before the helper
// ever ran. Leaving it would make the next launch report an update that is no longer waiting.
func ClearState(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return errs.Wrap(err)
	}
	return nil
}

// State builds the record handed to the helper and, after the swap, to the newly installed application.
func (p *Plan) State(staged *Staged, port int) *State {
	return &State{
		Schema:      StateSchema,
		Status:      StatusStaged,
		FromVersion: p.FromVersion,
		ToVersion:   p.ToVersion,
		Target:      p.Target.Path,
		Payload:     staged.Payload,
		Backup:      p.Target.FreeBackupPath(),
		WorkDir:     staged.WorkDir,
		Helper:      staged.Helper,
		LogPath:     staged.LogPath,
		HandoffPort: port,
		ParentPID:   os.Getpid(),
		Bundle:      p.Target.Kind == KindBundle,
	}
}
