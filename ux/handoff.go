// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"log/slog"
	"net"
	"path/filepath"
	"time"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
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
				atexit.Exit(1)
			}
		}
		// Port is in use, try connecting as a client and handing off our file list
		if conn, err := net.DialTimeout("tcp4", address, time.Second); err == nil && handoff(conn, pathsBuffer) {
			atexit.Exit(0)
		}
		// Client can't reach the server, loop around and start the process handoff again
	}
	slog.Error("failed to become primary instance and unable to handoff to another copy of GCS")
	atexit.Exit(1)
}

func handoff(conn net.Conn, pathsBuffer []byte) bool {
	slog.Info("handing off to primary instance")
	defer xio.CloseIgnoringErrors(conn)
	buffer := make([]byte, len(cmdline.AppIdentifier))
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		errs.Log(err)
		return false
	}
	n, err := conn.Read(buffer)
	if err != nil {
		errs.Log(err)
		return false
	}
	if n != len(buffer) || !bytes.Equal(buffer, []byte(cmdline.AppIdentifier)) {
		errs.Log(errs.New("unexpected app identifier"))
		return false
	}
	buffer = make([]byte, 5)
	buffer[0] = 22
	binary.LittleEndian.PutUint32(buffer[1:], uint32(len(pathsBuffer))) //nolint:gosec // No, this won't overflow
	n, err = conn.Write(buffer)
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
		slog.Info("app is ready", "elapsed", elapsed)
		if elapsed > 10*time.Second {
			slog.Warn("app took an excessive amount of time" + driverNote)
		}
	case <-time.After(2 * time.Minute):
		// This is here to try and ensure GCS doesn't hang around in the background if something goes wrong at startup.
		slog.Error("timed out waiting for app" + driverNote)
		atexit.Exit(1)
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
	if _, err := conn.Write([]byte(cmdline.AppIdentifier)); err != nil {
		errs.Log(err)
		return
	}
	var single [1]byte
	n, err := conn.Read(single[:])
	if err != nil {
		errs.Log(err)
		return
	}
	if n != 1 {
		errs.Log(errs.Newf("unexpected value for n: %d", n))
		return
	}
	if single[0] != 22 {
		errs.Log(errs.Newf("unexpected value for single[0]: %d", single[0]))
		return
	}
	var sizeBuffer [4]byte
	if n, err = conn.Read(sizeBuffer[:]); err != nil {
		errs.Log(err)
		return
	}
	if n != 4 {
		errs.Log(errs.Newf("unexpected value for n: %d", n))
		return
	}
	size := int(binary.LittleEndian.Uint32(sizeBuffer[:]))
	buffer := make([]byte, size)
	if n, err = conn.Read(buffer); err != nil {
		errs.Log(err)
		return
	}
	if n != size {
		errs.Log(errs.Newf("unexpected value for n: %d, size: %d", n, size))
		return
	}
	var paths []string
	if err = json.Unmarshal(buffer, &paths); err != nil {
		errs.Log(err)
		return
	}
	slog.Info("received handoff", "paths", paths)
	pathsChan <- paths
}
