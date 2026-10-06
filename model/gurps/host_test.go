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
	"encoding/json/jsontext"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/thememode"
	"github.com/richardwilkes/toolbox/v2/check"
)

// recordingHost is a Host that records what the model hands it and answers with whatever a test sets.
type recordingHost struct {
	NoHost
	themeColors      map[string]string
	onSaving         func(s *Settings)
	loaded           []*Settings
	tooltipDelay     time.Duration
	tooltipDismissal time.Duration
	cursorSize       int
	ppi              int
	ppiCalls         int
	focusForReading  bool
}

func (h *recordingHost) SetTooltipTiming(delay, dismissal time.Duration) {
	h.tooltipDelay = delay
	h.tooltipDismissal = dismissal
}

func (h *recordingHost) SetCursorSize(size int) { h.cursorSize = size }

func (h *recordingHost) SetFocusForReading(enabled bool) { h.focusForReading = enabled }

func (h *recordingHost) PrimaryDisplayPPI() int {
	h.ppiCalls++
	return h.ppi
}

func (h *recordingHost) ThemeColor(id string) (css string, ok bool) {
	css, ok = h.themeColors[id]
	return css, ok
}

func (h *recordingHost) SettingsLoaded(s *Settings) { h.loaded = append(h.loaded, s) }

func (h *recordingHost) SettingsSaving(s *Settings) {
	if h.onSaving != nil {
		h.onSaving(s)
	}
}

// installRecordingHost makes a recordingHost the model's host for the rest of the test. It is put in place directly
// rather than through SetHost, so that it starts out having been handed nothing.
func installRecordingHost(t *testing.T) *recordingHost {
	t.Helper()
	h := &recordingHost{}
	saved := currentHost()
	var installed Host = h
	host.Store(&installed)
	t.Cleanup(func() { host.Store(&saved) })
	return h
}

// TestSetHostHandsOverLoadedSettings verifies that a host installed after the global settings were loaded is still
// given them, so that the order the two happen in doesn't matter.
func TestSetHostHandsOverLoadedSettings(t *testing.T) {
	c := check.New(t)
	saved := currentHost()
	t.Cleanup(func() { host.Store(&saved) })

	global := GlobalSettings()
	h := &recordingHost{}
	SetHost(h)
	c.Equal(1, len(h.loaded), "the loaded settings are handed to the new host")
	c.True(h.loaded[0] == global, "and they are the global settings")
	c.Equal(global.General.CursorSize, h.cursorSize, "along with the cursor size")
	c.Equal(fxp.SecondsToDuration(global.General.TooltipDelay), h.tooltipDelay, "the tooltip delay")
	c.Equal(fxp.SecondsToDuration(global.General.TooltipDismissal), h.tooltipDismissal, "the tooltip dismissal")
	c.Equal(global.General.FocusForReading, h.focusForReading, "and focus for reading")

	SetHost(nil)
	_, isDefault := currentHost().(NoHost)
	c.True(isDefault, "a nil host restores the default")
}

// TestHandSettingsToGivesEverything verifies that the hand-over GlobalSettings and SetHost share gives a host every
// value it keeps a copy of as well as the settings, so that a host installed while the settings were being loaded,
// after their validation had pushed those values to the host of that moment, is given them all the same.
func TestHandSettingsToGivesEverything(t *testing.T) {
	c := check.New(t)
	h := &recordingHost{}
	s := FactorySettings()
	s.General.CursorSize = CursorSizeMax
	s.General.TooltipDelay = fxp.Two
	s.General.TooltipDismissal = fxp.Ten
	s.General.FocusForReading = true
	handSettingsTo(h, &s)
	c.Equal(1, len(h.loaded))
	c.True(h.loaded[0] == &s)
	c.Equal(CursorSizeMax, h.cursorSize)
	c.Equal(fxp.SecondsToDuration(fxp.Two), h.tooltipDelay)
	c.Equal(fxp.SecondsToDuration(fxp.Ten), h.tooltipDismissal)
	c.True(h.focusForReading)
}

