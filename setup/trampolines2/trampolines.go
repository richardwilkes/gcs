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

package trampolines2

import (
	"sync"

	"github.com/richardwilkes/gcs/v5/model/library"
)

// These functions are here to break what would otherwise be circular dependencies.

var (
	lock                sync.RWMutex
	showLibrarySettings func(lib *library.Library)
)

// CallShowLibrarySettings calls the trampoline that shows the library settings.
func CallShowLibrarySettings(lib *library.Library) {
	lock.RLock()
	f := showLibrarySettings //nolint:ifshort // Can't use short syntax
	lock.RUnlock()
	if f != nil {
		f(lib)
	}
}

// SetShowLibrarySettings sets the trampoline that shows the library settings.
func SetShowLibrarySettings(f func(lib *library.Library)) {
	lock.Lock()
	showLibrarySettings = f
	lock.Unlock()
}
