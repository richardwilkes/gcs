// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fonts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// preserveLiveFonts restores the live theme fonts when the test finishes, since these tests deliberately alter them to
// prove that serialization doesn't depend on them.
func preserveLiveFonts(t *testing.T) {
	t.Helper()
	saved := make(map[string]unison.Font, len(CurrentFonts()))
	for _, one := range CurrentFonts() {
		saved[one.ID] = one.Font.Font
	}
	t.Cleanup(func() {
		for _, one := range CurrentFonts() {
			one.Font.Font = saved[one.ID]
		}
	})
}

// fontsOf builds a Fonts that defines just the given fonts, as loading a partially specified .fonts file would.
func fontsOf(t *testing.T, m map[string]unison.FontDescriptor) *Fonts {
	t.Helper()
	c := check.New(t)
	data, err := jio.Marshal(m)
	c.NoError(err)
	var result Fonts
	c.NoError(jio.Unmarshal(data, &result))
	return &result
}

// stored returns what the Fonts would write.
func stored(t *testing.T, fonts *Fonts) map[string]unison.FontDescriptor {
	t.Helper()
	c := check.New(t)
	data, err := jio.Marshal(fonts)
	c.NoError(err)
	var m map[string]unison.FontDescriptor
	c.NoError(jio.Unmarshal(data, &m))
	return m
}

// TestSaveWritesReceiverRatherThanLiveTheme covers the `gcs --convert` path: a .fonts file is loaded and written back
// out, and must keep its own fonts instead of being overwritten with whatever fonts the app is currently using.
func TestSaveWritesReceiverRatherThanLiveTheme(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	ff := FactoryFonts()
	id := ff[0].ID
	larger := ff[0].Font.Descriptor()
	larger.Size += 5

	// A font file whose first font is 5 points larger than the factory default.
	dir := t.TempDir()
	p := filepath.Join(dir, "test.fonts")
	c.NoError(fontsOf(t, map[string]unison.FontDescriptor{id: larger}).Save(p))

	// The running app's fonts are something else entirely.
	for _, one := range CurrentFonts() {
		desc := one.Font.Descriptor()
		desc.Size += 11
		one.Font.Font = desc.Font()
	}

	loaded, err := NewFromFS(os.DirFS(dir), "test.fonts")
	c.NoError(err)
	c.Equal(larger, stored(t, loaded)[id], "the file's font is loaded, not the live one")

	// This is what --convert does: load the file, then write it back to the same path.
	c.NoError(loaded.Save(p))

	reloaded, err := NewFromFS(os.DirFS(dir), "test.fonts")
	c.NoError(err)
	m := stored(t, reloaded)
	c.Equal(larger, m[id], "converting preserves the file's font")

	// Fonts the file never mentioned fall back to the factory values, not the live ones.
	c.Equal(ff[1].Font.Descriptor(), m[ff[1].ID], "an unspecified font is written as its factory value")
}

// TestCaptureCurrentRecordsLiveEdits verifies the other half of the contract: the settings UI edits the live
// IndirectFonts in place, so the global settings must be able to pull those edits in before saving or they'd be lost.
func TestCaptureCurrentRecordsLiveEdits(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	id := CurrentFonts()[0].ID
	edited := CurrentFonts()[0].Font.Descriptor()
	edited.Size += 3
	CurrentFonts()[0].Font.Font = edited.Font() // As the font panel's FontModifiedCallback does.

	var saved Fonts // A zero value, as first-run settings hold.
	saved.CaptureCurrent()
	c.Equal(edited, stored(t, &saved)[id], "capturing records the live edit")

	// And a capture survives the round trip to disk.
	dir := t.TempDir()
	p := filepath.Join(dir, "captured.fonts")
	c.NoError(saved.Save(p))
	reloaded, err := NewFromFS(os.DirFS(dir), "captured.fonts")
	c.NoError(err)
	c.Equal(edited, stored(t, reloaded)[id], "the captured edit round-trips through a save")
}

// TestMakeCurrentAppliesToLiveFonts verifies that loaded fonts reach the live IndirectFonts the UI draws with.
func TestMakeCurrentAppliesToLiveFonts(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	first := CurrentFonts()[0]
	larger := first.Font.Descriptor()
	larger.Size += 7
	loaded := fontsOf(t, map[string]unison.FontDescriptor{first.ID: larger})
	loaded.MakeCurrent()
	c.Equal(larger, first.Font.Descriptor(), "the loaded font is applied")
	c.Equal(FactoryFonts()[1].Font.Descriptor(), CurrentFonts()[1].Font.Descriptor(),
		"fonts the file didn't mention are set to their factory values")
}

// TestResetWithNoLoadedData verifies that resetting a Fonts that has never been through UnmarshalJSONFrom -- the
// first-run case, where no settings file exists and the map was therefore never allocated -- populates the factory
// values rather than panicking with "assignment to entry in nil map".
func TestResetWithNoLoadedData(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	// Alter the live fonts so that a reset that mistakenly captured them would be caught.
	for _, one := range CurrentFonts() {
		desc := one.Font.Descriptor()
		desc.Size += 11
		one.Font.Font = desc.Font()
	}

	var f Fonts
	f.Reset()
	c.NotPanics(f.MakeCurrent)
	for _, one := range FactoryFonts() {
		c.Equal(one.Font.Descriptor(), stored(t, &f)[one.ID], "reset uses the factory descriptor for "+one.ID)
	}

	var one Fonts
	one.ResetOne(FactoryFonts()[0].ID)
	c.Equal(FactoryFonts()[0].Font.Descriptor(), stored(t, &one)[FactoryFonts()[0].ID],
		"reset of a single font uses the factory descriptor")
}
