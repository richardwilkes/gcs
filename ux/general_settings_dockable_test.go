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
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
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

// Toggling a "Use Separate Windows" checkbox must update only the OpenInWindow setting. It used to also rebuild the
// Library Explorer's deep-search extension map, a copy/paste leftover from the deep-search checkboxes.
func TestOpenInWindowCheckboxLeavesDeepSearchAlone(t *testing.T) {
	c := check.New(t)
	nav, settings := installSentinelNavigator(t)

	d := &generalSettingsDockable{}
	d.createOpenInWindowCheckboxes(unison.NewPanel())
	c.True(len(d.openInWindowCheckbox) != 0, "expected at least one separate-window checkbox")

	box, group := d.openInWindowCheckbox[0].box, d.openInWindowCheckbox[0].item

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

// updateCheckStub stands in for the repeating checks' scheduler and work, so a test can see what the checks were asked
// to do without a real timer being started or the network being reached.
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
// timer pending in the test process, and turning the checks on would reach out to the network. The ticks recorded in
// between are never fired.
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

// prepareUpdateCheckSettings installs a Navigator without an update button, so the ApplyUpdateCheckSettings() call the
// popups make finds nothing to sync, stubs out the repeating checks, and restores the update check settings the popups
// mutate when the test finishes. It returns the general settings, with both update checks set to their default, and the
// stub the repeating checks now record onto.
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

// Each update check popup must start out showing its own setting and write back to only that setting. The two popups
// are built by the same helper, so a mixed-up getter or setter would leave one silently driving the other's setting.
func TestUpdateCheckPopupsApplyTheirSelection(t *testing.T) {
	c := check.New(t)
	gs, _ := prepareUpdateCheckSettings(t)
	// Settings that differ from each other and from the default, so that each popup can show only its own.
	gs.AppUpdateCheck = updatecheck.Hourly
	gs.LibraryUpdateCheck = updatecheck.Daily

	d := &generalSettingsDockable{}
	d.createUpdateCheckPopups(unison.NewPanel())
	selected, ok := d.appUpdateCheckPopup.Selected()
	c.True(ok, "the app update popup should have a selection")
	c.Equal(updatecheck.Hourly, selected, "the app update popup should start out showing the current setting")
	selected, ok = d.libraryUpdateCheckPopup.Selected()
	c.True(ok, "the library update popup should have a selection")
	c.Equal(updatecheck.Daily, selected, "the library update popup should start out showing the current setting")

	d.appUpdateCheckPopup.Select(updatecheck.Never)
	c.Equal(updatecheck.Never, gs.AppUpdateCheck, "selecting in the app update popup should record the choice")
	c.Equal(updatecheck.Daily, gs.LibraryUpdateCheck,
		"the app update popup must not disturb the library update setting")

	d.libraryUpdateCheckPopup.Select(updatecheck.AtLaunch)
	c.Equal(updatecheck.AtLaunch, gs.LibraryUpdateCheck,
		"selecting in the library update popup should record the choice")
	c.Equal(updatecheck.Never, gs.AppUpdateCheck, "the library update popup must not disturb the app update setting")
}

// Choosing anything other than Never in a popup that was at Never must run that check right away, for the app and the
// libraries independently, while other changes must not. The popups are the way the checks get turned on mid-session,
// and the user who does that expects something to happen.
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

// The popups must follow settings that were replaced wholesale, as a reset or a load of a settings file does, once
// sync() brings the view into line with them, and selecting the replaced value must bring the repeating check into line
// with it as well, through the popup's own callback. TestPeriodicCheckTickRunsTheWorkAndSchedulesTheNext covers
// re-applying an unchanged setting, which must leave the pending tick alone.
func TestUpdateCheckPopupsFollowTheSettings(t *testing.T) {
	c := check.New(t)
	d, stub := newTestGeneralSettingsDockable(t)
	gs := gurps.GlobalSettings().General
	ApplyUpdateCheckSettings() // What launch does, after making its own checks
	c.Equal(0, len(stub.ticks), "checking at launch must not schedule a repeating check")

	gs.AppUpdateCheck = updatecheck.Hourly
	gs.LibraryUpdateCheck = updatecheck.Hourly
	d.sync()

	selected, ok := d.appUpdateCheckPopup.Selected()
	c.True(ok, "the app update popup should have a selection")
	c.Equal(updatecheck.Hourly, selected, "the app update popup should show the replaced setting")
	selected, ok = d.libraryUpdateCheckPopup.Selected()
	c.True(ok, "the library update popup should have a selection")
	c.Equal(updatecheck.Hourly, selected, "the library update popup should show the replaced setting")
	c.True(libraryUpdateCheck.pending, "selecting the replaced setting must reschedule the check")
	c.Equal(updatecheck.Hourly, libraryUpdateCheck.scheduled)
}

// The counterpart to TestOpenInWindowCheckboxLeavesDeepSearchAlone: the deep-search checkboxes are the ones that must
// rebuild the Library Explorer's extension map.
func TestDeepSearchCheckboxRebuildsDeepSearch(t *testing.T) {
	c := check.New(t)
	nav, settings := installSentinelNavigator(t)
	RegisterKnownFileTypes()

	d := &generalSettingsDockable{}
	d.createDeepSearchCheckboxes(unison.NewPanel())
	c.True(len(d.deepSearchableCheckbox) != 0, "expected at least one deep search checkbox")

	box, ext := d.deepSearchableCheckbox[0].box, d.deepSearchableCheckbox[0].item

	box.State = uncheck.On
	box.ClickCallback()
	c.True(slices.Contains(settings.DeepSearch, ext), "checking the box should record the extension")
	c.True(nav.deepSearch[ext], "checking the box should add the extension to the deep search map")
	c.True(!nav.deepSearch[sentinelDeepSearchKey], "the deep search map should have been rebuilt from scratch")
}

// newTestGeneralSettingsDockable returns a general settings dockable with its content built on a bare panel, the way
// Setup would build it, but without a workspace to dock it in. The settings its widgets and its reset mutate are put
// back when the test ends, and the update checks its popups drive are stubbed by the returned stub.
func newTestGeneralSettingsDockable(t *testing.T) (*generalSettingsDockable, *updateCheckStub) {
	t.Helper()
	_, stub := prepareUpdateCheckSettings(t)
	RegisterKnownFileTypes()
	gs := gurps.GlobalSettings().General
	saved := *gs
	t.Cleanup(func() {
		*gs = saved
		gs.EnsureValidity() // Restores unison's copies of the settings, which a reset in the test may have changed.
		gurps.SyncScriptExecTimeLimit()
	})
	uxtest.SwapForTest(t, &languageSetting, "")
	d := &generalSettingsDockable{}
	d.initContent(unison.NewPanel())
	return d, stub
}

// Every checkbox in the checkbox block and every initial scale field must edit its own field of the live general
// settings, and keep doing so after a reset has replaced the settings wholesale, which also has to bring each of them
// back into line. The widgets hold pointers into the settings, which is only right because reset and load copy into the
// settings struct rather than swapping it out.
func TestGeneralSettingsCheckBoxesAndScaleFieldsEditTheLiveSettings(t *testing.T) {
	c := check.New(t)
	d, _ := newTestGeneralSettingsDockable(t)
	gs := gurps.GlobalSettings().General
	c.Equal(9, len(d.checkBoxes), "expected one checkbox per general setting in the block")
	c.Equal(6, len(d.initialScaleFields), "expected one initial scale field per document kind")

	// Each checkbox must have its own field of the live settings, and show its value.
	seenBools := make(map[*bool]bool)
	for _, one := range d.checkBoxes {
		title := one.box.Text.String()
		c.True(!seenBools[one.value], "%q shares its setting with another checkbox", title)
		seenBools[one.value] = true
		c.Equal(uncheck.FromBool(*one.value), one.box.State, "%q should show its setting", title)
		before := *gs
		*one.value = !*one.value
		c.True(before != *gs, "%q should edit the live settings", title)
		*one.value = !*one.value
	}

	// Likewise for each initial scale field.
	seenInts := make(map[*int]bool)
	for _, one := range d.initialScaleFields {
		title := one.field.undoTitle
		c.True(!seenInts[one.value], "%q shares its setting with another field", title)
		seenInts[one.value] = true
		c.Equal(one.field.Format(*one.value), one.field.Text(), "%q should show its setting", title)
		before := *gs
		*one.value++
		c.True(before != *gs, "%q should edit the live settings", title)
		*one.value--
	}

	// Move a checkbox without a side effect and a scale field off their defaults through the widgets themselves. The
	// checkboxes with side effects reach into a workspace this test does not have.
	var box settingCheckBox
	for _, one := range d.checkBoxes {
		if one.box.OnSet == nil {
			box = one
			break
		}
	}
	c.NotNil(box.box, "expected a checkbox without a side effect")
	defaults := gurps.NewGeneralSettings()
	*gs = *defaults
	d.sync()
	box.box.State = uncheck.FromBool(!*box.value)
	box.box.ClickCallback()
	c.True(*gs != *defaults, "clicking the checkbox should change the live settings")
	c.Equal(box.box.State, uncheck.FromBool(*box.value), "clicking the checkbox should record its state")
	field := d.initialScaleFields[0]
	scale := gurps.InitialUIScaleMin
	if *field.value == scale {
		scale = gurps.InitialUIScaleMax
	}
	field.field.SetText(field.field.Format(scale))
	c.Equal(scale, *field.value, "editing the field should record its value")

	// The reset copies the defaults over the live settings and must bring every widget back into line.
	d.reset()
	c.Equal(*defaults, *gs, "the reset should restore the defaults")
	for _, one := range d.checkBoxes {
		c.Equal(uncheck.FromBool(*one.value), one.box.State, "%q should show the reset setting", one.box.Text.String())
	}
	for _, one := range d.initialScaleFields {
		c.Equal(one.field.Format(*one.value), one.field.Text(), "%q should show the reset setting", one.field.undoTitle)
	}

	// The widgets must still be editing the live settings, not a copy the reset left behind.
	box.box.State = uncheck.FromBool(!*box.value)
	box.box.ClickCallback()
	c.True(*gurps.GlobalSettings().General != *defaults, "clicking the checkbox after the reset should change the live settings")
	field.field.SetText(field.field.Format(scale))
	c.Equal(scale, *field.value, "editing the field after the reset should record its value")
}

func TestFocusForReadingCheckBoxAppliesToUnison(t *testing.T) {
	c := check.New(t)
	d, _ := newTestGeneralSettingsDockable(t)
	gs := gurps.GlobalSettings().General
	*gs = *gurps.NewGeneralSettings()
	gs.EnsureValidity()
	d.sync()
	var box settingCheckBox
	for _, one := range d.checkBoxes {
		if one.value == &gs.FocusForReading {
			box = one
			break
		}
	}
	c.NotNil(box.box, "expected a checkbox for the focus for reading setting")
	c.Equal(uncheck.Off, box.box.State, "the setting is off by default")
	c.False(unison.FocusForReading(), "unison starts with the switch off")

	box.box.State = uncheck.On
	box.box.ClickCallback()
	c.True(gs.FocusForReading, "clicking the checkbox should turn the setting on")
	c.True(unison.FocusForReading(), "clicking the checkbox should hand the setting to unison")

	d.reset()
	c.False(gs.FocusForReading, "the reset should turn the setting off")
	c.False(unison.FocusForReading(), "the reset should hand the setting to unison")
	c.Equal(uncheck.Off, box.box.State, "the reset should bring the checkbox back into line")
}

func TestGeneralSettingsResetRunsTheSideEffectsOfTheBoxesItChanges(t *testing.T) {
	c := check.New(t)
	d, _ := newTestGeneralSettingsDockable(t)
	gs := gurps.GlobalSettings().General
	*gs = *gurps.NewGeneralSettings()
	gs.EnsureValidity()
	d.sync()
	fired := make(map[*bool]int)
	for _, one := range d.checkBoxes {
		one.box.OnSet = func() { fired[one.value]++ }
	}
	var grouping settingCheckBox
	for _, one := range d.checkBoxes {
		if one.value == &gs.GroupContainersOnSort {
			grouping = one
			break
		}
	}
	if grouping.box == nil {
		t.Fatal("expected a checkbox for the grouping setting")
	}
	grouping.box.State = uncheck.FromBool(!gs.GroupContainersOnSort)
	grouping.box.ClickCallback()
	c.Equal(1, fired[&gs.GroupContainersOnSort], "clicking the box has its side effect")
	c.Equal(1, len(fired), "and only its own")

	d.reset()
	c.Equal(uncheck.FromBool(gurps.NewGeneralSettings().GroupContainersOnSort), grouping.box.State,
		"the reset brings the box back into line")
	c.Equal(2, fired[&gs.GroupContainersOnSort], "the reset has the side effect of the box it changed")
	c.Equal(1, len(fired), "the boxes the reset left alone have no side effect")
}

// TestGeneralSettingsResetClearsTheListsAndLocale verifies that a reset also clears the file types searched deeply, the
// groups opened in separate windows and the interface locale, which live outside the general settings proper, bringing
// their widgets and the Library Explorer's deep search into line.
func TestGeneralSettingsResetClearsTheListsAndLocale(t *testing.T) {
	c := check.New(t)
	d, _ := newTestGeneralSettingsDockable(t)
	settings := gurps.GlobalSettings()
	*settings.General = *gurps.NewGeneralSettings()
	settings.General.EnsureValidity()
	if len(d.deepSearchableCheckbox) == 0 || len(d.openInWindowCheckbox) == 0 {
		t.Fatal("expected deep search and separate window checkboxes")
	}
	deepSearch, separateWindow := d.deepSearchableCheckbox[0], d.openInWindowCheckbox[0]
	settings.DeepSearch = []string{deepSearch.item}
	settings.OpenInWindow = []dgroup.Group{separateWindow.item}
	languageSetting = "fr"
	d.sync()
	c.Equal(uncheck.On, deepSearch.box.State, "precondition: the deep search checkbox shows its setting")
	c.Equal(uncheck.On, separateWindow.box.State, "precondition: the separate window checkbox shows its setting")
	c.Equal("fr", d.localeField.Text(), "precondition: the locale field shows its setting")
	c.True(Workspace.Navigator.deepSearch[sentinelDeepSearchKey], "precondition: the deep search map is the sentinel")

	d.reset()
	c.Equal(0, len(settings.DeepSearch), "the reset clears the deep search setting")
	c.Equal(uncheck.Off, deepSearch.box.State, "and its checkbox shows that")
	c.Equal(0, len(settings.OpenInWindow), "the reset clears the separate window setting")
	c.Equal(uncheck.Off, separateWindow.box.State, "and its checkbox shows that")
	c.Equal("", languageSetting, "the reset clears the locale")
	c.Equal("", d.localeField.Text(), "and its field shows that")
	c.Equal(map[string]bool{}, Workspace.Navigator.deepSearch,
		"the reset rebuilds the deep search map from the cleared setting")
}

// TestGeneralSettingsLoadCopiesIntoTheLiveSettings verifies that loading a settings file copies it into the live
// general settings, which the widgets hold pointers into, rather than replacing them, and brings the widgets into line.
func TestGeneralSettingsLoadCopiesIntoTheLiveSettings(t *testing.T) {
	c := check.New(t)
	d, _ := newTestGeneralSettingsDockable(t)
	gs := gurps.GlobalSettings().General
	*gs = *gurps.NewGeneralSettings()
	gs.EnsureValidity()
	d.sync()
	var box settingCheckBox
	for _, one := range d.checkBoxes {
		if one.value == &gs.AutoFillProfile {
			box = one
			break
		}
	}
	if box.box == nil {
		t.Fatal("expected a checkbox for the initial description setting")
	}
	file := gurps.NewGeneralSettings()
	file.DefaultPlayerName = "Loaded Player"
	file.AutoFillProfile = !gs.AutoFillProfile
	dir := t.TempDir()
	name := "General" + gurps.GeneralSettingsExt
	c.NoError(file.Save(filepath.Join(dir, name)))

	c.NoError(d.load(os.DirFS(dir), name))
	c.Equal("Loaded Player", d.nameField.Text(), "the name field shows the loaded setting")
	c.Equal(uncheck.FromBool(file.AutoFillProfile), box.box.State, "the checkbox shows the loaded setting")
	box.box.State = uncheck.FromBool(!file.AutoFillProfile)
	box.box.ClickCallback()
	c.Equal(!file.AutoFillProfile, gurps.GlobalSettings().General.AutoFillProfile,
		"the checkbox still edits the live settings")
}
