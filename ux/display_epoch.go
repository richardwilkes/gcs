// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import "sync/atomic"

// displayEpoch counts the changes reported to the data on display. Everything derived from the model for display --
// a table cell's data above all, which is derived once per change rather than on every draw and every sizing pass --
// is stamped with the epoch it was derived in and derived afresh once the epoch has moved on. The rule the stamp
// keeps is the one the display has always followed: the draw after a change is reported shows the change. Every root
// a change is reported to (the MarkModified and Rebuild of the sheets, templates, loot sheets, editors and library
// lists) moves the epoch on, as does the broadcast of a settings change, so a change that is reported is never shown
// stale, while scrolling, hovering and the blinking of a caret, which report nothing, draw from what was derived
// before. It is atomic only so that tests that build nodes off the UI thread can't race it.
var displayEpoch atomic.Uint64

// invalidateDisplayData notes that the data on display may have changed, so that everything derived from the model
// for display is derived afresh the next time it is wanted; see displayEpoch.
func invalidateDisplayData() {
	displayEpoch.Add(1)
}

// currentDisplayStamp returns the stamp that marks data derived from the model as current. It is never zero, so a
// zero stamp means the data has never been derived.
func currentDisplayStamp() uint64 {
	return displayEpoch.Load() + 1
}
