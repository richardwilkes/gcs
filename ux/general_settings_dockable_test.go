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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// sentinelDeepSearchKey is a key that mapDeepSearch would never produce, so its survival proves the map was left alone.
const sentinelDeepSearchKey = "sentinel"

// installSentinelNavigator swaps in a Navigator whose deep-search map holds only the sentinel key, restoring the prior
// Navigator and the settings the checkboxes mutate when the test finishes.
func installSentinelNavigator(t *testing.T) (*Navigator, *gurps.Settings) {
	t.Helper()
	settings := gurps.GlobalSettings()
	savedOpenInWindow := settings.OpenInWindow
	savedDeepSearch := settings.DeepSearch
	savedNavigator := Workspace.Navigator
	t.Cleanup(func() {
		settings.OpenInWindow = savedOpenInWindow
		settings.DeepSearch = savedDeepSearch
		Workspace.Navigator = savedNavigator
	})
	settings.OpenInWindow = nil
	settings.DeepSearch = nil
	nav := &Navigator{deepSearch: map[string]bool{sentinelDeepSearchKey: true}}
	Workspace.Navigator = nav
	return nav, settings
}

// TestOpenInWindowCheckboxLeavesDeepSearchAlone verifies that toggling a "Use Separate Windows" checkbox updates only
// the OpenInWindow setting. It used to also rebuild the Library Explorer's deep-search extension map, a copy/paste
// leftover from the deep-search checkboxes that has nothing to do with which documents open in their own window.
func TestOpenInWindowCheckboxLeavesDeepSearchAlone(t *testing.T) {
	c := check.New(t)
	nav, settings := installSentinelNavigator(t)

	d := &generalSettingsDockable{}
	d.createOpenInWindowCheckboxes(unison.NewPanel())
	c.True(len(d.openInWindowCheckbox) != 0, "expected at least one separate-window checkbox")

	box := d.openInWindowCheckbox[0]
	group, ok := box.ClientData()["group"].(dgroup.Group)
	c.True(ok, "expected the checkbox to carry its group")

	box.State = uncheck.On
	box.ClickCallback()
	c.True(slices.Contains(settings.OpenInWindow, group), "checking the box should record the group")
	c.Equal(map[string]bool{sentinelDeepSearchKey: true}, nav.deepSearch,
		"toggling a separate-window checkbox must not rebuild the deep search map")

	box.State = uncheck.Off
	box.ClickCallback()
	c.True(!slices.Contains(settings.OpenInWindow, group), "unchecking the box should drop the group")
	c.Equal(map[string]bool{sentinelDeepSearchKey: true}, nav.deepSearch,
		"toggling a separate-window checkbox must not rebuild the deep search map")
}

// TestDeepSearchCheckboxRebuildsDeepSearch is the counterpart to TestOpenInWindowCheckboxLeavesDeepSearchAlone: the
// deep-search checkboxes are the ones that must rebuild the Library Explorer's extension map.
func TestDeepSearchCheckboxRebuildsDeepSearch(t *testing.T) {
	c := check.New(t)
	nav, settings := installSentinelNavigator(t)
	RegisterKnownFileTypes()

	d := &generalSettingsDockable{}
	d.createDeepSearchCheckboxes(unison.NewPanel())
	c.True(len(d.deepSearchableCheckbox) != 0, "expected at least one deep search checkbox")

	box := d.deepSearchableCheckbox[0]
	ext, ok := box.ClientData()["ext"].(string)
	c.True(ok, "expected the checkbox to carry its extension")

	box.State = uncheck.On
	box.ClickCallback()
	c.True(slices.Contains(settings.DeepSearch, ext), "checking the box should record the extension")
	c.True(nav.deepSearch[ext], "checking the box should add the extension to the deep search map")
	c.True(!nav.deepSearch[sentinelDeepSearchKey], "the deep search map should have been rebuilt from scratch")
}
