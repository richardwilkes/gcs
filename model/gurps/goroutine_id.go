// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"runtime"
)

var goroutineHeaderPrefix = []byte("goroutine ")

// currentGoroutineID returns the ID of the calling goroutine, or 0 if it cannot be determined.
//
// Go deliberately offers no API for this, but the runtime assigns every goroutine a 64-bit ID that is never reused for
// the life of the process, and the header line runtime.Stack writes for the current goroutine ("goroutine 4707
// [running]:") carries it. That header has had the same form since Go 1.0, and reading the ID out of it is the same
// technique golang.org/x/net/http2 uses (vendored into net/http as http2curGoroutineID). Only the header is needed, so
// the buffer is sized for just that; the runtime still unwinds the goroutine's frames to produce the trace it then
// discards, so a call costs a few microseconds -- small next to compiling or running a script, which is what the
// callers are about to do. With all=false, runtime.Stack does not stop the world.
//
// Should the header ever change form, every caller receives 0. That collapses them onto a single shared ID rather than
// misattributing one goroutine's work to another, so a caller that keys per-goroutine state by the result degrades to
// process-wide state instead of corrupting it.
func currentGoroutineID() uint64 {
	var buf [64]byte
	header, ok := bytes.CutPrefix(buf[:runtime.Stack(buf[:], false)], goroutineHeaderPrefix)
	if !ok {
		return 0
	}
	var id uint64
	for _, ch := range header {
		if ch == ' ' {
			return id
		}
		if ch < '0' || ch > '9' {
			return 0
		}
		id = id*10 + uint64(ch-'0')
	}
	return 0
}
