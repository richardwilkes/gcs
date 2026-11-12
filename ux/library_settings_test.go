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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
)

// TestLibraryCheckWantedAfterApply verifies when applying a library's settings is followed by a background check of its
// releases: only for a library that needs one -- its repository having just been changed, or never having been checked
// -- and only when the periodic checks are on. With them set to Never, the Library Explorer checks when its update
// buttons are clicked, and a check made here would be one the user has asked not to have.
func TestLibraryCheckWantedAfterApply(t *testing.T) {
	c := check.New(t)
	unchecked := library.NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.True(unchecked.NeedsUpgradeCheck())
	for _, option := range []updatecheck.Option{updatecheck.AtLaunch, updatecheck.Hourly, updatecheck.Daily} {
		c.True(libraryCheckWantedAfterApply(unchecked, option), "an unchecked library must be checked under %s",
			option.Key())
	}
	c.False(libraryCheckWantedAfterApply(unchecked, updatecheck.Never),
		"with the checks off, an unchecked library is left for the on-click check")

	local := library.NewLibrary("Local", "", "", "local", t.TempDir())
	c.False(local.NeedsUpgradeCheck())
	for _, option := range updatecheck.Options {
		c.False(libraryCheckWantedAfterApply(local, option), "a library with nothing to check must not be checked under %s",
			option.Key())
	}
}

// TestLibrarySettingsTitle verifies that the settings view's title names the library, falling back to a placeholder for
// a library that has not been named yet, so neither the tab nor the save prompt trails off after the colon.
func TestLibrarySettingsTitle(t *testing.T) {
	c := check.New(t)
	c.Equal("Library Settings: Master Library", librarySettingsTitle("Master Library"))
	c.Equal("Library Settings: Untitled Library", librarySettingsTitle(""))
}

// TestLibraryKeyTakenByOther verifies the collision check that keeps apply() from storing a library under an
// account/repo pair another library already uses, which would silently replace that library in the global set.
func TestLibraryKeyTakenByOther(t *testing.T) {
	c := check.New(t)
	libs := library.NewLibraries()
	dir := t.TempDir()
	existing := library.NewLibrary("Existing", "someone", "", "stuff", dir)
	libs.Store(existing.Key(), existing)

	// A library keeps its own key without it counting as a collision.
	c.False(libraryKeyTakenByOther(libs, "someone", "stuff", existing))

	// A different library pointed at that same account/repo pair collides.
	other := library.NewLibrary("Other", "elsewhere", "", "misc", dir)
	c.True(libraryKeyTakenByOther(libs, "someone", "stuff", other))

	// A brand new library (as created via Navigator.addLibrary) collides with any existing key, including the master
	// and user library keys, but is free to take an unused one.
	fresh := &library.Library{}
	c.True(libraryKeyTakenByOther(libs, "someone", "stuff", fresh))
	master := libs.Master()
	masterConfig := master.Config()
	c.True(libraryKeyTakenByOther(libs, masterConfig.GitHubAccountName, masterConfig.RepoName, fresh))
	c.False(libraryKeyTakenByOther(libs, "someone", "other-stuff", fresh))
	c.False(libraryKeyTakenByOther(libs, "", "local-only", fresh))
}

// TestLibrarySettingsFieldsShrinkToFitTheView verifies that the text in the view's fields has no say in how narrow its
// content can become: an access token far longer than the view is wide neither raises the content's minimum width nor
// makes the content wider than a view that has room for that minimum.
func TestLibrarySettingsFieldsShrinkToFitTheView(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	contentWidth := func(token string, viewWidth float32) float32 {
		d := newLibrarySettingsDockable(library.NewLibrary("Test", "someone", token, "repo", dir))
		content := newSettingsContent(d.initContent)
		scroller := d.newContentScroller(content)
		// Tall enough that no vertical scroll bar takes any of the width.
		scroller.SetFrameRect(geom.NewRect(0, 0, viewWidth, 2000))
		scroller.ValidateLayout()
		return content.FrameRect().Width
	}
	token := "github_pat_" + strings.Repeat("M", 82)
	minimum := contentWidth("", 100)
	c.Equal(minimum, contentWidth(token, 100), "the token does not raise the content's minimum width")
	c.Equal(minimum+50, contentWidth(token, minimum+50), "the content is no wider than a view with room for its minimum")
}
