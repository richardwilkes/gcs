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
	"context"
	"net/http"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibraryUpdateButtonsEnabledBeforeAnyCheck verifies that the Library Explorer's update and release-notes buttons
// are usable for a library with a repository before any check has found a release, and after a check that couldn't
// reach the repository, since with the periodic checks set to Never nothing else fills in a library's releases. A
// library with no repository stays disabled.
func TestLibraryUpdateButtonsEnabledBeforeAnyCheck(t *testing.T) {
	c := check.New(t)
	local := library.NewLibrary("Local", "", "", "local", t.TempDir())
	c.False(libraryUpdateButtonsEnabled(local), "a library without a repository has nothing to offer")

	lib := library.NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.True(libraryUpdateButtonsEnabled(lib), "an unchecked library with a repository must be offered a check")

	// A check that fails -- the context is already canceled, so no request is ever made -- must not disable the
	// buttons, since trying again is the only way to find out what the repository has.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
	c.True(libraryUpdateButtonsEnabled(lib), "a failed check must leave the buttons usable")
}

// TestCheckLibraryReleasesSkipsLibrariesWithNothingToCheck verifies that libraries with no check to make are passed
// straight through, without the window that reports on a check being opened -- which, without a UI, it can't be.
func TestCheckLibraryReleasesSkipsLibrariesWithNothingToCheck(t *testing.T) {
	c := check.New(t)
	local := library.NewLibrary("Local", "", "", "local", t.TempDir())
	c.True(checkLibraryReleases(nil))
	c.True(checkLibraryReleases([]*library.Library{local}))
}
