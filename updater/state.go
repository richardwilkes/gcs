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
	"encoding/json/v2"
	"io"
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// StateSchema is the version of the State layout. Three different builds of GCS read this file -- the one that staged
// the update, the helper that applies it, and the newly installed one that reports on it -- so the layout is versioned
// and every reader tolerates fields it does not recognize. A file whose schema is not understood is discarded rather
// than acted on.
const StateSchema = 1

// Status is where an update has got to.
type Status string

const (
	// StatusStaged means the replacement has been downloaded and verified but not yet swapped in.
	StatusStaged Status = "staged"
	// StatusApplied means the swap succeeded.
	StatusApplied Status = "applied"
	// StatusFailed means the update was abandoned; Reason says why.
	StatusFailed Status = "failed"
)

// Reason identifies why an update could not be applied. These are stable identifiers rather than messages: the file
// they travel in is written by one build of GCS and read by another, possibly running in a different language, so the
// translation has to happen at the point of display rather than here.
type Reason string

const (
	// ReasonNone means nothing went wrong.
	ReasonNone Reason = ""
	// ReasonPredecessorRunning means the application that staged the update never exited, so nothing was touched.
	ReasonPredecessorRunning Reason = "predecessor-still-running"
	// ReasonSwapFailed means the installation could not be replaced. The original is still in place.
	ReasonSwapFailed Reason = "swap-failed"
	// ReasonRelaunchFailed means the new version was installed but could not be started again.
	ReasonRelaunchFailed Reason = "relaunch-failed"
	// ReasonVersionMismatch means the swap reported success but the running version is not the one that was installed.
	ReasonVersionMismatch Reason = "version-mismatch"
	// ReasonAbandoned means an update was staged but never applied, and the staging has now been discarded.
	ReasonAbandoned Reason = "abandoned"
)

// State is the record shared between the application staging an update, the helper applying it, and the application
// that comes back afterwards. It is deliberately flat and made only of strings and integers, so that a build which
// predates or postdates any given field can still read the rest of it.
type State struct {
	Schema      int    `json:"schema"`
	Status      Status `json:"status"`
	Reason      Reason `json:"reason,omitzero"`
	Detail      string `json:"detail,omitzero"`
	FromVersion string `json:"from_version,omitzero"`
	ToVersion   string `json:"to_version,omitzero"`
	Target      string `json:"target"`
	Payload     string `json:"payload,omitzero"`
	Backup      string `json:"backup,omitzero"`
	WorkDir     string `json:"work_dir,omitzero"`
	Helper      string `json:"helper,omitzero"`
	LogPath     string `json:"log_path,omitzero"`
	HandoffPort int    `json:"handoff_port,omitzero"`
	ParentPID   int    `json:"parent_pid,omitzero"`
	Bundle      bool   `json:"bundle,omitzero"`
}

// StatePath returns where the update state lives. It sits in the application data directory rather than in the staging
// directory so that it survives the staging directory being swept, which is what lets a failure still be reported after
// everything else has been cleaned up.
func StatePath() string {
	return filepath.Join(xos.AppDataDir(true), stateName)
}

// Save writes the state atomically. A torn state file would be read by a different build than the one that wrote it,
// with no way to tell a truncated file from a stale one, so the write must not be observable half-done.
func (s *State) Save(path string) error {
	s.Schema = StateSchema
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return errs.Wrap(err)
	}
	return xos.WriteSafeFile(path, func(w io.Writer) error {
		if err := json.MarshalWrite(w, s); err != nil {
			return errs.Wrap(err)
		}
		return nil
	})
}

// LoadState reads the update state. A missing file returns a nil state and no error, since not having an update in
// progress is the normal case rather than a failure.
func LoadState(path string) (*State, error) {
	f, err := os.Open(path) //nolint:gosec // The path is derived from the app data directory, not from user input
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil //nolint:nilnil // No state is the ordinary case, and is not an error
		}
		return nil, errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	var state State
	if err = json.UnmarshalRead(f, &state); err != nil {
		return nil, errs.NewWithCause("unable to read the update state", err)
	}
	if state.Schema != StateSchema {
		return nil, errs.Newf("update state schema %d is not understood", state.Schema)
	}
	return &state, nil
}

// TargetInfo reconstructs what the state describes as the installation being replaced.
func (s *State) TargetInfo() Target {
	kind := KindExecutable
	if s.Bundle {
		kind = KindBundle
	}
	return Target{
		Path:   s.Target,
		Parent: filepath.Dir(s.Target),
		Exec:   s.Target,
		Kind:   kind,
	}
}
