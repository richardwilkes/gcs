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

	"github.com/richardwilkes/toolbox/v2/check"
)

// Everything the helper needs must survive being written and read back. The file is the only channel between the
// application that stages an update and the process that applies it, so a field lost here is one the helper silently
// does without.
func TestStateRoundTrip(t *testing.T) {
	c := check.New(t)
	path := filepath.Join(t.TempDir(), stateName)
	want := &State{
		Status:      StatusStaged,
		Reason:      ReasonNone,
		FromVersion: "5.45.2",
		ToVersion:   "5.46.0",
		Target:      "/Applications/GCS.app",
		Payload:     "/Applications/.gcs-update-abc/GCS.app",
		Backup:      "/Applications/.GCS.app.old-123",
		WorkDir:     "/Applications/.gcs-update-abc",
		Helper:      "/Applications/.gcs-update-abc/helper",
		LogPath:     "/Applications/.gcs-update-abc/update.log",
		HandoffPort: 13322,
		ParentPID:   4242,
		Bundle:      true,
	}
	c.NoError(want.Save(path))

	got, err := LoadState(path)
	c.NoError(err)
	c.NotNil(got)
	c.Equal(StateSchema, got.Schema, "Save must stamp the schema")
	want.Schema = StateSchema
	c.Equal(want, got)
}

// Having no update in progress is the ordinary case rather than an error; every launch takes this path.
func TestLoadStateWithNoFile(t *testing.T) {
	c := check.New(t)
	state, err := LoadState(filepath.Join(t.TempDir(), "does-not-exist.json"))
	c.NoError(err)
	c.Nil(state)
}

// A corrupt file, or one written by a build using a layout this one does not know, must be refused rather than acted
// on: acting on a half-understood state could mean deleting the wrong directory.
func TestLoadStateRejectsUnusableFiles(t *testing.T) {
	c := check.New(t)
	for name, content := range map[string]string{
		"truncated":     `{"schema":1,"status":"sta`,
		"not json":      `this is not json at all`,
		"empty":         ``,
		"future schema": `{"schema":99,"status":"staged","target":"/Applications/GCS.app"}`,
		"no schema":     `{"status":"staged","target":"/Applications/GCS.app"}`,
	} {
		path := filepath.Join(t.TempDir(), stateName)
		c.NoError(os.WriteFile(path, []byte(content), 0o600), name)
		state, err := LoadState(path)
		c.HasError(err, name)
		c.Nil(state, name)
	}
}

// Forward tolerance: a state written by a newer build carrying unknown fields must still be readable, or an update
// that installs a newer GCS could leave the newly installed application unable to report on its own installation.
func TestLoadStateIgnoresUnknownFields(t *testing.T) {
	c := check.New(t)
	path := filepath.Join(t.TempDir(), stateName)
	c.NoError(os.WriteFile(path, []byte(`{
		"schema": 1,
		"status": "applied",
		"target": "/Applications/GCS.app",
		"to_version": "5.46.0",
		"something_added_later": {"nested": [1, 2, 3]},
		"another_new_field": "value"
	}`), 0o600))
	state, err := LoadState(path)
	c.NoError(err)
	c.NotNil(state)
	c.Equal(StatusApplied, state.Status)
	c.Equal("5.46.0", state.ToVersion)
	c.Equal("/Applications/GCS.app", state.Target)
}

// The helper must be able to reconstruct what it is replacing without re-deriving it from its own executable path,
// which would be wrong, since the helper runs from the staging directory rather than from the installation.
func TestStateTargetInfo(t *testing.T) {
	c := check.New(t)

	bundle := (&State{Target: hostPath("/Applications/GCS.app"), Bundle: true}).TargetInfo()
	c.Equal(KindBundle, bundle.Kind)
	c.Equal(hostPath("/Applications"), bundle.Parent)

	exe := (&State{Target: hostPath("/home/someone/bin/gcs")}).TargetInfo()
	c.Equal(KindExecutable, exe.Kind)
	c.Equal(hostPath("/home/someone/bin"), exe.Parent)
}

// A state file must never be observed half-written. Readers are separate processes, and a truncated file is
// indistinguishable from a stale one.
func TestStateSaveIsAtomic(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	path := filepath.Join(dir, stateName)
	c.NoError((&State{Status: StatusStaged, Target: "/one"}).Save(path))
	c.NoError((&State{Status: StatusApplied, Target: "/two"}).Save(path))

	state, err := LoadState(path)
	c.NoError(err)
	c.Equal(StatusApplied, state.Status)
	c.Equal("/two", state.Target)

	entries, err := os.ReadDir(dir)
	c.NoError(err)
	c.Equal(1, len(entries), "the atomic write must not leave a temporary file behind")
}
