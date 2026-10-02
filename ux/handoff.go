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
	"bytes"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
)

const (
	// handoffPort is the port the primary instance holds for its entire life. The updater's helper waits until it can
	// bind it, which is how it knows the application has exited and that the replacement it starts can become the
	// primary instance.
	handoffPort = 13322
	// handoffMarker precedes the length-prefixed payload a secondary instance hands off to the primary one.
	handoffMarker = 22
	// maxHandoffPayloadSize bounds the payload a handoff may carry. The payload is only a JSON array of the absolute
	// paths named on the command line plus those of the files the OS asked the process to open at launch, so this is
	// far more than any real invocation can produce, given that the OS limits both the command line and such a launch
	// request to a fraction of this. Without a bound, any local process able to connect to the handoff port could claim
	// a payload of up to 4GB and force an allocation of that size.
	maxHandoffPayloadSize = 4 << 20
	// readyTimeout bounds how long startup may take before the app is presumed to be wedged.
	readyTimeout = 2 * time.Minute
	// orderlyExitGrace bounds how long the orderly shutdown triggered by readyTimeout expiring is given to finish
	// before the process is terminated outright.
	orderlyExitGrace = 5 * time.Second
)

func startHandoffService(readyChan chan struct{}, pathsChan chan<- []string, paths []string) {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(handoffPort))
	var pathsBuffer []byte
	now := time.Now()
	for time.Since(now) < time.Minute {
		// First, try to establish our port and become the primary GCS instance
		if listener, err := net.Listen("tcp4", address); err == nil {
			go waitForReady(readyChan)
			go acceptHandoff(listener, pathsChan)
			return
		}
		if pathsBuffer == nil {
			// Files the Finder or LaunchServices hands to GCS on macOS (a double-clicked file, an "Open With" choice, a
			// file dropped on the Dock icon) are not named on the command line. They arrive as an Apple Event, which
			// AppKit delivers only once the application has run its launch phase -- normally done by unison.Start,
			// which this instance never reaches if the handoff succeeds. FilesRequestedAtLaunch performs that phase and
			// returns them, so they can be handed off along with the command-line paths rather than silently dropped.
			// It is called only here, once another copy has been found holding the port, and on the main thread, as it
			// requires. Should the handoff fail and this instance later become the primary after all, nothing is lost
			// or duplicated: unison keeps those files queued and delivers them to the OpenFilesCallback once Start is
			// called, while the paths StartOptions opens itself are only the command-line ones.
			handoffList := handoffPaths(paths, unison.FilesRequestedAtLaunch())
			var err error
			if pathsBuffer, err = jio.Marshal(handoffList); err != nil {
				errs.Log(err, "paths", handoffList)
				xos.Exit(1)
			}
		}
		// Port is in use, try connecting as a client and handing off our file list
		if conn, err := net.DialTimeout("tcp4", address, time.Second); err == nil && handoff(conn, pathsBuffer) {
			xos.Exit(0)
		}
		// Client can't reach the server, loop around and start the process handoff again
	}
	slog.Error("failed to become primary instance and unable to handoff to another copy of GCS")
	xos.Exit(1)
}

// handoffPaths returns the paths a secondary instance hands off to the primary one: the absolute forms of the paths
// named on the command line (each kept as given should it not be resolvable), followed by the files the OS asked this
// process to open at launch, which are already absolute and so are kept as given.
func handoffPaths(cmdLinePaths, launchFiles []string) []string {
	paths := make([]string, 0, len(cmdLinePaths)+len(launchFiles))
	for _, p := range cmdLinePaths {
		if absPath, err := filepath.Abs(p); err == nil {
			p = absPath
		}
		paths = append(paths, p)
	}
	return append(paths, launchFiles...)
}

