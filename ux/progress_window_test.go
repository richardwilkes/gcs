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

// backgroundHandshakeTimeout is how long these tests wait for the background goroutine. It is far longer than anything
// that should ever be needed, since it exists only so that a broken hand-off fails the test rather than hanging it.
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

// TestRunInBackgroundReportsFailure guards against failed work being observed as a success, which a hand-off through
// variables shared with the UI thread could still be racing over when the modal loop is stopped.
func TestRunInBackgroundReportsFailure(t *testing.T) {
	c := check.New(t)
	want := errors.New("download failed")
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	runInBackground(resultChan, func() error { return want }, func() { close(finished) })
	awaitBackgroundFinish(t, finished)
	c.Equal(want, awaitBackgroundResult(t, resultChan))
}

// TestRunInBackgroundReportsSuccess checks the same hand-off for work that succeeded, with a result made of more than
// an error.
func TestRunInBackgroundReportsSuccess(t *testing.T) {
	c := check.New(t)
	want := map[string][]*rule{"B": {{Rule: "Dodge", Book: "B", Page: "374"}}}
	resultChan := make(chan rulesLookupResult, 1)
	runInBackground(resultChan, func() rulesLookupResult { return rulesLookupResult{rules: want} }, func() {})
	result := awaitBackgroundResult(t, resultChan)
	c.NoError(result.err)
	c.Equal(want, result.rules)
}

// TestRunInBackgroundDeliversResultBeforeFinishing checks the invariant the hand-off rests on: finish stops the modal
// loop and the UI thread reads the result as soon as that loop exits, so the result must already be in the channel by
// the time finish runs. The receive is non-blocking on purpose -- an empty channel means the UI thread could have been
// released with nothing to read.
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

// TestRunInBackgroundDoesNotBlockWithoutAReceiver checks that the goroutine completes with nothing receiving yet. Only
// finish stops the modal loop the UI thread must leave before it can receive, so a send that blocked would deadlock.
func TestRunInBackgroundDoesNotBlockWithoutAReceiver(t *testing.T) {
	c := check.New(t)
	resultChan := make(chan error, 1)
	finished := make(chan struct{})
	runInBackground(resultChan, func() error { return nil }, func() { close(finished) })
	awaitBackgroundFinish(t, finished)
	c.NoError(awaitBackgroundResult(t, resultChan))
}

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
