// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build !windows && !linux

package library

import "github.com/rjeczalik/notify"

// watchFolderEntry watches for moves of the folder at p, which only Windows and Linux need: elsewhere the platform
// watcher watches a folder by its path.
func watchFolderEntry(_ string, _ func(what notify.Event)) (stop func()) {
	return func() {}
}
