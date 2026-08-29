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

// TestLibrarySettingsTitle verifies that the settings view's title names the library, falling back to a placeholder
// for a library that has not been named yet, so that neither the tab nor the save prompt trails off after the colon.
func TestLibrarySettingsTitle(t *testing.T) {
	c := check.New(t)
	c.Equal("Library Settings: Master Library", librarySettingsTitle("Master Library"))
	c.Equal("Library Settings: Untitled Library", librarySettingsTitle(""))
}

// TestLibraryKeyTakenByOther verifies the collision check that keeps apply() from storing a library under an
// account/repo pair another library already uses, which would silently replace that library in the global set.
func TestLibraryKeyTakenByOther(t *testing.T) {
	c := check.New(t)
	libs := gurps.NewLibraries()
	dir := t.TempDir()
	existing := gurps.NewLibrary("Existing", "someone", "", "stuff", dir)
	libs.Store(existing.Key(), existing)

	// A library keeps its own key without it counting as a collision.
	c.False(libraryKeyTakenByOther(libs, "someone", "stuff", existing))

	// A different library pointed at that same account/repo pair collides.
	other := gurps.NewLibrary("Other", "elsewhere", "", "misc", dir)
	c.True(libraryKeyTakenByOther(libs, "someone", "stuff", other))

	// A brand new library (as created via Navigator.addLibrary) collides with any existing key, including the master
	// and user library keys, but is free to take an unused one.
	fresh := &gurps.Library{}
	c.True(libraryKeyTakenByOther(libs, "someone", "stuff", fresh))
	master := libs.Master()
	masterConfig := master.Config()
	c.True(libraryKeyTakenByOther(libs, masterConfig.GitHubAccountName, masterConfig.RepoName, fresh))
	c.False(libraryKeyTakenByOther(libs, "someone", "other-stuff", fresh))
	c.False(libraryKeyTakenByOther(libs, "", "local-only", fresh))
}
