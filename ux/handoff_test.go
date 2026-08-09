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
	"encoding/binary"
	"encoding/json/v2"
	"io"
	"math"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// recordingReader hands out a fixed set of bytes and then records the size of every further read that is attempted, so
// a test can tell whether the code under test tried to read (and therefore allocate) a payload it should have refused.
type recordingReader struct {
	data        []byte
	pos         int
	extraReads  []int
	extraErrOut error
}

func (r *recordingReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		r.extraReads = append(r.extraReads, len(p))
		return 0, r.extraErrOut
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// handoffHeader builds the marker byte plus the little-endian byte count that precedes a handoff payload.
func handoffHeader(size uint32) []byte {
	header := make([]byte, 5)
	header[0] = handoffMarker
	binary.LittleEndian.PutUint32(header[1:], size)
	return header
}

// TestReadHandoffPathsRejectsOversizedPayload verifies that a byte count larger than the maximum is refused before it
// is handed to make(). The handoff port is reachable by any local process, so an unchecked count let one of them ask
// for an allocation of up to 4GB, likely OOM-killing GCS.
func TestReadHandoffPathsRejectsOversizedPayload(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		size uint32
	}{
		{name: "one byte past the maximum", size: maxHandoffPayloadSize + 1},
		{name: "the largest value the count can hold", size: math.MaxUint32},
	} {
		r := &recordingReader{data: handoffHeader(one.size), extraErrOut: io.EOF}
		paths, err := readHandoffPaths(r)
		c.HasError(err, one.name)
		c.Nil(paths, one.name)
		// An empty extraReads proves the payload was never read, and therefore never allocated. Without the size
		// check, io.ReadFull would ask for the full claimed count here.
		c.Equal(0, len(r.extraReads), one.name+": the payload must not be read or allocated")
	}
}

// TestReadHandoffPathsRejectsMalformedFraming verifies the rest of the framing checks, so that a bad marker or a
// truncated stream is reported rather than silently treated as an empty path list.
func TestReadHandoffPathsRejectsMalformedFraming(t *testing.T) {
	c := check.New(t)
	payload, err := json.Marshal([]string{"/tmp/a.gcs"})
	c.NoError(err)
	for _, one := range []struct {
		name string
		data []byte
	}{
		{name: "nothing at all", data: nil},
		{name: "wrong marker", data: []byte{handoffMarker + 1, 0, 0, 0, 0}},
		{name: "truncated byte count", data: []byte{handoffMarker, 0, 0}},
		{name: "truncated payload", data: append(handoffHeader(uint32(len(payload))), payload[:len(payload)-1]...)},
		{name: "payload that isn't a path list", data: append(handoffHeader(3), []byte("nope")...)},
		{name: "empty payload", data: handoffHeader(0)},
	} {
		r := &recordingReader{data: one.data, extraErrOut: io.EOF}
		paths, rErr := readHandoffPaths(r)
		c.HasError(rErr, one.name)
		c.Nil(paths, one.name)
	}
}

// TestReadHandoffPathsAcceptsMaximumSizedPayload verifies the bound is inclusive, so a payload of exactly the maximum
// is still accepted.
func TestReadHandoffPathsAcceptsMaximumSizedPayload(t *testing.T) {
	c := check.New(t)
	// A single-element array of one long path: two brackets and two quotes plus the path itself.
	payload := []byte(`["` + strings.Repeat("a", maxHandoffPayloadSize-4) + `"]`)
	c.Equal(maxHandoffPayloadSize, len(payload))
	r := &recordingReader{data: append(handoffHeader(uint32(len(payload))), payload...), extraErrOut: io.EOF}
	paths, err := readHandoffPaths(r)
	c.NoError(err)
	c.Equal(1, len(paths))
}

// useTestAppIdentifier installs an app identifier for the duration of a test. The main entry point normally sets it,
// so it is empty in the test binary, and the handoff handshake exchanges it before anything else.
func useTestAppIdentifier(t *testing.T) {
	t.Helper()
	saved := xos.AppIdentifier
	t.Cleanup(func() { xos.AppIdentifier = saved })
	xos.AppIdentifier = "com.trollworks.gcs"
}

// TestHandoffRefusesOversizedPayload verifies the sending side applies the same bound, so it reports the problem
// against the paths it can name instead of writing a payload the receiver is going to throw away.
func TestHandoffRefusesOversizedPayload(t *testing.T) {
	c := check.New(t)
	useTestAppIdentifier(t)
	client, server := net.Pipe()
	defer xio.CloseIgnoringErrors(server)
	go func() {
		if _, err := server.Write([]byte(xos.AppIdentifier)); err != nil {
			return
		}
		// Drain whatever arrives, so that a sender which fails to apply the bound fails the assertion below instead
		// of blocking forever against the unbuffered pipe.
		if _, err := io.Copy(io.Discard, server); err != nil {
			return
		}
	}()
	c.False(handoff(client, make([]byte, maxHandoffPayloadSize+1)), "an oversized payload must not be sent")
}

// TestHandoffRoundTrip runs the sending and receiving halves against each other, verifying they agree on the wire
// format and that a normal path list still makes it across.
func TestHandoffRoundTrip(t *testing.T) {
	c := check.New(t)
	useTestAppIdentifier(t)
	want := []string{"/tmp/one.gcs", "/tmp/two.gct"}
	payload, err := json.Marshal(want)
	c.NoError(err)
	client, server := net.Pipe()
	defer xio.CloseIgnoringErrors(client)
	pathsChan := make(chan []string, 1)
	go processHandoff(server, pathsChan)
	c.True(handoff(client, payload), "the handoff should succeed")
	select {
	case got := <-pathsChan:
		c.Equal(want, got)
	case <-time.After(5 * time.Second):
		c.True(false, "timed out waiting for the handoff to be delivered")
	}
}
