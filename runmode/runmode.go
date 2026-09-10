// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package runmode lets any package register an alternate mode main() can dispatch to instead of the normal UI --
// something that takes over the process and exits, the way -convert, -sync and -text already do.
package runmode

import "flag"

// Mode describes one alternate run mode.
type Mode struct {
	// Name identifies this mode in the "cannot combine modes" error.
	Name string
	// HiddenFlagNames lists the flags that xflag.SetUsage should hide.
	HiddenFlagNames []string
	// Requested reports whether this mode has been requested. Only meaningful after flag.Parse.
	Requested func() bool
	// Start starts the mode with the file list from the command line. Never returns.
	Start func(files []string)
}

// Factory registers a run mode's command line flags as a side effect of being called, and returns the Mode
// describing it. main calls every factory in Factories ahead of flag.Parse, exactly once each.
type Factory func(*flag.FlagSet) Mode

// Factories holds the run mode factories registered by init functions elsewhere, typically one per package that
// wants to offer an alternate mode. main iterates this slice rather than referring to any specific mode by name, so
// which alternate modes exist, if any, is entirely up to what got compiled into this build.
var Factories []Factory
