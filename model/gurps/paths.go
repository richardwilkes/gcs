// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"os/user"
	"path/filepath"
)

// DefaultRootLibraryPath returns the default root library path.
func DefaultRootLibraryPath() string {
	var home string
	if u, err := user.Current(); err != nil {
		home = os.Getenv("HOME")
	} else {
		home = u.HomeDir
	}
	return filepath.Join(home, "GCS")
}

// DefaultMasterLibraryPath returns the default master library path.
func DefaultMasterLibraryPath() string {
	return filepath.Join(DefaultRootLibraryPath(), "Master Library")
}

// DefaultUserLibraryPath returns the default user library path.
func DefaultUserLibraryPath() string {
	return filepath.Join(DefaultRootLibraryPath(), "User Library")
}
