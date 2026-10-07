// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMonitorPPIFromHost verifies that the monitor PPI comes from the host and falls back to the default when the host
// has no usable value. The host reports 0 when there is no display, which happens on some Linux configurations.
func TestMonitorPPIFromHost(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)
	s := &GeneralSettings{}

	c.Equal(defaultMonitorPPI, s.MonitorPPI(), "an unknown PPI falls back to the default")

	h.ppi = -72
	c.Equal(defaultMonitorPPI, s.MonitorPPI(), "a non-positive PPI falls back to the default")

	h.ppi = 216
	c.Equal(216, s.MonitorPPI(), "a usable PPI is passed through")
}

// TestMonitorPPIUsesSettingOverride verifies that an explicit monitor resolution setting is honored and doesn't ask
// the host at all.
func TestMonitorPPIUsesSettingOverride(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)
	h.ppi = 216
	s := &GeneralSettings{MonitorResolution: 150}
	c.Equal(150, s.MonitorPPI())
	c.Equal(0, h.ppiCalls, "the host wasn't asked")
}

// TestCursorSizeValidation verifies that the cursor size setting is kept within the permitted range, that the zero
// value found in settings files written before the setting existed is replaced with the default, and that validation
// hands the resulting size to the host.
func TestCursorSizeValidation(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)

	s := NewGeneralSettings()
	c.Equal(CursorSizeDef, s.CursorSize, "new settings start at the default cursor size")

	s.CursorSize = 0 // settings files from before the setting existed load as zero
	s.EnsureValidity()
	c.Equal(CursorSizeDef, s.CursorSize, "a missing cursor size is reset to the default")
	c.Equal(CursorSizeDef, h.cursorSize, "validation hands the size to the host")

	s.CursorSize = CursorSizeMax + 1
	s.EnsureValidity()
	c.Equal(CursorSizeDef, s.CursorSize, "an out-of-range cursor size is reset to the default")

	s.CursorSize = CursorSizeMin
	s.EnsureValidity()
	c.Equal(CursorSizeMin, s.CursorSize, "an in-range cursor size is preserved")
	c.Equal(CursorSizeMin, h.cursorSize, "validation hands the size to the host")
}

// TestToolTipTimingValidation verifies that validation hands the tooltip timing to the host.
func TestToolTipTimingValidation(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)
	s := NewGeneralSettings()
	s.TooltipDelay = fxp.Two
	s.TooltipDismissal = fxp.Ten
	s.EnsureValidity()
	c.Equal(2*time.Second, h.tooltipDelay)
	c.Equal(10*time.Second, h.tooltipDismissal)
}

// TestUpdateCheckSettings verifies the app and library update check settings: new settings start at the default, an
// out-of-range value is reset rather than left to produce nonsense, a settings file written before these settings
// existed loads as the default, chosen values survive a save/load round trip, and the default doesn't bloat the saved
// file.
func TestUpdateCheckSettings(t *testing.T) {
	c := check.New(t)

	s := NewGeneralSettings()
	c.Equal(updatecheck.AtLaunch, s.AppUpdateCheck, "new settings check for app updates at launch")
	c.Equal(updatecheck.AtLaunch, s.LibraryUpdateCheck, "new settings check for library updates at launch")

	s.AppUpdateCheck = updatecheck.Option(200)
	s.LibraryUpdateCheck = updatecheck.Option(200)
	s.EnsureValidity()
	c.Equal(updatecheck.AtLaunch, s.AppUpdateCheck, "an out-of-range app update check is reset to the default")
	c.Equal(updatecheck.AtLaunch, s.LibraryUpdateCheck, "an out-of-range library update check is reset to the default")

	// A settings file written before these settings existed has no such keys, so they load as the default.
	p := filepath.Join(t.TempDir(), "general.json")
	c.NoError(os.WriteFile(p, []byte(`{"version":2}`), 0o600))
	loaded, err := NewGeneralSettingsFromFile(nil, p)
	c.NoError(err)
	c.Equal(updatecheck.AtLaunch, loaded.AppUpdateCheck,
		"a pre-existing settings file checks for app updates at launch")
	c.Equal(updatecheck.AtLaunch, loaded.LibraryUpdateCheck,
		"a pre-existing settings file checks for library updates at launch")

	// Chosen values survive a round trip through the file.
	s.AppUpdateCheck = updatecheck.Daily
	s.LibraryUpdateCheck = updatecheck.Never
	c.NoError(s.Save(p))
	loaded, err = NewGeneralSettingsFromFile(nil, p)
	c.NoError(err)
	c.Equal(updatecheck.Daily, loaded.AppUpdateCheck, "the app update check survives a save and load")
	c.Equal(updatecheck.Never, loaded.LibraryUpdateCheck, "the library update check survives a save and load")

	// The default is omitted from the saved file.
	c.NoError(NewGeneralSettings().Save(p))
	data, err := os.ReadFile(p)
	c.NoError(err)
	c.NotContains(string(data), "app_update_check", "the default app update check isn't written")
	c.NotContains(string(data), "library_update_check", "the default library update check isn't written")
}

func TestFocusForReadingSetting(t *testing.T) {
	c := check.New(t)

	h := installRecordingHost(t)

	s := NewGeneralSettings()
	c.False(s.FocusForReading, "new settings leave static text and disabled controls out of the tab order")

	p := filepath.Join(t.TempDir(), "general.json")
	c.NoError(os.WriteFile(p, []byte(`{"version":2}`), 0o600))
	// Turned on first, so that the load is what turns it off.
	h.focusForReading = true
	loaded, err := NewGeneralSettingsFromFile(nil, p)
	c.NoError(err)
	c.False(loaded.FocusForReading, "a pre-existing settings file leaves them out of the tab order")
	c.False(h.focusForReading, "loading hands the setting to the host")

	s.FocusForReading = true
	s.EnsureValidity()
	c.True(h.focusForReading, "validation hands the setting to the host")

	c.NoError(s.Save(p))
	h.focusForReading = false
	loaded, err = NewGeneralSettingsFromFile(nil, p)
	c.NoError(err)
	c.True(loaded.FocusForReading, "the choice survives a save and load")
	c.True(h.focusForReading, "loading hands the choice to the host")

	c.NoError(NewGeneralSettings().Save(p))
	data, err := os.ReadFile(p)
	c.NoError(err)
	c.NotContains(string(data), "focus_for_reading", "the default isn't written")
}
