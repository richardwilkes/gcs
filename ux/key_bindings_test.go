// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestApplyKeyBindingsStoresCanonicalText verifies that a binding spelled differently than unison would spell it is
// recognized for what it is, rather than being kept as a change from the factory default it is equal to.
func TestApplyKeyBindingsStoresCanonicalText(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	var id, key, respelled string
	for _, one := range currentKeyBindings() {
		key = one.KeyBinding.Key()
		if respelled = strings.ToUpper(key); respelled == key {
			respelled = strings.ToLower(key)
		}
		if respelled != key && one.KeyBinding == keyBindingEntries[one.ID].KeyBinding {
			id = one.ID
			break
		}
	}
	c.NotEqual("", id, "there must be a key binding at its factory default whose text has letters in it")
	t.Cleanup(func() { applyKeyBindings(&gurps.GlobalSettings().KeyBindings) })

	var untouched gurps.KeyBindings
	applyKeyBindings(&untouched)
	c.True(untouched.IsZero(), "applying the factory defaults must not record any of them as a change")
	for _, entry := range keyBindingEntries {
		c.Equal(entry.KeyBinding, entry.Action.KeyBinding,
			"%s: the binding must survive the trip through its text, not just the text", entry.ID)
	}

	var b gurps.KeyBindings
	b.Set(id, respelled)
	c.False(b.IsZero(), "the model can only compare the text")
	applyKeyBindings(&b)
	c.Equal(key, b.Current(id))
	c.True(b.IsZero(), "once respelled, the binding is seen to be the factory default")
	c.Equal(keyBindingEntries[id].KeyBinding, keyBindingEntries[id].Action.KeyBinding)
}

// TestApplyKeyBindingsDropsUnregisteredIDs verifies that a binding for an ID no action registered, which a settings
// file or an exported key bindings file from a build that had such an action may hold, is dropped rather than carried
// along for good, where the Menu Keys view couldn't show it and a later action given the ID would take it on.
func TestApplyKeyBindingsDropsUnregisteredIDs(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	t.Cleanup(func() { applyKeyBindings(&gurps.GlobalSettings().KeyBindings) })
	entry := currentKeyBindings()[0]
	custom := "ctrl+alt+shift+F19"
	c.NotEqual(custom, entry.KeyBinding.Key())

	var b gurps.KeyBindings
	c.NoError(jio.Unmarshal([]byte(`{"no.such.action":"ctrl+alt+shift+F18","`+entry.ID+`":"`+custom+`"}`), &b))
	applyKeyBindings(&b)
	data, err := jio.Marshal(&b)
	c.NoError(err)
	c.Equal(`{"`+entry.ID+`":"`+custom+`"}`, string(data), "the unregistered binding is dropped and the registered one kept")
	c.Equal(custom, entry.Action.KeyBinding.Key(), "and applied")
}
