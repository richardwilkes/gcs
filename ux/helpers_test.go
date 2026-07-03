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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison/enums/mod"
)

func TestNoModifiersDown(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name string
		mods mod.Modifiers
		want bool
	}{
		{name: "none", mods: mod.None, want: true},
		// A latched CapsLock or NumLock must be ignored so that plain keypresses are still recognized.
		{name: "caps lock latched", mods: mod.CapsLock, want: true},
		{name: "num lock latched", mods: mod.NumLock, want: true},
		{name: "both lock keys latched", mods: mod.CapsLock | mod.NumLock, want: true},
		// Any non-sticky modifier being down means it is not a plain keypress.
		{name: "shift", mods: mod.Shift, want: false},
		{name: "control", mods: mod.Control, want: false},
		{name: "option", mods: mod.Option, want: false},
		{name: "command", mods: mod.Command, want: false},
		// A real modifier still counts even when combined with a latched lock key.
		{name: "shift with caps lock", mods: mod.Shift | mod.CapsLock, want: false},
		{name: "command with num lock", mods: mod.Command | mod.NumLock, want: false},
	} {
		c.Equal(one.want, noModifiersDown(one.mods), one.name)
	}
}
