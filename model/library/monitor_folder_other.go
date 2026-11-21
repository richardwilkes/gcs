// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build !windows

package library

import "errors"

// folderID identifies a folder, which only Windows needs (see keepsStoppedWatches).
type folderID struct{}

// idOfFolder returns the id of the folder at p.
func idOfFolder(_ string) (folderID, error) {
	return folderID{}, errors.ErrUnsupported
}

// findFolder returns the path of the folder with the id, looking for it on the volume that holds the folder at hint, or
// an empty string if no folder there has the id.
func findFolder(_ folderID, _ string) (string, error) {
	return "", errors.ErrUnsupported
}

// shortPath returns p, since only Windows gives names short (8.3) ones.
func shortPath(p string) string {
	return p
}
