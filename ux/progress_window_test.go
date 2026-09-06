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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// backgroundHandshakeTimeout is how long these tests wait for the background goroutine before giving up. It is far
// longer than anything that should ever be needed, since it exists only so that a broken hand-off fails the test rather
// than hanging it.
const backgroundHandshakeTimeout = 10 * time.Second

// awaitBackgroundResult receives the background work's result, failing rather than hanging if it never arrives.
func awaitBackgroundResult[T any](t *testing.T, resultChan <-chan T) T {
	t.Helper()
	select {
	case result := <-resultChan:
		return result
	case <-time.After(backgroundHandshakeTimeout):
		t.Fatal("timed out waiting for the background result")
		var zero T
		return zero
	}
}

// awaitBackgroundFinish waits for the background work's finish callback to have closed finished, failing rather than
// hanging if it never does.
func awaitBackgroundFinish(t *testing.T, finished <-chan struct{}) {
	t.Helper()
	select {
	case <-finished:
	case <-time.After(backgroundHandshakeTimeout):
		t.Fatal("timed out waiting for the background work to finish")
	}
}

// TestRunInBackgroundReportsFailure verifies that failed work is reported as a failure. The UI thread receives the
// result only after its modal loop has been stopped, which is the moment a hand-off through shared variables could still
// be racing with the goroutine that produced it and read back a nil error.
func TestRunInBackgroundReportsFailure(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	runInBackground(resultChan, func() error { return want }, func() { close(finished) })
	awaitBackgroundFinish(t, finished)
	c.Equal(want, awaitBackgroundResult(t, resultChan))
}

// TestRunInBackgroundReportsSuccess verifies the same hand-off for work that succeeded, including that a result made
// of more than an error makes it across whole.
func TestRunInBackgroundReportsSuccess(t *testing.T) {
	c := check.New(t)
	want := map[string][]*rule{"B": {{Rule: "Dodge", Book: "B", Page: "374"}}}
	resultChan := make(chan rulesLookupResult, 1)
	runInBackground(resultChan, func() rulesLookupResult { return rulesLookupResult{rules: want} }, func() {})
	result := awaitBackgroundResult(t, resultChan)
	c.NoError(result.err)
	c.Equal(want, result.rules)
}

// TestRunInBackgroundDeliversResultBeforeFinishing is the invariant the hand-off rests on: finish is what ultimately
// stops the modal loop, and the UI thread reads the result as soon as that loop exits, so the result has to already be
// in the channel by the time finish runs. The receive here is non-blocking on purpose -- an empty channel means the UI
// thread could have been released with nothing to read.
func TestRunInBackgroundDeliversResultBeforeFinishing(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan error, 1)
	var (
		delivered bool
		got       error
	)
	done := make(chan struct{})
	runInBackground(resultChan, func() error { return want }, func() {
		select {
		case got = <-resultChan:
			delivered = true
		default:
		}
		close(done)
	})
	awaitBackgroundFinish(t, done)
	c.True(delivered, "the result must be in the channel before the modal loop can be stopped")
	c.Equal(want, got)
}

// TestRunInBackgroundDoesNotBlockWithoutAReceiver verifies that the goroutine completes even though nothing is
// receiving yet. The UI thread cannot receive until its modal loop has been stopped, and the loop is only stopped by
// finish, so a send that blocked would deadlock the application.
func TestRunInBackgroundDoesNotBlockWithoutAReceiver(t *testing.T) {
	c := check.New(t)
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	runInBackground(resultChan, func() error { return nil }, func() { close(finished) })
	awaitBackgroundFinish(t, finished)
	c.NoError(awaitBackgroundResult(t, resultChan))
}

// TestProgressWindowCancelButton verifies that the window offers a Cancel button exactly when the operation can be
// canceled, and that pressing it calls the cancel function, disables itself and rewrites the label to say the
// operation is being canceled.
func TestProgressWindowCancelButton(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	screen.Do(func() {
		wnd, label, err := newProgressWindow("Downloading…", "Downloading the data…", unison.NewProgressBar(0), nil)
		c.NoError(err)
		defer wnd.Dispose()
		c.Equal(0, len(panelsOfType[*unison.Button](wnd.Content())), "no Cancel button without a cancel function")
		c.Equal("Downloading the data…", label.String())
	})
	screen.Do(func() {
		canceled := false
		wnd, label, err := newProgressWindow("Updating…", "Downloading the library…", unison.NewProgressBar(0),
			func() { canceled = true })
		c.NoError(err)
		defer wnd.Dispose()
		buttons := panelsOfType[*unison.Button](wnd.Content())
		c.Equal(1, len(buttons), "exactly one Cancel button with a cancel function")
		c.True(buttons[0].Enabled())
		buttons[0].Click()
		c.True(canceled, "pressing Cancel must call the cancel function")
		c.False(buttons[0].Enabled(), "Cancel must disable itself so it cannot be pressed twice")
		c.Equal("Canceling…", label.String())
	})
}
