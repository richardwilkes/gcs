// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package colors

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// preserveLiveTheme restores the live theme colors when the test finishes, since these tests deliberately alter them to
// prove that serialization doesn't depend on them.
func preserveLiveTheme(t *testing.T) {
	t.Helper()
	saved := make(map[string]unison.ThemeColor, len(Current()))
	for _, one := range Current() {
		saved[one.ID] = *one.Color
	}
	t.Cleanup(func() {
		for _, one := range Current() {
			*one.Color = saved[one.ID]
		}
	})
}

// TestSaveWritesReceiverRatherThanLiveTheme covers the `gcs --convert` path: a .colors file is loaded and written back
// out, and must keep its own colors instead of being overwritten with whatever theme the app is currently running.
func TestSaveWritesReceiverRatherThanLiveTheme(t *testing.T) {
	c := check.New(t)
	preserveLiveTheme(t)

	red := unison.RGB(255, 0, 0)
	green := unison.RGB(0, 255, 0)

	// A theme file whose surface color is red.
	dir := t.TempDir()
	p := filepath.Join(dir, "test.colors")
	c.NoError((&Colors{data: map[string]*unison.ThemeColor{"surface": {Light: red, Dark: red}}}).Save(p))

	// The running app's theme is something else entirely.
	for _, one := range Current() {
		one.Color.Light = green
		one.Color.Dark = green
	}

	loaded, err := NewFromFS(os.DirFS(dir), "test.colors")
	c.NoError(err)
	c.Equal(red, loaded.data["surface"].Light, "the file's color is loaded, not the live one")

	// This is what --convert does: load the file, then write it back to the same path.
	c.NoError(loaded.Save(p))

	reloaded, err := NewFromFS(os.DirFS(dir), "test.colors")
	c.NoError(err)
	c.Equal(red, reloaded.data["surface"].Light, "converting preserves the file's light color")
	c.Equal(red, reloaded.data["surface"].Dark, "converting preserves the file's dark color")

	// Colors the file never mentioned fall back to the factory values, not the live theme.
	c.Equal(Factory()[1].Color.Light, reloaded.data[Factory()[1].ID].Light,
		"an unspecified color is written as its factory value")
}

// TestCaptureCurrentRecordsLiveEdits verifies the other half of the contract: the settings UI edits the live
// ThemeColors in place, so the global settings must be able to pull those edits in before saving or they'd be lost.
func TestCaptureCurrentRecordsLiveEdits(t *testing.T) {
	c := check.New(t)
	preserveLiveTheme(t)

	blue := unison.RGB(0, 0, 255)
	id := Current()[0].ID
	Current()[0].Color.Light = blue // As the color well's InkChangedCallback does.

	var saved Colors // A zero value, as first-run settings hold.
	saved.CaptureCurrent()
	c.Equal(blue, saved.data[id].Light, "capturing records the live edit")
	c.Equal(len(Current()), len(saved.data), "capturing records every color")

	// The captured value is a copy, so later live edits don't retroactively change what will be written.
	Current()[0].Color.Light = unison.RGB(1, 2, 3)
	c.Equal(blue, saved.data[id].Light, "the captured color is a copy, not an alias of the live color")

	// And a capture survives the round trip to disk.
	dir := t.TempDir()
	p := filepath.Join(dir, "captured.colors")
	c.NoError(saved.Save(p))
	reloaded, err := NewFromFS(os.DirFS(dir), "captured.colors")
	c.NoError(err)
	c.Equal(blue, reloaded.data[id].Light, "the captured edit round-trips through a save")
}

// TestCaptureCurrentOverwritesStaleData verifies that capturing replaces previously loaded values rather than only
// filling in absent ones.
func TestCaptureCurrentOverwritesStaleData(t *testing.T) {
	c := check.New(t)
	preserveLiveTheme(t)

	id := Current()[0].ID
	stale := unison.RGB(9, 9, 9)
	live := unison.RGB(7, 7, 7)
	loaded := Colors{data: map[string]*unison.ThemeColor{id: {Light: stale, Dark: stale}}}
	Current()[0].Color.Light = live

	loaded.CaptureCurrent()
	c.Equal(live, loaded.data[id].Light, "capturing replaces the stale loaded color with the live one")
}
