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
)

// TestKeyBindings verifies that only the bindings which differ from their factory default are held and written, and
// that unregistered IDs are ignored.
func TestKeyBindings(t *testing.T) {
	c := check.New(t)
	const (
		first  = "test.key_bindings.first"
		second = "test.key_bindings.second"
	)
	RegisterKeyBinding(first, "cmd+A")
	RegisterKeyBinding(second, "")
	RegisterKeyBinding(first, "cmd+Z")
	t.Cleanup(func() {
		delete(factoryBindings, first)
		delete(factoryBindings, second)
	})

	var b KeyBindings
	c.True(b.IsZero())
	c.Equal("cmd+A", b.Current(first), "a second registration of an ID is ignored")
	c.Equal("", b.Current(second))
	c.Equal("", b.Current("test.key_bindings.unknown"))

	b.Set(second, "shift+B")
	b.Set("test.key_bindings.unknown", "shift+C")
	c.False(b.IsZero())
	c.Equal("shift+B", b.Current(second))
	data, err := jio.Marshal(&b)
	c.NoError(err)
	c.Equal(`{"test.key_bindings.second":"shift+B"}`, string(data), "only the changed, registered binding is written")

	var loaded KeyBindings
	c.NoError(jio.Unmarshal([]byte(`{"test.key_bindings.first":"cmd+A","test.key_bindings.second":"shift+B"}`), &loaded))
	c.False(loaded.IsZero())
	c.Equal("shift+B", loaded.Current(second))
	loaded.ResetOne(second)
	c.True(loaded.IsZero(), "a binding stored with its factory value doesn't count as a change")

	b.Set(second, "")
	c.True(b.IsZero(), "setting a binding to its factory value removes it")
	b.Set(first, "cmd+Q")
	b.Reset()
	c.Equal("cmd+A", b.Current(first))
}

// TestKeyBindingsKeepUnregisteredIDs verifies that bindings for IDs nothing has registered survive a load and save, as
// happens to every binding when a key bindings file is converted without the user interface, which is what registers
// them.
func TestKeyBindingsKeepUnregisteredIDs(t *testing.T) {
	c := check.New(t)
	var b KeyBindings
	c.NoError(jio.Unmarshal([]byte(`{"cut":"cmd+Y","save":"ctrl+alt+S"}`), &b))
	c.False(b.IsZero(), "unregistered bindings count as something to write")
	c.Equal("", b.Current("cut"), "but are not answered for")
	data, err := jio.Marshal(&b)
	c.NoError(err)
	c.Equal(`{"cut":"cmd+Y","save":"ctrl+alt+S"}`, string(data))

	p := filepath.Join(t.TempDir(), "test"+KeySettingsExt)
	c.NoError(os.WriteFile(p, data, 0o600))
	c.NoError(Convert(p))
	data, err = os.ReadFile(p)
	c.NoError(err)
	var converted KeyBindings
	c.NoError(jio.Unmarshal(data, &converted))
	c.Equal(b.data, converted.data, "a conversion with nothing registered rewrites the file with every binding in it")
}
