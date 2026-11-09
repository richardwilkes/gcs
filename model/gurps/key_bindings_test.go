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

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestKeyBindingsFileRoundTrip verifies that a key bindings file loads whether it holds the bare map written before
// such files carried a header or the current form, and that saving one with no factory bindings registered, as in the
// --convert run mode, writes the current form with every binding rather than none of them.
func TestKeyBindingsFileRoundTrip(t *testing.T) {
	c := check.New(t)
	c.Equal(0, len(factoryBindings), "the model tests register no factory bindings")
	want := map[string]unison.KeyBinding{
		"cut":     {KeyCode: unison.KeyX, Modifiers: mod.Command},
		"unknown": {KeyCode: unison.KeyK, Modifiers: mod.Control | mod.Shift},
	}
	p := filepath.Join(t.TempDir(), "bindings.keys")
	c.NoError(os.WriteFile(p, []byte(`{"cut":"`+want["cut"].Key()+`","unknown":"`+want["unknown"].Key()+`"}`), 0o600))
	b, err := NewKeyBindingsFromFS(nil, p)
	c.NoError(err)
	c.Equal(want, b.data, "the bare map loads")

	c.NoError(b.Save(p))
	h, _, err := jio.ReadHeader(nil, p)
	c.NoError(err)
	c.Equal(jio.NewHeader(KeyBindingsDataKind), h)
	b, err = NewKeyBindingsFromFS(nil, p)
	c.NoError(err)
	c.Equal(want, b.data, "every binding survives a save and load")
}
