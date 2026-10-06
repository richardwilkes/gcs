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
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/richardwilkes/toolbox/v2/xos"
)

var (
	// host holds the current Host. A nil pointer means NoHost. It is read from any goroutine, so it is only ever
	// replaced atomically.
	host atomic.Pointer[Host]
	// hostLock serializes SetHost with the hand-over of the global settings in GlobalSettings, so that a host installed
	// while they are being loaded is still given them exactly once.
	hostLock sync.Mutex
	// globalSettingsLoaded is set once GlobalSettings has loaded the settings and handed them to the host.
	globalSettingsLoaded atomic.Bool
)

// Host is implemented by the user interface. The model hands it the settings that alter how the interface behaves and
// asks it for what only the interface knows, which keeps the model free of any dependency on a UI toolkit.
type Host interface {
	// SetTooltipTiming sets how long the pointer must rest before a tooltip is shown and how long it then stays up.
	SetTooltipTiming(delay, dismissal time.Duration)
	// SetCursorSize sets the size, in logical points, that cursors are built at.
	SetCursorSize(size int)
	// SetFocusForReading sets whether static text and disabled controls are part of the tab order.
	SetFocusForReading(enabled bool)
	// DefaultScrollWheelMultiplier returns the multiplier applied to mouse wheel deltas when the user hasn't chosen one.
	DefaultScrollWheelMultiplier() float32
	// PrimaryDisplayPPI returns the pixels per inch of the primary display divided by its content scale, or 0 if that
	// isn't known. It is only called on the UI thread, since a toolkit typically permits its displays to be queried
	// from no other.
	PrimaryDisplayPPI() int
	// ThemeColor returns the current value of the theme color with the given ID, in CSS form.
	ThemeColor(id string) (css string, ok bool)
	// SettingsLoaded is called once the global settings have been loaded and validated, so that the theme mode, colors
	// and fonts they hold can be applied. It must not call GlobalSettings.
	SettingsLoaded(s *Settings)
	// SettingsSaving is called before the global settings are written, so that the live theme colors and fonts can be
	// stored into them.
	SettingsSaving(s *Settings)
}

// SetHost sets the Host the model works with. Passing nil restores NoHost. If the global settings have already been
// loaded, they are handed to the new host, which puts the theme they hold into effect in place of whatever the live
// theme has become since. It may be called at any time, from any goroutine.
func SetHost(h Host) {
	if h == nil {
		h = NoHost{}
	}
	hostLock.Lock()
	defer hostLock.Unlock()
	host.Store(&h)
	if globalSettingsLoaded.Load() {
		globalSettings.General.UpdateToolTipTiming()
		globalSettings.General.UpdateCursorSize()
		globalSettings.General.UpdateFocusForReading()
		h.SettingsLoaded(&globalSettings)
	}
}

// currentHost returns the Host the model works with.
func currentHost() Host {
	if h := host.Load(); h != nil {
		return *h
	}
	return NoHost{}
}

// NoHost is the Host used when there is no user interface. It applies nothing and knows nothing.
type NoHost struct{}

// SetTooltipTiming implements Host.
func (NoHost) SetTooltipTiming(_, _ time.Duration) {}

// SetCursorSize implements Host.
func (NoHost) SetCursorSize(_ int) {}

// SetFocusForReading implements Host.
func (NoHost) SetFocusForReading(_ bool) {}

// DefaultScrollWheelMultiplier implements Host. It returns the values the user interface uses, so that settings first
// written without one suit it.
func (NoHost) DefaultScrollWheelMultiplier() float32 {
	if runtime.GOOS == xos.MacOS {
		return 8
	}
	return 24
}

// PrimaryDisplayPPI implements Host.
func (NoHost) PrimaryDisplayPPI() int { return 0 }

// ThemeColor implements Host.
func (NoHost) ThemeColor(_ string) (css string, ok bool) { return "", false }

// SettingsLoaded implements Host.
func (NoHost) SettingsLoaded(_ *Settings) {}

// SettingsSaving implements Host.
func (NoHost) SettingsSaving(_ *Settings) {}
