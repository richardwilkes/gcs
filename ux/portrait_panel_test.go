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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// Space and a press open a file chooser, so only the keys and requests the portrait ignores are exercised.
func TestPortraitPanelIsAKeyboardControl(t *testing.T) {
	c := check.New(t)
	p := NewPortraitPanel(gurps.NewEntity())
	c.True(p.Focusable(), "the portrait takes the keyboard focus")
	c.Equal(portraitPanelRefKey, p.RefKey, "the focus is put back on the portrait by its reference key")
	c.Equal(role.Image, p.Accessibility.Role)
	c.Equal("Portrait", p.Accessibility.Name)
	c.NotEqual("", p.Accessibility.Description, "how to change the portrait is described")
	node := &accessibility.Node{}
	p.Accessibility.Callback(node)
	c.True(node.Actions.Has(accessibility.Press), "a screen reader is offered the press")
	c.False(p.KeyDownCallback(unison.KeyA, mod.None, false), "keys other than the control action are left to others")
	c.False(p.Accessibility.ActionCallback(accessibility.ActionRequest{Action: accessibility.ScrollIntoView}),
		"requests other than the press are left to unison")
}