// TestSetHostIsConcurrencySafe verifies that the host can be replaced while other goroutines are using it, since the
// model asks it for things from whatever goroutine happens to need them.
func TestSetHostIsConcurrencySafe(t *testing.T) {
	c := check.New(t)
	saved := currentHost()
	t.Cleanup(func() { host.Store(&saved) })
	GlobalSettings() // So that each SetHost also hands the settings over.

	const workers = 4
	var wg sync.WaitGroup
	var handed atomic.Int64
	for range workers {
		wg.Go(func() {
			for range 50 {
				h := &recordingHost{}
				SetHost(h)
				handed.Add(int64(len(h.loaded)))
				NewGeneralSettings()
				_, ok := currentHost().ThemeColor("header")
				c.False(ok)
			}
		})
	}
	wg.Wait()
	c.Equal(int64(workers*50), handed.Load(), "every host installed after the load is handed the settings once")
}

// TestSettingsCarryHostOwnedData verifies that the parts of the settings only the host understands are collected from
// it when saving and survive a save and load untouched, including when there is no host to interpret them.
func TestSettingsCarryHostOwnedData(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)
	prevPath := SettingsPath
	SettingsPath = filepath.Join(t.TempDir(), "settings.json")
	t.Cleanup(func() { SettingsPath = prevPath })

	h.onSaving = func(s *Settings) {
		s.Colors = jsontext.Value(`{"surface":{"light":"#fff","dark":"#000"}}`)
		s.Fonts = jsontext.Value(`{"system":"Roboto 10 regular standard upright"}`)
	}
	settings := FactorySettings()
	settings.ThemeMode = thememode.Dark
	settings.TopDockState = jsontext.Value(`{"type":"layout","children":[{"type":"dockable","key":"nav"}]}`)
	c.NoError(settings.Save())

	data, err := os.ReadFile(SettingsPath)
	c.NoError(err)
	c.Contains(string(data), `"theme_mode": "dark"`)
	c.NotContains(string(data), "doc_dock_state", "a dock state that was never recorded isn't written")

	// Loaded without a host that could tell what any of it means.
	var none Host = NoHost{}
	host.Store(&none)
	reloaded := loadSettingsOrDefaults(SettingsPath)
	c.Equal(thememode.Dark, reloaded.ThemeMode)
	for _, one := range []struct {
		name string
		want jsontext.Value
		got  jsontext.Value
	}{
		{name: "theme colors", want: settings.Colors, got: reloaded.Colors},
		{name: "fonts", want: settings.Fonts, got: reloaded.Fonts},
		{name: "top dock state", want: settings.TopDockState, got: reloaded.TopDockState},
	} {
		c.Equal(string(one.want), compactJSON(c, one.got), one.name)
	}
	c.Equal(0, len(reloaded.DocDockState))

	// Saving again without a host must write back what was loaded rather than dropping it.
	c.NoError(reloaded.Save())
	again := loadSettingsOrDefaults(SettingsPath)
	c.Equal(string(settings.Colors), compactJSON(c, again.Colors), "the theme colors survive a save without a host")
}

// compactJSON returns the value with its insignificant whitespace removed, leaving the value itself untouched.
func compactJSON(c check.Checker, value jsontext.Value) string {
	value = slices.Clone(value)
	c.NoError(value.Compact())
	return string(value)
}

// TestLegacyExportThemeColor verifies that the COLOR_ keys of the legacy export are answered by the host, since the
// theme colors are its to know.
func TestLegacyExportThemeColor(t *testing.T) {
	c := check.New(t)
	h := installRecordingHost(t)
	h.themeColors = map[string]string{"header": "#505050"}
	e := NewEntity()
	c.Equal("#505050", runLegacyExport(t, c, e, "@COLOR_HEADER"))
	c.Contains(runLegacyExport(t, c, e, "@COLOR_BOGUS"), "Unidentified key")
}
