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
	"encoding/json/jsontext"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/thememode"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/colors"
	"github.com/richardwilkes/gcs/v5/ux/fonts"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	unthememode "github.com/richardwilkes/unison/enums/thememode"
)

// The model knows nothing of unison, so everything it needs from the user interface is installed here, before anything
// can ask for the global settings.
func init() {
	gurps.SetHost(modelHost{})
	gurps.RegisterConverter(gurps.ColorSettingsExt, colors.NewFromFS, (*colors.Colors).Save)
	gurps.RegisterConverter(gurps.FontSettingsExt, fonts.NewFromFS, (*fonts.Fonts).Save)
}

// modelHost is the gurps.Host that connects the model to unison.
type modelHost struct{}

// SetTooltipTiming implements gurps.Host.
func (modelHost) SetTooltipTiming(delay, dismissal time.Duration) {
	unison.DefaultTooltipTheme.Delay = delay
	unison.DefaultTooltipTheme.Dismissal = dismissal
}

// SetCursorSize implements gurps.Host.
func (modelHost) SetCursorSize(size int) {
	unison.SetCursorSize(geom.NewSize(float32(size), float32(size)))
}

// SetFocusForReading implements gurps.Host.
func (modelHost) SetFocusForReading(enabled bool) {
	unison.SetFocusForReading(enabled)
}

// DefaultScrollWheelMultiplier implements gurps.Host.
func (modelHost) DefaultScrollWheelMultiplier() float32 {
	return unison.MouseWheelMultiplier
}

// PrimaryDisplayPPI implements gurps.Host.
func (modelHost) PrimaryDisplayPPI() int {
	return displayPPI(unison.PrimaryDisplay())
}

// displayPPI returns the display's pixels per inch divided by its content scale, or 0 when the display is unavailable
// (on some Linux configurations PrimaryDisplay() returns nil when no monitor is enumerated) or has no content scale.
func displayPPI(d *unison.Display) int {
	if d == nil || d.Scale.X == 0 {
		return 0
	}
	return int(float32(d.PPI) / d.Scale.X)
}

// ThemeColor implements gurps.Host.
func (modelHost) ThemeColor(id string) (css string, ok bool) {
	for _, c := range colors.Current() {
		if c.ID == id {
			return c.Color.GetColor().String(), true
		}
	}
	return "", false
}

// SettingsLoaded implements gurps.Host.
func (modelHost) SettingsLoaded(s *gurps.Settings) {
	unison.SetThemeMode(unisonThemeMode(s.ThemeMode))
	var c colors.Colors
	loadThemeSet(s.Colors, &c)
	c.MakeCurrent()
	var f fonts.Fonts
	loadThemeSet(s.Fonts, &f)
	f.MakeCurrent()
	unison.DefaultScrollPanelTheme.MouseWheelMultiplier = func() float32 {
		return s.General.ScrollWheelMultiplier.AsFloat[float32]()
	}
	unison.DefaultFieldTheme.InitialClickSelectsAll = func(_ *unison.Field) bool {
		return s.General.InitialFieldClickSelectsAll
	}
}

// SettingsSaving implements gurps.Host.
func (modelHost) SettingsSaving(s *gurps.Settings) {
	var c colors.Colors
	c.CaptureCurrent()
	s.Colors = storeThemeSet(&c, s.Colors)
	var f fonts.Fonts
	f.CaptureCurrent()
	s.Fonts = storeThemeSet(&f, s.Fonts)
}

// loadThemeSet decodes a set of theme values the settings carried as raw JSON. Each value that can't be decoded is
// logged and left at its factory value, which the next save then stores in its place; the rest are kept.
func loadThemeSet(data jsontext.Value, set any) {
	if len(data) != 0 {
		if err := jio.Unmarshal(data, set); err != nil {
			errs.Log(errs.NewWithCause("unable to load theme settings", err))
		}
	}
}

// storeThemeSet encodes a set of theme values as raw JSON for the settings to carry, returning existing if that can't
// be done.
func storeThemeSet(set any, existing jsontext.Value) jsontext.Value {
	data, err := jio.Marshal(set)
	if err != nil {
		errs.Log(errs.NewWithCause("unable to store theme settings", err))
		return existing
	}
	return data
}

// unisonThemeMode returns unison's equivalent of the model's theme mode.
func unisonThemeMode(mode thememode.Mode) unthememode.Enum {
	switch mode {
	case thememode.Dark:
		return unthememode.Dark
	case thememode.Light:
		return unthememode.Light
	default:
		return unthememode.Auto
	}
}
