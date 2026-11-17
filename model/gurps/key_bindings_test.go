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
// -convert run mode, writes every binding rather than none of them.
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
	b, err = NewKeyBindingsFromFS(nil, p)
	c.NoError(err)
	c.Equal(want, b.data, "every binding survives a save and load")
}

// TestKeyBindingsWriteOnlyOverrides verifies that, once factory bindings are registered, only the bindings that differ
// from them are written, so that one left at its default follows any later change to that default, while one for an ID
// no longer registered is dropped.
func TestKeyBindingsWriteOnlyOverrides(t *testing.T) {
	c := check.New(t)
	saved := factoryBindings
	t.Cleanup(func() { factoryBindings = saved })
	factoryBindings = make(map[string]*Binding)
	cut := unison.KeyBinding{KeyCode: unison.KeyX, Modifiers: mod.Command}
	RegisterKeyBinding("cut", &unison.Action{Title: "Cut", KeyBinding: cut})
	RegisterKeyBinding("copy", &unison.Action{Title: "Copy", KeyBinding: unison.KeyBinding{KeyCode: unison.KeyC}})
	b := &KeyBindings{data: map[string]unison.KeyBinding{"cut": cut}}
	c.True(b.IsZero(), "bindings left at their defaults leave nothing to write")
	custom := unison.KeyBinding{KeyCode: unison.KeyF19, Modifiers: mod.Shift}
	b.data["copy"] = custom
	b.data["gone"] = custom
	data, err := jio.Marshal(b)
	c.NoError(err)
	var loaded KeyBindings
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(map[string]unison.KeyBinding{"copy": custom}, loaded.data)
}
