// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package state

import (
	"sync/atomic"
	"time"
)

var state atomic.Int32

// State is the state of the server.
type State int32

// Possible states for the server.
const (
	Stopped State = iota
	Starting
	Running
	Stopping
)

// Current returns the current state of the server.
func Current() State {
	return State(state.Load())
}

// Set the state of the server.
func Set(newState State) {
	state.Store(int32(newState))
}

// WaitUntil the server is in one of the specified states.
func WaitUntil(state ...State) {
	for {
		current := Current()
		for _, s := range state {
			if current == s {
				return
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}
