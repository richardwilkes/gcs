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
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/richardwilkes/canvas/codecs"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// TestPortraitImage verifies that a portrait whose bytes the current build cannot decode is neither discarded, since a
// different build may be able to decode it, nor decoded over and over, and that replacing the data lets a valid image
// be loaded afterwards.
func TestPortraitImage(t *testing.T) {
	c := check.New(t)
	data := []byte("this is not a valid image")
	var p gurps.Profile
	p.PortraitData = data
	c.Nil(portraitImage(&p))
	c.Equal(data, p.PortraitData)
	cached := p.PortraitCache
	c.NotNil(cached, "the failure is remembered")

	c.Nil(portraitImage(&p))
	c.True(cached == p.PortraitCache, "a second call must not decode again")
	c.Equal(data, p.PortraitData)

	codecs.Register() // unison installs the image decoders as the app starts, which a plain test never does.
	var buffer bytes.Buffer
	c.NoError(png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	p.SetPortraitData(buffer.Bytes())
	img := portraitImage(&p)
	c.NotNil(img)
	c.True(img == portraitImage(&p), "the decoded image is reused")

	p.SetPortraitData(nil)
	c.Nil(portraitImage(&p))
}

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
