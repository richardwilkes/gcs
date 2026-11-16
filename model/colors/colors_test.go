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

	"github.com/richardwilkes/gcs/v5/model/jio"
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

// colorsOf builds a Colors that defines just the given colors, as loading a partially specified .colors file would.
func colorsOf(t *testing.T, m map[string]unison.ThemeColor) *Colors {
	t.Helper()
	c := check.New(t)
	data, err := jio.Marshal(m)
	c.NoError(err)
	var result Colors
	c.NoError(jio.Unmarshal(data, &result))
	return &result
}

// stored returns what the Colors would write.
func stored(t *testing.T, colors *Colors) map[string]unison.ThemeColor {
	t.Helper()
	c := check.New(t)
	data, err := jio.Marshal(colors)
	c.NoError(err)
	var m map[string]unison.ThemeColor
	c.NoError(jio.Unmarshal(data, &m))
	return m
}

// TestSaveWritesReceiverRatherThanLiveTheme covers the `gcs --convert` path: a .colors file is loaded and written back
// out, and must keep its own colors instead of being overwritten with whatever theme the app is currently running.
func TestSaveWritesReceiverRatherThanLiveTheme(t *testing.T) {
	c := check.New(t)
	preserveLiveTheme(t)

	red := unison.RGB(255, 0, 0)
	green := unison.RGB(0, 255, 0)

	// A theme file whose surface color is red, written out literally to pin the form existing files have.
	dir := t.TempDir()
	p := filepath.Join(dir, "test.colors")
	c.NoError(os.WriteFile(p, []byte(`{"version":5,"colors":{"surface":{"light":"#FF0000","dark":"#FF0000"}}}`), 0o600))

	// The running app's theme is something else entirely.
	for _, one := range Current() {
		one.Color.Light = green
		one.Color.Dark = green
	}

	loaded, err := NewFromFS(os.DirFS(dir), "test.colors")
	c.NoError(err)
	c.Equal(red, stored(t, loaded)["surface"].Light, "the file's color is loaded, not the live one")

	// This is what --convert does: load the file, then write it back to the same path.
	c.NoError(loaded.Save(p))

	reloaded, err := NewFromFS(os.DirFS(dir), "test.colors")
	c.NoError(err)
	m := stored(t, reloaded)
	c.Equal(red, m["surface"].Light, "converting preserves the file's light color")
	c.Equal(red, m["surface"].Dark, "converting preserves the file's dark color")
}

// TestMakeCurrentAppliesToLiveTheme verifies that loaded colors reach the live ThemeColors the UI draws with.
func TestMakeCurrentAppliesToLiveTheme(t *testing.T) {
	c := check.New(t)
	preserveLiveTheme(t)

	purple := unison.RGB(128, 0, 128)
	orange := unison.RGB(255, 128, 0)
	first := Current()[0]
	second := Current()[1]
	want := *Factory()[1].Color // Taken before the live edit, so a factory that aliases the live colors is caught.
	second.Color.Light = orange // A live edit that the file, which doesn't mention this color, will replace.
	loaded := colorsOf(t, map[string]unison.ThemeColor{first.ID: {Light: purple, Dark: purple}})
	loaded.MakeCurrent()
	c.Equal(purple, first.Color.Light, "the loaded light color is applied")
	c.Equal(purple, first.Color.Dark, "the loaded dark color is applied")
	c.Equal(want, *second.Color, "colors the file didn't mention are set to their factory values")
}

// TestNewFromFSRefusesOldFormats verifies that theme color files from before the theme rework are refused rather than
// silently loaded as a mostly factory theme.
func TestNewFromFSRefusesOldFormats(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, "old.colors"), []byte(`{"version":4,"colors":{}}`), 0o600))
	_, err := NewFromFS(os.DirFS(dir), "old.colors")
	c.HasError(err)

	c.NoError(os.WriteFile(filepath.Join(dir, "new.colors"), []byte(`{"version":5,"colors":{}}`), 0o600))
	_, err = NewFromFS(os.DirFS(dir), "new.colors")
	c.NoError(err)
}
