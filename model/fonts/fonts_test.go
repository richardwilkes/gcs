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

// TestSaveWritesReceiverRatherThanLiveTheme covers the `gcs --convert` path: a .fonts file is loaded and written back
// out, and must keep its own fonts instead of being overwritten with whatever fonts the app is currently using.
func TestSaveWritesReceiverRatherThanLiveTheme(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	ff := FactoryFonts()
	id := ff[0].ID
	stored := ff[0].Font.Descriptor()
	stored.Size += 5

	// A font file whose first font is 5 points larger than the factory default.
	dir := t.TempDir()
	p := filepath.Join(dir, "test.fonts")
	c.NoError((&Fonts{data: map[string]unison.FontDescriptor{id: stored}}).Save(p))

	// The running app's fonts are something else entirely.
	for _, one := range CurrentFonts() {
		desc := one.Font.Descriptor()
		desc.Size += 11
		one.Font.Font = desc.Font()
	}

	loaded, err := NewFromFS(os.DirFS(dir), "test.fonts")
	c.NoError(err)
	c.Equal(stored, loaded.data[id], "the file's font is loaded, not the live one")

	// This is what --convert does: load the file, then write it back to the same path.
	c.NoError(loaded.Save(p))

	reloaded, err := NewFromFS(os.DirFS(dir), "test.fonts")
	c.NoError(err)
	c.Equal(stored, reloaded.data[id], "converting preserves the file's font")

	// Fonts the file never mentioned fall back to the factory values, not the live ones.
	c.Equal(ff[1].Font.Descriptor(), reloaded.data[ff[1].ID], "an unspecified font is written as its factory value")
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
	c.Equal(edited, saved.data[id], "capturing records the live edit")
	c.Equal(len(CurrentFonts()), len(saved.data), "capturing records every font")

	// And a capture survives the round trip to disk.
	dir := t.TempDir()
	p := filepath.Join(dir, "captured.fonts")
	c.NoError(saved.Save(p))
	reloaded, err := NewFromFS(os.DirFS(dir), "captured.fonts")
	c.NoError(err)
	c.Equal(edited, reloaded.data[id], "the captured edit round-trips through a save")
}

// TestResetWithNoLoadedData verifies that resetting a Fonts that has never been through UnmarshalJSONFrom -- the
// first-run case, where no settings file exists and the map was therefore never allocated -- populates the factory
// values rather than panicking with "assignment to entry in nil map".
func TestResetWithNoLoadedData(t *testing.T) {
	c := check.New(t)

	var f Fonts
	c.Equal(0, len(f.data), "a fresh Fonts starts with no data")
	f.Reset()

	ff := FactoryFonts()
	c.Equal(len(ff), len(f.data), "reset fills in every factory font")
	for _, one := range ff {
		v, ok := f.data[one.ID]
		c.True(ok, "reset defines "+one.ID)
		c.Equal(one.Font.Descriptor(), v, "reset uses the factory descriptor for "+one.ID)
	}
}

// TestResetOneWithNoLoadedData verifies the same first-run case for ResetOne.
func TestResetOneWithNoLoadedData(t *testing.T) {
	c := check.New(t)

	id := FactoryFonts()[0].ID
	var f Fonts
	f.ResetOne(id)

	v, ok := f.data[id]
	c.True(ok, "reset of a single font defines it")
	c.Equal(FactoryFonts()[0].Font.Descriptor(), v, "reset of a single font uses the factory descriptor")

	// An unknown ID is simply ignored and must not panic on the nil map either.
	var other Fonts
	other.ResetOne("no.such.font")
	c.Equal(0, len(other.data), "resetting an unknown font adds nothing")
}

// TestResetRestoresModifiedFonts verifies that a Fonts holding user modifications is returned to the factory values,
// including entries that the loaded data never mentioned.
func TestResetRestoresModifiedFonts(t *testing.T) {
	c := check.New(t)

	ff := FactoryFonts()
	altered := ff[0].Font.Descriptor()
	altered.Size += 5
	f := Fonts{data: map[string]unison.FontDescriptor{
		ff[0].ID:        altered,
		"stale.font.id": altered,
	}}

	f.Reset()

	c.Equal(ff[0].Font.Descriptor(), f.data[ff[0].ID], "a modified font is restored to its factory descriptor")
	for _, one := range ff {
		v, ok := f.data[one.ID]
		c.True(ok, "reset defines "+one.ID)
		c.Equal(one.Font.Descriptor(), v, "reset uses the factory descriptor for "+one.ID)
	}
}

// TestResetOneRestoresOnlyTheNamedFont verifies that ResetOne leaves the other entries alone.
func TestResetOneRestoresOnlyTheNamedFont(t *testing.T) {
	c := check.New(t)

	ff := FactoryFonts()
	first := ff[0].Font.Descriptor()
	first.Size += 5
	second := ff[1].Font.Descriptor()
	second.Size += 5
	f := Fonts{data: map[string]unison.FontDescriptor{ff[0].ID: first, ff[1].ID: second}}

	f.ResetOne(ff[0].ID)

	c.Equal(ff[0].Font.Descriptor(), f.data[ff[0].ID], "the named font is restored")
	c.Equal(second, f.data[ff[1].ID], "the other fonts are left untouched")
}
