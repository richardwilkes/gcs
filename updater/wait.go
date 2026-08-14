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
	"net"
	"strconv"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
)

const (
	// waitTimeout bounds how long the helper will wait for the application to exit before giving up. Giving up means
	// abandoning the update with nothing touched, so a generous bound costs only a delayed update, while too short a
	// one abandons updates whenever quitting takes a while.
	waitTimeout = 90 * time.Second
	// waitPollInterval is how often the port is tried. The wait is normally over within a second or two of the
	// application exiting, so this is short enough to be imperceptible.
	waitPollInterval = 50 * time.Millisecond
	// pidGraceInterval is how long to wait before believing that a departed process leaves the port to someone else.
	// A socket outlives its process by a moment, so checking immediately would mistake that for another instance.
	pidGraceInterval = 2 * time.Second
)

// waitForPredecessor blocks until the application that staged this update has exited.
//
// It waits on the single-instance port rather than on the process, because that is the resource that actually governs
// what happens next: a new instance started while the port is still held hands its arguments to whoever holds it and
// exits immediately, so relaunching too early would look exactly like the update having done nothing at all. Being able
// to bind the port is the direct answer to "can the replacement become the primary instance?", and unlike watching a
// process identifier it cannot be fooled by that identifier being reused.
//
// The listener is closed as soon as it is obtained rather than held through the swap. Holding it would make any GCS the
// user launches meanwhile connect to the helper and stall on a handoff that will never be answered. Releasing it leaves
// a sub-second window in which something else could take the port, which the relaunched instance recovers from on its
// own -- it retries the bind for a full minute before giving up.
func waitForPredecessor(pid, port int) error {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(waitTimeout)
	started := time.Now()
	for time.Now().Before(deadline) {
		listener, err := net.Listen("tcp4", address)
		if err == nil {
			xio.CloseIgnoringErrors(listener)
			return nil
		}
		// The port being held by something else entirely -- a second GCS installation, say -- would otherwise stall
		// this until the timeout. Once the process we are actually waiting for is definitively gone, the port is no
		// longer evidence about it.
		if time.Since(started) > pidGraceInterval && !processExists(pid) {
			slog.Warn("the application has exited but its port is held by something else; continuing", "port", port)
			return nil
		}
		time.Sleep(waitPollInterval)
	}
	return errs.Newf("the running copy of %s did not exit within %s", AppName, waitTimeout)
}

// processExists reports whether a process with the given identifier is still around. A zero or negative identifier
// means the state file carried no process to watch, in which case there is nothing to check and the port alone decides.
//
// This is only ever a tiebreaker for a port held by something other than what we are waiting for, so erring towards
// "still running" is the safe direction: it means waiting longer, never swapping too early.
func processExists(pid int) bool {
	if pid <= 0 {
		return true
	}
	return processIsAlive(pid)
}
