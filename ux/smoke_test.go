// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build smoke

package ux

import (
	"testing"
)

// TestSmokeLaunch starts the application with nothing to open and checks that it comes up with the empty workspace:
// the navigator showing both libraries and nothing open in the document dock.
func TestSmokeLaunch(t *testing.T) {
	s := startSmoke(t)
	var open int
	s.screen.Do(func() { open = len(AllDockables()) })
	s.c.Equal(0, open, "nothing should be open in the document dock")
	s.expectGUI("workspace", nil)
	s.expectScreenshot("workspace", nil)
}
