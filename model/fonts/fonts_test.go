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
	"fmt"
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

	// A font file whose first font is 5 points larger than the factory default, written out literally to pin the form
	// existing files have.
	text, err := jio.Marshal(larger)
	c.NoError(err)
	dir := t.TempDir()
	p := filepath.Join(dir, "test.fonts")
	c.NoError(os.WriteFile(p, fmt.Appendf(nil, `{"version":5,"fonts":{%q:%s}}`, id, text), 0o600))

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
	c.Equal(larger, stored(t, reloaded)[id], "converting preserves the file's font")
}

// TestMakeCurrentAppliesToLiveFonts verifies that loaded fonts reach the live IndirectFonts the UI draws with.
func TestMakeCurrentAppliesToLiveFonts(t *testing.T) {
	c := check.New(t)
	preserveLiveFonts(t)

	first := CurrentFonts()[0]
	second := CurrentFonts()[1]
	larger := first.Font.Descriptor()
	larger.Size += 7
	// Taken before the live edit, so a factory that aliases the live fonts is caught.
	want := FactoryFonts()[1].Font.Descriptor()
	edited := second.Font.Descriptor()
	edited.Size += 2
	second.Font.Font = edited.Font() // A live edit that the file, which doesn't mention this font, will replace.
	loaded := fontsOf(t, map[string]unison.FontDescriptor{first.ID: larger})
	loaded.MakeCurrent()
	c.Equal(larger, first.Font.Descriptor(), "the loaded font is applied")
	c.Equal(want, second.Font.Descriptor(), "fonts the file didn't mention are set to their factory values")
}
