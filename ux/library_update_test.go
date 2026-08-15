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
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// libraryUpdateHandshakeTimeout is how long these tests wait for the background goroutine before giving up. It is far
// longer than anything that should ever be needed, since it exists only so that a broken hand-off fails the test rather
// than hanging it.
const libraryUpdateHandshakeTimeout = 10 * time.Second

// awaitLibraryUpdateResult receives the update's result, failing rather than hanging if it never arrives.
func awaitLibraryUpdateResult(t *testing.T, resultChan <-chan error) error {
	t.Helper()
	select {
	case err := <-resultChan:
		return err
	case <-time.After(libraryUpdateHandshakeTimeout):
		t.Fatal("timed out waiting for the library update result")
		return nil
	}
}

// TestLibraryPhaseTitleDistinguishesPhases verifies that the window says which part of the update is running. The bar
// starts over when the phase changes, so a label that didn't change with it would read as a bar that had lost its
// place.
func TestLibraryPhaseTitleDistinguishesPhases(t *testing.T) {
	c := check.New(t)
	downloading := libraryPhaseTitle(gurps.LibraryUpdateDownloading, "Master Library", "5.13.0")
	installing := libraryPhaseTitle(gurps.LibraryUpdateInstalling, "Master Library", "5.13.0")
	c.NotEqual(downloading, installing)
	for _, title := range []string{downloading, installing} {
		c.Contains(title, "Master Library")
		c.Contains(title, "5.13", "the version the user is being given must be in the title")
	}
}

// TestRunLibraryUpdateReportsDownloadFailure verifies that a failed download is reported as a failure. The UI thread
// receives the result only after its modal loop has been stopped, which is the moment the old code could still be
// racing with the goroutine that produced it and read back a nil error.
func TestRunLibraryUpdateReportsDownloadFailure(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	go runLibraryUpdate(resultChan, func() error { return want }, func() { close(finished) })
	select {
	case <-finished:
	case <-time.After(libraryUpdateHandshakeTimeout):
		t.Fatal("timed out waiting for the library update to finish")
	}
	c.Equal(want, awaitLibraryUpdateResult(t, resultChan))
}

// TestRunLibraryUpdateReportsSuccess verifies the same hand-off for a download that worked.
func TestRunLibraryUpdateReportsSuccess(t *testing.T) {
	c := check.New(t)
	resultChan := make(chan error, 1)
	go runLibraryUpdate(resultChan, func() error { return nil }, func() {})
	c.NoError(awaitLibraryUpdateResult(t, resultChan))
}

// TestRunLibraryUpdateDeliversResultBeforeFinishing is the invariant the fix rests on: finish is what ultimately stops
// the modal loop, and the UI thread reads the result as soon as that loop exits, so the result has to already be in the
// channel by the time finish runs. The receive here is non-blocking on purpose -- an empty channel means the UI thread
// could have been released with nothing to read.
func TestRunLibraryUpdateDeliversResultBeforeFinishing(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan error, 1)
	var (
		delivered bool
		got       error
	)
	done := make(chan struct{})
	go runLibraryUpdate(resultChan, func() error { return want }, func() {
		select {
		case got = <-resultChan:
			delivered = true
		default:
		}
		close(done)
	})
	select {
	case <-done:
	case <-time.After(libraryUpdateHandshakeTimeout):
		t.Fatal("timed out waiting for the library update to finish")
	}
	c.True(delivered, "the result must be in the channel before the modal loop can be stopped")
	c.Equal(want, got)
}

// TestRunLibraryUpdateDoesNotBlockWithoutAReceiver verifies that the goroutine completes even though nothing is
// receiving yet. The UI thread cannot receive until its modal loop has been stopped, and the loop is only stopped by
// finish, so a send that blocked would deadlock the application.
func TestRunLibraryUpdateDoesNotBlockWithoutAReceiver(t *testing.T) {
	c := check.New(t)
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	go runLibraryUpdate(resultChan, func() error { return nil }, func() { close(finished) })
	select {
	case <-finished:
	case <-time.After(libraryUpdateHandshakeTimeout):
		t.Fatal("the library update goroutine blocked before finishing")
	}
	c.NoError(awaitLibraryUpdateResult(t, resultChan))
}
