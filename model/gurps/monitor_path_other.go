// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build !darwin

package gurps

import "path/filepath"

// resolvePath returns the given path as the OS names it. Resolving the symlinks is all that is needed here: on Windows,
// filepath.EvalSymlinks also restores the case each component has on disk, and everywhere else paths are case-sensitive
// to begin with.
func resolvePath(p string) (string, error) {
	return filepath.EvalSymlinks(p)
}
