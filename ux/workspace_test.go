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
	"github.com/richardwilkes/unison"
)

// TestResolveDockable verifies that a Dockable is resolved to its outermost implementation, since the placement code is
// handed the embedded SettingsDockable rather than the settings view itself.
func TestResolveDockable(t *testing.T) {
	c := check.New(t)
	d := &pageRefMappingsDockable{}
	d.Self = d
	c.Equal(d, resolveDockable(&d.SettingsDockable), "an inner layer resolves to the Dockable that contains it")
	c.Equal(d, resolveDockable(d), "an already-resolved Dockable is returned unchanged")
	c.Nil(resolveDockable(nil), "a nil Dockable is returned unchanged")
}

// TestResolveDockableRestoresOptionalInterfaces verifies that the optional interfaces a settings view implements are
// visible again once the Dockable has been resolved. Only the settings view implements GroupedCloser; the embedded
// SettingsDockable that gets handed to the placement code does not, so a view given a window of its own used to be
// skipped by traverseGroup and was left open when the sheet that owns it was closed.
func TestResolveDockableRestoresOptionalInterfaces(t *testing.T) {
	c := check.New(t)
	d := &sheetSettingsDockable{}
	d.Self = d
	var placed unison.Dockable = &d.SettingsDockable
	_, ok := placed.(GroupedCloser)
	c.False(ok, "the embedded SettingsDockable is not a GroupedCloser")
	_, ok = resolveDockable(placed).(GroupedCloser)
	c.True(ok, "the resolved settings view is a GroupedCloser")
}
