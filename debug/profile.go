// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package debug

import (
	"os"
	"runtime/pprof"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
)

var (
	lock      sync.Mutex
	pprofFile *os.File
)

// StartCPUProfile starts CPU profiling, writing to "cpu_profile.prof". Does nothing if profiling is already started.
func StartCPUProfile() {
	lock.Lock()
	defer lock.Unlock()
	if pprofFile == nil {
		if f, err := os.Create("cpu_profile.prof"); err == nil {
			pprofFile = f
			if err = pprof.StartCPUProfile(f); err != nil {
				errs.Log(err)
				xio.CloseLoggingErrors(pprofFile)
				pprofFile = nil
			}
		} else {
			errs.Log(err)
		}
	}
}

// StopCPUProfile stops CPU profiling and closes the profile file. Does nothing if profiling isn't started.
func StopCPUProfile() {
	lock.Lock()
	defer lock.Unlock()
	if pprofFile != nil {
		pprof.StopCPUProfile()
		xio.CloseLoggingErrors(pprofFile)
		pprofFile = nil
	}
}