func handoff(conn net.Conn, pathsBuffer []byte) bool {
	slog.Info("handing off to primary instance")
	defer xio.CloseIgnoringErrors(conn)
	buffer := make([]byte, len(xos.AppIdentifier))
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		errs.Log(err)
		return false
	}
	if _, err := io.ReadFull(conn, buffer); err != nil {
		errs.Log(err)
		return false
	}
	if !bytes.Equal(buffer, []byte(xos.AppIdentifier)) {
		errs.Log(errs.New("unexpected app identifier"))
		return false
	}
	if len(pathsBuffer) > maxHandoffPayloadSize {
		errs.Log(errs.Newf("handoff payload size of %d exceeds the maximum of %d", len(pathsBuffer),
			maxHandoffPayloadSize))
		return false
	}
	buffer = make([]byte, 5)
	buffer[0] = handoffMarker
	binary.LittleEndian.PutUint32(buffer[1:], uint32(len(pathsBuffer))) //nolint:gosec // No, this won't overflow
	n, err := conn.Write(buffer)
	if err != nil {
		errs.Log(err)
		return false
	}
	if n != len(buffer) {
		errs.Log(errs.Newf("unexpected value for n: %d, len(buffer): %d", n, len(buffer)))
		return false
	}
	if n, err = conn.Write(pathsBuffer); err != nil {
		errs.Log(err)
		return false
	}
	if n != len(pathsBuffer) {
		errs.Log(errs.Newf("unexpected value for n: %d, len(pathsBuffer): %d", n, len(pathsBuffer)))
		return false
	}
	return true
}

func waitForReady(readyChan <-chan struct{}) {
	const driverNote = " to become ready; this may be due to defective graphics or input device drivers, or to " +
		"third-party software that hooks them"
	started := time.Now()
	select {
	case <-readyChan:
		elapsed := time.Since(started)
		if elapsed > 10*time.Second {
			slog.Warn("app took an excessive amount of time" + driverNote)
		}
	case <-time.After(readyTimeout):
		// This is here to try and ensure GCS doesn't hang around in the background if something goes wrong at startup.
		slog.Error("timed out waiting for app"+driverNote, "workaround",
			"set "+unison.CPURenderingEnvKey+"=1 in the environment to bypass hardware-accelerated rendering")
		// xos.Exit() runs the registered exit functions, one of which hands the window teardown to the UI thread and
		// waits without bound for it to complete. Arriving here means the UI thread is almost certainly stuck, so
		// that wait would never return and the process would linger in the background -- the very thing this timeout
		// exists to prevent. Arm an unconditional exit first so that termination happens either way.
		time.AfterFunc(orderlyExitGrace, func() {
			slog.Error("orderly shutdown did not complete; terminating")
			os.Exit(1)
		})
		xos.Exit(1)
	}
}

func acceptHandoff(listener net.Listener, pathsChan chan<- []string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			errs.Log(err)
			break
		}
		go processHandoff(conn, pathsChan)
	}
}

func processHandoff(conn net.Conn, pathsChan chan<- []string) {
	defer xio.CloseIgnoringErrors(conn)
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		errs.Log(err)
		return
	}
	if _, err := conn.Write([]byte(xos.AppIdentifier)); err != nil {
		errs.Log(err)
		return
	}
	paths, err := readHandoffPaths(conn)
	if err != nil {
		errs.Log(err)
		return
	}
	slog.Info("received handoff", "paths", paths)
	pathsChan <- paths
}

// readHandoffPaths reads a handoff payload -- the marker byte, a little-endian 32-bit byte count, then that many bytes
// of JSON -- and returns the paths it holds. The byte count is checked before it is used to allocate, since it arrives
// from an unauthenticated local connection.
func readHandoffPaths(r io.Reader) ([]string, error) {
	var single [1]byte
	if _, err := io.ReadFull(r, single[:]); err != nil {
		return nil, errs.Wrap(err)
	}
	if single[0] != handoffMarker {
		return nil, errs.Newf("unexpected value for single[0]: %d", single[0])
	}
	var sizeBuffer [4]byte
	if _, err := io.ReadFull(r, sizeBuffer[:]); err != nil {
		return nil, errs.Wrap(err)
	}
	size := binary.LittleEndian.Uint32(sizeBuffer[:])
	if size > maxHandoffPayloadSize {
		return nil, errs.Newf("handoff payload size of %d exceeds the maximum of %d", size, maxHandoffPayloadSize)
	}
	buffer := make([]byte, size)
	if _, err := io.ReadFull(r, buffer); err != nil {
		return nil, errs.Wrap(err)
	}
	var paths []string
	if err := jio.Unmarshal(buffer, &paths); err != nil {
		return nil, errs.Wrap(err)
	}
	return paths, nil
}
