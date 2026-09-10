// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibraryPhaseTitleDistinguishesPhases verifies that the window says which part of the update is running. The bar
// starts over when the phase changes, so a label that didn't change with it would read as a bar that had lost its
// place.
func TestLibraryPhaseTitleDistinguishesPhases(t *testing.T) {
	c := check.New(t)
	downloading := libraryPhaseTitle(library.UpdateDownloading, "Master Library", "5.13.0")
	installing := libraryPhaseTitle(library.UpdateInstalling, "Master Library", "5.13.0")
	c.NotEqual(downloading, installing)
	for _, title := range []string{downloading, installing} {
		c.Contains(title, "Master Library")
		c.Contains(title, "5.13", "the version the user is being given must be in the title")
	}
}
