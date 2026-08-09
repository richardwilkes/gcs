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
	"encoding/json/v2"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
)

const (
	// handoffMarker precedes the length-prefixed payload a secondary instance hands off to the primary one.
	handoffMarker = 22
	// maxHandoffPayloadSize bounds the payload a handoff may carry. The payload is only a JSON array of the absolute
	// paths named on the command line, so this is far more than any real invocation can produce, given that the
	// command line itself is limited to a fraction of this by the OS. Without a bound, any local process able to
	// connect to the handoff port could claim a payload of up to 4GB and force an allocation of that size.
	maxHandoffPayloadSize = 4 << 20
)

func startHandoffService(readyChan chan struct{}, pathsChan chan<- []string, paths []string) {
	const address = "127.0.0.1:13322"
	var pathsBuffer []byte
	slog.Info("starting handoff service")
	now := time.Now()
	for time.Since(now) < time.Minute {
		// First, try to establish our port and become the primary GCS instance
		if listener, err := net.Listen("tcp4", address); err == nil {
			slog.Info("became primary instance")
			go waitForReady(readyChan)
			go acceptHandoff(listener, pathsChan)
			return
		}
		if pathsBuffer == nil {
			var err error
			absPaths := make([]string, len(paths))
			for i, p := range paths {
				if absPaths[i], err = filepath.Abs(p); err != nil {
					absPaths[i] = p
				}
			}
			if pathsBuffer, err = json.Marshal(absPaths); err != nil {
				errs.Log(err, "paths", absPaths)
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
	const driverNote = " to become ready; this may be due to defective input device drivers"
	started := time.Now()
	select {
	case <-readyChan:
		elapsed := time.Since(started)
		slog.Info("app is ready", "elapsed", elapsed.String())
		if elapsed > 10*time.Second {
			slog.Warn("app took an excessive amount of time" + driverNote)
		}
	case <-time.After(2 * time.Minute):
		// This is here to try and ensure GCS doesn't hang around in the background if something goes wrong at startup.
		slog.Error("timed out waiting for app" + driverNote)
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
		slog.Info("handoff connection accepted")
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
	if err := json.Unmarshal(buffer, &paths); err != nil {
		return nil, errs.Wrap(err)
	}
	return paths, nil
}
