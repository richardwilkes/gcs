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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestMonitorPPIForDisplay verifies that deriving the monitor PPI is robust against a missing display. On some Linux
// configurations unison.PrimaryDisplay() can return nil when no monitor is enumerated; previously that caused a nil
// dereference (and, when reached from an unrecovered background goroutine such as the markdown image loader, an
// unlogged process crash). It also guards against a zero content scale, which would otherwise divide by zero.
func TestMonitorPPIForDisplay(t *testing.T) {
	c := check.New(t)

	// A nil display must not panic and must fall back to the default.
	c.Equal(108, monitorPPIForDisplay(nil))

	// A display reporting a zero content scale must not divide by zero; it falls back to the default.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.Point{}}))

	// A display that computes a non-positive PPI falls back to the default.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 0, Scale: geom.NewPoint(2, 2)}))

	// A normal display yields its scaled PPI.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.NewPoint(2, 2)}))
	c.Equal(216, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.NewPoint(1, 1)}))
}

// TestMonitorPPIUsesSettingOverride verifies that an explicit monitor resolution setting is honored and doesn't touch
// the display at all.
func TestMonitorPPIUsesSettingOverride(t *testing.T) {
	s := &GeneralSettings{MonitorResolution: 150}
	check.New(t).Equal(150, s.MonitorPPI())
}

// TestCursorSizeValidation verifies that the cursor size setting is kept within the range unison permits, that the
// zero value found in settings files written before the setting existed is replaced with the default, and that
// validation pushes the resulting size to unison.
func TestCursorSizeValidation(t *testing.T) {
	c := check.New(t)

	savedSize := unison.CursorSize()
	defer unison.SetCursorSize(savedSize)

	c.Equal(int(unison.DefaultCursorSize().Width), CursorSizeDef, "the default tracks unison's default cursor size")

	s := NewGeneralSettings()
	c.Equal(CursorSizeDef, s.CursorSize, "new settings start at the default cursor size")

	s.CursorSize = 0 // settings files from before the setting existed load as zero
	s.EnsureValidity()
	c.Equal(CursorSizeDef, s.CursorSize, "a missing cursor size is reset to the default")

	s.CursorSize = CursorSizeMax + 1
	s.EnsureValidity()
	c.Equal(CursorSizeDef, s.CursorSize, "an out-of-range cursor size is reset to the default")

	s.CursorSize = CursorSizeMin
	s.EnsureValidity()
	c.Equal(CursorSizeMin, s.CursorSize, "an in-range cursor size is preserved")
	c.Equal(geom.NewSize(CursorSizeMin, CursorSizeMin), unison.CursorSize(), "validation applies the size to unison")
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
