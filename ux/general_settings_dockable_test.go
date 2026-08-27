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
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
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

// updateCheckStub stands in for the repeating checks' scheduler and work, so that a test can see what the checks were
// asked to do without a real timer being started or the network being reached.
type updateCheckStub struct {
	ticks       []recordedTick
	appRuns     int
	libraryRuns int
}

func (s *updateCheckStub) schedule(f func(), after time.Duration) {
	s.ticks = append(s.ticks, recordedTick{fire: f, after: after})
}

// stubUpdateCheckSchedules replaces the repeating checks with fresh ones that record onto the returned stub for the
// duration of the test. A popup selection made through the real callback would otherwise leave a real hour- or day-long
// timer pending in the test process, and turning the checks on would reach out to the network. The checks are put back
// exactly as they were afterwards; the ticks recorded in between are never fired.
func stubUpdateCheckSchedules(t *testing.T) *updateCheckStub {
	t.Helper()
	stub := &updateCheckStub{}
	savedApp, savedLibrary := appUpdateCheck, libraryUpdateCheck
	t.Cleanup(func() { appUpdateCheck, libraryUpdateCheck = savedApp, savedLibrary })
	appUpdateCheck = periodicCheck{
		frequency: savedApp.frequency,
		run:       func() { stub.appRuns++ },
		schedule:  stub.schedule,
	}
	libraryUpdateCheck = periodicCheck{
		frequency: savedLibrary.frequency,
		run:       func() { stub.libraryRuns++ },
		schedule:  stub.schedule,
	}
	return stub
}

// prepareUpdateCheckSettings installs a Navigator without an update button, so that the ApplyUpdateCheckSettings() call
// the popups make finds nothing to sync, stubs out the repeating checks, and restores the update check settings the
// popups mutate when the test finishes. Returns the general settings, with both update checks set to their default,
// and the stub the repeating checks now record onto.
func prepareUpdateCheckSettings(t *testing.T) (*gurps.GeneralSettings, *updateCheckStub) {
	t.Helper()
	installSentinelNavigator(t)
	stub := stubUpdateCheckSchedules(t)
	gs := gurps.GlobalSettings().General
	savedAppUpdateCheck := gs.AppUpdateCheck
	savedLibraryUpdateCheck := gs.LibraryUpdateCheck
	t.Cleanup(func() {
		gs.AppUpdateCheck = savedAppUpdateCheck
		gs.LibraryUpdateCheck = savedLibraryUpdateCheck
	})
	gs.AppUpdateCheck = updatecheck.AtLaunch
	gs.LibraryUpdateCheck = updatecheck.AtLaunch
	return gs, stub
}

// TestUpdateCheckPopupsApplyTheirSelection verifies that each update check popup starts out showing its own setting and
// writes back to only that setting. The two popups are built by the same helper, so a mixed-up getter or setter would
// leave one of them silently driving the other's setting.
func TestUpdateCheckPopupsApplyTheirSelection(t *testing.T) {
	c := check.New(t)
	gs, _ := prepareUpdateCheckSettings(t)

	d := &generalSettingsDockable{}
	d.createUpdateCheckPopups(unison.NewPanel())
	selected, ok := d.appUpdateCheckPopup.Selected()
	c.True(ok, "the app update popup should have a selection")
	c.Equal(updatecheck.AtLaunch, selected, "the app update popup should start out showing the current setting")
	selected, ok = d.libraryUpdateCheckPopup.Selected()
	c.True(ok, "the library update popup should have a selection")
	c.Equal(updatecheck.AtLaunch, selected, "the library update popup should start out showing the current setting")

	d.appUpdateCheckPopup.Select(updatecheck.Daily)
	c.Equal(updatecheck.Daily, gs.AppUpdateCheck, "selecting in the app update popup should record the choice")
	c.Equal(updatecheck.AtLaunch, gs.LibraryUpdateCheck,
		"the app update popup must not disturb the library update setting")

	d.libraryUpdateCheckPopup.Select(updatecheck.Never)
	c.Equal(updatecheck.Never, gs.LibraryUpdateCheck, "selecting in the library update popup should record the choice")
	c.Equal(updatecheck.Daily, gs.AppUpdateCheck, "the library update popup must not disturb the app update setting")
}

// TestUpdateCheckPopupsRunACheckWhenTurnedOn verifies that choosing anything other than Never in a popup that was at
// Never runs that check right away, for the app and the libraries independently, while other changes don't. The
// popups are the way the checks get turned on mid-session, and the user who does that expects something to happen.
func TestUpdateCheckPopupsRunACheckWhenTurnedOn(t *testing.T) {
	c := check.New(t)
	gs, stub := prepareUpdateCheckSettings(t)
	gs.AppUpdateCheck = updatecheck.Never
	gs.LibraryUpdateCheck = updatecheck.Never
	ApplyUpdateCheckSettings() // What launch does, after making its own checks

	d := &generalSettingsDockable{}
	d.createUpdateCheckPopups(unison.NewPanel())
	c.Equal(0, stub.appRuns, "building the popups must not run a check")
	c.Equal(0, stub.libraryRuns, "building the popups must not run a check")

	d.appUpdateCheckPopup.Select(updatecheck.AtLaunch)
	c.Equal(1, stub.appRuns, "turning the app check on must run it")
	c.Equal(0, stub.libraryRuns, "turning the app check on must not run the library check")
	c.Equal(0, len(stub.ticks), "checking at launch must not schedule a repeating check")

	d.appUpdateCheckPopup.Select(updatecheck.Hourly)
	c.Equal(1, stub.appRuns, "moving between settings that check must not run the check again")

	d.libraryUpdateCheckPopup.Select(updatecheck.Daily)
	c.Equal(1, stub.libraryRuns, "turning the library check on must run it")
	c.Equal(1, stub.appRuns, "turning the library check on must not run the app check")

	d.libraryUpdateCheckPopup.Select(updatecheck.Never)
	d.libraryUpdateCheckPopup.Select(updatecheck.AtLaunch)
	c.Equal(2, stub.libraryRuns, "turning the library check on again must run it again")
}

// TestUpdateCheckPopupsFollowTheSettings verifies that the popups can be brought back into line with settings that were
// replaced wholesale, as a reset or a load of a settings file does. This is what sync() performs; the rest of sync()
// isn't exercised here, since a bare dockable has none of the other widgets it touches.
func TestUpdateCheckPopupsFollowTheSettings(t *testing.T) {
	c := check.New(t)
	gs, _ := prepareUpdateCheckSettings(t)

	d := &generalSettingsDockable{}
	d.createUpdateCheckPopups(unison.NewPanel())

	gs.AppUpdateCheck = updatecheck.Hourly
	gs.LibraryUpdateCheck = updatecheck.Hourly
	d.appUpdateCheckPopup.Select(gs.AppUpdateCheck)
	d.libraryUpdateCheckPopup.Select(gs.LibraryUpdateCheck)

	selected, ok := d.appUpdateCheckPopup.Selected()
	c.True(ok, "the app update popup should have a selection")
	c.Equal(updatecheck.Hourly, selected, "the app update popup should show the replaced setting")
	selected, ok = d.libraryUpdateCheckPopup.Selected()
	c.True(ok, "the library update popup should have a selection")
	c.Equal(updatecheck.Hourly, selected, "the library update popup should show the replaced setting")
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
