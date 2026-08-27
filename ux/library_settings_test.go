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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibraryCheckWantedAfterApply verifies when applying a library's settings is followed by a background check of its
// releases: only for a library that needs one -- its repository having just been changed, or never having been checked
// -- and only when the periodic checks are on. With them set to Never, the Library Explorer checks when its update
// buttons are clicked, and a check made here would be one the user has asked not to have; a library that has already
// been checked and whose repository didn't change has nothing to ask.
func TestLibraryCheckWantedAfterApply(t *testing.T) {
	c := check.New(t)
	unchecked := gurps.NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.True(unchecked.NeedsUpgradeCheck())
	for _, option := range []updatecheck.Option{updatecheck.AtLaunch, updatecheck.Hourly, updatecheck.Daily} {
		c.True(libraryCheckWantedAfterApply(unchecked, option), "an unchecked library must be checked under %s",
			option.Key())
	}
	c.False(libraryCheckWantedAfterApply(unchecked, updatecheck.Never),
		"with the checks off, an unchecked library is left for the on-click check")

	local := gurps.NewLibrary("Local", "", "", "local", t.TempDir())
	c.False(local.NeedsUpgradeCheck())
	for _, option := range updatecheck.Options {
		c.False(libraryCheckWantedAfterApply(local, option), "a library with nothing to check must not be checked under %s",
			option.Key())
	}
}
