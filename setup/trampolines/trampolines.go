/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package trampolines

import (
	"sync"

	"github.com/richardwilkes/unison"
)

// These functions are here to break what would otherwise be circular dependencies.

var (
	lock                    sync.RWMutex
	menuSetup               func(wnd *unison.Window)
	libraryUpdatesAvailable func()
)

// CallMenuSetup calls the trampoline that sets up the menus for the given window.
func CallMenuSetup(wnd *unison.Window) {
	lock.RLock()
	f := menuSetup //nolint:ifshort // Can't use short syntax
	lock.RUnlock()
	if f != nil {
		f(wnd)
	}
}

// SetMenuSetup sets the trampoline that sets up the menus for the given window.
func SetMenuSetup(f func(wnd *unison.Window)) {
	lock.Lock()
	menuSetup = f
	lock.Unlock()
}

// CallLibraryUpdatesAvailable calls the trampoline that notifies of library updates becoming available. The trampoline
// will be called via unison.InvokeTask, so is safe to call from a non-UI thread.
func CallLibraryUpdatesAvailable() {
	lock.RLock()
	f := libraryUpdatesAvailable //nolint:ifshort // Can't use short syntax
	lock.RUnlock()
	if f != nil {
		unison.InvokeTask(f)
	}
}

// SetLibraryUpdatesAvailable sets the trampoline that notifies of library updates becoming available.
func SetLibraryUpdatesAvailable(f func()) {
	lock.Lock()
	libraryUpdatesAvailable = f
	lock.Unlock()
}
