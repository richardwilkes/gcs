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
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xio"
)

// holdPort takes a port and returns it along with the listener holding it. A port allocated by the system is used
// rather than the real handoff port, so that these tests cannot interfere with a copy of GCS the developer happens to
// be running.
func holdPort(t *testing.T) (listener net.Listener, port int) {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected a TCP address, got %T", listener.Addr())
	}
	return listener, addr.Port
}

// TestWaitForPredecessorReturnsWhenThePortIsFree verifies the ordinary case: the application has already exited, so
// there is nothing to wait for.
func TestWaitForPredecessorReturnsWhenThePortIsFree(t *testing.T) {
	c := check.New(t)
	listener, port := holdPort(t)
	xio.CloseIgnoringErrors(listener)

	started := time.Now()
	c.NoError(waitForPredecessor(os.Getpid(), port))
	c.True(time.Since(started) < 5*time.Second, "a free port should be noticed immediately")
}

// TestWaitForPredecessorWaitsUntilThePortIsReleased is the property the whole wait exists for. Relaunching while the
// port is still held would make the new instance hand its arguments to the old one and exit, so the update would look
// as though it had done nothing at all.
func TestWaitForPredecessorWaitsUntilThePortIsReleased(t *testing.T) {
	c := check.New(t)
	listener, port := holdPort(t)

	done := make(chan error, 1)
	go func() { done <- waitForPredecessor(os.Getpid(), port) }()

	// The wait must still be in progress while the port is held.
	select {
	case <-done:
		t.Fatal("the wait finished while the port was still held")
	case <-time.After(500 * time.Millisecond):
	}

	xio.CloseIgnoringErrors(listener)

	select {
	case err := <-done:
		c.NoError(err)
	case <-time.After(15 * time.Second):
		t.Fatal("the wait did not finish after the port was released")
	}
}

// TestWaitForPredecessorGivesUpOnAPortHeldByAnotherProcess verifies the tiebreaker. Once the application being waited
// for is definitively gone, a port still held belongs to something else -- a second GCS installation, say -- and
// waiting the full timeout for it would abandon an update for no reason.
func TestWaitForPredecessorGivesUpOnAPortHeldByAnotherProcess(t *testing.T) {
	c := check.New(t)
	listener, port := holdPort(t)
	defer xio.CloseIgnoringErrors(listener)

	// A process that has certainly exited, so that the identifier being watched is definitively gone.
	//nolint:gosec // Re-executing this test binary, purely to obtain a process that has certainly exited
	cmd := exec.Command(os.Args[0], "-test.run", "TestWaitForPredecessorGivesUpOnAPortHeldByAnotherProcess", "-h")
	cmd.Env = append(os.Environ(), "GCS_UPDATER_NOOP=1")
	c.NoError(cmd.Start())
	deadPID := cmd.Process.Pid
	_ = cmd.Wait() //nolint:errcheck // Only that it has exited matters, not how

	started := time.Now()
	c.NoError(waitForPredecessor(deadPID, port))
	elapsed := time.Since(started)
	c.True(elapsed >= pidGraceInterval, "the socket outlives its process briefly, so the check must not be immediate")
	c.True(elapsed < waitTimeout, "a port held by something else must not consume the whole timeout")
}

// TestWaitForPredecessorIgnoresAnAbsentProcessID verifies that a state file carrying no process to watch falls back to
// the port alone rather than treating "no process" as "already gone".
func TestWaitForPredecessorIgnoresAnAbsentProcessID(t *testing.T) {
	c := check.New(t)
	c.True(processExists(0))
	c.True(processExists(-1))

	listener, port := holdPort(t)
	xio.CloseIgnoringErrors(listener)
	c.NoError(waitForPredecessor(0, port))
}

// TestProcessExists verifies the liveness check both ways round, since a wrong answer either stalls every update or
// starts the swap while the application is still running.
func TestProcessExists(t *testing.T) {
	c := check.New(t)
	c.True(processExists(os.Getpid()), "this process is certainly running")

	//nolint:gosec // Re-executing this test binary, purely to obtain a process that has certainly exited
	cmd := exec.Command(os.Args[0], "-test.run", "TestProcessExists", "-h")
	c.NoError(cmd.Start())
	pid := cmd.Process.Pid
	_ = cmd.Wait() //nolint:errcheck // Only that it has exited matters, not how
	c.False(processExists(pid), "a process that has exited must not be reported as running")
}
