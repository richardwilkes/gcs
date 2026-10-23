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

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// backgroundHandshakeTimeout is a generous bound on the wait for the background goroutine, so that a broken hand-off
// fails the test rather than hanging it.
const backgroundHandshakeTimeout = 10 * time.Second

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

// TestRunInBackgroundDeliversResultBeforeFinishing checks the invariant the hand-off rests on: finish stops the modal
// loop and the UI thread reads the result as soon as that loop exits, so the result, whether a failure or a success
// made of more than an error, must already be in the channel by the time finish runs. The receive is non-blocking on
// purpose -- an empty channel means the UI thread could have been released with nothing to read.
func TestRunInBackgroundDeliversResultBeforeFinishing(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		checkResultDeliveredBeforeFinishing(t, errors.New("download failed"))
	})
	t.Run("success", func(t *testing.T) {
		checkResultDeliveredBeforeFinishing(t,
			rulesLookupResult{rules: map[string][]*rule{"B": {{Rule: "Dodge", Book: "B", Page: "374"}}}})
	})
}

// checkResultDeliveredBeforeFinishing runs work returning want in the background and checks that finish finds it
// already waiting in the channel.
func checkResultDeliveredBeforeFinishing[T any](t *testing.T, want T) {
	t.Helper()
	c := check.New(t)
	resultChan := make(chan T, 1)
	var (
		delivered bool
		got       T
	)
	done := make(chan struct{})
	runInBackground(resultChan, func() T { return want }, func() {
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

func TestProgressWindowCancelButton(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		wnd, label, err := newProgressWindow("Downloading…", "Downloading the data…", unison.NewProgressBar(0), nil)
		c.NoError(err)
		defer wnd.Dispose()
		c.Equal(0, len(uxtest.PanelsOfType[*unison.Button](wnd.Content())), "no Cancel button without a cancel function")
		c.Equal("Downloading the data…", label.String())
	})
	screen.Do(func() {
		canceled := false
		wnd, label, err := newProgressWindow("Updating…", "Downloading the library…", unison.NewProgressBar(0),
			func() { canceled = true })
		c.NoError(err)
		defer wnd.Dispose()
		buttons := uxtest.PanelsOfType[*unison.Button](wnd.Content())
		c.Equal(1, len(buttons), "exactly one Cancel button with a cancel function")
		c.True(buttons[0].Enabled())
		buttons[0].Click()
		c.True(canceled, "pressing Cancel must call the cancel function")
		c.False(buttons[0].Enabled(), "Cancel must disable itself so it cannot be pressed twice")
		c.Equal("Canceling…", label.String())
	})
}
